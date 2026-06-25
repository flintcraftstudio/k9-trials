// Package judge holds the judge-side scoring UI: view types consumed by
// the B1–B6 templ panels. Construction of these view structs lives in
// internal/handler — this package is purely the rendering surface.
package judge

// Discipline codes match internal/scoring.Discipline values
// (OB / PR / TR / DT).
type Discipline string

const (
	DiscOB Discipline = "OB"
	DiscPR Discipline = "PR"
	DiscTR Discipline = "TR"
	DiscDT Discipline = "DT"
)

// EventKey maps Discipline → the data-event attribute value so the CSS
// event scope rebinds the accent tokens. Keep in sync with
// colors_and_type.css §11.
func (d Discipline) EventKey() string {
	switch d {
	case DiscOB:
		return "obedience"
	case DiscPR:
		return "protection"
	case DiscTR:
		return "tracking"
	case DiscDT:
		return "detection"
	}
	return "obedience"
}

func (d Discipline) Label() string {
	switch d {
	case DiscOB:
		return "Obedience"
	case DiscPR:
		return "Protection"
	case DiscTR:
		return "Tracking"
	case DiscDT:
		return "Detection"
	}
	return string(d)
}

// Status is the run's current life-cycle state used by the queue + chrome.
type Status string

const (
	StatusPending   Status = "pending"
	StatusInProg    Status = "inprog"
	StatusInPocket  Status = "unsynced"
	StatusDelivered Status = "synced"
	StatusScratch   Status = "scratch"
)

func (s Status) Label() string {
	switch s {
	case StatusPending:
		return "Pending"
	case StatusInProg:
		return "Scoring now"
	case StatusInPocket:
		return "In pocket"
	case StatusDelivered:
		return "Delivered"
	case StatusScratch:
		return "Scratched"
	}
	return string(s)
}

// Trial is the top-level competition envelope visible across all panels.
type Trial struct {
	Name       string
	Class      string // "Open class", "Interior", etc.
	Discipline Discipline
	JudgeName  string
	JudgeInits string
	InPocket   int    // unsynced runs waiting on the iPad
	LastSync   string // human time string like "09.12"
}

// Run is one queue entry. Score and Pass mirror what the engine would return.
type Run struct {
	ID         string
	Number     int
	DogName    string
	DogInit    string
	DogVariant int // 1..5, picks the avatar gradient
	HandlerSh  string
	Breed      string
	K9ID       string
	Scheduled  string
	Status     Status
	Score      string // empty if unscored; "187" or "scored 187"
	Note       string // optional extra line, e.g. "at the gate"
}

// Exercise is one row inside a scoresheet.
type Exercise struct {
	Num         int
	Code        string // stable rulebook code, e.g. "1.1" — used for nav + scoring posts
	Name        string
	Max         float64
	Score       float64 // -1 sentinel for unscored
	Scored      bool
	Active      bool
	Criteria    []Criterion // per-criterion entry rows for the active pane
	Triggers    []Trigger   // auto-NQ triggers available on this exercise
	NQ          bool        // exercise score forced to 0 by a fired trigger (own or phase)
	NQReason    string      // human-readable cause, e.g. "NQ — Dog leaves the field"
	DeductLabel string      // e.g. "anticipated −0.5", surfaces on B4/B6
	Deductions  []Deduction
	Voice       *VoiceNote
}

// Trigger is one auto-NQ trigger the judge can flag against an exercise.
// Scope is "exercise" (zeroes this exercise), "phase" (zeroes the whole
// phase), or "trial" (NQs the entire run).
type Trigger struct {
	Code  string
	Desc  string
	Scope string
	Fired bool
}

// Criterion is one scored line within a CriteriaSum exercise. The judge
// taps a value 0..Max; the exercise score is the sum across criteria.
type Criterion struct {
	Code   string // e.g. "1.1.a"
	Desc   string // judge-facing description
	Max    int    // whole points
	Points int    // current value (0 when unscored)
	Scored bool   // an input row exists for this criterion
}

// Deduction is one applied or available deduction in the catalog.
type Deduction struct {
	Label   string
	Value   float64 // negative, e.g. -0.5
	Applied bool    // currently selected
	Recent  bool    // surfaced as "recently used at this trial"
}

// VoiceNote is the inline per-exercise recording.
type VoiceNote struct {
	State    string // "idle" | "recording" | "done"
	Duration string // "0:08"
	Label    string // "Recording" | "Saved" | "Tap mic"
}

// DetectionHide is one find on the detection scoresheet (B3-D).
type DetectionHide struct {
	Num    int
	Name   string // odor name, e.g. "Pseudo-Eucalyptol"
	Loc    string // placement description
	State  string // "find" | "active" | "queued"
	Time   string // "0:42" or "—"
	Detail string // "handler call", "1 FA logged", "— running"
}

// AuditStep is one entry in the locked-run audit chain (B6).
type AuditStep struct {
	Title string
	Body  string
	When  string
	Tone  string // "" | "green" | "lock"
}

// Counts is the tally for the filter chips on B1.
type Counts struct {
	All, Pending, Scoring, InPocket, Delivered, Scratch int
}

// Tally counts runs by status. Exported so handlers can build it from
// real db rows without reimplementing the enum mapping.
func Tally(rs []Run) Counts {
	c := Counts{All: len(rs)}
	for _, r := range rs {
		switch r.Status {
		case StatusPending:
			c.Pending++
		case StatusInProg:
			c.Scoring++
		case StatusInPocket:
			c.InPocket++
		case StatusDelivered:
			c.Delivered++
		case StatusScratch:
			c.Scratch++
		}
	}
	return c
}

// === Per-screen view structs ===

type QueueViewData struct {
	Trial  Trial
	Runs   []Run
	Counts Counts
}

type GateViewData struct {
	Trial Trial
	Run   Run
}

type ScoreViewData struct {
	Trial      Trial
	Run        Run
	Discipline Discipline

	// Obedience two-pane state
	Exercises  []Exercise
	ActiveIdx  int  // 0-based, references Exercises
	TrialNQ    bool // a trial-scoped trigger fired — the whole run is NQ
	Score      float64
	ScoreMax   float64
	NeedToPass float64

	// Detection hides
	Hides       []DetectionHide
	HideActive  int
	FindsMade   int
	FindsTotal  int
	FalseAlerts int
	Elapsed     string
}

type ReviewViewData struct {
	Trial            Trial
	Run              Run
	Exercises        []ReviewExercise
	Provisional      float64
	Max              float64
	Qualifying       float64
	Deductions       []ReviewDeduction
	UnscoredCount    int
	UnscoredExercise string
}

type ReviewExercise struct {
	Num     int
	Name    string
	Score   float64
	Max     float64
	Scored  bool
	Flagged bool
	Note    string // small text like "anticipated −0.5" / "no score entered"
}

type ReviewDeduction struct {
	Where string // "Ex 2 · Figure 8"
	Value float64
	Tag   string // "(anticipated)"
}

type SubmitViewData struct {
	Trial       Trial
	Run         Run
	Total       float64
	Qualifying  bool
	ExercisesOK string // "8/8 exercises"
	DedSummary  string // "1 deduction"
}

type LockedViewData struct {
	Trial       Trial
	Run         Run
	Total       float64
	Max         float64
	Qualifying  float64
	SyncedAt    string
	SubmittedBy struct {
		Name   string
		Inits  string
		At     string
		Method string // "PIN ✓"
	}
	Exercises  []ReviewExercise
	Deductions []ReviewDeduction
	Audit      []AuditStep
	Critique   struct {
		Duration   string
		RecordedAt string
		Transcript string
	}
	PerExerciseNotes []struct {
		Title    string
		Quote    string
		Duration string
	}
}
