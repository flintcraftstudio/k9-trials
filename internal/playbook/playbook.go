// Package playbook serves the K9 Elements demo playbook — the client-facing
// feature walkthrough and scripted user stories (source: docs/demo-playbook.md)
// rendered to a standalone, K9 Elements-branded HTML page — as an in-app page.
//
// It exists purely as a development and demo convenience so the walkthrough is
// reachable from the running app, and is only wired into the router when
// DEMO_MODE is enabled (see cmd/server/main.go). The HTML is embedded into the
// binary so it ships with the app and needs no files on disk at runtime.
//
// To regenerate the page, edit docs/demo-playbook.md and rebuild the HTML into
// internal/playbook/demo-playbook.html.
package playbook

import (
	_ "embed"
	"net/http"
)

//go:embed demo-playbook.html
var page []byte

// Handler serves the embedded demo playbook HTML.
func Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(page)
	}
}
