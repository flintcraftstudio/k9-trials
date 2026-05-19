package scoring

import "fmt"

// ExerciseKind discriminates how an exercise is scored.
// Used as the tag on the discriminated union in ExerciseTemplate.
//
// Within ExerciseTemplate, exactly one of {Criteria, Events, AggregateOf}
// is populated according to Kind. The Validate method enforces this.
type ExerciseKind int

const (
	// CriteriaSum: exercise score = sum of criterion scores.
	// Used by virtually all L1 OB exercises, most Protection scenario
	// exercises, and any exercise where the judge awards points up
	// from 0 against a list of positive line items.
	CriteriaSum ExerciseKind = iota

	// PenaltyLedger: exercise starts at MaxPoints and is reduced by
	// logged penalty events. The judge logs occurrences (e.g.
	// "missed hide", "false alert", "exceeded time"); the system
	// computes the score. Used by:
	//   - L1/L2/L3 DET Handler Accuracy
	//   - L1/L2/L3 TRK Completion (binary — Deduction == MaxPoints)
	//   - other event-driven exercises
	PenaltyLedger

	// Aggregate: exercise has no direct scores. Its score is the sum
	// of its component sibling exercises' scores. Used by L2/L3 DET
	// aggregated category lines (Hunt Drive, Odor Commitment, etc.,
	// where per-search 5-point scores roll up into a category total).
	Aggregate
)

// Criterion is a single positive line item within a CriteriaSum exercise.
type Criterion struct {
	Code        string // stable identifier, e.g., "1.1.a"
	Description string // displayed to the judge on the tablet
	MaxPoints   Points // sums with siblings to ExerciseTemplate.MaxPoints
}

// PenaltyEvent is a discrete fault the judge can log against a
// PenaltyLedger exercise. Each logged occurrence subtracts Deduction
// from the exercise's running total. The total floors at 0.
type PenaltyEvent struct {
	Code        string // "missed-hide", "false-alert-active", "exceeded-time"
	Description string // judge-facing label
	Deduction   Points // amount subtracted per occurrence

	// OverflowTo, when non-empty, names the sibling exercise code that
	// absorbs deduction excess once this exercise hits its floor (0).
	// Used by L2/L3 DET false-alert-in-blank: the -20 hits Handler
	// Accuracy first, excess flows to Blank Area Performance.
	// Empty for the common case.
	OverflowTo string

	// AutoInsufficientOnOccurrence, when true, marks the exercise
	// Insufficient as soon as a single occurrence is logged regardless
	// of remaining points. For binary completion exercises (e.g. TRK
	// Completion) the deduction equal to MaxPoints already produces
	// Insufficient via the math; this flag is for cases where the
	// rulebook says "Insufficient" but the math wouldn't otherwise
	// reach that tier.
	AutoInsufficientOnOccurrence bool
}

// AutoTrigger is a non-scoring rule that, when fired by the judge,
// forces an outcome regardless of point totals. Examples:
//   - "Dog destroys muzzle" (zeroes the exercise)
//   - "Aggression toward people" (§4.5 trial-level NQ)
type AutoTrigger struct {
	Code        string // e.g., "1.1.nq.destroy-muzzle"
	Description string
	Scope       AutoTriggerScope
}

// AutoTriggerScope determines how far an AutoTrigger's effect reaches.
type AutoTriggerScope int

const (
	// AutoNQExercise zeros this exercise only and marks it Insufficient.
	// Other exercises and the rest of the trial continue.
	AutoNQExercise AutoTriggerScope = iota
	// AutoNQPhase zeros the entire phase containing this exercise and
	// marks every exercise in the phase Insufficient.
	AutoNQPhase
	// AutoNQTrial ends the trial. The team is NQ regardless of any
	// other scoring (§4.5).
	AutoNQTrial
)

// ExerciseTemplate is the rulebook spec for a single scored exercise.
// Discriminated union: the Kind field determines which of
// Criteria / Events / AggregateOf is populated.
type ExerciseTemplate struct {
	Kind      ExerciseKind
	Code      string // "1.1"
	Name      string // "Muzzle Acceptance & Heeling Pattern"
	MaxPoints Points

	// Populated when Kind == CriteriaSum. Sum of Criteria.MaxPoints
	// must equal MaxPoints.
	Criteria []Criterion

	// Populated when Kind == PenaltyLedger. The exercise's running
	// score starts at MaxPoints and is reduced by logged events.
	Events []PenaltyEvent

	// Populated when Kind == Aggregate. Lists sibling exercise codes
	// (within the same PhaseTemplate) whose scores sum into this one.
	// MaxPoints must equal the sum of those siblings' MaxPoints.
	AggregateOf []string

	// IsAggregateComponent, when true, means this exercise's points
	// roll up into a sibling Aggregate exercise. The exercise is
	// still scored normally (criteria or events), but its tier is
	// NOT individually evaluated for the scoresheet's Insufficient
	// count — only the Aggregate parent's tier is.
	// Must be false when Kind == Aggregate.
	IsAggregateComponent bool

	AutoTriggers []AutoTrigger
}

// Validate checks that the discriminated union is well-formed and that
// cross-references resolve. Called at process startup against every
// hardcoded template.
//
// The siblings argument is the list of all exercises within the same
// PhaseTemplate, used to resolve AggregateOf and PenaltyEvent.OverflowTo
// references.
//
// Validation rules per Kind:
//
//	CriteriaSum:   Criteria non-empty and sums to MaxPoints;
//	               Events nil; AggregateOf nil.
//	PenaltyLedger: Events non-empty; Criteria nil; AggregateOf nil.
//	               Each Event.OverflowTo (if set) must resolve to a
//	               sibling exercise code.
//	Aggregate:     AggregateOf non-empty and each entry resolves to a
//	               sibling marked IsAggregateComponent;
//	               sum of components' MaxPoints equals this MaxPoints;
//	               Criteria nil; Events nil; IsAggregateComponent false.
func (t ExerciseTemplate) Validate(siblings []ExerciseTemplate) error {
	if t.Code == "" {
		return fmt.Errorf("exercise has no Code")
	}
	switch t.Kind {
	case CriteriaSum:
		if len(t.Criteria) == 0 {
			return fmt.Errorf("exercise %s: CriteriaSum requires criteria", t.Code)
		}
		if len(t.Events) != 0 {
			return fmt.Errorf("exercise %s: CriteriaSum must not have events", t.Code)
		}
		if len(t.AggregateOf) != 0 {
			return fmt.Errorf("exercise %s: CriteriaSum must not have AggregateOf", t.Code)
		}
		var sum Points
		for _, c := range t.Criteria {
			sum += c.MaxPoints
		}
		if sum != t.MaxPoints {
			return fmt.Errorf("exercise %s: criteria sum %d != MaxPoints %d", t.Code, sum, t.MaxPoints)
		}
	case PenaltyLedger:
		if len(t.Events) == 0 {
			return fmt.Errorf("exercise %s: PenaltyLedger requires events", t.Code)
		}
		if len(t.Criteria) != 0 {
			return fmt.Errorf("exercise %s: PenaltyLedger must not have criteria", t.Code)
		}
		if len(t.AggregateOf) != 0 {
			return fmt.Errorf("exercise %s: PenaltyLedger must not have AggregateOf", t.Code)
		}
		siblingCodes := make(map[string]bool, len(siblings))
		for _, s := range siblings {
			siblingCodes[s.Code] = true
		}
		for _, ev := range t.Events {
			if ev.OverflowTo != "" && !siblingCodes[ev.OverflowTo] {
				return fmt.Errorf("exercise %s event %s: OverflowTo %q does not resolve to a sibling",
					t.Code, ev.Code, ev.OverflowTo)
			}
		}
	case Aggregate:
		if t.IsAggregateComponent {
			return fmt.Errorf("exercise %s: Aggregate cannot also be IsAggregateComponent", t.Code)
		}
		if len(t.AggregateOf) == 0 {
			return fmt.Errorf("exercise %s: Aggregate requires AggregateOf entries", t.Code)
		}
		if len(t.Criteria) != 0 {
			return fmt.Errorf("exercise %s: Aggregate must not have criteria", t.Code)
		}
		if len(t.Events) != 0 {
			return fmt.Errorf("exercise %s: Aggregate must not have events", t.Code)
		}
		siblingByCode := make(map[string]ExerciseTemplate, len(siblings))
		for _, s := range siblings {
			siblingByCode[s.Code] = s
		}
		var compSum Points
		for _, code := range t.AggregateOf {
			sib, ok := siblingByCode[code]
			if !ok {
				return fmt.Errorf("exercise %s: AggregateOf %q does not resolve to a sibling", t.Code, code)
			}
			if !sib.IsAggregateComponent {
				return fmt.Errorf("exercise %s: AggregateOf %q is not marked IsAggregateComponent", t.Code, code)
			}
			compSum += sib.MaxPoints
		}
		if compSum != t.MaxPoints {
			return fmt.Errorf("exercise %s: component sum %d != MaxPoints %d", t.Code, compSum, t.MaxPoints)
		}
	default:
		return fmt.Errorf("exercise %s: unknown Kind %d", t.Code, t.Kind)
	}
	return nil
}

// PhaseTemplate is a named grouping of exercises within a scoresheet.
// Covers both OB Phases and Protection Scenarios — only the Name
// distinguishes them in display. Tracking uses "Track" and "Safety
// Preparedness" phases; Detection uses per-search phases plus a
// Blank Area phase.
type PhaseTemplate struct {
	Code      string             // "P1", "S1", "TRACK", "SAFETY"
	Name      string             // "Phase 1: Muzzle & Stability"
	Exercises []ExerciseTemplate
}

// MaxPoints returns the sum of contributing exercises' MaxPoints.
// Components do not double-count: their points are subsumed by the
// Aggregate parent's MaxPoints, so only non-component exercises and
// Aggregate parents contribute to the phase total.
func (p PhaseTemplate) MaxPoints() Points {
	var total Points
	for _, ex := range p.Exercises {
		if ex.IsAggregateComponent {
			continue
		}
		total += ex.MaxPoints
	}
	return total
}

// SelectionMode governs how a judge constructs a ConcretePhase from a
// PhaseTemplate's menu of exercises at trial-setup time.
type SelectionMode int

const (
	// SelectAll: every exercise in the phase is run. Default for
	// L1 OB, all L1/L2/L3 TRK, all Protection scenarios at L1/L2.
	SelectAll SelectionMode = iota

	// SelectFromInventory: judge selects Min-to-Max exercises from
	// the phase's exercises at trial setup. Used by L2 OB Phase 3
	// (3-4 obstacles from inventory) and every phase of L3 EN
	// (2-3 behaviors per domain).
	SelectFromInventory
)

// PhaseSelection describes how a single phase's concrete exercise
// list is built from its template.
type PhaseSelection struct {
	Mode SelectionMode
	// Min, Max: how many exercises the judge selects from the phase's
	// menu. Only consulted when Mode == SelectFromInventory.
	Min, Max int
}

// SelectionRule governs how a ConcreteScoresheet is built from a
// ScoresheetTemplate at trial-setup time.
//
// Phases absent from PerPhase use SelectAll. For the common case where
// every phase is SelectAll (L1 OB, L1/L2/L3 TRK, L1/L2 Protection),
// the SelectionRule may be the zero value.
type SelectionRule struct {
	PerPhase map[string]PhaseSelection // keyed by PhaseTemplate.Code
}

// ScoresheetModifier is a post-aggregation effect on a scoresheet.
// Modifiers apply after EvaluateScoresheet produces the raw result,
// and may deduct points from the total and/or cap the awarded tier.
//
// V1 has one modifier: L2 TRK Lifeline (-20 points, tier cap at
// TierVeryGood — cannot earn Excellent).
//
// Modifiers are attached to ScoresheetTemplate.AvailableModifiers; a
// judge may only apply modifiers from that list during evaluation.
type ScoresheetModifier struct {
	Code        string // "L2-TRK-LIFELINE"
	Description string
	PointDelta  Points // negative for deductions; 0 if tier-cap-only
	MaxTier     *Tier  // optional ceiling on overall awarded tier
}

// ScoresheetTemplate is the rulebook spec for a (Discipline, Level)
// pair at a given rulebook version. It is a MENU — the full set of
// exercises that may be scored at this level. Concrete trials select
// from the menu via BuildConcrete.
type ScoresheetTemplate struct {
	Version    TemplateVersion
	Discipline Discipline
	Level      Level
	Phases     []PhaseTemplate

	// SelectionRule describes how a ConcreteScoresheet is built from
	// this template. Zero value means SelectAll for every phase.
	SelectionRule SelectionRule

	// AvailableModifiers lists modifiers a judge may apply when
	// evaluating a scoresheet against this template. Nil/empty means
	// no modifiers exist for this (Discipline, Level).
	AvailableModifiers []ScoresheetModifier

	// PassThresholdPct: §2.2. 70 (L1/L2), 75 (L3).
	PassThresholdPct int

	// MaxInsufficients: §2.3. 1 (L1), 0 (L2/L3).
	MaxInsufficients int
}

// Validate checks the entire template tree: every phase, every
// exercise, every cross-reference. Run at process startup before
// serving traffic. Returns the first error encountered.
func (t ScoresheetTemplate) Validate() error {
	if t.Version == "" {
		return fmt.Errorf("template has no Version")
	}
	if !t.Level.IsValid() {
		return fmt.Errorf("template %s: invalid level %d", t.Version, t.Level)
	}
	if len(t.Phases) == 0 {
		return fmt.Errorf("template %s: no phases", t.Version)
	}
	seenExercise := make(map[string]bool)
	seenPhase := make(map[string]bool)
	for _, ph := range t.Phases {
		if ph.Code == "" {
			return fmt.Errorf("phase has no Code")
		}
		if seenPhase[ph.Code] {
			return fmt.Errorf("duplicate phase code %q", ph.Code)
		}
		seenPhase[ph.Code] = true
		for _, ex := range ph.Exercises {
			if seenExercise[ex.Code] {
				return fmt.Errorf("duplicate exercise code %q across phases", ex.Code)
			}
			seenExercise[ex.Code] = true
			if err := ex.Validate(ph.Exercises); err != nil {
				return err
			}
		}
	}
	// AvailableModifiers: codes must be unique.
	seenMod := make(map[string]bool)
	for _, m := range t.AvailableModifiers {
		if m.Code == "" {
			return fmt.Errorf("modifier has no Code")
		}
		if seenMod[m.Code] {
			return fmt.Errorf("duplicate modifier code %q", m.Code)
		}
		seenMod[m.Code] = true
	}
	return nil
}

// FindExercise returns the named exercise within this template, or
// false if not present.
func (t ScoresheetTemplate) FindExercise(code string) (ExerciseTemplate, bool) {
	for _, ph := range t.Phases {
		for _, ex := range ph.Exercises {
			if ex.Code == code {
				return ex, true
			}
		}
	}
	return ExerciseTemplate{}, false
}

// FindCriterion returns the named criterion within the named exercise,
// or false if not present.
func (t ScoresheetTemplate) FindCriterion(exerciseCode, criterionCode string) (Criterion, bool) {
	ex, ok := t.FindExercise(exerciseCode)
	if !ok {
		return Criterion{}, false
	}
	for _, c := range ex.Criteria {
		if c.Code == criterionCode {
			return c, true
		}
	}
	return Criterion{}, false
}

// FindModifier returns the named modifier from AvailableModifiers, or
// false if not present.
func (t ScoresheetTemplate) FindModifier(code string) (ScoresheetModifier, bool) {
	for _, m := range t.AvailableModifiers {
		if m.Code == code {
			return m, true
		}
	}
	return ScoresheetModifier{}, false
}

// BuildConcrete applies SelectionRule and produces a ConcreteScoresheet
// from this template plus the judge's selections.
//
// For SelectAll phases (the default), passes through all exercises;
// the corresponding entry in selections is ignored.
//
// For SelectFromInventory phases, takes the selected exercise codes
// from selections[phaseCode]. Errors when:
//   - the count of selections falls outside [Min, Max]
//   - a selected code does not resolve to an exercise in that phase
//
// L3 EN's complementary-coverage rule (across two legs of a title day)
// is NOT enforced here. BuildConcrete validates one scoresheet at a
// time; cross-leg coverage is a title-evaluation concern.
func (t ScoresheetTemplate) BuildConcrete(
	selections map[string][]string, // phaseCode -> selected exerciseCodes
) (ConcreteScoresheet, error) {
	out := ConcreteScoresheet{
		TemplateVersion:  t.Version,
		Discipline:       t.Discipline,
		Level:            t.Level,
		PassThresholdPct: t.PassThresholdPct,
		MaxInsufficients: t.MaxInsufficients,
		Phases:           make([]ConcretePhase, 0, len(t.Phases)),
	}

	for _, ph := range t.Phases {
		sel, hasRule := t.SelectionRule.PerPhase[ph.Code]
		// Default Mode is SelectAll (zero value); applies when no rule
		// or when the phase rule's Mode is SelectAll.
		if !hasRule || sel.Mode == SelectAll {
			out.Phases = append(out.Phases, ConcretePhase{
				Code:      ph.Code,
				Name:      ph.Name,
				Exercises: append([]ExerciseTemplate(nil), ph.Exercises...),
			})
			out.MaxPoints += ph.MaxPoints()
			continue
		}

		// SelectFromInventory: resolve judge-selected codes.
		picked := selections[ph.Code]
		if len(picked) < sel.Min || len(picked) > sel.Max {
			return ConcreteScoresheet{}, fmt.Errorf(
				"phase %s: selected %d exercises, must be in [%d, %d]",
				ph.Code, len(picked), sel.Min, sel.Max)
		}

		byCode := make(map[string]ExerciseTemplate, len(ph.Exercises))
		for _, ex := range ph.Exercises {
			byCode[ex.Code] = ex
		}

		cp := ConcretePhase{Code: ph.Code, Name: ph.Name}
		seen := make(map[string]bool, len(picked))
		for _, code := range picked {
			if seen[code] {
				return ConcreteScoresheet{}, fmt.Errorf(
					"phase %s: duplicate selection %q", ph.Code, code)
			}
			seen[code] = true
			ex, ok := byCode[code]
			if !ok {
				return ConcreteScoresheet{}, fmt.Errorf(
					"phase %s: selected exercise %q not in phase", ph.Code, code)
			}
			cp.Exercises = append(cp.Exercises, ex)
		}
		out.Phases = append(out.Phases, cp)

		// MaxPoints for the concrete phase: same component-exclusion
		// rule as PhaseTemplate.MaxPoints.
		var phMax Points
		for _, ex := range cp.Exercises {
			if ex.IsAggregateComponent {
				continue
			}
			phMax += ex.MaxPoints
		}
		out.MaxPoints += phMax
	}

	return out, nil
}

// ConcreteScoresheet is the scoresheet shape that gets scored. Built
// from a ScoresheetTemplate at trial-setup time. For most disciplines
// (L1 OB, all TRK, L1/L2 Protection), ConcreteScoresheet is the
// template's phases passed through verbatim — the abstraction is
// near-zero-cost where it isn't needed.
type ConcreteScoresheet struct {
	TemplateVersion TemplateVersion
	Discipline      Discipline
	Level           Level

	// Phases mirror the template's phases, filtered to the judge's
	// selection. Phase codes preserved.
	Phases []ConcretePhase

	// MaxPoints is the sum across phases. Stored here because L3 EN's
	// max varies per leg by judge selection.
	MaxPoints Points

	PassThresholdPct int
	MaxInsufficients int
}

// ConcretePhase is the selected subset of a PhaseTemplate.
type ConcretePhase struct {
	Code      string
	Name      string
	Exercises []ExerciseTemplate // the judge-selected subset
}

// FindExercise returns the named exercise within this concrete
// scoresheet, or false if not present.
func (s ConcreteScoresheet) FindExercise(code string) (ExerciseTemplate, bool) {
	for _, ph := range s.Phases {
		for _, ex := range ph.Exercises {
			if ex.Code == code {
				return ex, true
			}
		}
	}
	return ExerciseTemplate{}, false
}
