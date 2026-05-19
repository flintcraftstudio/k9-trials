package scoring

import (
	"strings"
	"testing"
)

func TestEvaluateExercise_CriteriaSum_Basic(t *testing.T) {
	tpl := ExerciseTemplate{
		Kind:      CriteriaSum,
		Code:      "1.1",
		MaxPoints: 10,
		Criteria: []Criterion{
			{Code: "1.1.a", MaxPoints: 4},
			{Code: "1.1.b", MaxPoints: 6},
		},
	}
	in := ExerciseInputs{
		CriterionScores: []CriterionScore{
			{ExerciseCode: "1.1", CriterionCode: "1.1.a", Points: 3},
			{ExerciseCode: "1.1", CriterionCode: "1.1.b", Points: 6},
		},
	}
	res, err := EvaluateExercise(tpl, in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Points != 9 || res.MaxPoints != 10 {
		t.Errorf("Points/MaxPoints = %d/%d, want 9/10", res.Points, res.MaxPoints)
	}
	if res.Tier != TierVeryGood {
		t.Errorf("Tier = %v, want VeryGood (9/10)", res.Tier)
	}
	if res.IsInsufficient {
		t.Error("IsInsufficient should be false")
	}
}

func TestEvaluateExercise_CriteriaSum_UnknownCriterion(t *testing.T) {
	tpl := ExerciseTemplate{
		Kind:      CriteriaSum,
		Code:      "1.1",
		MaxPoints: 5,
		Criteria:  []Criterion{{Code: "1.1.a", MaxPoints: 5}},
	}
	in := ExerciseInputs{
		CriterionScores: []CriterionScore{
			{CriterionCode: "1.1.z", Points: 3},
		},
	}
	_, err := EvaluateExercise(tpl, in)
	if err == nil || !strings.Contains(err.Error(), "unknown criterion") {
		t.Fatalf("expected unknown-criterion error, got %v", err)
	}
}

func TestEvaluateExercise_CriteriaSum_InsufficientBand(t *testing.T) {
	// 6/10 = 60%, below Sufficient threshold of 70%.
	tpl := ExerciseTemplate{
		Kind:      CriteriaSum,
		Code:      "1.1",
		MaxPoints: 10,
		Criteria:  []Criterion{{Code: "1.1.a", MaxPoints: 10}},
	}
	in := ExerciseInputs{
		CriterionScores: []CriterionScore{{CriterionCode: "1.1.a", Points: 6}},
	}
	res, _ := EvaluateExercise(tpl, in)
	if res.Tier != TierInsufficient {
		t.Errorf("Tier = %v, want Insufficient", res.Tier)
	}
	if !res.IsInsufficient {
		t.Error("IsInsufficient should be true")
	}
}

func TestEvaluateExercise_AggregateComponent_NotInsufficient(t *testing.T) {
	// Even if a component's tier is Insufficient, IsInsufficient
	// is false (parent Aggregate carries the insufficiency).
	tpl := ExerciseTemplate{
		Kind:                 CriteriaSum,
		Code:                 "c1",
		MaxPoints:            10,
		IsAggregateComponent: true,
		Criteria:             []Criterion{{Code: "c1.a", MaxPoints: 10}},
	}
	in := ExerciseInputs{
		CriterionScores: []CriterionScore{{CriterionCode: "c1.a", Points: 3}},
	}
	res, _ := EvaluateExercise(tpl, in)
	if res.Tier != TierInsufficient {
		t.Errorf("Tier = %v, want Insufficient", res.Tier)
	}
	if res.IsInsufficient {
		t.Error("IsInsufficient must be false for components")
	}
}

func TestEvaluateExercise_PenaltyLedger_Basic(t *testing.T) {
	tpl := ExerciseTemplate{
		Kind:      PenaltyLedger,
		Code:      "ACC",
		MaxPoints: 25,
		Events:    []PenaltyEvent{{Code: "miss", Deduction: 5}},
	}
	in := ExerciseInputs{
		PenaltyOccurrences: []PenaltyOccurrence{
			{EventCode: "miss"},
			{EventCode: "miss"},
		},
	}
	res, _ := EvaluateExercise(tpl, in)
	if res.Points != 15 {
		t.Errorf("Points = %d, want 15 (25 - 2*5)", res.Points)
	}
}

func TestEvaluateExercise_PenaltyLedger_FloorsAtZero(t *testing.T) {
	tpl := ExerciseTemplate{
		Kind:      PenaltyLedger,
		Code:      "ACC",
		MaxPoints: 10,
		Events:    []PenaltyEvent{{Code: "big", Deduction: 6}},
	}
	in := ExerciseInputs{
		PenaltyOccurrences: []PenaltyOccurrence{
			{EventCode: "big"},
			{EventCode: "big"}, // 10 - 6 - 6 would be -2; floor at 0
		},
	}
	res, _ := EvaluateExercise(tpl, in)
	if res.Points != 0 {
		t.Errorf("Points = %d, want 0", res.Points)
	}
	if res.CascadeOverflow != 0 {
		t.Errorf("CascadeOverflow = %d, want 0 (no OverflowTo)", res.CascadeOverflow)
	}
}

func TestEvaluateExercise_PenaltyLedger_CascadeOverflowCaptured(t *testing.T) {
	// L2/L3 DET scenario: -20 deduction lands on a 25-point exercise
	// that already has 5 points left, so 15 flows out.
	tpl := ExerciseTemplate{
		Kind:      PenaltyLedger,
		Code:      "ACC",
		MaxPoints: 25,
		Events: []PenaltyEvent{
			{Code: "drain", Deduction: 20},
			{Code: "fa", Deduction: 20, OverflowTo: "BLANK"},
		},
	}
	in := ExerciseInputs{
		PenaltyOccurrences: []PenaltyOccurrence{
			{EventCode: "drain"}, // running 25 -> 5
			{EventCode: "fa"},    // running 5 -> 0, excess 15 flows
		},
	}
	res, _ := EvaluateExercise(tpl, in)
	if res.Points != 0 {
		t.Errorf("Points = %d, want 0", res.Points)
	}
	if res.CascadeOverflow != 15 {
		t.Errorf("CascadeOverflow = %d, want 15", res.CascadeOverflow)
	}
}

func TestEvaluateExercise_PenaltyLedger_AbsorbsCascadeInflow(t *testing.T) {
	tpl := ExerciseTemplate{
		Kind:      PenaltyLedger,
		Code:      "BLANK",
		MaxPoints: 25,
		Events:    []PenaltyEvent{{Code: "miss", Deduction: 3}},
	}
	in := ExerciseInputs{
		PenaltyOccurrences: []PenaltyOccurrence{{EventCode: "miss"}}, // 25 -> 22
		CascadeInflow:      15,                                       // 22 -> 7
	}
	res, _ := EvaluateExercise(tpl, in)
	if res.Points != 7 {
		t.Errorf("Points = %d, want 7", res.Points)
	}
	if res.AbsorbedOverflow != 15 {
		t.Errorf("AbsorbedOverflow = %d, want 15", res.AbsorbedOverflow)
	}
}

func TestEvaluateExercise_PenaltyLedger_AutoInsufficientFlag(t *testing.T) {
	// PenaltyEvent.AutoInsufficientOnOccurrence forces Insufficient
	// even if Points math hasn't reached the band's floor.
	tpl := ExerciseTemplate{
		Kind:      PenaltyLedger,
		Code:      "TRK-COMP",
		MaxPoints: 25,
		Events: []PenaltyEvent{
			{Code: "abandon", Deduction: 1, AutoInsufficientOnOccurrence: true},
		},
	}
	in := ExerciseInputs{
		PenaltyOccurrences: []PenaltyOccurrence{{EventCode: "abandon"}},
	}
	res, _ := EvaluateExercise(tpl, in)
	if res.Tier != TierInsufficient {
		t.Errorf("Tier = %v, want Insufficient (auto flag)", res.Tier)
	}
	if !res.IsInsufficient {
		t.Error("IsInsufficient should be true")
	}
	if res.Points != 24 {
		t.Errorf("Points = %d, want 24 (math unaffected by auto flag)", res.Points)
	}
}

func TestEvaluateExercise_Aggregate_SumsSiblings(t *testing.T) {
	tpl := ExerciseTemplate{
		Kind:        Aggregate,
		Code:        "total",
		MaxPoints:   15,
		AggregateOf: []string{"c1", "c2"},
	}
	in := ExerciseInputs{
		SiblingResults: map[string]ExerciseResult{
			"c1": {Points: 4},
			"c2": {Points: 8},
		},
	}
	res, _ := EvaluateExercise(tpl, in)
	if res.Points != 12 {
		t.Errorf("Points = %d, want 12", res.Points)
	}
}

func TestEvaluateExercise_Aggregate_MissingSibling(t *testing.T) {
	tpl := ExerciseTemplate{
		Kind:        Aggregate,
		Code:        "total",
		MaxPoints:   15,
		AggregateOf: []string{"c1", "c2"},
	}
	in := ExerciseInputs{
		SiblingResults: map[string]ExerciseResult{"c1": {Points: 4}},
	}
	_, err := EvaluateExercise(tpl, in)
	if err == nil || !strings.Contains(err.Error(), "missing sibling") {
		t.Fatalf("expected missing-sibling error, got %v", err)
	}
}

func TestEvaluateExercise_AutoNQExercise_ZerosScore(t *testing.T) {
	tpl := ExerciseTemplate{
		Kind:      CriteriaSum,
		Code:      "1.1",
		MaxPoints: 10,
		Criteria:  []Criterion{{Code: "1.1.a", MaxPoints: 10}},
		AutoTriggers: []AutoTrigger{
			{Code: "1.1.nq", Scope: AutoNQExercise},
		},
	}
	in := ExerciseInputs{
		CriterionScores: []CriterionScore{{CriterionCode: "1.1.a", Points: 10}},
		AutoTriggers:    []AutoTriggerFiring{{TriggerCode: "1.1.nq"}},
	}
	res, _ := EvaluateExercise(tpl, in)
	if res.Points != 0 {
		t.Errorf("Points = %d, want 0 (AutoNQExercise zeros)", res.Points)
	}
	if res.Tier != TierInsufficient {
		t.Errorf("Tier = %v, want Insufficient", res.Tier)
	}
	if len(res.AutoNQFired) != 1 || res.AutoNQFired[0] != "1.1.nq" {
		t.Errorf("AutoNQFired = %v, want [1.1.nq]", res.AutoNQFired)
	}
}

func TestEvaluateExercise_AutoNQTrial_RecordedButPointsKept(t *testing.T) {
	// AutoNQTrial doesn't zero the exercise — that's handled at the
	// scoresheet level. The firing is just recorded.
	tpl := ExerciseTemplate{
		Kind:      CriteriaSum,
		Code:      "1.2",
		MaxPoints: 5,
		Criteria:  []Criterion{{Code: "1.2.a", MaxPoints: 5}},
		AutoTriggers: []AutoTrigger{
			{Code: "trial-nq", Scope: AutoNQTrial},
		},
	}
	in := ExerciseInputs{
		CriterionScores: []CriterionScore{{CriterionCode: "1.2.a", Points: 4}},
		AutoTriggers:    []AutoTriggerFiring{{TriggerCode: "trial-nq"}},
	}
	res, _ := EvaluateExercise(tpl, in)
	if res.Points != 4 {
		t.Errorf("Points = %d, want 4 (AutoNQTrial doesn't zero exercise)", res.Points)
	}
	if len(res.AutoNQFired) != 1 {
		t.Errorf("AutoNQFired = %v, want one entry", res.AutoNQFired)
	}
}

func TestEvaluateExercise_UnknownAutoTrigger(t *testing.T) {
	tpl := ExerciseTemplate{
		Kind:      CriteriaSum,
		Code:      "1.1",
		MaxPoints: 5,
		Criteria:  []Criterion{{Code: "1.1.a", MaxPoints: 5}},
	}
	in := ExerciseInputs{
		AutoTriggers: []AutoTriggerFiring{{TriggerCode: "ghost"}},
	}
	_, err := EvaluateExercise(tpl, in)
	if err == nil || !strings.Contains(err.Error(), "unknown AutoTrigger") {
		t.Fatalf("expected unknown-trigger error, got %v", err)
	}
}
