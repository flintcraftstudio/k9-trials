// Package competitors holds the public competitor directory (P5) and
// competitor profile (P6). View structs are presentational — counts,
// relative dates, initials, and per-row scores are precomputed in
// internal/handler.
package competitors

// DirectoryCard is one competitor in the search/directory results (P5).
type DirectoryCard struct {
	Handle         string
	DisplayName    string
	DogCount       int
	FinalizedCount int
	LastCompeted   string // "2 days ago", or "" when never competed
}

// SearchViewData backs the directory page (P5) and its results fragment.
// Query is the current search term (echoed into the input); Searched marks
// whether a term was submitted, which selects the results vs. the
// recently-active heading.
type SearchViewData struct {
	Query    string
	Searched bool
	Cards    []DirectoryCard
}

// DogTag is one dog in the profile dogs section (P6). Meta is the
// precomposed "Breed · age" line; RegNo is shown when recorded.
type DogTag struct {
	ID       int64
	CallName string
	Meta     string
	RegNo    string
}

// HistoryRow is one finalized entry in a profile event-history list.
// HasScore is false when the entry couldn't be evaluated (no template);
// the row then shows no points. Qualified drives the Q / NQ marker.
type HistoryRow struct {
	EntryID   int64
	Title     string // "Vex · Obedience · Level 2"
	Sub       string // "Cedar Creek · 14 Mar"
	Points    int
	Qualified bool
	HasScore  bool
}

// ProfileViewData backs the competitor profile page (P6).
type ProfileViewData struct {
	Handle      string
	DisplayName string
	Initials    string
	Bio         string
	DogCount    int
	Dogs        []DogTag
	History     []HistoryRow
}
