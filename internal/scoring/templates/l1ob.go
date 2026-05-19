// Package templates holds the hardcoded ScoresheetTemplate
// implementations for each (Discipline, Level) the K9 Elements rulebook
// defines. Each file in this package encodes one (Discipline, Level)
// pair, with every ExerciseTemplate annotated by rulebook section.
//
// Template construction is via factory functions (not package-level
// vars) so that Validate runs at construction time with clear error
// messages, and so that template hot-reloading is possible in tests
// without process restart.
//
// Rulebook citations refer to K9_Elements_Rulebook_Working_Draft.
// The Version field on each template marks the rulebook revision
// the encoding tracks; bump on every rulebook revision that affects
// scoring.
package templates

import (
	"fmt"

	"github.com/flintcraftstudio/k9-trials/internal/scoring"
)

// L1OB returns the Level 1 Obedience scoresheet template encoding
// §7.2 of the rulebook.
//
// Structure:
//   - 4 Phases, 19 Exercises total
//   - Discipline total: 120 points (100 exercise + 20 neutrality)
//   - Pass threshold: 70% = 84/120 (§2.2)
//   - Max Insufficients: 1 (§2.3, L1 rule)
//
// Phase totals:
//   - Phase 1 (Muzzle & Stability): 30 pts (25 exercise + 5 neutrality)
//   - Phase 2 (Scent Work & Field Search): 30 pts (25 + 5)
//   - Phase 3 (Obstacles & Distance Work): 35 pts (30 + 5)
//   - Phase 4 (Retrieve & Final Composure): 25 pts (20 + 5)
//
// All four Decoy Neutrality exercises are encoded as standalone
// CriteriaSum exercises within their respective phases. The rulebook
// §7.2 master section explicitly treats them as scored line items
// eligible for the "one Insufficient permitted" rule, not as
// components of a discipline-wide aggregate.
//
// Re-attempt mechanics for Long Stay (Exercise 1.3) and reset-cue
// mechanics for several Phase 2/3 exercises are NOT modeled here
// (open question: see project notes). For v1 the judge applies these
// caps by adjusting the final criterion score; structured modeling
// is a candidate for v2.
func L1OB() scoring.ScoresheetTemplate {
	t := scoring.ScoresheetTemplate{
		Version:          "2026.1",
		Discipline:       scoring.DisciplineOB,
		Level:            scoring.LevelOne,
		PassThresholdPct: 70,
		MaxInsufficients: 1,
		Phases: []scoring.PhaseTemplate{
			l1obPhase1(),
			l1obPhase2(),
			l1obPhase3(),
			l1obPhase4(),
		},
		// SelectionRule zero-value: every phase is SelectAll.
		// AvailableModifiers: none for L1 OB.
	}
	if err := t.Validate(); err != nil {
		panic(fmt.Sprintf("L1 OB template invalid: %v", err))
	}
	return t
}

// l1obPhase1 encodes §7.2 Phase 1: Muzzle & Stability.
// Phase total: 30 pts (10 + 5 + 5 + 5 + 5 = 30).
func l1obPhase1() scoring.PhaseTemplate {
	return scoring.PhaseTemplate{
		Code: "P1",
		Name: "Phase 1: Muzzle & Stability",
		Exercises: []scoring.ExerciseTemplate{
			// Exercise 1.1 - Muzzle Acceptance & Heeling Pattern (10 points)
			// Rulebook: §7.2 Phase 1 Exercise 1.1
			{
				Kind:      scoring.CriteriaSum,
				Code:      "1.1",
				Name:      "Muzzle Acceptance & Heeling Pattern",
				MaxPoints: 10, // 2+2+2+2+2 = 10
				Criteria: []scoring.Criterion{
					{Code: "1.1.a", MaxPoints: 2, Description: "Muzzle accepted at start line, no resistance during fitting"},
					{Code: "1.1.b", MaxPoints: 2, Description: "Position maintained at handler's left, no forging or lagging"},
					{Code: "1.1.c", MaxPoints: 2, Description: "Smooth turns and pace consistency"},
					{Code: "1.1.d", MaxPoints: 2, Description: "Automatic sit at halt — straight, prompt"},
					{Code: "1.1.e", MaxPoints: 2, Description: "Overall engagement and attitude — dog works as if muzzle is irrelevant; no muzzle pawing, rubbing, or fixation"},
				},
				AutoTriggers: []scoring.AutoTrigger{
					{
						Code:        "1.1.nq.destroy-muzzle",
						Description: "Dog actively destroys or dislodges the muzzle during the exercise",
						Scope:       scoring.AutoNQExercise,
					},
					{
						Code:        "1.1.nq.aggression-muzzle",
						Description: "Aggressive response to muzzling",
						Scope:       scoring.AutoNQExercise,
					},
				},
			},

			// Exercise 1.2 - Muzzle Stability Under Approach (5 points)
			// Rulebook: §7.2 Phase 1 Exercise 1.2
			{
				Kind:      scoring.CriteriaSum,
				Code:      "1.2",
				Name:      "Muzzle Stability Under Approach",
				MaxPoints: 5, // 2+1+1+1 = 5
				Criteria: []scoring.Criterion{
					{Code: "1.2.a", MaxPoints: 2, Description: "Dog maintains position at handler's side without breaking"},
					{Code: "1.2.b", MaxPoints: 1, Description: "No avoidance — no shrinking, hiding, attempted retreat"},
					{Code: "1.2.c", MaxPoints: 1, Description: "No reactivity — no growling, lunging, sustained fixed staring"},
					{Code: "1.2.d", MaxPoints: 1, Description: "Calm body language — relaxed posture, normal breathing, no muzzle fixation"},
				},
				AutoTriggers: []scoring.AutoTrigger{
					{
						Code:        "1.2.nq.aggression-steward",
						Description: "Aggressive display toward steward (lunge, snap, sustained growl)",
						Scope:       scoring.AutoNQTrial, // §4.5: aggression toward people is trial-level NQ
					},
				},
			},

			// Exercise 1.3 - Long Stay (5 points)
			// Rulebook: §7.2 Phase 1 Exercise 1.3
			// Re-attempt rule (max 3 pts on re-attempt, half points, ties round up)
			// NOT modeled in v1 — judge applies cap manually via final criterion scores.
			{
				Kind:      scoring.CriteriaSum,
				Code:      "1.3",
				Name:      "Long Stay",
				MaxPoints: 5, // 3+1+1 = 5
				Criteria: []scoring.Criterion{
					{Code: "1.3.a", MaxPoints: 3, Description: "Dog remains in declared position the full 30 seconds"},
					{Code: "1.3.b", MaxPoints: 1, Description: "Position is clean — no shifting, no creeping, no whining"},
					{Code: "1.3.c", MaxPoints: 1, Description: "Calm demeanor — not visibly stressed, not fixating on environment"},
				},
				AutoTriggers: []scoring.AutoTrigger{
					{
						Code:        "1.3.nq.approach-handler",
						Description: "Dog leaves position and approaches handler",
						Scope:       scoring.AutoNQExercise,
					},
					{
						Code:        "1.3.nq.leaves-field",
						Description: "Dog leaves the field",
						Scope:       scoring.AutoNQTrial, // §4.5
					},
				},
			},

			// Exercise 1.4 - Food Refusal (5 points)
			// Rulebook: §7.2 Phase 1 Exercise 1.4
			{
				Kind:      scoring.CriteriaSum,
				Code:      "1.4",
				Name:      "Food Refusal",
				MaxPoints: 5, // 2+2+1 = 5
				Criteria: []scoring.Criterion{
					{Code: "1.4.a", MaxPoints: 2, Description: "Dog does not approach or attempt to take food"},
					{Code: "1.4.b", MaxPoints: 2, Description: "Single command effective, or natural neutrality without command"},
					{Code: "1.4.c", MaxPoints: 1, Description: "Calm demeanor — not fixated, not whining, not body-blocking"},
				},
				// Rulebook notes "Dog takes food" is Insufficient at exercise
				// level, not auto-trial-NQ. Modeled as a natural scoring outcome
				// (judge enters 0 on criterion 1.4.a), not an AutoTrigger.
			},

			// Exercise 1.5 - Phase 1 Decoy Neutrality (5 points)
			// Rulebook: §7.2 Phase 1 Exercise 1.5
			// Encoded as a CriteriaSum exercise with a single criterion
			// representing the phase-wide neutrality score. The rulebook
			// describes neutrality scoring as "begins at 5, deducted by
			// observed behaviors," which is conceptually a penalty ledger,
			// but the deduction schedule (§7.2 Decoy Neutrality Scoring)
			// is judge-discretionary in degree (-1 to -3 ranges) rather
			// than fixed per-event. Until those ranges are nailed down,
			// modeling as a single CriteriaSum line is cleaner.
			// REVISIT: if Director clarifies fixed-amount deductions, this
			// becomes a PenaltyLedger exercise.
			{
				Kind:      scoring.CriteriaSum,
				Code:      "1.5",
				Name:      "Phase 1 Decoy Neutrality",
				MaxPoints: 5,
				Criteria: []scoring.Criterion{
					{Code: "1.5.a", MaxPoints: 5, Description: "Dog stays task-focused throughout Phase 1; decoys appear irrelevant"},
				},
				AutoTriggers: []scoring.AutoTrigger{
					{
						Code:        "1.5.nq.charge-bite",
						Description: "Position break to charge a decoy, making contact or attempting a bite",
						Scope:       scoring.AutoNQPhase, // per §7.2 Decoy Neutrality Scoring
					},
				},
			},
		},
	}
}

// l1obPhase2 encodes §7.2 Phase 2: Scent Work & Field Search.
// Phase total: 30 pts (5 + 15 + 5 + 5 = 30).
func l1obPhase2() scoring.PhaseTemplate {
	return scoring.PhaseTemplate{
		Code: "P2",
		Name: "Phase 2: Scent Work & Field Search",
		Exercises: []scoring.ExerciseTemplate{
			// Exercise 2.1 - Heel to Search Area & Release (5 points)
			// Rulebook: §7.2 Phase 2 Exercise 2.1
			{
				Kind:      scoring.CriteriaSum,
				Code:      "2.1",
				Name:      "Heel to Search Area & Release",
				MaxPoints: 5, // 2+1+1+1 = 5
				Criteria: []scoring.Criterion{
					{Code: "2.1.a", MaxPoints: 2, Description: "Controlled heel from table area to search area entrance"},
					{Code: "2.1.b", MaxPoints: 1, Description: "Calm leash removal / pre-search setup, no breaking"},
					{Code: "2.1.c", MaxPoints: 1, Description: "Single, clear search cue (no repeated commands)"},
					{Code: "2.1.d", MaxPoints: 1, Description: "Dog enters search area on cue, no handler push or escort"},
				},
			},

			// Exercise 2.2 - Independent Search & Article Retrieve (15 points)
			// Rulebook: §7.2 Phase 2 Exercise 2.2
			//
			// Note on the Time Component criterion: rulebook defines a
			// bucket table (≤10s=5pts, 11-15s=4pts, etc.). The tablet UI
			// presents the buckets and writes the corresponding integer
			// to this criterion's points. The domain model sees only the
			// final point value.
			//
			// Reset-cue cap (max 8 points on reset) is NOT modeled in v1.
			// Judge applies the cap by adjusting final criterion scores.
			{
				Kind:      scoring.CriteriaSum,
				Code:      "2.2",
				Name:      "Independent Search & Article Retrieve",
				MaxPoints: 15, // 3+2+2+3+5 = 15
				Criteria: []scoring.Criterion{
					{Code: "2.2.a", MaxPoints: 3, Description: "Drive to hunt — dog enters area and immediately engages in active search behavior"},
					{Code: "2.2.b", MaxPoints: 2, Description: "Independence — dog works without ongoing handler input or direction"},
					{Code: "2.2.c", MaxPoints: 2, Description: "Coverage — dog systematically works the area, including behind blinds"},
					{Code: "2.2.d", MaxPoints: 3, Description: "Article location accuracy — dog locates the actual article (not a distraction or false indication)"},
					{Code: "2.2.e", MaxPoints: 5, Description: "Time component — ≤10s=5, 11-15s=4, 16-20s=3, 21-25s=2, 26-30s=1, >30s=NQ on this exercise"},
				},
				AutoTriggers: []scoring.AutoTrigger{
					{
						Code:        "2.2.nq.handler-enters-area",
						Description: "Handler enters the search area",
						Scope:       scoring.AutoNQExercise,
					},
					{
						Code:        "2.2.nq.directional-cues",
						Description: "Handler gives directional cues (pointing, leading the dog to article)",
						Scope:       scoring.AutoNQExercise,
					},
					{
						Code:        "2.2.nq.no-return",
						Description: "Dog leaves search area and does not return / does not respond to recall",
						Scope:       scoring.AutoNQExercise,
					},
					{
						Code:        "2.2.nq.time-cap",
						Description: "Time exceeds 30 seconds with no article pickup",
						Scope:       scoring.AutoNQExercise,
					},
					{
						Code:        "2.2.nq.destroys-article",
						Description: "Dog destroys the article in the search area",
						Scope:       scoring.AutoNQExercise,
					},
				},
			},

			// Exercise 2.3 - Delivery & Recovery (5 points)
			// Rulebook: §7.2 Phase 2 Exercise 2.3
			{
				Kind:      scoring.CriteriaSum,
				Code:      "2.3",
				Name:      "Delivery & Recovery",
				MaxPoints: 5, // 1+1+2+1 = 5
				Criteria: []scoring.Criterion{
					{Code: "2.3.a", MaxPoints: 1, Description: "Clean carry out of search area (no dropping, no chewing, no playing)"},
					{Code: "2.3.b", MaxPoints: 1, Description: "Direct return to handler (no detours, no decoy fixation)"},
					{Code: "2.3.c", MaxPoints: 2, Description: "Clean delivery per declared style (front-sit-and-hold OR direct-to-hand)"},
					{Code: "2.3.d", MaxPoints: 1, Description: "Recovery to heel position under handler control"},
				},
				AutoTriggers: []scoring.AutoTrigger{
					{
						Code:        "2.3.nq.destroys-article",
						Description: "Dog destroys article during carry",
						Scope:       scoring.AutoNQExercise,
					},
					{
						Code:        "2.3.nq.refuses-delivery",
						Description: "Dog refuses to deliver article (drops and walks away)",
						Scope:       scoring.AutoNQExercise,
					},
					{
						Code:        "2.3.nq.refuses-return",
						Description: "Dog refuses to return to handler",
						Scope:       scoring.AutoNQExercise,
					},
				},
			},

			// Exercise 2.4 - Phase 2 Decoy Neutrality (5 points)
			// Rulebook: §7.2 Phase 2 Exercise 2.4
			{
				Kind:      scoring.CriteriaSum,
				Code:      "2.4",
				Name:      "Phase 2 Decoy Neutrality",
				MaxPoints: 5,
				Criteria: []scoring.Criterion{
					{Code: "2.4.a", MaxPoints: 5, Description: "Dog stays task-focused throughout Phase 2; decoys appear irrelevant"},
				},
				AutoTriggers: []scoring.AutoTrigger{
					{
						Code:        "2.4.nq.charge-bite",
						Description: "Position break to charge a decoy, making contact or attempting a bite",
						Scope:       scoring.AutoNQPhase,
					},
				},
			},
		},
	}
}

// l1obPhase3 encodes §7.2 Phase 3: Obstacles & Distance Work.
// Phase total: 35 pts (5 + 5 + 8 + 7 + 5 + 5 = 35).
func l1obPhase3() scoring.PhaseTemplate {
	return scoring.PhaseTemplate{
		Code: "P3",
		Name: "Phase 3: Obstacles & Distance Work",
		Exercises: []scoring.ExerciseTemplate{
			// Exercise 3.1 - Tunnel (5 points)
			// Rulebook: §7.2 Phase 3 Exercise 3.1
			{
				Kind:      scoring.CriteriaSum,
				Code:      "3.1",
				Name:      "Tunnel",
				MaxPoints: 5, // 2+2+1 = 5
				Criteria: []scoring.Criterion{
					{Code: "3.1.a", MaxPoints: 2, Description: "Dog enters tunnel on first cue, no hesitation"},
					{Code: "3.1.b", MaxPoints: 2, Description: "Smooth traversal — no stopping mid-tunnel, no backing out"},
					{Code: "3.1.c", MaxPoints: 1, Description: "Clean exit, dog returns to handler under control"},
				},
				AutoTriggers: []scoring.AutoTrigger{
					{
						Code:        "3.1.nq.refusal",
						Description: "Dog refuses to enter tunnel after multiple cues",
						Scope:       scoring.AutoNQExercise,
					},
					{
						Code:        "3.1.nq.backout",
						Description: "Dog enters tunnel and exits the way it came",
						Scope:       scoring.AutoNQExercise,
					},
				},
			},

			// Exercise 3.2 - Jump (5 points)
			// Rulebook: §7.2 Phase 3 Exercise 3.2
			{
				Kind:      scoring.CriteriaSum,
				Code:      "3.2",
				Name:      "Jump",
				MaxPoints: 5, // 2+2+1 = 5
				Criteria: []scoring.Criterion{
					{Code: "3.2.a", MaxPoints: 2, Description: "Outbound jump clean — no refusal, no knock-down"},
					{Code: "3.2.b", MaxPoints: 2, Description: "Return over jump clean — no refusal, no go-around"},
					{Code: "3.2.c", MaxPoints: 1, Description: "Recovery to heel under control"},
				},
				AutoTriggers: []scoring.AutoTrigger{
					{
						Code:        "3.2.nq.refusal",
						Description: "Dog refuses to attempt jump after multiple cues",
						Scope:       scoring.AutoNQExercise,
					},
					{
						Code:        "3.2.nq.cannot-clear",
						Description: "Dog cannot physically clear the jump after re-cue",
						Scope:       scoring.AutoNQExercise,
					},
				},
			},

			// Exercise 3.3 - Send-Away to Target (8 points)
			// Rulebook: §7.2 Phase 3 Exercise 3.3
			// Note: 8-point scale produces band overlap at 7 (VG/Good).
			// Reset-cue cap (max 4 pts after reset) NOT modeled in v1.
			{
				Kind:      scoring.CriteriaSum,
				Code:      "3.3",
				Name:      "Send-Away to Target",
				MaxPoints: 8, // 2+2+2+1+1 = 8
				Criteria: []scoring.Criterion{
					{Code: "3.3.a", MaxPoints: 2, Description: "Drive on send — dog leaves handler immediately on cue"},
					{Code: "3.3.b", MaxPoints: 2, Description: "Directness of line — dog runs toward target without veering"},
					{Code: "3.3.c", MaxPoints: 2, Description: "Target acquisition — dog mounts/touches platform"},
					{Code: "3.3.d", MaxPoints: 1, Description: "Stop or down on target — clean, prompt response"},
					{Code: "3.3.e", MaxPoints: 1, Description: "Recovery to heel on recall"},
				},
				AutoTriggers: []scoring.AutoTrigger{
					{
						Code:        "3.3.nq.refuses-send",
						Description: "Dog refuses to leave handler",
						Scope:       scoring.AutoNQExercise,
					},
					{
						Code:        "3.3.nq.wrong-target",
						Description: "Dog runs to a non-target object (decoy zone, gate, exit)",
						Scope:       scoring.AutoNQExercise,
					},
					{
						Code:        "3.3.nq.handler-enters-field",
						Description: "Handler enters the field beyond the start point",
						Scope:       scoring.AutoNQExercise,
					},
				},
			},

			// Exercise 3.4 - Directional (Two-Target Discrimination) (7 points)
			// Rulebook: §7.2 Phase 3 Exercise 3.4
			{
				Kind:      scoring.CriteriaSum,
				Code:      "3.4",
				Name:      "Directional (Two-Target Discrimination)",
				MaxPoints: 7, // 3+2+1+1 = 7
				Criteria: []scoring.Criterion{
					{Code: "3.4.a", MaxPoints: 3, Description: "Dog responds to directional cue, commits to correct target"},
					{Code: "3.4.b", MaxPoints: 2, Description: "Drive on send — dog leaves handler immediately, no hesitation"},
					{Code: "3.4.c", MaxPoints: 1, Description: "Directness — dog runs directly to correct marker, no detour"},
					{Code: "3.4.d", MaxPoints: 1, Description: "Recovery (heel or stop/down on marker per declared style)"},
				},
				AutoTriggers: []scoring.AutoTrigger{
					{
						Code:        "3.4.nq.refuses-send",
						Description: "Dog refuses to leave handler",
						Scope:       scoring.AutoNQExercise,
					},
					{
						Code:        "3.4.nq.wrong-target",
						Description: "Dog goes to a non-target object (decoy zone, exit)",
						Scope:       scoring.AutoNQExercise,
					},
					{
						Code:        "3.4.nq.inconsistent-cues",
						Description: "Handler gives multiple inconsistent cues that confuse the dog",
						Scope:       scoring.AutoNQExercise,
					},
				},
			},

			// Exercise 3.5 - Change of Position at Distance (5 points)
			// Rulebook: §7.2 Phase 3 Exercise 3.5
			{
				Kind:      scoring.CriteriaSum,
				Code:      "3.5",
				Name:      "Change of Position at Distance",
				MaxPoints: 5, // 2+1+2 = 5
				Criteria: []scoring.Criterion{
					{Code: "3.5.a", MaxPoints: 2, Description: "Position change executed on first cue"},
					{Code: "3.5.b", MaxPoints: 1, Description: "Clean execution — proper position assumed correctly"},
					{Code: "3.5.c", MaxPoints: 2, Description: "No creeping forward / dog stays in original spot"},
				},
				AutoTriggers: []scoring.AutoTrigger{
					{
						Code:        "3.5.nq.returns-to-handler",
						Description: "Dog leaves position and returns to handler without cue",
						Scope:       scoring.AutoNQExercise,
					},
					{
						Code:        "3.5.nq.refuses-change",
						Description: "Dog refuses position change after multiple cues",
						Scope:       scoring.AutoNQExercise,
					},
				},
			},

			// Exercise 3.6 - Phase 3 Decoy Neutrality (5 points)
			// Rulebook: §7.2 Phase 3 Exercise 3.6
			{
				Kind:      scoring.CriteriaSum,
				Code:      "3.6",
				Name:      "Phase 3 Decoy Neutrality",
				MaxPoints: 5,
				Criteria: []scoring.Criterion{
					{Code: "3.6.a", MaxPoints: 5, Description: "Dog stays task-focused throughout Phase 3; decoys appear irrelevant during all distance work"},
				},
				AutoTriggers: []scoring.AutoTrigger{
					{
						Code:        "3.6.nq.charge-bite",
						Description: "Position break to charge a decoy, making contact or attempting a bite",
						Scope:       scoring.AutoNQPhase,
					},
				},
			},
		},
	}
}

// l1obPhase4 encodes §7.2 Phase 4: Retrieve & Final Composure.
// Phase total: 25 pts (8 + 8 + 4 + 5 = 25).
func l1obPhase4() scoring.PhaseTemplate {
	return scoring.PhaseTemplate{
		Code: "P4",
		Name: "Phase 4: Retrieve & Final Composure",
		Exercises: []scoring.ExerciseTemplate{
			// Exercise 4.1 - Retrieve on the Flat (8 points)
			// Rulebook: §7.2 Phase 4 Exercise 4.1
			{
				Kind:      scoring.CriteriaSum,
				Code:      "4.1",
				Name:      "Retrieve on the Flat",
				MaxPoints: 8, // 2+1+2+1+1+1 = 8
				Criteria: []scoring.Criterion{
					{Code: "4.1.a", MaxPoints: 2, Description: "Dog remains in heel-sit during throw, no breaking"},
					{Code: "4.1.b", MaxPoints: 1, Description: "Send on first cue, immediate departure"},
					{Code: "4.1.c", MaxPoints: 2, Description: "Direct line to object, clean pickup"},
					{Code: "4.1.d", MaxPoints: 1, Description: "Direct return, clean carry (no dropping, no chewing)"},
					{Code: "4.1.e", MaxPoints: 1, Description: "Clean delivery per declared style"},
					{Code: "4.1.f", MaxPoints: 1, Description: "Recovery to heel"},
				},
				AutoTriggers: []scoring.AutoTrigger{
					{
						Code:        "4.1.nq.destroys-object",
						Description: "Dog destroys object",
						Scope:       scoring.AutoNQExercise,
					},
					{
						Code:        "4.1.nq.refuses-retrieve",
						Description: "Dog refuses to retrieve after multiple cues",
						Scope:       scoring.AutoNQExercise,
					},
					{
						Code:        "4.1.nq.wrong-target",
						Description: "Dog runs to a non-target object (decoy zone, etc.)",
						Scope:       scoring.AutoNQExercise,
					},
				},
			},

			// Exercise 4.2 - Retrieve over Jump (8 points)
			// Rulebook: §7.2 Phase 4 Exercise 4.2
			{
				Kind:      scoring.CriteriaSum,
				Code:      "4.2",
				Name:      "Retrieve over Jump",
				MaxPoints: 8, // 2+1+1+1+1+2 = 8
				Criteria: []scoring.Criterion{
					{Code: "4.2.a", MaxPoints: 2, Description: "Dog remains in heel-sit during throw and over jump, no breaking"},
					{Code: "4.2.b", MaxPoints: 1, Description: "Send on first cue, immediate departure"},
					{Code: "4.2.c", MaxPoints: 1, Description: "Outbound jump clean (no refusal, no go-around)"},
					{Code: "4.2.d", MaxPoints: 1, Description: "Direct line to object, clean pickup"},
					{Code: "4.2.e", MaxPoints: 1, Description: "Return jump clean (no refusal, no go-around)"},
					{Code: "4.2.f", MaxPoints: 2, Description: "Clean delivery and recovery to heel"},
				},
				AutoTriggers: []scoring.AutoTrigger{
					{
						Code:        "4.2.nq.refuses-outbound",
						Description: "Dog refuses outbound jump",
						Scope:       scoring.AutoNQExercise,
					},
					{
						Code:        "4.2.nq.around-on-return",
						Description: "Dog cannot complete return jump and goes around with object",
						Scope:       scoring.AutoNQExercise,
					},
					{
						Code:        "4.2.nq.destroys-object",
						Description: "Dog destroys object",
						Scope:       scoring.AutoNQExercise,
					},
					{
						Code:        "4.2.nq.refuses-retrieve",
						Description: "Dog refuses to retrieve",
						Scope:       scoring.AutoNQExercise,
					},
				},
			},

			// Exercise 4.3 - Final Stand for Examination & Report Off (4 points)
			// Rulebook: §7.2 Phase 4 Exercise 4.3
			{
				Kind:      scoring.CriteriaSum,
				Code:      "4.3",
				Name:      "Final Stand for Examination & Report Off",
				MaxPoints: 4, // 1+2+1 = 4
				Criteria: []scoring.Criterion{
					{Code: "4.3.a", MaxPoints: 1, Description: "Dog holds stand during judge approach (no breaking)"},
					{Code: "4.3.b", MaxPoints: 2, Description: "Dog accepts close-proximity presence without flinching, growling, or moving"},
					{Code: "4.3.c", MaxPoints: 1, Description: "Clean report-off and heel off the field"},
				},
				AutoTriggers: []scoring.AutoTrigger{
					{
						Code:        "4.3.nq.aggression-judge",
						Description: "Dog growls or shows aggression during judge approach",
						Scope:       scoring.AutoNQTrial, // §4.5: aggression toward judge
					},
					{
						Code:        "4.3.nq.retreats",
						Description: "Dog leaves position to retreat from judge",
						Scope:       scoring.AutoNQExercise,
					},
				},
			},

			// Exercise 4.4 - Phase 4 Decoy Neutrality (5 points)
			// Rulebook: §7.2 Phase 4 Exercise 4.4
			{
				Kind:      scoring.CriteriaSum,
				Code:      "4.4",
				Name:      "Phase 4 Decoy Neutrality",
				MaxPoints: 5,
				Criteria: []scoring.Criterion{
					{Code: "4.4.a", MaxPoints: 5, Description: "Dog stays task-focused throughout Phase 4; decoys appear irrelevant"},
				},
				AutoTriggers: []scoring.AutoTrigger{
					{
						Code:        "4.4.nq.charge-bite",
						Description: "Position break to charge a decoy, making contact or attempting a bite",
						Scope:       scoring.AutoNQPhase,
					},
				},
			},
		},
	}
}
