package handler

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/flintcraftstudio/k9-trials/internal/db"
	"github.com/flintcraftstudio/k9-trials/internal/session"
	"github.com/flintcraftstudio/k9-trials/internal/store"
	"github.com/flintcraftstudio/k9-trials/internal/view/admin"
)

// AdminDashboard serves GET /admin — the operations landing page (D1).
func AdminDashboard(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		events, err := st.ListEvents(r.Context())
		if err != nil {
			slog.Error("admin dashboard events", "err", err)
			http.Error(w, "admin unavailable", http.StatusInternalServerError)
			return
		}
		pending, err := st.CountAllPendingRegistrations(r.Context())
		if err != nil {
			slog.Error("admin pending count", "err", err)
			http.Error(w, "admin unavailable", http.StatusInternalServerError)
			return
		}
		openCh, err := st.CountAllOpenChallenges(r.Context())
		if err != nil {
			slog.Error("admin challenge count", "err", err)
			http.Error(w, "admin unavailable", http.StatusInternalServerError)
			return
		}
		renderPublic(w, r, admin.DashboardPage(toAdminDashboardVD(r.Context(), st, events, int(pending), int(openCh))))
	}
}

// AdminEvents serves GET /admin/events — the events list (D2), with status
// filter chips. htmx filter requests receive only the table fragment.
func AdminEvents(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		events, err := st.ListEvents(r.Context())
		if err != nil {
			slog.Error("admin events", "err", err)
			http.Error(w, "admin unavailable", http.StatusInternalServerError)
			return
		}
		filter := r.URL.Query().Get("status")
		if !validEventFilter(filter) {
			filter = ""
		}
		q := r.URL.Query().Get("q")
		data := toAdminEventsVD(r.Context(), st, events, filter, q)
		if r.Header.Get("HX-Request") == "true" {
			renderPublic(w, r, admin.EventsResults(data))
			return
		}
		renderPublic(w, r, admin.EventsListPage(data))
	}
}

// AdminEventsNew serves GET /admin/events/new — the empty create form (D3).
func AdminEventsNew() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderPublic(w, r, admin.EventsFormPage(admin.EventFormViewData{Status: "draft"}))
	}
}

// AdminEventsCreate serves POST /admin/events — validates and inserts a new
// event, then redirects to its editor. Errors re-render the form fragment.
func AdminEventsCreate(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		in, vd, ok := parseEventForm(r, admin.EventFormViewData{})
		if !ok {
			renderPublic(w, r, admin.EventForm(vd))
			return
		}

		available, err := st.EventSlugAvailable(r.Context(), in.Slug, 0)
		if err != nil {
			slog.Error("event slug check", "err", err)
			vd.Err = "Something went wrong. Please try again."
			renderPublic(w, r, admin.EventForm(vd))
			return
		}
		if !available {
			vd.Err = "That slug is already taken. Pick another."
			renderPublic(w, r, admin.EventForm(vd))
			return
		}

		u := session.FromContext(r.Context())
		event, err := st.CreateEvent(r.Context(), in, u.ID)
		if err != nil {
			if isUniqueViolation(err) {
				vd.Err = "That slug is already taken. Pick another."
				renderPublic(w, r, admin.EventForm(vd))
				return
			}
			slog.Error("create event", "err", err)
			vd.Err = "Something went wrong. Please try again."
			renderPublic(w, r, admin.EventForm(vd))
			return
		}
		hxRedirect(w, r, "/admin/events/"+strconv.FormatInt(event.ID, 10)+"/edit")
	}
}

// AdminEventsEdit serves GET /admin/events/{id}/edit — the editor (D3).
func AdminEventsEdit(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		event, ok := loadAdminEvent(w, r, st)
		if !ok {
			return
		}
		renderPublic(w, r, admin.EventsFormPage(editEventVD(r.Context(), st, event)))
	}
}

// AdminEventsUpdate serves POST /admin/events/{id} — saves metadata edits
// (the slug is immutable) and re-renders the form with a confirmation.
func AdminEventsUpdate(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		event, ok := loadAdminEvent(w, r, st)
		if !ok {
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		base := editEventVD(r.Context(), st, event)
		in, vd, ok := parseEventForm(r, base)
		if !ok {
			renderPublic(w, r, admin.EventForm(vd))
			return
		}
		// The slug does not change on update; keep the stored one.
		in.Slug = event.Slug
		if _, err := st.UpdateEvent(r.Context(), event.ID, in); err != nil {
			slog.Error("update event", "event", event.ID, "err", err)
			vd.Err = "Something went wrong. Please try again."
			renderPublic(w, r, admin.EventForm(vd))
			return
		}
		// Opening registration (draft/closed → published) fires the notify-me
		// hook for subscribers (Q4 / R1c).
		if event.Status != "published" && in.Status == "published" {
			notifyEventSubscribers(r.Context(), st, event.ID, event.Name)
		}
		vd.Saved = true
		vd.Err = ""
		renderPublic(w, r, admin.EventForm(vd))
	}
}

// notifyEventSubscribers logs the recipients who asked to be notified when an
// event opens registration, then marks them notified so a later re-publish
// does not re-notify. Email delivery is not wired (the mail client only
// targets the contact-form recipient), mirroring the D6 notify-judges stub.
func notifyEventSubscribers(ctx context.Context, st *store.Store, eventID int64, eventName string) {
	subs, err := st.ListEventSubscribers(ctx, eventID)
	if err != nil {
		slog.Error("list event subscribers", "event", eventID, "err", err)
		return
	}
	if len(subs) == 0 {
		return
	}
	emails := make([]string, 0, len(subs))
	for _, s := range subs {
		emails = append(emails, s.Email)
	}
	slog.Info("event registration opened — notifying subscribers (delivery pending mail setup)",
		"event", eventID, "name", eventName, "count", len(emails), "recipients", emails)
	if err := st.MarkEventSubscribersNotified(ctx, eventID); err != nil {
		slog.Error("mark subscribers notified", "event", eventID, "err", err)
	}
}

// AdminEventsArchive serves POST /admin/events/{id}/archive — the D3 archive
// lifecycle action. Archiving files the event away (hidden from public lists,
// excluded from the default admin view) while retaining its row and history.
func AdminEventsArchive(st *store.Store) http.HandlerFunc {
	return setEventStatusHandler(st, "archived")
}

// AdminEventsRestore serves POST /admin/events/{id}/unarchive — returns an
// archived event to draft so it can be edited and re-published.
func AdminEventsRestore(st *store.Store) http.HandlerFunc {
	return setEventStatusHandler(st, "draft")
}

// setEventStatusHandler builds a handler that transitions an event to the
// given status and reloads its editor.
func setEventStatusHandler(st *store.Store, status string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		event, ok := loadAdminEvent(w, r, st)
		if !ok {
			return
		}
		if _, err := st.SetEventStatus(r.Context(), event.ID, status); err != nil {
			slog.Error("set event status", "event", event.ID, "status", status, "err", err)
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}
		hxRedirect(w, r, "/admin/events/"+strconv.FormatInt(event.ID, 10)+"/edit")
	}
}

// AdminEventsSlugCheck serves GET /admin/events/slug-check — the live
// availability probe for the create form.
func AdminEventsSlugCheck(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("slug")))
		if slug == "" {
			renderPublic(w, r, admin.SlugStatus("", false, false))
			return
		}
		if !handlePattern.MatchString(slug) {
			renderPublic(w, r, admin.SlugStatus(slug, false, true))
			return
		}
		available, err := st.EventSlugAvailable(r.Context(), slug, 0)
		if err != nil {
			slog.Error("slug check", "err", err)
			renderPublic(w, r, admin.SlugStatus(slug, false, false))
			return
		}
		renderPublic(w, r, admin.SlugStatus(slug, available, true))
	}
}

// AdminTrials serves GET /admin/events/{id}/trials — the trials list (D4).
func AdminTrials(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		event, ok := loadAdminEvent(w, r, st)
		if !ok {
			return
		}
		trials, err := st.TrialsByEvent(r.Context(), event.ID)
		if err != nil {
			slog.Error("admin trials", "event", event.ID, "err", err)
			http.Error(w, "admin unavailable", http.StatusInternalServerError)
			return
		}
		renderPublic(w, r, admin.TrialsListPage(toAdminTrialsVD(r.Context(), st, event, trials)))
	}
}

// AdminTrialsNew serves GET /admin/events/{id}/trials/new — the create form.
func AdminTrialsNew(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		event, ok := loadAdminEvent(w, r, st)
		if !ok {
			return
		}
		renderPublic(w, r, admin.TrialsFormPage(admin.TrialFormViewData{
			EventID:         event.ID,
			EventName:       event.Name,
			Discipline:      "OB",
			Level:           "1",
			TemplateVersion: "2026.1",
		}))
	}
}

// AdminTrialsCreate serves POST /admin/events/{id}/trials — inserts a trial
// and redirects to the trials list. Errors re-render the form fragment.
func AdminTrialsCreate(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		event, ok := loadAdminEvent(w, r, st)
		if !ok {
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		in, vd, ok := parseTrialForm(r, event.ID, event.Name)
		if !ok {
			renderPublic(w, r, admin.TrialsFormPage(vd))
			return
		}
		if _, err := st.CreateTrial(r.Context(), event.ID, in); err != nil {
			if isUniqueViolation(err) {
				vd.Err = "A trial with that discipline, level, and date already exists."
				renderPublic(w, r, admin.TrialsFormPage(vd))
				return
			}
			slog.Error("create trial", "event", event.ID, "err", err)
			vd.Err = "Something went wrong. Please try again."
			renderPublic(w, r, admin.TrialsFormPage(vd))
			return
		}
		hxRedirect(w, r, "/admin/events/"+strconv.FormatInt(event.ID, 10)+"/trials")
	}
}

// AdminTrialsDelete serves POST /admin/events/{id}/trials/{tid}/delete.
func AdminTrialsDelete(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		eventID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || eventID <= 0 {
			http.NotFound(w, r)
			return
		}
		trialID, err := strconv.ParseInt(r.PathValue("tid"), 10, 64)
		if err != nil || trialID <= 0 {
			http.NotFound(w, r)
			return
		}
		// Confirm the trial belongs to the event in the URL before deleting.
		trial, err := st.GetTrial(r.Context(), trialID)
		if errors.Is(err, sql.ErrNoRows) || (err == nil && trial.EventID != eventID) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			slog.Error("trial delete load", "trial", trialID, "err", err)
			http.Error(w, "admin unavailable", http.StatusInternalServerError)
			return
		}
		if err := st.DeleteTrial(r.Context(), trialID); err != nil {
			slog.Error("delete trial", "trial", trialID, "err", err)
			http.Error(w, "admin unavailable", http.StatusInternalServerError)
			return
		}
		hxRedirect(w, r, "/admin/events/"+strconv.FormatInt(eventID, 10)+"/trials")
	}
}

// loadAdminEvent parses the {id} path segment and loads the event. Writes a
// 404 and returns ok=false on a missing id or row.
func loadAdminEvent(w http.ResponseWriter, r *http.Request, st *store.Store) (db.Event, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.NotFound(w, r)
		return db.Event{}, false
	}
	event, err := st.GetEvent(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return db.Event{}, false
	}
	if err != nil {
		slog.Error("admin event load", "event", id, "err", err)
		http.Error(w, "admin unavailable", http.StatusInternalServerError)
		return db.Event{}, false
	}
	return event, true
}

// parseEventForm reads and validates the event form fields. base carries
// the IsEdit / EventID / glance context. Returns ok=false with a populated
// view (values + error) on a validation failure.
func parseEventForm(r *http.Request, base admin.EventFormViewData) (store.EventInput, admin.EventFormViewData, bool) {
	vd := base
	name := strings.TrimSpace(r.FormValue("name"))
	location := strings.TrimSpace(r.FormValue("location"))
	startStr := strings.TrimSpace(r.FormValue("start_date"))
	endStr := strings.TrimSpace(r.FormValue("end_date"))
	status := strings.TrimSpace(r.FormValue("status"))

	vd.Name = name
	vd.Location = location
	vd.StartDate = startStr
	vd.EndDate = endStr
	vd.Status = status

	slug := vd.Slug
	if !vd.IsEdit {
		slug = strings.ToLower(strings.TrimSpace(r.FormValue("slug")))
		if slug == "" {
			slug = slugify(name)
		}
		vd.Slug = slug
	}

	if name == "" {
		vd.Err = "Name is required."
		return store.EventInput{}, vd, false
	}
	if !vd.IsEdit && (slug == "" || !handlePattern.MatchString(slug)) {
		vd.Err = "Slug can use only lowercase letters, digits, and hyphens."
		return store.EventInput{}, vd, false
	}
	if !validEventStatus(status) {
		vd.Err = "Pick a valid status."
		return store.EventInput{}, vd, false
	}
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		vd.Err = "Start date must be a valid date."
		return store.EventInput{}, vd, false
	}
	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		vd.Err = "End date must be a valid date."
		return store.EventInput{}, vd, false
	}
	if end.Before(start) {
		vd.Err = "End date cannot be before the start date."
		return store.EventInput{}, vd, false
	}

	return store.EventInput{
		Slug:      slug,
		Name:      name,
		Location:  location,
		StartDate: start,
		EndDate:   end,
		Status:    status,
	}, vd, true
}

// parseTrialForm reads and validates the create-trial form.
func parseTrialForm(r *http.Request, eventID int64, eventName string) (store.TrialInput, admin.TrialFormViewData, bool) {
	discipline := strings.TrimSpace(r.FormValue("discipline"))
	levelStr := strings.TrimSpace(r.FormValue("level"))
	dateStr := strings.TrimSpace(r.FormValue("trial_date"))
	version := strings.TrimSpace(r.FormValue("template_version"))

	vd := admin.TrialFormViewData{
		EventID:         eventID,
		EventName:       eventName,
		Discipline:      discipline,
		Level:           levelStr,
		Date:            dateStr,
		TemplateVersion: version,
	}

	if !validDiscipline(discipline) {
		vd.Err = "Pick a valid discipline."
		return store.TrialInput{}, vd, false
	}
	level, err := strconv.ParseInt(levelStr, 10, 64)
	if err != nil || level < 1 || level > 3 {
		vd.Err = "Level must be 1, 2, or 3."
		return store.TrialInput{}, vd, false
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		vd.Err = "Date must be a valid date."
		return store.TrialInput{}, vd, false
	}
	if version == "" {
		vd.Err = "Template version is required."
		return store.TrialInput{}, vd, false
	}

	return store.TrialInput{
		Discipline:      discipline,
		Level:           level,
		TrialDate:       date,
		TemplateVersion: version,
		Status:          "pending",
	}, vd, true
}

// validEventStatus reports whether status is a recognized event status.
func validEventStatus(status string) bool {
	switch status {
	case "draft", "published", "closed", "archived":
		return true
	}
	return false
}
