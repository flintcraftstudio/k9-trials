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
	HasInput bool // any criterion/penalty/trigger logged against this code
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
	for _, cs := range inputs.CriterionScores {
		hasInputByCode[cs.ExerciseCode] = true
	}
	for _, po := range inputs.PenaltyOccurrences {
		hasInputByCode[po.ExerciseCode] = true
	}
	for _, at := range inputs.AutoTriggers {
		hasInputByCode[at.ExerciseCode] = true
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
			out = append(out, flattenedExercise{
				Num:      n,
				Code:     ex.Code,
				Name:     ex.Name,
				MaxPts:   ex.MaxPoints,
				Result:   resultsByCode[ex.Code],
				HasInput: hasInputByCode[ex.Code],
			})
		}
	}
	return out
}

// toExercises produces the B3-O scoresheet exercise list. The "active"
// exercise is the first one with no inputs (judge's natural cursor); if
// every exercise has inputs, the active row stays on the first.
func toExercises(flat []flattenedExercise) ([]judge.Exercise, int) {
	out := make([]judge.Exercise, len(flat))
	activeIdx := 0
	pickedActive := false
	for i, fx := range flat {
		out[i] = judge.Exercise{
			Num:    fx.Num,
			Name:   fx.Name,
			Max:    float64(fx.MaxPts),
			Score:  float64(fx.Result.Points),
			Scored: fx.HasInput,
		}
		if !pickedActive && !fx.HasInput {
			activeIdx = i
			pickedActive = true
		}
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
