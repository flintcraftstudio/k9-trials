package handler

import (
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/flintcraftstudio/k9-trials/internal/view/judge"
)

func renderJudge(w http.ResponseWriter, r *http.Request, c templ.Component) {
	if err := c.Render(r.Context(), w); err != nil {
		slog.Error("render error", "err", err)
	}
}

// JudgeQueue renders the run queue (B1).
func JudgeQueue() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderJudge(w, r, judge.QueuePage(judge.QueueData()))
	}
}

// JudgeGate renders the identity gate (B2) for a given entry.
func JudgeGate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, ok := judge.GateData(r.PathValue("id"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		renderJudge(w, r, judge.GatePage(data))
	}
}

// JudgeScore renders the active scoresheet (B3-O or B3-D, picked by discipline).
func JudgeScore() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, ok := judge.ScoreData(r.PathValue("id"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		if data.Discipline == "DT" {
			renderJudge(w, r, judge.DetectionScorePage(data))
			return
		}
		renderJudge(w, r, judge.ObedienceScorePage(data))
	}
}

// JudgeReview renders run review (B4).
func JudgeReview() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, ok := judge.ReviewData(r.PathValue("id"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		renderJudge(w, r, judge.ReviewPage(data))
	}
}

// JudgeSubmit renders submit confirmation (B5).
func JudgeSubmit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, ok := judge.SubmitData(r.PathValue("id"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		renderJudge(w, r, judge.SubmitPage(data))
	}
}

// JudgeLocked renders the read-only locked run (B6).
func JudgeLocked() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, ok := judge.LockedData(r.PathValue("id"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		renderJudge(w, r, judge.LockedPage(data))
	}
}
