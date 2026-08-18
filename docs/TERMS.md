# Terms

Domain vocabulary for beep scheduling, delivery, and tenancy.

| Term    | Meaning                                                                              |
| ------- | ------------------------------------------------------------------------------------ |
| Account | Tenant (personal or team). All data belongs to an account.                           |
| Beep    | User-created reminder. Either one-shot (`once`) or `recurring`.                      |
| Run     | One actual fire of a Beep.                                                           |
| Channel | How a Run is delivered (email, web push, …).                                         |
| Poller  | Claims due Beeps and starts Runs.                                                    |
| Deliver | Sends a Run through each Channel.                                                    |
