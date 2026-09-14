# Review 记录 — feature/channel（muse）

范围：`main...HEAD`（`feature/channel`），117 文件，约 +5103 / -711。
重点：`core/` Channel / Beep / Device Flow、`apps/cli/` channel / runner / daemon / exec、`apps/web/` channels / push / device。
方法：三路并行读码 + 抽查验证（已确认下述 High 项真实存在）。未跑全量测试，仅 `tsc` 概念验证与读码。

## 修复状态（2026-09-14）

已按“待修建议顺序”全部修复并验证：

| 验证                 | 结果                                   |
|----------------------|----------------------------------------|
| `apps/web tsc`       | 通过                                   |
| `apps/cli go vet/test` | 通过（`./...` 全绿）                 |
| `core rails test`    | 通过（84 runs，0 failures，含更新后的 delivery 状态机测试） |

| 编号 | 修复内容                                                                                          |
|------|---------------------------------------------------------------------------------------------------|
| H1   | `channels.ts:61`、`push.ts:49` `{)` → `{ method: "POST" }`                                        |
| H2   | `beeps_controller` 去掉 `:source_type/:source_id/:intent`；`Beep#validate_source` 加同 account 校验；title 禁 CRLF |
| H3   | `ChannelsController#set_channel`、`TestsController` 改 `where(user: Current.user)`；`channel_params` 去掉 `:config` |
| H4   | `Authorization` 加 `consumed` 态、一次性 `consume_token!`、`poll_interval_exceeded?`（slow_down）、`approve!` 事务、`deny!` 守卫、`create_request!` 重试；两 controller 加 `rate_limit` |
| H5   | `notify.go` AppleScript 转义 + 限长 500                                                            |
| H6   | `CheckRedirect` 跨 host 同删三 token                                                               |
| H7   | `DispatchOnBeepHook` 拒空 root + 属主/权限/可执行位校验；hook 输出截断 2048                        |
| M16  | `/device` 去自动验证、URL 变化同步、验证后清 search、防钓鱼提示、展示 `expires_in`、成功页用服务端名 |
| M17  | 切 slug 清 `createdChannel` + Dismiss、token 提示改 `&lt;token&gt;` 占位                            |
| M18  | banner 只跳转不自动 `enable()`                                                                     |
| M3   | allowlist 加点边界                                                                                 |
| M5   | `Delivery#claim!` 原子、`succeed!/fail!` 限 pending/claimed、error 进 payload 截断；inbox 原子 claim；ack 非法转移 422 |
| M6   | `Beep::Run#claim_delivery?` 只认原子 claim                                                         |
| M4/M7/L | `upsert_for!` 走 `for_endpoint`；`approve!` 事务；`masked_token` 只露前缀；backfill `down` 改 no-op |
| CLI  | token 文件/目录 0600/0700 回收、`browser.Open` 限 https（localhost 例外 http）、`DisconnectChannel` 显式双头、`DeleteJob`/ack `PathEscape`、`PollDeviceToken` 信号/默认/`errors.As`、runner 单 tick 上限 + nil workspace 守卫、ack error 截断 |
| Web  | 路径段 `encodeURIComponent`、`copy-button` 按成功态、`load` 竞态守卫、timer ref 清理、按钮忙时禁用     |

## 结论

| 等级   | 数量 | 说明                                   |
|--------|------|----------------------------------------|
| High   | 7    | 构建失败 1，其余为越权 / 伪造 / 注入   |
| Medium | 13   | 限流缺失、竞态、Token 泄露、UX 安全    |
| Low    | 9    | 强参数、日志、i18n、竞态边角           |

最先修：Web 两个语法错误（构建已断）、Beep source 伪造、Channel 越权删除 / 测试、CLI `osascript` 注入、Device token 可重复领取。

---

## High

### H1 [Web] 构建失败：`testChannel` / `testPushSubscription` 语法错误

- 位置：`apps/web/src/lib/api/channels.ts:61`、`apps/web/src/lib/api/push.ts:49`
- 现象：
  ```ts
  await apiFetch(`.../test`, {)
      method: "POST",
  });
  ```
  `{)` 应为 `{ method: "POST" }`。已读码确认两处完全一致的笔误。
- 影响：`tsc --noEmit` 失败（TS1136/TS1109/TS1128），整个 `apps/web` 无法构建。
- 修复：
  ```ts
  await apiFetch(`.../test`, { method: "POST" });
  ```

### H2 [Core] Beep source 可伪造（mass assignment）

- 位置：`core/app/controllers/api/v1/beeps_controller.rb:48`、`core/app/models/beep.rb:189-193`
- 现象：`beep_params` 放行 `:source_type, :source_id, :intent` + `metadata: {}`，而 `validate_source` 只校验 `source_type.in?(SOURCE_TYPES)`，不校验 `source_id` 是否属于 `Current.account`。
- 影响：同 account 成员可伪造他人 `Beeper` / `Runner::Job` 为来源，污染 `_beep.json.jbuilder` 展示的 provenance；`intent` / `metadata` 任意写。
- 修复：`beep_params` 去掉 `:source_type, :source_id, :intent`，服务端设置 source；`intent` 用 allowlist，`source_id` 必须 scope 到 `Current.account`。

### H3 [Core] 任意成员可删除 / 测试他人 Channel

- 位置：`core/app/controllers/api/v1/channels_controller.rb:32`、`core/app/controllers/api/v1/channels/tests_controller.rb:11`
- 现象：都用 `Current.account.channels.find(...)`，仅 account 级隔离。对比 `PushSubscriptionsController:3,20` 与 `PushSubscriptions::TestsController:18` 正确 scope 到 `Current.user`。而 `Channel belongs_to :user`、`User has_many :channels` 表明支持按 user 归属。
- 影响：成员 A 可 `DELETE` 成员 B 的 CLI channel（对方 daemon 掉线，DoS），可 `POST .../test` 骚扰 B 的端点。
- 修复：scope 到 `Current.user.channels`（或跨用户操作要求 admin），与 push 订阅保持一致。

### H4 [Core] Device token 可无限重复领取，无 interval 限流

- 位置：`core/app/controllers/api/v1/channels/cli/authorizations/tokens_controller.rb:5-29`、`core/app/models/channel/authorization.rb:56-61`、`core/app/views/api/v1/channels/cli/authorizations/create.json.jbuilder:6`
- 现象：`TokensController#create` 在 `approved` 状态下每次轮询都返回完整 channel token，`device_code` 首次兑换后不消费 / 不失效；`DEFAULT_INTERVAL`（5s）只告知客户端，服务端从不比较 `last_polled_at`，无 RFC 8628 `slow_down`；`channels/cli/*` 与 `push_subscriptions/*` 全无 `rate_limit`（grep 仅命中 session/beeper/ping）。
- 影响：泄露的 `device_code` 可反复换取长期 CLI token；未认证 poll 端点与 `user_code` show/update/destroy 可全速爆破。
- 修复：首次发放后消费 `device_code`（状态置 `consumed` 或轮换）；服务端强制 `last_polled_at + interval` 并返回 `authorization_pending` / `slow_down`；按 `device_code` / IP 加 `rate_limit`。

### H5 [CLI] macOS `osascript` 命令注入

- 位置：`apps/cli/internal/exec/notify.go:21-25`
- 现象：
  ```go
  script := fmt.Sprintf("display notification %q with title %q", body, title)
  ```
  `title` / `body` 来自服务端下发（`channel/channel.go:107-111`）。Go `%q` 不是 AppleScript 转义，`"` / `\` / 换行可跳出字符串执行任意 AppleScript（可再调 shell）。Linux `notify-send` 走独立 argv，安全。
- 影响：恶意服务端（或被污染的 beep 内容）→ macOS 本地 RCE。
- 修复：darwin 侧不用字符串拼接（写临时脚本文件或换 argv 语义的 notifier）；至少做 AppleScript 转义（`\` → `\\`、`"` → `\"`、去控制字符）并限长。

### H6 [CLI] 跨 host 重定向泄露 Channel token

- 位置：`apps/cli/internal/client/client.go:30-38`、`325-326`、`375-376`
- 现象：`CheckRedirect` 跨 host 只删 `X-Runner-Token`，而 `FetchCliInbox` / `AckCliDelivery` 用的 `X-Channel-Token` / `X-CLI-Token` 会被带到重定向目标。
- 影响：恶意服务端（或 dev 明文 HTTP 中间人）回 `302` 到攻击者 host 即窃取 channel token。
- 修复：同样 `Del("X-Channel-Token")`、`Del("X-CLI-Token")`；最好禁止跨 host 重定向或校验 host 白名单。

### H7 [CLI] Hook 无归属 / 权限校验 + 空 root 回退到 CWD

- 位置：`apps/cli/internal/exec/hook.go:48-95`、`apps/cli/internal/workspace/workspace.go:50`、`apps/cli/internal/channel/channel.go:128-132`
- 现象：`DispatchOnBeepHook` 找到候选即 `exec.CommandContext`，不校验属主 / world-writable / 可执行位；workspace 目录 `MkdirAll(..., 0o755)` 世界可读；`workspace == nil` 时传 `wsRoot = ""`，`filepath.Join("", ".beep/hooks/...")` 相对 CWD 解析。
- 影响：同机其他用户可 planted `hooks/on_beep*` → 下次投递 RCE；仓库内恶意 `hooks/on_beep` 被意外执行。
- 修复：拒绝空 `workspaceRoot`；执行前 `Stat` 要求 regular file、属主为 `os.Getuid()`、`mode & 0o022 == 0` 且有可执行位，否则跳过并告警。

---

## Medium

### M1 [Core] `user_code` 低熵 + 未认证可枚举

- 位置：`core/app/controllers/api/v1/channels/cli/authorizations_controller.rb:11-38`、`core/app/models/channel/authorization.rb:4,63-73`
- 现象：`user_code` 8 字符 × 28 字母表（约 38 bit），未认证 `create` 可无限签发，`show`/`update`/`destroy` 对任意登录身份接受猜测，`show` 以 200/404 构成 oracle。`approve!` 绑定 `user.account`。
- 影响：枚举 pending code 后批准，可把受害者 CLI 绑到攻击者 account，或拒绝造成 DoS。
- 修复：四个动作加 `rate_limit`；状态变更尽量同时要求高熵 `device_code`；缩短 TTL、限制单 IP 并发 pending 数。

### M2 [Core] `deny!` 覆盖终态

- 位置：`core/app/models/channel/authorization.rb:50-54`（对比 `approve! :31-32` 有 `pending` 守卫）
- 现象：`deny!` 仅 `return false if expired?`，对已 `approved` 的授权调用 `DELETE` 会翻转为 `access_denied`（channel 还在，状态说谎），`poll!` 随后对已发 token 报拒绝。
- 修复：`return false unless status == "pending" && !expired?`。

### M3 [Core] Web Push allowlist 后缀匹配绕过

- 位置：`core/app/models/channel/web_push.rb:102-105`
- 现象：`host&.end_with?(permitted)` 无点边界，`evilfcm.googleapis.com` 可通过 `fcm.googleapis.com` 校验。另 `SsrfProtection.blocked_address?` 在 development 直接 `false`（`ssrf_protection.rb:69`），本地无 DNS 兜底。
- 修复：`host == permitted || host.end_with?(".#{permitted}")`。

### M4 [Core] Web Push upsert 全表扫描 + 无唯一约束

- 位置：`core/app/models/channel/web_push.rb:37-63`、`core/db/schema.rb:295-299`
- 现象：去重用 `user.channels.where(kind: :web_push).find { |c| c.endpoint == endpoint }`（Ruby 侧全量加载，应走 `Channel.for_endpoint`）；`(user_id, endpoint)` 无唯一索引（仅 `[user_id, kind]` 非唯一），并发 upsert 产生重复。
- 修复：按 `json_extract(config, '$.endpoint')` 查询；加唯一索引或去重键并处理 `RecordNotUnique` 重试。

### M5 [Core] CLI 投递 claim/ack 竞态，`fail!` 丢错误信息

- 位置：`core/app/models/channel/delivery.rb:20-32`、`core/app/controllers/api/v1/channels/cli/inboxes_controller.rb:2-5`、`core/app/controllers/api/v1/channels/cli/deliveries/acks_controller.rb:3-11`
- 现象：`claim!` 先读后写非原子；`InboxesController#show` 取 `due_for_cli.limit(10)` 再 `each(&:claim!)`，并发轮询拿同一批 → 重复投递，且忽略返回值仍渲染；`succeed!`/`fail!` 无 `claimed?` 守卫、无幂等，重放 / 乱序 ack 照收；`fail!(msg)` 只 `update!(status: :failed)`，丢掉 message。
- 修复：原子 claim（`status: pending` 守卫的 `update_all` 并返回成功行）；强制 `pending→claimed→(succeeded|failed)`；落库 error 列。

### M6 [Core] `Beep::Run` 可投递两次

- 位置：`core/app/models/beep/run.rb:39-42,80-82`
- 现象：`claimed || running?` —— 第二个 worker 发现已 `running` 也会进入 `deliver_notifications_now`，造成 web push / email / CLI 重复；`persist_result` 用 `update_columns` 跳过校验 / 锁。
- 修复：只返回 `claimed`；孤儿 `running` 靠 `reclaim_stale_firing` 回收。

### M7 [Core] `approve!` 非事务

- 位置：`core/app/models/channel/authorization.rb:31-48`
- 现象：先 `user.account.channels.create!` 再单独 `update!`，第二写失败则留下带 live token 的孤儿 CLI channel，而授权仍 `pending`。
- 修复：包 `transaction`。

### M8 [Core] `masked_token` 泄露 4 字符随机量

- 位置：`core/app/models/channel.rb:44-48`，`index`/`show`/`update` 视图对所有 account 成员可见
- 现象：`"#{token.first(12)}••••"`，前缀 `beep_ct_` 占 8 字符，等于暴露 4 字符随机量（约 23 bit 熵）。
- 修复：只显示前缀（`beep_ct_••••`）或后 4 位，不暴露前导随机字符。

### M9 [CLI] Token 文件已有时权限不收紧

- 位置：`apps/cli/internal/config/config.go:96-106`
- 现象：`os.WriteFile(path, data, 0o600)` 只对新建文件生效，已存在（如旧版 0644 或用户手动创建）则保持 world-readable，`SaveFile` 从不补 `Chmod`。
- 修复：写后 `os.Chmod(configPath, 0o600)`；工作区目录考虑 `0700`。

### M10 [CLI] Hook 输出随 ack 上传服务端，无大小限制

- 位置：`apps/cli/internal/channel/channel.go:134-142`、`apps/cli/internal/exec/hook.go:90-93`
- 现象：hook 失败错误拼入完整 stdout/stderr，`PollInbox` 以 ack `error` 字段上传，同时写日志文件。本地脚本输出（含 env / secret / PII）外发，且无截断。
- 修复：ack 只发固定格式（exit code + 末尾 N 字节，上限 2–4KB），完整输出仅留本地。

### M11 [CLI] `DisconnectChannel` 未显式带 channel token

- 位置：`apps/cli/internal/client/client.go:395-421`（对比 `:325-326`、`:375-376`）
- 现象：inbox/ack 显式设 `X-Channel-Token` + `X-CLI-Token`，`DisconnectChannel` 只调 `setHeaders`（仅 `cfg.CliToken != ""` 时带 `X-CLI-Token`）。仅 `ChannelToken` 的 `Config`（`cmd/up.go:227-236`、`channel.go:35-44` 合法）会无认证发出；今天能用全靠 `config.Load` 三字段互为别名（`config.go:164-166`）。
- 修复：与 inbox/ack 同样解析 token 并显式设双头，缺失时报明确错误。

### M12 [CLI] `browser.Open` 直接打开服务端 URL

- 位置：`apps/cli/internal/browser/browser.go:9-19`、`apps/cli/internal/channel/channel.go:62-71`
- 现象：`VerificationURIComplete` / `VerificationURI` 不经校验直传 `xdg-open` / `open` / `rundll32`（argv 无 shell 注入，但 scheme/handler 可被滥用，如 `file://` / 自定义 scheme）；失败错误被丢弃（`:71`），测试无法感知。
- 修复：解析 URL，仅允许 `https:`（localhost 例外 `http:`，与 `requireSafeServerURL` 一致），打开前提示确认；不要吞错。

### M13 [CLI] Device 轮询：无 SIGINT、边界与 error 类型脆弱

- 位置：`apps/cli/cmd/channel.go:75-91,127-145`
- 现象：`ctx` 无 `signal.NotifyContext`，Ctrl-C 靠杀进程；`interval < 2s → 5s` 无视服务端合法 `interval=1`；`ExpiresIn=0`/缺失则 deadline 即现在 → 秒过期；`err.(*client.OAuthErrorResponse)` 类型断言遇 `fmt.Errorf %w` 包裹即失效，掉进无限“网络错误重试”。
- 修复：接 `signal.NotifyContext`；非正 `ExpiresIn` / `Interval` 给默认（900/5）；用 `errors.As`；非 OAuth 错误加最大重试 / backoff 上限。

### M14 [CLI] Runner 单 tick 无界 drain

- 位置：`apps/cli/internal/runner/runner.go:100-130`
- 现象：`PollAndExecute` 循环 `Poll → execute` 无单 tick 上限、无 sleep，服务端持续返回任务即高频 POST，打满 server 并饿死 shutdown（仅靠 poll 错误退出）。
- 修复：单 tick 上限（如 `2*concurrency`），到上限即回 ticker。

### M15 [CLI] deliveryID / slug 直接拼 URL path

- 位置：`apps/cli/internal/client/client.go:362`、`:130`
- 现象：`fmt.Sprintf("%s/api/v1/channels/cli/deliveries/%s/ack", ..., deliveryID)`，`deliveryID` 服务端可控，`/`、`?`、`#` 可改请求路径；`DeleteJob(slug)` 同理。
- 修复：`url.PathEscape(deliveryID)` / `url.PathEscape(slug)`。

### M16 [Web] `/device?code=` 自动验证可被钓鱼利用

- 位置：`apps/web/src/routes/device.tsx:22-30,45,74-78`
- 现象：`code` 取自 URL 并自动 `handleVerifyCode`，OTP 留在 history / 日志 / Referer，验证后不清除 search。
- 攻击：攻击者在自己机器 `beep channel connect` 得码，发受害者 `https://web…/device?code=ABCD-EFGH`，受害者登录态打开并点 Authorize → 攻击者设备绑到受害者 account（`handleApprove :80-97` 用 URL 传入的 `user_code`）。
- 修复：URL code 不自动验证（至少明确提示来源并二次确认），验证后 `router.replace({ search: {} })`，渲染 `expires_in`（已取未展示）助用户识别过期码。

### M17 [Web] 一次性 channel token 常驻 state，跨 account 泄露

- 位置：`apps/web/src/components/settings/channel-management-settings.tsx:30,62,134-158`
- 现象：`createdChannel`（含完整 `token`）创建后不清；切换 `slug` 触发 `load()`（`:36-45`）但不清除，account A 的 secret 留在 account B 页面 DOM（`CopyableCode` + `beep config set channel_token {token}`）。
- 修复：`useEffect(() => setCreatedChannel(null), [slug])`；加 Dismiss；提示语不重印原文（copy 按钮持有即可）或掩码。

### M18 [Web] Banner 边启用边跳转，中断授权提示

- 位置：`apps/web/src/components/settings/web-push-setup-banner.tsx:76-79`（结合 `:40-42` settings 下 `fuzzy: true` 即 unmount）
- 现象：点击先 `enable()`（含 `Notification.requestPermission()`）紧接着 `openSettings()` 跳走并 unmount，权限弹窗被打断，`error` 态（`:114-118`）随组件卸载丢失。
- 修复：二选一 —— `await enable()` 仅失败 / 拒绝才跳转，或 banner 只做纯链接，启用逻辑留给 `WebPushSettings`。

---

## Low

### L1 [Core] `Channels#create` 强参数形状错误 + `kind` 无约束

- 位置：`core/app/controllers/api/v1/channels_controller.rb:35-37`、`core/app/models/channel/webhook.rb:1-16`、`core/app/models/channel/email.rb:1-19`
- 现象：`permit(:name, :kind, :config)` 把 `:config` 当标量，`web_push` 的 endpoint/key 嵌套被 strip（静默 `{}`）；同时 `:kind` 放行 `webhook` / `email`，而其 `validate_config` 为空 stub（webhook 全 TODO）。可建出惰性 / 误导性 channel；`Email.deliver_test!` 无操作却回 204。
- 修复：按 kind 放行 `config: [...]` 形状（或走 kind 专属端点），`kind` 明确 allowlist，未实现 `deliver_test!` 应抛 422。

### L2 [Core] 邮件 subject 取未消毒 title

- 位置：`core/app/mailers/beep_mailer.rb:11`、`core/app/models/beep.rb:20,27`
- 现象：`subject: @beep.title`，title 仅 presence/length/strip，无换行 / CRLF 拒绝（Mail gem 会编码，属 hardening，非确认注入）。
- 修复：`validates :title, format: { without: /[\r\n]/ }`。

### L3 [Core] Backfill `down`  destructive + 并发建码可 500

- 位置：`core/db/migrate/20260911110000_backfill_email_and_web_push_channels.rb:44-46`、`core/app/models/channel/authorization.rb:23-25,63-73`
- 现象：`down` 直接 `DELETE FROM channels WHERE kind IN ('email','web_push')`，回滚即删真实用户数据；`generate_user_code` 的 `exists?` 循环并发下竞态，DB 唯一索引（`schema.rb:265` 有）会转成未捕获 `RecordNotUnique` 500。
- 修复：`down` 改 no-op 或备份表恢复；`create_request!` rescue `RecordNotUnique` 重试。

### L4 [Core] `user_code` 进 URL query

- 位置：`core/app/views/api/v1/channels/cli/authorizations/create.json.jbuilder:4`
- 现象：`verification_uri_complete` 拼 `?code=<user_code>`，进浏览器历史 / 服务端日志。TTL 15 分钟尚合理。
- 修复：`/device` 页让用户手输 code，或接受日志风险并保持短 TTL。

### L5 [CLI] Runner 端点附带多余 CLI token + `--token` 进子进程列表

- 位置：`apps/cli/internal/client/client.go:279-289`、`apps/cli/cmd/up.go:279-298`、`apps/cli/cmd/runner.go:111-117`
- 现象：`setHeaders` 给 runner 端点（Ping/Poll/jobs）也带 `X-CLI-Token`（与 channel token 互为别名），暴露面过大；`up` 重启 daemon 原样转发 argv，`--token <secret>` 留在 `ps`。
- 修复：token 按端点族各发各的；子进程参数剥掉 `--token` / `-t`（走 env / 配置文件）或至少文档警示。

### L6 [CLI] Socket 先解码后验权；非 Linux 无校验；短暂宽权限窗口

- 位置：`apps/cli/internal/daemon/socket.go:70-80,169-174`、`apps/cli/internal/daemon/peercred_other.go:9-11`
- 现象：对端 status 先 JSON 解码后 `verifyPeerCredentials`；macOS 等为 no-op，任意本地用户可伪造 daemon status（含 `status` 显示的 PID）；socket 默认 umask 创建后 `Chmod 0600`，中间有短暂世界可访问窗口；stale-socket `Remove` + `Listen` 有 TOCTOU（双启动其一失败，属良性）。
- 修复：能验先验；文档注明仅 Linux 强制；`Listen` 前后 `unix.Umask(0o077)` 或确保工作区 `0700`（`AcquireSocket` 已有，但 `workspace.Open` 预建父目录 `0755`）。

### L7 [CLI] Runner `workspace` nil 即 panic

- 位置：`apps/cli/internal/runner/runner.go:53-58`（对比 channel 侧有 nil 检查）
- 现象：`Run` 无条件 `r.workspace.Root`，当前调用方皆非空，测试 / 未来调用方一疏忽即 panic。
- 修复：同 channel 加 nil 检查，或 `New` 要求非空。

### L8 [Web] `load()` 竞态、timer 泄露、delete 未防抖

- 位置：`apps/web/src/components/settings/channel-management-settings.tsx:36-45,72-96,230-238`
- 现象：无 `AbortController` / request-id，多次 `load()`（create+delete+切 slug）可 last-write-wins 脏数据，可 unmount 后 `setState`；`setTimeout(...3000)`（`:88-90`）卸载不清；`handleDelete` 用阻塞 `confirm()`（硬编码英文，绕 i18n），按钮缺 `type="button"` 且忙时不禁用，连点双发。
- 修复：busy 态禁用双按钮 + `pendingDeleteId`；mount 标记 / `AbortController`；timer ref + cleanup `clearTimeout`。

### L9 [Web] device 页状态陈旧、`expires_in` 未用、成功页回显本地名

- 位置：`apps/web/src/routes/device.tsx:45,74-78,124-131,168-175`、`apps/web/src/lib/api/channels.ts:70`
- 现象：`code` state 仅初始化一次，URL 后变不同步；`expires_in` 已取从不展示，过期码只有原始 `ApiError.message`；成功屏回显本地 `channelName` 输入而非服务端返回名。
- 修复：search 变化同步 / 清空 `code`；渲染倒计时；成功态用服务端 `channel.name`。

### L10 [Web] Copy 按钮谎报成功

- 位置：`apps/web/src/components/ui/copy-button.tsx:70-77`
- 现象：`void navigator.clipboard.writeText(text); setCopied(true)` 不 await、无 catch，剪贴板被拒也打勾。此处拷贝的是 channel token，影响大。
- 修复：`await` / `.then/.catch`，成功才 `setCopied(true)`，失败提示。

### L11 [Web] `accountSlug` / `channelId` 未编码（既有模式）

- 位置：`apps/web/src/lib/api/channels.ts:26-27,39,52,61`
- 现象：`kind` 编码了但路径段未编码（`/api/v1/${accountSlug}/channels/${channelId}`），与 `beepers.ts` / `runners.ts` / `beeps.ts` 既有模式一致；slug 正常 `[a-z0-9-]` 低风险，构造 `$account_slug`（`..` / `/` / `%2f`）可路径注入（客户端侧）。
- 修复：新 `channels.ts` 用 `encodeURIComponent(accountSlug)` / `encodeURIComponent(channelId)`，顺手 elsewhere。

---

## 已查无问题（免回查）

- SQL 注入：唯一原生 SQL `Channel.for_endpoint` 参数化，无注入。
- XSS：`apps/web/src` 无 `dangerouslySetInnerHTML` / `innerHTML`，新建 channel/device 名均为 JSX 文本，React 转义。
- 认证传输：新建 API 全走 `apiFetch`（`lib/api/client.ts:64`），`credentials: "include"` 集中设置。
- Web Push 密钥 / SW：本 diff 无 VAPID / SW scope 变更（`lib/web-push.ts:11,234-237` 未动）。
- CLI TLS：默认 `http.Client`，无 `InsecureSkipVerify`。
- CLI Hook 经 payload 的 shell 注入：`hook.go:85` 用 `exec.CommandContext(hookPath)` 无 shell，payload 走 env，不可利用（仅 macOS `osascript` 例外，见 H5）。
- CLI 回调 SSRF：`ReportLog` / `ReportResult` 有同 scheme+host+`/api/v1/runner/tasks/` 前缀校验（`client.go:214-233`）。
- CLI `up.go` runner/channel error 变量：各 goroutine 写各自变量，主协程 `wg.Wait()` 后读，有 happens-before，非 data race。
- CSRF / IDOR（需后端确认，非本 diff 可单修）：`lib/api/client.ts:41-67` cookie 无 `X-CSRF-Token`，`apps/web/src` 无 `csrf` 字样；若 Core 靠 cookie session 且无 `SameSite=strict` / CSRF token，新 `POST/PATCH/DELETE` 可被跨站伪造。新路由 slug 直传 API，无客户端成员检查，授权必须服务端强制（前端 401/403 直显 `withAuthRedirects` 正确），错误文案避免区分“无此 account”与“无权”。

## 待修建议顺序

1. H1（构建断）→ 2. H2/H3/H4（越权 / 伪造 / 爆破）→ 3. H5/H6/H7（本地 RCE / token 泄露）→ 4. M16/M17（钓鱼绑定 / token 常驻）→ 5. M3/M5/M6（allowlist / 重复投递）→ 6. 其余 Medium/Low。
