package scoring

import "testing"

// TestRoundPoints verifies §3.2: nearest whole point, ties round up
// (toward positive infinity). Distinct from math.Round, which rounds
// half away from zero (and would send -0.5 to -1).
func TestRoundPoints(t *testing.T) {
	tests := []struct {
		name string
		raw  float64
		want Points
	}{
		// Rulebook §3.2 published example.
		{"rulebook_example_17.5_to_18", 17.5, 18},

		// Half values at scales encountered in L1 OB
		// (exercise maxes 5/7/8, phase totals 20-35).
		{"half_0.5_to_1", 0.5, 1},
		{"half_4.5_to_5", 4.5, 5},
		{"half_7.5_to_8", 7.5, 8},
		{"half_24.5_to_25", 24.5, 25},

		// Near-half values: guard against naive "always round up"
		// or "always round down" implementations.
		{"below_half_17.4_to_17", 17.4, 17},
		{"above_half_17.6_to_18", 17.6, 18},
		{"just_below_half_0.49_to_0", 0.49, 0},
		{"just_above_half_0.51_to_1", 0.51, 1},

		// Integer-valued inputs pass through unchanged.
		{"integer_0.0_to_0", 0.0, 0},
		{"integer_5.0_to_5", 5.0, 5},
		{"integer_25.0_to_25", 25.0, 25},

		// Negative inputs: documented contract per RoundPoints
		// docstring. Half-up rounds toward positive infinity, so
		// -0.5 -> 0 (not -1 as math.Round would yield).
		{"negative_half_-0.5_to_0", -0.5, 0},
		{"negative_below_half_-0.4_to_0", -0.4, 0},
		{"negative_above_half_-0.6_to_-1", -0.6, -1},
		{"negative_half_-1.5_to_-1", -1.5, -1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := RoundPoints(tc.raw)
			if got != tc.want {
				t.Errorf("RoundPoints(%v) = %d, want %d", tc.raw, got, tc.want)
			}
		})
	}
}
