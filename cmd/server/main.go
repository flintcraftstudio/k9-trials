package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/flintcraftstudio/k9-trials/internal/config"
	"github.com/flintcraftstudio/k9-trials/internal/handler"
	"github.com/flintcraftstudio/k9-trials/internal/mail"
	"github.com/flintcraftstudio/k9-trials/internal/middleware"
	"github.com/flintcraftstudio/k9-trials/internal/seeddemo"
	"github.com/flintcraftstudio/k9-trials/internal/session"
	"github.com/flintcraftstudio/k9-trials/internal/store"
	"github.com/flintcraftstudio/k9-trials/internal/view"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := loadEnv(".env"); err != nil {
		slog.Error("env error", "err", err)
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config error", "err", err)
		os.Exit(1)
	}

	// Tracking pixels
	view.GtagID = cfg.GtagID
	view.PixelID = cfg.PixelID
	if cfg.GtagID == "" {
		slog.Warn("GTAG_ID not set, Google Analytics disabled")
	}
	if cfg.PixelID == "" {
		slog.Warn("PIXEL_ID not set, Facebook Pixel disabled")
	}

	// Turnstile
	view.TurnstileSiteKey = cfg.TurnstileSiteKey
	if cfg.TurnstileSiteKey == "" || cfg.TurnstileSecretKey == "" {
		slog.Warn("TURNSTILE_SITE_KEY or TURNSTILE_SECRET_KEY not set, Turnstile disabled")
	}

	// Session cookie security
	session.Secure = cfg.CookieSecure
	if !cfg.CookieSecure {
		slog.Warn("COOKIE_INSECURE set, session cookie Secure flag disabled (dev only)")
	}

	// Demo mode — exposes the admin "Reset demo data" endpoint.
	view.DemoMode = cfg.DemoMode
	if cfg.DemoMode {
		slog.Warn("DEMO_MODE enabled, admin can wipe and reseed all data")
	}

	// Database
	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0755); err != nil {
		slog.Error("failed to create database directory", "err", err)
		os.Exit(1)
	}
	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		slog.Error("database open error", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// Enable WAL mode and foreign keys for SQLite
	if _, err := db.Exec("PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;"); err != nil {
		slog.Error("database pragma error", "err", err)
		os.Exit(1)
	}

	// Run migrations
	goose.SetBaseFS(nil)
	if err := goose.SetDialect("sqlite3"); err != nil {
		slog.Error("goose dialect error", "err", err)
		os.Exit(1)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		slog.Error("migration error", "err", err)
		os.Exit(1)
	}
	slog.Info("migrations applied")

	// Store
	st := store.New(db)

	// Demo bootstrap: a fresh deploy has only the schema and no users, so
	// nobody could log in to trigger the admin reset. When DEMO_MODE is on,
	// seed the demo world if the database is empty. Subsequent resets go
	// through the admin "Reset demo data" button.
	if cfg.DemoMode {
		var users int
		if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&users); err != nil {
			slog.Error("demo bootstrap: count users", "err", err)
		} else if users == 0 {
			if _, err := seeddemo.Run(context.Background(), st); err != nil {
				slog.Error("demo bootstrap: seed", "err", err)
			} else {
				slog.Info("demo bootstrap: empty database seeded (admin@example.com / admin1234)")
			}
		}
	}

	// Mail client (nil if Postmark is not configured)
	var mailer *mail.Client
	if cfg.PostmarkToken != "" {
		mailer = mail.NewClient(cfg.PostmarkToken, cfg.PostmarkFrom, cfg.PostmarkTo)
		slog.Info("postmark configured")
	} else {
		slog.Info("postmark not configured, contact form emails disabled")
	}

	mux := http.NewServeMux()

	// Static files
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Pages
	mux.Handle("GET /{$}", handler.Home())
	mux.Handle("GET /", handler.Fallback())
	mux.Handle("GET /contact", handler.Contact())
	mux.Handle("POST /contact", handler.ContactSubmit(mailer, cfg.TurnstileSecretKey))

	// Auth
	mux.Handle("GET /login", handler.LoginPage())
	mux.Handle("POST /login", handler.LoginSubmit(st))
	mux.Handle("POST /logout", handler.Logout(st))
	mux.Handle("GET /signup", handler.SignupPage())
	mux.Handle("POST /signup", handler.SignupSubmit(st))
	mux.Handle("GET /signup/handle", handler.SignupHandleCheck(st))

	// Public — events, competitors, dogs
	mux.Handle("GET /events", handler.EventsList(st))
	mux.Handle("GET /events/{slug}", handler.EventDetail(st))
	mux.Handle("GET /events/{slug}/trials/{id}", handler.TrialDetail(st))
	mux.Handle("GET /entries/{id}", handler.EntryDetail(st))
	mux.Handle("GET /competitors", handler.CompetitorSearch(st))
	mux.Handle("GET /competitors/{handle}", handler.CompetitorProfile(st))
	mux.Handle("GET /dogs/{id}", handler.DogProfile(st))

	// Competitor account — requires competitor or admin role
	competitor := func(h http.Handler) http.Handler {
		return session.RequireRole(h, "competitor", "admin")
	}
	mux.Handle("GET /account", competitor(handler.AccountDashboard(st)))
	mux.Handle("GET /account/profile", competitor(handler.AccountProfile(st)))
	mux.Handle("POST /account/profile", competitor(handler.AccountProfileSave(st)))
	mux.Handle("GET /account/dogs", competitor(handler.AccountDogs(st)))
	mux.Handle("POST /account/dogs", competitor(handler.AccountDogsCreate(st)))
	mux.Handle("GET /account/dogs/new", competitor(handler.AccountDogsNew(st)))
	mux.Handle("GET /account/dogs/{id}/edit", competitor(handler.AccountDogsEdit(st)))
	mux.Handle("POST /account/dogs/{id}", competitor(handler.AccountDogsUpdate(st)))
	mux.Handle("POST /account/dogs/{id}/delete", competitor(handler.AccountDogsDelete(st)))
	mux.Handle("GET /account/entries", competitor(handler.AccountEntries(st)))
	mux.Handle("GET /account/entries/{id}", competitor(handler.AccountEntryDetail(st)))
	mux.Handle("POST /account/entries/{id}/withdraw", competitor(handler.AccountEntryWithdraw(st)))
	mux.Handle("GET /account/challenges", competitor(handler.AccountChallenges(st)))
	mux.Handle("GET /account/challenges/new", competitor(handler.AccountChallengeNew(st)))
	mux.Handle("POST /account/challenges", competitor(handler.AccountChallengeSubmit(st)))

	// Event registration (competitor-side) — lives under /events/{slug}/register
	mux.Handle("GET /events/{slug}/register", competitor(handler.RegisterPage(st)))
	mux.Handle("GET /events/{slug}/register/trials", competitor(handler.RegisterTrials(st)))
	mux.Handle("POST /events/{slug}/register", competitor(handler.RegisterSubmit(st)))
	mux.Handle("POST /events/{slug}/register/notify", competitor(handler.RegisterNotify(st)))

	// Judge-side scoring UI (B1–B6 panels). All routes load real entries
	// from store + run the scoring engine; access requires the judge or
	// admin role.
	mux.Handle("GET /judge", session.RequireJudge(handler.JudgeQueue(st)))
	mux.Handle("GET /judge/entry/{id}/gate", session.RequireJudge(handler.JudgeGate(st)))
	mux.Handle("GET /judge/entry/{id}/score", session.RequireJudge(handler.JudgeScore(st)))
	mux.Handle("GET /judge/entry/{id}/review", session.RequireJudge(handler.JudgeReview(st)))
	mux.Handle("GET /judge/entry/{id}/submit", session.RequireJudge(handler.JudgeSubmit(st)))
	mux.Handle("GET /judge/entry/{id}/locked", session.RequireJudge(handler.JudgeLocked(st)))

	// Admin — events, trials, registrations, challenges, users
	mux.Handle("GET /admin", session.RequireAdmin(handler.AdminDashboard(st)))
	mux.Handle("GET /admin/events", session.RequireAdmin(handler.AdminEvents(st)))
	mux.Handle("POST /admin/events", session.RequireAdmin(handler.AdminEventsCreate(st)))
	mux.Handle("GET /admin/events/new", session.RequireAdmin(handler.AdminEventsNew()))
	mux.Handle("GET /admin/events/slug-check", session.RequireAdmin(handler.AdminEventsSlugCheck(st)))
	mux.Handle("GET /admin/events/{id}/edit", session.RequireAdmin(handler.AdminEventsEdit(st)))
	mux.Handle("POST /admin/events/{id}", session.RequireAdmin(handler.AdminEventsUpdate(st)))
	mux.Handle("POST /admin/events/{id}/archive", session.RequireAdmin(handler.AdminEventsArchive(st)))
	mux.Handle("POST /admin/events/{id}/unarchive", session.RequireAdmin(handler.AdminEventsRestore(st)))
	mux.Handle("GET /admin/events/{id}/trials", session.RequireAdmin(handler.AdminTrials(st)))
	mux.Handle("POST /admin/events/{id}/trials", session.RequireAdmin(handler.AdminTrialsCreate(st)))
	mux.Handle("GET /admin/events/{id}/trials/new", session.RequireAdmin(handler.AdminTrialsNew(st)))
	mux.Handle("POST /admin/events/{id}/trials/{tid}/delete", session.RequireAdmin(handler.AdminTrialsDelete(st)))
	mux.Handle("GET /admin/events/{id}/registrations", session.RequireAdmin(handler.AdminRegistrations(st)))
	mux.Handle("GET /admin/events/{id}/assignments", session.RequireAdmin(handler.AdminAssignments(st)))
	mux.Handle("POST /admin/events/{id}/trials/{tid}/judge", session.RequireAdmin(handler.AdminAssignJudge(st)))
	mux.Handle("POST /admin/events/{id}/notify-judges", session.RequireAdmin(handler.AdminNotifyJudges(st)))
	mux.Handle("POST /admin/registrations/{rid}/accept", session.RequireAdmin(handler.AdminRegistrationAccept(st)))
	mux.Handle("POST /admin/registrations/{rid}/waitlist", session.RequireAdmin(handler.AdminRegistrationWaitlist(st)))
	mux.Handle("POST /admin/registrations/{rid}/reject", session.RequireAdmin(handler.AdminRegistrationReject(st)))
	mux.Handle("POST /admin/registrations/{rid}/confirm-withdrawal", session.RequireAdmin(handler.AdminRegistrationConfirmWithdrawal(st)))
	mux.Handle("GET /admin/challenges", session.RequireAdmin(handler.AdminChallenges(st)))
	mux.Handle("GET /admin/challenges/{id}", session.RequireAdmin(handler.AdminChallenges(st)))
	mux.Handle("POST /admin/challenges/{id}/status", session.RequireAdmin(handler.AdminChallengeStatus(st)))
	mux.Handle("GET /admin/users", session.RequireAdmin(handler.AdminUsers(st)))
	mux.Handle("POST /admin/users/{id}/role", session.RequireAdmin(handler.AdminUserRole(st)))

	// Demo reset — only registered when DEMO_MODE=1, and still admin-gated.
	if cfg.DemoMode {
		mux.Handle("POST /admin/seed-demo", session.RequireAdmin(handler.AdminSeedDemo(st)))
	}

	// Session + logging middleware
	srv := session.Middleware(st)(mux)
	srv = middleware.Logging(logger)(srv)

	// --- Graceful shutdown sequence ---

	// 1. Configure the HTTP server with timeouts to bound slow clients.
	//    ReadTimeout:  max time to read the entire request (headers + body).
	//    WriteTimeout: max time to write the response.
	//    IdleTimeout:  max time a keep-alive connection sits idle.
	server := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      srv,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 2. Start serving in a background goroutine. Any fatal listen error
	//    (e.g. port already in use) is sent to errCh so we can react.
	errCh := make(chan error, 1)
	go func() {
		slog.Info("server starting", "addr", cfg.Addr())
		fmt.Printf("listening on %s\n", cfg.Addr())
		errCh <- server.ListenAndServe()
	}()

	// 3. Register for SIGINT (Ctrl-C) and SIGTERM (Docker/systemd stop).
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 4. Block until we receive a shutdown signal or a server error.
	select {
	case sig := <-quit:
		slog.Info("shutdown signal received", "signal", sig)
	case err := <-errCh:
		// ErrServerClosed is expected after Shutdown(); anything else is fatal.
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}

	// 5. Begin graceful shutdown: stop accepting new connections and give
	//    in-flight requests up to 10 seconds to complete.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		// 6. Deadline exceeded — force-close remaining connections.
		slog.Error("shutdown deadline exceeded, forcing close", "err", err)
		server.Close()
		os.Exit(1)
	}

	slog.Info("server stopped gracefully")
}

// loadEnv reads a .env file and sets environment variables if not already set.
func loadEnv(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, line := range splitLines(string(data)) {
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		key, val, ok := splitOnce(line, '=')
		if !ok {
			continue
		}
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
	return nil
}

// splitLines splits a string into non-empty lines.
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// splitOnce splits a string on the first occurrence of sep.
func splitOnce(s string, sep byte) (string, string, bool) {
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			return s[:i], s[i+1:], true
		}
	}
	return "", "", false
}
