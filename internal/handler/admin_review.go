package handler

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/flintcraftstudio/k9-trials/internal/session"
	"github.com/flintcraftstudio/k9-trials/internal/store"
	"github.com/flintcraftstudio/k9-trials/internal/view/admin"
)

// AdminRegistrations serves GET /admin/events/{id}/registrations — the
// registration review screen (D5).
func AdminRegistrations(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		event, ok := loadAdminEvent(w, r, st)
		if !ok {
			return
		}
		rows, err := st.ListRegistrationsByEvent(r.Context(), event.ID)
		if err != nil {
			slog.Error("admin registrations", "event", event.ID, "err", err)
			http.Error(w, "admin unavailable", http.StatusInternalServerError)
			return
		}
		renderPublic(w, r, admin.RegistrationsPage(toRegistrationsVD(event, rows)))
	}
}

// AdminRegistrationAccept serves POST /admin/registrations/{rid}/accept —
// the bridge: creates the entry and marks the registration accepted.
func AdminRegistrationAccept(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reg, ok := loadPendingRegistration(w, r, st)
		if !ok {
			return
		}
		u := session.FromContext(r.Context())
		if _, err := st.AcceptRegistration(r.Context(), reg.ID, u.ID); err != nil {
			slog.Error("accept registration", "reg", reg.ID, "err", err)
			http.Error(w, "could not accept registration", http.StatusInternalServerError)
			return
		}
		hxRedirect(w, r, registrationsURL(reg.EventID))
	}
}

// AdminRegistrationWaitlist serves POST /admin/registrations/{rid}/waitlist.
func AdminRegistrationWaitlist(st *store.Store) http.HandlerFunc {
	return setRegistrationStatusHandler(st, "waitlisted")
}

// AdminRegistrationReject serves POST /admin/registrations/{rid}/reject.
func AdminRegistrationReject(st *store.Store) http.HandlerFunc {
	return setRegistrationStatusHandler(st, "rejected")
}

// setRegistrationStatusHandler builds a handler that moves a pending
// registration to the given status.
func setRegistrationStatusHandler(st *store.Store, status string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reg, ok := loadPendingRegistration(w, r, st)
		if !ok {
			return
		}
		u := session.FromContext(r.Context())
		if err := st.SetRegistrationStatus(r.Context(), reg.ID, u.ID, status); err != nil {
			slog.Error("set registration status", "reg", reg.ID, "status", status, "err", err)
			http.Error(w, "could not update registration", http.StatusInternalServerError)
			return
		}
		hxRedirect(w, r, registrationsURL(reg.EventID))
	}
}

// loadPendingRegistration loads the {rid} registration and confirms it is
// still pending. Writes 404/409 and returns ok=false otherwise.
func loadPendingRegistration(w http.ResponseWriter, r *http.Request, st *store.Store) (store.RegistrationRef, bool) {
	rid, err := strconv.ParseInt(r.PathValue("rid"), 10, 64)
	if err != nil || rid <= 0 {
		http.NotFound(w, r)
		return store.RegistrationRef{}, false
	}
	reg, err := st.GetRegistrationDetail(r.Context(), rid)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return store.RegistrationRef{}, false
	}
	if err != nil {
		slog.Error("load registration", "reg", rid, "err", err)
		http.Error(w, "admin unavailable", http.StatusInternalServerError)
		return store.RegistrationRef{}, false
	}
	if reg.Status != "pending" {
		http.Error(w, "registration already reviewed", http.StatusConflict)
		return store.RegistrationRef{}, false
	}
	return store.RegistrationRef{ID: reg.ID, EventID: reg.EventID}, true
}

func registrationsURL(eventID int64) string {
	return "/admin/events/" + strconv.FormatInt(eventID, 10) + "/registrations"
}

// AdminAssignments serves GET /admin/events/{id}/assignments — judge
// assignment per trial (D6).
func AdminAssignments(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		event, ok := loadAdminEvent(w, r, st)
		if !ok {
			return
		}
		trials, err := st.TrialsByEvent(r.Context(), event.ID)
		if err != nil {
			slog.Error("assignments trials", "event", event.ID, "err", err)
			http.Error(w, "admin unavailable", http.StatusInternalServerError)
			return
		}
		judges, err := st.AssignableJudges(r.Context())
		if err != nil {
			slog.Error("assignable judges", "err", err)
			http.Error(w, "admin unavailable", http.StatusInternalServerError)
			return
		}
		renderPublic(w, r, admin.AssignmentsPage(toAssignmentsVD(r.Context(), st, event, trials, judges)))
	}
}

// AdminAssignJudge serves POST /admin/events/{id}/trials/{tid}/judge —
// assigns a judge to a trial (bulk-updating its entries). Selecting the
// blank option is a no-op.
func AdminAssignJudge(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		event, ok := loadAdminEvent(w, r, st)
		if !ok {
			return
		}
		trialID, err := strconv.ParseInt(r.PathValue("tid"), 10, 64)
		if err != nil || trialID <= 0 {
			http.NotFound(w, r)
			return
		}
		// Guard that the trial belongs to the event in the URL.
		trial, err := st.GetTrial(r.Context(), trialID)
		if errors.Is(err, sql.ErrNoRows) || (err == nil && trial.EventID != event.ID) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			slog.Error("assign judge trial", "trial", trialID, "err", err)
			http.Error(w, "admin unavailable", http.StatusInternalServerError)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		judgeID, _ := strconv.ParseInt(r.FormValue("judge"), 10, 64)
		if judgeID > 0 {
			if err := st.AssignTrialJudge(r.Context(), trialID, judgeID); err != nil {
				slog.Error("assign judge", "trial", trialID, "judge", judgeID, "err", err)
				http.Error(w, "could not assign judge", http.StatusInternalServerError)
				return
			}
		}
		hxRedirect(w, r, "/admin/events/"+strconv.FormatInt(event.ID, 10)+"/assignments")
	}
}

// AdminNotifyJudges serves POST /admin/events/{id}/notify-judges — the D6
// "Notify judges" action. It collects the distinct judges assigned across the
// event's trials and returns an htmx confirmation. Email delivery is not wired
// yet (the mail client only targets the contact-form recipient), so the
// recipients are logged and the confirmation says delivery is pending.
func AdminNotifyJudges(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		event, ok := loadAdminEvent(w, r, st)
		if !ok {
			return
		}
		trials, err := st.TrialsByEvent(r.Context(), event.ID)
		if err != nil {
			slog.Error("notify judges trials", "event", event.ID, "err", err)
			http.Error(w, "admin unavailable", http.StatusInternalServerError)
			return
		}
		seen := make(map[string]bool)
		for _, t := range trials {
			email, err := st.TrialJudgeEmail(r.Context(), t.ID)
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			if err != nil {
				slog.Error("notify judges email", "trial", t.ID, "err", err)
				continue
			}
			if email != "" {
				seen[email] = true
			}
		}
		recipients := make([]string, 0, len(seen))
		for e := range seen {
			recipients = append(recipients, e)
		}
		slog.Info("notify judges requested (delivery not yet wired)", "event", event.ID, "recipients", recipients)
		renderPublic(w, r, admin.NotifyJudgesResult(len(seen)))
	}
}

// challengesPageSize is the number of queue rows shown per page (D7).
const challengesPageSize = 12

// challengeStatusFilters whitelists the status filter values, in chip order.
// "" means all.
var challengeStatusFilters = []string{"", "open", "under_review", "resolved", "dismissed"}

func validChallengeFilter(status string) bool {
	for _, s := range challengeStatusFilters {
		if s == status {
			return true
		}
	}
	return false
}

// AdminChallenges serves GET /admin/challenges and
// GET /admin/challenges/{id} — the cross-event review queue (D7), with the
// selected challenge open in the detail panel. The queue is filtered by
// status, sorted, and paginated via query params (status, sort, page).
func AdminChallenges(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		q := r.URL.Query()

		status := q.Get("status")
		if !validChallengeFilter(status) {
			status = ""
		}
		sort := q.Get("sort")
		if !store.ChallengeSortValid(sort) {
			sort = "newest"
		}
		page, _ := strconv.Atoi(q.Get("page"))
		if page < 1 {
			page = 1
		}

		total, err := st.CountChallenges(ctx, status)
		if err != nil {
			slog.Error("admin challenges count", "err", err)
			http.Error(w, "admin unavailable", http.StatusInternalServerError)
			return
		}
		lastPage := int((total + challengesPageSize - 1) / challengesPageSize)
		if lastPage < 1 {
			lastPage = 1
		}
		if page > lastPage {
			page = lastPage
		}
		offset := int64((page - 1) * challengesPageSize)

		rows, err := st.ListChallengesPage(ctx, status, sort, challengesPageSize, offset)
		if err != nil {
			slog.Error("admin challenges", "err", err)
			http.Error(w, "admin unavailable", http.StatusInternalServerError)
			return
		}
		counts, err := st.ChallengeStatusCounts(ctx)
		if err != nil {
			slog.Error("admin challenges counts", "err", err)
			http.Error(w, "admin unavailable", http.StatusInternalServerError)
			return
		}

		var selectedID int64
		var detail *admin.ChalDetail
		if idStr := r.PathValue("id"); idStr != "" {
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil || id <= 0 {
				http.NotFound(w, r)
				return
			}
			c, err := st.GetChallengeDetail(ctx, id)
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			if err != nil {
				slog.Error("challenge detail", "challenge", id, "err", err)
				http.Error(w, "admin unavailable", http.StatusInternalServerError)
				return
			}
			selectedID = id
			d := chalDetailVD(r, st, c)
			detail = &d
		}

		vd := toChallengesVD(challengesVDInput{
			rows:       rows,
			counts:     counts,
			total:      int(total),
			status:     status,
			sort:       sort,
			page:       page,
			offset:     int(offset),
			selectedID: selectedID,
			detail:     detail,
		})
		renderPublic(w, r, admin.ChallengesPage(vd))
	}
}

// AdminChallengeStatus serves POST /admin/challenges/{id}/status — advances
// a challenge to under_review, resolved, or dismissed (with optional notes).
func AdminChallengeStatus(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id <= 0 {
			http.NotFound(w, r)
			return
		}
		c, err := st.GetChallengeDetail(r.Context(), id)
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			slog.Error("challenge status load", "challenge", id, "err", err)
			http.Error(w, "admin unavailable", http.StatusInternalServerError)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		status := r.FormValue("status")
		if !validChallengeTarget(status) {
			http.Error(w, "invalid status", http.StatusBadRequest)
			return
		}
		// Only an unresolved challenge can change state.
		if c.Status == "resolved" || c.Status == "dismissed" {
			http.Error(w, "challenge already closed", http.StatusConflict)
			return
		}
		notes := strings.TrimSpace(r.FormValue("notes"))
		u := session.FromContext(r.Context())
		var resolvedBy int64
		var resolvedAt time.Time
		if status == "resolved" || status == "dismissed" {
			resolvedBy = u.ID
			resolvedAt = time.Now()
		}
		if err := st.UpdateChallengeStatus(r.Context(), id, status, notes, resolvedBy, resolvedAt); err != nil {
			slog.Error("update challenge", "challenge", id, "err", err)
			http.Error(w, "could not update challenge", http.StatusInternalServerError)
			return
		}
		hxRedirect(w, r, "/admin/challenges/"+strconv.FormatInt(id, 10))
	}
}

// validChallengeTarget reports whether status is a valid workflow target.
func validChallengeTarget(status string) bool {
	switch status {
	case "under_review", "resolved", "dismissed":
		return true
	}
	return false
}

// AdminUsers serves GET /admin/users — accounts and roles (D8), with a role
// filter. htmx filter requests receive only the table fragment.
func AdminUsers(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := st.ListUsers(r.Context())
		if err != nil {
			slog.Error("admin users", "err", err)
			http.Error(w, "admin unavailable", http.StatusInternalServerError)
			return
		}
		filter := r.URL.Query().Get("role")
		if !validUserFilter(filter) {
			filter = ""
		}
		q := r.URL.Query().Get("q")
		self := session.FromContext(r.Context())
		data := toUsersVD(rows, self.ID, filter, q)
		if r.Header.Get("HX-Request") == "true" {
			renderPublic(w, r, admin.UsersResults(data))
			return
		}
		renderPublic(w, r, admin.UsersPage(data))
	}
}

// AdminUserRole serves POST /admin/users/{id}/role — inline role change.
// An admin cannot change their own role (lockout guard). Returns the
// re-rendered role control fragment.
func AdminUserRole(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id <= 0 {
			http.NotFound(w, r)
			return
		}
		self := session.FromContext(r.Context())
		if id == self.ID {
			http.Error(w, "cannot change your own role", http.StatusForbidden)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		role := r.FormValue("role")
		if !validRole(role) {
			http.Error(w, "invalid role", http.StatusBadRequest)
			return
		}
		if err := st.UpdateUserRole(r.Context(), id, role); err != nil {
			slog.Error("update user role", "user", id, "err", err)
			http.Error(w, "could not update role", http.StatusInternalServerError)
			return
		}
		renderPublic(w, r, admin.UserRoleControl(admin.UserRow{ID: id, Role: role}))
	}
}
