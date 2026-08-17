# AGENTS.md

Guidelines for coding agents in this beep monorepo.

## Layout

- [`core/`](core/) — Rails 8.1 backend → [`docs/core/DEVELOP.md`](docs/core/DEVELOP.md), [`docs/core/STYLE.md`](docs/core/STYLE.md), [`docs/core/ACCOUNT.md`](docs/core/ACCOUNT.md)
- [`apps/web/`](apps/web/) — TanStack Router → [`.agents/web.md`](.agents/web.md)
- `apps/` — other client apps (admin, mobile, desktop)
- `packages/` — shared TypeScript packages
- `scripts/` — monorepo setup / automation

## Development

```bash
mise setup
mise dev
```

`mise dev` → `scripts/dev.sh` prints subdomain URLs, then starts [`Procfile.dev`](Procfile.dev) via overmind:

- Core: http://core.${APP_HOST}:${CORE_PORT}
- Web: http://web.${APP_HOST}:${WEB_PORT}

`*.localhost` resolves to `127.0.0.1` (no hosts file). Login: `john@example.com`. Development is host-locked to `web.${APP_HOST}` / `core.${APP_HOST}`; `localhost` and any wrong `*.localhost` fail fast with a page showing the canonical URL.

Local CORS for web → core `/api/v1` is development-only: [`core/config/initializers/development_cors.rb`](core/config/initializers/development_cors.rb). Details: [`docs/core/DEVELOP.md`](docs/core/DEVELOP.md#local-cors-development-only).

## Rules

- Git commit title format: `emoji [scope] The main change` — example: `✨ [core] Adopt shared account slug tenancy for personal and team`
- Do not add agent trailers to commits (`Made-with:`, `Co-Authored-By: Claude`, Cursor, etc.). Message body only when it adds real context.
- PR title follows the same format as the git commit title.
- Markdown tables must be auto-aligned (pad columns so pipes line up).
- Do not use superpower or other speculative-driven skills unless explicitly declared.
- If something is unclear, ask questions. Keep everything from design to code as simple as possible.

## Core

- API JSON responses use jbuilder views (`.json.jbuilder`), not inline hashes in controllers.
- Do not add gems or other package dependencies without asking first.
