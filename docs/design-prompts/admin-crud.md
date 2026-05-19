# Design prompt: Admin CRUD

Design the admin section of a K9 trial scoring application. The user is a club admin who sets up trial events, schedules trials within events, registers competitor entries, and manages judge accounts. They work from a desk on desktop; tablet support is nice-to-have but not the priority.

## Tech context

Go + templ + htmx + Alpine.js + Tailwind. Server-rendered HTML with htmx for partial updates (modals via `hx-get`, inline edits via `hx-post`). No SPA framework. Design must work with HTML-over-the-wire patterns — every meaningful interaction is an HTTP round trip returning HTML fragments.

## Data model

**Event** — the umbrella for a trial weekend.
- `slug` (URL-safe, e.g., `spring-2026-bozeman`)
- `name` (e.g., "Spring 2026 Bozeman Trial")
- `location` (free text, optional)
- `start_date`, `end_date`
- `status`: `draft` | `published` | `closed`

**Trial** — one (discipline, level) judged on one day, under an event.
- `discipline`: OB (Obedience) | PR (Protection) | TR (Tracking) | DT (Detection)
- `level`: 1 | 2 | 3
- `trial_date`
- `template_version` (rulebook revision, e.g., "2026.1" — usually defaulted, rarely edited)
- `status`: `pending` | `in_progress` | `complete`
- A single event hosts multiple trials (e.g., OB-L1 and OB-L2 on Saturday, TR-L1 on Sunday). Unique on (event, discipline, level, date).

**Entry** — a handler+dog registered to run a specific trial.
- `entry_number` (sequential within trial, e.g., 1-25)
- `handler_name`, `dog_name`, `dog_breed` (breed optional)
- `judge_id` — a user with `role=judge`; nullable (may be unassigned at creation)
- `status`: `registered` | `scoring` | `finalized`

**User (judge)** — email + password account, `role=judge`. Created by an admin; no self-registration.

## Screens needed

1. **Admin dashboard** — landing page after admin login. At-a-glance view of active events, trials in progress, and entries pending scoring. Quick links to recently-touched events. Should answer "what should I do next?"

2. **Events list** — all events the admin has created. Filter by status. Each row links to event detail. Top-level "New event" button. Sort by start date, newest first.

3. **Event detail** — header with event name / location / dates / status. Body: trials grouped by date. Each trial row shows discipline+level, date, status, entry count, judge-assignment progress (e.g., "8 of 12 entries have a judge"), and links into trial detail. "New trial" button. Inline status badge editor (htmx-driven dropdown).

4. **Trial detail** — header with trial info (discipline, level, date, status). Body: entries table — entry #, handler, dog, breed, assigned judge (or "unassigned"), entry status. Each row opens an edit modal. "New entry" button. Bulk action: assign judge to multiple selected entries via checkbox + dropdown.

5. **Entry edit** — modal or side drawer triggered from trial detail. Fields: entry number, handler, dog name, dog breed, judge dropdown (filtered to users with `role=judge`). Save / cancel.

6. **Judges page** — list of judge accounts: email, created date, count of currently-assigned entries, count of finalized entries. "Create judge" form on the same page: email + temporary password. (No edit / delete flows yet — locked-down on purpose.)

## Cross-cutting UX

- **Status badges** must be visually distinct and used consistently across pages. Suggested mapping (designer can refine):
  - Event: `draft` (neutral), `published` (info), `closed` (muted)
  - Trial: `pending` (neutral), `in_progress` (warning/active), `complete` (success/muted)
  - Entry: `registered` (neutral), `scoring` (active), `finalized` (success)
- **Empty states** on every list (no events yet, no trials yet, no entries yet, no judges yet) with a clear primary action that bootstraps the next thing.
- **Confirmation modals** for destructive actions. Deleting an event cascades to its trials and entries; deleting a trial cascades to its entries. The confirmation must surface that scope.
- **Form validation** is server-side. Inline errors render below each field after htmx-driven submit failure. No client-side validation logic.
- **Navigation**: persistent left sidebar or top nav with Events, Judges, and admin's email + logout. Current section highlighted.

## Out of scope

- The score-entry UI (separate prompt — `judge-scoring.md`).
- Public spectator results pages.
- Authentication / login screens (already designed; admin section assumes the user is authenticated).
- Multi-tenancy / multi-club organizations.

## Deliverables

Wireframes or high-fidelity mockups for each of the six screens, plus the global navigation shell. Annotate htmx interactions where it changes the layout decision (e.g., "this dropdown is a partial-swap, not a JS-driven menu").

Show populated states and empty states for at least the events list, trial detail, and judges page.
