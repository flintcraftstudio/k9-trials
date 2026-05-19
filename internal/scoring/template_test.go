package scoring

import (
	"strings"
	"testing"
)

// validCriteriaSum builds a known-good CriteriaSum exercise for use as
// a base in negative-validation tests.
func validCriteriaSum() ExerciseTemplate {
	return ExerciseTemplate{
		Kind:      CriteriaSum,
		Code:      "1.1",
		Name:      "Test",
		MaxPoints: 5,
		Criteria: []Criterion{
			{Code: "1.1.a", MaxPoints: 3, Description: "first"},
			{Code: "1.1.b", MaxPoints: 2, Description: "second"},
		},
	}
}

func TestValidate_CriteriaSum_Valid(t *testing.T) {
	ex := validCriteriaSum()
	if err := ex.Validate(nil); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestValidate_CriteriaSum_SumMismatch(t *testing.T) {
	ex := validCriteriaSum()
	ex.MaxPoints = 6 // criteria sum to 5
	err := ex.Validate(nil)
	if err == nil || !strings.Contains(err.Error(), "criteria sum") {
		t.Fatalf("expected sum mismatch error, got %v", err)
	}
}

func TestValidate_CriteriaSum_RejectsEvents(t *testing.T) {
	ex := validCriteriaSum()
	ex.Events = []PenaltyEvent{{Code: "e1", Deduction: 1}}
	err := ex.Validate(nil)
	if err == nil || !strings.Contains(err.Error(), "must not have events") {
		t.Fatalf("expected events rejection, got %v", err)
	}
}

func TestValidate_PenaltyLedger_OverflowToResolves(t *testing.T) {
	siblings := []ExerciseTemplate{
		{Kind: CriteriaSum, Code: "sink", MaxPoints: 5,
			Criteria: []Criterion{{Code: "sink.a", MaxPoints: 5}}},
	}
	ex := ExerciseTemplate{
		Kind:      PenaltyLedger,
		Code:      "source",
		MaxPoints: 25,
		Events:    []PenaltyEvent{{Code: "fa", Deduction: 20, OverflowTo: "sink"}},
	}
	if err := ex.Validate(append(siblings, ex)); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestValidate_PenaltyLedger_OverflowToUnresolved(t *testing.T) {
	ex := ExerciseTemplate{
		Kind:      PenaltyLedger,
		Code:      "source",
		MaxPoints: 25,
		Events:    []PenaltyEvent{{Code: "fa", Deduction: 20, OverflowTo: "nonexistent"}},
	}
	err := ex.Validate([]ExerciseTemplate{ex})
	if err == nil || !strings.Contains(err.Error(), "OverflowTo") {
		t.Fatalf("expected OverflowTo error, got %v", err)
	}
}

func TestValidate_Aggregate_ComponentSumMatches(t *testing.T) {
	siblings := []ExerciseTemplate{
		{Kind: CriteriaSum, Code: "c1", MaxPoints: 5, IsAggregateComponent: true,
			Criteria: []Criterion{{Code: "c1.a", MaxPoints: 5}}},
		{Kind: CriteriaSum, Code: "c2", MaxPoints: 10, IsAggregateComponent: true,
			Criteria: []Criterion{{Code: "c2.a", MaxPoints: 10}}},
	}
	agg := ExerciseTemplate{
		Kind:        Aggregate,
		Code:        "total",
		MaxPoints:   15,
		AggregateOf: []string{"c1", "c2"},
	}
	if err := agg.Validate(append(siblings, agg)); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestValidate_Aggregate_ComponentNotMarked(t *testing.T) {
	siblings := []ExerciseTemplate{
		{Kind: CriteriaSum, Code: "c1", MaxPoints: 5,
			Criteria: []Criterion{{Code: "c1.a", MaxPoints: 5}}}, // not IsAggregateComponent
	}
	agg := ExerciseTemplate{
		Kind:        Aggregate,
		Code:        "total",
		MaxPoints:   5,
		AggregateOf: []string{"c1"},
	}
	err := agg.Validate(append(siblings, agg))
	if err == nil || !strings.Contains(err.Error(), "IsAggregateComponent") {
		t.Fatalf("expected IsAggregateComponent error, got %v", err)
	}
}

func TestPhaseTemplate_MaxPoints_ExcludesComponents(t *testing.T) {
	ph := PhaseTemplate{
		Code: "P1",
		Exercises: []ExerciseTemplate{
			{Kind: CriteriaSum, Code: "c1", MaxPoints: 5, IsAggregateComponent: true,
				Criteria: []Criterion{{Code: "c1.a", MaxPoints: 5}}},
			{Kind: CriteriaSum, Code: "c2", MaxPoints: 10, IsAggregateComponent: true,
				Criteria: []Criterion{{Code: "c2.a", MaxPoints: 10}}},
			{Kind: Aggregate, Code: "total", MaxPoints: 15,
				AggregateOf: []string{"c1", "c2"}},
			{Kind: CriteriaSum, Code: "standalone", MaxPoints: 8,
				Criteria: []Criterion{{Code: "standalone.a", MaxPoints: 8}}},
		},
	}
	got := ph.MaxPoints()
	want := Points(15 + 8) // aggregate + standalone, components excluded
	if got != want {
		t.Errorf("MaxPoints() = %d, want %d", got, want)
	}
}

// validTemplate returns a small known-good ScoresheetTemplate.
func validTemplate() ScoresheetTemplate {
	return ScoresheetTemplate{
		Version:          "test.1",
		Discipline:       DisciplineOB,
		Level:            LevelOne,
		PassThresholdPct: 70,
		MaxInsufficients: 1,
		Phases: []PhaseTemplate{
			{
				Code: "P1",
				Exercises: []ExerciseTemplate{
					validCriteriaSum(),
				},
			},
		},
	}
}

func TestScoresheetTemplate_Validate_AcceptsValid(t *testing.T) {
	if err := validTemplate().Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestScoresheetTemplate_Validate_DuplicateExerciseCode(t *testing.T) {
	tpl := validTemplate()
	tpl.Phases = append(tpl.Phases, PhaseTemplate{
		Code: "P2",
		Exercises: []ExerciseTemplate{
			validCriteriaSum(), // same code "1.1" as in P1
		},
	})
	err := tpl.Validate()
	if err == nil || !strings.Contains(err.Error(), "duplicate exercise") {
		t.Fatalf("expected duplicate exercise error, got %v", err)
	}
}

func TestScoresheetTemplate_Validate_InvalidLevel(t *testing.T) {
	tpl := validTemplate()
	tpl.Level = 99
	err := tpl.Validate()
	if err == nil || !strings.Contains(err.Error(), "invalid level") {
		t.Fatalf("expected level error, got %v", err)
	}
}

func TestScoresheetTemplate_FindExercise(t *testing.T) {
	tpl := validTemplate()
	ex, ok := tpl.FindExercise("1.1")
	if !ok || ex.Code != "1.1" {
		t.Fatalf("FindExercise(1.1) = (%v, %v)", ex, ok)
	}
	if _, ok := tpl.FindExercise("nonexistent"); ok {
		t.Fatal("expected miss on nonexistent code")
	}
}

func TestScoresheetTemplate_FindCriterion(t *testing.T) {
	tpl := validTemplate()
	c, ok := tpl.FindCriterion("1.1", "1.1.a")
	if !ok || c.Code != "1.1.a" {
		t.Fatalf("FindCriterion(1.1, 1.1.a) = (%v, %v)", c, ok)
	}
	if _, ok := tpl.FindCriterion("1.1", "nope"); ok {
		t.Fatal("expected miss on bad criterion code")
	}
	if _, ok := tpl.FindCriterion("nope", "1.1.a"); ok {
		t.Fatal("expected miss on bad exercise code")
	}
}

func TestScoresheetTemplate_FindModifier(t *testing.T) {
	tpl := validTemplate()
	tier := TierVeryGood
	tpl.AvailableModifiers = []ScoresheetModifier{
		{Code: "L2-TRK-LIFELINE", PointDelta: -20, MaxTier: &tier},
	}
	m, ok := tpl.FindModifier("L2-TRK-LIFELINE")
	if !ok || m.PointDelta != -20 {
		t.Fatalf("FindModifier got (%v, %v)", m, ok)
	}
	if _, ok := tpl.FindModifier("nope"); ok {
		t.Fatal("expected miss on bad modifier code")
	}
}

func TestConcreteScoresheet_FindExercise(t *testing.T) {
	cs := ConcreteScoresheet{
		Phases: []ConcretePhase{
			{Code: "P1", Exercises: []ExerciseTemplate{validCriteriaSum()}},
		},
	}
	if _, ok := cs.FindExercise("1.1"); !ok {
		t.Fatal("expected hit")
	}
	if _, ok := cs.FindExercise("nope"); ok {
		t.Fatal("expected miss")
	}
}
