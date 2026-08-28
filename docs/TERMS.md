# Terms

Domain vocabulary for beep scheduling, delivery, tenancy, and Beepers.

| Term        | Meaning                                                                                                    |
| ----------- | ---------------------------------------------------------------------------------------------------------- |
| Account     | Tenant (personal or team). All data belongs to an account.                                                 |
| Beep        | A notification. `once` or `recurring`. Never produces a signal.                                            |
| Run         | One delivery attempt of a Beep (`BeepRun`).                                                                |
| Channel     | Delivery type for a Beep: `email`, `web_push`. Destination is the account owner in this period.            |
| Poller      | Two jobs: Beep poller claims due notification Beeps; Beeper poller claims due Beepers.                     |
| Deliver     | Sends a BeepRun through each Channel listed on the Beep.                                                   |
| Beeper App  | Catalog definition: manifest + receiver implementation. Official (seeded) or account custom (`BeeperApp`). |
| Beeper      | Account-owned running instance: config, cron, alert state, default channels (`Beeper`).                     |
| Beeper Run  | One execution of a Beeper. Produces a **Signal** (`ok` / `alerting` / `error`). Does not deliver.          |
| Signal      | Beeper-internal logic outcome for one Beeper Run: what the pager “heard.” Replaces Check / Checker meaning. |
| Alert state | Whether a Beeper is `ok` or `alerting`. Lives on the Beeper.                                               |
| Threshold   | Consecutive non-`ok` Beeper Runs required before the first notification Beep.                              |
