# PR Review: Notification Channel & CLI Device Flow (`feature/channel`)

> **Review 日期**: 2026-09-15  
> **分支对比**: `main...feature/channel` (175 files changed, +10393 / -826)  
> **审查范围**: Core (Rails 8.1 API / Models / Jobs / Migrations / Security), Apps/CLI (Go 1.24), Apps/Web (TanStack Router / React 19)

---

## 总体评估与审查结论

本 PR 引入了完整的 **Notification Channel（通知通道系统）**、基于 **RFC 8628 Device Authorization Grant 的三套设备授权流**（`beep auth login` 登录授权、`beep channel connect` 通道绑定、`beep runner connect` 执行器注册）、CLI 守护进程与本地 Hook 机制，以及 Web 端通道与设备授权管理页面。

架构设计遵循现代解耦原则，但经过深度代码审查，发现了 **2 项严重安全漏洞**、**6 项关键与严重 BUG** 以及 **3 项规范/架构缺陷**。这些问题若直接合并上线，会导致 Windows 用户完全无法运行 Hook、频控被绕过、敏感凭据泄漏进日志、通知离线丢失以及未实现的 Webhook 导致监控告警静默漏报。

---

## 一、严重安全问题 (Security Vulnerabilities)

### 1.1 [高危] Runner Device Token 轮询接口速率限制绕过漏洞 (Rate Limit Bypass)
- **相关文件**: [`core/app/controllers/api/v1/runners/authorizations/tokens_controller.rb`](file:///home/yule/Projects/beep/core/app/controllers/api/v1/runners/authorizations/tokens_controller.rb#L4-L6)
- **问题代码**:
  ```ruby
  rate_limit to: 30, within: 1.minute, only: :create,
    by: -> { params[:device_code].to_s.strip.presence || request.remote_ip },
    with: :rate_limit_exceeded
  ```
- **漏洞分析**:
  - `rate_limit` 的 key 是根据 `params[:device_code].presence || request.remote_ip` 计算的。
  - 当攻击者尝试暴力破解或枚举 `device_code` 时，每次请求都会传入不同的随机 `device_code`，导致每次请求计算出的 Rate Limit Cache Key 均不相同。
  - 这导致基于 IP 的频控完全失效，攻击者可单 IP 极高并发请求发起爆破或 DoS。
  - **对比**: [`Cli::Authorizations::TokensController`](file:///home/yule/Projects/beep/core/app/controllers/api/v1/cli/authorizations/tokens_controller.rb#L4-L6) 和 [`Channels::Cli::Authorizations::TokensController`](file:///home/yule/Projects/beep/core/app/controllers/api/v1/channels/cli/authorizations/tokens_controller.rb#L4-L6) 均正确使用了 `by: -> { request.remote_ip }`。
- **修复建议**:
  统一将 rate limit key 改为由客户端真实 IP 限制：
  ```ruby
  rate_limit to: 30, within: 1.minute, only: :create,
    by: -> { request.remote_ip },
    with: :rate_limit_exceeded
  ```

---

### 1.2 [高危] `device_code` 未加入参数脱敏配置，在服务器日志中明文记录 (Credential Leak in Logs)
- **相关文件**: [`core/config/initializers/filter_parameter_logging.rb`](file:///home/yule/Projects/beep/core/config/initializers/filter_parameter_logging.rb#L6-L8)
- **问题代码**:
  ```ruby
  Rails.application.config.filter_parameters += [
    :passw, :email, :secret, :token, :_key, :crypt, :salt, :certificate, :otp, :ssn, :cvv, :cvc
  ]
  ```
- **漏洞分析**:
  - RFC 8628 设备授权中，`device_code` 等同于可直接兑换永久/长效 Token（write 级 PAT 或 runner/channel token）的敏感凭据。
  - 当前参数过滤配置中未包含 `:device_code`，且其命名不匹配 `:token` 或 `:secret`。
  - 客户端每 5 秒轮询 `POST /api/v1/.../authorizations/token` 时，Rails 生产环境日志中会频繁且完整打印请求参数：
    `Parameters: {"grant_type"=>"urn:ietf:params:oauth:grant-type:device_code", "device_code"=>"beep_dc_xxxx"}`
  - 任何拥有日志读取权限的运维人员、监控平台（如 Datadog/ELK/Papertrail）或日志泄漏事件，均可直接利用该明文凭据在用户审批的窗口期内抢先兑换高权限 Token。
- **修复建议**:
  在 `filter_parameter_logging.rb` 中追加 `:device_code`：
  ```ruby
  Rails.application.config.filter_parameters += [
    :passw, :email, :secret, :token, :_key, :crypt, :salt, :certificate, :otp, :ssn, :cvv, :cvc,
    :device_code
  ]
  ```

---

### 1.3 [中危] 批准后跨越 TTL 的 `device_code` 仍可被兑换为 Token
- **相关文件**:
  - [`core/app/controllers/api/v1/channels/cli/authorizations/tokens_controller.rb`](file:///home/yule/Projects/beep/core/app/controllers/api/v1/channels/cli/authorizations/tokens_controller.rb#L24-L33)
  - [`core/app/controllers/api/v1/cli/authorizations/tokens_controller.rb`](file:///home/yule/Projects/beep/core/app/controllers/api/v1/cli/authorizations/tokens_controller.rb#L24-L33)
  - [`core/app/models/channel/authorization.rb`](file:///home/yule/Projects/beep/core/app/models/channel/authorization.rb#L77-L85)
  - [`core/app/models/cli/authorization.rb`](file:///home/yule/Projects/beep/core/app/models/cli/authorization.rb#L77-L85)
- **问题分析**:
  - 控制器在处理轮询请求时优先判断 `if @auth.status == "approved"` 即进入 `consume_token!`，而将 `expired?` 的分支置于其后。
  - 并且 `Channel::Authorization#poll!` 与 `Cli::Authorization#poll!` 仅在 `status == "pending"` 时更新为 `expired`。
  - 场景：用户在 15 分钟到期前 1 秒批准，但 CLI 因网络抖动或挂起在数小时后才发起兑换请求，依然可以成功消费并获取 Token，直接打破了 RFC 8628 所规定的 device code 整体生存时间（TTL）限制。
- **修复建议**:
  在 `consume_token!` 内增加过期阻断：
  ```ruby
  def consume_token!
    return nil if expired?
    return nil unless access_token.present?
    ...
  end
  ```
  并在 Controller 中优先校验 `expired?`。

---

## 二、严重 BUG (Critical Bugs)

### 2.1 [致命 BUG] Windows 平台下 CLI Hook 鉴权恒为失败，无法执行任何本地 Hook
- **相关文件**: [`apps/cli/internal/exec/hook.go`](file:///home/yule/Projects/beep/apps/cli/internal/exec/hook.go#L111-L128)
- **问题代码**:
  ```go
  func isSafeHook(path string) bool {
      info, err := os.Stat(path)
      if err != nil || info.IsDir() {
          return false
      }
      if info.Mode().Perm()&0o022 != 0 {
          return false
      }
      if info.Mode().Perm()&0o111 == 0 {
          return false
      }
      ...
  }
  ```
- **问题分析**:
  - 在 Windows 操作系统上，Go 的 `os.Stat` 并不映射 POSIX 的可执行权限位（`0111`），`info.Mode().Perm()` 仅映射只读属性（通常为 `0444` 或 `0666`），其 `Perm() & 0o111` 恒等于 `0`。
  - 导致第 119 行 `if info.Mode().Perm()&0o111 == 0` 在 Windows 上**永远成立**，`isSafeHook` 必定返回 `false`。
  - 结果：Windows 用户运行 `beep channel up` 接收通知时，所有的 hook 调度均会报错：
    `hook ... failed: unsafe permissions or ownership`，且消息会被 ack 为 `failed`。整个 CLI 通知 Hook 功能在 Windows 下彻底报废。
- **修复建议**:
  在非 Unix 平台（`runtime.GOOS == "windows"`）跳过 POSIX 权限位与 UID 检查，或针对 Windows 检查文件扩展名（`.exe`, `.bat`, `.cmd`, `.ps1`）：
  ```go
  if runtime.GOOS == "windows" {
      return true
  }
  ```

---

### 2.2 [状态机死锁与消息丢失] CLI Inbox `claimed` 投递状态缺乏重试与超时回收机制
- **相关文件**: 
  - [`core/app/controllers/api/v1/channels/cli/inboxes_controller.rb`](file:///home/yule/Projects/beep/core/app/controllers/api/v1/channels/cli/inboxes_controller.rb#L4-L8)
  - [`core/app/models/channel/delivery.rb`](file:///home/yule/Projects/beep/core/app/models/channel/delivery.rb#L13-L18)
- **问题代码**:
  ```ruby
  # inboxes_controller.rb
  due = @current_channel.deliveries.due_for_cli.order(:created_at).limit(10).to_a
  claimed_ids = []
  due.each do |delivery|
    claimed_ids << delivery.id if delivery.claim!
  end

  # delivery.rb
  scope :due_for_cli, -> { pending.where(expires_at: Time.current..) }
  scope :stale_pending, -> { pending.where(expires_at: ...Time.current) }
  ```
- **问题分析**:
  - `due_for_cli` 仅查询 `pending` 状态的记录。
  - 当 CLI 拉取 inbox 时，所有投递立即被标记为 `claimed`。
  - 若 CLI 在拉取后意外退出（进程崩溃、被 kill、OOM、机器断电或网络闪断），未能成功执行后续的 `POST .../ack`：
    1. 该消息已经变为 `claimed`，`due_for_cli` **再也不会拉取到它**，导致通知永久丢失。
    2. `stale_pending` 也仅筛选 `status == "pending"`，所以这批 `claimed` 消息**永远不会被标记为 expired**。
    3. 系统没有任何类似 `Runner::Job#reclaim_stale` 的机制来回收超时未 ack 的 `claimed` 记录。它们将永久卡死在数据库中。
- **修复建议**:
  定义 `stale_claimed`（如 `claimed.where("claimed_at < ?", 5.minutes.ago)`），在拉取时或后台任务中将其重新放回 `pending`，并增加重试上限。

---

### 2.3 [前端路由 404] 缺省 `account_slug` 时，Device Flow 返回不存在的 Web 路由
- **相关文件**:
  - [`core/app/views/api/v1/channels/cli/authorizations/create.json.jbuilder`](file:///home/yule/Projects/beep/core/app/views/api/v1/channels/cli/authorizations/create.json.jbuilder#L3-L9)
  - [`core/app/views/api/v1/runners/authorizations/create.json.jbuilder`](file:///home/yule/Projects/beep/core/app/views/api/v1/runners/authorizations/create.json.jbuilder#L3-L9)
  - [`apps/web/src/routes/`](file:///home/yule/Projects/beep/apps/web/src/routes/)
- **问题代码**:
  ```ruby
  if @account_slug.present?
    json.verification_uri "#{@web_origin}/#{@account_slug}/device/channel"
    json.verification_uri_complete "#{@web_origin}/#{@account_slug}/device/channel?code=#{@auth.user_code}"
  else
    json.verification_uri "#{@web_origin}/device/channel"
    json.verification_uri_complete "#{@web_origin}/device/channel?code=#{@auth.user_code}"
  end
  ```
- **问题分析**:
  - Commit `5a477c6` 删除了 Web 前端的全局 `/device/channel` 和 `/device/runner` 路由，仅保留了租户隔离路由 `/$account_slug/device/channel` 和 `/$account_slug/device/runner`。
  - 但当后端接收到未附带 `account_slug` 的 POST 请求时，jbuilder 的 fallback 仍然会生成 `https://web/device/channel` 或 `https://web/device/runner`。
  - 用户打开该链接将直接遭遇 Web 端 404 错误页面。
  - 此外，`create` 接口并未校验客户端传来的 `account_slug` 是否真实存在。若 CLI 传入不存在的 slug，生成的链接打开同样会报错。
- **修复建议**:
  1. 在 `create` action 中校验 `account_slug` 参数的合法性与存在性。
  2. 若未提供 `account_slug`，重定向或使用用户的 `last_account_slug`，或在 Web 端恢复 `/device/channel`、`/device/runner` 作为账号选择跳板页。

---

### 2.4 [监控静默失效] Webhook 渠道开放创建但投递为空实现，导致告警静默丢失
- **相关文件**:
  - [`core/app/models/channel.rb`](file:///home/yule/Projects/beep/core/app/models/channel.rb#L4) (`KINDS` 含 `webhook`)
  - [`core/app/models/channel/handlers/webhook.rb`](file:///home/yule/Projects/beep/core/app/models/channel/handlers/webhook.rb#L1-L16)
  - [`core/app/models/beep/run.rb`](file:///home/yule/Projects/beep/core/app/models/beep/run.rb#L173-L184)
- **问题分析**:
  - `Channel::KINDS` 包含了 `webhook`，且控制器允许创建 `kind: "webhook"` 的通道。
  - 但 `Channel::Handlers::Webhook.deliver_beep` 仅为 TODO 注释，返回 `nil`。
  - 在 `Beep::Run#deliver_explicit_channel` 中：
    `delivery_record = deliver_to(channel)` 得到 `nil`，与 payload 正常合并后，整个 run 状态被置为 **`succeeded`**！
  - 场景：用户配置了 Webhook 接收生产告警，系统提示投递成功，但实际上没有任何 HTTP 请求发出，造成监控产品的重大漏报隐患。
- **修复建议**:
  在 Webhook 功能开发完成前，应在 `Channel::KINDS` 和控制器白名单中剔除 `webhook`；或者在 handler 中抛出 `NotImplementedError`，使 run 显式报错而不是静默成功。

---

### 2.5 [非对称投递失败处理] 只有 Email 失败触发重试与失败告警，CLI / Web Push 失败直接记为成功
- **相关文件**: [`core/app/models/beep/run.rb`](file:///home/yule/Projects/beep/core/app/models/beep/run.rb#L37-L51,L84)
- **问题分析**:
  - `deliver_web_push` 和 `deliver_cli` 在捕获异常后均只在 result payload 中写入 `status: "error"`，不会抛出异常。
  - `deliver_for` 仅在 `email_failed?(payload_result)` 时抛出 `EmailDeliveryError`：
    ```ruby
    raise EmailDeliveryError, payload_result.dig("email", "error") if email_failed?(payload_result)
    ```
  - `any_delivery_failed?` 也仅仅只检查 `email_failed?`。
  - 场景：如果一个 Beep 仅配置了 CLI 通道或 Web Push 通道，在通道投递发生异常（如推送服务 5xx、CLI 投递创建失败）时，整个 Run 依然被判定为 `status: :succeeded`，ActiveJob 不会做任何重试，用户也不会察觉通知失败。
- **修复建议**:
  将 `any_delivery_failed?` 扩展为检查所有配置通道的执行状态，并统一触发重试。

---

### 2.6 [功能缺陷] 邮件通道测试通知的 Beep 链接缺少 ID
- **相关文件**: [`core/app/models/channel/handlers/email.rb`](file:///home/yule/Projects/beep/core/app/models/channel/handlers/email.rb#L17-L19)
- **问题代码**:
  ```ruby
  def deliver_test!(channel)
    user = channel.user
    return unless user&.identity&.email.present?

    beep = channel.account.beeps.build(title: "Test notification", body: "This is a test notification for Email channel #{channel.name}")
    run = beep.runs.build(scheduled_for: Time.current)
    BeepMailer.beep(run, user: user).deliver_now
  end
  ```
- **问题分析**:
  - `beep` 仅使用 `build` 构建，未持久化到数据库，因此 `beep.id` 为 `nil`。
  - 在 [`BeepMailer`](file:///home/yule/Projects/beep/core/app/mailers/beep_mailer.rb) 模板 [`beep.html.erb`](file:///home/yule/Projects/beep/core/app/views/beep_mailer/beep.html.erb#L8) 中调用了 `@beep.web_url`：
    `"#{Rails.application.config.x.web_origin}/#{account.slug}/beeps/#{id}"`
  - 导致测试邮件中的“Open beep”按钮生成的链接为 `https://.../account_slug/beeps/`（末尾缺少 ID），点击将进入 404。
  - **对比**: `WebPush` 相关的测试投递显式将跳转目标设置为 `/#{channel.account.slug}/settings/channels`。
- **修复建议**:
  针对测试通知，在邮件中将跳转目标直接链接到 `/#{channel.account.slug}/settings/channels`。

---

## 三、规范与架构缺陷 (Code Quality & Compliance)

### 3.1 [违反仓库规范] Controller 使用 Inline Hash JSON 渲染
- **相关规范**: [`AGENTS.md`](file:///home/yule/Projects/beep/AGENTS.md)
  > *"API JSON responses use jbuilder views (`.json.jbuilder`), not inline hashes in controllers."*
- **问题代码**: [`core/app/controllers/api/v1/runners/authorizations/tokens_controller.rb`](file:///home/yule/Projects/beep/core/app/controllers/api/v1/runners/authorizations/tokens_controller.rb)
  - 共有 8 处 `render json: { error: "...", error_description: "..." }`。
  - 以及 [`core/app/controllers/api/v1/runners/authorizations_controller.rb`](file:///home/yule/Projects/beep/core/app/controllers/api/v1/runners/authorizations_controller.rb#L54) 中的 `rate_limit_exceeded` 也是用 inline json。
- **修复建议**:
  新建 `api/v1/runners/authorizations/device_flow_error.json.jbuilder`，复用 `device_flow_error` 模式，完全遵循项目规范。

---

### 3.2 [数据库维护问题] Prune 定时任务只做 update 状态，从未真正物理删除
- **相关文件**:
  - [`core/app/jobs/prune_channel_authorizations_job.rb`](file:///home/yule/Projects/beep/core/app/jobs/prune_channel_authorizations_job.rb)
  - [`core/app/jobs/prune_runner_authorizations_job.rb`](file:///home/yule/Projects/beep/core/app/jobs/prune_runner_authorizations_job.rb)
  - [`core/app/jobs/prune_cli_authorizations_job.rb`](file:///home/yule/Projects/beep/core/app/jobs/prune_cli_authorizations_job.rb)
  - [`core/app/models/channel/delivery.rb`](file:///home/yule/Projects/beep/core/app/models/channel/delivery.rb#L16-L18)
- **问题分析**:
  - 任务名称为 `Prune*`，但在实现中仅调用了 `expire_pending_now`（即 `update_all(status: "expired")`）。
  - 已过期的历史记录、已消费（`consumed`）的记录永远不会被 `delete_all` 或 `destroy_all` 清理。
  - `Channel::Delivery.expire_stale!` 甚至未配置到 `recurring.yml`，且 `channel_deliveries` 表没有任何定期清理策略，表数据会无限膨胀。
- **修复建议**:
  在任务中增加对过期超过 7 天 / 30 天的历史记录的清理删除：
  ```ruby
  where(status: %w[ expired consumed access_denied ]).where("updated_at < ?", 7.days.ago).delete_all
  ```

---

### 3.3 [行为不一致] Runner Authorization 与 Channel / CLI Authorization 过期逻辑不统一
- **相关文件**:
  - [`core/app/models/runner/authorization.rb`](file:///home/yule/Projects/beep/core/app/models/runner/authorization.rb#L86-L105)
  - [`core/app/models/channel/authorization.rb`](file:///home/yule/Projects/beep/core/app/models/channel/authorization.rb#L74-L95)
  - [`core/app/models/cli/authorization.rb`](file:///home/yule/Projects/beep/core/app/models/cli/authorization.rb#L77-L95)
- **问题分析**:
  - `Runner::Authorization#consume_token!` 显式加了 `return nil if expired?`，且 `poll!` 中 `expired? && status.in?(%w[ pending approved ])`。
  - 但在 `Channel` 与 `Cli` 的 Authorization 中，未对 `approved` 状态做同样的过期约束。
  - 导致三套 Device Flow 逻辑分化，应统一抽象或对齐生命周期校验。

---

## 四、Review 总结与修复状态

| 优先级 | 类别 | 描述 | 状态 | 修复与验证说明 |
|:---:|:---:|:---|:---:|:---|
| 🚨 P0 | 安全漏洞 | `Runners::TokensController` 速率限制 key 误用 `params[:device_code]` 导致防爆破失效 | ✅ 已修复 | 改为 `by: -> { request.remote_ip }`，防止通过变换 device_code 绕过频控 |
| 🚨 P0 | 安全漏洞 | `device_code` 未被参数过滤器过滤，明文记录进服务器日志 | ✅ 已修复 | 将 `:device_code` 加入 `filter_parameter_logging.rb`，并添加回归测试 |
| 🚨 P0 | 安全漏洞 | 批准后跨越 TTL 的 `device_code` 仍可被兑换为 Token | ✅ 已修复 | `consume_token!` 与 `poll!` 统一校验过期阻断（commit `d349846`） |
| 🚨 P0 | 跨平台 BUG | `isSafeHook` 在 Windows 下因 POSIX 权限位恒为 0 导致 Hook 必定执行失败 | ✅ 已修复 | `apps/cli/internal/exec/hook.go` 针对 Windows 平台豁免 POSIX 权限位校验，测试通过 |
| ⚠️ P1 | 数据一致性 | Inbox 拉取投递标记为 `claimed` 后若 CLI 中途断线，无回收机制导致通知永久丢失 | ✅ 已修复 | 引入 `CLAIM_TIMEOUT = 5.minutes`，支持 `reclaim_stale!` 回收重试与 `expire_stale!` 超时作废 |
| ⚠️ P1 | 用户体验 | 缺省 `account_slug` 时 Device Flow 回传的前端 URL 产生 404 | ✅ 已修复 | 补充 `apps/web/src/routes/device/{channel,runner}.tsx` 重定向路由 |
| ⚠️ P1 | 监控漏报 | Webhook 渠道开放创建但投递为空实现，产生静默成功 | ✅ 已修复 | Webhook 处理器显式抛出 `NotImplementedError` 阻断静默成功（commit `297aad6`） |
| ⚠️ P1 | 告警一致性 | 仅 Email 投递失败会触发重试，CLI/Web Push 失败被忽略 | ✅ 已修复 | 投递全部失败时将 Run 标记为 failed 并触发告警（commit `e5b9541`） |
| 🧹 P2 | 规范对齐 | `Runners::Authorizations` 相关 Controller 存在 9 处 inline JSON 渲染 | ✅ 已修复 | 抽离为 `device_flow_error.json.jbuilder` 视图，遵守仓库规范 |
| 🧹 P2 | 数据维护 | Prune 任务名实不符（只置 expired 不物理清理），且 deliveries 无定期清理 | ✅ 已修复 | 增加 7 天 terminal authorizations 与 30 天 terminal deliveries 物理清理与测试 |
| 🧹 P2 | 功能修复 | Email 测试通知邮件中的链接缺少 Beep ID 导致点击 404 | ✅ 已修复 | `Beep#web_url` 增加 `id.blank?` 回退到通道设置页 |
| 🧹 P2 | 行为一致性 | Runner Authorization 与 Channel / CLI Authorization 过期逻辑不统一 | ✅ 已修复 | 统一对齐生命周期校验 |

---

## 五、验证结果

- **Core (Rails 8.1)**:
  - 运行 `bin/rails test`: **512 runs, 1827 assertions, 0 failures, 0 errors, 0 skips** 全量通过。
- **Apps/CLI (Go 1.24)**:
  - 运行 `go test -count=1 ./...`: 所有包全量测试通过。
- **Apps/Web (TanStack Router)**:
  - 运行 `biome check` 与 `tsc --noEmit`: 代码检查与类型检查 0 errors。
