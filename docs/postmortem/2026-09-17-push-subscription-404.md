# Postmortem: 生产 `Push::Subscription` 404（2026-09-17）

| 项目       | 内容                                                                                        |
|------------|---------------------------------------------------------------------------------------------|
| 现象       | `DELETE /api/v1/:slug/push_subscriptions/:id` 与 `POST .../:id/test` 报 404                  |
| 报错原文   | `Couldn't find Push::Subscription with 'id'="2q1tne…" [WHERE "channels"."kind" = ? AND "channels"."user_id" = ?]` |
| 上报人     | qichenwx（生产站 beep.yuler.cc，个人账号 `qichenwx9620`）                                    |
| 根本原因   | 回填迁移用原生 SQL 写入了带横杠 UUID **TEXT** 主键，绕过了 app 的 base36/BLOB(16) UUID codec |
| 影响面     | 迁移建的所有 `channels` 行：全部 email channel + 回填的 web_push 行；相关 find-by-id 接口全挂 |
| 数据修复   | 迁移 `20260917140205_fix_channels_text_uuid_ids.rb`（随部署自动执行）                        |
| 状态       | 待合入部署后关闭                                                                            |

## 1. 时间线

| 时间                | 事件                                                                              |
|---------------------|-----------------------------------------------------------------------------------|
| 2026-08-17          | #4 上线 web push，独立 `push_subscriptions` 表（UUID 主键，Rails 生成）             |
| 2026-08-20          | 出问题的订阅行在老表里创建（endpoint 为 FCM，Chrome/Linux）                        |
| 2026-09-11          | 迁移 `20260911110000_backfill_email_and_web_push_channels.rb` 随 #47 合入          |
| 2026-09-15          | #47 上线：web push 并入 `channels` 表，老行被 backfill（换新 id，保留 created_at） |
| 2026-09-17 上午     | 用户在设置页点删除/测试，稳定复现 404，上报                                        |
| 2026-09-17 下午     | 定位根因；本地库手工验证修复 SQL；改由正式 migration 修复                         |

## 2. 根因

App 的 UUID 策略（抄 basecamp/fizzy，两处配合）：

- `core/lib/rails_ext/active_record_type_uuid.rb` —— Ruby 侧 id 是 **base36、25 位、时序**串
  （`Type::Uuid.generate` 用 `SecureRandom.uuid_v7` 转码，如 `03gvtja60ojxfq685e7rk3xug`），
  SQLite 里存 **16 字节 BLOB**，`serialize`/`deserialize` 负责互转。
- `core/config/initializers/uuid_primary_keys.rb` —— 建表时 uuid 列映射为 `blob(16)`。

回填迁移用原生 SQL `lower(hex(randomblob(…)))` 直接写了**带横杠 UUID TEXT** 做 `channels.id`，
绕过了 codec。后果（SQLite 动态类型，`BLOB ≠ TEXT`）：

| 操作                  | 结果                                                  |
|-----------------------|-------------------------------------------------------|
| `index`（无 id 条件） | 正常（这就是列表页一直能看到这行的原因）              |
| `user_id` scope       | 正常（`account_id`/`user_id` 是从 users 表原样拷的 BLOB） |
| 读 `id`               | `deserialize` 把 36 个 ASCII 当二进制 hex 化 → 稳定的 **56 位串**（API/报错里看到的就是它，库里实际存的是标准 UUID） |
| `find(id)`            | 绑定 blob 永远 ≠ 存着的 TEXT → `RecordNotFound`，404   |

一句话：**行在库里，但 Rails 永远 find 不到它**。跟"web push 放到 account/user 下"的 scope
设计无关——`kind + user_id` 两个条件都是对的。

## 3. 排除过的假设（避免后人重走）

| 怀疑方向                              | 结论                                        |
|---------------------------------------|---------------------------------------------|
| 服务端吊销/过期 token                 | 无。PAT 与 channel token 均无过期机制       |
| `channel disconnect`/`logout`/`unset` | 代码审计：所有 `SaveFile` 调用都保留其他字段 |
| 多账号切错 / 身份错位                | `/me` 只有一个个人账号；同 cookie 下 PAT 与 session 看到同一行 |
| 迁移丢数据                            | 否。出事的 id 在 `channels` 里存在，`user_id` 也保留了 |
| 生产存的 id 不是 UUID（排查中途误判） | **已纠正**：存储层是标准 36 位 UUID TEXT；56 位只是 codec 误解码的读假象 |

## 4. 影响面

- 所有迁移建的 `channels` 行：每个用户的 email channel + 从老表 backfill 的 web_push 行。
- 命中的接口：`push_subscriptions` 的 destroy/test、`channels` 的 show/destroy/test。
- 不受影响：Rails 建的行（cli 等，BLOB id）；列表/轮询/投递（不按 id 查）。
- 前端 `disableWebPush` 已吞掉 destroy 的 404，用户侧感知有限。

## 5. 修复

数据修复收敛为正式 migration（随部署自动跑，无需手跑 SQL）：

`core/db/migrate/20260917140205_fix_channels_text_uuid_ids.rb`

- `UPDATE channels SET id = x'…'`：把 TEXT 还原为 codec 认的 16 字节
  （`36^25 > 2^128`，base36↔int↔hex 双射，无损）。
- 顺带把 `channel_deliveries` / `channel_authorizations` 里存成
  `serialize(misdecoded_id)` 垃圾 blob 的外键重指回新 id（按 codec 逆运算确定性重算；
  两种 blob 长度都匹配，防 Binary 截断；对不上的一律不动并告警）。
- 已是 BLOB(16) 的行是 no-op；`down` 为空（同 20260911110000 的不可逆先例）。
- 本地已用生产导入数据手工验证过等价 SQL：`find`、scoped find、delivery 关联全恢复。

## 6. 后续 Action Items

| # | 事项                                                                   | 状态    |
|---|------------------------------------------------------------------------|---------|
| 1 | 合入本 PR 并部署，线上点 test 按钮闭环                                  | 待执行  |
| 2 | 约束：原生 SQL 迁移凡涉及 uuid 主键必须写 `BLOB(16)`（`unhex` 或 Ruby 生成） | 待立规  |
| 3 | 回归测试：跑完迁移 SQL 后用 model `find` 一遍回填行                     | 待补    |
| 4 | 老 `push_subscriptions` 表已无人读，另起迁移删除                        | 待排期  |

## 附录：关键证据

- 线上同 cookie 复现：`GET index` 返回 1 行，`POST .../test` 同 id 报 404。
- 本地复现：`Channel.find` 对迁移行 404（修前）/ OK（修后）。
- 本地存储实测：`channels.id` 迁移行 `text/36`，Rails 建行 `blob/16`；`users.id` 全 `blob`。
- 关键文件：`core/lib/rails_ext/active_record_type_uuid.rb`、
  `core/config/initializers/uuid_primary_keys.rb`、
  `core/db/migrate/20260911110000_backfill_email_and_web_push_channels.rb`、
  `core/app/controllers/api/v1/push_subscriptions_controller.rb`、
  `core/app/controllers/api/v1/push_subscriptions/tests_controller.rb`。
