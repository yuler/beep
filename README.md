# beep

Monorepo for the beep project, based on the MonoSolo template.

## Struct

```bash
.
├── core/                # Rails backend (API, Background Jobs & Core Logic)
├── apps/                # Client applications (Consumer-facing)
│   ├── web/             # TanStack Router (Primary web application / SaaS product)
│   ├── admin/           # Vite + Vue3 (Internal dashboard & B2B management)
│   ├── mobile/          # Mobile application (Capacitor/React Native)
│   └── desktop/         # Desktop application (Tauri/Electron)
├── packages/            # Shared workspace packages
│   ├── api-client/      # Auto-generated SDK/Hooks for Rails API
│   ├── shared-utils/    # Shared TypeScript types, constants & helpers
│   └── ui-kit/          # Design system & shared component library
├── scripts/             # Global automation (Setup, CI/CD, Dev-ops)
├── package.json         # Root workspace manifest & global task runner
└── README.md            # Project documentation & beep philosophy
```

## Core (Rails)

Opinionated Rails 8.1 template for rapid development.

- **Solid:** Solid Queue, Solid Cache, and Solid Cable on SQLite
- **Multi-tenancy:** Shared slug namespace `/{slug}/...` for personal and team accounts (see [docs/core/ACCOUNT.md](docs/core/ACCOUNT.md), [AccountSlug::Extractor](core/config/initializers/account_slug.rb))
- **API:** JWT-authenticated controllers

## 🛠️ Development

```bash
mise setup
mise dev
```

`mise dev` prints local subdomain URLs on start, then runs [`Procfile.dev`](Procfile.dev) via overmind (`*.localhost` resolves to `127.0.0.1` — no `/etc/hosts` needed):

| App  | Subdomain URL                       | Plain localhost       |
| ---- | ----------------------------------- | --------------------- |
| web  | http://web.beep.localhost:3000  | http://localhost:3000 |
| core | http://core.beep.localhost:3001 | http://localhost:3001 |

Ports come from root `.env` (`CORE_PORT`, `WEB_PORT`). Rails allows any `*.localhost` host in development; Vite allows `.localhost` via `allowedHosts`.

Login: `john@example.com`

## Docker Compose (local quick start)

```bash
cp .env.example .env
# Set SECRET_KEY_BASE (`cd core && bin/rails secret`).

cp compose.example.yml compose.yml
docker compose pull    # or: docker compose build
docker compose up -d
```

Compose forces Mode B (host-only cookies; Nitro proxies `/api` → core), so a Mode A `mise` `.env` is fine to reuse.

- Web: http://localhost:3000 (TanStack Start Node SSR; `/api` proxied to core)
- Core: http://localhost:3001 (optional direct access)

Images: `ghcr.io/yuler/beep` and `ghcr.io/yuler/beep-web` (see `core-push-image.yml` / `web-push-image.yml`).
