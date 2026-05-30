package account

// This file carries the view-data structs for the competitor account
// surface (A1–A8). Handlers do all domain work (scoring, formatting) and
// hand these precomputed structs to the templates, which stay free of
// business logic.

// DashboardViewData backs the account landing page (A1).
type DashboardViewData struct {
	DisplayName    string
	Handle         string
	DogCount       int
	FinalizedCount int
	OpenChallenges int
	UpNext         *UpNextCard
	Recent         []RecentRow
}

// UpNextCard is the prominent "up next" entry on the dashboard — the
// nearest entry that has not been finalized. Nil when the competitor has
// nothing scheduled.
type UpNextCard struct {
	EntryID     int64
	EventName   string
	Meta        string // "Obedience · Level 2 · Sat 14 Mar · entry #17"
	DogName     string
	EventKey    string // discipline key for the data-event accent
	StatusLabel string // "Upcoming" / "Scoring"
	StatusKind  string // pill variant: wait / scoring
}

// RecentRow is one finalized result in the dashboard recent-results list.
type RecentRow struct {
	EntryID   int64
	Title     string // "Kestrel · Detection · Level 3"
	Sub       string // "Brindle Bay · 8 Feb"
	Points    int
	Qualified bool
	HasScore  bool
}

// ProfileViewData backs the profile editor (A2).
type ProfileViewData struct {
	DisplayName string
	Handle      string
	Bio         string
	PublicURL   string // "/competitors/ltanaka"
	Saved       bool   // render the just-saved confirmation
	Err         string // validation message, empty when none
}

// DogsListViewData backs the dogs list (A3).
type DogsListViewData struct {
	Count int
	Dogs  []DogCard
}

// DogCard is one dog row on the list.
type DogCard struct {
	ID        int64
	CallName  string
	RegNo     string
	Meta      string // "Czech GSD · 4y · last ran 2 days ago" or "no entries yet"
	PublicURL string // "/dogs/12"
}

// DogFormViewData backs the add/edit dog form (A4). On a fresh add every
// value field is empty and IsEdit is false.
type DogFormViewData struct {
	IsEdit         bool
	DogID          int64
	CallName       string
	RegisteredName string
	Breed          string
	DOB            string // ISO yyyy-mm-dd for the date input
	RegNo          string
	PublicURL      string // dog public page, edit only
	Err            string
}
