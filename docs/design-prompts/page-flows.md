# Design prompt: Page flows — public, account, registration, challenges, admin

This brief documents every page currently stubbed in the application, organized by persona and connected via the shared navigation harness. It is a handoff document for a UX/UI agent producing hi-fi templates.

## Relationship to existing prompts

Three earlier prompts in this folder cover parts of the app and remain authoritative for their scope:

- **`judge-scoring.md`** — the judge tablet UI (B1–B6). Still current. The pages exist and have been wired through real persistence.
- **`public-results.md`** — public event index, event detail, trial leaderboard, entry detail (read-only). Still current for what it covers; this doc extends it with competitor profiles and dog profiles.
- **`admin-crud.md`** — early admin sketches. The data model has expanded since (see "Updated data model" below) — where this doc conflicts with `admin-crud.md`, this doc is authoritative.

What this doc adds: the **competitor self-service surface** (`/account/*`), **event registration** as a competitor-driven flow with admin review, **score challenges** as a dispute workflow, **competitor + dog public profiles**, **signup**, and the **navigation harness** that ties everything together.

## Tech context

Go + templ + htmx + Alpine.js + Tailwind. Server-rendered HTML with htmx partial swaps; Alpine for small client state (drawer open/close, inline edit toggles). The K9 Elements design system (already vendored) provides typography, colors, and shared component primitives (`Button`, `Container`, `Eyebrow`, `Pill`, `Field`). Every meaningful interaction is an HTTP round trip returning HTML.

## Personas

1. **Public visitor** — anonymous; browsing events, competitors, and dogs. No auth, mostly phone-first.
2. **Competitor** — handler/owner who registers dogs, reviews their scores, files challenges. Self-signup.
3. **Judge** — tablet-focused; see `judge-scoring.md`.
4. **Admin** — operates events. Desktop-focused.

A single user can in principle hold multiple roles (judges often compete), but the v1 schema enforces a single `role` per user. The nav harness is built to grow into multi-role; don't design visuals that assume exactly one role chip.

## Updated data model

Entities that exist beyond what `admin-crud.md` describes:

- **Competitors** — public identity for handlers and owners. `handle` (URL slug), `display_name`, `bio`. `user_id` is nullable so admins can pre-create competitor rows for handlers without accounts (junior handlers, club guests).
- **Dogs** — `owner_id → competitors.id`. Call name, registered name, breed, DOB (nullable), registration number. Co-ownership is deferred — a single owner per dog for now.
- **Registrations** — a request to enter a dog in a trial. Status: `pending` / `accepted` / `waitlisted` / `withdrawn` / `rejected`. `submitted_by` is the user account that filed the request, which may not be the handler — clubs can register on behalf. On admin "accept", an `entries` row is created and linked.
- **Entries (updated)** — now carries nullable `dog_id` and `handler_id` FKs to the entities above. The legacy `dog_name` / `handler_name` strings are retained as a print-program snapshot for historical display.
- **Challenges** — a competitor's dispute of a finalized scoresheet. Status: `open` / `under_review` / `resolved` / `dismissed`. Resolution flows through the normal scoring path (admin reopens the entry; judge re-enters scores; new `score_inputs` rows naturally form the audit trail). There is no separate "edit a score from the challenge page" UI.

## Navigation harness

Four contexts, four chrome variants — each tuned to the user's task. The harness is already implemented; the visual treatment is what the UX agent should refine.

### Global top bar — used by public, account, and admin

- **Left**: K9 Elements logo, links to `/`.
- **Center**: persistent public links (Home, Events, Competitors, Contact). Never change with role.
- **Right (anon)**: Sign up + Log in.
- **Right (authenticated)**: a **role chip** pointing at the user's primary section ("My account" / "Judge" / "Admin"), followed by Log out.
- The chip is the entry point from any public page into the user's workspace.

### Account sub-tabs — used by `/account/*` only

Horizontal pill row immediately below the top bar, against a tinted band so it reads as a sub-section header. Tabs: **Dashboard / Profile / Dogs / Entries / Challenges**. Active tab gets `aria-current="page"` and primary fill.

### Admin sidebar — used by `/admin/*` only

240px fixed left rail. Top group (always shown): **Dashboard / Events / Challenges / Users**. When inside a specific event (`/admin/events/{id}/...`), a second group labeled **"This event"** appears below with **Settings / Trials / Registrations / Judges**. The active page is highlighted in whichever group it belongs to. No breadcrumbs in v1.

### Judge drawer — used by judge pages

The judge shell deliberately omits the marketing nav — judges are mid-trial and don't need the wayfinder. The drawer is the only off-ramp:

- Trigger: the hamburger in the queue's app bar (other judge pages keep a contextual `← Queue` link).
- Items (minimal by design): **Switch trial / Admin (admins only) / My account / Log out**.
- Closes on backdrop click or Escape.

---

## Public surface

`/events`, `/events/{slug}`, `/events/{slug}/trials/{id}`, and entry detail are described in `public-results.md`. New public pages stubbed here:

### `/competitors` — Competitor search/directory

Anonymous. Lets a visitor find a handler or a dog by name.

- Search input over: competitor handle/display_name, dog call_name / registered_name, dog registration number.
- Results: list of competitor cards (display name, @handle, dog count, finalized-entry count).
- A "Recently active" cluster above the search may help discovery — designer's call.
- Empty state: a one-line tip on what's searchable.

### `/competitors/{handle}` — Competitor public profile

The handler's public page. Linked from search, leaderboards, dog profile pages.

- **Header**: display name, @handle, optional bio.
- **Dogs** section: a card per dog → link to `/dogs/{id}`.
- **Event history**: chronological list of finalized entries (most recent first). Each row: event, trial discipline+level, dog, tier, pass/fail, points. Filter by discipline / year.
- No edit affordances — read-only public.

### `/dogs/{id}` — Dog public profile

- **Header**: call name, registered name, breed, owner link → `/competitors/{handle}`.
- DOB / age if known; registration number if recorded.
- **Trial history**: chronological list of finalized entries; same fields as competitor history but ordered by date. Handler is per-entry (not always the owner), so display the handler name on each row.

---

## Auth

### `/signup` — Competitor self-signup

The only self-service auth path. Admins and judges are still seeded by `mage seed`.

- Fields: email, password, display name, handle (with availability check via htmx).
- Inline validation: email uniqueness, handle uniqueness, handle slug format (lowercase letters, digits, hyphens).
- On submit: creates `users` (role=competitor), creates matching `competitors` row, starts a session, redirects to `/account`.
- Below the form: link to `/login` for existing users.

Login lives at `/login` and already has a template.

---

## Competitor account surface

Wrapped in the sub-tab strip. The role chip "My account" in the top bar is active when on any `/account/*` page.

### `/account` — Dashboard (tab: **Dashboard**)

The "what should I do?" landing.

- **Up next**: the soonest upcoming accepted entry — event, trial, dog, date, entry number (when assigned). Big and obvious.
- **Recent results**: last 2–3 finalized entries with pass/fail + score.
- **Open challenges**: count + link to `/account/challenges` if any.
- **Quick actions**: Register for an event, Add a dog, View your public profile.

Empty state (brand new account): a friendly nudge — "Add your first dog to get started" linking to `/account/dogs/new`.

### `/account/profile` — Profile editor (tab: **Profile**)

Form for display name, handle, bio. Show the public URL (`/competitors/{handle}`) so the user sees the consequence of changing the handle. Save via htmx; success toast inline.

### `/account/dogs` — Dogs list (tab: **Dogs**)

Cards or table rows for each dog: call name, breed, age (derived from DOB), registration number, last competed date.

- Prominent "Add a dog" CTA.
- Per-row → `/account/dogs/{id}/edit`.
- Empty state: large CTA card.

### `/account/dogs/new` and `/account/dogs/{id}/edit` — Dog form (tab: **Dogs**)

Single form, used for both add and edit.

- **Required**: call name.
- **Optional**: registered name, breed (autocomplete from common breeds), DOB, registration number.
- Save → back to `/account/dogs` with a success state. Edit mode also shows a "Public profile" link to `/dogs/{id}`.

### `/account/entries` — Entries list (tab: **Entries**)

Every entry across all events the competitor has ever been part of.

- Filter chips: **All / Upcoming / In progress / Finalized**.
- Each row: event, trial, dog, date, status badge, score (if finalized).
- Click through → entry detail.
- Empty state: link to `/events` to register.

### `/account/entries/{id}` — Entry detail (tab: **Entries**)

The competitor's read-only view of their own scoresheet.

- **Header**: event, trial, dog, handler, entry number, status, date.
- Behavior by status:
  - `registered` — show running order info when assigned; "Withdraw" action (stub for v1).
  - `scoring` — "Scoring in progress — results appear when the judge finalizes." No partial scores.
  - `finalized` — full scoresheet breakdown like the public entry detail, plus a **"Challenge this score"** CTA → `/account/challenges/new?entry={id}`.
- If any challenges have been filed against this entry, show their status inline ("Challenge under review · filed 3 days ago").

### `/account/challenges` — Challenges list (tab: **Challenges**)

All challenges the user has filed.

- Filter chips: **Open / Under review / Resolved / Dismissed**.
- Each row: entry (with event/trial/dog), filed date, status badge, last update.
- Empty state: "You haven't filed any challenges. From a finalized entry, you can dispute a score if something looks wrong." Link to `/account/entries` filtered to finalized.

### `/account/challenges/new?entry={id}` — File a challenge (tab: **Challenges**)

- **Header**: the entry being challenged, with the relevant scoresheet excerpt or summary so the user has the data in front of them.
- **Reason** field — long textarea, required. Encourage specificity: "Which exercise? What do you think the score should have been?"
- **What happens next** callout: admin reviews; may dismiss or resolve by reopening the entry for re-scoring; user will be notified when status changes.
- Submit → `/account/challenges` with a success state.

---

## Registration bridge

### `/events/{slug}/register` — Register a dog for trials in this event

Auth-gated (requires role=competitor or admin). Linked from `/events/{slug}` when the event is `published` and registration is open. Lives in URL space under `/events/*` because it's tied to a specific event, but is functionally a competitor flow — the page does **not** show the account sub-tabs (it's a focused flow, not a section), but it does carry the global top bar with the "My account" chip active.

- **Header**: event name, location, dates.
- **Step 1 — Pick a dog**: radio list of the user's dogs. If empty, show "Add a dog first" → `/account/dogs/new` with a return param.
- **Step 2 — Pick trials within the event**: checkbox list showing discipline, level, date for each trial. A dog can only be entered once per trial (enforced by DB uniqueness).
- Optional notes field (qualification status, special accommodations).
- Submit creates `registrations` rows in `pending` status, redirects to `/account/entries` with a success message: "Registered for N trials — admin will confirm shortly."

Empty state: if the event has no published trials yet, show "Registration not yet open" instead of the form, with a link back to `/events/{slug}`.

---

## Admin surface

Wrapped in the sidebar. The "Admin" chip in the top bar is the entry point.

The original `admin-crud.md` covers events + trials + entries CRUD in detail. The new surfaces below were added for registrations, challenges, and judge assignment. Where they conflict with `admin-crud.md`, this doc is authoritative.

### `/admin` — Dashboard (sidebar: **Dashboard**)

The "what needs attention" board.

- **Today**: events live or starting today.
- **Pending review**: counts of pending registrations + open challenges, each linked to its review queue.
- **Recently created**: events the admin started but hasn't published yet.
- **Quick actions**: New event, View all events, Review challenges.

### `/admin/events` — Events list (sidebar: **Events**)

All events including drafts.

- Filter by status; default sort by `start_date` desc.
- Each row: name, slug, location, dates, status badge, trial count.
- "New event" button top-right.

### `/admin/events/new` and `/admin/events/{id}/edit` — Event form (sidebar: **Events** / **Settings**)

- Fields: name, slug (auto-derive from name with manual override + uniqueness check via htmx), location, start_date, end_date, status.
- Save → events list (new) or stay on edit (existing).
- Edit mode also surfaces the "This event" sidebar group; the form itself is just metadata. Trials / registrations / judge assignments live on their own pages within the event context.

### `/admin/events/{id}/trials` — Trials in event (sidebar: **This event → Trials**)

- Group by date.
- Each trial row: discipline+level, date, template_version, status, entry count.
- "New trial" CTA.

### `/admin/events/{id}/trials/new` — New trial form

- Discipline (OB/PR/TR/DT), level (1/2/3), date, template_version (defaults to latest; rarely edited).

### `/admin/events/{id}/registrations` — Registration review (sidebar: **This event → Registrations**)

The bridge between competitor self-service and the trial running.

- Group by trial (a sub-tab strip or accordion — designer's call).
- Each registration row: dog (with owner link), handler, submitted_by (shown when not the same as handler — i.e. a club secretary registered), submitted_at, status.
- Per-row actions for `pending` registrations: **Accept** (assigns next `entry_number` and creates the `entries` row), **Waitlist**, **Reject** (with optional note).
- Per-row action for `accepted`: View the entry, Withdraw.
- Empty state: "No registrations yet."

### `/admin/events/{id}/assignments` — Judge assignments (sidebar: **This event → Judges**)

- One section per trial: pick a judge from the dropdown of users with role=judge.
- Save updates `entries.judge_id` for that trial's entries (bulk operation).
- A trial without a judge can't be scored; surface this state clearly with a warning chip.

### `/admin/challenges` — Challenge review queue (sidebar: **Challenges**)

Cross-event view of every challenge.

- Default filter: `open` and `under_review`.
- Each row: entry (with link to entry detail), event, competitor (with @handle), filed_at, status, last_update.
- Open a challenge → detail view with the dispute text, the entry's current scoresheet, and three actions:
  - **Start review** (status → `under_review`) — signals it's being looked at.
  - **Resolve** — opens a confirmation; on confirm, marks resolved with notes. **The actual score change happens by reopening the entry through the judge flow** — there is no scoresheet edit UI on this page. This page is the dispute workflow only.
  - **Dismiss** — closes the challenge with notes; entry score unchanged.

### `/admin/users` — Users + roles (sidebar: **Users**)

- List all users: email, display name (from linked competitor when role=competitor), role, created date.
- Change role inline (dropdown, htmx-driven).
- Link to competitor public profile when role=competitor.
- "Create user" affordance for admins managing volunteer judges is a v1.5 concern — the seed command works for now.

---

## Cross-cutting UX considerations

### Visibility cascade

| Event status | Public sees | Competitor sees (own data) | Admin sees |
|---|---|---|---|
| draft | nothing | nothing | full |
| published | event, trials, entries, finalized scores | their own entries any status | full |
| closed | everything read-only | everything read-only | edit-locked except corrections via challenge resolution |

In-progress entries (`scoring`) appear on the public trial page as "scoring" with **no partial points**. The competitor sees their own in-progress entry on `/account/entries/{id}` — also without partial points.

### Challenge resolution does not have a scoring UI

When an admin "resolves" a challenge with a score change, they reopen the entry through the existing judge flow. The append-only `score_inputs` table provides the audit trail automatically. The admin challenge UI is for **workflow** (open → under review → resolved/dismissed) not for **scoring**.

### Multi-role users — design for multiple chips

The role chip is plural under the hood (`SectionAnchors []NavBarLink`). When the schema later supports a user with multiple roles, the top bar grows two chips ("Admin" + "My account" for a competing admin) without component changes. Avoid visual designs that assume exactly one chip on the right.

### Mobile

- **Public** is phone-first (per `public-results.md`).
- **Account** sub-tabs should overflow-scroll horizontally on small screens rather than wrap.
- **Admin** sidebar collapses to an icon rail at medium widths and a drawer at phone widths — a v1.5 concern. Desktop-only is acceptable for the first hi-fi pass.
- **Judge** is tablet-fixed (landscape primary); see `judge-scoring.md`.

### Empty states are the common case at launch

Most lists will be empty until the first event runs. Give them friendly, instructive copy with an obvious next step — not just "No data."

### Trust signals on public profiles

`/competitors/{handle}` and `/dogs/{id}` are linkable from search, social media, and word-of-mouth, and they double as "is this dog real?" verification for buyers, clubs, and registries. Even with zero finalized entries, the profile should feel intentional, not under-construction.
