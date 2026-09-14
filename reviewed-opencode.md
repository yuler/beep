# Review 记录 — feature/channel（opencode）

- 分支：`feature/channel`，范围 `main...HEAD`（124 文件，约 +6356 / -736）。
- 方法：三路并行实读（core / web / cli）+ 抽查验证。对 `reviewed-muse.md` / `reviewed-agy.md` 声称已修的项逐项复核，重点看最新 4 个提交（`144e604`、`6a5be1b`、`c25c644`、`d1e131c`）引入的新问题。
- 抽查确认：`channels.ts:110-117` deny 未走 account 路径、`beep/run.rb:31/39-42/84-85` 多接收人守卫与 claim 语义、`beep.rb:159-161` 单接收人现状、`up.go:345-371` token 转发，均已亲自读码。

## 结论

| 等级   | 数量 | 说明                                         |
|--------|------|----------------------------------------------|
| High   | 7    | 并发重复投递、多接收人静默丢弃、approve 双建、deny 路径不对称、CSRF、无 hook 误报送达、token 进 ps |
| Medium | 14   | 频控分桶、ack 非原子、显式 channel 重试重复、show/destroy 跨账号、push 未编码、code 跨 slug 漫游等 |
| Low    | 8    | 空 test 实现、日志前缀、UTF-8 截断、browser 大小写等 |

最先修：C-H2（多接收人丢弃是静默丢数据）、C-H1（并发重复投递）、W-H2（CSRF 二选一必居其一）、L-H2（token 进 ps）。

---

## High

### C-H1 [Core] `claim_delivery?` 的 `|| running?` 导致并发重复投递

- 位置：`core/app/models/beep/run.rb:39-42`
- 现象：
  ```ruby
  def claim_delivery?
    claimed = self.class.where(id: id, status: :pending).update_all(status: "running", updated_at: Time.current) == 1
    claimed || running?
  end
  ```
  两个 `DeliverBeepRunJob` 并发时，第一个 `pending→running`，第二个 `update_all==0` 但 `running?` 为 true（job 经 GlobalID 重新 find，读到 `running`），同样进入 `deliver_notifications_now`。而 `deliver_to`（如 `Channel::Cli.deliver_beep`）每次都会 `create!` 新 `Channel::Delivery`，即重复建投递行 + 重复推送。
- 影响：重复打扰/重复计费副作用；重试语义被污染（无法区分“并发重复”与“合法重试”）。
- 说明：`144e604` 从 muse 版的“只认 claimed”改回 `claimed || running?`，修好了重试测试（`run_deliver_test.rb:113`），但把并发重复投递又带了回来。两边各对一半。
- 修复：只返回 `claimed` 做 single-flight；合法重试走显式路径（retry 时 `with_lock` 重查，或 `running?` + 结果幂等 guard 双条件），不要用一个表达式同时表达两个语义。

### C-H2 [Core] 多接收人通知被静默丢弃（agy U2，当前仍在）

- 位置：`core/app/models/beep/run.rb:31-33,84-85,95-96,116-117`
- 现象：`payload_result` 跨 user 复用，`deliver_web_push` / `deliver_cli` 以顶层 key（`key?("web_push")` / `key?("cli")`）、email 以 `dig("email","status")` 做幂等守卫。第 2 个 user 进循环直接全部 `return`。
- 影响：目前被 `beep.rb:159-161` 的 `recipient_users == [account.owner_user]` 掩盖（一直接收人恒为 1）；一旦扩展为多成员，第 2 人起的 web_push/email/cli 全部跳过且无日志无错误。静默丢数据，比抛错更危险。
- 修复：结果按 user 隔离（如 `payload_result[user.id]["web_push"]`），或每 user 独立 `payload_result` 再合并。

### C-H3 [Core] `approve!` 非原子，并发双重 approve 建出两个 channel

- 位置：`core/app/models/channel/authorization.rb:37-56`
- 现象：内存检查 `return false if expired? || status != "pending"` 通过后，`transaction` 内先 `channels.create!` 再 `update!`。两个并发 `PATCH` 都通过检查，各建一个 `cli` channel，后写覆盖 `channel_id`，前一个成带 live token 的孤儿。
- 影响：孤儿 CLI token 泄漏归属；与 H4 同类（claim 原子了，approve 没原子）。
- 修复：`with_lock` 或先 `where(id:, status: "pending").update_all(status: "approved")==1` 抢占，抢到才建 channel。

### C-H4 [Core] `succeed!/fail!` 非原子（claim 原子了，ack 没原子）

- 位置：`core/app/models/channel/delivery.rb:31-43`
- 现象：`return false unless pending? || claimed?` 是内存判断，随后 `update!`。并发 ack（或 inbox claim 与 ack 竞争）都通过检查后依次写入，终态看时序。
- 修复：`where(id:, status: %w[pending claimed]).update_all(...) == 1` 条件写。

### W-H1 [Web] `denyDeviceAuth` 未走 account-slug 路径，与 approve 不对称

- 位置：`apps/web/src/lib/api/channels.ts:110-117`；调用方 `apps/web/src/routes/$account_slug/device.tsx:100`
- 现象：`approveDeviceAuth(accountSlug, ...)` 已是 `/api/v1/:slug/channels/cli/authorizations/:code`（d1e131c 修好），但 `denyDeviceAuth(user_code)` 仍是全局 `/api/v1/channels/cli/authorizations/:code`，不带 slug。
- 影响：deny 绕过“URL slug → 成员身份”约束；结合后端 `show/destroy` 全局查 code（C-M4），任意猜中 user_code 的登录用户可 deny 他人 device flow（DoS）。即使服务端有校验，前后端契约不一致迟早一方改错。
- 修复：`denyDeviceAuth(accountSlug, user_code)`，路径与 approve 对称；后端 destroy 同样要求 account scope。

### W-H2 [Web] Cookie 会话 + 无 CSRF Token 的变异请求（二选一必居其一）

- 位置：`apps/web/src/lib/api/client.ts:41-67`（`credentials: "include"`，无 `X-CSRF-Token`）；服务端 `core/app/controllers/concerns/request_forgery_protection.rb:7,19-23`（`header_only`，仅 `Sec-Fetch-Site` + `Sec-Fetch-Mode` 双缺失才放行）。
- 现象：`createChannel` / `deleteChannel` / `testChannel` / `approve/denyDeviceAuth` / push 订阅增删测，全是 cookie 凭据 POST/PATCH/DELETE，前端从不发 CSRF 头。
- 影响：(a) 若浏览器请求被正常校验 → 生产直接 422（功能性 BUG）；(b) 若靠 Sec-Fetch 缺失放行 → CSRF 防护依赖浏览器指纹头而非法线 token。必居其一。
- 修复：服务端下发、前端 `client.ts:request()` 统一附 `X-CSRF-Token`（推荐）；或明确 SameSite + 自定义头并两端写死，加回归测试。

### L-H1 [CLI] c25c644 后无 hook / nil workspace 仍 ACK `succeeded`，且桌面通知已删

- 位置：`apps/cli/internal/channel/channel.go:139,145-154`；`apps/cli/internal/exec/hook.go:44-46`；`notify.go` 已删除
- 现象：`DispatchHook` 空 root 直接 `("", "", nil)`；`hookName == ""` 只打一条 log，随后 `_ = AckCliDelivery(delivery.ID, "succeeded", "")`。c25c644 删掉了 `NotifyDesktop` 与 `if title != "" || body != ""` 分支。
- 影响：零 hook 用户从“至少有 OS 弹窗”退化为“一条 log + 服务端记已送达”。nil workspace 同理静默丢。`channel.test` 也一样，测试页点“发送”显示成功但用户侧无任何感知。
- 修复：无 hook 时 ACK `failed`/`skipped` 或不 ACK（保留到过期）；或恢复可选桌面通知（默认开，加 `--no-notify`）。

### L-H2 [CLI] `--token/-t` 原样转发给子进程，`ps` 可见（muse L5 声称修好，实未修）

- 位置：`apps/cli/cmd/up.go:361-370`（takesArg 名单含 `-t/--token` 但照样 append 进 flags）；`startServiceBackgroundDaemon:292-293` 的 `exec.Command(exe, childArgs...)`
- 现象：`beep up -d -t <secret>` 进子进程 argv，同机 `ps -ef` 泄露长期 token。
- 修复：`buildServiceChildArgs` 剥掉 `-t/--token` 及其值（走 env/配置文件），加测试断言输出不含 secret。

---

## Medium

### C-M1 approve 名过长抛 500，无 422 映射

- 位置：`core/app/controllers/api/v1/channels/cli/authorizations_controller.rb:28` + `authorization.rb:42`
- 现象：`approve!` 内 `create!` 失败抛 `RecordInvalid`，controller 只处理返回 false，未 rescue → 500。
- 修复：`rescue ActiveRecord::RecordInvalid → 422 VALIDATION_ERROR`。

### C-M2 显式 channel ID 分支无幂等守卫，重试重复投递

- 位置：`core/app/models/beep/run.rb:68-75,131-139`
- 现象：`deliver_web_push/cli/email` 都有 guard，但 `deliver_explicit_channel` 只是追加，无跳过检查。email 失败触发 job retry 后显式 channel 再投一次。
- 修复：在 `payload_result` 记录已投递 `channel.id` 并跳过。

### C-M3 device token 频控按 code 分桶，轮换 code 即绕过

- 位置：`core/app/controllers/api/v1/channels/cli/authorizations/tokens_controller.rb:4-6`
- 现象：`rate_limit by: device_code || ip`，攻击者轮换 code 即得全新配额，IP 级防护失效。
- 修复：按 IP 限流，或复合键 `"#{ip}:#{device_code}"`。

### C-M4 `show/destroy` 全局查 code，任意登录用户可探查/deny 他人 flow

- 位置：`core/app/controllers/api/v1/channels/cli/authorizations_controller.rb:6,17,37`（`skip_account_scope only: %i[create show destroy]` + `active.find_by(user_code:)`）
- 现象：`update` 已要求 account scope（d1e131c），但 `show` 泄漏 `channel_name/status`，`destroy→deny!` 允许任意登录用户 deny 任意 code。user_code 仅 8 字符，按 code 限流减缓但不消除跨账号干扰。
- 修复：`destroy` 同样要求 account scope，或 deny 仅允许同账号语义。

### C-M5 `claim!` 未原子校验过期

- 位置：`core/app/models/channel/delivery.rb:20-29`
- 现象：内存 `return false if expired?`，`WHERE` 无 `expires_at` 条件；`due_for_cli` 只管读不管写，读到 claim 之间过期仍被 claim。
- 修复：`update_all` 加 `where("expires_at > ?", Time.current)`。

### C-M6 CLI ack 404 错误格式不一致

- 位置：`core/app/controllers/api/v1/channels/cli/deliveries/acks_controller.rb:3`；`base_controller.rb:1`（`ActionController::API`，无 `Api::V1::BaseController` 的 `rescue_from RecordNotFound`）
- 现象：`@current_channel.deliveries.find` 抛 `RecordNotFound`，回退框架默认处理而非 JSON 404。
- 修复：`Cli::BaseController` 加 `rescue_from ActiveRecord::RecordNotFound`。

### C-M7 Web Push upsert 竞争产生重复行 + SQLite 方言

- 位置：`core/app/models/channel/web_push.rb:37-63`；`channel.rb:27`（`json_extract`）
- 现象：无唯一索引，并发 upsert 同一 endpoint 都 `first→nil` 各插一条；`for_endpoint` 用 `json_extract`，Postgres 下直接 SQL 错误。
- 修复：endpoint 唯一/函数索引 + `rescue RecordNotUnique retry`。

### W-M1 `push.ts` 路径段未编码，与 `channels.ts` 不一致

- 位置：`apps/web/src/lib/api/push.ts:22-51`（slug/id 全直接拼接）；对比 `channels.ts:26,39,53,65,81,102,112` 全量 `encodeURIComponent`
- 修复：`push.ts` 全部加 `encodeURIComponent`。

### W-M2 换账号时 `code` 被原样携带，放大钓鱼面

- 位置：`apps/web/src/routes/$account_slug/device.tsx:109-113,156-163`（`switchSearch` 把 `user_code ?? search.code` 带回 `/device` 再进另一 slug）
- 现象：结合 W-M3 的自动验证，切一次账号即在新账号下自动验证同一 code。
- 修复：Switch 链接不带 `code`，或切换后强制手动重输。

### W-M3 `$account_slug/device.tsx` 仍从 URL 自动验证且不清 search（M16 部分回归）

- 位置：`apps/web/src/routes/$account_slug/device.tsx:67-72`（`search.code` 即 `handleVerifyCode`）；验证后无 `navigate({search:{}})` 清理
- 现象：`https://web/<victim-slug>/device?code=ATTACKER-CODE` 打开即自动验证攻击者通道名，受害者再点 Authorize 即把攻击者 CLI 绑进自己账户（RFC 8628 §8.4 phishing）；code 留在 URL，进历史/录屏/转发。
- 修复：URL code 仅预填输入框不自动 verify（或自动 verify 但 Authorize 前二次确认归属 account）；成功后清 search；已有 `expires_in` 展示（Verified）+ 加“勿点陌生 code 链接”提示。

### L-M1 runner 端点附带 channel token（L5 残留）

- 位置：`apps/cli/internal/client/client.go:281-289`（`setHeaders` 无条件 `X-CLI-Token`）；`config.go:169-171`（别名）
- 修复：按端点族设头（runner 族只发 `X-Runner-Token`，channel 族只发 channel 头），或 `setHeaders(kind)` 参数。

### L-M2 device 授权/轮询请求附带已有 token

- 位置：`apps/cli/internal/client/client.go:475-499`（`RequestDeviceAuthorization` / `PollDeviceToken` 经 `setHeaders`）
- 修复：这两个用不带 auth 头的构造。

### L-M3 回调 POST 跟随跨站重定向，body 照发

- 位置：`apps/cli/internal/client/client.go:30-40,193-214`（`CheckRedirect` 只剥 token 头；`ReportLog/ReportResult` 默认跟随 307/308；SSRF 校验只管首跳 `allowedCallbackURL:216-235`）
- 修复：回调专用 `http.Client{CheckRedirect: http.ErrUseLastResponse}`，重定向即报错；补 307 测试。

### L-M4 socket 先解码后验权 + umask/目录权限窗口（L6 残留）

- 位置：`apps/cli/internal/daemon/socket.go:70-78,169-174`；`workspace.go:50`（`MkdirAll 0755`）
- 现象：`json.Decode` 在 `verifyPeerCredentials` 之前；`Listen` 后才 `Chmod 0600`；workspace/hooks/jobs 默认 0755。
- 修复：先验权再读；`Listen` 前 `unix.Umask(0o077)`；workspace 根改 `0700`。

---

## Low

- L1 [Core] `Channel::Email#deliver_test!` 仍空实现（`channel/email.rb:13-14`）：`POST channels/:id/test` 对 email 回 204 误导“已发送”。发测试邮件或对 email 回 422。
- L2 [Core] `web_push.rb:86-94` 通道名直接存截断 UA（含 PII），`curl` 特判无意义。脱敏/映射设备名。
- L3 [Core] `delivery.rb:14` `stale_pending` 无任务消费，过期 CLI 投递永久堆积。定时 `pending→expired`。
- L4 [CLI] `channel.test` 误跑遗留 `on_beep_fired`（`hook.go:18-21` 全局 `on_channel > on_beep > on_beep_fired`）：只有老 hook 的用户点“测试通道”触发生产动作。`channel.test` 只选 `on_channel*`。
- L5 [CLI] `BEEP_EVENT*` 来自服务端 payload 未消毒即进 hook env（`hook.go:47-82`）：脚本若 `eval "$BEEP_EVENT_JSON"` 即注入。示例 + `docs/architecture/channel.md` 加“勿 eval”红字；`BEEP_EVENT` 加 allowlist。
- L6 [CLI] `buildServiceChildArgs` verb 过滤误伤值（`up.go:356`）：`beep up --hostname channel`（`--hostname` 不在 takesArg 名单）值被吞。用 cobra 解析后重建 argv 或补全名单 + 测试。
- L7 [CLI] `slow_down` 无上限 + 非 OAuth 错误无 backoff（`cmd/channel.go:139-152`）：interval 膨胀到分钟级；抖动网络空转到 900s deadline。cap 30s + 指数 backoff。
- L8 [CLI/Web] 日志与小问题：`DispatchHook` 永返 `"on_channel"`（`hook.go:58,94,96`，应返实际 basename）；`up.go:272` 残留 `[beep-cli]`（channel 侧已统一 `[beep-channel]`）；ACK 按字节截断 UTF-8（`client.go:363-365` + `hook.go:99-105` 双重截断，应按 rune 且只截一处）；`browser.go:17` scheme 大小写敏感 + 传 rawURL 未 trim（`channel.go:73` 还吞掉 open 错误）；`config.go:68-73` `~otheruser` 误映射（仅处理 `~` / `~/`）。

---

## 已修好（本轮实读确认，不必回查）

- 越权删除/测试：`channels_controller.rb:32`、`tests_controller.rb:11` 均为 `where(user: Current.user)`。OK。
- Device 一次性消费原子：`authorization.rb:64-72` 条件 `update_all`。OK（并发 approve 仍见 C-H3）。
- slow_down + rate_limit 存在：`tokens_controller.rb:20-23` + 两处 `rate_limit`。OK（分桶键见 C-M3）。
- deny 终态守卫：`authorization.rb:58-62`。OK。
- user_code 熵 + 冲突重试：28 字符集 ×8 + `create_request!` rescue `RecordNotUnique`。OK。
- Delivery claim 原子：`delivery.rb:23`。OK（过期条件见 C-M5）。
- fail 落库 error：`delivery.rb:37-43` merge + truncate。OK。
- Web Push allowlist 点边界：`web_push.rb:102-105`。OK。
- Beep source 伪造已关：`beeps_controller.rb:47-49` + `beep.rb:189-195`。OK。
- N+1：`channels_controller.rb:5` `includes(:user)`。OK。
- migration down 非破坏：backfill `down` 为 no-op。OK。
- 级联删除：`channel.rb` `dependent: :destroy`（U1）。OK。
- `*.localhost`：`browser.go:34-40` + 测试。OK。
- H5 osascript：`notify.go` 已删除，无其他引用。OK（代价见 L-H1）。
- H6 跨 host 剥离三头：`client.go:34-38`。OK。
- H7 hook 属主/权限/空 root：`hook.go:44-45,107-124`。OK。
- M11 双头、M15 PathEscape、M9 0600/0700、M13 signal/errors.As/默认、M12 browser 校验、M10 ack 截断、M14 单 tick 上限：均 OK（细节残留见 L3/L7/L8）。
- Web：approve 走 slug 路径、新旧 device 页无逻辑分叉（旧页现为 picker）、slug 门禁 `beforeLoad`、channels 全量编码、`testChannel` 语法、M17 跨 slug 清 token + Dismiss、copy-button 成功态、banner 不再边启用边跳转、`expires_in` 展示、无 `dangerouslySetInnerHTML`、i18n 键齐备：均 OK。
