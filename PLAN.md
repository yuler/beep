# Notification Channel System Implementation Plan

Roadmap for delivering the **Notification Channel System**, enabling user-scoped channels, device notifications, local hook triggers (`.beep/hooks/on_beep`), and the `get-off-work` automated clock-in / screen-blanking integration.

---

## Phases Overview

| Phase | Milestone | Scope | Deliverables |
| :--- | :--- | :--- | :--- |
| **Phase 1** | **Core Domain Models** | `core/app/models/channel*` | Models `Channel`, `ChannelDelivery`, associations with `User`, `Account`, and `BeepRun`. |
| **Phase 2** | **Core REST APIs** | `core/app/controllers/api/v1/` | Channel CRUD, device queue polling (`GET .../device/inbox`), delivery ACK (`POST .../device/ack`). |
| **Phase 3** | **CLI Daemon & Hook Dispatch** | `apps/cli/internal/` | `beep channel` command suite, polling loop in `beep up`, hook runner with TTL check. |
| **Phase 4** | **Web UI Integration** | `apps/web/src/routes/` | User settings channel management, Beep creation channel selector. |
| **Phase 5** | **E2E & Script Integration** | `~/.beep/jobs/` & `id5.cn` | Deploy cloud `get-off-work` job and configure local `on_beep` hook with `checkin-blank.sh`. |

---

## Phase 1: Core Domain Models & Delivery Router

### 1.1 Migrations & Models
- `Channel`: `account_id`, `user_id`, `kind` (`device`, `email`, `web_push`, `webhook`), `name`, `token`, `status` (`active`, `disabled`).
- `ChannelDelivery`: `beep_run_id`, `channel_id`, `status` (`pending`, `claimed`, `succeeded`, `failed`, `expired`), `payload`, `expires_at`.
- Associations on `User`, `Account`, and `BeepRun`.

### 1.2 Router & Delivery Logic
- Update `BeepRun#deliver_now` to fan-out to configured channels (email, web_push, device queue).

---

## Phase 2: Core REST APIs

- `GET/POST/DELETE /api/v1/:account/channels` — Channel management.
- `GET /api/v1/device/inbox` — Device polling for pending deliveries.
- `POST /api/v1/device/deliveries/:id/ack` — Execution ACK.

---

## Phase 3: CLI Daemon & Local Hook System (`apps/cli`)

- `beep channel` CLI commands (list, create, test).
- Pull loop in `beep up` for device notifications with TTL verification.
- Local execution of `$WORKSPACE/.beep/hooks/on_beep`.

---

## Phase 4: Web UI Integration (`apps/web`)

- Channel management under User Settings.
- Channel multi-selector in Beep creation form.

---

## Phase 5: End-to-End Verification & Script Integration

- Deploy cloud `get-off-work` cron job.
- Configure `.beep/hooks/on_beep` to trigger `checkin-blank.sh`.
- Test normal triggering, reboot resilience, and TTL expiration.
