package handler

import (
	"context"
	"log/slog"
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

// toChallengesVD builds the D7 queue with optional selected detail.
func toChallengesVD(rows []db.ListAllChallengesRow, selectedID int64, detail *admin.ChalDetail) admin.ChallengesViewData {
	var counts admin.ChalStatusCounts
	out := make([]admin.ChalRow, 0, len(rows))
	for _, c := range rows {
		switch c.Status {
		case "open":
			counts.Open++
		case "under_review":
			counts.UnderReview++
		case "resolved":
			counts.Resolved++
		case "dismissed":
			counts.Dismissed++
		}
		out = append(out, admin.ChalRow{
			ID:       c.ID,
			Title:    c.DogName + " · " + disciplineLevelLabel(c.Discipline, c.Level),
			Sub:      c.EventName + " · @" + c.FilerHandle + " · " + relativeTime(c.FiledAt),
			Status:   c.Status,
			Selected: c.ID == selectedID,
		})
	}
	return admin.ChallengesViewData{Counts: counts, Rows: out, Selected: detail}
}

// chalDetailVD maps a challenge detail row into the view.
func chalDetailVD(c db.GetChallengeDetailRow) admin.ChalDetail {
	return admin.ChalDetail{
		ID:              c.ID,
		Title:           c.DogName + " · " + disciplineLevelLabel(c.Discipline, c.Level),
		Status:          c.Status,
		Filed:           "Filed by @" + c.FilerHandle + " · " + relativeTime(c.FiledAt),
		EntryID:         c.EntryID,
		EntryTitle:      c.EventName + " · " + disciplineLevelLabel(c.Discipline, c.Level) + " · " + entryNumberLabel(c.EntryNumber),
		EntrySub:        "Entry is " + c.EntryStatus,
		EventKey:        disciplineKey(c.Discipline),
		Reason:          c.Reason,
		ResolutionNotes: c.ResolutionNotes,
		CanStart:        c.Status == "open",
		CanClose:        c.Status == "open" || c.Status == "under_review",
	}
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
