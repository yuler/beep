# PR #47 Review — Notification Channel system (`feature/channel` → `main`)

Scope: `git diff main...feature/channel` (45 files, +1840/−144). Focus: bugs and security issues.
All findings below are verified against the code; items marked **[verified]** were reproduced by running code/tests.

## High severity

### 1. `User#destroy` raises FK violation when web_push channels have deliveries [verified]
- `core/app/models/user.rb:10-11` — `has_many :push_subscriptions, dependent: :delete_all` and
  `has_many :channels, dependent: :destroy` target the same `channels` table.
  `delete_all` runs first (association order) and deletes web_push rows **without**
  destroying their `channel_deliveries`, violating `channel_deliveries.channel_id` FK.
- Repro: user + web_push channel + delivery, then `user.destroy!` →
  `ActiveRecord::InvalidForeignKey: SQLite3::ConstraintException: FOREIGN KEY constraint failed`.
- Fix: drop `dependent: :delete_all` from `push_subscriptions` (keep the association read-only;
  `channels, dependent: :destroy` already covers those rows and cascades deliveries).

### 2. CLI inbox claim is racy; claimed-but-unacked deliveries are stuck forever
- `core/app/controllers/api/v1/cli/inboxes_controller.rb:3-4` — fetch-then-`claim!` in a loop,
  no transaction/row lock. Two `beep up` processes sharing one token (or overlapping polls) can
  fetch the same `pending` rows before either claims → duplicate hook execution.
- `ChannelDelivery.claim!` uses non-bang `update` and the inbox ignores its return value, so a
  delivery that expired between fetch and claim is still rendered to the client.
- Nothing ever reaps `claimed` rows: `stale_pending` scope (`channel_delivery.rb:12`) has zero
  callers, and there is no timeout reclaiming `claimed` → crash between claim and ack loses the
  notification permanently (client never sees it again, server never retries).
- Fix: atomic claim (`update_all(status: claimed) … WHERE status = pending` returning claimed ids),
  and either return `claimed` rows to the same poller on a grace window or expire/requeue stale
  `claimed` rows.

### 3. Invalid channel `kind` → unhandled 500 instead of 422 [verified]
- `core/app/controllers/api/v1/channels_controller.rb:10` + `core/app/models/channel.rb:17`.
  `enum :kind` raises `ArgumentError` on assignment of an unknown value (verified in console),
  so `POST /api/v1/:slug/channels` with `kind: "bogus"` raises before validation runs.
  The `validates :kind, inclusion:` check is dead code for param assignment.
- Fix: rescue `ArgumentError` in `create` → 422, or whitelist `kind` before assignment.

### 4. Lost endpoint uniqueness → duplicate web_push channels; `.sole` raises
- Old `push_subscriptions` had `UNIQUE(user_id, endpoint)` (`schema.rb:327`) and
  `rescue RecordNotUnique` retry. New `channels` table has no endpoint constraint, and
  `Channel.upsert_web_push_for!` (`channel.rb:34-60`) does a Ruby-side linear scan
  (`user.channels.where(kind: :web_push).find { … }`) with no race protection.
- Concurrent subscribes (double-click, two tabs) create duplicate rows; the updated test helper
  `push_subscriptions.for_endpoint(ep).sole` then raises on duplicates instead of self-healing.
- Fix: add a unique expression index on the endpoint (SQLite supports
  `CREATE UNIQUE INDEX … ON channels(user_id, json_extract(config, '$.endpoint'))`) and restore
  find-or-retry logic. At minimum handle `>1` row without raising.

### 5. Channels API is account-scoped, not user-scoped (cross-user read/delete)
- `channels_controller.rb:4-5,30-31` — `index` returns **all** `Current.account.channels`
  (names, masked tokens, `last_seen_at` of other users); `destroy` deletes any channel in the
  account. Severity depends on team accounts having multiple members, but `TERMS.md` explicitly
  states teams do not aggregate members' channels and beeps target only the user's channels.
- Fix: scope to `Current.user.channels` (or authorize membership/ownership explicitly).

## Medium severity

### 6. `beep up` auto-executes workspace-controlled hooks (trust boundary)
- `apps/cli/internal/exec/hook.go:16-31,33-69` — on every delivery the daemon executes
  `$WORKSPACE/.beep/hooks/on_beep` (plus three fallback paths) with no executable-bit check,
  no ownership/signature check, and no user confirmation. A cloned/malicious workspace gets
  arbitrary code execution via normal daemon operation. `CombinedOutput` also buffers unbounded
  hook output (OOM vector against a compromised server).
- Related liveness bug: `daemon.go:271-315` runs `pollCliInbox` **synchronously** inside
  `pollAndExecute` on every tick; each hook may run up to 60s, starving runner job polling.
- Fix: check executable bit + warn on first use / require opt-in per workspace; cap output buffer;
  dispatch hooks asynchronously or on a separate ticker from runner polling.

### 7. CLI token handling weaknesses
- `apps/cli/internal/client/client.go:30-39` — `CheckRedirect` strips `X-Runner-Token` on
  cross-host redirect but **not** `X-CLI-Token` (or the `Authorization: Bearer` used by
  `ListChannels`/`CreateChannel`): credential leak on redirect.
- `setHeaders` (`client.go:279-289`) attaches `X-CLI-Token` to **every** request, including runner
  ping/poll/task endpoints — unnecessary credential spread.
- `ListChannels`/`CreateChannel` interpolate `accountSlug` and `AckCliDelivery` interpolates
  `deliveryID` into URLs without escaping (`client.go:312,343,423`). Slugs are server-restricted,
  so impact is low, but use `url.PathEscape`.
- `core/app/controllers/api/v1/cli/base_controller.rb:37-41` — bearer regex `/^Bearer /` is
  case-sensitive (RFC 7235 scheme is case-insensitive); also token lookup is a plain string
  compare (timing) — low risk given token entropy, note only.

## Low severity / correctness nits

### 8. `fail!(error_message)` silently discards the message
- `channel_delivery.rb:30-32` — parameter is accepted and ignored; `channel_deliveries` has no
  error column. Client diagnostics (`hookErr.Error()`, `"expired"`) are lost server-side.
  Either persist it (e.g. in `payload`) or drop the parameter.

### 9. No delivery state-machine guards
- `deliveries_controller.rb:2-19` — ack allows `pending → succeeded/failed` (no claim required),
  double-ack, and `succeeded → failed` flips. Decide legal transitions and enforce in the model.

### 10. `touch_last_seen` writes on every poll
- `cli/base_controller.rb:29` → `channel.rb:62-64` — each inbox poll (default 3s per device) does a
  DB write bumping `updated_at`. Throttle (e.g. only if older than 1 minute) or use
  `update_column`.

### 11. Silent notification loss on unresolvable channel IDs
- `beep.rb:233-237` accepts any 10–40 char `[0-9a-zA-Z_-]` string; `beep_run.rb:63-65` then
  silently skips IDs matching no channel. A typo'd channel ID passes validation and the beep
  reports success with nothing delivered. Consider validating existence at beep-create/update time.

### 12. Default-channel fallback looks dead, contradicting ID-targeting direction
- `beep.rb:189-198` — `target_user.notification_channels.presence || …pluck(:id)`: users always
  have `DEFAULT_NOTIFICATION_CHANNELS = ["email"]` (`user.rb:62-66`), so the `pluck(:id)` branch
  is unreachable unless notification_channels is explicitly emptied. New beeps therefore default
  to legacy kind strings while `TERMS.md` says beeps target channel IDs. Clarify intended default.

### 13. Migrations
- `20260911110000 BackfillEmailAndWebPushChannels#down` deletes **all** `email`/`web_push`
  channels, including ones created after the migration. Scope the delete to backfilled rows or
  make it a no-op with a warning.
- `20260911100000 RenameDeviceChannelsToCli#down` converts all `cli` rows back to `device`,
  including rows that were originally `cli`. Lossy rollback.
- Backfill token generation has no collision retry (unlikely; migration would fail midway).

### 14. Web UI nits
- `apps/web/src/components/settings/channel-management-settings.tsx:733-757` — the raw channel
  token stays in component state/DOM indefinitely after creation; clear it on unmount or after
  copy. `onClick={() => handleDelete(ch.id)}` is fine (errors caught internally) but uses
  blocking `confirm()`.
- `apps/web/src/lib/api/channels.ts:846,857` — `accountSlug` interpolated without
  `encodeURIComponent`.

## What looks good (no action)
- `create.json.jbuilder` returns the raw token once; `index`/`show` expose only `masked_token`.
- CLI `DeliveriesController#ack` scopes via `@current_channel.deliveries.find` (no cross-channel ack).
- Web-push SSRF allowlist and HTTPS validation preserved in the move to `Channel`.
- `config.json` written `0600` (`apps/cli/internal/config/config.go:89`).
- New tests pass: 38 runs, 136 assertions, 0 failures (channel, delivery, beep_run, channels + cli
  controller tests).
