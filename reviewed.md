# PR #47 Review — Notification Channel System

- PR:     https://github.com/yuler/beep/pull/47
- Title:  ✨ [core,cli,web] Implement Notification Channel system and device actions
- Branch: `feature/channel` → `main` (5 commits, 34 files, +1586 / -4)
- Date:   2026-09-10

## Commits

| Commit  | Scope | Subject                                                     |
|---------|-------|-------------------------------------------------------------|
| 71c195c | core  | Add Channel domain models and delivery routing              |
| cf47ebd | core  | Add Channel management and Device inbox/ack APIs            |
| 5828f1b | cli   | Add channel commands, daemon pull loop, local hook dispatch |
| 8edb482 | web   | Add Notification Channel settings and device management     |
| 59b5f1e | web   | Fix TypeScript and linting errors in channel settings       |

## CI Status (latest run 09:54 UTC)

| Workflow  | Result  |
|-----------|---------|
| Core CI   | SUCCESS |
| CLI CI    | SUCCESS |
| Web CI    | SUCCESS |
| PR Autofix| SUCCESS |

Earlier run (09:50 UTC) failed Lint / Typecheck / Autofix on
`apps/web/src/components/settings/channel-management-settings.tsx`
(unused `setKind`, `useExhaustiveDependencies`, `CopyableCode` props,
import order, formatting). Fixed by 59b5f1e; all checks green since.

## What was reviewed

- Core models: `core/app/models/channel.rb`, `channel_delivery.rb`, `beep_run.rb`,
  `user.rb`, `account.rb`, `beep.rb`, migration `20260910180000`, `schema.rb`.
- Core API: `channels_controller.rb`, `device/{base,inboxes,deliveries}_controller.rb`,
  4 jbuilder views, `config/routes.rb`.
- Core tests: `channel_test.rb`, `channel_delivery_test.rb`,
  `beep_run_deliver_test.rb`, `channels_controller_test.rb`, `device_controller_test.rb`.
- CLI: `cmd/channel.go`, `internal/client/client.go` (channel + device inbox/ack),
  `internal/config/config.go`, `internal/daemon/daemon.go`, `internal/exec/hook.go` + test.
- Web: `channel-management-settings.tsx`, `lib/api/channels.ts`, settings route wiring.
- Docs: `docs/architecture/channel.md`, `PLAN.md`.

## Findings

### Strengths

- Clean separation: scheduling (Core) stays platform-agnostic; device delivery is a
  pull-only queue with 30m TTL; local execution is declarative via `.beep/hooks/on_beep`.
- Good test coverage for new behavior: token generation, account scoping,
  claim/expire transitions, fan-out to `device` and `device:<name>`, inbox claim + ack.
- Auth design is sound: user Bearer token for channel CRUD (account-scoped),
  `X-Device-Token` / Bearer for device endpoints, `touch_last_seen` on auth,
  masked token in list views, raw token only on create.
- CLI reuses existing `setHeaders` + timeout patterns; expired deliveries are
  dropped client-side and acked as failed; hook runs with 60s timeout and process-group setup.

### Issues / risks (non-blocking, suggested follow-ups)

1. `InboxesController#show` claims deliveries with `each(&:claim!)` outside a
   transaction and without row locking — two concurrent `beep up` polls could
   double-deliver. Consider `SELECT ... FOR UPDATE SKIP LOCKED` or atomic claim.
2. `ChannelDelivery#claim!` uses non-bang `update` (returns false on failure, silently
   ignored by the controller loop); `fail!` ignores its `error_message` argument —
   error context is lost. Persist error into payload/result if needed.
3. Device auth has no per-account scoping: any valid device token can inbox/ack only
   its own channel (good), but there is no rate limiting or token rotation/revocation
   UI beyond delete.
4. `BeepRun#deliver_device` fans out synchronously inside `deliver_now`; a user with
   many device channels creates N rows in-request. Fine for now, watch for slow runs.
5. `channel_params` permits `:config` but the model has no `config` column and the
   jbuilders never render it — dead param; either add the column or drop it.
6. Web `handleDelete` uses blocking `confirm()`; no per-row pending state, and the
   created-token panel never dismisses after copy. Minor UX polish.
7. `expires_at` is computed from `scheduled_for` in `BeepRun#deliver_to_channel` but
   the model default is `30m from_now` — two TTL sources; keep the explicit one and
   document precedence.

## Verdict

Approve. Scope matches `docs/architecture/channel.md` + `PLAN.md`, tests cover the
new paths, and CI is green after the lint/type fix. The items above are follow-ups,
none block merge.
