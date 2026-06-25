# Architecture: Multi-role accounts (per-trial roles)

Status: **Approved, ready for implementation.** Authored by the architect agent.
This is a boundary/contract handoff for the stack agent — not implementation code.

## Problem

Judges are frequently also competitors who enter their own dogs in *other*
trials. The original schema stored a **single** `users.role` string
(`CHECK IN ('admin','judge','competitor')`), and the auth middleware matched it
exclusively. Consequences today:

- `competitor()` gate = `RequireRole("competitor","admin")` → a `judge`-role
  account is locked out of `/account/dogs`, `/account/entries`, registration.
- `RequireJudge` = `RequireRole("judge","admin")` → a `competitor`-role account
  can't reach `/judge`.

So a judge literally cannot enter their own dog. We fix this by treating role as
**additive capabilities** plus **per-trial assignment**, never as one exclusive
string.

## What already exists (and stays)

- **Competitor identity** (`competitors` table) links to a user via nullable
  `UNIQUE user_id` — already independent of role. **Unchanged.**
- **Per-trial judge assignment** via `entries.judge_id` (bulk-assigned trial-wide
  by `AssignTrialJudge`; admin UI at `/admin/events/{id}/assignments`).
  **Kept as the source of truth** (no new `trial_judges` table).
- **Scoring engine** (`internal/scoring/`) is pure; `judged_by` is recorded at
  the append-only input layer. **Unchanged.**

## Decisions

| # | Decision | Choice |
|---|----------|--------|
| 1 | Capability storage | **`user_roles` join table** `(user_id, capability)` |
| 2 | Competitor baseline | **Universal** — every authenticated account is a competitor |
| 3 | Competitor row | **Implicit** — never stored; `user_roles` holds only `judge`/`admin` grants |
| 4 | Conflict-of-interest | **Warn only** — advisory at assign-time, non-blocking |
| 5 | Trial→judge source of truth | **Keep `entries.judge_id`** |
| 6 | `users.role` after backfill | **Drop** (SQLite table rebuild) |

## Three concepts

1. **Account capabilities** (global, additive): `competitor` baseline (implicit),
   plus `judge` and `admin` grants. A person holds any combination.
2. **Per-trial judge assignment** (`entries.judge_id`): authoritative grant to
   score a specific trial's entries. Only judge-eligible accounts are assignable.
3. **Computed authorization**: surface access asks "do I hold capability X?";
   scoring a *specific* entry additionally asks "am I that entry's assigned
   judge?" (admins exempt), with a warn-only COI advisory at assign-time.

## Boundary map

```
[ Account Capabilities ]   NEW — replaces users.role single-string
  Owns:    Which global capabilities a user holds (competitor implicit;
           judge/admin stored). Additive, not exclusive.
  Uses:    DB access (user_roles)
  Used by: Auth middleware, Admin user-role handler

[ Auth / Session ]         CHANGED — exclusive match → capability check
  Owns:    Session validation + capability checks on the request user
  Uses:    Account Capabilities
  Used by: HTTP layer (route gates)

[ Trial Judge Assignment ] EXISTS — gains eligibility check, COI advisory,
                           per-entry authority guard
  Owns:    Who judges a trial/entry; only judge-eligible are assignable
  Uses:    DB access (entries.judge_id), Account Capabilities, Competitor identity
  Used by: Admin assignment handler, Judge scoring handlers

[ Competitor Identity ]    UNCHANGED
[ Scoring engine ]         UNCHANGED
[ DB access / Store ]      schema migration only (user_roles, drop users.role)
```

## Contracts

### Account Capabilities (NEW)

One-line: a user holds a set of additive capabilities; `competitor` is the
universal implicit baseline, `judge`/`admin` are explicit stored grants.

Surface (store):
```go
func (s *Store) UserCapabilities(ctx context.Context, userID int64) ([]string, error)
func (s *Store) GrantCapability(ctx context.Context, userID int64, cap string) error   // idempotent
func (s *Store) RevokeCapability(ctx context.Context, userID int64, cap string) error  // idempotent
```

Guarantees:
- Capabilities are a set — additive, order-independent, no exclusivity.
- `competitor` is implicit for every authenticated user; never stored.
  `user_roles` holds only `judge` and `admin`.
- Grant/revoke are idempotent (no-op on duplicate/absent).

Caller rules: only the admin user handler grants/revokes; read once per request
at session load.

Must not: touch `entries.judge_id`, `competitors`, or scoring tables; never
store a `competitor` row.

Owned files: `migrations/018_create_user_roles.sql` (new — 014–017 were
already taken when this landed), `queries/users.sql`, `internal/store/store.go`,
generated `internal/db/`.

### Auth / Session (CHANGED)

One-line: authorization checks whether the request user *holds* a capability,
not whether a single role *equals* one.

Surface:
```go
// session.User gains:
Caps []string
func (u *User) Has(cap string) bool      // membership
func (u *User) IsAdmin() bool            // Has("admin")
func (u *User) IsJudge() bool            // Has("judge")  — judge-eligible
func (u *User) IsCompetitor() bool       // u != nil      — baseline

func RequireAuth(next http.Handler) http.Handler                       // any logged-in user = competitor
func RequireCapability(next http.Handler, caps ...string) http.Handler // holds ANY cap; admin always passes
func RequireAdmin(next http.Handler) http.Handler                      // RequireCapability(next, "admin")
func RequireJudge(next http.Handler) http.Handler                      // RequireCapability(next, "judge")
```

Guarantees: wrapped handler runs only if the user satisfies the requirement;
`admin` is a superset satisfying every capability check. `RequireAuth` is the
new competitor gate.

Caller rules: in `main.go`, `competitor(...)` becomes `RequireAuth` (drop the
`RequireRole("competitor","admin")` wrapper); judge/admin routes unchanged in
spelling. Session middleware loads `Caps` when resolving the cookie user.

Must not: read `users.role` (going away); make per-entry authority decisions.

Owned files: `internal/session/session.go`, `internal/session/session_test.go`,
wiring in `cmd/server/main.go`.

### Trial Judge Assignment (EXISTS — extended)

One-line: a trial's entries carry the assigned judge; only judge-eligible
accounts are assignable, and scoring an entry requires being that entry's
assigned judge.

Guarantees:
- Admin may only assign a user holding the `judge` capability.
- COI is advisory: at assign-time, if the candidate judge handles any dog
  entered in that trial, return a non-blocking ⚠ warning. Assignment proceeds.
- A judge may score/finalize a specific entry only if
  `entry.judge_id == currentUser.ID` (admins exempt). `RequireJudge` gates the
  surface; this gates the row.

Caller rules:
- `AdminAssignJudge`: verify target holds `judge` before writing `judge_id`;
  compute + surface COI warning; do not block on COI.
- Judge scoring handlers (`/judge/entry/{id}/score|review|submit`): after
  `RequireJudge`, load the entry and reject with 403 if `judge_id` isn't the
  current user (admin bypass).

Must not: grant global capabilities; hard-block on COI.

Owned files: `AdminAssignJudge`/`AdminAssignments` handler, the `Judge*`
scoring handlers, `queries/entries.sql` (existing).

## Implementation order (one commit each)

1. **DB/Store** — `migrations/018_create_user_roles.sql`. Backfill: `judge`→row,
   `admin`→row, `competitor` implicit (no row). Capability queries + store
   wrappers. ✅ DONE. (The `users.role` drop was deferred to step 2 — session/
   store still read the column, so dropping it in step 1 would break the build.)
2. **Auth/Session** — `User.Caps` + `Has`/`IsAdmin`/`IsJudge`/`IsCompetitor`;
   `RequireAuth` + `RequireCapability`; swap `competitor()` → `RequireAuth` in
   `main.go`. Update `session_test.go`. Once session/store no longer read
   `users.role`, add a migration that **drops `users.role`** (SQLite table
   rebuild) — deferred here from step 1.
3. **Admin user handler** — `POST /admin/users/{id}/role` becomes grant/revoke.
4. **Trial Judge Assignment** — eligibility check + COI warn at assign-time;
   per-entry `judge_id` authority guard in `Judge*` scoring handlers.

Each boundary depends on the one above it; implement and review in order.
