package handler

import "net/http"

// hxRedirect performs a post-action redirect that is correct for both htmx
// and plain form submissions.
//
// For an htmx request, emitting a 303 is wrong: htmx issues the request
// over XHR, and XHR transparently follows the 303 to the destination
// before htmx can inspect the response. htmx then swaps the destination
// page into the triggering element (the form) instead of navigating. So
// for htmx we return 200 with an HX-Redirect header and let htmx perform a
// client-side location change; non-htmx callers get a normal 303.
func hxRedirect(w http.ResponseWriter, r *http.Request, url string) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", url)
		return
	}
	http.Redirect(w, r, url, http.StatusSeeOther)
}
