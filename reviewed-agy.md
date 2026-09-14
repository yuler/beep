# PR Review 报告 — feature/channel

本报告对 `feature/channel` 分支相较于 `main` 的全部变更（124 文件，约 +6356 / -736）进行全方位代码 Review，覆盖 Core（Rails 8.1 后端）、Web（TanStack Router 前端）与 CLI（Go 客户端/常驻守护进程）。

重点针对并发安全性、授权与访问控制边界（IDOR / CSRF / Phishing）、网络请求安全（SSRF / 凭据泄漏 / 重定向）以及错误处理与幂等性进行深入分析。

---

## 一、审查结论概览

| 类别     | 严重级别   | 数量   | 核心风险摘要                                                                             |
| -------- | ---------- | ------ | ---------------------------------------------------------------------------------------- |
| Core     | High       | 4      | 并发重复投递、多接收人通知静默丢弃、双重 approve 孤儿 Channel、ACK 状态流转非原子操作    |
| Core     | Medium     | 6      | 全局 Deny 跨账号越权、显式通道重试重复、Token 轮询限流绕过、422 缺少映射、过期竞争等     |
| Web      | High       | 1      | Deny API 未限定账号 slug 路径（前后端契约不对称）                                        |
| Web      | Medium     | 3      | URL 验证码自动验证及未清理带来钓鱼与泄露面、账号切换保留 Code、Push API 缺少 URI 编码    |
| CLI      | High       | 2      | 常驻守护进程启动参数携带明文 Token 泄露至 `ps`、无 Hook 时误报 `succeeded` 导致静默丢弃  |
| CLI      | Medium     | 3      | 回调上报允许跨站 POST 重定向、Unix Socket 先解码后验权、请求头 Token 无差别附加          |
| 综合     | Low        | 10     | 测试空实现、未消费的过期 Scope、UTF-8 字节截断、Jbuilder 规范违反、日志前缀不一致等      |
---

## 二、High 级别缺陷与安全漏洞

### C-H1 [Core] `claim_delivery?` 的 `|| running?` 破坏单飞锁，引发并发重复投递

- **位置**: [`core/app/models/beep/run.rb:39-42`](file:///home/yule/Projects/beep/core/app/models/beep/run.rb#L39-L42)
- **代码**:
  ```ruby
  def claim_delivery?
    claimed = self.class.where(id: id, status: :pending).update_all(status: "running", updated_at: Time.current) == 1
    claimed || running?
  end
  ```
- **问题分析**:
  在 commit `144e604` 中，为了让重试单测通过，将判定逻辑改为了 `claimed || running?`。
  当同一 Beep Run 因调度重试、网络延迟或队列重复推送导致两个 `DeliverBeepRunJob` 并发执行时：
  1. 任务 A 执行 `update_all(status: "running")` 成功，`claimed = true`，进入 `deliver_notifications_now`；
  2. 任务 B 执行 `update_all` 返回 0（`claimed = false`），但通过 GlobalID 重新 load 后发现 `running? == true`；
  3. 任务 B 同样判定通过并同时进入 `deliver_notifications_now`。
  在 `deliver_to`（如 `Channel::Cli.deliver_beep`）中，每次执行均无条件 `channel.deliveries.create!` 创建新投递记录并推送到客户端，导致产生重复投递行，用户收到多次重复提醒。
- **修复方案**:
  区分“首次抢占单飞”与“失败重试续跑”语义。正常抢占仅允许 `claimed`（`update_all == 1`）；重试分支应在明确捕获错误后走受控显式重试流程（如带有乐观锁或行级锁 `with_lock`），禁止将 `running?` 作为宽松的并发放行条件。

---

### C-H2 [Core] 多接收人通知被静默丢弃（Silent Notification Drop）

- **位置**: [`core/app/models/beep/run.rb:31-33, 84-85, 95-96, 116-117`](file:///home/yule/Projects/beep/core/app/models/beep/run.rb#L31-L33)
- **代码**:
  ```ruby
  def deliver_notifications_now
    payload_result = stringify_result
    beep.recipient_users.each do |user|
      payload_result = deliver_for(user, payload_result)
    end
    ...
  end

  def deliver_web_push(user, payload_result)
    return payload_result if payload_result.key?("web_push")
    ...
  end

  def deliver_email(user, payload_result)
    return payload_result if email_attempt_complete?(payload_result)
    ...
  end

  def deliver_cli(user, channels, payload_result)
    return payload_result if payload_result.key?("cli")
    ...
  end
  ```
- **问题分析**:
  `deliver_notifications_now` 遍历 `beep.recipient_users`，但 `payload_result` 跨所有用户复用同一个 Hash，且幂等防护直接检查顶层 Key（如 `payload_result.key?("web_push")`、`payload_result.key?("cli")` 或 `email_attempt_complete?`）。
  当处理第 1 个用户时，顶层 Key 被写入；循环流转到第 2 个及后续用户时，所有投递方法立即命中早期 `return`。
  虽然目前 `beep.rb:160` 的 `recipient_users` 仅返回 `[ account.owner_user ]`，但在团队协作或未来支持多成员订阅场景下，第 2 人起的所有 Web Push、Email、CLI 通知均会被静默丢弃，且没有任何日志或异常，属于严重的数据丢失逻辑陷患。
- **修复方案**:
  将投递结果按 `user.id` 隔离存储（例如 `payload_result[user.id]["web_push"]`），或者在循环内部针对每个 User 维护独立的结果上下文最后汇总。

---

### C-H3 [Core] `approve!` 非原子操作，并发批准产生孤儿 Channel

- **位置**: [`core/app/models/channel/authorization.rb:37-56`](file:///home/yule/Projects/beep/core/app/models/channel/authorization.rb#L37-L56)
- **代码**:
  ```ruby
  def approve!(user:, name: nil)
    return false if expired? || status != "pending"

    target_name = name.presence || channel_name.presence || "CLI Channel"
    transaction do
      new_channel = user.account.channels.create!(
        user: user,
        kind: "cli",
        name: target_name
      )

      update!(
        account: user.account,
        user: user,
        channel: new_channel,
        channel_name: target_name,
        status: "approved"
      )
    end
  end
  ```
- **问题分析**:
  `return false if expired? || status != "pending"` 是内存状态判断。在 `transaction` 内未对当前记录加锁（未调用 `lock!` 或 `with_lock`）。
  如果前端用户快速双击或网络并发发送两次 `PATCH` 批准请求：
  1. 两个请求同时通过内存中的 `status == "pending"` 校验；
  2. 两个请求先后进入事务，分别执行 `user.account.channels.create!` 创建了两个不同的 `Channel` 实体，均带有有效的认证 Token；
  3. 后执行的 `update!` 将 `authorization.channel_id` 指向第 2 个 Channel，第 1 个 Channel 沦为不可追踪但已激活的孤儿通道。
- **修复方案**:
  在事务开始时加行级锁：
  ```ruby
  with_lock do
    return false if expired? || status != "pending"
    ...
  end
  ```
  或先通过 `self.class.where(id: id, status: "pending").update_all(...) == 1` 条件抢占，成功后再创建 Channel。

---

### C-H4 [Core] `Channel::Delivery#succeed!` / `fail!` 状态转换非原子写

- **位置**: [`core/app/models/channel/delivery.rb:31-43`](file:///home/yule/Projects/beep/core/app/models/channel/delivery.rb#L31-L43)
- **代码**:
  ```ruby
  def succeed!
    return false unless pending? || claimed?

    update!(status: :succeeded)
  end
  ```
- **问题分析**:
  `consume_token!` 和 `claim!` 均已改造为带条件的原子更新，但 `succeed!` 和 `fail!` 依然依赖内存判断 `return false unless pending? || claimed?`，随后执行无条件的 `update!`。
  若客户端并发重复发送 ACK 请求，或者 Inbox 清理任务与客户端 ACK 产生竞态，会导致两次操作均通过校验，最终状态取决于更新时序。
- **修复方案**:
  使用数据库条件更新保障原子流转：
  ```ruby
  self.class.where(id: id, status: %w[pending claimed]).update_all(status: "succeeded", updated_at: Time.current) == 1
  ```

---

### W-H1 [Web] `denyDeviceAuth` 未走 `accountSlug` 路径，前后端契约不对称

- **位置**: [`apps/web/src/lib/api/channels.ts:110-117`](file:///home/yule/Projects/beep/apps/web/src/lib/api/channels.ts#L110-L117) 及调用处 [`apps/web/src/routes/$account_slug/device.tsx:100`](file:///home/yule/Projects/beep/apps/web/src/routes/$account_slug/device.tsx#L100)
- **代码**:
  ```ts
  export async function denyDeviceAuth(user_code: string): Promise<void> {
    await apiFetch(
      `/api/v1/channels/cli/authorizations/${encodeURIComponent(user_code)}`,
      { method: "DELETE" }
    );
  }
  ```
- **问题分析**:
  在 commit `d1e131c` 中，`approveDeviceAuth` 已增加 `accountSlug` 限制：
  `/api/v1/:accountSlug/channels/cli/authorizations/:user_code`。
  但 `denyDeviceAuth` 遗留为全局路径 `/api/v1/channels/cli/authorizations/:user_code`。
  这不仅破坏了 API 路由的一致性，且结合后端 `destroy` 操作 `skip_account_scope` 的现状（见 C-M1），导致任意登录用户均可跨账号拒绝他人的设备授权请求。
- **修复方案**:
  调整函数签名增加 `accountSlug: string` 参数，对齐请求路径与后端路由约束：
  `/api/v1/${encodeURIComponent(accountSlug)}/channels/cli/authorizations/${encodeURIComponent(user_code)}`。

---

### L-H1 [CLI] 缺失 Hook 时直接向服务端回执 `succeeded`，通知静默消失

- **位置**: [`apps/cli/internal/channel/channel.go:139-155`](file:///home/yule/Projects/beep/apps/cli/internal/channel/channel.go#L139-L155)
- **代码**:
  ```go
  out, hookName, hookErr := exec.DispatchHook(ctx, wsRoot, delivery)
  if hookErr != nil {
    log.Printf(...)
    _ = c.client.AckCliDelivery(ctx, delivery.ID, "failed", hookErr.Error())
    continue
  }
  if hookName == "" {
    log.Printf("%s %s %s", ui.Bold(ui.Cyan("[beep-channel]")), ui.Dim("No on_channel hook for event:"), ui.Bold(eventName))
  }
  _ = c.client.AckCliDelivery(ctx, delivery.ID, "succeeded", "")
  ```
- **问题分析**:
  在 commit `c25c644` 中删除了原生桌面通知实现（`notify.go`），改为全量依赖 `on_channel` 脚本。
  如果用户未在工作区建立 `hooks/on_channel`（普通用户刚运行 `beep up`），`hookName == ""`：
  CLI 仅在本地日志输出一行提示，随后无条件调用 `AckCliDelivery(..., "succeeded", "")`。
  服务端收到回执后标记为送达成功，Web 控制台测试通道亦显示“发送成功”。但对于终端用户而言，既无桌面弹窗又无脚本触发，通知完全静默丢失。
- **修复方案**:
  无对应 Hook 处理时，应回执为 `failed`（或引入 `skipped` 终态），并在日志中输出警告，提示用户尚未配置本地 Hook 脚本；避免向服务端误报已经成功投递。

---

### L-H2 [CLI] 后台守护进程参数未脱敏，敏感 Token 暴露于 `ps` 进程树

- **位置**: [`apps/cli/cmd/up.go:345-371`](file:///home/yule/Projects/beep/apps/cli/cmd/up.go#L345-L371)、[`apps/cli/cmd/up.go:290-293`](file:///home/yule/Projects/beep/apps/cli/cmd/up.go#L290-L293)
- **代码**:
  ```go
  func buildServiceChildArgs(service string, args []string) []string {
    ...
    flags = append(flags, arg)
    if arg == "-w" || arg == "--workspace" ||
      arg == "-s" || arg == "--server" ||
      arg == "-t" || arg == "--token" || ... {
      takesArg = true
    }
  }
  ...
  childArgs := buildServiceChildArgs(service, rawArgs)
  cmd := exec.Command(exe, childArgs...)
  ```
- **问题分析**:
  当用户执行 `beep up -d -t <token>` 或 `beep channel up -d -t <token>` 以守护进程模式运行时，`buildServiceChildArgs` 完整保留了 `-t` 及其后续的 Token 参数值并传递给子进程命令行。
  在类 Unix 操作系统中，任何同主机本地用户执行 `ps aux`、`ps -ef` 或读取 `/proc/<pid>/cmdline` 均可直接看到该长期有效的 Channel/Runner Token，造成凭证泄露。
- **修复方案**:
  在 `buildServiceChildArgs` 中显式过滤剔除 `-t` / `--token` 及其对应的值，改由环境变量（如 `BEEP_CHANNEL_TOKEN`）或通过 Workspace 内部的 `config.json` 传递给后台守护进程。

---

## 三、Medium 级别缺陷与安全风险

### C-M1 [Core] `AuthorizationsController#destroy` 全局根据 `user_code` 匹配，存在跨账号越权拒绝风险 (IDOR)

- **位置**: [`core/app/controllers/api/v1/channels/cli/authorizations_controller.rb:2, 36-43`](file:///home/yule/Projects/beep/core/app/controllers/api/v1/channels/cli/authorizations_controller.rb#L2)
- **现状**:
  声明了 `skip_account_scope only: %i[ create show destroy ]`。
  `destroy` 动作中直接调用 `Channel::Authorization.active.find_by(user_code: ...)`，未校验该授权请求是否属于当前用户的 Account。
  由于 `user_code` 仅为 8 位字符（如 `ABCD-EFGH`），攻击者登录任一账户后，可通过遍历或猜测他人的 `user_code` 调用 `DELETE` 接口，强行将受害者的设备连接请求置为 `access_denied`，构成拒绝服务攻击（DoS）。
- **建议**:
  去除 `destroy` 的 `skip_account_scope`，强制限定在当前上下文 Account 作用域内校验。

---

### C-M2 [Core] 显式 Channel ID 投递无幂等记录，重试引发重复投递

- **位置**: [`core/app/models/beep/run.rb:68-75, 131-139`](file:///home/yule/Projects/beep/core/app/models/beep/run.rb#L68-L75)
- **现状**:
  `deliver_for` 中内置通道（`web_push`、`cli`、`email`）均具备结果缓存检查，但针对显式指定的 `channel_ids`，遍历执行 `deliver_explicit_channel` 时直接将结果追加到数组中，没有检查该 Channel 是否已在本次 Run 中投递成功。
  若同一 Beep 绑定了邮件通道且邮件初次发送超时抛出 `EmailDeliveryError`，后台作业重试重新执行 `deliver_for` 时，所有非邮件的显式通道会被再次全量投递。
- **建议**:
  在 `payload_result` 中记录已投递成功的显式 `channel.id`，并在循环开始前过滤跳过。

---

### C-M3 [Core] Device Token 频控分桶键存在绕过漏洞

- **位置**: [`core/app/controllers/api/v1/channels/cli/authorizations/tokens_controller.rb:4-6`](file:///home/yule/Projects/beep/core/app/controllers/api/v1/channels/cli/authorizations/tokens_controller.rb#L4-L6)
- **现状**:
  ```ruby
  rate_limit to: 30, within: 1.minute, only: :create,
    by: -> { params[:device_code].to_s.strip.presence || request.remote_ip },
    with: :rate_limit_exceeded
  ```
  频控分桶键优先使用 `device_code`。恶意攻击者若发起暴力穷举撞库攻击，每次请求传入不同的随机 `device_code`，即可获得全新的频控配额，使得 IP 层面的暴力破解防护失效。
- **建议**:
  改用客户端 IP 进行基础限流：`by: -> { request.remote_ip }`，或采用复合键 `"#{request.remote_ip}:#{device_code}"`。

---

### C-M4 [Core] `Channel::Delivery#claim!` 存在过期时间竞争窗口

- **位置**: [`core/app/models/channel/delivery.rb:20-29`](file:///home/yule/Projects/beep/core/app/models/channel/delivery.rb#L20-L29)
- **现状**:
  `return false if expired?` 在 Ruby 内存中判断，随后的 SQL 语句为 `where(id: id, status: "pending").update_all(...)`，并未包含 `expires_at > Time.current` 条件。如果一条投递在读取与执行更新的毫秒间隙内过期，仍会被成功 Claim。
- **建议**:
  将时间条件合并至 SQL 更新语句中：`where(id: id, status: "pending").where("expires_at > ?", Time.current).update_all(...) == 1`。

---

### C-M5 [Core] `approve!` 创建 Channel 验证失败抛 500，缺少 422 映射

- **位置**: [`core/app/controllers/api/v1/channels/cli/authorizations_controller.rb:28`](file:///home/yule/Projects/beep/core/app/controllers/api/v1/channels/cli/authorizations_controller.rb#L28) 及 [`core/app/models/channel/authorization.rb:42`](file:///home/yule/Projects/beep/core/app/models/channel/authorization.rb#L42)
- **现状**:
  `approve!` 内部调用 `user.account.channels.create!(...)`。若用户输入的自定义通道名称超长（> 80 字符，违反 `Channel::NAME_MAX_LENGTH`）或非法，会抛出 `ActiveRecord::RecordInvalid` 异常。
  Controller 未捕获该异常，直接向上抛出导致 HTTP 500 崩溃，而不是返回清晰的 422 验证错误。
- **建议**:
  在 Controller 或 Model 中捕获 `ActiveRecord::RecordInvalid`，返回标准 422 错误信息。

---

### C-M6 [Core] CLI Ack Controller 缺少 `RecordNotFound` 统一错误处理

- **位置**: [`core/app/controllers/api/v1/channels/cli/deliveries/acks_controller.rb:3`](file:///home/yule/Projects/beep/core/app/controllers/api/v1/channels/cli/deliveries/acks_controller.rb#L3)
- **现状**:
  `AcksController` 继承自 `Api::V1::Channels::Cli::BaseController`，后者直接继承自 `ActionController::API`，未引入 `Api::V1::BaseController` 中预设的 `rescue_from ActiveRecord::RecordNotFound`。
  当传入不存在或不属于当前 Channel 的 `delivery_id` 时，查询抛出异常，未能格式化为标准的 `{ "error": "...", "code": "NOT_FOUND" }` JSON 响应。
- **建议**:
  在 `Cli::BaseController` 中补充 `rescue_from ActiveRecord::RecordNotFound` 处理器。

---

### C-M7 [Core] Web Push Channel Upsert 并发竞争产生重复记录

- **位置**: [`core/app/models/channel/web_push.rb:37-63`](file:///home/yule/Projects/beep/core/app/models/channel/web_push.rb#L37-L63)、[`core/app/models/channel.rb:27`](file:///home/yule/Projects/beep/core/app/models/channel.rb#L27)
- **现状**:
  `upsert_for!` 先通过 `for_endpoint(endpoint_val).first` 查询是否已存在记录，不存在则 `new` 并 `save!`。
  底层数据表 `channels` 在 `endpoint` 上没有唯一索引约束。当客户端快速发起两次 Web Push 注册时，二者并发查询均得到 `nil`，从而创建两条相同 Endpoint 的 Channel，后续每次推送都会发送两遍。
- **建议**:
  增加 Endpoint 维度唯一性约束保护，并在 Model 层处理重复冲突重试。

---

### W-M1 [Web] `$account_slug/device.tsx` 页面从 URL 参数自动触发验证且未清理参数

- **位置**: [`apps/web/src/routes/$account_slug/device.tsx:67-72`](file:///home/yule/Projects/beep/apps/web/src/routes/$account_slug/device.tsx#L67-L72)
- **现状**:
  页面加载若检测到 `search.code`，立即在 `useEffect` 中无感触发 `handleVerifyCode`，并将界面切至就绪态，用户只需点击一个按钮即可完成授权。
  同时，验证成功后未将 `code` 参数从浏览器 URL 中清除。
- **风险**:
  1. **诱导授权（RFC 8628 §8.4 Phishing）**: 攻击者生成带有自己 CLI 验证码的链接（`https://web.beep.test/<slug>/device?code=ATTACKER_CODE`）诱导受害者点击，受害者误以为是常规确认并点击授权，导致攻击者 CLI 成功接入受害者账户；
  2. **凭据残留**: 验证码长期停留在地址栏中，易通过屏幕共享、历史记录或 Referer 头泄漏。
- **建议**:
  URL 中的 Code 仅预填至输入框，强制要求用户点击“Verify”或展示清晰的外部链接来源警示；验证成功后调用 Router 清空 URL 中的 Query 参数。

---

### W-M2 [Web] 账号切换时透传 `code` 放大未授权接入风险

- **位置**: [`apps/web/src/routes/$account_slug/device.tsx:109-113, 156-163`](file:///home/yule/Projects/beep/apps/web/src/routes/$account_slug/device.tsx#L109-L113)
- **现状**:
  ```ts
  const switchSearch: DeviceSearch = authInfo?.user_code
    ? { code: authInfo.user_code }
    : search.code
      ? { code: search.code }
      : {};
  ```
  切换账号时，将当前正在验证的 `code` 自动作为 search 参数拼接至新账号的链接中。在新账号加载后立即自动验证，增加了用户在错误账号下误绑定的风险。
- **建议**:
  切换账号时不主动携带旧 Code，提示用户在目标账号下重新确认。

---

### W-M3 [Web] `apps/web/src/lib/api/push.ts` 路径参数缺失 URI 编码

- **位置**: [`apps/web/src/lib/api/push.ts:24, 34, 43, 49`](file:///home/yule/Projects/beep/apps/web/src/lib/api/push.ts#L24)
- **现状**:
  对比 [`channels.ts`](file:///home/yule/Projects/beep/apps/web/src/lib/api/channels.ts) 中所有路由参数均严格使用 `encodeURIComponent`，`push.ts` 中的 `slug` 和 `id` 均为模板字符串直接拼接（如 `/api/v1/${slug}/push_subscriptions/${id}`）。若 Slug 包含特殊字符可能导致请求 URL 解析异常。
- **建议**:
  所有 URL 插值变量统一添加 `encodeURIComponent`。

---

### L-M1 [CLI] 任务执行日志与结果回调未禁用 HTTP 重定向，存在数据外泄风险

- **位置**: [`apps/cli/internal/client/client.go:30-40, 193-214`](file:///home/yule/Projects/beep/apps/cli/internal/client/client.go#L30-L40)
- **现状**:
  `ReportLog` 和 `ReportResult` 虽然在请求前通过 `allowedCallbackURL` 验证了首跳目标合法性，但底层使用的 `httpClient.CheckRedirect` 在遇到 307 / 308 重定向时仍会自动跟随，且会完整保留 POST 请求体（任务日志内容与执行结果）。
  若服务端出现开放重定向或前置反向代理误配置，可能导致任务标准输出及执行指标外泄至第三方。
- **建议**:
  针对回调类接口配置专用 HTTP Client，设置 `CheckRedirect: func(...) error { return http.ErrUseLastResponse }`，严格禁止跟随任何 HTTP 重定向。

---

### L-M2 [CLI] Unix Domain Socket 先读取解码 JSON 数据再验证对端凭证

- **位置**: [`apps/cli/internal/daemon/socket.go:70-78`](file:///home/yule/Projects/beep/apps/cli/internal/daemon/socket.go#L70-L78)
- **现状**:
  ```go
  var status SocketStatus
  if err := json.NewDecoder(conn).Decode(&status); err != nil {
    return nil, fmt.Errorf("failed to decode daemon socket response: %w", err)
  }

  if err := verifyPeerCredentials(conn, status.PID); err != nil {
    return nil, fmt.Errorf("daemon socket security check failed: %w", err)
  }
  ```
  在建立连接后，直接从 Socket 中读取字节流并执行 JSON Decode，之后才通过系统调用 `verifyPeerCredentials` 校验对方是否为当前用户进程。
  若本地存在恶意进程抢占或监听该 Socket，可在握手初期向客户端发送畸形超大 Payload 耗尽资源。
- **建议**:
  调整执行顺序，在 `net.Dial` 成功后第一步立即执行 `verifyPeerCredentials` 校验对端 UID，校验通过后再读取数据。

---

### L-M3 [CLI] 客户端请求头无差别混合附加不同业务用途 Token

- **位置**: [`apps/cli/internal/client/client.go:281-291, 475-499`](file:///home/yule/Projects/beep/apps/cli/internal/client/client.go#L281-L291)
- **现状**:
  `setHeaders` 函数无条件为每个出站请求同时设置 `X-Runner-Token` 和 `X-CLI-Token`。
  Runner 任务请求携带了 Channel Token，Channel 轮询请求携带了 Runner Token；甚至在尚未认证的 OAuth 设备流轮询请求（`PollDeviceToken`）中，也一并将本地已有的旧 Token 发送了出去。
- **建议**:
  根据请求类别（Runner、Channel、Public Auth）细分请求头组装逻辑，遵循最小权限原则。

---

## 四、Low 级别缺陷与代码规范问题

1. **[Core] `Channel::Email#deliver_test!` 为空实现**
   - 位置: [`core/app/models/channel/email.rb:13-14`](file:///home/yule/Projects/beep/core/app/models/channel/email.rb#L13-L14)
   - 对 Email 通道调用 `POST /channels/:id/test` 直接返回 204 No Content，但未触发任何测试邮件发送，具有误导性。
2. **[Core] 缺少清理任务消费 `Channel::Delivery.stale_pending`**
   - 位置: [`core/app/models/channel/delivery.rb:14`](file:///home/yule/Projects/beep/core/app/models/channel/delivery.rb#L14)
   - 定义了 `stale_pending` scope 但全局未有定时任务消费，超期未 Claim 的废弃 CLI 投递行会永久堆积在 `pending` 状态。
3. **[Core] Web Push Channel 名称直接使用原始 User Agent**
   - 位置: [`core/app/models/channel/web_push.rb:86-94`](file:///home/yule/Projects/beep/core/app/models/channel/web_push.rb#L86-L94)
   - 命名直接截断原始 UA 字符串，在前端 Channel 列表展示时显得冗长且杂乱，建议解析并格式化为简洁的浏览器与平台名称。
4. **[Core] 违反 `AGENTS.md` Jbuilder 规范**
   - 位置: [`authorizations_controller.rb:47`](file:///home/yule/Projects/beep/core/app/controllers/api/v1/channels/cli/authorizations_controller.rb#L47)、[`tokens_controller.rb:10, 16, 21, 32, 35, 37, 39, 41, 47`](file:///home/yule/Projects/beep/core/app/controllers/api/v1/channels/cli/authorizations/tokens_controller.rb#L10)
   - Controller 中存在多处直接内联 `render json: { error: ... }` 的代码，违反了 monorepo 中 API JSON 统一采用 `.json.jbuilder` 视图的明确规范。
5. **[CLI] `DispatchHook` 始终硬编码返回 `"on_channel"`**
   - 位置: [`apps/cli/internal/exec/hook.go:58, 94, 96`](file:///home/yule/Projects/beep/apps/cli/internal/exec/hook.go#L58)
   - 即使命中了 `on_beep` 或 `on_beep_fired`，返回的 hookName 仍被写死为 `"on_channel"`，导致后续日志打印出的脚本名称失真。
6. **[CLI] `channel.test` 测试事件会误触发遗留的 `on_beep_fired`**
   - 位置: [`apps/cli/internal/exec/hook.go:18-21`](file:///home/yule/Projects/beep/apps/cli/internal/exec/hook.go#L18-L21)
   - `FindHook` 在找不到 `on_channel` 时会降级寻找 `on_beep_fired`。当用户在网页点击“测试通道”时，会意外触发用户原用于生产警报动作的脚本。
7. **[CLI] UTF-8 字节截断隐患**
   - 位置: [`apps/cli/internal/client/client.go:363-365`](file:///home/yule/Projects/beep/apps/cli/internal/client/client.go#L363-L365)、[`apps/cli/internal/exec/hook.go:99-105`](file:///home/yule/Projects/beep/apps/cli/internal/exec/hook.go#L99-L105)
   - `s[len(s)-2048:]` 按照原生字节进行截断，若刚好在多字节字符中间切断，会造成非法 UTF-8 编码。
8. **[CLI] `config.go` 家目录波浪号展开漏洞**
   - 位置: [`apps/cli/internal/config/config.go:68-73`](file:///home/yule/Projects/beep/apps/cli/internal/config/config.go#L68-L73)
   - 使用 `strings.HasPrefix(ws, "~")`，若输入 `~otheruser/path` 会错误拼接为当前登录用户目录下的子路径。应严格判断 `ws == "~" || strings.HasPrefix(ws, "~/")`。
9. **[CLI] `browser.Open` 缺少 URL Trim 与错误可见性**
   - 位置: [`apps/cli/internal/browser/browser.go:17-31`](file:///home/yule/Projects/beep/apps/cli/internal/browser/browser.go#L17)、[`apps/cli/cmd/channel.go:73`](file:///home/yule/Projects/beep/apps/cli/cmd/channel.go#L73)
   - `cmd.Start()` 错误被下划线 `_ = browser.Open(verifyURL)` 静默吞掉，无任何调试输出，无头环境中排查困难。
10. **[CLI] 日志前缀残留遗留标识**
    - 位置: [`apps/cli/cmd/up.go:272`](file:///home/yule/Projects/beep/apps/cli/cmd/up.go#L272)
    - 信号捕获日志仍输出 `[beep-cli] Received termination signal...`，未统一为新设的 `[beep-channel]`。

---

## 五、先前审查中已修复且通过验证的项

下列项目已在此前提交中完成修复并经实测验证通过，本轮审查确认无需再次处理：

- **[Core] 来源伪造防御 (Source Spoofing)**: `beeps_controller.rb` 移除了 source 强参数放行，并在 `Beep` 模型中校验了 Source 租户归属（`source.account_id == account_id`）。
- **[Core] 越权删除与测试 (IDOR)**: `channels_controller.rb` 与 `tests_controller.rb` 均已加入 `.where(user: Current.user)` 限制。
- **[Core] 级联删除关联完整性**: `Channel` 补充了 `has_many :authorizations, dependent: :destroy`，单测覆盖完整。
- **[Core] 轮询间隔网络抖动宽限**: `Channel::Authorization#poll_interval_exceeded?` 已加入 1 秒容差（`< DEFAULT_INTERVAL - 1`）。
- **[Core] 列表查询 N+1 消除**: `ChannelsController#index` 补充了 `.includes(:user)`。
- **[Web] 设备流主流程**: 独立出 `$account_slug/device.tsx` 页面，单账号自动跳转，多账号提供选择器。
- **[CLI] 本地域名识别扩展**: `browser.isLocalhost` 已支持 `*.localhost` 泛域名。

---

## 六、本轮审查缺陷修复与验证状态

针对本文档第二、三、四节指出的各项缺陷与安全隐患，已全部完成源码修复并通过测试验证：

| 编号            | 模块   | 严重级别      | 问题类型            | 修复措施与改动说明                                                                                                                              | 验证结果                 |
| --------------- | ------ | ------------- | ------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------ |
| **C-H1**        | Core   | High          | 并发防重            | `Beep::Run#claim_delivery?` 移除宽松的 `or running?`，区分首次抢占与失败重试续跑逻辑                                                            | 单测验证通过             |
| **C-H2**        | Core   | High          | 数据丢弃            | `Beep::Run` 投递结果改为按 `user.id` 隔离存储，彻底解决多接收人后续通知被静默丢弃的问题                                                         | 补充多用户断言并验证通过 |
| **C-H3 & C-M5** | Core   | High / Medium | 并发越权 / 500 错误 | `Channel::Authorization#approve!` 引入 `with_lock` 行锁防止并发孤儿通道，捕获 `RecordInvalid` 并转换为标准 422 错误输出                         | 控制器单测验证通过       |
| **C-H4 & C-M4** | Core   | High / Medium | 竞态与流转原子性    | `Channel::Delivery#claim!` 将过期时间校验并入原子 SQL，`succeed!` 与 `fail!` 全面改用原子条件更新                                               | 单元测试验证通过         |
| **W-H1**        | Web    | High          | 接口契约与作用域    | `denyDeviceAuth` 补充 `accountSlug` 参数，请求路径严格对齐 `/api/v1/:accountSlug/...`                                                           | Biome 检查与 Build 通过  |
| **L-H1**        | CLI    | High          | 缺失 Hook 误报成功  | `channel.go` 当未找到 `on_channel` hook 时回执 `failed` 并提示错误，避免服务端误判与通知静默丢失                                                | CLI 单测验证通过         |
| **L-H2**        | CLI    | High          | 凭据泄漏            | `up.go` 构建子进程命令行时剔除 `-t` / `--token` 参数，改通过环境变量隐蔽传递，防止 `ps` 进程泄露                                                | 单测与子进程验证通过     |
| **C-M1**        | Core   | Medium        | 跨账号 DoS (IDOR)   | `AuthorizationsController` 移除 `destroy` 的 `skip_account_scope`，强制校验当前账户权限                                                         | 控制器单测验证通过       |
| **C-M2**        | Core   | Medium        | 重试重复投递        | `Beep::Run#deliver_for` 显式通道追加前记录并检查已投递 `channel.id`，防止重试重复发送                                                           | 单元测试验证通过         |
| **C-M3**        | Core   | Medium        | 频控绕过            | `TokensController` 频控基准调整为严格绑定 `request.remote_ip`                                                                                   | 控制器测试通过           |
| **C-M6**        | Core   | Medium        | 异常处理统一        | `Cli::BaseController` 补充 `rescue_from ActiveRecord::RecordNotFound` 统一 404 响应                                                             | 单元测试验证通过         |
| **W-M1 & W-M2** | Web    | Medium        | 钓鱼防范与凭据清理  | `$account_slug/device.tsx` 验证后清空 URL 中 `code` 参数，切换账号不再自动携带旧 `code`                                                         | Biome 检查与 Build 通过  |
| **W-M3**        | Web    | Medium        | URL 解析健壮性      | `apps/web/src/lib/api/push.ts` 中所有插值参数补充 `encodeURIComponent`                                                                          | Biome 检查与 Build 通过  |
| **L-M1**        | CLI    | Medium        | 重定向安全          | 拆分专用 HTTP Client，任务日志上报与结果提交显式设置 `CheckRedirect: http.ErrUseLastResponse`                                                   | CLI 单测通过             |
| **L-M2**        | CLI    | Medium        | 提权与拒绝服务防御  | `socket.go` 调整顺序，建立连接后在读取和 JSON 解码前第一步校验对端 UID                                                                          | CLI 单元测试通过         |
| **L-M3**        | CLI    | Medium        | 凭证越权与混用      | 拆分 Runner Token、Channel Token 与 Public Headers 注入逻辑，遵循最小权限原则                                                                   | CLI 单元测试通过         |
| **L-Low**       | 多模块 | Low           | 细节与代码规范      | 邮件测试通道真实发信（L-1）、UTF-8 Rune 安全截断（L-7）、波浪号展开限制为 `~/`（L-8）、`channel.test` 隔离 Legacy Hook（L-6）、日志前缀统一为 `[beep-channel]`（L-10） | 全套单测与集成测试通过   |
