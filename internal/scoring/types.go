// Package scoring encodes the K9 Elements rulebook scoring model.
//
// The package is structured around a clear separation between the
// rulebook spec (ScoresheetTemplate, ExerciseTemplate, etc. — what
// the rulebook says is possible at a given discipline and level)
// and the actual scoring of a run (ConcreteScoresheet plus inputs —
// what a particular team did on a particular day).
//
// Templates are hardcoded in Go and versioned (TemplateVersion). When
// the rulebook revises, a new template version ships with a release.
// Historical scoresheets retain their original TemplateVersion so they
// remain interpretable after rule changes.
//
// All scoring evaluation is pure: no I/O, no time, no randomness.
// The package is safe to call from anywhere — request handlers,
// background jobs, tests.
//
// The rulebook references in this package follow the format "§3.2",
// pointing to sections of K9_Elements_Rulebook_Working_Draft.
package scoring

import "math"

// Discipline identifies one of the four K9 Elements disciplines (§1).
type Discipline string

const (
	DisciplineOB Discipline = "OB" // Obedience
	DisciplinePR Discipline = "PR" // Protection
	DisciplineTR Discipline = "TR" // Tracking
	DisciplineDT Discipline = "DT" // Detection
)

// Level is the progression tier. Level 0 is intentionally invalid;
// callers should guard with IsValid.
type Level int

const (
	LevelOne   Level = 1
	LevelTwo   Level = 2
	LevelThree Level = 3
)

// IsValid reports whether the level is one of the three defined tiers.
func (l Level) IsValid() bool {
	return l == LevelOne || l == LevelTwo || l == LevelThree
}

// Tier is the qualitative band a score falls into (§3.1).
// Ordered worst-to-best so callers can write `tier >= TierSufficient`
// for passing checks. The zero value is TierInsufficient — the safe
// default for an unscored exercise.
type Tier int

const (
	TierInsufficient Tier = iota //  0–69%
	TierSufficient               // 70–75%
	TierGood                     // 76–85%
	TierVeryGood                 // 86–95%
	TierExcellent                // 96–100%
)

// String returns the tier's rulebook name.
func (t Tier) String() string {
	switch t {
	case TierInsufficient:
		return "Insufficient"
	case TierSufficient:
		return "Sufficient"
	case TierGood:
		return "Good"
	case TierVeryGood:
		return "Very Good"
	case TierExcellent:
		return "Excellent"
	}
	return "unknown"
}

// IsPassing reports whether the tier qualifies as a pass at the
// per-line-item level. A scoresheet still has to satisfy the
// scoresheet-wide pass conditions (§3.3) on top of line-item tiers.
func (t Tier) IsPassing() bool {
	return t >= TierSufficient
}

// Points is a whole-number point value. The rulebook rounds to nearest
// whole point with ties rounding up (§3.2). Fractional points never
// appear in the domain; rounding happens at the moment of conversion
// via RoundPoints.
type Points int

// RoundPoints applies §3.2: nearest whole, ties round up.
// Examples:
//
//	17.5 -> 18  (tie rounds up)
//	17.4 -> 17
//	17.6 -> 18
//	-0.5 -> 0   (tie rounds up toward positive infinity)
//
// This is "round half up" semantics, distinct from Go's math.Round
// (which rounds half away from zero) for negative values. Negative
// points should not occur in practice but the implementation is
// well-defined.
func RoundPoints(raw float64) Points {
	return Points(math.Floor(raw + 0.5))
}

// TemplateVersion identifies the rulebook revision a template encodes.
// Stored on every scoresheet so historical scores remain interpretable
// after rulebook revisions. Format: "YYYY.N" — year of the competition
// season, sequential revision number within that season.
type TemplateVersion string
