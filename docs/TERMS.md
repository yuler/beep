# Terms

Domain vocabulary for beep scheduling, delivery, tenancy, and Beepers.

| Term             | Meaning                                                                                                                                              |
| ---------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| Account          | Tenant (personal or team). All data belongs to an account. Team does not aggregate members' Channels.                                                |
| Beep             | A notification. `once` or `recurring`. Never produces a signal. Targets Channel IDs, not kinds.                                                      |
| Run              | One delivery attempt of a Beep (`BeepRun`).                                                                                                          |
| Channel          | One user-owned delivery destination (`channels` row). Adapter `kind`: `cli`, `email`, `web_push` (planned: `webhook`, `ios`, `android`, `desktop`).  |
| CLI Channel      | Channel with `kind=cli`. Pulled by `beep up` via `GET /cli/inbox`.                                                                                   |
| Beep CLI         | The `beep` binary (`apps/cli`). More than inbox pull (jobs / Beepers).                                                                               |
| Email Channel    | Channel with `kind=email`. Each user gets one default row for the account email.                                                                     |
| Web Push Channel | Channel with `kind=web_push`. One browser subscription is one row.                                                                                   |
| Webhook Channel  | Planned, not implemented yet; creation is rejected.                                                                                                  |
| Channel token    | Auth secret on a Channel (`beep_ct_`). CLI copy says CLI token.                                                                                      |
| ChannelDelivery  | One BeepRun sent to one Channel.                                                                                                                     |
| Poller           | Two jobs: Beep poller claims due notification Beeps; Beeper poller claims due Beepers.                                                               |
| Deliver          | Sends a BeepRun through each targeted Channel.                                                                                                       |
| Beeper App       | Catalog definition: manifest + receiver implementation. Official (seeded) or account custom (`BeeperApp`).                                           |
| Beeper           | Account-owned running instance: config, cron, alert state, default channels (`Beeper`).                                                              |
| Beeper Run       | One execution of a Beeper. Produces a **Signal** (`ok` / `alerting` / `error`). Does not deliver.                                                    |
| Signal           | Beeper-internal logic outcome for one Beeper Run: what the pager “heard.” Replaces Check / Checker meaning.                                          |
| Alert state      | Whether a Beeper is `ok` or `alerting`. Lives on the Beeper.                                                                                         |
| Threshold        | Consecutive non-`ok` Beeper Runs required before the first notification Beep.                                                                        |

## Trigger semantics

- `once`: detail-page **Send now** is a real delivery that consumes the schedule — `run_at` updates to send time and after success the Beep is `completed` with `next_run_at = nil`. If `run_at` is still in the future the UI confirms this before sending.
- `recurring`: detail-page **Trigger run** is an extra run — the schedule captured before firing is restored on completion (unless already due, in which case it recalculates to avoid duplicate firing), so the next slot is preserved.
- Channel settings **Test** (`deliver_test!`) only verifies delivery and rendering: no `beep_run`, no status or schedule change.
