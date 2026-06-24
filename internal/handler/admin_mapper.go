package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/flintcraftstudio/k9-trials/internal/db"
	"github.com/flintcraftstudio/k9-trials/internal/store"
	"github.com/flintcraftstudio/k9-trials/internal/view/admin"
)

// dashboardMaxList caps how many events the dashboard shows per cluster.
const dashboardMaxList = 6

// toAdminDashboardVD builds the D1 view: status tallies, the published and
// draft event clusters (each with a trial-count meta line), and the
// needs-review counters.
func toAdminDashboardVD(ctx context.Context, st *store.Store, events []db.Event, pendingRegs, openCh int) admin.DashboardViewData {
	var counts admin.EventStatusCounts
	counts.Total = len(events)
	published := make([]admin.EventLine, 0)
	drafts := make([]admin.EventLine, 0)
	for _, e := range events {
		switch e.Status {
		case "published":
			counts.Published++
			if len(published) < dashboardMaxList {
				published = append(published, eventLine(ctx, st, e))
			}
		case "draft":
			counts.Draft++
			if len(drafts) < dashboardMaxList {
				drafts = append(drafts, eventLine(ctx, st, e))
			}
		case "closed":
			counts.Closed++
		}
	}
	return admin.DashboardViewData{
		PendingRegs:    pendingRegs,
		OpenChallenges: openCh,
		Counts:         counts,
		Published:      published,
		Drafts:         drafts,
	}
}

// eventLine builds a compact dashboard event row with its trial count.
func eventLine(ctx context.Context, st *store.Store, e db.Event) admin.EventLine {
	return admin.EventLine{
		ID:     e.ID,
		Name:   e.Name,
		Meta:   eventMeta(ctx, st, e),
		Status: e.Status,
	}
}

// eventMeta composes "location · dates · N trials", dropping the location
// when unset.
func eventMeta(ctx context.Context, st *store.Store, e db.Event) string {
	parts := []string{}
	if e.Location != "" {
		parts = append(parts, e.Location)
	}
	parts = append(parts, dateRange(e.StartDate, e.EndDate))
	n, err := st.CountTrialsByEvent(ctx, e.ID)
	if err != nil {
		slog.Error("count trials", "event", e.ID, "err", err)
	}
	parts = append(parts, trialsCountWord(int(n)))
	return strings.Join(parts, " · ")
}

// trialsCountWord renders "4 trials" / "1 trial" / "no trials".
func trialsCountWord(n int) string {
	switch n {
	case 0:
		return "no trials"
	case 1:
		return "1 trial"
	default:
		return fmt.Sprintf("%d trials", n)
	}
}

// validEventFilter reports whether key is a recognized event status filter
// (empty means all).
func validEventFilter(key string) bool {
	switch key {
	case "", "draft", "published", "closed":
		return true
	}
	return false
}

// toAdminEventsVD builds the D2 list: status filter chips with counts and the
// rows matching the active status filter and search term, each with its trial
// count. Status counts span all events (independent of the search), so the
// chips stay stable; the search narrows the visible rows by name or slug.
func toAdminEventsVD(ctx context.Context, st *store.Store, events []db.Event, active, q string) admin.EventsListViewData {
	var draft, published, closed int
	for _, e := range events {
		switch e.Status {
		case "draft":
			draft++
		case "published":
			published++
		case "closed":
			closed++
		}
	}

	needle := strings.ToLower(strings.TrimSpace(q))
	rows := make([]admin.EventRow, 0, len(events))
	for _, e := range events {
		if active != "" && e.Status != active {
			continue
		}
		if needle != "" && !strings.Contains(strings.ToLower(e.Name), needle) && !strings.Contains(strings.ToLower(e.Slug), needle) {
			continue
		}
		n, err := st.CountTrialsByEvent(ctx, e.ID)
		if err != nil {
			slog.Error("count trials", "event", e.ID, "err", err)
		}
		rows = append(rows, admin.EventRow{
			ID:       e.ID,
			Name:     e.Name,
			Slug:     e.Slug,
			Location: e.Location,
			Dates:    dateRange(e.StartDate, e.EndDate),
			Trials:   int(n),
			Status:   e.Status,
		})
	}

	return admin.EventsListViewData{
		Total:   len(events),
		Active:  active,
		Query:   q,
		Filters: eventFilters(active, q, len(events), draft, published, closed),
		Rows:    rows,
	}
}

// eventsListURL composes an events-list URL preserving the status filter and
// search term, omitting whichever is empty.
func eventsListURL(status, q string) string {
	v := url.Values{}
	if status != "" {
		v.Set("status", status)
	}
	if q != "" {
		v.Set("q", q)
	}
	if len(v) == 0 {
		return "/admin/events"
	}
	return "/admin/events?" + v.Encode()
}

// eventFilters builds the status chip row with per-status counts, preserving
// the active search term in each chip's href.
func eventFilters(active, q string, total, draft, published, closed int) []admin.EventFilter {
	defs := []struct {
		key, label string
		count      int
	}{
		{"", "All", total},
		{"draft", "Draft", draft},
		{"published", "Published", published},
		{"closed", "Closed", closed},
	}
	out := make([]admin.EventFilter, 0, len(defs))
	for _, d := range defs {
		out = append(out, admin.EventFilter{
			Key:    d.key,
			Label:  d.label,
			Count:  d.count,
			Href:   eventsListURL(d.key, q),
			Active: active == d.key,
		})
	}
	return out
}

// editEventVD builds the D3 edit-form view from an event row plus its
// trial and pending-registration counts.
func editEventVD(ctx context.Context, st *store.Store, e db.Event) admin.EventFormViewData {
	trials, err := st.CountTrialsByEvent(ctx, e.ID)
	if err != nil {
		slog.Error("count trials", "event", e.ID, "err", err)
	}
	pending, err := st.CountPendingRegistrationsByEvent(ctx, e.ID)
	if err != nil {
		slog.Error("count pending", "event", e.ID, "err", err)
	}
	return admin.EventFormViewData{
		IsEdit:      true,
		EventID:     e.ID,
		Name:        e.Name,
		Slug:        e.Slug,
		Location:    e.Location,
		StartDate:   e.StartDate.UTC().Format("2006-01-02"),
		EndDate:     e.EndDate.UTC().Format("2006-01-02"),
		Status:      e.Status,
		TrialCount:  int(trials),
		PendingRegs: int(pending),
		PublicURL:   "/events/" + e.Slug,
	}
}

// toAdminTrialsVD groups an event trials by date for D4, resolving each
// trial entry count and assigned judge.
func toAdminTrialsVD(ctx context.Context, st *store.Store, e db.Event, trials []db.Trial) admin.TrialsViewData {
	days := make([]admin.TrialDay, 0)
	var cur *admin.TrialDay
	curKey := ""
	for _, t := range trials {
		key := t.TrialDate.UTC().Format("2006-01-02")
		if key != curKey {
			days = append(days, admin.TrialDay{Label: t.TrialDate.UTC().Format("Monday · 2 January")})
			cur = &days[len(days)-1]
			curKey = key
		}
		cur.Trials = append(cur.Trials, trialLine(ctx, st, t))
		cur.Count++
	}
	return admin.TrialsViewData{
		EventID:     e.ID,
		EventName:   e.Name,
		EventStatus: e.Status,
		EventSlug:   e.Slug,
		TrialCount:  len(trials),
		Days:        days,
	}
}

// trialLine builds one D4 trial row with entry count and judge name.
func trialLine(ctx context.Context, st *store.Store, t db.Trial) admin.TrialLine {
	entries, err := st.CountEntriesForTrial(ctx, t.ID)
	if err != nil {
		slog.Error("count entries", "trial", t.ID, "err", err)
	}
	meta := fmt.Sprintf("Template v%s · %s", t.TemplateVersion, entriesCountWord(int(entries)))

	judge := ""
	if email, err := st.TrialJudgeEmail(ctx, t.ID); err == nil {
		judge = judgeName(email)
	} else if !errors.Is(err, sql.ErrNoRows) {
		slog.Error("trial judge", "trial", t.ID, "err", err)
	}

	return admin.TrialLine{
		ID:       t.ID,
		Title:    disciplineLevelLabel(t.Discipline, t.Level),
		Meta:     meta,
		EventKey: disciplineKey(t.Discipline),
		Status:   t.Status,
		Judge:    judge,
	}
}

// entriesCountWord renders "12 entries" / "1 entry" / "no entries".
func entriesCountWord(n int) string {
	switch n {
	case 0:
		return "no entries"
	case 1:
		return "1 entry"
	default:
		return fmt.Sprintf("%d entries", n)
	}
}

// slugify derives a URL slug from a name: lowercase, alphanumerics kept,
// runs of other characters collapsed to single hyphens.
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	lastHyphen := false
	for _, r := range s {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			lastHyphen = false
		case r == ' ' || r == '-' || r == '_':
			if !lastHyphen && b.Len() > 0 {
				b.WriteByte('-')
				lastHyphen = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
