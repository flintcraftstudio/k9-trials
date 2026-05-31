package admin

// eventStatusLabel maps an event status to its display label.
func eventStatusLabel(status string) string {
	switch status {
	case "draft":
		return "Draft"
	case "published":
		return "Published"
	case "closed":
		return "Closed"
	}
	return status
}

// eventStatusKind maps an event status to a status-pill variant: a draft is
// muted, a published event reads as open, a closed event reads as closed.
func eventStatusKind(status string) string {
	switch status {
	case "draft":
		return "muted"
	case "published":
		return "open"
	case "closed":
		return "closed"
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
