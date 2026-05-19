# Design prompt: Judge scoring

Design the score-entry UI for a K9 trial scoring application. The user is a judge actively scoring a competitor's run — handler + dog working through a structured scoresheet at a trial field. The judge enters scores in real time as the dog performs each exercise.

## Critical context

- **Device**: tablet (iPad-class, landscape primary). Touch input only.
- **Environment**: outdoor, possibly bright sunlight; the judge may be standing, walking, holding a clipboard in the other hand. Hit targets must be generous and high-contrast.
- **Mental state**: the judge is watching a live dog, not the screen — UI must be glanceable, with point selections and totals obvious at arm's length.
- **High stakes**: a misfired disqualification can void a competitor's title attempt. Disqualification triggers must be visually distinct from normal scoring and require an explicit confirmation step.

## Tech context

Go + templ + htmx + Alpine.js + Tailwind. Each judge tap is a server round-trip. No offline queue. No client-side scoring logic — the server is authoritative for all totals and tier bands. Designs must work as htmx partial swaps (no SPA framework).

## Domain model

A judge scores one **Entry** at a time. An entry belongs to a **Trial** which determines the **Scoresheet template** for a given (discipline, level, rulebook version).

A scoresheet is structured:

```
Scoresheet                      (e.g., L1 Obedience, 120 points)
└── Phase                       (e.g., "Phase 1: Muzzle & Stability", 30 points)
    └── Exercise                (e.g., "1.1 Muzzle Acceptance", 10 points)
        ├── Criteria            (positive-points line items, used by CriteriaSum exercises)
        │   └── e.g., "1.1.a Muzzle accepted at start line" (0–2 points)
        ├── Penalty events      (deductions, used by PenaltyLedger exercises)
        │   └── e.g., "Missed hide" (−5 pts), logged by occurrence count
        └── Auto-triggers       (disqualifications)
            └── e.g., "Dog destroys muzzle" — fires once; scope: exercise / phase / trial
```

Each exercise is exactly one Kind:

- **CriteriaSum** — judge enters point values per criterion. Exercise score = sum of criteria. Most common.
- **PenaltyLedger** — exercise starts at MaxPoints. Each tap on a penalty event subtracts that event's deduction. Floors at 0.
- **Aggregate** — derived from sibling exercises' scores. Read-only for the judge (no direct entry). Rare; L2/L3 Detection only.

Each criterion has its own MaxPoints (small integer, typically 1–5). Each exercise has a MaxPoints which is the sum of its criteria's maxes. The exercise score is banded into a **Tier**: Excellent / Very Good / Good / Sufficient / Insufficient. Bands are percentage-based on the exercise's own max, so a 5-pt and a 25-pt exercise have different integer cutoffs.

The scoresheet has scoresheet-wide modifiers in some cases (e.g., L2 Tracking "Lifeline" = −20 pts + tier cap at Very Good). Modifier UI shows in the finalize step, not during scoring.

## Screens needed

### 1. Judge dashboard

Lists trials the judge is assigned to. Each entry's status visible: `registered` (not started), `scoring` (in progress, draft state), `finalized` (locked).

- "Next up" prominently highlighted — the lowest-numbered un-finalized entry the judge is assigned to.
- Filter or group by trial / status.
- One tap to open the entry pre-run view.

### 2. Entry pre-run view

Shown after the judge picks an entry, before they tap "Start scoring."

- Entry header: entry number, handler name, dog name, breed, trial info (discipline + level + date).
- Scoresheet preview: phase names + their max points, collapsed.
- Total points possible for this scoresheet.
- "Start scoring" — primary action.
- "Back to entries" — secondary.

### 3. Entry scoring view — THE central screen

The judge spends most of their time here. Auto-saves every tap.

**Persistent header** (always visible while scrolling):
- Entry number, handler, dog
- Discipline + level
- Running scoresheet total: `points / max`, percentage, current tier band (with color)
- Pass-condition indicator (at-a-glance "would pass" / "would fail" given current scores)
- Pause / back button (returns to entry pre-run; scoresheet auto-saves continuously)

**Body, organized by phase.** Three plausible patterns — designer's call which fits best for tablet landscape:
- (a) Vertical scroll with sticky phase headings (one long page)
- (b) Horizontal stepper / tabs across phases (one phase visible at a time)
- (c) Accordion (one phase expanded at a time)

Within a phase, each exercise renders as a card containing:

- **Exercise header**: code + name (e.g., "1.1 Muzzle Acceptance & Heeling Pattern"), running exercise total `points / max`, current tier band.

- **For CriteriaSum exercises**:
  - Each criterion as a row: code (e.g., `1.1.a`), description, current selection, and a row of point-stepper buttons (one button per integer from 0 to criterion.MaxPoints — NOT a numeric input field). Selected value visually emphasized.
  - Example: a 2-point criterion shows three buttons: 0, 1, 2.
  - Tapping a value overwrites the previous selection (latest-write-wins).

- **For PenaltyLedger exercises**:
  - Banner: "Starts at X points."
  - Each penalty event as a button: event name + deduction value (e.g., `Missed hide  −5`). Tap to log an occurrence. A counter shows occurrence count. Running exercise score updates.
  - Undo last occurrence action available.

- **For Aggregate exercises**:
  - Read-only display: name, current value, max, tier. Note that the value is derived from sibling exercises (e.g., "Calculated from search-1 + search-2").

- **Auto-triggers** (when the exercise has any): visually distinct from criterion entry. Located at the bottom of the exercise card OR behind an "Advanced" disclosure to prevent accidental taps. Tap → confirmation modal showing the trigger's scope ("This will NQ the exercise / phase / trial — are you sure?"). After firing, the exercise card shows a banner: "NQ FIRED — exercise zeroed: [trigger name]" with an undo affordance while the entry is still in `scoring` (draft) status.

**Persistent footer**:
- Running scoresheet total (duplicate of header, for easy bottom-of-screen visibility while thumb-scrolling)
- "Finalize entry" — primary action, opens screen 4.

### 4. Finalize confirm

Modal or full-screen review before locking the entry.

- Final point total + percentage + final tier (after any modifier caps applied)
- Pass / fail result with the reason (e.g., "PASS — 87%, 0 Insufficients" or "FAIL — Trial NQ: dog left field")
- Per-exercise summary table: exercise code, name, points / max, tier band
- If the scoresheet template has available modifiers (rare — only L2 TRK Lifeline in v1): a checklist to apply any modifier here, with live preview of the final tier
- "Finalize" — primary, locks the entry (`status=finalized`), makes it publicly visible
- "Back to scoring" — secondary

## Cross-cutting UX

- **Hit targets**: minimum 44×44px. Stepper buttons spaced so a gloved or sun-warmed finger does not fat-finger adjacent values.
- **Tier as semantic color**: Excellent (green), Very Good (cool blue or teal), Good (neutral), Sufficient (amber), Insufficient (red). Use consistently for exercise / phase / scoresheet totals so judges learn the color shorthand.
- **Auto-trigger styling**: clearly differentiated from criterion entry. Red border, warning icon, confirmation required. Fired triggers visually take over the exercise card so the judge cannot miss the state.
- **Latest-write-wins**: tapping a different criterion value overwrites the previous one visually. (Server records it as a new append-only row; the UI shows only the most recent.)
- **No save buttons within exercises**: every tap auto-submits via htmx. Subtle "saving…" indicator briefly during the round-trip.
- **Network failure**: if a tap fails, show an inline retry near the affected control. Don't lose the judge's intent or silently revert.
- **Reachability**: with iPad in landscape, keep frequently-tapped controls (point steppers, "next phase" navigation) in the lower half of the screen where thumbs rest.

## Out of scope

- Admin CRUD (separate prompt — `admin-crud.md`).
- Public spectator results page.
- Multi-judge concurrent scoring (one judge per entry in v1).
- Offline / queued submission (future iteration; v1 requires connectivity).
- The scoring engine math — implemented and treated as a black box that returns updated totals after each tap.

## Deliverables

Wireframes or high-fidelity mockups for:
1. Judge dashboard
2. Entry pre-run view
3. Entry scoring view (priority): full layout, plus zoomed-in details of
   - one CriteriaSum exercise card
   - one PenaltyLedger exercise card
   - the auto-trigger confirmation modal
4. Finalize confirm screen

Show both a **fresh entry** state (nothing scored yet) and a **mid-run** state (some exercises fully scored, one partially scored, one penalty event logged) — the mid-run state is where the running totals and tier indicators do most of their work.
