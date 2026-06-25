package handler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/flintcraftstudio/k9-trials/internal/db"
	"github.com/flintcraftstudio/k9-trials/internal/scoring"
	"github.com/flintcraftstudio/k9-trials/internal/view/judge"
)

// statusToView maps the DB entry.status enum to the view-side Status enum.
// The DB has only three life-cycle states (registered, scoring, finalized);
// the view has additional pocket/scratch states that aren't yet persisted
// — they collapse onto the closest DB-backed peer.
func statusToView(s string) judge.Status {
	switch s {
	case "registered":
		return judge.StatusPending
	case "scoring":
		return judge.StatusInProg
	case "finalized":
		return judge.StatusDelivered
	}
	return judge.StatusPending
}

// dogInitial returns the first uppercase letter of dog's name, defaulting
// to '?' for empty names. The B1 queue uses this for the avatar disc.
func dogInitial(name string) string {
	for _, r := range name {
		return strings.ToUpper(string(r))
	}
	return "?"
}

// dogVariant picks a deterministic avatar gradient (1..5) for a dog name.
// Stable across renders so the queue chips don't shimmer on refresh.
func dogVariant(name string) int {
	var sum int
	for _, r := range name {
		sum += int(r)
	}
	return (sum % 5) + 1
}

// handlerShort renders "Jane Marsh" as "J. Marsh" for the dense queue rows.
// Falls back to the original string if it can't be split.
func handlerShort(full string) string {
	parts := strings.Fields(full)
	if len(parts) < 2 {
		return full
	}
	first := parts[0]
	last := parts[len(parts)-1]
	if len(first) == 0 {
		return last
	}
	return string(first[0]) + ". " + last
}

// judgeInitials returns up to two uppercase initials from an email or
// name. "h.vance@example.com" → "HV", "Logan Williams" → "LW".
func judgeInitials(nameOrEmail string) string {
	if nameOrEmail == "" {
		return "??"
	}
	// Strip domain if email.
	local := nameOrEmail
	if i := strings.IndexByte(local, '@'); i >= 0 {
		local = local[:i]
	}
	// Split on common separators.
	local = strings.ReplaceAll(local, ".", " ")
	local = strings.ReplaceAll(local, "_", " ")
	local = strings.ReplaceAll(local, "-", " ")
	parts := strings.Fields(local)
	switch len(parts) {
	case 0:
		return "??"
	case 1:
		s := strings.ToUpper(parts[0])
		if len(s) >= 2 {
			return s[:2]
		}
		return s
	default:
		return strings.ToUpper(string(parts[0][0])) + strings.ToUpper(string(parts[len(parts)-1][0]))
	}
}

// judgeName returns the human-readable judge name. For now this is just
// the local part of the email with separators turned to spaces and title-cased.
// When real `users.display_name` lands, swap this helper.
func judgeName(email string) string {
	if email == "" {
		return ""
	}
	local := email
	if i := strings.IndexByte(local, '@'); i >= 0 {
		local = local[:i]
	}
	local = strings.ReplaceAll(local, ".", " ")
	local = strings.ReplaceAll(local, "_", " ")
	local = strings.ReplaceAll(local, "-", " ")
	parts := strings.Fields(local)
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
	}
	return strings.Join(parts, " ")
}

// scheduledTime formats the entry's CreatedAt into "HH.MM" for queue display.
// The schema doesn't carry a separate scheduled time yet — created_at is a
// stable stand-in.
func scheduledTime(t time.Time) string {
	return t.Format("15.04")
}

// toTrial maps DB rows + the judge identity into the view's Trial struct.
func toTrial(trial db.Trial, event db.Event, judgeEmail string) judge.Trial {
	class := ""
	switch scoring.Level(trial.Level) {
	case scoring.LevelOne:
		class = "Level 1"
	case scoring.LevelTwo:
		class = "Level 2"
	case scoring.LevelThree:
		class = "Level 3"
	}
	return judge.Trial{
		Name:       event.Name,
		Class:      class,
		Discipline: judge.Discipline(trial.Discipline),
		JudgeName:  judgeName(judgeEmail),
		JudgeInits: judgeInitials(judgeEmail),
	}
}

// toRun maps a DB entry into the view's Run struct. `scoreLabel` is the
// pre-formatted score string for the queue row (e.g. "scored 84") and is
// empty when the entry has no evaluated total to show.
func toRun(entry db.Entry, scoreLabel string) judge.Run {
	return judge.Run{
		ID:         strconv.FormatInt(entry.ID, 10),
		Number:     int(entry.EntryNumber),
		DogName:    entry.DogName,
		DogInit:    dogInitial(entry.DogName),
		DogVariant: dogVariant(entry.DogName),
		HandlerSh:  handlerShort(entry.HandlerName),
		Breed:      entry.DogBreed,
		K9ID:       "",
		Scheduled:  scheduledTime(entry.CreatedAt),
		Status:     statusToView(entry.Status),
		Score:      scoreLabel,
	}
}

// flattenExercises walks the template's phases and returns each
// (numbered) exercise row paired with its evaluated result. The row
// number is a flat 1..N counter so the UI can show "Exercise 3 · Stand
// for exam" without exposing phase structure.
type flattenedExercise struct {
	Num      int
	Code     string
	Name     string
	MaxPts   scoring.Points
	Result   scoring.ExerciseResult
	HasInput bool                 // any criterion/penalty/trigger logged against this code
	Criteria []flattenedCriterion // per-criterion rows (CriteriaSum exercises)
	Triggers []flattenedTrigger   // auto-NQ triggers with fired state
	NQ       bool                 // score forced to 0 by a fired trigger (own or phase)
	NQReason string
}

// flattenedCriterion is one criterion of a CriteriaSum exercise with its
// current judge-entered value projected from the append-only inputs.
type flattenedCriterion struct {
	Code   string
	Desc   string
	Max    scoring.Points
	Points scoring.Points
	Scored bool
}

// flattenedTrigger is one auto-NQ trigger with whether it has fired.
type flattenedTrigger struct {
	Code  string
	Desc  string
	Scope string // "exercise" | "phase" | "trial"
	Fired bool
}

// scopeStr renders an AutoTriggerScope as a short view-facing label.
func scopeStr(s scoring.AutoTriggerScope) string {
	switch s {
	case scoring.AutoNQPhase:
		return "phase"
	case scoring.AutoNQTrial:
		return "trial"
	default:
		return "exercise"
	}
}

func flattenExercises(
	sheet scoring.ConcreteScoresheet,
	result scoring.ScoresheetResult,
	inputs scoring.ScoresheetInputs,
) []flattenedExercise {
	resultsByCode := make(map[string]scoring.ExerciseResult, len(result.PerExercise))
	for _, r := range result.PerExercise {
		resultsByCode[r.ExerciseCode] = r
	}
	hasInputByCode := make(map[string]bool)
	// pointsByCriterion is the latest-write-wins projection keyed by
	// (exerciseCode, criterionCode); the store loader already collapses
	// the append-only rows, so a present key means the criterion was scored.
	pointsByCriterion := make(map[string]map[string]scoring.Points)
	for _, cs := range inputs.CriterionScores {
		hasInputByCode[cs.ExerciseCode] = true
		if pointsByCriterion[cs.ExerciseCode] == nil {
			pointsByCriterion[cs.ExerciseCode] = make(map[string]scoring.Points)
		}
		pointsByCriterion[cs.ExerciseCode][cs.CriterionCode] = cs.Points
	}
	for _, po := range inputs.PenaltyOccurrences {
		hasInputByCode[po.ExerciseCode] = true
	}
	// firedTriggers[exerciseCode][triggerCode] = true for any logged firing.
	firedTriggers := make(map[string]map[string]bool)
	for _, at := range inputs.AutoTriggers {
		hasInputByCode[at.ExerciseCode] = true
		if firedTriggers[at.ExerciseCode] == nil {
			firedTriggers[at.ExerciseCode] = make(map[string]bool)
		}
		firedTriggers[at.ExerciseCode][at.TriggerCode] = true
	}

	// A phase is NQ'd when any phase-scoped trigger inside it has fired;
	// the engine then zeroes every exercise in that phase.
	phaseNQ := make(map[string]bool)
	for _, ph := range sheet.Phases {
		for _, ex := range ph.Exercises {
			for _, at := range ex.AutoTriggers {
				if at.Scope == scoring.AutoNQPhase && firedTriggers[ex.Code][at.Code] {
					phaseNQ[ph.Code] = true
				}
			}
		}
	}

	out := make([]flattenedExercise, 0)
	n := 0
	for _, ph := range sheet.Phases {
		for _, ex := range ph.Exercises {
			// Aggregate components are folded into their parent and shouldn't
			// be shown as standalone rows on the scoresheet.
			if ex.IsAggregateComponent {
				continue
			}
			n++
			crits := make([]flattenedCriterion, 0, len(ex.Criteria))
			for _, c := range ex.Criteria {
				pts, scored := pointsByCriterion[ex.Code][c.Code]
				crits = append(crits, flattenedCriterion{
					Code:   c.Code,
					Desc:   c.Description,
					Max:    c.MaxPoints,
					Points: pts,
					Scored: scored,
				})
			}
			trigs := make([]flattenedTrigger, 0, len(ex.AutoTriggers))
			exNQ := false
			reason := ""
			for _, at := range ex.AutoTriggers {
				fired := firedTriggers[ex.Code][at.Code]
				trigs = append(trigs, flattenedTrigger{
					Code:  at.Code,
					Desc:  at.Description,
					Scope: scopeStr(at.Scope),
					Fired: fired,
				})
				if fired && at.Scope == scoring.AutoNQExercise {
					exNQ = true
					reason = "NQ — " + at.Description
				}
			}
			if phaseNQ[ph.Code] {
				exNQ = true
				if reason == "" {
					reason = "Phase NQ"
				}
			}
			out = append(out, flattenedExercise{
				Num:      n,
				Code:     ex.Code,
				Name:     ex.Name,
				MaxPts:   ex.MaxPoints,
				Result:   resultsByCode[ex.Code],
				HasInput: hasInputByCode[ex.Code],
				Criteria: crits,
				Triggers: trigs,
				NQ:       exNQ,
				NQReason: reason,
			})
		}
	}
	return out
}

// toExercises produces the B3-O scoresheet exercise list. When activeCode
// names an exercise, that one is the active cursor (the judge navigated to
// it or just scored a criterion on it). Otherwise the active exercise is
// the first one with no inputs; if every exercise has inputs, it stays on
// the first.
func toExercises(flat []flattenedExercise, activeCode string) ([]judge.Exercise, int) {
	out := make([]judge.Exercise, len(flat))
	activeIdx := 0
	pickedActive := false
	requestedIdx := -1
	for i, fx := range flat {
		crits := make([]judge.Criterion, 0, len(fx.Criteria))
		for _, c := range fx.Criteria {
			crits = append(crits, judge.Criterion{
				Code:   c.Code,
				Desc:   c.Desc,
				Max:    int(c.Max),
				Points: int(c.Points),
				Scored: c.Scored,
			})
		}
		trigs := make([]judge.Trigger, 0, len(fx.Triggers))
		for _, t := range fx.Triggers {
			trigs = append(trigs, judge.Trigger{
				Code:  t.Code,
				Desc:  t.Desc,
				Scope: t.Scope,
				Fired: t.Fired,
			})
		}
		out[i] = judge.Exercise{
			Num:      fx.Num,
			Code:     fx.Code,
			Name:     fx.Name,
			Max:      float64(fx.MaxPts),
			Score:    float64(fx.Result.Points),
			Scored:   fx.HasInput,
			Criteria: crits,
			Triggers: trigs,
			NQ:       fx.NQ,
			NQReason: fx.NQReason,
		}
		if activeCode != "" && fx.Code == activeCode {
			requestedIdx = i
		}
		if !pickedActive && !fx.HasInput {
			activeIdx = i
			pickedActive = true
		}
	}
	if requestedIdx >= 0 {
		activeIdx = requestedIdx
	}
	if len(out) > 0 {
		out[activeIdx].Active = true
	}
	return out, activeIdx
}

// toReviewExercises is the B4/B6 summary list. Same flatten as B3 but
// with the Flagged + Note channels for "no score entered" callouts.
func toReviewExercises(flat []flattenedExercise) []judge.ReviewExercise {
	out := make([]judge.ReviewExercise, len(flat))
	for i, fx := range flat {
		re := judge.ReviewExercise{
			Num:    fx.Num,
			Name:   fx.Name,
			Score:  float64(fx.Result.Points),
			Max:    float64(fx.MaxPts),
			Scored: fx.HasInput,
		}
		if !fx.HasInput {
			re.Flagged = true
			re.Note = "no score entered"
		}
		out[i] = re
	}
	return out
}

// unscoredSummary returns (count, "first unscored exercise label") for
// the B4 banner.
func unscoredSummary(flat []flattenedExercise) (int, string) {
	count := 0
	first := ""
	for _, fx := range flat {
		if fx.HasInput {
			continue
		}
		count++
		if first == "" {
			first = fmt.Sprintf("Ex %d · %s", fx.Num, fx.Name)
		}
	}
	return count, first
}

// scoreLabel formats "scored N" for finalized entries with a non-nil
// result, empty otherwise — the B1 queue uses an empty string to mean
// "no number to show yet."
func scoreLabel(entry db.Entry, total scoring.Points) string {
	if entry.Status == "finalized" {
		return fmt.Sprintf("scored %d", int(total))
	}
	return ""
}

// qualifyingThreshold is the pass cutoff in points: percent threshold ×
// max points, rounded per §3.2.
func qualifyingThreshold(tpl scoring.ScoresheetTemplate, sheet scoring.ConcreteScoresheet) float64 {
	return float64(scoring.RoundPoints(
		float64(sheet.MaxPoints) * float64(tpl.PassThresholdPct) / 100.0,
	))
}
