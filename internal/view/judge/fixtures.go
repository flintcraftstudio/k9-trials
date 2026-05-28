// Package judge holds the judge-side scoring UI: views and the hardcoded
// fixtures they currently render against. When the real persistence layer
// lands, fixtures.go is replaced by handlers that read from store + scoring.
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

// EventKey maps Discipline → the data-event attribute value so the CSS event
// scope rebinds the accent tokens. Keep in sync with colors_and_type.css §11.
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
	Name        string
	Max         float64
	Score       float64 // -1 sentinel for unscored
	Scored      bool
	Active      bool
	DeductLabel string // e.g. "anticipated −0.5", surfaces on B4/B6
	Deductions  []Deduction
	Voice       *VoiceNote
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

// === Single hardcoded fixture ===
// A spring-trial day with one judge (H. Vance) running Cedar Creek Obedience.

func trialOB() Trial {
	return Trial{
		Name:       "Cedar Creek Spring Trial",
		Class:      "Open class",
		Discipline: DiscOB,
		JudgeName:  "H. Vance",
		JudgeInits: "HV",
		InPocket:   3,
		LastSync:   "09.12",
	}
}

func trialDT() Trial {
	t := trialOB()
	t.Class = "Interior"
	t.Discipline = DiscDT
	return t
}

func allRuns() []Run {
	return []Run{
		{ID: "14", Number: 14, DogName: "Echo", DogInit: "E", DogVariant: 3, HandlerSh: "J. Marsh", Breed: "German Shepherd", K9ID: "K9-2419", Scheduled: "09.40", Status: StatusDelivered, Score: "scored 193"},
		{ID: "15", Number: 15, DogName: "Atlas", DogInit: "A", DogVariant: 1, HandlerSh: "R. Okafor", Breed: "Belgian Malinois", K9ID: "K9-3041", Scheduled: "09.55", Status: StatusDelivered, Score: "scored 187"},
		{ID: "16", Number: 16, DogName: "Saber", DogInit: "S", DogVariant: 4, HandlerSh: "K. Hessel", Breed: "Dutch Shepherd", K9ID: "K9-2877", Scheduled: "10.10", Status: StatusInPocket, Score: "scored 189"},
		{ID: "17", Number: 17, DogName: "Vex", DogInit: "V", DogVariant: 2, HandlerSh: "L. Tanaka", Breed: "Czech GSD", K9ID: "K9-3187", Scheduled: "10.25", Status: StatusInProg, Note: "at the gate"},
		{ID: "18", Number: 18, DogName: "Birch", DogInit: "B", DogVariant: 5, HandlerSh: "M. Pereira", Breed: "Rottweiler", K9ID: "K9-1944", Scheduled: "10.40", Status: StatusPending},
		{ID: "19", Number: 19, DogName: "Lumen", DogInit: "L", DogVariant: 1, HandlerSh: "S. Yi", Breed: "Belgian Malinois", K9ID: "K9-3502", Scheduled: "10.55", Status: StatusPending},
		{ID: "20", Number: 20, DogName: "Junebug", DogInit: "J", DogVariant: 3, HandlerSh: "D. Royo", Breed: "German Shepherd", K9ID: "K9-1812", Scheduled: "11.10", Status: StatusPending},
		{ID: "21", Number: 21, DogName: "Karma", DogInit: "K", DogVariant: 4, HandlerSh: "T. Frye", Breed: "Dutch Shepherd", K9ID: "withdrew 07.52", Scheduled: "11.25", Status: StatusScratch},
		{ID: "22", Number: 22, DogName: "Ferro", DogInit: "F", DogVariant: 1, HandlerSh: "A. Reyes", Breed: "Belgian Malinois", K9ID: "K9-2218", Scheduled: "11.40", Status: StatusPending},
	}
}

func findRun(id string) (Run, bool) {
	for _, r := range allRuns() {
		if r.ID == id {
			return r, true
		}
	}
	return Run{}, false
}

// Counts returns a tally for the filter chips on B1.
type Counts struct {
	All, Pending, Scoring, InPocket, Delivered, Scratch int
}

func tally(rs []Run) Counts {
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

// === Per-screen data builders ===

type QueueViewData struct {
	Trial  Trial
	Runs   []Run
	Counts Counts
}

func QueueData() QueueViewData {
	rs := allRuns()
	return QueueViewData{Trial: trialOB(), Runs: rs, Counts: tally(rs)}
}

type GateViewData struct {
	Trial Trial
	Run   Run
}

func GateData(id string) (GateViewData, bool) {
	r, ok := findRun(id)
	if !ok {
		return GateViewData{}, false
	}
	return GateViewData{Trial: trialOB(), Run: r}, true
}

type ScoreViewData struct {
	Trial      Trial
	Run        Run
	Discipline Discipline

	// Obedience two-pane state
	Exercises  []Exercise
	ActiveIdx  int // 0-based, references Exercises
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

func ScoreData(id string) (ScoreViewData, bool) {
	// Synthetic detection slot — id "09" maps to Junebug in Detection Interior
	// so the same /score route can demo the B3-D layout.
	if id == "09" {
		t := trialDT()
		dogRun := Run{ID: "09", Number: 9, DogName: "Junebug", DogInit: "J", DogVariant: 3, HandlerSh: "D. Royo", K9ID: "K9-1812"}
		return ScoreViewData{
			Trial:       t,
			Run:         dogRun,
			Discipline:  DiscDT,
			Hides:       detectionHides(),
			HideActive:  2,
			FindsMade:   2,
			FindsTotal:  4,
			FalseAlerts: 1,
			Elapsed:     "1:23",
		}, true
	}
	r, ok := findRun(id)
	if !ok {
		return ScoreViewData{}, false
	}
	return ScoreViewData{
		Trial:      trialOB(),
		Run:        r,
		Discipline: DiscOB,
		Exercises:  obedienceExercises(),
		ActiveIdx:  3, // "Recall over jump"
		Score:      94,
		ScoreMax:   200,
		NeedToPass: 170,
	}, true
}

func obedienceExercises() []Exercise {
	return []Exercise{
		{Num: 1, Name: "Heel free", Max: 10, Score: 9.0, Scored: true},
		{Num: 2, Name: "Figure 8", Max: 10, Score: 8.5, Scored: true},
		{Num: 3, Name: "Stand for exam", Max: 10, Score: 10.0, Scored: true},
		{
			Num: 4, Name: "Recall over jump", Max: 10, Score: 8.5, Scored: true, Active: true,
			Deductions: []Deduction{
				{Label: "Anticipated", Value: -0.5, Applied: true},
				{Label: "Slow recall", Value: -0.5},
				{Label: "Wide turn", Value: -1.0},
				{Label: "Touched jump", Value: -1.0},
				{Label: "Crooked sit", Value: -0.5},
				{Label: "Extra command", Value: -1.0},
				{Label: "Wide return", Value: -0.5, Recent: true},
			},
			Voice: &VoiceNote{State: "recording", Duration: "0:08", Label: "Recording"},
		},
		{Num: 5, Name: "Drop on recall", Max: 10},
		{Num: 6, Name: "Retrieve on flat", Max: 10},
		{Num: 7, Name: "Retrieve over jump", Max: 10},
		{Num: 8, Name: "Broad jump", Max: 10},
	}
}

func detectionHides() []DetectionHide {
	return []DetectionHide{
		{Num: 1, Name: "Hide 1", State: "find", Time: "0:42", Detail: "handler call"},
		{Num: 2, Name: "Hide 2", State: "find", Time: "1:08", Detail: "1 FA logged"},
		{Num: 3, Name: "Hide 3", State: "active", Time: "—", Detail: "running"},
		{Num: 4, Name: "Hide 4", State: "queued", Time: "—", Detail: "—"},
	}
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

func ReviewData(id string) (ReviewViewData, bool) {
	r, ok := findRun(id)
	if !ok {
		return ReviewViewData{}, false
	}
	return ReviewViewData{
		Trial: trialOB(),
		Run:   r,
		Exercises: []ReviewExercise{
			{Num: 1, Name: "Heel free", Score: 9.0, Max: 10, Scored: true},
			{Num: 2, Name: "Figure 8", Score: 8.5, Max: 10, Scored: true},
			{Num: 3, Name: "Stand for exam", Score: 10.0, Max: 10, Scored: true},
			{Num: 4, Name: "Recall over jump", Score: 8.5, Max: 10, Scored: true, Note: "anticipated −0.5"},
			{Num: 5, Name: "Drop on recall", Score: 9.5, Max: 10, Scored: true},
			{Num: 6, Name: "Retrieve on flat", Max: 10, Scored: false, Flagged: true, Note: "no score entered"},
			{Num: 7, Name: "Retrieve over jump", Score: 9.0, Max: 10, Scored: true},
			{Num: 8, Name: "Broad jump", Score: 9.5, Max: 10, Scored: true},
		},
		Provisional:      182.5,
		Max:              200,
		Qualifying:       170,
		UnscoredCount:    1,
		UnscoredExercise: "Retrieve on flat",
		Deductions: []ReviewDeduction{
			{Where: "Ex 2 · Figure 8", Value: -1.5},
			{Where: "Ex 4 · Recall", Value: -1.5, Tag: "anticipated"},
			{Where: "Ex 7 · Retrieve jump", Value: -1.0},
			{Where: "Ex 8 · Broad jump", Value: -0.5},
		},
	}, true
}

type SubmitViewData struct {
	Trial       Trial
	Run         Run
	Total       float64
	Qualifying  bool
	ExercisesOK string // "8/8 exercises"
	DedSummary  string // "1 deduction"
}

func SubmitData(id string) (SubmitViewData, bool) {
	r, ok := findRun(id)
	if !ok {
		return SubmitViewData{}, false
	}
	return SubmitViewData{
		Trial:       trialOB(),
		Run:         r,
		Total:       182.5,
		Qualifying:  true,
		ExercisesOK: "8/8 exercises",
		DedSummary:  "1 deduction",
	}, true
}

type LockedViewData struct {
	Trial      Trial
	Run        Run
	Total      float64
	Max        float64
	Qualifying float64
	SyncedAt   string
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

func LockedData(id string) (LockedViewData, bool) {
	r, ok := findRun(id)
	if !ok {
		return LockedViewData{}, false
	}
	d := LockedViewData{
		Trial:      trialOB(),
		Run:        r,
		Total:      193,
		Max:        200,
		Qualifying: 170,
		SyncedAt:   "13.42",
		Exercises: []ReviewExercise{
			{Num: 1, Name: "Heel free", Score: 9.5, Max: 10, Scored: true},
			{Num: 2, Name: "Figure 8", Score: 9.0, Max: 10, Scored: true},
			{Num: 3, Name: "Stand for exam", Score: 10.0, Max: 10, Scored: true},
			{Num: 4, Name: "Recall over jump", Score: 9.5, Max: 10, Scored: true, Note: "crooked sit −0.5"},
			{Num: 5, Name: "Drop on recall", Score: 9.5, Max: 10, Scored: true},
			{Num: 6, Name: "Retrieve on flat", Score: 9.0, Max: 10, Scored: true},
			{Num: 7, Name: "Retrieve over jump", Score: 9.0, Max: 10, Scored: true},
			{Num: 8, Name: "Broad jump", Score: 9.5, Max: 10, Scored: true},
		},
		Deductions: []ReviewDeduction{
			{Where: "Ex 1 · Tight turn", Value: -0.5},
			{Where: "Ex 2 · Slow sit", Value: -1.0},
			{Where: "Ex 4 · Crooked sit", Value: -0.5},
			{Where: "Ex 5 · Forged on call", Value: -0.5},
			{Where: "Ex 6 · Slow pickup", Value: -1.0},
			{Where: "Ex 7 · Touched jump", Value: -1.0},
			{Where: "Ex 8 · Touched broad", Value: -0.5},
		},
		Audit: []AuditStep{
			{Tone: "green", Title: "Sync · accepted by secretary", Body: "H. Vance iPad → Secretary tablet · payload checksummed", When: "13.42"},
			{Tone: "lock", Title: "Run submitted · locked", Body: "H. Vance · PIN ✓ · final total 193 / 200 · qualifying", When: "09.58"},
			{Title: "End-of-run critique recorded", Body: "Voice note · 0:48 · auto-transcribed (see Critique tab)", When: "09.57"},
			{Title: "Run started · identity confirmed", Body: "H. Vance confirmed " + r.DogName + " / " + r.HandlerSh + " / " + r.K9ID + " at the gate", When: "09.40"},
			{Title: "Run created from schedule", Body: "Trial secretary K. Sundaram · run #" + r.ID + " of 23", When: "07.02"},
		},
	}
	d.SubmittedBy.Name = "H. Vance"
	d.SubmittedBy.Inits = "HV"
	d.SubmittedBy.At = "09.58"
	d.SubmittedBy.Method = "PIN ✓"
	d.Critique.Duration = "0:48"
	d.Critique.RecordedAt = "09.57"
	d.Critique.Transcript = "Strong overall performance from this team. Echo's heelwork was crisp through the figure 8, holding position cleanly on both inside and outside turns. The stand for exam was textbook — dog held steady, accepted contact without weight shift. Where I'd push this team is on the retrieve sequence; pickups felt deliberate but not yet eager, costing them on the flat retrieve. Recall over the jump was clean but the front sit drifted slightly — a half-point off there. Recommend continued work on motivation drills for the retrieve and tightening the front-position finish."
	d.PerExerciseNotes = []struct {
		Title    string
		Quote    string
		Duration string
	}{
		{Title: "Exercise 4 · Recall over jump", Quote: "\"Front sit drifted left by maybe four inches…\"", Duration: "0:11"},
		{Title: "Exercise 6 · Retrieve on flat", Quote: "\"Pickup was slow — second time today I've seen that…\"", Duration: "0:08"},
	}
	return d, true
}
