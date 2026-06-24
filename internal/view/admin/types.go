package admin

// View-data structs for the admin surface (D1–D8). Handlers do the data
// loading and formatting; templates stay logic-free.

// DashboardViewData backs the admin landing page (D1).
type DashboardViewData struct {
	PendingRegs    int
	OpenChallenges int
	Counts         EventStatusCounts
	Published      []EventLine
	Drafts         []EventLine
}

// EventStatusCounts is the event tally by status for the dashboard and the
// events-list filter chips.
type EventStatusCounts struct {
	Total     int
	Draft     int
	Published int
	Closed    int
}

// EventLine is a compact event row on the dashboard.
type EventLine struct {
	ID     int64
	Name   string
	Meta   string // "Cedar Creek, OR · 14–16 Mar · 4 trials"
	Status string
}

// EventsListViewData backs the admin events list (D2). Active is the current
// status filter ("" = all) and Query the current search term; both feed the
// chip hrefs and the search box so filter + search compose.
type EventsListViewData struct {
	Total   int
	Active  string
	Query   string
	Filters []EventFilter
	Rows    []EventRow
}

// EventFilter is one status chip on the events list.
type EventFilter struct {
	Key    string
	Label  string
	Count  int
	Href   string
	Active bool
}

// EventRow is one event line on the admin list.
type EventRow struct {
	ID       int64
	Name     string
	Slug     string
	Location string
	Dates    string
	Trials   int
	Status   string
}

// EventFormViewData backs the create/edit event form (D3).
type EventFormViewData struct {
	IsEdit    bool
	EventID   int64
	Name      string
	Slug      string
	Location  string
	StartDate string // yyyy-mm-dd
	EndDate   string // yyyy-mm-dd
	Status    string
	Err       string
	Saved     bool

	// Edit-only at-a-glance.
	TrialCount  int
	PendingRegs int
	PublicURL   string
}

// TrialsViewData backs the admin trials list for an event (D4).
type TrialsViewData struct {
	EventID     int64
	EventName   string
	EventStatus string
	EventSlug   string
	TrialCount  int
	Days        []TrialDay
}

// TrialDay groups trials run on the same date.
type TrialDay struct {
	Label  string // "Saturday · 14 March"
	Count  int
	Trials []TrialLine
}

// TrialLine is one trial row on the admin trials list.
type TrialLine struct {
	ID       int64
	Title    string // "Obedience · Level 2"
	Meta     string // "Template v2026.1 · 12 entries"
	EventKey string
	Status   string
	Judge    string // judge display name, empty when unassigned
}

// TrialFormViewData backs the create-trial form (D4).
type TrialFormViewData struct {
	EventID         int64
	EventName       string
	Discipline      string
	Level           string
	Date            string
	TemplateVersion string
	Err             string
}

// --- D5 Registrations review ---

// RegistrationsViewData backs the registration review screen (D5).
type RegistrationsViewData struct {
	EventID   int64
	EventName string
	Counts    RegStatusCounts
	Trials    []RegTrialGroup
}

// RegStatusCounts is the registration tally by status.
type RegStatusCounts struct {
	Total      int
	Pending    int
	Accepted   int
	Waitlisted int
	Rejected   int
	Withdrawn  int
}

// RegTrialGroup is a trial accordion section with its registration rows.
type RegTrialGroup struct {
	Title    string
	EventKey string
	Pending  int
	Rows     []RegRow
}

// RegRow is one registration in the review list.
type RegRow struct {
	ID          int64
	DogName     string
	DogMeta     string // "K9-3187 · Czech GSD"
	Owner       string // "owner @ltanaka"
	SubmittedBy string // "Submitted by ... · 3 hours ago"
	Status      string
	EntryNumber string // "entry #17" once accepted, else ""
	Pending     bool
}

// --- D6 Judge assignments ---

// AssignmentsViewData backs the judge assignment screen (D6).
type AssignmentsViewData struct {
	EventID        int64
	EventName      string
	Unassigned     int
	AssignedJudges int // distinct judges currently assigned (drives Notify)
	Trials         []AssignTrial
	Judges         []JudgeOption
}

// AssignTrial is one trial's assignment row.
type AssignTrial struct {
	ID       int64
	Title    string
	EventKey string
	Entries  int
	JudgeID  int64 // current judge, 0 when none
	Assigned bool
}

// JudgeOption is one selectable judge.
type JudgeOption struct {
	ID   int64
	Name string
}

// --- D7 Challenge review ---

// ChallengesViewData backs the admin challenge queue (D7). Selected is nil
// when no challenge is open in the detail panel. Filters, Sorts, and Page
// drive the queue's status filtering, sorting, and pagination controls.
type ChallengesViewData struct {
	Counts   ChalStatusCounts
	Filters  []ChalFilter
	Sorts    []ChalSortLink
	Rows     []ChalRow
	Page     ChalPage
	Selected *ChalDetail
}

// ChalStatusCounts is the challenge tally by status.
type ChalStatusCounts struct {
	Open        int
	UnderReview int
	Resolved    int
	Dismissed   int
}

// ChalFilter is one status filter chip with its match count.
type ChalFilter struct {
	Label  string
	Count  int
	Href   string
	Active bool
}

// ChalSortLink is one sort option in the queue's sort control.
type ChalSortLink struct {
	Label  string
	Href   string
	Active bool
}

// ChalPage is the queue's pagination state. From/To are 1-based row indices
// within the filtered total (both 0 when the page is empty).
type ChalPage struct {
	From     int
	To       int
	Total    int
	HasPrev  bool
	HasNext  bool
	PrevHref string
	NextHref string
}

// ChalRow is one challenge in the queue.
type ChalRow struct {
	ID       int64
	Title    string // "Vex · Obedience L2"
	Sub      string // "Cedar Creek · @ltanaka · 5 days ago"
	Status   string
	Href     string // detail link, preserving the active filter/sort/page
	Selected bool
}

// ChalDetail is the selected challenge in the detail panel.
type ChalDetail struct {
	ID              int64
	Title           string // "Vex · Obedience · Level 2"
	Status          string
	Filed           string // "Filed by @ltanaka · 5 days ago · review started yesterday"
	EntryID         int64
	EntryTitle      string // "Cedar Creek · Obedience · Level 2 · Entry 08 · 12 Jan"
	EntrySub        string // "Judged by H. Vance · finalized · result NQ"
	EventKey        string

	// Disputed-entry excerpt: the NQ reason (or score summary) the judge's
	// result produced, so the admin sees what is being disputed. Empty when
	// the score could not be evaluated.
	ExcerptLabel string // "NQ reason —" / "Result —"
	Excerpt      string // "\"Ring departure during courage test…\""

	Reason          string
	ResolutionNotes string
	CanStart        bool // status is open
	CanClose        bool // status is open or under_review

	// Timeline is the audit trail of the dispute: entry finalized → challenge
	// filed → review started → terminal/pending state.
	Timeline []ChalAuditStep
}

// ChalAuditStep is one entry in the challenge audit timeline.
type ChalAuditStep struct {
	Title string // "Entry finalized · result NQ"
	Meta  string // "H. Vance" / "@ltanaka · re-score request"
	When  string // "12 Jan" / "5 days ago" / "—"
	Kind  string // dot color: lock / warn / green / muted / "" (pending)
}

// --- D8 Users and roles ---

// UsersViewData backs the users and roles admin (D8). Active is the current
// role filter ("" = all) and Query the current search term; both feed the
// chip hrefs and the search box so role filter + search compose.
type UsersViewData struct {
	Total   int
	Active  string
	Query   string
	Filters []UserFilter
	Rows    []UserRow
}

// UserFilter is one role filter chip.
type UserFilter struct {
	Key    string
	Label  string
	Count  int
	Href   string
	Active bool
}

// UserRow is one user line with an inline role control.
type UserRow struct {
	ID      int64
	Name    string // display name, or email local part
	Sub     string // "@handle" or role note
	Email   string
	Created string
	Role    string
	Handle  string // public-profile handle, empty when none
	IsSelf  bool   // the logged-in admin cannot change their own role
}
