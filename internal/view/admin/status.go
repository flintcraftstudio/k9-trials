package admin

import "fmt"

// countLabel renders "1 singular" for n == 1 and "N plural" otherwise, with a
// single space (never a middot) between the count and its noun.
func countLabel(n int, singular, plural string) string {
	if n == 1 {
		return "1 " + singular
	}
	return fmt.Sprintf("%d %s", n, plural)
}

// eventStatusLabel maps an event status to its display label.
func eventStatusLabel(status string) string {
	switch status {
	case "draft":
		return "Draft"
	case "published":
		return "Published"
	case "closed":
		return "Closed"
	case "archived":
		return "Archived"
	}
	return status
}

// eventStatusKind maps an event status to a status-pill variant. Color
// carries lifecycle state on the canonical scale (text still names the
// status): published is active → green (open); draft, closed, and archived
// are all neutral/inactive → gray (muted). Red (closed) is reserved for
// danger/rejected, which no event status is.
func eventStatusKind(status string) string {
	switch status {
	case "draft":
		return "muted"
	case "published":
		return "open"
	case "closed":
		return "muted"
	case "archived":
		return "muted"
	}
	return "muted"
}

// trialStatusLabel maps a trial status to its display label.
func trialStatusLabel(status string) string {
	switch status {
	case "pending":
		return "Accepting"
	case "in_progress":
		return "Running"
	case "complete":
		return "Complete"
	}
	return status
}

// trialStatusKind maps a trial status to a status-pill variant.
func trialStatusKind(status string) string {
	switch status {
	case "pending":
		return "wait"
	case "in_progress":
		return "scoring"
	case "complete":
		return "muted"
	}
	return "muted"
}
