package scoring

import (
	"strings"
	"testing"
)

func TestBuildConcrete_SelectAll_PassThrough(t *testing.T) {
	tpl := validTemplate() // single phase P1, single exercise 1.1
	cs, err := tpl.BuildConcrete(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cs.Phases) != 1 || cs.Phases[0].Code != "P1" {
		t.Fatalf("phases = %v", cs.Phases)
	}
	if len(cs.Phases[0].Exercises) != 1 || cs.Phases[0].Exercises[0].Code != "1.1" {
		t.Fatalf("exercises = %v", cs.Phases[0].Exercises)
	}
	if cs.MaxPoints != 5 {
		t.Errorf("MaxPoints = %d, want 5", cs.MaxPoints)
	}
	if cs.TemplateVersion != tpl.Version {
		t.Errorf("TemplateVersion = %q, want %q", cs.TemplateVersion, tpl.Version)
	}
}

func TestBuildConcrete_SelectFromInventory_Resolves(t *testing.T) {
	tpl := ScoresheetTemplate{
		Version:          "test.1",
		Discipline:       DisciplineOB,
		Level:            LevelTwo,
		PassThresholdPct: 70,
		Phases: []PhaseTemplate{
			{
				Code: "P3",
				Exercises: []ExerciseTemplate{
					mkCriteriaExercise("3.1", 5),
					mkCriteriaExercise("3.2", 5),
					mkCriteriaExercise("3.3", 5),
					mkCriteriaExercise("3.4", 5),
				},
			},
		},
		SelectionRule: SelectionRule{
			PerPhase: map[string]PhaseSelection{
				"P3": {Mode: SelectFromInventory, Min: 3, Max: 4},
			},
		},
	}
	if err := tpl.Validate(); err != nil {
		t.Fatalf("template invalid: %v", err)
	}
	cs, err := tpl.BuildConcrete(map[string][]string{
		"P3": {"3.1", "3.2", "3.4"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cs.Phases[0].Exercises) != 3 {
		t.Errorf("got %d exercises, want 3", len(cs.Phases[0].Exercises))
	}
	if cs.MaxPoints != 15 {
		t.Errorf("MaxPoints = %d, want 15", cs.MaxPoints)
	}
}

func TestBuildConcrete_SelectFromInventory_CountOutOfRange(t *testing.T) {
	tpl := mkInventoryTemplate()
	_, err := tpl.BuildConcrete(map[string][]string{
		"P3": {"3.1", "3.2"}, // below Min=3
	})
	if err == nil || !strings.Contains(err.Error(), "must be in") {
		t.Fatalf("expected range error, got %v", err)
	}
}

func TestBuildConcrete_SelectFromInventory_UnknownCode(t *testing.T) {
	tpl := mkInventoryTemplate()
	_, err := tpl.BuildConcrete(map[string][]string{
		"P3": {"3.1", "3.2", "9.9"},
	})
	if err == nil || !strings.Contains(err.Error(), "not in phase") {
		t.Fatalf("expected unknown-code error, got %v", err)
	}
}

func TestBuildConcrete_SelectFromInventory_DuplicateSelection(t *testing.T) {
	tpl := mkInventoryTemplate()
	_, err := tpl.BuildConcrete(map[string][]string{
		"P3": {"3.1", "3.1", "3.2"},
	})
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate error, got %v", err)
	}
}

func TestBuildConcrete_L1OB_FullPassthrough(t *testing.T) {
	// L1 OB uses SelectAll everywhere; BuildConcrete should produce a
	// 120-point concrete sheet with every exercise carried over.
	tpl := mkL1OBClone(t)
	cs, err := tpl.BuildConcrete(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cs.MaxPoints != 120 {
		t.Errorf("MaxPoints = %d, want 120", cs.MaxPoints)
	}
	if len(cs.Phases) != 4 {
		t.Errorf("len(Phases) = %d, want 4", len(cs.Phases))
	}
}

// Helpers.

func mkCriteriaExercise(code string, maxPts Points) ExerciseTemplate {
	return ExerciseTemplate{
		Kind:      CriteriaSum,
		Code:      code,
		Name:      code,
		MaxPoints: maxPts,
		Criteria:  []Criterion{{Code: code + ".a", MaxPoints: maxPts}},
	}
}

func mkInventoryTemplate() ScoresheetTemplate {
	return ScoresheetTemplate{
		Version:          "test.1",
		Discipline:       DisciplineOB,
		Level:            LevelTwo,
		PassThresholdPct: 70,
		Phases: []PhaseTemplate{
			{
				Code: "P3",
				Exercises: []ExerciseTemplate{
					mkCriteriaExercise("3.1", 5),
					mkCriteriaExercise("3.2", 5),
					mkCriteriaExercise("3.3", 5),
				},
			},
		},
		SelectionRule: SelectionRule{
			PerPhase: map[string]PhaseSelection{
				"P3": {Mode: SelectFromInventory, Min: 3, Max: 4},
			},
		},
	}
}

// mkL1OBClone inlines a minimal stand-in for the real L1OB template
// — same shape (4 phases summing to 120 points) without the templates
// import cycle.
func mkL1OBClone(t *testing.T) ScoresheetTemplate {
	t.Helper()
	return ScoresheetTemplate{
		Version:          "test.1",
		Discipline:       DisciplineOB,
		Level:            LevelOne,
		PassThresholdPct: 70,
		MaxInsufficients: 1,
		Phases: []PhaseTemplate{
			{Code: "P1", Exercises: []ExerciseTemplate{mkCriteriaExercise("1.1", 30)}},
			{Code: "P2", Exercises: []ExerciseTemplate{mkCriteriaExercise("2.1", 30)}},
			{Code: "P3", Exercises: []ExerciseTemplate{mkCriteriaExercise("3.1", 35)}},
			{Code: "P4", Exercises: []ExerciseTemplate{mkCriteriaExercise("4.1", 25)}},
		},
	}
}
