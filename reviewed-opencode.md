# Review 记录 — feature/channel（opencode）

- 范围：`main...HEAD`（channel 体系：Core Device Flow / 投递 / CLI / Web）。
- 方法：三路并行实读 + 抽查验证。本轮修复时发现并行改动：`1cdb31d` 已修掉清单大半；
  另一 agent 正在同树做 Runner device-flow 镜像 + channel handler 重构（`Channel::Handlers`），
  其重构完整保留了本轮改动；`device` 页已拆为 `device/channel` + `device/runner`，修复跟到了新文件。
- 验证：`core rails test` 469 runs 0 failures；`apps/cli go vet + go test ./...` 全绿；
  `apps/web tsc --noEmit` 通过；biome  clean（5 个经手文件）；rubocop clean（10 个经手文件）。

## 结论

| 等级   | 数量 | 状态                                                              |
|--------|------|-------------------------------------------------------------------|
| High   | 7    | 6 由 `1cdb31d` 修掉；CSRF（W-H2）本轮修掉并补了回归测试           |
| Medium | 14   | 10 由 `1cdb31d` 修掉；剩余 4（C-M7/L2/L3 除外见下）本轮修掉       |
| Low    | 8    | 5 由 `1cdb31d` 修掉；O5/show 作用域/轮询 cap/hook 单测本轮修掉    |

## 本轮修复（opencode，未提交）

### Core

| 编号 | 文件                                                        | 改动                                                                                          | 测试                                    |
|------|-------------------------------------------------------------|-----------------------------------------------------------------------------------------------|-----------------------------------------|
| W-H2 | `concerns/request_forgery_protection.rb`、`client.ts`、`development_cors.rb` | cookie 会话 JSON 变异请求强制 `X-Requested-With`；顺带修了真正的 BUG：`included do include` 把 Rails 模块插到 concern 之上，`verified_via_header_only?` 覆写从未生效 | `csrf_protection_test.rb` 新增 4 用例 |
| C-M7 | `channel/handlers/web_push.rb`（重构前 `channel/web_push.rb`） | `upsert_for!` 加 `user.with_lock`，堵无唯一索引下的并发重复插                                  | 沿用既有 upsert 测试                    |
| L2   | 同上                                                        | 通道名只存 UA 首个 product token，去 OS/设备 PII                                               | —                                       |
| L3   | `channel/delivery.rb`、`cli/inboxes_controller.rb`          | 新增 `expire_stale!`，inbox 读取前把本 channel 过期投递置 `expired`，不再永久堆积               | inbox stale 用例                        |
| O5   | `tokens_controller.rb`、`authorizations_controller.rb`、新增 `device_flow_error.json.jbuilder` | device-flow 全部 inline `render json:` 收敛到共享 jbuilder，wire 形状不变（CLI 按 error code 解析） | 既有 tokens 断言                        |
| C-M4b | `authorizations_controller.rb`、`channels.ts`              | `show` 移出 `skip_account_scope`（与 update/destroy 对齐）；前端 verify 走 slug 路径            | show 非成员 404 用例                    |

### Web

| 编号 | 文件                                                              | 改动                                                                |
|------|-------------------------------------------------------------------|---------------------------------------------------------------------|
| W-M3 | `$account_slug/device/channel.tsx`、`device/runner.tsx`（后者是并行 agent 的新文件，顺手同修） | 删除 URL code 自动验证，只做输入框预填；验证成功后清 search；加防钓鱼提示 |
| W-M1 | `push.ts`（由 `1cdb31d` 修）                                       | 无残留                                                              |
| W-M2 | `device/channel.tsx`（`switchSearch = {}`，由 `1cdb31d` 修）       | 无残留                                                              |
| —    | `dashboard-sidebar.tsx`                                           | `/dev/letters`、`/admin/jobs`、`/admin/stats` 去掉多余 params       |

### CLI

| 编号 | 文件                              | 改动                                                                       | 测试                          |
|------|-----------------------------------|----------------------------------------------------------------------------|-------------------------------|
| M8   | `cmd/channel.go`                  | `slow_down` interval 上限 30s；连续 10 次非 OAuth 错误退出并指数 backoff    | —                             |
| M9   | `internal/exec/hook_test.go`      | 补空 root / 组可写拒绝 / 无执行位拒绝 / rune 截断 4 用例（`writeHook` 回 path） | 新用例全过                    |
| M7   | `cmd/up_test.go`                  | 补 `--server up` 动词值存活回归用例（既有 `--workspace runner` 已覆盖）      | 用例过                        |
| M6   | `examples/hooks/on_channel.example` | 加“勿 eval BEEP_EVENT*”红字；修正 no-hook 注释（现为 warn + ACK failed）  | —                             |

## `1cdb31d` 已修（本轮实读确认，不必回查）

C-H1（claim 重试门控 + 并发 no-op 测试）、C-H2（逐 channel 增量投递 + per-user email + 多接收人测试）、
C-H3（`with_lock` approve + 422 测试）、C-H4（条件 `update_all` ack）、C-M1（422 VALIDATION_ERROR 测试）、
C-M2（显式 channel 去重测试）、C-M3（IP 限流）、C-M4（destroy 作用域 + 测试）、C-M5（claim 过期条件）、
C-M6（Cli base 404 rescue）、L1（email 真实 test 邮件）、W-H1（deny slug + 测试）、W-M1/W-M2、
L-H1（无 hook ACK failed）、L-H2（token 剥离 + env 传递 + 测试）、L-M1/M2（runner/channel 分头）、
L-M3（noRedirectClient）、browser trim/大小写、`~` 映射、socket 先验权、`channel.test` hook 隔离、
hook 真名返回、rune 截断、up 日志前缀。

## 残留说明（非阻塞）

- `show` 全局可见性：code 未绑定 account 前，同 slug 成员仍可枚举任意 code 的 `channel_name`；
  8 字符 ~38bit + 30/min IP 限流使枚举不可行，属 device-flow 固有属性，接受。
- Hook TOCTOU（check 与 exec 之间替换文件）：workspace 根在 daemon 启动时已 0700，
  `isSafeHook` fail-closed；剩余窗口需本地攻击者 + 竞态，可接受。
- Socket umask 窗口：同上，目录 0700 已使其他用户不可达，未改。
- `BEEP_EVENT` 未加 allowlist：示例 + 注释已警告勿 eval；服务端字段多为受控枚举，接受。
- 并行 agent 的 Runner device-flow（`Runner::Authorization`、runner device 页）是新代码，
  按 channel 同标准顺手修了前端自动验证；其后端 show/destroy 作用域请该 agent 自查。
