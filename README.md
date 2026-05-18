# Advanced Template

Project template for Firefly Software **Advanced tier** client projects. Superset of the standard tier — adds a persistence layer, migrations, session-backed auth, and a working login flow.

## Stack

- **Go 1.25** stdlib `net/http` router
- **templ** for server-side rendered components
- **Tailwind CSS** via standalone CLI
- **htmx** + **Alpine.js** (vendored)
- **SQLite** via `modernc.org/sqlite` (pure Go, no CGo) — [PostgreSQL upgrade path](POSTGRES.md)
- **sqlc** for type-safe query generation
- **goose** for migrations
- **bcrypt** password hashing, cookie-based sessions
- **Postmark** for transactional email
- **Cloudflare Turnstile** for contact form spam protection
- **Mage** build system

## Getting Started

### Prerequisites

Open in the devcontainer — all tools are installed automatically. Otherwise install manually:

- Go 1.25+
- [templ](https://templ.guide)
- [goose](https://github.com/pressly/goose)
- [sqlc](https://sqlc.dev)
- [Mage](https://magefile.org)

### Setup

```bash
cp .env.example .env        # edit with your values
mage installtailwind        # download Tailwind standalone CLI
mage migrateup              # run database migrations
mage seed admin@example.com yourpassword  # create an admin user
```

### Development

```bash
# Terminal 1: watch and rebuild Tailwind CSS
mage dev

# Terminal 2: run the server
go run ./cmd/server
```

### Production Build

```bash
mage build        # compiles Tailwind, generates templ + sqlc, builds Go binary
./bin/server
```

## Project Structure

```
advanced-template/
├── cmd/
│   ├── server/           # main application entry point
│   └── seed/             # CLI tool for creating admin users
├── internal/
│   ├── config/           # env-based config loader
│   ├── handler/          # HTTP handlers (home, contact, auth)
│   ├── mail/             # Postmark email client
│   ├── middleware/        # request logging
│   ├── session/          # session middleware + context helpers
│   ├── store/            # database query wrappers
│   ├── view/             # templ components and layouts
│   └── db/               # sqlc generated code (gitignored)
├── migrations/           # goose SQL migrations
├── queries/              # sqlc SQL query definitions
├── web/static/           # CSS, JS, images
├── tailwind/             # Tailwind config + standalone CLI
├── sqlc.yaml
├── magefile.go
├── Dockerfile
├── docker-compose.yml
└── POSTGRES.md
```

## Mage Targets

| Target | Description |
|---|---|
| `mage build` | Full production build (CSS + templ + sqlc + Go binary) |
| `mage buildcss` | Compile Tailwind CSS |
| `mage dev` | Tailwind watch mode |
| `mage generate` | Run `templ generate` and `sqlc generate` |
| `mage migrateup` | Run all pending migrations |
| `mage migratedown` | Roll back the last migration |
| `mage migratestatus` | Show current migration state |
| `mage createmigration <name>` | Scaffold a new migration file |
| `mage seed <email> <password>` | Create an admin user |

## Auth

### Routes

| Route | Purpose |
|---|---|
| `GET /login` | Render login form |
| `POST /login` | Validate credentials, create session, redirect |
| `POST /logout` | Destroy session, clear cookie, redirect |

The login form uses htmx (`hx-post`, `hx-swap="outerHTML"`) for inline error feedback without a full page reload.

### Creating Users

There is no registration flow — admin users are created via the seed CLI:

```bash
mage seed admin@example.com yourpassword
```

Or directly:

```bash
go run ./cmd/seed admin@example.com yourpassword
```

### Protecting Routes

Wrap any handler with `session.RequireAuth` to redirect unauthenticated users to `/login`:

```go
mux.Handle("GET /admin", session.RequireAuth(handler.AdminDashboard()))
```

For a group of routes, wrap the sub-mux:

```go
admin := http.NewServeMux()
admin.Handle("GET /admin/dashboard", handler.Dashboard())
admin.Handle("GET /admin/settings", handler.Settings())

mux.Handle("/admin/", session.RequireAuth(admin))
```

### Accessing the Current User

The session middleware runs on every request and attaches the user to the context when a valid session cookie is present. Access it from any handler:

```go
func Dashboard() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        user := session.FromContext(r.Context())
        if user != nil {
            // user.ID, user.Email are available
        }
    }
}
```

`session.FromContext` returns `nil` for unauthenticated requests — safe to call on public pages (e.g. to show/hide a nav login link).

### Sessions

- Stored in SQLite (`sessions` table), persist across server restarts
- Cookie: `session_token`, HttpOnly, Secure, SameSite=Lax, 7-day expiry
- Passwords hashed with bcrypt via `golang.org/x/crypto`

## Middleware

All middleware lives in `internal/middleware/` and follows the `func(http.Handler) http.Handler` pattern. They compose by wrapping — outermost runs first.

### Logging

Wraps every request with a structured JSON log line (method, path, status, duration, request ID).

```go
srv := middleware.Logging(logger)(mux)
```

### CORS

Handles cross-origin requests. Preflight (OPTIONS) gets `Allow-Methods`/`Allow-Headers`/`Max-Age`; actual requests get `Allow-Origin` only. Disallowed origins receive no CORS headers (browser-enforced).

```go
srv = middleware.CORS(middleware.CORSConfig{
    AllowedOrigins:   []string{"https://example.com"},
    AllowedMethods:   []string{"GET", "POST"},
    AllowedHeaders:   []string{"Content-Type", "Authorization"},
    AllowCredentials: true,
    MaxAge:           3600,
})(srv)
```

Use `"*"` in `AllowedOrigins` to allow any origin.

### Rate Limiting

Per-client token bucket throttling. Each unique client gets its own bucket with the configured rate (requests/sec) and burst (max instant). Returns `429 Too Many Requests` with `Retry-After` when exceeded.

```go
srv = middleware.RateLimit(middleware.RateLimitConfig{
    Rate:           10,    // 10 requests/sec steady state
    Burst:          20,    // allow bursts up to 20
    TrustedProxies: 1,     // 1 = behind Caddy (reads X-Forwarded-For)
    CleanupInterval: 5 * time.Minute,
})(srv)
```

**`TrustedProxies`**: Set to the number of reverse proxies in front of the app. `0` uses `RemoteAddr` directly (no proxy). `1` trusts the rightmost `X-Forwarded-For` entry (Caddy on Hetzner). `2` for CDN + Caddy.

**Custom key function**: Rate limit by something other than IP (e.g. API key):

```go
middleware.RateLimitConfig{
    Rate:  5,
    Burst: 5,
    KeyFunc: func(r *http.Request) string {
        return r.Header.Get("X-API-Key")
    },
}
```

### CSRF Protection

Signed double-submit cookie pattern. A token is issued on every GET and validated on POST/PUT/DELETE. Uses HMAC-SHA256 to prevent cookie tampering.

```go
srv = middleware.CSRF(middleware.CSRFConfig{
    Secret: []byte(cfg.SessionSecret), // must be >= 32 bytes
})(srv)
```

**In forms** — embed the token as a hidden field. In templ components:

```go
// In your handler, pass the token to the template:
token := middleware.Token(r)

// In your .templ file, add a hidden input:
<input type="hidden" name="csrf_token" value={ token }/>
```

Or use the `TemplateField` helper in Go `html/template`:

```go
template.HTML(middleware.TemplateField(r))
```

**With htmx** — send the token in a header. Configure globally:

```html
<body hx-headers='{"X-CSRF-Token": "{{ token }}"}'>
```

Or per-element with `hx-headers`.

**Config options:**
- `FieldName` — form field name (default: `"csrf_token"`)
- `HeaderName` — header name (default: `"X-CSRF-Token"`)
- `InsecureDev` — set `true` for local HTTP without TLS
- `ErrorHandler` — custom 403 response

### Panic Recovery

Catches panics in handlers, logs the stack trace, and returns a 500 instead of crashing the server. Safe with WebSocket upgrades and partial responses.

```go
srv = middleware.Recovery(middleware.RecoveryConfig{})(srv)
```

With a custom error page:

```go
srv = middleware.Recovery(middleware.RecoveryConfig{
    ErrorHandler: func(w http.ResponseWriter, r *http.Request, val any) {
        w.WriteHeader(http.StatusInternalServerError)
        view.ErrorPage().Render(r.Context(), w)
    },
})(srv)
```

Edge cases handled automatically:
- **`http.ErrAbortHandler`** — re-panics (intentional connection abort, not a bug)
- **WebSocket upgrades** — logs but writes no HTTP response (would corrupt the connection)
- **Headers already sent** — logs but skips the error response (can't change status mid-stream)

### Middleware Order

In `main.go`, middleware wraps inside-out. A typical stack:

```go
srv := session.Middleware(st)(mux)      // innermost: attach user to context
srv = middleware.CSRF(csrfCfg)(srv)     // CSRF after session (needs cookie access)
srv = middleware.RateLimit(rlCfg)(srv)  // rate limit before heavy work
srv = middleware.CORS(corsCfg)(srv)     // CORS before rate limit (preflight is cheap)
srv = middleware.Recovery(recCfg)(srv)  // catch panics before they reach the logger
srv = middleware.Logging(logger)(srv)   // outermost: log everything including 500s
```

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Server listen port |
| `DB_PATH` | `./data/app.db` | SQLite database file path |
| `SESSION_SECRET` | — | Secret for signing session tokens |
| `POSTMARK_SERVER_TOKEN` | — | Postmark API token |
| `POSTMARK_FROM` | — | Sender email address |
| `POSTMARK_TO` | — | Recipient email address |
| `GTAG_ID` | — | Google Analytics measurement ID |
| `PIXEL_ID` | — | Facebook Pixel ID |
| `TURNSTILE_SITE_KEY` | — | Cloudflare Turnstile site key |
| `TURNSTILE_SECRET_KEY` | — | Cloudflare Turnstile secret key |

## Deployment

Built for deployment on Hetzner behind Caddy via Docker Compose. See `Dockerfile` and `docker-compose.yml`.
