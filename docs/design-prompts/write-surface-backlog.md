# Write-surface implementation backlog

Resume doc for upgrading the already-built write-heavy screens to their NEW hi-fi
mockups in `panels/K9 Public Panels/*.dc.html`. Companion to `design-panels-brief.md`.

Every surface (P/U/A/R/D) is already implemented and wired — this is a polish/redesign
pass, NOT greenfield. The mockups are a hi-fi refresh of working screens.

## Conventions (carry into a fresh context)
- Stack: Go + templ + htmx + Alpine + Tailwind. Edit `.templ` source only; `*_templ.go`
  is gitignored and regenerated. After editing: `templ generate` → `go build ./...` →
  `go vet ./internal/...`. Commit only `.templ`/`.go` sources.
- Extract a mockup's text: strip `<script>/<style>/<svg>`, then tags. (See git history of
  this session, or just read the `.dc.html`.)
- Pill classes in use: `pill-qual` (green), `pill-closed` (red/NQ), `pill-scoring` (+`dot`),
  `pill-active` / `pill-muted` (toggles/filters), `pill-wait`. Tier palette: Excellent green,
  Very Good teal, Good neutral, Sufficient amber, Insufficient red — never color alone.
- A5 entries list (`internal/view/account/entries_list.templ` + `account_entries_mapper.go`)
  is the reference pattern for htmx filter chips with live counts (`Filters []EntryFilter`).
- `view.SiteName = "K9 Elements"`; canonical domain is `k9elements.com`.

## Done (do not redo)
- P3 leaderboard — sort toggle, NQ divider, live poll (commit 08b2919).
- P4 / A5 / A6 — already at mockup fidelity, untouched.
- Auth + profile copy/bug pass (commit 60c1095): competitor login → `/account`;
  U2 routing banner + create-account link; U1 URL preview; A2 toast + 30-day note.

## Tranche 2 — unblocked, no schema/migration (do these next)

### A7 · Challenges list — `internal/view/account/challenges_list.templ`, `account_challenges_mapper.go:toChallengesListVD`
- GAP [feature]: no filter-chip strip. Mockup: `All · 2 / Open · 1 / Review · 1 / Resolved · 0 / Dismissed · 0`.
- Clone A5's pattern: add per-status counts + a `Filters` struct to `ChallengesListViewData`
  (currently only `Total` + `Rows`), render chips, filter by `?status=` with htmx swap of a
  `#challenges-results` fragment. Count in-memory from the loaded challenges (like A5).
- GAP [copy]: header "Last update {relative}" line; per-row "admin started review yesterday" detail.
- Effort: M.

### A8 · File a challenge — `internal/view/account/challenges_new.templ`, `account_challenges_mapper.go:toChallengeNewVD`
- GAP [feature] (headline): the disputing card must show the **scoresheet excerpt** — the NQ
  reason quote (e.g. "Ring departure during courage test, 1.4s outside the working line") and a
  **"View full scoresheet →"** link. `ChallengeNewViewData` has no NQ-reason/excerpt/scoresheet-link
  fields. The mapper already runs `evalFinalizedScore`; extract the NQ reason / per-exercise excerpt
  from the scoring result and add fields.
- GAP [copy]: disputing sub should read "{dog} · {date} · judged by {judge} · finalized" — the
  struct comment already promises "judged by H. Vance" but the mapper never populates judge name.
- Effort: M.

### D7 · Challenge review — `internal/view/admin/challenges.templ`, `admin_review_mapper.go:chalDetailVD`
- Two-pane is already built (and exceeds mockup: has filter+sort+pagination). Gaps only:
- GAP [feature]: detail has no **audit timeline** (Entry finalized → Challenge filed → Review
  started → Pending). `ChalDetail` has no timeline data; mapper doesn't load it.
- GAP [copy]: entry-disputed card omits the NQ-reason line (`EntrySub` is just "Entry is {status}");
  "Filed by … review started yesterday by {admin}" attribution missing.
- Effort: M.

### Quick wins
- D6 assignments (`admin/assignments.templ`): add a "Notify judges" button. [S]
- D2 events / D8 users: add a search box (`?q=`) + render as a true table with header; D8 already
  has an unrendered `UserRow.Created` field. Also D2 missing an Archived filter chip. [M]

## Tranche 3 — bigger; three blocked on product decisions
- **D1 dashboard** — recent-activity feed (needs a new query/data source) + quick-actions card +
  2-col board layout. [L, unblocked but needs new data]
- **D5 registrations** — accordion + lifecycle strip + Export CSV + "Add manual entry" +
  **club-secretary badge** when `submitted_by ≠ handler`. ⛔ *Withdraw* action blocked on **open
  question Q1** (void-and-free-number vs. retain-for-audit). [L]
- **R1 register** — stepped-checkout chrome (step indicator, selected-dog 2px discipline border,
  avatars), live "N trials selected for {dog}" count, per-trial entry-count/judge metadata. R1c
  ⛔ "Notify me" + open-date email promise blocked on **open question Q4** (what it subscribes to). [L]
- **A4 dog form** — missing **Sex** field. ⛔ Needs a migration + a decision to add it now. Also
  breed autocomplete. [M]
- **D3 event form** — audit block (created/published/last-edited), Archive lifecycle action,
  fuller at-a-glance (judge-coverage + total entries), `archived` status. [L, needs timestamps]
- **D4 trials** — new-trial as slide-over (currently full page), pill-chip discipline/level
  selectors, "1 trial without a judge" summary. [M]

## Open questions to resolve before Tranche 3's blocked items
- **Q1** (D5/A6): withdrawal semantics after accept — void+free entry number, or retain for audit?
- **Q4** (R1c): what does "Notify me" subscribe to, and does it require login?
- **A4**: add a `sex` column to dogs now? (migration + store/parse/view wiring.)
