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
	Sex            string // "male", "female", or "" when unrecorded
	PublicURL      string // dog public page, edit only
	Err            string
}

// EntriesListViewData backs the account entries list (A5).
type EntriesListViewData struct {
	Total   int
	Filters []EntryFilter
	Rows    []EntryRow
}

// EntryFilter is one status chip in the entries filter row.
type EntryFilter struct {
	Key    string // "" (all) / upcoming / scoring / finalized
	Label  string
	Count  int
	Href   string
	Active bool
}

// EntryRow is one row on the list (A5) — an entry, or a pending
// registration that has no entry yet. Href is where the row links: the
// entry detail for entries, the event page for registrations.
type EntryRow struct {
	Href        string
	Title       string // "Cedar Creek · Obedience · Level 2"
	Sub         string // "Vex · entry #17 · Sat 14 Mar" (+ "· scored 182")
	EventKey    string
	StatusLabel string
	StatusKind  string // pill variant
}

// EntryDetailViewData backs the read-only entry page a competitor sees for
// their own entry (A6).
type EntryDetailViewData struct {
	EntryID   int64
	Eyebrow   string // "Obedience · Level 2 · Entry 14"
	EventName string
	EventKey  string
	DogMeta   string // "Vex · handled by L. Tanaka · 14 Mar 2026"

	// Exactly one of these states is true.
	Finalized bool
	Scoring   bool
	Pending   bool

	// Finalized payload.
	Points        int
	MaxPoints     int
	Passed        bool
	Threshold     int
	Exercises     []ExerciseLine
	JudgedBy      string
	FinalizedDate string

	// Challenge affordance (finalized only).
	CanChallenge    bool   // within window and not already filed
	ChallengeHref   string // /account/challenges/new?entry=ID
	WindowClosed    bool   // finalized but past the dispute window
	AlreadyFiled    bool
	ChallengeStatus string // status label when AlreadyFiled
}

// ExerciseLine is one scored exercise row on the entry breakdown.
type ExerciseLine struct {
	Num   int
	Name  string
	Score int
	Max   int
}

// ChallengesListViewData backs the challenges list (A7). Filters carries the
// status chip row with per-status counts; LastUpdate is the relative time of
// the most recently touched challenge, shown in the header when any exist.
type ChallengesListViewData struct {
	Total      int
	LastUpdate string // "1 hour ago" / "today", empty when no challenges
	Filters    []ChallengeFilter
	Rows       []ChallengeRow
}

// ChallengeFilter is one status chip in the challenges filter row.
type ChallengeFilter struct {
	Key    string // "" (all) / open / under_review / resolved / dismissed
	Label  string
	Count  int
	Href   string
	Active bool
}

// ChallengeRow is one filed dispute on the list. Status is the raw stored
// status; the template derives its label and pill variant.
type ChallengeRow struct {
	EntryID int64
	Title   string // "Hopkins Mill · Protection · Level 2"
	Sub     string // "Vex · entry #08 · 12 Jan"
	Filed   string // "Filed 5 days ago · admin started review yesterday"
	Status  string // open / under_review / resolved / dismissed
}

// RegisterViewData backs the event registration form (R1). Exactly one of
// the form / NoDogs / NotOpen presentations renders.
type RegisterViewData struct {
	EventName string
	EventSlug string
	DateRange string
	EventKey  string

	Dogs   []RegDogOption
	Trials []RegTrialOption
	Notes  string
	Err    string

	NoDogs     bool   // competitor owns no dogs yet
	NotOpen    bool   // event not accepting registrations
	NotOpenMsg string // reason shown in the not-open state
}

// RegDogOption is one selectable dog in the registration radio list.
type RegDogOption struct {
	ID       int64
	CallName string
	Meta     string // "Czech GSD · 4y · K9-3187"
	Selected bool
}

// RegTrialOption is one trial checkbox. Disabled is set when the selected
// dog already holds a registration for that trial.
type RegTrialOption struct {
	ID       int64
	Label    string // "Obedience · Level 2"
	Date     string // "Sat 14 Mar"
	EventKey string
	Disabled bool
	Checked  bool
}

// RegisterDoneViewData backs the post-submit confirmation (R1).
type RegisterDoneViewData struct {
	EventName string
	EventSlug string
	DogName   string
	Count     int
}

// ChallengeNewViewData backs the file-a-challenge form (A8). The disputing
// card carries a scoresheet excerpt — the result pill plus the NQ reason (or
// score summary) — so the competitor has the data in front of them while
// writing, with a link through to the full scoresheet (A6).
type ChallengeNewViewData struct {
	EntryID      int64
	DisputeTitle string // "Hopkins Mill · Protection · Level 2 · Entry 08"
	DisputeSub   string // "Vex · 12 Jan · judged by H. Vance · finalized"
	EventKey     string
	Reason       string
	Err          string

	// Scoresheet excerpt. ResultLabel/ResultKind drive the Q/NQ pill;
	// ExcerptLabel/Excerpt are the reason quote or score summary. All empty
	// when the score could not be evaluated. ScoresheetHref links to A6.
	ResultLabel    string // "NQ" / "Q"
	ResultKind     string // pill variant: closed (NQ) / qual (Q)
	ExcerptLabel   string // "NQ reason —" / "Result —"
	Excerpt        string // "\"Ring departure during courage test…\""
	ScoresheetHref string // /account/entries/{id}

	// When the entry already has a challenge from this filer, the form is
	// replaced with this notice.
	AlreadyFiled    bool
	ChallengeStatus string
}
