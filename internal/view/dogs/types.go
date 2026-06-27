// Package dogs holds the public dog profile (P7). View structs are
// presentational — age, owner link, and per-row scores are precomputed in
// internal/handler.
package dogs

// HistoryRow is one finalized entry in the dog trial-history list. The
// handler of record is carried per row because it isn't always the owner.
// HasScore is false when the entry couldn't be evaluated (no template).
type HistoryRow struct {
	EntryID   int64
	Title     string // "Obedience · Level 2"
	Sub       string // "Cedar Creek · L. Tanaka · 14 Mar"
	Points    int
	Max       int
	Qualified bool
	HasScore  bool
}

// ProfileViewData backs the dog profile page (P7).
type ProfileViewData struct {
	CallName       string
	RegisteredName string
	Breed          string
	Age            string // "4y", or "" when DOB unknown
	RegNo          string
	OwnerHandle    string
	OwnerName      string
	History        []HistoryRow
}
