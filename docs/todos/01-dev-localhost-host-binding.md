# Bind local `*.localhost` Host to this project

Local Mode A uses `web.<app>.localhost:3000` and `core.<app>.localhost:3001`. `*.localhost` all resolve to loopback, and **TCP only cares about the port**. Whichever process holds 3000 answers every Host on that port.

Vite currently allows any `.localhost` (`apps/web/vite.config.ts`). Rails development allows any `*.localhost` (`core/config/environments/development.rb`). A wrong URL still loads the running app.

## Symptom that triggered this

Opened `http://web.monosolo.localhost:3000/` while **beep** was on 3000/3001. The landing page was beep. Sign-in `POST http://core.beep.localhost:3001/api/v1/session` returned:

```json
{"error":"Invalid cross-origin request","code":"INVALID_CROSS_ORIGIN"}
```

That JSON is **CSRF**, not the CORS middleware:

- [`Api::V1::BaseController`](../../core/app/controllers/api/v1/base_controller.rb) rescues `ActionController::InvalidCrossOriginRequest`
- [`RequestForgeryProtection`](../../core/app/controllers/concerns/request_forgery_protection.rb) uses Rails `protect_from_forgery using: :header_only`
- Browser sends `Sec-Fetch-Site: cross-site` when the page origin is not same-site as `core.beep.localhost`
- [`development_cors.rb`](../../core/config/initializers/development_cors.rb) still allows any `*.localhost`, so the 422 body is visible to JS

Same 422 for `http://localhost:3000` → `core.beep.localhost:3001`. `http://web.beep.localhost:3000` is `same-site` and succeeds. Cookie domain is `.beep.localhost`, so a monosolo Host would not get the session cookie even if CSRF passed.

| Page origin                             | `POST /api/v1/session` to core.beep | Result                  |
| --------------------------------------- | ----------------------------------- | ----------------------- |
| `http://web.beep.localhost:3000`        | same-site                           | 200                     |
| `http://web.monosolo.localhost:3000`    | cross-site                          | 422 INVALID_CROSS_ORIGIN |
| `http://localhost:3000`                 | cross-site                          | 422 INVALID_CROSS_ORIGIN |

## What to do (beep)

Lock Host first. Unique ports are optional. No Caddy / `/etc/hosts` — they do not bind a Host to a process on a shared port.

- [ ] Single source of truth for the local site, e.g. `APP_HOST=beep.localhost` (or derive from `SITE_NAME`)
- [ ] Vite `allowedHosts`: `web.${APP_HOST}` only (optional extra: `localhost`, documented as cookie/CSRF-unsafe for Mode A login)
- [ ] Rails `config.hosts` in development: `core.${APP_HOST}`, plus loopback; drop the catch-all `*.localhost` regex (keep trycloudflare if still needed)
- [ ] `scripts/dev.sh` prints only the canonical URLs (`web.beep.localhost` / `core.beep.localhost`)
- [ ] Wrong Host (`web.monosolo.localhost:3000` while beep owns 3000) must fail fast at Vite/Rails, not serve this app

Optional, separate concern (two projects running at once):

- [ ] Give other repos different default `WEB_PORT` / `CORE_PORT` (e.g. monosolo 3100/3101). Does not replace Host allowlist.

monosolo (and any sibling that copied this stack) needs the same Host lock, or the mistake can still happen the other way.

## Out of scope

- Changing CSRF (`header_only`) or production Mode B (same-origin `/api` proxy)
- Broadening CORS to paper over a wrong Host
