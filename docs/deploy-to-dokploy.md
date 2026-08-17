# Deploy to Dokploy

Production is **Mode B**: one public hostname, host-only session cookies, Nitro proxies `/api` to Rails. Do not deploy this monorepo as a single Dokploy **Application** (one container). Use **Compose**.

```
Browser  →  Traefik (Dokploy)  →  web:3000
                                      └── /api  →  http://core:80
Solid Queue  →  worker (`bin/jobs`) on the same SQLite volume as core
```

| Piece    | Image                    | Public? | Role                               |
| -------- | ------------------------ | ------- | ---------------------------------- |
| `web`    | `ghcr.io/yuler/beep-web` | Yes     | TanStack Start SSR; Traefik target |
| `core`   | `ghcr.io/yuler/beep`     | No      | Rails API (Thruster on port 80)    |
| `worker` | `ghcr.io/yuler/beep`     | No      | `bin/jobs` (Solid Queue)           |

Local `mise dev` stays **Mode A** (`web.*` / `core.*`). That setup is [docs/core/DEVELOP.md](core/DEVELOP.md). Do not paste a Mode A `.env` into Dokploy.

## Files

| File                   | In git? | Purpose                                                    |
| ---------------------- | ------- | ---------------------------------------------------------- |
| `compose.dokploy.yml`  | Yes     | Compose stack Dokploy runs                                 |
| `.env.dokploy.example` | Yes     | Shape of the Environment tab (no secrets)                  |
| `.env.dokploy`         | No      | Local production secrets (`cp .env .env.dokploy`)          |
| `.env.local`           | No      | Overrides for `mise dev`                                   |

Dokploy writes the Environment tab to a `.env` **next to the compose file on the server**. `compose.dokploy.yml` loads that file into `core` and `worker` via `env_file: .env`. It then forces Mode B:

- `SESSION_COOKIE_DOMAIN` empty
- `SOLID_QUEUE_IN_PUMA=false`
- `CORE_INTERNAL_URL=http://core:80` on `web`

`VITE_CORE_URL` is baked empty into the published web image ([`.github/workflows/web-push-image.yml`](../.github/workflows/web-push-image.yml)). Do not rebuild web on the VPS unless you pass the same build args.

## 1. Images

On push to `main` (and on `v*` tags), GitHub Actions publish:

- `ghcr.io/yuler/beep:<tag>`
- `ghcr.io/yuler/beep-web:<tag>`

Default tag in compose is `main` (`BEEP_IMAGE_TAG`). Pin a git tag (for example `v1.2.3`) for production if you do not want floating `main`.

If the packages are private, add a Dokploy **Registry**:

- Registry: `ghcr.io`
- Username: GitHub username
- Password: PAT with `read:packages`

## 2. Create the Compose service

In the Dokploy project / production environment:

1. **Add service → Compose** (Docker Compose, not Stack, not Application).
2. **Source:** GitHub → this repo → branch you want to track.
3. **Compose path:** `compose.dokploy.yml`.
4. **Environment:** paste `.env.dokploy` (or fill from `.env.dokploy.example`).
5. **Domains:** service `web`, container port `3000`, HTTPS / Let's Encrypt.
6. Deploy.

The public hostname must equal `SITE_DOMAIN` (no scheme), for example `beep.yuler.cc`.

Stop or delete any leftover **Application** that was bound to the same domain.

## 3. Environment

Minimum:

```bash
SITE_DOMAIN=beep.yuler.cc
SITE_NAME=beep
BEEP_IMAGE_TAG=main
SECRET_KEY_BASE=          # cd core && bin/rails secret
```

SQLite paths (defaults are fine):

```bash
DB_NAME=storage/production.sqlite3
DB_NAME_CACHE=storage/production_cache.sqlite3
DB_NAME_QUEUE=storage/production_queue.sqlite3
DB_NAME_CABLE=storage/production_cable.sqlite3
```

Also set SMTP, `MISSION_DASHBOARD_*`, and VAPID keys as needed. Generate VAPID with `mise run core` → `create-vapid-key`.

Do **not** set:

| Variable                 | Why                                    |
| ------------------------ | -------------------------------------- |
| `SESSION_COOKIE_DOMAIN`  | Parent-domain cookies are Mode A only  |
| `SOLID_QUEUE_IN_PUMA`    | Worker already runs `bin/jobs`         |
| `VITE_CORE_URL`          | Must stay empty in the **image build** |
| `APP_HOST` / `CORE_PORT` | Local Mode A only                      |

## 4. Domain and TLS

- Attach the domain only to **`web`**, port **3000**.
- Do not publish core (`:80`) or worker.
- `web` is on `dokploy-network` so Traefik can reach it; `core` and `worker` stay on the compose network.
- If Cloudflare sits in front, SSL/TLS mode **Full (strict)** — not Flexible.

## 5. Data and backups

`core` and `worker` share the named volume `core-storage` → `/rails/storage`. That is the database (SQLite + Solid Queue/Cache/Cable) and Active Storage.

Use Dokploy **Volume Backups** on `core-storage` (named volumes only; bind mounts are not backed up). Do not scale `worker` replicas while on SQLite.

`core` runs `db:prepare` on boot via [`core/bin/docker-entrypoint`](../core/bin/docker-entrypoint).

## 6. After deploy

- Open `https://$SITE_DOMAIN`. Browser calls same-origin `/api/v1`.
- Magic-link mail uses `SITE_DOMAIN` in `action_mailer.default_url_options`.
- Jobs: one `worker`; poller interval is in [`core/config/recurring.yml`](../core/config/recurring.yml).
- Redeploy after a new GHCR tag, or set `BEEP_IMAGE_TAG` and redeploy.

## Local vs Dokploy

|              | Local (`mise dev`)                         | Dokploy                                 |
| ------------ | ------------------------------------------ | --------------------------------------- |
| Env file     | `.env.local` overrides `.env`              | Compose Environment tab → server `.env` |
| API          | CORS to `core.*` (Mode A)                  | Nitro proxy, same origin (Mode B)       |
| Cookies      | `SESSION_COOKIE_DOMAIN=.${APP_HOST}`       | Host-only                               |
| Queue        | `SOLID_QUEUE_IN_PUMA=true`                 | Separate `worker`                       |
| Compose file | `compose.yml` / `compose.example.yml`      | `compose.dokploy.yml`                   |

Keep production secrets in `.env.dokploy` (gitignored). Refresh it with `cp .env .env.dokploy` only when `.env` is already the Dokploy Mode B file — never copy a Mode A `mise` `.env` into Dokploy.
