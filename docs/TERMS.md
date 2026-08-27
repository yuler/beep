# Terms

Domain vocabulary for beep scheduling, delivery, and tenancy.

| Term        | Meaning                                                                              |
| ----------- | ------------------------------------------------------------------------------------ |
| Account     | Tenant (personal or team). All data belongs to an account.                           |
| Beep        | User-created reminder. Either one-shot (`once`) or `recurring`.                      |
| Run         | One actual fire of a Beep.                                                           |
| Channel     | How a Run is delivered (email, web push, …).                                         |
| Poller      | Claims due Beeps and starts Runs.                                                    |
| Deliver     | Sends a Run through each Channel.                                                    |
| Plugin      | A check definition: manifest + implementation. Official (seeded) or custom.          |
| Check       | One execution of a Plugin against user config. Produces `ok` / `alerting` / `error`. |
| Alert state | Whether a plugin Beep is currently `ok` or `alerting`. Lives on the Beep.            |
| Threshold   | Consecutive non-`ok` Checks required before the first notification.                  |
