// Package events holds the public, read-only event surface: the events
// index (P1), event detail (P2), and trial leaderboard (P3). View structs
// here are purely presentational — every label, date, and accent key is
// precomputed in internal/handler so the templates carry no domain logic.
package events

// DisciplineTag is one discipline chip on an event card. Key is the
// data-event value (obedience/protection/tracking/detection) so the pill
// picks up the right accent.
type DisciplineTag struct {
	Label string
	Key   string
}

// DisciplineFilter is one chip in the index filter row. Code is the stored
// discipline code ("OB"…) or "" for the All chip. Href is the link the
// chip points at; Active marks the current selection.
type DisciplineFilter struct {
	Code   string
	Label  string
	Href   string
	Active bool
}

// EventCard is one event in the index list.
type EventCard struct {
	Slug        string
	Name        string
	Location    string
	DateRange   string
	TrialCount  int
	Disciplines []DisciplineTag
	RegOpen     bool
	EventKey    string // dominant discipline → rail/eyebrow accent
}

// ListViewData backs the events index page (P1) and its filtered grid
// fragment.
type ListViewData struct {
	Count   int
	Filters []DisciplineFilter
	Events  []EventCard
}

// TrialRow is one trial within an event-detail page (P2).
type TrialRow struct {
	ID            int64
	DisciplineLvl string // "Obedience · Level 2"
	EventKey      string
	Date          string // "Sat 14 Mar"
	EntryCount    int
	JudgeAssigned bool
	JudgeName     string
}

// DetailViewData backs the event detail page (P2).
type DetailViewData struct {
	Slug        string
	Name        string
	Location    string
	DateRange   string
	Disciplines []DisciplineTag
	RegOpen     bool
	LoggedIn    bool
	TrialCount  int
	Trials      []TrialRow
	EventKey    string
}

// LeaderRow is one entry on the trial leaderboard (P3). Rank is 0 for rows
// that don't carry a numbered placing (in-progress, NQ). Scoring/NQ flags
// drive the pill; Points is only meaningful when Finalized && !NQ.
type LeaderRow struct {
	EntryID   int64
	Rank      int
	DogName   string
	Handler   string
	K9ID      string
	Points    int
	Qualified bool
	NQ        bool
	Scoring   bool
	Finalized bool
}

// TrialDetailViewData backs the trial leaderboard page (P3).
type TrialDetailViewData struct {
	EventSlug      string
	EventName      string
	DisciplineLvl  string // "Obedience · Level 2"
	EventKey       string
	Date           string
	TotalEntries   int
	FinalizedCount int
	ScoringCount   int
	UpcomingCount  int
	Rows           []LeaderRow
}
