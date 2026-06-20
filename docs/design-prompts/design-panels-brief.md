# Design handoff: remaining hi-fi panels (Public, Auth, Account, Registration, Admin)

Companion to the seven judge panels (`B1–B6`) already drafted in `docs/panels/`. This brief
covers everything else on the flow map (`docs/panels/flow/F0 Flow map.html`): the four
non-judge surfaces, broken down panel-by-panel with **purpose** and **available user actions**.

Panel IDs match the flow map so they slot in alongside the judge set:
`P` public, `U` auth, `A` account, `R` registration, `D` admin.

## Shared context for every panel

- **Stack**: Go + templ + htmx + Alpine.js + Tailwind. Server-rendered HTML; htmx partial
  swaps for live regions and inline edits; Alpine for small client state (drawer/accordion
  open-close, inline-edit toggles). Every meaningful action is an HTTP round trip returning HTML.
- **Design system**: the K9 Elements system is vendored — reuse its typography, color tokens,
  and primitives (`Button`, `Container`, `Eyebrow`, `Pill`, `Field`). Match the judge UI's tier
  palette exactly: Excellent green, Very Good blue/teal, Good neutral, Sufficient amber,
  Insufficient red. **Never convey tier by color alone** — pair the badge text with the color.
- **Empty states are the default**, pre-launch and well past it. Every list panel needs a real,
  friendly, instructive empty-state sketch with an obvious next step — never a bare "No data."
- **Tier / pass-fail semantics** are read-only truth on public + account; only the judge flow
  changes scores. There is no scoresheet-edit UI anywhere outside the judge tablet.

## Navigation harness (already implemented — refine the visual treatment, don't redesign structure)

- **Global top bar** (public, account, admin): left = K9 Elements logo → `/`; center = persistent
  public links (Home, Events, Competitors, Contact), never role-dependent; right (anon) = Sign up +
  Log in; right (authed) = a **role chip** ("My account" / "Judge" / "Admin") + Log out. The chip
  area is **plural under the hood** — a competing admin shows two chips. Don't design a layout that
  assumes exactly one chip.
- **Account sub-tabs** (`/account/*` only): horizontal pill row on a tinted band below the top bar.
  Tabs: Dashboard / Profile / Dogs / Entries / Challenges. Active tab = primary fill +
  `aria-current="page"`. On phones the row overflow-scrolls horizontally, never wraps.
- **Admin sidebar** (`/admin/*` only): 240px fixed left rail. Top group always shown:
  Dashboard / Events / Challenges / Users. Inside a specific event, a second group **"This event"**
  appears: Settings / Trials / Registrations / Judges. No breadcrumbs in v1. Desktop-first is fine
  for the first pass (icon-rail/drawer collapse is v1.5).
- **Judge drawer**: out of scope here (judge set is done).

---

# Surface P — Public (anonymous, phone-first)

Spectators at the field, relatives watching from home, clubs posting links to social. No auth UI
visible. URLs are guessable and shareable; entry/leaderboard pages need Open Graph tags so links
unfurl well. Only `published`/`closed` events and `finalized` entries are visible — in-progress
entries appear as "scoring" with **no partial points**.

### P1 — Events index `/events`
- **Purpose**: front door / discovery; may double as the marketing landing if the club has no
  separate site (leave room for a hero above the list).
- **Actions**: scan upcoming + recent events; tap a card → P2. A "live now" event (any trial
  `in_progress`) is visually prominent. Cards show name, location, dates, status badge
  (upcoming / live now / complete).
- **States**: populated; empty ("No public events yet" + brief explanation).

### P2 — Event detail `/events/{slug}`
- **Purpose**: one event's header + per-trial summary.
- **Actions**: tap a trial row → P3; if a trial is live, a "Live now: OB-L1" pill jumps straight
  to its leaderboard; if registration is open and the visitor is authed as competitor/admin, the
  "Register" CTA → R1.
- **Content**: header (name, location, dates, overall status); trials grouped by date, each row =
  discipline + level, date, status, one-line summary ("12 of 18 scored — leader 92% Excellent"
  live; "Complete — 14 passed of 18" done).
- **States**: pre-trial ("Schedule coming soon"); live; complete.

### P3 — Trial leaderboard `/events/{slug}/trials/{id}` (priority panel)
- **Purpose**: the screen spectators refresh during a trial.
- **Actions**: toggle sort (by score desc / by entry number chronological); tap a row → P4;
  optional name search/filter (nice-to-have).
- **Content**: ranked finalized entries (rank, entry #, handler, dog + breed, points/max + percent,
  tier badge, pass/NQ). NQ'd entries sit below a divider with reason inline. While the trial runs it
  auto-refreshes (~15s htmx poll) and new finalizations animate in with a subtle highlight.
- **States**: pending/0 finalized ("Scoring starts soon…" + live indicator); live mixed; complete.
- Provide **both phone and desktop** layouts (this and P4 are where users actually live).

### P4 — Public entry detail `/entries/{id}` (priority panel)
- **Purpose**: the page handlers share with friends; full breakdown of one finalized entry.
- **Actions**: read the breakdown; navigate back to trial/event; optional Share button (the URL is
  the real share artifact). Judge identity is never shown.
- **Content**: header (entry #, handler, dog + breed, trial/event context) with the big result
  (points/max, percent, tier, pass/fail). Body = per-phase sections, each exercise with code+name,
  points/max, tier; optionally expandable to criterion level. Callouts for auto-triggers
  ("Trial NQ: dog left field") and modifiers ("Lifeline applied: −20, tier capped at Very Good").
- **States**: draw a passing (Very Good) and a failed (Trial NQ) example to show both treatments.
- Strong Open Graph tags here specifically; **both phone and desktop** layouts.

### P5 — Competitor search/directory `/competitors`
- **Purpose**: find a handler or dog.
- **Actions**: search over competitor handle/display_name + dog call/registered name + dog
  registration number (htmx-driven results); tap a result card → P6.
- **Content**: result cards (display name, @handle, dog count, finalized-entry count); optional
  "Recently active" cluster above the search for discovery.
- **States**: empty/landing (one-line tip on what's searchable); results; no-match.

### P6 — Competitor public profile `/competitors/{handle}`
- **Purpose**: the handler's public page — linked from search, leaderboards, dog profiles, social.
  Doubles as "is this dog/handler real?" verification, so it must feel intentional even at zero
  entries — never under-construction.
- **Actions**: read bio; tap a dog card → P7; filter event history by discipline / year. **No edit
  affordances** (read-only public).
- **Content**: header (display name, @handle, optional bio); Dogs section (card per dog); Event
  history (chronological finalized entries: event, discipline+level, dog, tier, pass/fail, points).
- **States**: full; zero-entry-but-intentional.

### P7 — Dog public profile `/dogs/{id}`
- **Purpose**: a dog's public record.
- **Actions**: tap owner link → P6; read trial history. Read-only.
- **Content**: header (call name, registered name, breed, owner link, DOB/age + registration # if
  known); trial history (chronological finalized entries; same fields as P6 but the **handler is
  per-entry** — not always the owner — so show the handler name on each row).

---

# Surface U — Auth

Self-signup is the only self-service path; admins/judges are seeded via `mage seed`.

### U1 — Sign up `/signup`
- **Purpose**: competitor self-registration.
- **Actions**: enter email, password, display name, handle; live inline validation via htmx
  (email uniqueness, handle uniqueness, handle slug format = lowercase letters/digits/hyphens).
  Submit → creates `users` (role=competitor) + matching `competitors` row, starts session,
  lands on A1. Link to `/login` below the form.
- **States**: clean; per-field error (taken handle, bad slug, dup email); submitting.

### U2 — Log in `/login`
- **Purpose**: existing-user sign in. (Template already exists — refine to match the system.)
- **Actions**: email + password → session; on success, competitors land on A1, admins on D1,
  judges on the judge shell. Link to `/signup`.

---

# Surface A — Competitor account (wrapped in the sub-tab strip)

The competitor loop: browse public → log in → register → wait → see results → maybe challenge.
"My account" chip is active throughout. Most lists start empty.

### A1 — Dashboard `/account` (tab: Dashboard)
- **Purpose**: the "what should I do?" landing.
- **Actions**: jump to the soonest accepted entry; view recent results; open challenges list; quick
  actions (Register for an event → R1/P-events, Add a dog → A4, View public profile → P6).
- **Content**: **Up next** (soonest upcoming accepted entry — event, trial, dog, date, entry # when
  assigned — big and obvious); Recent results (last 2–3 finalized w/ pass/fail + score);
  Open challenges (count + link).
- **States**: brand-new account empty state ("Add your first dog to get started" → A4).

### A2 — Profile editor `/account/profile` (tab: Profile)
- **Purpose**: edit public identity.
- **Actions**: edit display name, handle, bio; save via htmx with inline success toast. Show the
  resulting public URL (`/competitors/{handle}`) so the user sees the consequence of changing it.

### A3 — Dogs list `/account/dogs` (tab: Dogs)
- **Purpose**: manage the competitor's dogs.
- **Actions**: prominent "Add a dog" → A4; per-row → A4 edit mode.
- **Content**: card/row per dog (call name, breed, age from DOB, registration #, last competed date).
- **States**: empty = large CTA card.

### A4 — Dog form (add/edit) `/account/dogs/new` + `/account/dogs/{id}/edit` (tab: Dogs)
- **Purpose**: single form for both create and edit.
- **Actions**: required call name; optional registered name, breed (autocomplete from common breeds),
  DOB, registration #. Save → back to A3 with success. Edit mode also shows a "Public profile" link → P7.

### A5 — Entries list `/account/entries` (tab: Entries)
- **Purpose**: every entry across all events the competitor has been part of.
- **Actions**: filter chips (All / Upcoming / In progress / Finalized); row → A6.
- **Content**: each row = event, trial, dog, date, status badge, score (if finalized).
- **States**: empty → link to `/events` to register.

### A6 — Entry detail (own) `/account/entries/{id}` (tab: Entries)
- **Purpose**: the competitor's read-only view of their own scoresheet.
- **Actions** (status-dependent):
  - `registered` — see running order when assigned; "Withdraw" action (stub for v1).
  - `scoring` — message only ("Scoring in progress — results appear when the judge finalizes").
    **No partial scores.**
  - `finalized` — full breakdown like P4, plus **"Challenge this score"** → A8 (`?entry={id}`).
- **Content**: header (event, trial, dog, handler, entry #, status, date). If challenges exist
  against this entry, show their status inline ("Challenge under review · filed 3 days ago").

### A7 — Challenges list `/account/challenges` (tab: Challenges)
- **Purpose**: all challenges the user has filed.
- **Actions**: filter chips (Open / Under review / Resolved / Dismissed); row → its detail/status.
- **Content**: each row = entry (with event/trial/dog), filed date, status badge, last update.
- **States**: empty ("You haven't filed any challenges…" + link to A5 filtered to finalized).

### A8 — File a challenge `/account/challenges/new?entry={id}` (tab: Challenges)
- **Purpose**: dispute a finalized score.
- **Actions**: read the challenged entry + scoresheet excerpt; write a required long-form **Reason**
  (prompt for specificity — which exercise, expected score); submit → A7 with success.
- **Content**: a "What happens next" callout (admin reviews; may dismiss, or resolve by reopening
  the entry for re-scoring; user notified on status change).

---

# Surface R — Registration bridge (the flagged design risk)

The orange dashed handoff where competitor self-service meets admin review — the flow map calls it
**"the single biggest design risk in the app."** Auth-gated (competitor or admin). Lives under
`/events/*` but is a focused flow: carries the global top bar with "My account" chip active, but
**no account sub-tabs**.

### R1 — Register a dog `/events/{slug}/register`
- **Purpose**: enter a dog in one or more trials within an event.
- **Actions**:
  - **Step 1 — pick a dog**: radio list of the user's dogs; if none, "Add a dog first" → A4 (with a
    return param).
  - **Step 2 — pick trials**: checkbox list of trials (discipline, level, date). A dog can only be
    entered once per trial (DB-enforced — reflect this in the UI).
  - Optional notes (qualification status, special accommodations).
  - Submit → creates `pending` registration rows → redirect to A5 with "Registered for N trials —
    admin will confirm shortly."
- **States**: no published trials yet → "Registration not yet open" instead of the form, link back
  to P2. Make the **pending → accepted → entry** lifecycle legible so the user knows they're waiting
  on admin review (this is the handoff worth designing carefully).

---

# Surface D — Admin (wrapped in the sidebar, desktop-first)

The admin loop: open event → manage trials → review registrations → assign judges → review
challenges. "Admin" chip is the entry point.

### D1 — Dashboard `/admin` (sidebar: Dashboard)
- **Purpose**: the "what needs attention" board.
- **Actions**: jump to events live/starting today; open pending-registration + open-challenge
  queues from their counts; resume recently created (unpublished) events; quick actions
  (New event → D3, View all events → D2, Review challenges → D7).

### D2 — Events list `/admin/events` (sidebar: Events)
- **Purpose**: all events including drafts.
- **Actions**: filter by status (default sort start_date desc); "New event" top-right → D3; row → D3 edit.
- **Content**: each row = name, slug, location, dates, status badge, trial count.

### D3 — Event form (new/edit) `/admin/events/new` + `/admin/events/{id}/edit` (sidebar: Events / Settings)
- **Purpose**: event metadata.
- **Actions**: name, slug (auto-derive from name + manual override + htmx uniqueness check),
  location, start/end date, status. Save → D2 (new) or stay (edit). Edit mode surfaces the
  "This event" sidebar group (Trials / Registrations / Judges live on their own pages).

### D4 — Trials in event `/admin/events/{id}/trials` (sidebar: This event → Trials)
- **Purpose**: manage the event's trials.
- **Actions**: "New trial" → D4-new; per-row delete; row context into registrations/assignments.
- **Content**: grouped by date; each row = discipline+level, date, template_version, status, entry count.

### D4-new — New trial form `/admin/events/{id}/trials/new`
- **Purpose**: add a trial.
- **Actions**: pick discipline (OB/PR/TR/DT), level (1/2/3), date, template_version (defaults to
  latest, rarely edited). Save → D4.

### D5 — Registration review `/admin/events/{id}/registrations` (sidebar: This event → Registrations)
- **Purpose**: the receiving end of the R1 bridge — turn pending requests into entries.
- **Actions**:
  - `pending` row: **Accept** (assigns next `entry_number`, creates the `entries` row),
    **Waitlist**, **Reject** (optional note).
  - `accepted` row: View the entry, Withdraw.
- **Content**: grouped by trial (sub-tab strip or accordion); each row = dog (+ owner link), handler,
  `submitted_by` shown only when ≠ handler (club secretary case), submitted_at, status.
- **States**: "No registrations yet."

### D6 — Judge assignments `/admin/events/{id}/assignments` (sidebar: This event → Judges)
- **Purpose**: assign a judge per trial.
- **Actions**: per trial, pick a judge from the role=judge dropdown; save = bulk update of
  `entries.judge_id` for that trial. A trial with no judge can't be scored — surface that clearly
  with a warning chip. (Concurrent multi-judge trials is an open question — flag inline.)

### D7 — Challenge review queue `/admin/challenges` (sidebar: Challenges)
- **Purpose**: cross-event dispute **workflow** — not a scoring UI.
- **Actions**: default filter open + under_review; row → challenge detail. In detail, three actions:
  **Start review** (→ under_review), **Resolve** (confirmation + notes — but the actual score change
  happens by reopening the entry through the judge flow; no edit UI here), **Dismiss** (notes; score
  unchanged).
- **Content**: each row = entry (link to detail), event, competitor (+ @handle), filed_at, status,
  last_update. Detail = dispute text + the entry's current scoresheet + the three actions.

### D8 — Users & roles `/admin/users` (sidebar: Users)
- **Purpose**: see all users, adjust roles.
- **Actions**: change role inline (htmx dropdown); link to competitor public profile when
  role=competitor. (Admin-side user creation is v1.5 — seed command covers it now.)
- **Content**: each row = email, display name (from linked competitor when competitor), role,
  created date.

---

# Open questions to flag inline on the relevant panels

- **A6** — Withdrawal UX once an admin has already accepted the entry.
- **D6** — Concurrent multi-judge trials.
- **Top bar** — Does an admin get an "Admin" hat on the public site, or stay invisible there?

Mark these with the same orange "open question" chip treatment used elsewhere in the wireframes.

# Suggested drafting order

Follow the two loops from the flow map. **Competitor loop first** (P1–P7 → U1–U2 → A1–A8 → R1),
since it's where most user time goes and R1 is the riskiest handoff; then the **admin loop**
(D1–D8). Priority single panels if you need a starting wedge: **P3 (leaderboard)** and
**P4 (entry detail)** — both need phone + desktop layouts.
