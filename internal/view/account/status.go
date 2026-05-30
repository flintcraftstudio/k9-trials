package account

// challengeLabel maps a stored challenge status to its human label. Unknown
// statuses pass through so a bad row is visible rather than mislabeled.
func challengeLabel(status string) string {
	switch status {
	case "open":
		return "Open"
	case "under_review":
		return "Under review"
	case "resolved":
		return "Resolved"
	case "dismissed":
		return "Dismissed"
	}
	return status
}

// challengeKind maps a challenge status to a status-pill variant: open
// waits, under_review is in-progress, resolved is a positive close,
// dismissed is muted.
func challengeKind(status string) string {
	switch status {
	case "open":
		return "wait"
	case "under_review":
		return "scoring"
	case "resolved":
		return "qual"
	case "dismissed":
		return "muted"
	}
	return "muted"
}
