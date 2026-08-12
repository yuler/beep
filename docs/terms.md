# Terms

Domain vocabulary for beep scheduling, delivery, and tenancy.

| 概念                     | 英文 Term      | 模型 / 表              | 说明                                                                 |
| ------------------------ | -------------- | ---------------------- | -------------------------------------------------------------------- |
| 主实体（用户新建的对象） | Beep           | Beep / `beeps`         | 一次性或周期的提醒/通知配置                                          |
| 类型                     | kind           | 字段 `kind`            | `once` \| `recurring`                                                |
| 一次性                   | once           | `kind: "once"`         | 只触发一次，用 `run_at`                                              |
| 周期                     | recurring      | `kind: "recurring"`    | 按 cron 重复，用 `cron` + `timezone`                                 |
| 计划触发时间（一次）     | run_at         | 字段 `run_at`          | 仅 `once` 使用                                                       |
| 下次触发时间             | next_run_at    | 字段 `next_run_at`     | poller 扫描用；`once` 创建时等于 `run_at`                            |
| 上次触发时间             | last_run_at    | 字段 `last_run_at`     | 最近一次被领取/执行的时间                                            |
| 时区                     | timezone       | 字段 `timezone`        | 解释 cron / 展示用，默认 UTC                                         |
| 周期表达式               | cron           | 字段 `cron`            | 仅 `recurring`，标准 cron                                            |
| Beep 状态                | status (Beep)  | 字段 `status`          | `active` \| `paused` \| `completed` \| `cancelled`                   |
| 执行记录                 | Run            | BeepRun / `beep_runs`  | 每一次实际触发的记录                                                 |
| 计划执行点               | scheduled_for  | 字段 `scheduled_for`   | 幂等键之一（与 `beep_id` 唯一）                                      |
| Run 状态                 | status (Run)   | 字段 `status`          | `pending` \| `running` \| `succeeded` \| `failed` \| `skipped`     |
| 通知渠道                 | Channel        | Channel / `channels`   | Email、Web Push 等投递方式                                           |
| 渠道类型                 | channel type   | 字段 `type` 或等价     | `email` \| `web_push`（MVP）                                         |
| 投递结果摘要             | result         | `beep_runs.result`     | 各 channel 成功/失败信息                                             |
| 租户                     | Account        | 已有                   | 个人或团队；所有数据带 `account_id`                                  |
| 调度扫描                 | Poller         | `BeepPollerJob`        | Solid Queue 定时扫 `next_run_at <= now`                              |
| 投递任务                 | Deliver        | `DeliverBeepRunJob`    | 真正按 channel 发送                                                  |
