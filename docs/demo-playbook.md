# K9 Elements — Demo Playbook & Feature Walkthrough

A client-facing guide to the working K9 Elements application: what it does, and
a set of scripted **user stories** you can click through live to demonstrate the
capabilities and flow. Every story is grounded in the built-in demo dataset, so
the names, dogs, events, and scores below are exactly what you will see on
screen.

---

## Running the demo

1. **Start the app** with demo mode on (seeds a complete, realistic world on an
   empty database):

   ```bash
   DEMO_MODE=1 go run ./cmd/server     # serves on http://localhost:8080
   ```

   Or reset an existing demo database to the known state at any time:

   ```bash
   mage seeddemo
   ```

   When `DEMO_MODE=1`, the admin dashboard also has a **"↺ Reset demo data"**
   button — one click wipes everything and reseeds, so you can re-run a story
   from a clean slate mid-demo.

2. **Log in** at `/login`. All demo passwords are `demo1234`, except the admin.

| Role | Login | Password | Notes |
|------|-------|----------|-------|
| **Admin** | `admin@example.com` | `admin1234` | Full organizer surface |
| **Judge** | `judge@example.com` | `demo1234` | Assigned to the live Cedar Creek trial |
| **Judge** | `jpereira@example.com` | `demo1234` | Judged the closed Brindle Bay trial |
| **Competitor** | `ltanaka@example.com` | `demo1234` | @ltanaka — 3 dogs (Vex, Kestrel, Nyx), most active |
| **Competitor** | `rokafor@example.com` | `demo1234` | @rokafor — 2 dogs (Atlas, Echo) |
| **Competitor** | `khessel@example.com` | `demo1234` | @khessel — 1 dog (Saber) |
| **Competitor** | `syi@example.com` | `demo1234` | @syi — 1 dog (Lumen), first season |
| **Competitor** | `dfowler@example.com` | `demo1234` | @dfowler — **no dogs yet** (fresh-account persona) |

### The demo world at a glance

| Event | When | Status | What it shows |
|-------|------|--------|---------------|
| **Cedar Creek Spring Trial** | May 2026 | Published | A trial **mid-run** — finalized scores, a run still scoring, open registrations, and active challenges |
| **Hopkins Mill Tracking** | Jun 2026 | Published | Accepted entries **not yet run** — the registration & withdrawal surface |
| **Brindle Bay Autumn** | Oct 2025 | Closed | A **completed** event with final results and a resolved challenge |
| **Cedar Creek Summer** | Jul 2026 | **Draft** | An **unpublished** event — the "notify me when it opens" surface |

Scoring uses the **Level 1 Obedience** rulebook: 120 points max, **84 (70%)
to qualify**. In the live Cedar Creek trial that means Echo (104), Saber (98)
and Kestrel (95) qualify, while Atlas (74) is an NQ — which is exactly why
@rokafor filed a challenge on it.

---

## Part 1 — Feature inventory

Grouped by who uses it. Each line points at where to see it in the demo.

### Public site (no login required)
- **Event listing & detail** (`/events`) — published and closed events, their
  trials, schedule, and location. *See: Cedar Creek, Hopkins Mill, Brindle Bay.*
- **Live trial leaderboard** (open a trial from an event) — standings sorted by
  score, qualifying runs above an **NQ divider**, color-coded result tiers, and
  a **live auto-refresh** while a trial is in progress. *See: Cedar Creek → Obedience L1.*
- **Public scoresheet** (`/entries/{id}`) — the full per-exercise breakdown for
  any finalized run, with its Q/NQ verdict.
- **Competitor directory** (`/competitors`) — searchable list of handlers, each
  with a public profile: bio, dogs, and trial history. *See: @ltanaka, @rokafor.*
- **Dog profiles** (`/dogs/{id}`) — per-dog registry info and chronological
  trial history. *See: Vex, Atlas.*
- **Contact form** with spam protection.

### Accounts & onboarding
- **Self-service signup** (`/signup`) — a handler creates an account and claims a
  unique **@handle**, checked for availability as they type.
- **Login / logout**, role-aware (competitors land on their account, admins on
  the operations dashboard, judges on their scoring queue).

### Competitor account
- **Dashboard** — what's upcoming, recent results, and open challenges at a glance.
- **Profile editor** — display name, bio, public handle.
- **Dog roster** — add/edit dogs with call name, registered name, breed, date of
  birth, **sex**, and registration number.
- **Entries list** — every run across all events with status filter chips
  (Upcoming / In progress / Finalized / Withdrawn) and live counts.
- **Entry detail** — the competitor's own scoresheet, the qualifying threshold,
  the **"challenge this score" affordance** (open for 7 days after finalizing),
  and the **"request withdrawal"** action for upcoming runs.
- **Challenges** — file a dispute against a finalized score (with a scoresheet
  excerpt of what's being disputed) and track it through to resolution.

### Registration
- **Register for an event** — pick a dog, pick the trials to enter, submit; the
  request lands in the organizer's review queue.
- **"Notify me"** — for an event that hasn't opened registration yet, a
  logged-in competitor can subscribe to be emailed the moment it opens.

### Judge scoring
- **Scoring queue** — the runs assigned to a judge, in order.
- **Guided scoring flow** — a gate/eligibility step, per-exercise scoring against
  the rulebook, a review screen, a submit-and-lock step, and a locked read-only
  view. Scores are **append-only** and audit-tracked.

### Admin / organizer
- **Operations dashboard** — a "what needs attention" board: pending
  registrations and open challenges, the live events and unpublished drafts, a
  **quick-actions** card, and a cross-event **recent-activity feed**.
- **Events** — searchable list with status filters (Draft / Published / Closed /
  Archived); create and edit events with a slug-availability check, an **audit
  block** (created / published / last-edited), and an **archive lifecycle**.
- **Trials** — per-event trial management grouped by day, with a **slide-over
  new-trial form** (pill-chip discipline/level selectors) and an "N trials
  without a judge" flag.
- **Registration review** — accept (which assigns an entry number and creates the
  scoring entry), waitlist, or reject; **confirm withdrawal requests**.
- **Judge assignments** — assign a judge per trial and notify the judging panel.
- **Challenge queue** — review disputes with filtering, sorting, an audit
  timeline, and start-review / resolve / dismiss actions.
- **Users & roles** — search users and change roles (competitor / judge / admin).

> **Demo notes / honest stubs:** outbound email (notify-me, notify-judges,
> registration confirmations) is **logged to the server console, not actually
> sent** — the delivery backend is intentionally not wired yet. Everything else
> is live functionality against a real database.

---

## Part 2 — User-story playbook

Eight scripted walkthroughs. Run them in this order for a natural narrative, or
cherry-pick. Each lists the persona, the goal, the click-path, and what it
demonstrates. Tip: keep the **server console visible** for the email-stub stories.

---

### Story 1 — A new handler gets set up
**Persona:** @dfowler (fresh account, no dogs) · **Goal:** join and enter a trial.

1. Log in as `dfowler@example.com`. The account dashboard greets an empty roster.
2. Go to **Dogs → Add a dog**. Fill in call name, breed, date of birth, and the
   **Sex** field; save.
3. Go to **Events**, open **Hopkins Mill Tracking**, and click **Register**.
4. Pick the new dog, check the **Tracking L1** trial, add a note, and submit.
5. Open **My entries** — the new entry shows as **Pending review**.

*Demonstrates:* self-onboarding, the dog roster (incl. the new sex field), event
discovery, and the registration flow that feeds the organizer's queue.

> Want to show **signup from scratch**? Use `/signup` with a brand-new email and
> watch the live @handle availability check, then continue from step 2.

---

### Story 2 — The organizer works the registration queue
**Persona:** Admin · **Goal:** turn requests into scoreable entries.

1. Log in as `admin@example.com` → the **dashboard** shows pending registrations
   under "Needs review."
2. Open **Cedar Creek Spring → Registrations**. Requests are grouped by trial.
3. On **Obedience L2**, **Accept** @ltanaka's Vex — note it's assigned the next
   **entry number** and becomes an entry. **Waitlist** or **Reject** others to
   show those paths.
4. (Optional) Accept @dfowler's Hopkins Mill request from Story 1.

*Demonstrates:* the registration "bridge" — accept assigns a number and creates
the scoring entry; waitlist/reject are one click each; the dashboard counters
update.

---

### Story 3 — A judge scores a run
**Persona:** Judge (`judge@example.com`) · **Goal:** finalize a live run.

1. Log in as the judge → the **scoring queue** lists the Cedar Creek Obedience L1
   runs assigned to them (Vex is mid-score).
2. Open **Vex** → step through the **gate**, then **score** each exercise against
   the L1 Obedience criteria.
3. Continue to **review**, confirm, and **submit** — the run **locks** and a final
   score is produced.
4. Open the same trial's **public leaderboard** in another tab and refresh —
   Vex now appears in the standings.

*Demonstrates:* the guided, rulebook-driven judging flow; append-only,
audit-tracked scoring; and the result flowing straight to the public leaderboard.

---

### Story 4 — A competitor reviews a score and disputes it
**Persona:** @ltanaka · **Goal:** see results and file a challenge.

1. Log in as `ltanaka@example.com` → **My entries**. Filter to **Finalized**.
2. Open **Vex's Brindle Bay** run (a qualifying 110/120) — the **scoresheet**
   shows the per-exercise breakdown, the 84-point threshold, and the **Q** verdict.
3. Within the 7-day window, click **Challenge this score**, pick a reason, and
   submit — note the scoresheet **excerpt** of what's being disputed.
4. Visit **Challenges** to see it tracked; existing demo challenges show the
   other states (Kestrel's is *under review*, Saber's Brindle Bay one is
   *resolved* with the judge's notes).

*Demonstrates:* the competitor-facing scoresheet, the time-boxed challenge
window, and the dispute lifecycle from the handler's side.

---

### Story 5 — The organizer resolves a challenge
**Persona:** Admin · **Goal:** adjudicate a dispute. *(Pairs with Story 4.)*

1. As admin, open **Challenges** — the queue shows open and in-review disputes
   across all events, with filters and sorting.
2. Open @rokafor's **open** challenge on **Atlas** (the 74/120 NQ at Cedar Creek).
   The detail panel shows the disputed run, the **NQ reason excerpt**, and an
   **audit timeline**.
3. Click **Start review**, then **Resolve** (or **Dismiss**) with notes.
4. Back in the competitor account, the filer now sees the updated status.

*Demonstrates:* the full challenge-review surface — context, audit trail, and the
resolve/dismiss decision — closing the loop opened in Story 4.

---

### Story 6 — A competitor withdraws an entry
**Persona:** @ltanaka + Admin · **Goal:** the admin-confirmed withdrawal flow.

1. As `ltanaka@example.com`, open **My entries** and select **Kestrel at Hopkins
   Mill** (accepted, not yet run).
2. Click **Request withdrawal** — the page now reads **"Withdrawal requested ·
   pending admin."** The entry isn't voided yet.
3. Log in as admin → **Hopkins Mill → Registrations**. Kestrel's row shows a
   **"Withdrawal requested"** badge and a **Confirm withdrawal** button.
4. Click **Confirm**. The registration becomes **Withdrawn** — and the **entry
   number is retained** on the record for the audit history.
5. Back as @ltanaka, the entry now shows **Withdrawn**, and a **Withdrawn** filter
   chip appears on the entries list.

*Demonstrates:* withdrawal as a **request routed to the organizer**, not an
instant cancel — with the number retained for auditability.

---

### Story 7 — "Notify me" for an upcoming event
**Persona:** Any competitor + Admin · **Goal:** subscribe, then trigger the open.

1. As a logged-in competitor, go directly to
   **`/events/cedar-creek-summer/register`** (Cedar Creek Summer is still a draft,
   so it isn't in the public list — open it by link).
2. The page shows **"Not yet open."** Click **Notify me** — it confirms you're on
   the list.
3. Log in as admin → **Events → Cedar Creek Summer → Edit** → set status to
   **Published** and save.
4. Watch the **server console**: a line logs the registration-opened notification
   and its recipients.

*Demonstrates:* event subscriptions and the **publish-transition hook** that
notifies subscribers (delivery stubbed to the log for now).

---

### Story 8 — The organizer's event lifecycle & dashboard
**Persona:** Admin · **Goal:** the full operations surface.

1. As admin, start on the **dashboard** — the **recent-activity feed** (finalized
   runs, accepted registrations, filed challenges, published events), the
   **needs-review** counters, and **quick actions**.
2. Open **Events**: try the **search** box and the status **filter chips**.
3. Open **Cedar Creek Spring → Edit**: show the **audit block** (created /
   published / last-edited) and the fuller at-a-glance (judge coverage, total
   entries).
4. **Archive** the closed **Brindle Bay** event — it leaves the public listings,
   and the **Archived** chip on the events list now has a count. **Restore** it to
   show the lifecycle is reversible.
5. Open **Cedar Creek Spring → Trials**: click **+ New trial** to open the
   **slide-over** form with **pill-chip** discipline/level selectors; note the
   "trials without a judge" flag.
6. Open **Judge assignments**: assign a judge to an unjudged trial and use
   **Notify judges** (logged to the console).

*Demonstrates:* the complete organizer toolkit — dashboard, event lifecycle with
audit & archive, trial setup, and judge coordination.

---

### Spectator path (no login) — quick opener or closer
Browse **`/events`** → open **Cedar Creek Spring** → open the **Obedience L1**
trial for the **live leaderboard** (qualifiers above the NQ divider, auto-
refreshing) → click a name into a **competitor profile** → into a **dog profile**
and its trial history. A clean, zero-friction way to open or close a demo.

---

## Resetting between runs
Stories 2, 3, 6, and 7 change data. To return to the pristine demo world, click
**↺ Reset demo data** on the admin dashboard (DEMO_MODE), or run `mage seeddemo`.
Logins and the world above are restored exactly.
