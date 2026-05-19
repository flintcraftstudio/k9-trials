package scoring

import "fmt"

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
	// Group inputs by exercise code for O(1) per-exercise lookup later.
	csByExercise := make(map[string][]CriterionScore)
	for _, cs := range inputs.CriterionScores {
		csByExercise[cs.ExerciseCode] = append(csByExercise[cs.ExerciseCode], cs)
	}
	poByExercise := make(map[string][]PenaltyOccurrence)
	for _, po := range inputs.PenaltyOccurrences {
		poByExercise[po.ExerciseCode] = append(poByExercise[po.ExerciseCode], po)
	}
	atByExercise := make(map[string][]AutoTriggerFiring)
	for _, at := range inputs.AutoTriggers {
		atByExercise[at.ExerciseCode] = append(atByExercise[at.ExerciseCode], at)
	}

	// Step 1: strict validation that every input resolves to something
	// in the concrete sheet / template.
	for code := range csByExercise {
		if _, ok := sheet.FindExercise(code); !ok {
			return ScoresheetResult{}, fmt.Errorf("criterion score references unknown exercise %q", code)
		}
	}
	for code := range poByExercise {
		if _, ok := sheet.FindExercise(code); !ok {
			return ScoresheetResult{}, fmt.Errorf("penalty occurrence references unknown exercise %q", code)
		}
	}
	for code := range atByExercise {
		if _, ok := sheet.FindExercise(code); !ok {
			return ScoresheetResult{}, fmt.Errorf("auto trigger firing references unknown exercise %q", code)
		}
	}
	for _, mod := range inputs.Modifiers {
		if _, ok := template.FindModifier(mod.ModifierCode); !ok {
			return ScoresheetResult{}, fmt.Errorf("modifier application references unknown modifier %q", mod.ModifierCode)
		}
	}

	// Step 2: cascade inflows. Walk every PenaltyLedger exercise, replay
	// its occurrences against its running balance, route OverflowTo
	// excess into the inflow map keyed by target exercise code.
	inflows := computeCascadeInflows(sheet, poByExercise)

	// Steps 3 & 4: evaluate non-aggregate first, then aggregates (which
	// depend on sibling results).
	resultByCode := make(map[string]ExerciseResult)
	phaseNQ := make(map[string]bool)
	trialNQ := false

	for _, ph := range sheet.Phases {
		for _, ex := range ph.Exercises {
			if ex.Kind == Aggregate {
				continue
			}
			res, err := EvaluateExercise(ex, ExerciseInputs{
				CriterionScores:    csByExercise[ex.Code],
				PenaltyOccurrences: poByExercise[ex.Code],
				AutoTriggers:       atByExercise[ex.Code],
				CascadeInflow:      inflows[ex.Code],
			})
			if err != nil {
				return ScoresheetResult{}, fmt.Errorf("phase %s: %w", ph.Code, err)
			}
			resultByCode[ex.Code] = res
			recordAutoNQScopes(ex, res.AutoNQFired, ph.Code, phaseNQ, &trialNQ)
		}
	}

	for _, ph := range sheet.Phases {
		for _, ex := range ph.Exercises {
			if ex.Kind != Aggregate {
				continue
			}
			res, err := EvaluateExercise(ex, ExerciseInputs{
				AutoTriggers:   atByExercise[ex.Code],
				SiblingResults: resultByCode,
			})
			if err != nil {
				return ScoresheetResult{}, fmt.Errorf("phase %s: %w", ph.Code, err)
			}
			resultByCode[ex.Code] = res
			recordAutoNQScopes(ex, res.AutoNQFired, ph.Code, phaseNQ, &trialNQ)
		}
	}

	// Step 5: AutoNQPhase post-pass. Zero every exercise in any phase
	// that had a phase-scoped trigger fire; mark Insufficient (skipping
	// aggregate components per the IsInsufficient-evaluable rule).
	for _, ph := range sheet.Phases {
		if !phaseNQ[ph.Code] {
			continue
		}
		for _, ex := range ph.Exercises {
			r := resultByCode[ex.Code]
			r.Points = 0
			r.Tier = TierInsufficient
			r.IsInsufficient = !ex.IsAggregateComponent
			resultByCode[ex.Code] = r
		}
	}

	// Walk the concrete sheet in order to produce PerExercise.
	perExercise := make([]ExerciseResult, 0, totalExercises(sheet))
	for _, ph := range sheet.Phases {
		for _, ex := range ph.Exercises {
			perExercise = append(perExercise, resultByCode[ex.Code])
		}
	}

	// Step 6: scoresheet totals. Skip aggregate components — their
	// points are already counted via their parent Aggregate's total.
	var total Points
	var insufficientCount int
	for _, ph := range sheet.Phases {
		for _, ex := range ph.Exercises {
			r := resultByCode[ex.Code]
			if ex.IsAggregateComponent {
				continue
			}
			total += r.Points
			if r.IsInsufficient {
				insufficientCount++
			}
		}
	}

	// Step 7: apply modifiers in input order. PointDelta is signed
	// (typically negative); MaxTier is an optional ceiling on FinalTier.
	var appliedMods []ModifierApplicationResult
	var tierCap *Tier
	for _, app := range inputs.Modifiers {
		m, _ := template.FindModifier(app.ModifierCode)
		total += m.PointDelta
		if total < 0 {
			total = 0
		}
		appliedMods = append(appliedMods, ModifierApplicationResult{
			Code:       m.Code,
			PointDelta: m.PointDelta,
			TierCap:    m.MaxTier,
		})
		if m.MaxTier != nil && (tierCap == nil || *m.MaxTier < *tierCap) {
			tierCap = m.MaxTier
		}
	}

	// Step 8: Percent, FinalTier.
	var percent float64
	if sheet.MaxPoints > 0 {
		percent = float64(total) * 100.0 / float64(sheet.MaxPoints)
	}
	finalTier := BandFor(total, sheet.MaxPoints)
	if tierCap != nil && finalTier > *tierCap {
		finalTier = *tierCap
	}

	// Step 9: pass conditions per §3.3.
	passed := !trialNQ &&
		percent >= float64(sheet.PassThresholdPct) &&
		insufficientCount <= sheet.MaxInsufficients

	return ScoresheetResult{
		TotalPoints:       total,
		MaxPoints:         sheet.MaxPoints,
		Percent:           percent,
		InsufficientCount: insufficientCount,
		TrialNQ:           trialNQ,
		Passed:            passed,
		PerExercise:       perExercise,
		AppliedModifiers:  appliedMods,
		FinalTier:         finalTier,
	}, nil
}

// computeCascadeInflows walks every PenaltyLedger exercise in sheet and
// projects logged occurrences into per-target inflow totals. The walk
// here mirrors the running-balance logic in EvaluateExercise — kept
// duplicated so EvaluateExercise stays a single-pass function with no
// hidden side effects.
func computeCascadeInflows(
	sheet ConcreteScoresheet,
	occurrencesByExercise map[string][]PenaltyOccurrence,
) map[string]Points {
	inflows := map[string]Points{}
	for _, ph := range sheet.Phases {
		for _, ex := range ph.Exercises {
			if ex.Kind != PenaltyLedger {
				continue
			}
			occs := occurrencesByExercise[ex.Code]
			if len(occs) == 0 {
				continue
			}
			byCode := make(map[string]PenaltyEvent, len(ex.Events))
			for _, ev := range ex.Events {
				byCode[ev.Code] = ev
			}
			running := ex.MaxPoints
			for _, occ := range occs {
				ev, ok := byCode[occ.EventCode]
				if !ok {
					// Validated above by EvaluateExercise's strict path;
					// skip silently here to avoid double-erroring.
					continue
				}
				if running >= ev.Deduction {
					running -= ev.Deduction
				} else if ev.OverflowTo != "" {
					inflows[ev.OverflowTo] += ev.Deduction - running
					running = 0
				} else {
					running = 0
				}
			}
		}
	}
	return inflows
}

// recordAutoNQScopes inspects the codes that fired against an exercise
// and propagates their phase / trial scope flags upward.
func recordAutoNQScopes(ex ExerciseTemplate, firedCodes []string, phaseCode string, phaseNQ map[string]bool, trialNQ *bool) {
	if len(firedCodes) == 0 {
		return
	}
	byCode := make(map[string]AutoTrigger, len(ex.AutoTriggers))
	for _, at := range ex.AutoTriggers {
		byCode[at.Code] = at
	}
	for _, code := range firedCodes {
		at, ok := byCode[code]
		if !ok {
			continue
		}
		switch at.Scope {
		case AutoNQPhase:
			phaseNQ[phaseCode] = true
		case AutoNQTrial:
			*trialNQ = true
		}
	}
}

func totalExercises(sheet ConcreteScoresheet) int {
	n := 0
	for _, ph := range sheet.Phases {
		n += len(ph.Exercises)
	}
	return n
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
//
// Input ranges are NOT validated. A CriterionScore with Points outside
// [0, criterion.MaxPoints] is summed as-given; a Points value larger
// than the criterion's max can push the exercise above its MaxPoints
// and a negative value can sink it below 0. The submission boundary
// (handler / judge UI) is responsible for clamping or rejecting inputs
// before they reach this function.
func EvaluateExercise(tpl ExerciseTemplate, in ExerciseInputs) (ExerciseResult, error) {
	res := ExerciseResult{
		ExerciseCode: tpl.Code,
		Kind:         tpl.Kind,
		MaxPoints:    tpl.MaxPoints,
	}

	knownTriggers := make(map[string]AutoTrigger, len(tpl.AutoTriggers))
	for _, at := range tpl.AutoTriggers {
		knownTriggers[at.Code] = at
	}
	autoNQExerciseFired := false
	for _, fired := range in.AutoTriggers {
		at, ok := knownTriggers[fired.TriggerCode]
		if !ok {
			return res, fmt.Errorf("exercise %s: unknown AutoTrigger %q", tpl.Code, fired.TriggerCode)
		}
		res.AutoNQFired = append(res.AutoNQFired, fired.TriggerCode)
		if at.Scope == AutoNQExercise {
			autoNQExerciseFired = true
		}
	}

	autoInsuf := false
	switch tpl.Kind {
	case CriteriaSum:
		critCodes := make(map[string]bool, len(tpl.Criteria))
		for _, c := range tpl.Criteria {
			critCodes[c.Code] = true
		}
		var sum Points
		for _, cs := range in.CriterionScores {
			if !critCodes[cs.CriterionCode] {
				return res, fmt.Errorf("exercise %s: unknown criterion %q", tpl.Code, cs.CriterionCode)
			}
			sum += cs.Points
		}
		res.Points = sum

	case PenaltyLedger:
		knownEvents := make(map[string]PenaltyEvent, len(tpl.Events))
		for _, ev := range tpl.Events {
			knownEvents[ev.Code] = ev
		}
		running := tpl.MaxPoints
		for _, occ := range in.PenaltyOccurrences {
			ev, ok := knownEvents[occ.EventCode]
			if !ok {
				return res, fmt.Errorf("exercise %s: unknown penalty event %q", tpl.Code, occ.EventCode)
			}
			if ev.AutoInsufficientOnOccurrence {
				autoInsuf = true
			}
			if running >= ev.Deduction {
				running -= ev.Deduction
			} else if ev.OverflowTo != "" {
				res.CascadeOverflow += ev.Deduction - running
				running = 0
			} else {
				running = 0
			}
		}
		if in.CascadeInflow > 0 {
			absorbed := in.CascadeInflow
			if absorbed > running {
				absorbed = running
			}
			running -= absorbed
			res.AbsorbedOverflow = absorbed
		}
		res.Points = running

	case Aggregate:
		var sum Points
		for _, code := range tpl.AggregateOf {
			sib, ok := in.SiblingResults[code]
			if !ok {
				return res, fmt.Errorf("exercise %s: missing sibling result %q for Aggregate", tpl.Code, code)
			}
			sum += sib.Points
		}
		res.Points = sum

	default:
		return res, fmt.Errorf("exercise %s: unknown Kind %d", tpl.Code, tpl.Kind)
	}

	switch {
	case autoNQExerciseFired:
		res.Points = 0
		res.Tier = TierInsufficient
	case autoInsuf:
		res.Tier = TierInsufficient
	default:
		res.Tier = BandFor(res.Points, tpl.MaxPoints)
	}
	res.IsInsufficient = res.Tier == TierInsufficient && !tpl.IsAggregateComponent

	return res, nil
}
