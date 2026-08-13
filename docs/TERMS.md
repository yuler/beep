# Terms

Domain vocabulary for beep scheduling, delivery, and tenancy.

| Concept                          | Term           | Model / table          | Notes                                                                |
| -------------------------------- | -------------- | ---------------------- | -------------------------------------------------------------------- |
| Primary entity (user-created)    | Beep           | Beep / `beeps`         | One-shot or recurring reminder/notification config                   |
| Content                          | message        | field `message`        | What the reminder says                                               |
| Type                             | kind           | field `kind`           | `once` \| `recurring`                                                |
| One-shot                         | once           | `kind: "once"`         | Fires once; uses `run_at`                                            |
| Recurring                        | recurring      | `kind: "recurring"`    | Repeats on a cron; uses `cron` + `timezone`                          |
| Planned fire time (once)         | run_at         | field `run_at`         | Only for `once`                                                      |
| Next fire time                   | next_run_at    | field `next_run_at`    | Used by the poller; for `once`, equals `run_at` at create            |
| Last fire time                   | last_run_at    | field `last_run_at`    | When it was last claimed/executed                                    |
| Timezone                         | timezone       | field `timezone`       | Interprets cron / display; default UTC                               |
| Recurrence expression            | cron           | field `cron`           | Only for `recurring`; standard cron                                  |
| Beep status                      | status (Beep)  | field `status`         | `active` \| `paused` \| `completed` \| `cancelled`                   |
| Execution record                 | Run            | BeepRun / `beep_runs`  | One record per actual fire                                           |
| Planned execution point          | scheduled_for  | field `scheduled_for`  | Part of the idempotency key (unique with `beep_id`)                  |
| Run status                       | status (Run)   | field `status`         | `pending` \| `running` \| `succeeded` \| `failed` \| `skipped`       |
| Notification channel             | Channel        | Channel / `channels`   | Delivery methods such as Email, Web Push                             |
| Channel type                     | channel type   | field `type` or equiv. | `email` \| `web_push` (MVP)                                          |
| Delivery result summary          | result         | `beep_runs.result`     | Per-channel success/failure info                                     |
| Tenant                           | Account        | existing               | Personal or team; all data carries `account_id`                      |
| Schedule scan                    | Poller         | `BeepPollerJob`        | Solid Queue job scanning `next_run_at <= now`                        |
| Delivery job                     | Deliver        | `DeliverBeepRunJob`    | Actually sends via each channel                                      |
