# K9 Trials — Demo Walkthrough

A ~10-minute script for demoing the app to a client. It follows the full
loop — a competitor registers, an admin accepts, a judge scores, the
competitor sees the result and disputes it, the admin resolves it.

## Before the meeting

```bash
mage seeddemo     # wipes + reseeds ./data/app.db with the demo world
mage dev          # build + run the server (http://localhost:8080)
```

`mage seeddemo` resets to a known state — safe to re-run anytime, including
mid-demo if you want a clean slate. It **wipes all data**, so never point it
at a database you care about.

Open four browser tabs (or use one and log in/out). Logins:

| Role | Email | Password |
|------|-------|----------|
| Admin | `admin@example.com` | `admin1234` |
| Judge | `judge@example.com` | `demo1234` |
| Competitor | `ltanaka@example.com` | `demo1234` |
| (more competitors) | `rokafor@`, `khessel@`, `syi@`, `dfowler@example.com` | `demo1234` |

The demo world: 5 competitors with dogs, 4 events (one live, one upcoming,
one closed with results, one draft), scored entries (qualifying and NQ),
registrations awaiting review, and challenges in every state.

## The story (suggested order)

### 1. The public face — no login

- **`/events`** — the events anyone sees. Note the draft event (Cedar Creek
  Summer) is *not* listed; only published events are public. Filter by
  discipline.
- Open **Cedar Creek Spring → its OB L1 trial** — a live leaderboard with
  real scores, qualifying placements, and NQ runs.
- **`/competitors`** — search the directory. Open **L. Tanaka** → her dogs
  and event history. Click a dog (Vex) → its public record.

> Talking point: every result is computed by the scoring engine from the
> judge's logged inputs, not stored as a number — so it stays correct across
> rulebook revisions.

### 2. The competitor — `ltanaka@example.com`

- **`/account`** — dashboard: her dogs, recent results, an open challenge.
- **Dogs** → add/edit a dog (call name, breed, registration number).
- **Entries** — every entry across events: upcoming, in progress, finalized,
  and *pending registrations* she's filed. Filter by status.
- **Register for an event**: open **Hopkins Mill Tracking → Register** →
  pick a dog, pick a trial, submit. It becomes a **pending** registration —
  show it now appears on her Entries list as "pending review."

### 3. The admin — `admin@example.com` → `/admin`

- **Dashboard** — what needs attention: pending registrations and open
  challenges, plus live events and drafts.
- **Events** → open **Cedar Creek Spring**: edit settings, manage trials
  (grouped by day, with judge + entry counts).
- **Registrations** — the handoff. Find a **pending** row and click
  **Accept**. This creates the entry (assigns the next entry number) — the
  competitor's pending item is now a real entry, and it shows up in the
  judge's queue.
- **Judges** — assign a judge to a trial that has none.
- **Challenges** — open the queue, pick a dispute, **Start review**, then
  **Resolve** with a note. Point out: resolving closes the dispute; the
  actual re-score happens through the judge flow (the audit chain stays
  intact).
- **Users** — change a role inline (you can't demote yourself).

### 4. The judge — `judge@example.com` → `/judge`

- The scoring queue for the live trial. Open an entry and walk the
  scoresheet — gate, per-exercise scoring, review, submit. Finalizing locks
  the score.

> Close the loop: the score the judge just finalized is what the competitor
> sees on their entry, the public leaderboard reflects it, and the
> competitor can challenge it — which lands back in the admin's queue.

## Tips

- To reset between runs (or if you make a mess): `mage seeddemo` again.
- Real point scores only render for **OB L1** trials — that's the one
  rulebook template wired into the scoring engine so far. Other disciplines
  demo the registration/listing flows; their scoresheets come online as more
  templates are added.
