# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

A K9-trials scoring application being built on top of the Firefly Software **Advanced Tier** project template (module `github.com/flintcraftstudio/k9-trials`). Most of the surface area today (auth, contact form, Postmark, layouts, middleware) is the unmodified template — that boilerplate is documented exhaustively in `README.md`.

The K9-specific work currently lives only as a **design spec** in `docs/`, which is a *separate Go module* (`github.com/dirtybandana/k9elements`, `go 1.22`) containing a `scoring` package with full doc comments but bodies of `panic("not implemented")`. That spec is the source of truth for the scoring domain model — read `docs/types.go`, `docs/template.go`, `docs/evaluate.go`, `docs/tiers.go` before touching scoring logic. `docs/l1ob.go` is a fully-fleshed example template (Level 1 Obedience) showing how `ScoresheetTemplate` is meant to be populated.

The two modules do not yet import each other. When implementing scoring, the expected path is to lift `docs/` into `internal/scoring/` (or wherever it lands) under the main module rather than treat `docs/` as a vendored dependency.

## Common commands

Build / generate / run (via Mage):

```bash
mage installtailwind        # one-time: download standalone Tailwind CLI
mage generate               # templ generate + sqlc generate
mage buildcss               # compile Tailwind to web/static/css/site.css
mage build                  # buildcss + generate + go build -o ./bin/server ./cmd/server
mage dev                    # full Build then run ./bin/server (NOT a watcher despite the README)
go run ./cmd/server         # dev loop: run directly without rebuilding the binary
```

Database (goose against SQLite at `$DB_PATH`, default `./data/app.db`):

```bash
mage migrateup
mage migratedown
mage migratestatus
mage createmigration <name>
mage seed <email> <password>   # creates an admin user via cmd/seed
```

Note: `cmd/server/main.go` *also* runs `goose.Up` at startup, so migrations apply automatically on boot — the `mage migrate*` targets are only needed for inspection or rollback.

Tests:

```bash
go test ./...                                       # everything in the main module
go test ./internal/middleware/...                   # one package
go test ./internal/middleware -run TestCSRF         # one test
go test -race ./...                                 # race detector
cd docs && go test ./...                            # the scoring spec module is separate
```

There is no linter configured beyond `go vet`. The CI workflow lives in `.github/workflows/deploy.yml`.

## Architecture

### HTTP wiring (`cmd/server/main.go`)

Single `http.ServeMux` registered with stdlib method-prefixed routes (`GET /`, `POST /login`, …). The middleware stack is currently minimal:

```
session.Middleware(store) → middleware.Logging(logger) → mux
```

`internal/middleware/` contains additional middleware (CORS, CSRF, RateLimit, Recovery) with tests, but they are **not wired into `main.go` yet**. If you add CSRF or rate limiting, follow the ordering documented in `README.md` ("Middleware Order"). Graceful shutdown is already implemented: SIGINT/SIGTERM trigger `server.Shutdown` with a 10s deadline.

### Sessions and auth

- `internal/session/` — cookie-based sessions, 7-day expiry, HttpOnly+Secure+SameSite=Lax cookie named `session_token`. `session.Middleware` reads the cookie, looks the row up via the `Store` interface, and attaches `*User` to the request context.
- `session.FromContext(ctx)` returns `nil` for anonymous requests — safe to call on public pages.
- `session.RequireAuth(handler)` redirects unauthenticated callers to `/login`.
- No registration flow. Users are created via `cmd/seed` (`mage seed email password`) which writes a bcrypt hash to the `users` table.

### Data layer

- SQLite via `modernc.org/sqlite` (pure Go, no CGo). `PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;` are set on every boot.
- `migrations/` — goose SQL files. `queries/` — sqlc input queries.
- `sqlc.yaml` generates into `internal/db/` (gitignored — run `mage generate` to create it).
- `internal/store/store.go` currently hand-writes the queries it needs (sessions, users) using `database/sql` directly. **The sqlc-generated `internal/db/` package is not yet wired up.** When adding new tables, prefer adding sqlc queries under `queries/` and generating, then wrapping them in `store`.
- PostgreSQL migration path is documented in `POSTGRES.md` — driver swap + sqlc engine flip + goose dialect change.

### Views

`internal/view/` holds templ components (`.templ` files compile to `_templ.go` via `templ generate`). Process-global view state — `GtagID`, `PixelID`, `TurnstileSiteKey` — is set once at startup in `main.go` and consumed by templates via package vars in `internal/view/shared.go`. The site name is hardcoded in `view.SiteName`.

### Domain model (scoring, in `docs/`)

The scoring package separates **templates** (rulebook spec — what's possible at a given Discipline+Level) from **concrete scoresheets** (what a specific team ran on a specific day) and **inputs** (what the judge logged). `EvaluateScoresheet` is a pure function `(inputs, concrete, template) → result`. Key invariants from the spec:

- `ExerciseTemplate.Kind` is a discriminated union (`CriteriaSum` / `PenaltyLedger` / `Aggregate`) — exactly one of `Criteria`/`Events`/`AggregateOf` is populated per kind.
- `TemplateVersion` is stamped on every scoresheet so historical scores stay interpretable across rulebook revisions.
- Storage is append-only at the input layer; evaluation is latest-write-wins per `(scoresheet, exercise, criterion)`.
- `AutoTrigger` scopes (`AutoNQExercise` / `AutoNQPhase` / `AutoNQTrial`) cascade differently — trial NQ bypasses the point math.
- Point rounding is "round half up" (§3.2) via `RoundPoints`, distinct from Go's `math.Round` for negative values.

When implementing scoring, follow the evaluation order documented on `EvaluateScoresheet` (cascade inflows → non-aggregate exercises → aggregates → AutoNQPhase zeroing → totals → modifiers → percent/tier/passed).

## Project-specific conventions

- Module path is `github.com/flintcraftstudio/k9-trials`. The repo was forked from the Firefly Advanced Template and renamed — if you see stray references to the old `firefly-software-mt/advanced-template` path (the README still uses the old name in places), they're cosmetic and don't affect builds.
- `internal/view/` package vars (`GtagID`, `PixelID`, `TurnstileSiteKey`) are intentional process-globals. Set once in `main.go`, never reassigned at request time.
- `cmd/server/main.go` includes a hand-rolled `loadEnv` that reads `.env` into `os.Setenv` *without overriding* existing env vars — this is intentional so container env wins over the dotfile.
- `web/static/css/site.css` is a build artifact (gitignored). Always run `mage buildcss` after touching `tailwind/input.css` or any `.templ` file with new utility classes.
- Firefly-specific guidance for code style, architecture, and review lives in `.claude/skills/firefly-stack-agent/`, `.claude/skills/firefly-architect-agent/`, and `.claude/skills/firefly-review-agent/` — consult these for stack-wide conventions (Go + templ + htmx + Alpine + Tailwind, with Svelte for complex islands).
