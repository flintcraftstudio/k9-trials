package scoring

import (
	"strings"
	"testing"
)

// twoPhaseTemplate is a minimal but realistic template used across the
// scoresheet-level tests below. P1 has one 10-pt CriteriaSum (1.1).
// P2 has one 10-pt CriteriaSum (2.1) plus a phase-NQ-scoped trigger.
// L1-style pass conditions: 70%, 1 insufficient allowed.
func twoPhaseTemplate() (ScoresheetTemplate, ConcreteScoresheet) {
	tpl := ScoresheetTemplate{
		Version:          "test.1",
		Discipline:       DisciplineOB,
		Level:            LevelOne,
		PassThresholdPct: 70,
		MaxInsufficients: 1,
		Phases: []PhaseTemplate{
			{
				Code: "P1",
				Exercises: []ExerciseTemplate{
					{
						Kind: CriteriaSum, Code: "1.1", MaxPoints: 10,
						Criteria: []Criterion{
							{Code: "1.1.a", MaxPoints: 5},
							{Code: "1.1.b", MaxPoints: 5},
						},
					},
				},
			},
			{
				Code: "P2",
				Exercises: []ExerciseTemplate{
					{
						Kind: CriteriaSum, Code: "2.1", MaxPoints: 10,
						Criteria: []Criterion{{Code: "2.1.a", MaxPoints: 10}},
						AutoTriggers: []AutoTrigger{
							{Code: "p2.nq", Scope: AutoNQPhase},
							{Code: "trial.nq", Scope: AutoNQTrial},
						},
					},
				},
			},
		},
	}
	if err := tpl.Validate(); err != nil {
		panic(err)
	}
	sheet, err := tpl.BuildConcrete(nil)
	if err != nil {
		panic(err)
	}
	return tpl, sheet
}

func TestEvaluateScoresheet_PerfectRun_Passes(t *testing.T) {
	tpl, sheet := twoPhaseTemplate()
	inputs := ScoresheetInputs{
		CriterionScores: []CriterionScore{
			{ExerciseCode: "1.1", CriterionCode: "1.1.a", Points: 5},
			{ExerciseCode: "1.1", CriterionCode: "1.1.b", Points: 5},
			{ExerciseCode: "2.1", CriterionCode: "2.1.a", Points: 10},
		},
	}
	res, err := EvaluateScoresheet(inputs, sheet, tpl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.TotalPoints != 20 || res.MaxPoints != 20 {
		t.Errorf("Total/Max = %d/%d, want 20/20", res.TotalPoints, res.MaxPoints)
	}
	if res.Percent != 100.0 {
		t.Errorf("Percent = %f, want 100", res.Percent)
	}
	if res.FinalTier != TierExcellent {
		t.Errorf("FinalTier = %v, want Excellent", res.FinalTier)
	}
	if !res.Passed {
		t.Error("expected pass")
	}
}

func TestEvaluateScoresheet_BelowThreshold_Fails(t *testing.T) {
	tpl, sheet := twoPhaseTemplate()
	inputs := ScoresheetInputs{
		CriterionScores: []CriterionScore{
			{ExerciseCode: "1.1", CriterionCode: "1.1.a", Points: 3},
			{ExerciseCode: "1.1", CriterionCode: "1.1.b", Points: 3},
			{ExerciseCode: "2.1", CriterionCode: "2.1.a", Points: 7},
		},
	}
	// Total: 13/20 = 65% — below the 70% L1 threshold.
	res, _ := EvaluateScoresheet(inputs, sheet, tpl)
	if res.Passed {
		t.Errorf("expected fail (13/20 = 65%%), Percent=%f", res.Percent)
	}
}

func TestEvaluateScoresheet_TwoInsufficient_FailsCount(t *testing.T) {
	tpl, sheet := twoPhaseTemplate()
	// Both exercises at 6/10 = 60% → Insufficient on each.
	// L1 allows max 1 insufficient.
	inputs := ScoresheetInputs{
		CriterionScores: []CriterionScore{
			{ExerciseCode: "1.1", CriterionCode: "1.1.a", Points: 3},
			{ExerciseCode: "1.1", CriterionCode: "1.1.b", Points: 3},
			{ExerciseCode: "2.1", CriterionCode: "2.1.a", Points: 6},
		},
	}
	res, _ := EvaluateScoresheet(inputs, sheet, tpl)
	if res.InsufficientCount != 2 {
		t.Errorf("InsufficientCount = %d, want 2", res.InsufficientCount)
	}
	if res.Passed {
		t.Error("expected fail (2 insufficients > allowed 1)")
	}
}

func TestEvaluateScoresheet_AutoNQTrial_Fails(t *testing.T) {
	tpl, sheet := twoPhaseTemplate()
	inputs := ScoresheetInputs{
		CriterionScores: []CriterionScore{
			{ExerciseCode: "1.1", CriterionCode: "1.1.a", Points: 5},
			{ExerciseCode: "1.1", CriterionCode: "1.1.b", Points: 5},
			{ExerciseCode: "2.1", CriterionCode: "2.1.a", Points: 10},
		},
		AutoTriggers: []AutoTriggerFiring{
			{ExerciseCode: "2.1", TriggerCode: "trial.nq"},
		},
	}
	res, _ := EvaluateScoresheet(inputs, sheet, tpl)
	if !res.TrialNQ {
		t.Error("expected TrialNQ=true")
	}
	if res.Passed {
		t.Error("trial NQ must force fail regardless of percent")
	}
	// Points are NOT zeroed by trial NQ — only the pass flag flips.
	if res.TotalPoints != 20 {
		t.Errorf("TotalPoints = %d, want 20 (trial NQ doesn't zero points)", res.TotalPoints)
	}
}

func TestEvaluateScoresheet_AutoNQPhase_ZerosPhase(t *testing.T) {
	tpl, sheet := twoPhaseTemplate()
	inputs := ScoresheetInputs{
		CriterionScores: []CriterionScore{
			{ExerciseCode: "1.1", CriterionCode: "1.1.a", Points: 5},
			{ExerciseCode: "1.1", CriterionCode: "1.1.b", Points: 5},
			{ExerciseCode: "2.1", CriterionCode: "2.1.a", Points: 10},
		},
		AutoTriggers: []AutoTriggerFiring{
			{ExerciseCode: "2.1", TriggerCode: "p2.nq"},
		},
	}
	res, _ := EvaluateScoresheet(inputs, sheet, tpl)
	// P2 zeroed; only P1's 10 points remain.
	if res.TotalPoints != 10 {
		t.Errorf("TotalPoints = %d, want 10 (P2 zeroed)", res.TotalPoints)
	}
	// 2.1 is now Insufficient; 1.1 stays Excellent.
	if res.InsufficientCount != 1 {
		t.Errorf("InsufficientCount = %d, want 1", res.InsufficientCount)
	}
}

func TestEvaluateScoresheet_UnknownExerciseInInput(t *testing.T) {
	tpl, sheet := twoPhaseTemplate()
	inputs := ScoresheetInputs{
		CriterionScores: []CriterionScore{
			{ExerciseCode: "ghost", CriterionCode: "x.a", Points: 1},
		},
	}
	_, err := EvaluateScoresheet(inputs, sheet, tpl)
	if err == nil || !strings.Contains(err.Error(), "unknown exercise") {
		t.Fatalf("expected unknown-exercise error, got %v", err)
	}
}

func TestEvaluateScoresheet_UnknownModifier(t *testing.T) {
	tpl, sheet := twoPhaseTemplate()
	inputs := ScoresheetInputs{
		Modifiers: []ModifierApplication{{ModifierCode: "MADE-UP"}},
	}
	_, err := EvaluateScoresheet(inputs, sheet, tpl)
	if err == nil || !strings.Contains(err.Error(), "unknown modifier") {
		t.Fatalf("expected unknown-modifier error, got %v", err)
	}
}

// TestEvaluateScoresheet_Modifier_PointDeltaAndTierCap exercises the
// L2 TRK Lifeline scenario: -20 points, tier cap at Very Good.
func TestEvaluateScoresheet_Modifier_PointDeltaAndTierCap(t *testing.T) {
	veryGood := TierVeryGood
	mod := ScoresheetModifier{
		Code: "LIFELINE", PointDelta: -20, MaxTier: &veryGood,
	}
	tpl := ScoresheetTemplate{
		Version: "test.1", Discipline: DisciplineTR, Level: LevelTwo,
		PassThresholdPct:   70,
		MaxInsufficients:   0,
		AvailableModifiers: []ScoresheetModifier{mod},
		Phases: []PhaseTemplate{
			{Code: "TRK", Exercises: []ExerciseTemplate{
				{Kind: CriteriaSum, Code: "track", MaxPoints: 100,
					Criteria: []Criterion{{Code: "track.a", MaxPoints: 100}}},
			}},
		},
	}
	if err := tpl.Validate(); err != nil {
		t.Fatalf("template invalid: %v", err)
	}
	sheet, _ := tpl.BuildConcrete(nil)
	inputs := ScoresheetInputs{
		CriterionScores: []CriterionScore{
			{ExerciseCode: "track", CriterionCode: "track.a", Points: 100},
		},
		Modifiers: []ModifierApplication{{ModifierCode: "LIFELINE"}},
	}
	res, err := EvaluateScoresheet(inputs, sheet, tpl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.TotalPoints != 80 {
		t.Errorf("TotalPoints = %d, want 80 (100 - 20 lifeline)", res.TotalPoints)
	}
	// 80/100 = 80% would normally be Good, but should still award Good
	// (under VG cap). 80 is below VG cutoff of 86, so it's Good anyway.
	if res.FinalTier != TierGood {
		t.Errorf("FinalTier = %v, want Good", res.FinalTier)
	}
	if len(res.AppliedModifiers) != 1 || res.AppliedModifiers[0].PointDelta != -20 {
		t.Errorf("AppliedModifiers = %v", res.AppliedModifiers)
	}
	if !res.Passed {
		t.Error("80%% should pass at 70%% threshold")
	}
}

// TestEvaluateScoresheet_TierCapBitesExcellent: with a perfect score,
// the cap actually constrains FinalTier.
func TestEvaluateScoresheet_TierCapBitesExcellent(t *testing.T) {
	veryGood := TierVeryGood
	mod := ScoresheetModifier{
		Code: "CAP_ONLY", PointDelta: 0, MaxTier: &veryGood,
	}
	tpl := ScoresheetTemplate{
		Version: "test.1", Discipline: DisciplineOB, Level: LevelOne,
		PassThresholdPct: 70, MaxInsufficients: 1,
		AvailableModifiers: []ScoresheetModifier{mod},
		Phases: []PhaseTemplate{
			{Code: "P1", Exercises: []ExerciseTemplate{
				{Kind: CriteriaSum, Code: "1.1", MaxPoints: 10,
					Criteria: []Criterion{{Code: "1.1.a", MaxPoints: 10}}},
			}},
		},
	}
	sheet, _ := tpl.BuildConcrete(nil)
	inputs := ScoresheetInputs{
		CriterionScores: []CriterionScore{
			{ExerciseCode: "1.1", CriterionCode: "1.1.a", Points: 10},
		},
		Modifiers: []ModifierApplication{{ModifierCode: "CAP_ONLY"}},
	}
	res, _ := EvaluateScoresheet(inputs, sheet, tpl)
	if res.FinalTier != TierVeryGood {
		t.Errorf("FinalTier = %v, want VeryGood (capped down from Excellent)", res.FinalTier)
	}
}

// TestEvaluateScoresheet_AggregateNotDoubleCounted: an Aggregate plus
// its components must not double-count toward the scoresheet total.
func TestEvaluateScoresheet_AggregateNotDoubleCounted(t *testing.T) {
	tpl := ScoresheetTemplate{
		Version: "test.1", Discipline: DisciplineDT, Level: LevelTwo,
		PassThresholdPct: 70, MaxInsufficients: 0,
		Phases: []PhaseTemplate{
			{Code: "DT", Exercises: []ExerciseTemplate{
				{Kind: CriteriaSum, Code: "search1", MaxPoints: 5,
					IsAggregateComponent: true,
					Criteria:             []Criterion{{Code: "search1.a", MaxPoints: 5}}},
				{Kind: CriteriaSum, Code: "search2", MaxPoints: 5,
					IsAggregateComponent: true,
					Criteria:             []Criterion{{Code: "search2.a", MaxPoints: 5}}},
				{Kind: Aggregate, Code: "hunt-drive", MaxPoints: 10,
					AggregateOf: []string{"search1", "search2"}},
			}},
		},
	}
	if err := tpl.Validate(); err != nil {
		t.Fatalf("invalid: %v", err)
	}
	sheet, _ := tpl.BuildConcrete(nil)
	if sheet.MaxPoints != 10 {
		t.Errorf("sheet.MaxPoints = %d, want 10 (components excluded)", sheet.MaxPoints)
	}
	inputs := ScoresheetInputs{
		CriterionScores: []CriterionScore{
			{ExerciseCode: "search1", CriterionCode: "search1.a", Points: 4},
			{ExerciseCode: "search2", CriterionCode: "search2.a", Points: 5},
		},
	}
	res, err := EvaluateScoresheet(inputs, sheet, tpl)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.TotalPoints != 9 {
		t.Errorf("TotalPoints = %d, want 9 (aggregate only, no double-count)", res.TotalPoints)
	}
}

// TestEvaluateScoresheet_CascadeOverflowReachesTarget verifies that
// excess from a PenaltyLedger source actually deducts from its target.
func TestEvaluateScoresheet_CascadeOverflowReachesTarget(t *testing.T) {
	tpl := ScoresheetTemplate{
		Version: "test.1", Discipline: DisciplineDT, Level: LevelTwo,
		PassThresholdPct: 70, MaxInsufficients: 0,
		Phases: []PhaseTemplate{
			{Code: "DT", Exercises: []ExerciseTemplate{
				{Kind: PenaltyLedger, Code: "ACC", MaxPoints: 25,
					Events: []PenaltyEvent{
						{Code: "drain", Deduction: 20},
						{Code: "fa", Deduction: 20, OverflowTo: "BLANK"},
					}},
				{Kind: PenaltyLedger, Code: "BLANK", MaxPoints: 25,
					Events: []PenaltyEvent{{Code: "noop", Deduction: 1}}},
			}},
		},
	}
	if err := tpl.Validate(); err != nil {
		t.Fatalf("invalid: %v", err)
	}
	sheet, _ := tpl.BuildConcrete(nil)
	inputs := ScoresheetInputs{
		PenaltyOccurrences: []PenaltyOccurrence{
			{ExerciseCode: "ACC", EventCode: "drain"}, // ACC: 25 -> 5
			{ExerciseCode: "ACC", EventCode: "fa"},    // ACC: 5 -> 0, 15 flows
		},
	}
	res, err := EvaluateScoresheet(inputs, sheet, tpl)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	// ACC = 0, BLANK = 25 - 15 (absorbed) = 10. Total = 10.
	if res.TotalPoints != 10 {
		t.Errorf("TotalPoints = %d, want 10", res.TotalPoints)
	}
	var blank ExerciseResult
	for _, r := range res.PerExercise {
		if r.ExerciseCode == "BLANK" {
			blank = r
		}
	}
	if blank.AbsorbedOverflow != 15 {
		t.Errorf("BLANK.AbsorbedOverflow = %d, want 15", blank.AbsorbedOverflow)
	}
}
