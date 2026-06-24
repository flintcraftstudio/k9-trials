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
- **Tranche 2 (A7 / A8 / D7)** — done, smoke-tested against the DEMO_MODE seed across all
  status branches. See "Tranche 2 — DONE" below for what shipped.

## Tranche 2 — DONE

### A7 · Challenges list — `challenges_list.templ`, `account_challenges_mapper.go:toChallengesListVD`
- Added filter-chip strip (`All/Open/Review/Resolved/Dismissed` + counts), `ChallengeFilter`
  struct, `?status=` htmx swap of `#challenges-results` fragment (cloned A5). Reuses admin's
  `validChallengeFilter`. Empty-state gates on `Total==0` (teaching state, no chips); a filter
  with no matches shows "No challenges match this filter."
- Header "Last update {relative}" (max UpdatedAt across all rows). Per-row detail line is now
  status-dependent: open→"waiting on admin", under_review→"admin started review {rel}", etc.

### A8 · File a challenge — `challenges_new.templ`, `account_challenges_mapper.go:toChallengeNewVD`
- Disputing card now shows the **scoresheet excerpt**: a Q/NQ result pill, an adaptive reason
  line (`challengeExcerpt`: AutoNQ trigger description quote → Insufficient tally → below-threshold
  summary → Q score summary), and a "View full scoresheet →" link to A6.
- Sub reads "{dog} · {date} · judged by {judge} · finalized"; judge resolved via
  `st.TrialJudgeEmail` + `judgeName`, clause dropped when unavailable.
- New helpers `challengeExcerpt` / `firedTriggerReasons` / `challengeJudgeName` are shared with D7.

### D7 · Challenge review — `admin/challenges.templ`, `admin_review_mapper.go:chalDetailVD`
- `GetChallengeDetail` query extended with `ch.updated_at, t.id, t.template_version` (sqlc regen,
  no migration). `chalDetailVD` now takes `(r, st, c)` and re-evaluates the score.
- Entry-disputed card: result in sub ("Judged by … · finalized · result NQ"), trial date in title,
  NQ-reason excerpt line (reuses A8's `challengeExcerpt`), "View full scoresheet →" → `/entries/{id}`
  (public scoresheet; the old `/account/entries` link would 404 for a non-owner admin).
- **Audit timeline** (`ChalAuditStep` + `chalAudit`/`chalDotStyle`): finalized→filed→(branch by
  status). Only one `updated_at` exists, so intermediate transitions aren't reconstructable — the
  terminal step carries it. Filed line extends with "· review started/resolved/dismissed {rel}".
- Reviewer/resolver *name* is genuinely not in the schema (UpdateChallengeStatus records
  resolved_by only on resolve/dismiss, not on start-review) — "by {admin}" clause omitted, not faked.

### Quick wins — DONE
- **D6 assignments** (`admin/assignments.templ`): "Notify judges" header button → POST
  `/admin/events/{id}/notify-judges` → htmx confirmation into `#notify-result`. Disabled until a
  judge is assigned (`AssignedJudges` count); singular/plural copy. **No email backend yet** (mail
  client only targets the contact-form recipient) — recipients are logged and the banner says
  "Delivery pending mail setup." Real per-judge delivery is a later task.
- **D2 events** + **D8 users**: in-memory `?q=` search box (events: name/slug; users:
  email/name/handle) + rendered as true tables with header rows. D8 surfaces the `Created` column.
  Chips + search compose: the whole `#events-results`/`#users-results` block swaps so active
  status/role + query stay consistent; status counts span all rows (search only narrows visible).
  Shared `orDash` helper; `eventsListURL`/`usersListURL` (handler) bake `q` into chip hrefs.
- ⛔ **D2 Archived filter chip — DEFERRED**: the events CHECK constraint only allows
  `draft/published/closed`; `archived` needs a migration, which belongs with **D3's archive
  lifecycle** (Tranche 3). Add the chip when that status lands.

## Tranche 3 — bigger (all three blockers now resolved; see "Decisions" below)
- **D1 dashboard** — recent-activity feed (needs a new query/data source) + quick-actions card +
  2-col board layout. [L, unblocked but needs new data]
- **D5 registrations** — accordion + lifecycle strip + Export CSV + "Add manual entry" +
  **club-secretary badge** when `submitted_by ≠ handler`. *Withdraw* (Q1 decided): a competitor's
  Withdraw on an **accepted** entry **routes to admin for confirmation** (a request, not an
  immediate action); on confirm the entry becomes `status = withdrawn` with the **entry_number and
  row retained** for audit (number is NOT freed). Needs a `withdrawn` entry status + an
  admin-confirm step on D5; A6 shows "withdrawal requested / pending admin". [L]
- **R1 register** — stepped-checkout chrome (step indicator, selected-dog 2px discipline border,
  avatars), live "N trials selected for {dog}" count, per-trial entry-count/judge metadata. R1c
  "Notify me" (Q4 decided): **requires login**; a logged-in competitor subscribes to the **event**
  and is emailed when it opens registration (`status → published`). Needs a subscriptions table +
  a hook on the event publish transition (email delivery itself is still unwired — log like D6). [L]
- **A4 dog form** (A4 decided: **add `sex` now**) — migration `ALTER TABLE dogs ADD COLUMN sex TEXT`
  with a CHECK in `('male','female','')`; wire store insert/update + handler parse + the form Sex
  select. Also breed autocomplete (separate, can defer). [M]
- **D3 event form** — audit block (created/published/last-edited), Archive lifecycle action,
  fuller at-a-glance (judge-coverage + total entries), `archived` status. [L, needs timestamps].
  Adding the `archived` status here **unblocks the deferred D2 Archived filter chip** — do them
  together.
- **D4 trials** — new-trial as slide-over (currently full page), pill-chip discipline/level
  selectors, "1 trial without a judge" summary. [M]

## Decisions (resolved 2026-06-24 — were the Tranche 3 blockers)
- **Q1** (D5/A6) — withdrawal after accept: **retain for audit, admin-confirmed**. Withdraw is a
  request that routes to admin; on confirm, entry → `withdrawn`, entry_number + row retained, number
  NOT freed.
- **Q4** (R1c) — "Notify me": subscribes to **event registration opening**, **requires login**;
  emailed when the event publishes.
- **A4** — dogs `sex` column: **add it now** (migration + store/parse/form wiring).
