package templates

import (
	"testing"

	"github.com/flintcraftstudio/k9-trials/internal/scoring"
)

// TestL1OB_Validates is the smoke test the L1OB() doc comment specifies:
// the hardcoded rulebook template must build without panic and validate
// against the discriminated-union rules in ExerciseTemplate.Validate.
func TestL1OB_Validates(t *testing.T) {
	tpl := L1OB() // panics on validation failure
	if tpl.Discipline != scoring.DisciplineOB {
		t.Errorf("Discipline = %v, want OB", tpl.Discipline)
	}
	if tpl.Level != scoring.LevelOne {
		t.Errorf("Level = %v, want LevelOne", tpl.Level)
	}
	if len(tpl.Phases) != 4 {
		t.Fatalf("len(Phases) = %d, want 4", len(tpl.Phases))
	}
}

// TestL1OB_PhaseTotals locks in the phase totals documented in l1ob.go:
//
//	P1: 30, P2: 30, P3: 35, P4: 25. Discipline total: 120.
func TestL1OB_PhaseTotals(t *testing.T) {
	tpl := L1OB()
	wantByCode := map[string]scoring.Points{
		"P1": 30,
		"P2": 30,
		"P3": 35,
		"P4": 25,
	}
	var total scoring.Points
	for _, ph := range tpl.Phases {
		want, ok := wantByCode[ph.Code]
		if !ok {
			t.Errorf("unexpected phase code %q", ph.Code)
			continue
		}
		got := ph.MaxPoints()
		if got != want {
			t.Errorf("phase %s MaxPoints = %d, want %d", ph.Code, got, want)
		}
		total += got
	}
	if total != 120 {
		t.Errorf("discipline total = %d, want 120", total)
	}
}

// TestL1OB_KeyLookups checks that the documented exercise and
// criterion codes resolve.
func TestL1OB_KeyLookups(t *testing.T) {
	tpl := L1OB()
	if _, ok := tpl.FindExercise("1.1"); !ok {
		t.Error("FindExercise(1.1) missed")
	}
	if _, ok := tpl.FindCriterion("2.2", "2.2.e"); !ok {
		t.Error("FindCriterion(2.2, 2.2.e) missed — Time Component")
	}
	if _, ok := tpl.FindExercise("4.4"); !ok {
		t.Error("FindExercise(4.4) missed — Phase 4 Decoy Neutrality")
	}
}
