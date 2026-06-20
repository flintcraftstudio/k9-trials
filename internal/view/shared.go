package view

import "time"

// SiteName is the display name used in templates.
const SiteName = "K9 Elements"

// Tracking IDs and Turnstile site key, set once at startup from config.
var (
	PixelID          string
	GtagID           string
	TurnstileSiteKey string
)

// DemoMode mirrors config.DemoMode, set once at startup. When true the admin
// dashboard shows the "Reset demo data" control.
var DemoMode bool

// Year returns the current year for copyright notices.
func Year() int {
	return time.Now().Year()
}
