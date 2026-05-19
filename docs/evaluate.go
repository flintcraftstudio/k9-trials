package scoring

// CriterionScore is a judge's point assignment to one criterion within
// a CriteriaSum exercise. Append-only at the storage layer; later
// writes supersede earlier ones for the same
// (scoresheet, exercise, criterion) tuple.
type CriterionScore struct {
	ExerciseCode  string
	CriterionCode string
	Points        Points
}

// PenaltyOccurrence is a logged instance of a PenaltyEvent within a
// PenaltyLedger exercise. Each occurrence is a separate record (no
// batching). The append-only storage layer preserves every tap; the
// evaluation engine counts occurrences by (ExerciseCode, EventCode).
type PenaltyOccurrence struct {
	ExerciseCode string
	EventCode    string
}

// AutoTriggerFiring records that a judge invoked an AutoTrigger.
type AutoTriggerFiring struct {
	ExerciseCode string
	TriggerCode  string
}

// ModifierApplication records that a ScoresheetModifier was applied
// to a scoresheet. The modifier definition is looked up from
// ScoresheetTemplate.AvailableModifiers.
type ModifierApplication struct {
	ModifierCode string
}

// ScoresheetInputs bundles everything the judge has logged for a
// scoresheet at evaluation time. Each slice is the latest-write-wins
// projection of the underlying append-only storage rows.
type ScoresheetInputs struct {
	CriterionScores    []CriterionScore
	PenaltyOccurrences []PenaltyOccurrence
	AutoTriggers       []AutoTriggerFiring
	Modifiers          []ModifierApplication
}

// ExerciseInputs is the per-exercise slice of ScoresheetInputs needed
// by EvaluateExercise. EvaluateScoresheet builds these internally;
// callers who use EvaluateExercise directly (for testing or partial
// evaluation) construct them by filtering ScoresheetInputs to the
// relevant exercise.
type ExerciseInputs struct {
	// CriterionScores: scores logged against this exercise's criteria.
	// Used when ExerciseTemplate.Kind == CriteriaSum.
	CriterionScores []CriterionScore

	// PenaltyOccurrences: events logged against this exercise.
	// Used when ExerciseTemplate.Kind == PenaltyLedger.
	PenaltyOccurrences []PenaltyOccurrence

	// SiblingResults: results of already-evaluated sibling exercises,
	// keyed by exercise code. Used when ExerciseTemplate.Kind ==
	// Aggregate to look up the components named in AggregateOf.
	SiblingResults map[string]ExerciseResult

	// CascadeInflow: total points absorbed from sibling exercises
	// whose PenaltyEvent.OverflowTo named this exercise. Subtracted
	// from this exercise's score after its own penalties resolve.
	// Used only on PenaltyLedger exercises that are overflow targets.
	CascadeInflow Points

	// AutoTriggers: AutoTriggerFirings recorded against this exercise.
	AutoTriggers []AutoTriggerFiring
}

// ExerciseResult is the evaluation of one exercise.
type ExerciseResult struct {
	ExerciseCode string
	Kind         ExerciseKind
	Points       Points
	MaxPoints    Points
	Tier         Tier

	// IsInsufficient: this exercise's tier == TierInsufficient AND it
	// is Insufficient-evaluable. An exercise is Insufficient-evaluable
	// when it is NOT IsAggregateComponent (components don't count;
	// their parent Aggregate does).
	IsInsufficient bool

	// AutoNQFired lists the AutoTrigger codes that fired against this
	// exercise. The presence of any AutoNQTrial-scoped firing here is
	// surfaced separately on ScoresheetResult.TrialNQ.
	AutoNQFired []string

	// CascadeOverflow: deduction excess that overflowed FROM this
	// exercise (after it hit 0) to its sibling overflow target.
	// Diagnostic; the math is already reflected in Points.
	CascadeOverflow Points

	// AbsorbedOverflow: deduction excess this exercise absorbed FROM
	// a sibling (CascadeInflow). Diagnostic; already reflected in Points.
	AbsorbedOverflow Points
}

// ScoresheetResult is the full evaluation of a scoresheet. Pure
// function of (inputs, concrete, template); no I/O, no time.
type ScoresheetResult struct {
	TotalPoints Points
	MaxPoints   Points

	// Percent: 100 * TotalPoints / MaxPoints. Computed once and stored;
	// callers compare against PassThresholdPct.
	Percent float64

	// InsufficientCount: number of Insufficient-evaluable exercises
	// whose tier == TierInsufficient.
	InsufficientCount int

	// TrialNQ: any AutoTrigger with scope AutoNQTrial fired.
	TrialNQ bool

	// Passed satisfies §3.3:
	//   Percent >= template.PassThresholdPct
	//   AND InsufficientCount <= template.MaxInsufficients
	//   AND !TrialNQ
	//   AND modifier tier caps respected
	Passed bool

	// PerExercise: results for every exercise in the concrete scoresheet,
	// in concrete-scoresheet order.
	PerExercise []ExerciseResult

	// AppliedModifiers: modifiers applied, in application order,
	// with their effects materialized.
	AppliedModifiers []ModifierApplicationResult

	// FinalTier: the overall tier awarded after modifier caps.
	// Derived from Percent against the standard tier bands, then
	// capped by any modifier.MaxTier.
	FinalTier Tier
}

// ModifierApplicationResult records the effect of one applied modifier.
type ModifierApplicationResult struct {
	Code       string
	PointDelta Points
	TierCap    *Tier // nil if no cap applied
}

// EvaluateScoresheet computes the full scoresheet result. Pure
// function. Evaluation order:
//
//  1. Filter inputs to exercises present in the concrete scoresheet.
//  2. Compute cascade inflows by walking PenaltyEvent.OverflowTo
//     references across all PenaltyLedger exercises.
//  3. Evaluate non-Aggregate exercises (CriteriaSum, PenaltyLedger).
//  4. Evaluate Aggregate exercises, reading sibling results.
//  5. Apply AutoNQPhase scope: zero every exercise in the affected
//     phase and mark Insufficient.
//  6. Compute scoresheet totals (sum non-component, non-double-counted).
//  7. Apply modifiers in input order: deduct points, record tier caps.
//  8. Compute Percent, InsufficientCount, FinalTier, TrialNQ.
//  9. Compute Passed per §3.3.
//
// Errors: returns an error when inputs reference exercises/criteria/
// events/modifiers not present in the concrete scoresheet or template.
// The function is strict — silent-ignore of unknown codes would mask
// tablet-sync bugs.
func EvaluateScoresheet(
	inputs ScoresheetInputs,
	sheet ConcreteScoresheet,
	template ScoresheetTemplate,
) (ScoresheetResult, error) {
	panic("not implemented")
}

// EvaluateExercise computes the result for a single exercise. Exported
// for testing and for callers who need partial evaluation (e.g., the
// tablet showing a live running total as the judge enters criteria).
//
// For Aggregate kinds, ExerciseInputs.SiblingResults must contain entries
// for every code listed in tpl.AggregateOf.
//
// Returns an error when:
//   - CriterionScores reference criteria not on the template
//   - PenaltyOccurrences reference events not on the template
//   - Aggregate is missing a required sibling result
func EvaluateExercise(tpl ExerciseTemplate, in ExerciseInputs) (ExerciseResult, error) {
	panic("not implemented")
}
