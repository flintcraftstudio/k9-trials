package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/flintcraftstudio/k9-trials/internal/db"
	"github.com/flintcraftstudio/k9-trials/internal/store"
	"github.com/flintcraftstudio/k9-trials/internal/view/admin"
)

// toRegistrationsVD groups an event registrations by trial for D5 and
// tallies the status counts.
func toRegistrationsVD(event db.Event, rows []db.ListRegistrationsByEventRow) admin.RegistrationsViewData {
	var counts admin.RegStatusCounts
	groups := make([]admin.RegTrialGroup, 0)
	var cur *admin.RegTrialGroup
	curTrial := int64(0)

	for _, r := range rows {
		counts.Total++
		switch r.Status {
		case "pending":
			counts.Pending++
		case "accepted":
			counts.Accepted++
		case "waitlisted":
			counts.Waitlisted++
		case "rejected":
			counts.Rejected++
		case "withdrawn":
			counts.Withdrawn++
		}

		if r.TrialID != curTrial {
			groups = append(groups, admin.RegTrialGroup{
				Title:    disciplineLevelLabel(r.Discipline, r.Level) + " · " + shortDate(r.TrialDate),
				EventKey: disciplineKey(r.Discipline),
			})
			cur = &groups[len(groups)-1]
			curTrial = r.TrialID
		}

		entryNum := ""
		if r.EntryNumber.Valid {
			entryNum = entryNumberLabel(r.EntryNumber.Int64)
		}
		pending := r.Status == "pending"
		if pending {
			cur.Pending++
		}
		cur.Rows = append(cur.Rows, admin.RegRow{
			ID:          r.ID,
			DogName:     r.DogName,
			DogMeta:     regDogDetail(r.DogRegno, r.DogBreed),
			SubmittedBy: regSubmittedLine(r),
			Status:      r.Status,
			EntryNumber: entryNum,
			Pending:     pending,
		})
	}

	return admin.RegistrationsViewData{
		EventID:   event.ID,
		EventName: event.Name,
		Counts:    counts,
		Trials:    groups,
	}
}

// regDogDetail composes "K9-3187 · Czech GSD", dropping unknown parts.
func regDogDetail(regno, breed string) string {
	parts := []string{}
	if regno != "" {
		parts = append(parts, regno)
	}
	if breed != "" {
		parts = append(parts, breed)
	}
	return strings.Join(parts, " · ")
}

// regSubmittedLine renders "owner @handle · submitted 3 hours ago".
func regSubmittedLine(r db.ListRegistrationsByEventRow) string {
	return "owner @" + r.CompetitorHandle + " · submitted " + relativeTime(r.SubmittedAt)
}

// toAssignmentsVD builds the D6 view: each trial with its entry count and
// current judge, plus the assignable-judge options.
func toAssignmentsVD(ctx context.Context, st *store.Store, event db.Event, trials []db.Trial, judges []db.ListAssignableJudgesRow) admin.AssignmentsViewData {
	opts := make([]admin.JudgeOption, 0, len(judges))
	for _, j := range judges {
		opts = append(opts, admin.JudgeOption{ID: j.ID, Name: judgeName(j.Email)})
	}

	rows := make([]admin.AssignTrial, 0, len(trials))
	unassigned := 0
	for _, t := range trials {
		entries, err := st.CountEntriesForTrial(ctx, t.ID)
		if err != nil {
			slog.Error("assign count entries", "trial", t.ID, "err", err)
		}
		judgeID, assigned, err := st.TrialJudgeID(ctx, t.ID)
		if err != nil {
			slog.Error("assign trial judge", "trial", t.ID, "err", err)
		}
		if !assigned {
			unassigned++
		}
		rows = append(rows, admin.AssignTrial{
			ID:       t.ID,
			Title:    disciplineLevelLabel(t.Discipline, t.Level) + " · " + shortDate(t.TrialDate),
			EventKey: disciplineKey(t.Discipline),
			Entries:  int(entries),
			JudgeID:  judgeID,
			Assigned: assigned,
		})
	}

	return admin.AssignmentsViewData{
		EventID:    event.ID,
		EventName:  event.Name,
		Unassigned: unassigned,
		Trials:     rows,
		Judges:     opts,
	}
}

// challengesVDInput carries the queue state into the view-data mapper.
type challengesVDInput struct {
	rows       []store.ChallengeListRow
	counts     map[string]int // global tally by status
	total      int            // filtered total (for pagination)
	status     string         // active status filter ("" = all)
	sort       string         // active sort key
	page       int            // 1-based
	offset     int            // (page-1)*pageSize
	selectedID int64
	detail     *admin.ChalDetail
}

// challengesURL builds a queue URL, preserving the active status/sort/page and
// omitting defaults so canonical links stay clean.
func challengesURL(base, status, sort string, page int) string {
	q := url.Values{}
	if status != "" {
		q.Set("status", status)
	}
	if sort != "" && sort != "newest" {
		q.Set("sort", sort)
	}
	if page > 1 {
		q.Set("page", strconv.Itoa(page))
	}
	if len(q) == 0 {
		return base
	}
	return base + "?" + q.Encode()
}

// toChallengesVD builds the D7 queue — filter chips, sort links, pagination,
// and rows — from the current queue state and optional selected detail.
func toChallengesVD(in challengesVDInput) admin.ChallengesViewData {
	counts := admin.ChalStatusCounts{
		Open:        in.counts["open"],
		UnderReview: in.counts["under_review"],
		Resolved:    in.counts["resolved"],
		Dismissed:   in.counts["dismissed"],
	}
	allCount := counts.Open + counts.UnderReview + counts.Resolved + counts.Dismissed

	// Status filter chips. Switching filter resets to page 1 and keeps sort.
	filterDefs := []struct {
		key, label string
		count      int
	}{
		{"", "All", allCount},
		{"open", "Open", counts.Open},
		{"under_review", "Under review", counts.UnderReview},
		{"resolved", "Resolved", counts.Resolved},
		{"dismissed", "Dismissed", counts.Dismissed},
	}
	filters := make([]admin.ChalFilter, 0, len(filterDefs))
	for _, f := range filterDefs {
		filters = append(filters, admin.ChalFilter{
			Label:  f.label,
			Count:  f.count,
			Href:   challengesURL("/admin/challenges", f.key, in.sort, 1),
			Active: f.key == in.status,
		})
	}

	// Sort links. Switching sort keeps the filter and resets to page 1.
	sortDefs := []struct{ key, label string }{
		{"newest", "Newest"},
		{"oldest", "Oldest"},
		{"status", "By status"},
	}
	sorts := make([]admin.ChalSortLink, 0, len(sortDefs))
	for _, s := range sortDefs {
		sorts = append(sorts, admin.ChalSortLink{
			Label:  s.label,
			Href:   challengesURL("/admin/challenges", in.status, s.key, 1),
			Active: s.key == in.sort,
		})
	}

	out := make([]admin.ChalRow, 0, len(in.rows))
	for _, c := range in.rows {
		out = append(out, admin.ChalRow{
			ID:       c.ID,
			Title:    c.DogName + " · " + disciplineLevelLabel(c.Discipline, c.Level),
			Sub:      c.EventName + " · @" + c.FilerHandle + " · " + relativeTime(c.FiledAt),
			Status:   c.Status,
			Href:     challengesURL(fmt.Sprintf("/admin/challenges/%d", c.ID), in.status, in.sort, in.page),
			Selected: c.ID == in.selectedID,
		})
	}

	from, to := 0, 0
	if len(in.rows) > 0 {
		from = in.offset + 1
		to = in.offset + len(in.rows)
	}
	pageVD := admin.ChalPage{
		From:     from,
		To:       to,
		Total:    in.total,
		HasPrev:  in.page > 1,
		HasNext:  to < in.total,
		PrevHref: challengesURL("/admin/challenges", in.status, in.sort, in.page-1),
		NextHref: challengesURL("/admin/challenges", in.status, in.sort, in.page+1),
	}

	return admin.ChallengesViewData{
		Counts:   counts,
		Filters:  filters,
		Sorts:    sorts,
		Rows:     out,
		Page:     pageVD,
		Selected: in.detail,
	}
}

// chalDetailVD maps a challenge detail row into the view, re-evaluating the
// disputed entry's score for the result label and excerpt and assembling the
// audit timeline.
func chalDetailVD(r *http.Request, st *store.Store, c db.GetChallengeDetailRow) admin.ChalDetail {
	judge := challengeJudgeName(r, st, db.Trial{ID: c.TrialID})

	result, excerptLabel, excerpt := chalEntryExcerpt(r, st, c)
	d := admin.ChalDetail{
		ID:              c.ID,
		Title:           c.DogName + " · " + disciplineLevelLabel(c.Discipline, c.Level),
		Status:          c.Status,
		Filed:           chalFiledLine(c),
		EntryID:         c.EntryID,
		EntryTitle:      c.EventName + " · " + disciplineLevelLabel(c.Discipline, c.Level) + " · " + entryNumberLabel(c.EntryNumber) + " · " + shortDate(c.TrialDate),
		EntrySub:        chalEntrySub(c, judge, result),
		EventKey:        disciplineKey(c.Discipline),
		ExcerptLabel:    excerptLabel,
		Excerpt:         excerpt,
		Reason:          c.Reason,
		ResolutionNotes: c.ResolutionNotes,
		CanStart:        c.Status == "open",
		CanClose:        c.Status == "open" || c.Status == "under_review",
		Timeline:        chalTimeline(c, judge, result),
	}
	return d
}

// chalEntryExcerpt evaluates the disputed entry's score and returns the
// result label ("NQ"/"Q", empty when unevaluable) plus the excerpt label and
// text for the disputed-entry card, reusing the competitor-side excerpt.
func chalEntryExcerpt(r *http.Request, st *store.Store, c db.GetChallengeDetailRow) (result, label, text string) {
	trial := db.Trial{Discipline: c.Discipline, Level: c.Level, TemplateVersion: c.TemplateVersion}
	tpl, sheet, _, res, err := loadTemplateAndEvaluate(r, st, trial, c.EntryID)
	if err != nil {
		return "", "", ""
	}
	result = "NQ"
	if res.Passed {
		result = "Q"
	}
	label, text = challengeExcerpt(tpl, sheet, res)
	return result, label, text
}

// chalEntrySub renders the disputed-entry sub-line: judge, finalized state,
// and result. Clauses drop out gracefully when their data is unavailable.
func chalEntrySub(c db.GetChallengeDetailRow, judge, result string) string {
	parts := []string{}
	if judge != "" {
		parts = append(parts, "Judged by "+judge)
	}
	parts = append(parts, c.EntryStatus)
	if result != "" {
		parts = append(parts, "result "+result)
	}
	return strings.Join(parts, " · ")
}

// chalFiledLine renders the header attribution: who filed it and when, plus
// the review/resolution clause once the dispute has moved past open.
func chalFiledLine(c db.GetChallengeDetailRow) string {
	line := "Filed by @" + c.FilerHandle + " · " + relativeTime(c.FiledAt)
	switch c.Status {
	case "under_review":
		line += " · review started " + relativeTime(c.UpdatedAt)
	case "resolved":
		line += " · resolved " + relativeTime(c.UpdatedAt)
	case "dismissed":
		line += " · dismissed " + relativeTime(c.UpdatedAt)
	}
	return line
}

// chalTimeline assembles the audit trail. The schema records only the
// challenge's latest updated_at (not each intermediate transition), so the
// terminal step carries that timestamp while earlier transitions show their
// own recorded time.
func chalTimeline(c db.GetChallengeDetailRow, judge, result string) []admin.ChalAuditStep {
	finalized := admin.ChalAuditStep{
		Title: "Entry finalized",
		Meta:  judge,
		When:  shortDate(c.TrialDate),
		Kind:  "lock",
	}
	if result != "" {
		finalized.Title += " · result " + result
	}
	filed := admin.ChalAuditStep{
		Title: "Challenge filed",
		Meta:  "@" + c.FilerHandle,
		When:  relativeTime(c.FiledAt),
		Kind:  "warn",
	}
	steps := []admin.ChalAuditStep{finalized, filed}

	switch c.Status {
	case "open":
		steps = append(steps, admin.ChalAuditStep{
			Title: "Awaiting review",
			Meta:  "with admin",
			When:  "—",
		})
	case "under_review":
		steps = append(steps,
			admin.ChalAuditStep{Title: "Review started", When: relativeTime(c.UpdatedAt), Kind: "green"},
			admin.ChalAuditStep{Title: "Pending — resolve or dismiss", Meta: "awaiting admin decision", When: "—"},
		)
	case "resolved":
		steps = append(steps, admin.ChalAuditStep{Title: "Resolved", Meta: c.ResolutionNotes, When: relativeTime(c.UpdatedAt), Kind: "green"})
	case "dismissed":
		steps = append(steps, admin.ChalAuditStep{Title: "Dismissed", Meta: c.ResolutionNotes, When: relativeTime(c.UpdatedAt), Kind: "muted"})
	}
	return steps
}

// validUserFilter reports whether key is a recognized role filter (empty
// means all).
func validUserFilter(key string) bool {
	switch key {
	case "", "competitor", "judge", "admin":
		return true
	}
	return false
}

// toUsersVD builds the D8 list with role filter chips, marking the
// logged-in admin row so it cannot self-demote.
func toUsersVD(rows []db.ListUsersWithCompetitorRow, selfID int64, active string) admin.UsersViewData {
	var competitors, judges, admins int
	for _, u := range rows {
		switch u.Role {
		case "competitor":
			competitors++
		case "judge":
			judges++
		case "admin":
			admins++
		}
	}

	out := make([]admin.UserRow, 0, len(rows))
	for _, u := range rows {
		if active != "" && u.Role != active {
			continue
		}
		out = append(out, userRowVD(u, selfID))
	}

	return admin.UsersViewData{
		Total:   len(rows),
		Filters: userFilters(active, len(rows), competitors, judges, admins),
		Rows:    out,
	}
}

// userRowVD maps one user row, preferring the competitor display name and
// handle when present.
func userRowVD(u db.ListUsersWithCompetitorRow, selfID int64) admin.UserRow {
	name := u.DisplayName.String
	if name == "" {
		name = emailLocal(u.Email)
	}
	sub := ""
	handle := ""
	if u.Handle.Valid && u.Handle.String != "" {
		handle = u.Handle.String
		sub = "@" + handle
	}
	return admin.UserRow{
		ID:      u.ID,
		Name:    name,
		Sub:     sub,
		Email:   u.Email,
		Created: fullDate(u.CreatedAt),
		Role:    u.Role,
		Handle:  handle,
		IsSelf:  u.ID == selfID,
	}
}

// userFilters builds the role filter chip row with counts.
func userFilters(active string, total, competitors, judges, admins int) []admin.UserFilter {
	defs := []struct {
		key, label string
		count      int
	}{
		{"", "All", total},
		{"competitor", "Competitors", competitors},
		{"judge", "Judges", judges},
		{"admin", "Admins", admins},
	}
	out := make([]admin.UserFilter, 0, len(defs))
	for _, d := range defs {
		href := "/admin/users"
		if d.key != "" {
			href += "?role=" + d.key
		}
		out = append(out, admin.UserFilter{
			Key:    d.key,
			Label:  d.label,
			Count:  d.count,
			Href:   href,
			Active: active == d.key,
		})
	}
	return out
}

// emailLocal returns the local part of an email for a display fallback.
func emailLocal(email string) string {
	if i := strings.IndexByte(email, '@'); i >= 0 {
		return email[:i]
	}
	return email
}

// validRole reports whether role is one the admin may assign.
func validRole(role string) bool {
	switch role {
	case "competitor", "judge", "admin":
		return true
	}
	return false
}
