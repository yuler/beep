# Develop

Guidelines and commands for agents working in the Rails 8.1 `core` app.

## Development Commands

### Setup and Server

```bash
# Initial setup (installs deps)
bin/setup
# Start development server (from core/)
bin/dev
```

Development URLs (login with `john@example.com`, `APP_HOST` from repo-root `.env` / `.env.local`, default `beep.localhost`):

| URL                                      | Notes                                              |
| ---------------------------------------- | -------------------------------------------------- |
| http://core.${APP_HOST}:${CORE_PORT} | Preferred local subdomain (host-locked; required) |

`localhost` is deliberately refused (Mode A login from loopback is cookie/CSRF-unsafe). Use `core.${APP_HOST}`.

From the monorepo root prefer `mise setup` / `mise dev` (runs [`scripts/dev.sh`](../../scripts/dev.sh): prints subdomain URLs, then `overmind start -f Procfile.dev`).

### Local subdomains (`.localhost`)

Modern OS/browsers resolve `*.localhost` to `127.0.0.1` — no `/etc/hosts` entry required.

| App  | Subdomain                                        | Port env                    |
| ---- | ------------------------------------------------ | --------------------------- |
| web  | http://web.${APP_HOST}:${WEB_PORT}          | `WEB_PORT` (default `3000`) |
| core | http://core.${APP_HOST}:${CORE_PORT}        | `CORE_PORT` (default `3001`) |

Development is **host-locked**: [`config/environments/development.rb`](../../core/config/environments/development.rb) accepts only `core.${APP_HOST}` (plus trycloudflare) and Vite accepts only `web.${APP_HOST}`. Plain `localhost` and any other `*.localhost` host are refused — a wrong host fails fast with a page telling you the canonical URL, instead of silently serving beep. Multi-tenancy remains path-based (`/{slug}/...`), not subdomain-based. Blocked hosts on core render a custom 403 page (`config/initializers/development_host_authorization.rb`).

### Local CORS (development only)

`apps/web` and Core run on different origins locally (`web.*` vs `core.*`, or `localhost:3000` vs `localhost:3001`). Browser `fetch` to `/api/v1` needs CORS.

[`config/initializers/development_cors.rb`](../../core/config/initializers/development_cors.rb) registers `DevelopmentCors` **only when `Rails.env.development?`**. Test and production do not load it.

Allowed origins (credentials enabled):

| Origin                                  | Notes                         |
| --------------------------------------- | ----------------------------- |
| `http(s)://web.${APP_HOST}` (any port)  | Canonical web host (only)     |

`localhost` / loopback origins are refused — same host lock as `config.hosts`.

Rails has no built-in “allow CORS” config switch. This middleware is **development-only**. In production, TanStack Start (Nitro) proxies `/api` to core same-origin (Mode B).

### Testing

```bash
# Run unit tests (fast)
bin/rails test
# Run single test file
bin/rails test test/path/file_test.rb
# Run system tests (Capybara + Selenium)
bin/rails test:system
# Run full CI suite (style, security, tests)
bin/ci

# For parallel test execution issues, use:
PARALLEL_WORKERS=1 bin/rails test
```

### Database

```bash
# Load fixture data
bin/rails db:fixtures:load
# Run migrations
bin/rails db:migrate
# Drop, create, and load schema
bin/rails db:reset
# Drop, create, async schema to sqlite schema
ruby script/db_schema_fresh.rb
```

### Other Utilities

```bash
# Manage Solid Queue jobs
bin/jobs
# Deploy (requires 1Password CLI for secrets)
bin/kamal deploy
```

## Architecture Overview

### Multi-Tenancy (URL-Based)

Canonical design & vocabulary: [`ACCOUNT.md`](ACCOUNT.md).

URL path-based multi-tenancy via middleware:

- Personal and team Accounts share one slug namespace; URLs are `/{slug}/...` for both
- Middleware ([`AccountSlug::Extractor`](../../core/config/initializers/account_slug.rb)) mounts via `SCRIPT_NAME`, looks up the Account, sets `Current.account`; missing slug → 404
- Global routes (no slug) keep `Current.account` nil; do not fall back to personal as tenant truth
- Authz (login / membership) lives in controllers — unauthenticated → login; non-member → 404
- All tenant models include `account_id` for data isolation
- Background jobs automatically serialize and restore account context

This avoids subdomains or separate databases, which keeps local development and testing simpler.

### Authentication & Authorization

Passwordless magic-link authentication:

- Global `Identity` (email-based) can have `Users` in multiple Accounts
- Users belong to an Account and have roles: owner, admin, member, system
- Sessions managed via signed `session_id` cookies (browser) and optional Bearer tokens (mobile/CLI)
- Board-level access control via `Access` records

**Mode B (the only supported deployment):** the browser talks only to `apps/web` (TanStack Start / Node + Nitro). Rails-core is not exposed to the browser. Every `/api/v1` call originates on the Node server:

- **Server functions** (`src/server/*` → `createServerFn`) handle mutations and reads called from the browser; `src/server/core.ts` forwards the request `Cookie` header so core sees the same `session_id` as the user, and relays `Set-Cookie` back onto the browser response for session start/end.
- **SSR** (`beforeLoad` / `loader`) resolves data on Node during the document request; the session cookie lives on the web origin, so auth guards work server-side.
- Sessions use the HttpOnly `session_id` cookie (no bearer token in localStorage).

| Deployment | `VITE_CORE_URL`           | `SESSION_COOKIE_DOMAIN`  | API access                               |
| ---------- | ------------------------- | ------------------------ | ---------------------------------------- |
| Production | `""` (empty; unused)      | web-origin host cookie    | Node server fns → core (server-to-server) |

`VITE_CORE_URL` is retained only for `coreAppUrl()` external links (Mission Control, letter opener); it is not the browser's API target. Session cookies use `SameSite=Lax`; CSRF on core's JSON API uses header-only protection, which treats the server-to-server calls (no `Sec-Fetch-Site`) as allowed (`RequestForgeryProtection`).

### Core Domain Models

**Account** → The tenant/organization

- Has users, boards, cards, tags, webhooks
- Has entropy configuration for auto-postponement

**Identity** → Global user (email)

- Can have Users in multiple Accounts
- Session management tied to Identity

**User** → Account membership

- Belongs to Account and Identity
- Has role (owner/admin/member/system)
- Board access via explicit `Access` records

### UUID Primary Keys

All tables use UUIDs (UUIDv7, base36-encoded as 25-char strings):

- Custom fixture UUID generation keeps deterministic ordering in tests
- Fixtures are always older than runtime records
- `.first` / `.last` work correctly in tests

### Background Jobs (Solid Queue)

Database-backed job queue (no Redis):

- Development and production use `:solid_queue` with a dedicated `queue` DB (`storage/*_queue.sqlite3`)
- Locally set `SOLID_QUEUE_IN_PUMA=true` (see root `.env.example`) so the supervisor runs inside Puma
- Jobs automatically capture/restore `Current.account`
- Mission Control::Jobs for monitoring; apps/web staff pages at `/admin/jobs` and `/admin/stats`

Key recurring tasks (via [`config/recurring.yml`](../../core/config/recurring.yml)):

- `BeepPollerJob` every 10 seconds (due `once` beeps)
- Production: cleanup of finished Solid Queue jobs

### Chrome MCP (Local Dev)

App URL: http://core.${APP_HOST}:${CORE_PORT}  
Login: `john@example.com` (passwordless magic link — check Rails console for the link)

Use Chrome MCP tools against the running dev app for UI testing and debugging.

## Code Style Guidelines

See [`STYLE.md`](STYLE.md).
