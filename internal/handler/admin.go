package handler

import (
	"log/slog"
	"net/http"

	"github.com/flintcraftstudio/k9-trials/internal/view/admin"
)

// AdminDashboard renders the admin landing page.
func AdminDashboard() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := admin.DashboardPage().Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// AdminEvents lists every event with admin status.
func AdminEvents() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := admin.EventsListPage().Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// AdminEventsNew renders the create-event form.
func AdminEventsNew() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := admin.EventsFormPage("").Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// AdminEventsEdit renders the edit-event form.
func AdminEventsEdit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := admin.EventsFormPage(id).Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// AdminTrials lists trials within an event.
func AdminTrials() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := admin.TrialsListPage(id).Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// AdminTrialsNew renders the create-trial form for a given event.
func AdminTrialsNew() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := admin.TrialsFormPage(id).Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// AdminRegistrations renders the registration review queue for an event.
func AdminRegistrations() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := admin.RegistrationsPage(id).Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// AdminAssignments renders the judge-to-trial assignment screen.
func AdminAssignments() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := admin.AssignmentsPage(id).Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// AdminChallenges renders the cross-event challenge review queue.
func AdminChallenges() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := admin.ChallengesPage().Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// AdminUsers renders the user/role management page.
func AdminUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := admin.UsersPage().Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}
