# agy-todo: unify Channel terms in code

Handoff from grill-me. **Do the code changes.** Docs already match the glossary (except leftovers listed below).

Source of truth: [`docs/TERMS.md`](docs/TERMS.md). Also updated: [`docs/architecture/channel.md`](docs/architecture/channel.md), [`docs/architecture/web-push.md`](docs/architecture/web-push.md).

Do not use the word **Device** in UI, API, TERMS, or domain code (no `kind=device`, no `/device/inbox`, no “Device Channel”).

Branch: `feature/channel`, latest doc commit `c07f82a`.

---

## Confirmed glossary

| Term             | Meaning                                                                                                           |
| ---------------- | ----------------------------------------------------------------------------------------------------------------- |
| Channel          | One user-owned destination row (`channels`). Beep targets Channel IDs, not kinds.                                 |
| kind             | Adapter. Now: `cli`, `email`, `web_push`, `webhook`. Planned (TERMS only, not enum): `ios`, `android`, `desktop`. |
| CLI Channel      | `kind=cli`. Pulled by `beep up` via `GET /cli/inbox`.                                                             |
| Beep CLI         | The `beep` binary (`apps/cli`). More than inbox pull.                                                             |
| Email Channel    | `kind=email`. Auto-create one default row per user (account email).                                               |
| Web Push Channel | `kind=web_push`. One browser subscription = one row.                                                              |
| Webhook Channel  | `kind=webhook`. User-owned. No account-level shared webhook yet.                                                  |
| Channel token    | `beep_ct_`. CLI copy says “CLI token”.                                                                            |
| Account          | Tenant. Team does **not** aggregate members’ Channels. A Beep targets only the current user’s Channels.           |
| Device           | **Banned.** Not a table, not a kind, not UI copy.                                                                 |

**Beep CLI** vs **CLI Channel** are different words. Do not rename the binary.

Desktop (planned `kind=desktop`) ≠ `beep up`.

---

## Grill Q&A trace (what the user actually chose)

- Q1 Channel = ? → **A destination record.** One `channels` row per endpoint; `kind` is how it delivers.
- Q2 Device = ? → **A informal, not a table.** Device = “an app on hardware” class of Channel. Email / webhook are Channels, not Devices.
- Q3 `kind=device` means ? → user said **“C 去掉，直接改名字成 CLI”**: drop it, rename straight to `cli` (not `device_cli`). So `kind=cli`.
- Q4 Settings/docs name → **A CLI Channel.** API `kind=cli`; UI “CLI” / “Add CLI”.
- Q5 Email/Web Push placement → **A now, same table.** One mailbox = one `email` row; one browser subscription = one `web_push` row. (User asked why B was recommended first: B = rename words now, migrate tables later; user rejected the split and chose full A.)
- Q6 Future iOS/Android/Desktop → **A parallel kinds** (`ios` / `android` / `desktop`), TERMS only, not enum. Desktop ≠ CLI.
- Q7 Pull API/token → **A cli.** `GET /cli/inbox`; copy “CLI token”; keep prefix `beep_ct_`.
- Q8 `cli` vs `beep` binary collision → **A accept.** Beep CLI = binary; CLI Channel = kind; `beep up` pulls CLI Channels.
- Q9 Account email as row → **A auto-create.** One default `kind=email` row per user. Old “tick email” = that row’s ID.
- Q10 Team Beep targets whose Channels → user typed **“A” then “B” on separate lines; recorded as B but NEVER explicitly confirmed. MUST confirm with user before implementing team semantics.** Recorded: own Channels only; Team is a tenant, not a fan-out.
- Q11 “Device” in product copy → **A banned.** Browser = Web Push Channel.

---

## Already done (do not redo)

- Mermaid quotes in `docs/architecture/channel.md` (GitHub parse fix), committed `c07f82a` on `feature/channel`.
- TERMS table rewritten (16-col aligned) + `channel.md` (intro, diagram, decisions, data model, §4 title, §5 targeting, §6 roadmap) + `web-push.md` (§7 line, responsibilities, settings/status/next/later lines).
- `channel.md` / `web-push.md` / `TERMS.md` contain zero “device” after the rewrite (verified by grep).

Code is **still** `kind=device` and `/device/inbox`.

---

## Docs leftovers (fix while here)

- [`docs/architecture/beeper.md`](docs/architecture/beeper.md):33 — still says “N channel types on that Beep”. Rewrite as “N Channels on that Beep (same BeepRun)”.
- [`PLAN.md`](PLAN.md) — still full old terms. Fix lines 3 (device notifications), 12 + 34 + 35 (`/device/inbox`, `/device/ack`), 22 (`kind (device, …)`), 27 (device queue), 42 (device notifications).
- [`reviewed.md`](reviewed.md) — **historical PR #47 review artifact, leave it.** Do not rewrite history.
- [`docs/COMPETITORS.md`](docs/COMPETITORS.md) — “Device push” refers to Apple/Google market features, not our domain term. Leave it.

---

## Implementation (code)

Follow [`AGENTS.md`](AGENTS.md): no new gems; API JSON via jbuilder; do not edit `core/db/queue_schema.rb` / `cable_schema.rb` / `cache_schema.rb`; use `bin/rails db:prepare` not `db:migrate` that dumps empty secondary DBs.

### A. Rename device → cli (exact anchors)

Model / scopes:

- [`core/app/models/channel.rb`](core/app/models/channel.rb):4 — `KINDS = %w[ device … ]` → `cli`.
- `channel.rb`:10 — enum default `"device"` → `"cli"`.
- `channel.rb`:21 — scope `active_devices` → `active_cli` (check callers first).
- [`core/app/models/channel_delivery.rb`](core/app/models/channel_delivery.rb):13 — scope `due_for_device` → `due_for_cli`.
- Migration: `UPDATE channels SET kind = 'cli' WHERE kind = 'device'` (+ rollback).

Controllers / routes:

- [`core/config/routes.rb`](core/config/routes.rb):78 — `namespace :device` → `namespace :cli`.
- Move `core/app/controllers/api/v1/device/` → `api/v1/cli/` (`base_controller.rb`, `inboxes_controller.rb`, `deliveries_controller.rb`); `Api::V1::Device::` → `Api::V1::Cli::`.
- [`core/app/controllers/api/v1/device/base_controller.rb`](core/app/controllers/api/v1/device/base_controller.rb): `@current_device` → `@current_channel`; `authenticate_device!` → `authenticate_cli_channel!`; `extract_device_token` → `extract_cli_token`; header `X-Device-Token` → `X-CLI-Token` (Bearer stays); messages “Missing/Invalid device token” → “CLI token”; `Channel.active.device` → `Channel.active.cli`.
- `inboxes_controller.rb`:3 — `due_for_device` → `due_for_cli`.
- `deliveries_controller.rb`:3 — `@current_device` rename follows base.

Fan-out (`beep_run.rb`):

- [`core/app/models/beep_run.rb`](core/app/models/beep_run.rb):57 — `"device"` / `"device:"` prefix → `"cli"` / `"cli:"`.
- `beep_run.rb`:119 `deliver_device` → `deliver_cli`; :127 `device_payload` → `cli_payload`; :128 `kind: :device` → `kind: :cli`; :129 `"device:"` → `"cli:"`; result key `"device"` (:120, :135, :138) → `"cli"`.
- Keep `deliver_to_channel` name (already generic).

Validation:

- [`core/app/models/beep.rb`](core/app/models/beep.rb):235 — `User::NOTIFICATION_CHANNELS` + `"device:"` prefix → `"cli:"` prefix.
- [`core/app/models/user.rb`](core/app/models/user.rb):2 — `NOTIFICATION_CHANNELS = %w[ email web_push device ]` → `cli`. (`DEFAULT_NOTIFICATION_CHANNELS` stays `%w[ email ]` until step B remaps it.)

CLI:

- [`apps/cli/internal/client/client.go`](apps/cli/internal/client/client.go):308,335 — `/api/v1/device/inbox`, `/api/v1/device/deliveries/:id/ack` → `/cli/…`; any `X-Device-Token` header → `X-CLI-Token`.
- [`apps/cli/cmd/channel.go`](apps/cli/cmd/channel.go):160 — default `--kind "device"` → `"cli"`; help text `(device, webhook, email)` → `(cli, webhook, email)`.
- Grep `apps/cli` for remaining “device” in `beep up` help/output copy.

Web:

- [`apps/web/src/components/settings/channel-management-settings.tsx`](apps/web/src/components/settings/channel-management-settings.tsx):57 `kind: "device"` → `"cli"`; :84 “Device Channels” → “CLI Channels”; :114 “Device Channel Created” → “CLI Channel Created”; :86-89 description “desktop machines and CLI daemons” → CLI wording; token label → “CLI token”.
- [`apps/web/src/lib/auth/slugs.ts`](apps/web/src/lib/auth/slugs.ts):27 reserved slug `"device"` — only change if reserved for this feature; do not break unrelated slug bans.

Copy sweep (rename all):

- “Missing device token” / “Invalid device token” (base_controller).
- “This device is no longer subscribed” ([`core/app/controllers/api/v1/push_subscriptions_controller.rb`](core/app/controllers/api/v1/push_subscriptions_controller.rb):59) → “This browser is no longer subscribed” / Web Push Channel wording.

### B. Beep targets Channel IDs; email / web_push are rows

Today `notification_channels` is a JSON list of **types** (`email`, `web_push`, `device`, `device:laptop`) on User / Beep / Beeper.

Target model:

- Beep (and Beeper defaults) store **Channel IDs**.
- `BeepRun#deliver_for` fans out to those rows only (current user’s Channels). No expand-by-kind.
- Deliver by `channel.kind`: `email` → mailer, `web_push` → Web Push, `cli` → inbox queue, `webhook` → HTTP.

Email:

- On user create (and a one-shot backfill), insert default `channels` row `kind=email`, name = email address.
- Unsubscribe-from-email should disable/remove that Email Channel (today: [`core/app/controllers/email_channel_unsubscribes_controller.rb`](core/app/controllers/email_channel_unsubscribes_controller.rb):13 strips `"email"` from `users.notification_channels`).

Web Push:

- One browser subscription = one `kind=web_push` Channel row.
- Today: `push_subscriptions` table + `PushSubscriptionsController`. Fold into `channels` (config/endpoint on the row, or migrate columns). Delete or stop using `push_subscriptions` once equivalent.
- Settings “device list” → Web Push Channel list.
- 410 / expired subscription deletes that Channel row.

`users.notification_channels` / `beeps.notification_channels` / `beepers.notification_channels`: replace with Channel ID lists (or a join table). Preserve behavior:

- `beep.rb`:189-199 defaults from `Current.user` / owner; [`core/app/models/beeper.rb`](core/app/models/beeper.rb):201,210,234-236 copies owner defaults on notify — must not silently target other team members’ Channels.
- New Beep default = current user’s Channels (or previous default set mapped to IDs).

Read current fan-out in [`core/app/models/beep_run.rb`](core/app/models/beep_run.rb) (`deliver_for`, `cli:` prefix after step A, `result["cli"]`).

### C. Tests to update (non-exhaustive)

- `core/test/models/channel_test.rb`
- `core/test/models/channel_delivery_test.rb`
- `core/test/models/beep_run_deliver_test.rb` (`kind: :device`, `notification_channels: %w[device]`, `device:laptop`, `result.dig("device", …)`)
- `core/test/controllers/api/v1/channels_controller_test.rb`
- `core/test/controllers/api/v1/device_controller_test.rb` → rename file to cli + paths
- `core/test/controllers/email_channel_unsubscribes_controller_test.rb`
- `core/test/models/user_test.rb` (`DEFAULT_NOTIFICATION_CHANNELS`, allowed list)
- Beep / Beeper / settings controller tests that assert `%w[email web_push]` / `%w[device]`

### D. Out of scope

- Adding `ios` / `android` / `desktop` to the enum.
- Account-level shared webhooks.
- Team fan-out to members’ Channels (**pending Q10 confirmation — see above**).
- Renaming `beep_ct_` prefix.
- Committing unless the user asks.

---

## Suggested order

1. Confirm Q10 with the user (A-then-B ambiguity) before locking team semantics.
2. Rename `device` → `cli` (model, migration, routes, controllers, CLI client, web copy, tests). Keep current `notification_channels` type-list working with `cli` instead of `device` for a green suite.
3. Fix docs leftovers (`beeper.md:33`, `PLAN.md`).
4. Auto-create Email Channel rows; map `"email"` targeting to that row.
5. Move `push_subscriptions` onto `kind=web_push` Channel rows.
6. Change Beep/Beeper/User targeting from kind strings to Channel IDs; restrict to current user.
7. Grep for `device` / `Device Channel` / `/device/` / `X-Device-Token` / `@current_device` and clear leftovers (except `reviewed.md`, COMPETITORS market sense, unrelated words).
8. Run core tests + CLI build + web typecheck/lint.

---

## Conversation recap (why)

User: current work should be called more like “device CLI”; are web push and email also devices or channels? Future iOS / Android / Desktop. `/grill-me` to unify terms.

11 rounds, one question at a time, each with a recommendation. User picked A everywhere except Q3 (“C 去掉，直接改名字成 CLI”) and Q10 (“A” then “B”, unconfirmed). Mid-grill the user asked why B was recommended for Q5 vs choosing A directly — answered: A = migrate tables now, B = words now/tables later; recommending B was scope control, not a better model; user chose full A.

Then: implement that in code. This file is the spec for that work.
