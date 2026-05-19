# Design prompt: Public scoreboard & event results

Design the public-facing read-only views of a K9 trial scoring application. The audience is spectators at the trial field, family members watching remotely, and participants checking their own / their friends' results. No authentication is required to view these pages.

## Critical context

- **Device mix**: phone-first (spectators at the field on their phones) with desktop a strong second (relatives watching from home, club posting on social media). Tablet is a side case.
- **Liveness**: during a trial day, results update as judges finalize entries. Spectators refresh frequently. Either htmx polling (10–30s interval) or SSE — designer should assume polling unless a better case is made.
- **Visibility rule** (already wired in the data layer): only entries with `status = finalized` appear on these pages. Entries still being scored (`status = scoring`, draft) are hidden until the judge taps "Finalize."
- **No editing**: these are read-only views. No buttons to change anything. No login state visible (a public visitor never sees auth UI).

## Tech context

Go + templ + htmx + Alpine.js + Tailwind. Server-rendered HTML with htmx partial swaps for live polling. URLs must be guessable and shareable (clubs post them on Facebook; handlers send them to their relatives). Open Graph metadata expected so links unfurl nicely on social platforms.

## Data model (visible subset)

The public pages render finalized data only. The same shape used in admin / judge views applies, with these public-relevant fields:

- **Event**: slug, name, location, start_date, end_date, status (only `published` and `closed` events appear publicly; `draft` events are hidden).
- **Trial**: discipline (OB/PR/TR/DT), level (1/2/3), trial_date, status (`pending` / `in_progress` / `complete`).
- **Entry** (finalized only): entry_number, handler_name, dog_name, dog_breed, plus the evaluated scoresheet result:
  - total_points / max_points + percent
  - final_tier: Excellent / Very Good / Good / Sufficient / Insufficient
  - passed (bool) + reason if failed (trial NQ / below threshold / too many insufficients)
  - per-exercise breakdown: code, name, points, max, tier
  - applied modifiers (rare; L2 TRK Lifeline is the v1 case)

Judge identity is **not** shown publicly — that's an internal concern. The published result stands on its own.

## URL shape

Guessable, slug-based:

- `/` — public events index (or a marketing landing page; designer's call)
- `/events/:slug` — event page
- `/events/:slug/trials/:trial_id` — trial leaderboard
- `/events/:slug/trials/:trial_id/entries/:entry_id` — entry detail

## Screens needed

### 1. Public events index

Lists upcoming and recent published events. Most recent / soonest first.

- Each card: event name, location, start–end dates, status badge ("upcoming" / "live now" / "complete").
- A "live now" event (one where any trial has `status = in_progress`) should be visually prominent — that's what spectators are most likely looking for.
- Empty state: "No public events yet" with a brief explanation.
- This page may double as the marketing landing if the club doesn't have a separate site, so leave room for a hero section above the events list.

### 2. Event page

Header + per-trial summary.

- **Header**: event name, location, dates, overall status. If the event is live (some trial in progress), surface a "Live now: OB-L1" pill linking straight to that trial's leaderboard.
- **Trials list**: grouped by trial_date, each row showing discipline + level (e.g., "Obedience — Level 1"), trial date, status, and a one-line summary (e.g., "12 of 18 entries scored — leader: 92% Excellent" while in progress; "Complete — 14 passed of 18 entered" when done).
- Click a trial → leaderboard.
- Empty: "Schedule coming soon" if event has no trials yet.

### 3. Trial leaderboard — the spectator's main screen

The page people refresh during the trial.

**Header**:
- Trial info: event name (link back), discipline + level + date
- Status: in progress (with auto-refresh indicator) vs complete
- Counts: "12 of 18 finalized" while running; "All scored" when complete
- Sort toggle: by score (descending, default) vs by entry number (chronological — useful for spectators following along live)

**Leaderboard body**:
- A ranked list of finalized entries. Each row:
  - Rank (1, 2, 3… for passing entries; ties share a rank)
  - Entry number, handler name, dog name (breed as a secondary detail)
  - Total points / max + percent
  - Tier badge (Excellent green, Very Good blue/teal, Good neutral, Sufficient amber, Insufficient red — same palette as the judge UI)
  - Pass / NQ indicator
  - Row links to the entry detail page
- NQ'd entries are listed at the bottom under a divider (no rank). The reason (Trial NQ / below threshold / too many insufficients) is visible inline.
- Auto-refresh: while the trial is in progress, the leaderboard polls every ~15s. New finalizations animate in with a subtle highlight so refresher-watchers notice changes.

**Empty / partial states**:
- Trial pending, no entries finalized: "Scoring starts soon. Results will appear here as the judge finalizes each run."
- Trial in progress, 0 finalized: same message + live-refresh indicator.

**Search / filter** (nice-to-have, not v1-critical): a text search box that filters to entries matching handler or dog name. Useful for big events; degrade gracefully if omitted.

### 4. Entry detail

The page handlers share with their friends. Full breakdown of one finalized entry.

**Header**:
- Entry number, handler, dog name + breed
- Trial / event context (linked back)
- The big result: total points / max, percent, final tier, pass / fail badge

**Body**:
- Per-phase sections. Each phase header: name, phase total, phase tier (if applicable — depends on whether phase totals are scored independently).
- Within a phase, each exercise: code + name, points / max, tier badge. Optionally expandable to show criterion-level breakdown (designer's call — clean default vs. enthusiast detail).
- If any auto-triggers fired: highlighted callout ("Trial NQ: dog left field") with the rulebook context if known.
- If any modifiers applied: line showing the modifier and its effect ("Lifeline applied: −20 points, tier capped at Very Good").

**Social / sharing**:
- Open Graph tags on this page in particular — handlers share these links. The OG image and description should let a "92% Excellent" result render well on Facebook / iMessage / Twitter.
- Subtle "Share" button is optional; the URL itself is the share artifact.

## Cross-cutting UX

- **Tier as semantic color**, consistent with the judge UI — same palette so a spectator who's seen the judge's tablet feels at home. Especially important on the leaderboard where the tier badges do most of the visual sorting.
- **No login affordance**: the public site shouldn't surface "sign in" / "admin" links in the main nav. (Admins know to go to `/login` directly.) Optional: a small footer link.
- **Empty / loading states** matter more here than in admin UI — a trial sitting at 0 entries during setup is the first thing many visitors see. The message should reassure ("results will appear as runs are scored") rather than feel broken.
- **Print stylesheet** (low priority): some clubs print the final leaderboard. Plain-text-readable fallback on the trial page would be appreciated.
- **Accessibility**: public-facing, so screen-reader correctness matters. Tier should not be conveyed by color alone (badge text + color, not color alone).
- **Performance**: most of these pages are mostly-static after entries finalize. Designer doesn't need to plan for it, but heavy client-side animation is unnecessary — these pages should feel fast on a phone over LTE at a remote training field.

## Out of scope

- Admin CRUD (separate prompt — `admin-crud.md`).
- Judge scoring UI (separate prompt — `judge-scoring.md`).
- Login / authentication flows (already exist).
- Score editing / dispute flows — finalized is finalized in v1.
- Embedded / iframeable widgets for clubs to put on their own sites — future iteration.
- Spectator accounts, favorites, follow-a-handler features — future iteration.

## Deliverables

Wireframes or high-fidelity mockups for:
1. Public events index — populated and empty states
2. Event page — pre-trial, live (with one trial in progress), complete
3. **Trial leaderboard** (priority): live state with mixed finalized + still-to-come, plus complete state with NQs separated at the bottom
4. Entry detail — a passing result (Very Good) and a failed result (Trial NQ), to show both visual treatments

Mobile (phone) and desktop layouts for screens 3 and 4 specifically — that's where spectators and handlers actually live.
