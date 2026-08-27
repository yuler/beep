# Terms

Domain vocabulary for beep scheduling, delivery, tenancy, and Beepers.

| Term           | Meaning                                                                                          |
| -------------- | ------------------------------------------------------------------------------------------------ |
| Account        | Tenant (personal or team). All data belongs to an account.                                       |
| Beep           | A notification. `once` or `recurring`. Never runs a checker.                                     |
| Run            | One delivery attempt of a Beep (`BeepRun`).                                                      |
| Channel        | Delivery type for a Beep: `email`, `web_push`. Destination is the account owner in this period.  |
| Poller         | Two jobs: Beep poller claims due notification Beeps; Beeper poller claims due Installs.          |
| Deliver        | Sends a BeepRun through each Channel listed on the Beep.                                         |
| Beeper         | Catalog definition: manifest + implementation. Official (seeded) in this period.                 |
| Install        | Account-owned running Beeper: config, cron, alert state, default channels.                       |
| Beeper Run     | One execution of an Install. Produces `ok` / `alerting` / `error`. Does not deliver.             |
| Alert state    | Whether an Install is `ok` or `alerting`. Lives on the Install.                                  |
| Threshold      | Consecutive non-`ok` Beeper Runs required before the first notification Beep.                    |
