package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/flintcraftstudio/k9-trials/internal/view"
)

// Home handles GET /{$} (the root only) and renders the home page.
func Home() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := view.HomePage().Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// Fallback handles any GET path not matched by a specific route. It canonicalises
// trailing-slash URLs to their slashless form (so /admin/ reaches the /admin
// handler) and returns 404 for genuinely unknown paths — rather than silently
// rendering the home page for every unmatched request, which masks both typos
// and the trailing-slash mismatch.
func Fallback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if p := r.URL.Path; len(p) > 1 && strings.HasSuffix(p, "/") {
			target := strings.TrimRight(p, "/")
			if target == "" {
				target = "/"
			}
			if r.URL.RawQuery != "" {
				target += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, target, http.StatusMovedPermanently)
			return
		}
		http.NotFound(w, r)
	}
}
