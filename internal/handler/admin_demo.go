package handler

import (
	"log/slog"
	"net/http"

	"github.com/flintcraftstudio/k9-trials/internal/seeddemo"
	"github.com/flintcraftstudio/k9-trials/internal/session"
	"github.com/flintcraftstudio/k9-trials/internal/store"
	"github.com/flintcraftstudio/k9-trials/internal/view"
)

// AdminSeedDemo serves POST /admin/seed-demo — it wipes all application data
// and reseeds the deterministic demo world, then returns to the dashboard.
//
// The route is only registered when DEMO_MODE=1, but this re-checks
// view.DemoMode as defense in depth so the handler can never wipe data in a
// non-demo deployment even if it is wired up by mistake. The seed is
// idempotent: it always resets to the same known state.
func AdminSeedDemo(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !view.DemoMode {
			http.NotFound(w, r)
			return
		}

		if _, err := seeddemo.Run(r.Context(), st); err != nil {
			slog.Error("demo reset", "err", err)
			http.Error(w, "demo reset failed", http.StatusInternalServerError)
			return
		}

		// The reseed wiped the sessions table, including this admin's own
		// session. Re-establish one for the freshly created admin so the
		// operator stays logged in across repeated demo run-throughs. If this
		// fails we simply fall through logged out and /admin bounces to /login.
		if uid, _, _, err := st.GetUserByEmail(r.Context(), "admin@example.com"); err == nil {
			if err := session.Create(r.Context(), w, st, uid); err != nil {
				slog.Error("demo reset: re-establish admin session", "err", err)
			}
		}

		slog.Info("demo data reset by admin")
		hxRedirect(w, r, "/admin")
	}
}
