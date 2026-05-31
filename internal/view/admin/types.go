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

// EventsListViewData backs the admin events list (D2).
type EventsListViewData struct {
	Total   int
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
