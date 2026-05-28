// seed-trial seeds a demo event + trial + entries against the local
// SQLite database so the judge UI has something real to render. Safe to
// re-run: users, events, trials, and entries are upserted by their
// natural keys; criterion scores are only inserted when an entry has none.
//
// Usage:
//
//	mage seedtrial            # uses DB_PATH or ./data/app.db
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"

	"github.com/flintcraftstudio/k9-trials/internal/db"
	"github.com/flintcraftstudio/k9-trials/internal/scoring"
	"github.com/flintcraftstudio/k9-trials/internal/scoring/templates"
	"github.com/flintcraftstudio/k9-trials/internal/store"
)

const (
	adminEmail = "admin@example.com"
	adminPass  = "admin1234"
	judgeEmail = "judge@example.com"
	judgePass  = "judge1234"

	eventSlug    = "cedar-creek-spring-2026"
	eventName    = "Cedar Creek Spring Trial"
	eventLoc     = "Cedar Creek Field, MT"
	eventStart   = "2026-05-15"
	eventEnd     = "2026-05-16"
	trialDateStr = "2026-05-15"
	templateVer  = "2026.1"
)

// seedEntry is one row from the fixture roster, kept in seed code so we
// can replicate the original demo dataset deterministically.
type seedEntry struct {
	Number  int64
	Dog     string
	Breed   string
	Handler string
	Status  string // registered | scoring | finalized
}

var seedRoster = []seedEntry{
	{14, "Echo", "German Shepherd", "Jordan Marsh", "finalized"},
	{15, "Atlas", "Belgian Malinois", "Rita Okafor", "finalized"},
	{16, "Saber", "Dutch Shepherd", "Kai Hessel", "finalized"},
	{17, "Vex", "Czech GSD", "Lia Tanaka", "scoring"},
	{18, "Birch", "Rottweiler", "Mara Pereira", "registered"},
	{19, "Lumen", "Belgian Malinois", "Sam Yi", "registered"},
	{20, "Junebug", "German Shepherd", "Dev Royo", "registered"},
	{21, "Karma", "Dutch Shepherd", "Theo Frye", "registered"},
	{22, "Ferro", "Belgian Malinois", "Ana Reyes", "registered"},
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/app.db"
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return fmt.Errorf("ensure db dir: %w", err)
	}

	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer conn.Close()
	if _, err := conn.Exec("PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;"); err != nil {
		return fmt.Errorf("pragma: %w", err)
	}

	// Run migrations so a fresh DB has its schema before seeding.
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("goose dialect: %w", err)
	}
	if err := goose.Up(conn, "migrations"); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}

	st := store.New(conn)
	ctx := context.Background()

	adminID, err := ensureUser(ctx, st, adminEmail, adminPass, "admin")
	if err != nil {
		return fmt.Errorf("admin user: %w", err)
	}
	judgeID, err := ensureUser(ctx, st, judgeEmail, judgePass, "judge")
	if err != nil {
		return fmt.Errorf("judge user: %w", err)
	}

	event, err := ensureEvent(ctx, st, adminID)
	if err != nil {
		return fmt.Errorf("event: %w", err)
	}
	trial, err := ensureTrial(ctx, st, event.ID)
	if err != nil {
		return fmt.Errorf("trial: %w", err)
	}
	created, err := ensureEntries(ctx, st, trial.ID, judgeID)
	if err != nil {
		return fmt.Errorf("entries: %w", err)
	}

	tpl := templates.L1OB()
	for _, entry := range created {
		if entry.Status != "finalized" {
			continue
		}
		if err := ensureScores(ctx, st, entry, judgeID, tpl); err != nil {
			return fmt.Errorf("scores for entry %d: %w", entry.EntryNumber, err)
		}
	}

	fmt.Printf("Seeded: %s / %s / event=%s / trial=%d / entries=%d\n",
		adminEmail, judgeEmail, eventSlug, trial.ID, len(created))
	return nil
}

func ensureUser(ctx context.Context, st *store.Store, email, password, role string) (int64, error) {
	id, _, _, _, err := st.GetUserByEmail(ctx, email)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	return st.CreateUser(ctx, email, password, role)
}

func ensureEvent(ctx context.Context, st *store.Store, adminID int64) (db.Event, error) {
	existing, err := st.Q().GetEventBySlug(ctx, eventSlug)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return db.Event{}, err
	}
	start, _ := time.Parse("2006-01-02", eventStart)
	end, _ := time.Parse("2006-01-02", eventEnd)
	return st.Q().CreateEvent(ctx, db.CreateEventParams{
		Slug:      eventSlug,
		Name:      eventName,
		Location:  eventLoc,
		StartDate: start,
		EndDate:   end,
		Status:    "published",
		CreatedBy: adminID,
	})
}

func ensureTrial(ctx context.Context, st *store.Store, eventID int64) (db.Trial, error) {
	trials, err := st.Q().ListTrialsByEvent(ctx, eventID)
	if err != nil {
		return db.Trial{}, err
	}
	for _, t := range trials {
		if t.Discipline == "OB" && t.Level == 1 && t.TemplateVersion == templateVer {
			return t, nil
		}
	}
	trialDate, _ := time.Parse("2006-01-02", trialDateStr)
	return st.Q().CreateTrial(ctx, db.CreateTrialParams{
		EventID:         eventID,
		Discipline:      "OB",
		Level:           1,
		TrialDate:       trialDate,
		TemplateVersion: templateVer,
		Status:          "in_progress",
	})
}

func ensureEntries(ctx context.Context, st *store.Store, trialID, judgeID int64) ([]db.Entry, error) {
	existing, err := st.Q().ListEntriesByTrial(ctx, trialID)
	if err != nil {
		return nil, err
	}
	byNumber := make(map[int64]db.Entry, len(existing))
	for _, e := range existing {
		byNumber[e.EntryNumber] = e
	}

	out := make([]db.Entry, 0, len(seedRoster))
	for _, row := range seedRoster {
		if e, ok := byNumber[row.Number]; ok {
			// Keep existing entry; update status only if it drifted from seed
			// expectation (so re-runs of the seed restore the demo state).
			if e.Status != row.Status {
				updated, err := st.Q().UpdateEntryStatus(ctx, db.UpdateEntryStatusParams{
					Status: row.Status,
					ID:     e.ID,
				})
				if err != nil {
					return nil, err
				}
				e = updated
			}
			out = append(out, e)
			continue
		}
		created, err := st.Q().CreateEntry(ctx, db.CreateEntryParams{
			TrialID:     trialID,
			JudgeID:     sql.NullInt64{Int64: judgeID, Valid: true},
			EntryNumber: row.Number,
			HandlerName: row.Handler,
			DogName:     row.Dog,
			DogBreed:    row.Breed,
			Status:      row.Status,
		})
		if err != nil {
			return nil, err
		}
		out = append(out, created)
	}
	return out, nil
}

// ensureScores writes one criterion_scores row per criterion on the
// L1OB template for the given entry. Most criteria get full marks; a
// couple are dropped by 1 so the totals come out interesting (not
// 120/120). Skips entirely when the entry already has any criterion
// scores — we don't want to bloat the append-only ledger on re-seed.
func ensureScores(ctx context.Context, st *store.Store, entry db.Entry, judgeID int64, tpl scoring.ScoresheetTemplate) error {
	existing, err := st.Q().ListCriterionScoresByEntry(ctx, entry.ID)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}

	// Deliberate drops keyed by criterion code, mapped per entry number so
	// each finalized entry shows a slightly different final tally.
	drops := map[int64]map[string]int64{
		14: {"1.1.b": 1, "3.3.a": 1}, // Echo: 118/120
		15: {"2.2.c": 1, "3.2.b": 1, "4.1.d": 1}, // Atlas: 117/120
		16: {"1.3.a": 2, "4.2.b": 1}, // Saber: 117/120
	}
	entryDrops := drops[entry.EntryNumber]

	for _, ph := range tpl.Phases {
		for _, ex := range ph.Exercises {
			for _, c := range ex.Criteria {
				points := int64(c.MaxPoints)
				if d, ok := entryDrops[c.Code]; ok {
					points -= d
					if points < 0 {
						points = 0
					}
				}
				if _, err := st.Q().RecordCriterionScore(ctx, db.RecordCriterionScoreParams{
					EntryID:       entry.ID,
					ExerciseCode:  ex.Code,
					CriterionCode: c.Code,
					Points:        points,
					JudgedBy:      judgeID,
				}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
