// Package entries holds the public, read-only per-entry page (P4) at
// /entries/{id}. It mirrors the judge's B6 read-only view at a
// public-trust level of detail. View structs are presentational — the
// handler precomputes every label and runs the scoring engine.
package entries

// ExerciseLine is one row in the public score breakdown.
type ExerciseLine struct {
	Num   int
	Name  string
	Score int
	Max   int
}

// DetailViewData backs the public entry page. Exactly one of Finalized /
// Scoring / Pending is true, selecting which body the page renders.
type DetailViewData struct {
	// Header
	Eyebrow   string // "Obedience · Level 2 · Entry 14"
	EventName string
	EventSlug string
	TrialID   int64
	EventKey  string
	DogName   string
	DogMeta   string // "German Shepherd · K9-2419 · handled by J. Marsh"

	// State (exactly one true)
	Finalized bool
	Scoring   bool
	Pending   bool

	// Finalized payload
	Points        int
	MaxPoints     int
	Passed        bool
	Threshold     int
	Exercises     []ExerciseLine
	JudgedBy      string
	FinalizedDate string
}
