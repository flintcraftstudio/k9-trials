package handler

import (
	"fmt"
	"time"

	"github.com/flintcraftstudio/k9-trials/internal/scoring"
)

// disciplineLabel maps the stored discipline code (OB/PR/TR/DT) to its
// human label. Unknown codes pass through unchanged so a bad row is
// visible rather than silently relabeled.
func disciplineLabel(code string) string {
	switch code {
	case "OB":
		return "Obedience"
	case "PR":
		return "Protection"
	case "TR":
		return "Tracking"
	case "DT":
		return "Detection"
	}
	return code
}

// disciplineKey maps the discipline code to the data-event attribute value
// that rebinds the accent tokens (see colors_and_type.css §11). Defaults
// to obedience so the accent never resolves to an undefined var.
func disciplineKey(code string) string {
	switch code {
	case "OB":
		return "obedience"
	case "PR":
		return "protection"
	case "TR":
		return "tracking"
	case "DT":
		return "detection"
	}
	return "obedience"
}

// levelLabel renders the integer level as "Level N". Levels outside 1–3
// still render so misconfigured trials are visible.
func levelLabel(level int64) string {
	return fmt.Sprintf("Level %d", level)
}

// disciplineLevelLabel is the compact "Obedience · Level 2" string used in
// eyebrows and list rows.
func disciplineLevelLabel(code string, level int64) string {
	return disciplineLabel(code) + " · " + levelLabel(level)
}

// dateRange formats an event's [start, end] span for public display,
// collapsing shared month/year so a three-day trial reads "14–16 Mar 2026"
// rather than repeating the month. A zero or equal end date renders a
// single day.
func dateRange(start, end time.Time) string {
	start = start.UTC()
	end = end.UTC()
	if end.IsZero() || end.Equal(start) {
		return start.Format("2 Jan 2006")
	}
	sameYear := start.Year() == end.Year()
	sameMonth := sameYear && start.Month() == end.Month()
	switch {
	case sameMonth:
		// 14–16 Mar 2026
		return fmt.Sprintf("%d–%d %s", start.Day(), end.Day(), end.Format("Jan 2006"))
	case sameYear:
		// 28 Feb – 2 Mar 2026
		return fmt.Sprintf("%s – %s", start.Format("2 Jan"), end.Format("2 Jan 2006"))
	default:
		// 30 Dec 2025 – 2 Jan 2026
		return fmt.Sprintf("%s – %s", start.Format("2 Jan 2006"), end.Format("2 Jan 2006"))
	}
}

// shortDate renders a single date as "14 Mar" for dense trial/leaderboard
// rows where the year is implied by the event context.
func shortDate(t time.Time) string {
	return t.UTC().Format("2 Jan")
}

// fullDate renders "14 Mar 2026" for headers where the year matters.
func fullDate(t time.Time) string {
	return t.UTC().Format("2 Jan 2006")
}

// registrationOpen reports whether an event currently accepts
// registrations. Published events are open; closed/draft/archived are not.
func registrationOpen(eventStatus string) bool {
	return eventStatus == "published"
}

// qualified reports whether an evaluated result is a passing run. Thin
// wrapper for readability at call sites mapping leaderboard rows.
func qualified(res scoring.ScoresheetResult) bool {
	return res.Passed
}
