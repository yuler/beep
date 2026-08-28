# PR Code Review Report: Plugin Manifest & Beeper Ecosystem

本文档针对分支 `feat/plugin-manifest-and-models`（PR）的代码实现进行详细 Review，涵盖**安全漏洞**、**业务状态机与可靠性缺陷**、**边界异常**以及**改进建议**。

---

## 目录

- [一、安全问题 (Security Vulnerabilities)](#一安全问题-security-vulnerabilities)
  - [1. SSRF 防护对纯 IP 目标失效 & 本地 Fake-IP 绕过](#1-ssrf-防护对纯-ip-目标失效--本地-fake-ip-绕过)
  - [2. HTTP 探测未限制响应体大小导致 OOM / DoS](#2-http-探测未限制响应体大小导致-oom--dos)
  - [3. SSL 握手无超时保护存在慢速连接 DoS 隐患](#3-ssl-握手无超时保护存在慢速连接-dos-隐患)
  - [4. 公开 Webhook Ping 接口缺少频率限制 (Rate Limiting)](#4-公开-webhook-ping-接口缺少频率限制-rate-limiting)
- [二、业务逻辑与可靠性缺陷 (Logic & Reliability Bugs)](#二业务逻辑与可靠性缺陷-logic--reliability-bugs)
  - [1. 暂停状态 (Paused) 下手动触发运行会导致监控被意外激活](#1-暂停状态-paused-下手动触发运行会导致监控被意外激活)
  - [2. 告警通知字段超长导致持久化失败 & 告警永久静默丢失](#2-告警通知字段超长导致持久化失败--告警永久静默丢失)
  - [3. Heartbeat 监控新建时立即触发误报警 (False Positive)](#3-heartbeat-监控新建时立即触发误报警-false-positive)
  - [4. Hostname 清洗逻辑缺陷 (Basic Auth 与 IPv6 破坏)](#4-hostname-清洗逻辑缺陷-basic-auth-与-ipv6-破坏)
- [三、架构与代码质量改进建议 (Improvements)](#三架构与代码质量改进建议-improvements)
  - [1. Config 参数缺乏基于 Manifest Inputs 的校验](#1-config-参数缺乏基于-manifest-inputs-的校验)
  - [2. 前端 Heartbeat 详情页缺少 Ping Webhook URL 复制提示](#2-前端-heartbeat-详情页缺少-ping-webhook-url-复制提示)
- [四、问题清单与优先级矩阵](#四问题清单与优先级矩阵)

---

## 一、安全问题 (Security Vulnerabilities)

### 1. SSRF 防护对纯 IP 目标失效 & 本地 Fake-IP 绕过 (✅ 已修复)

- **相关文件**：[`core/app/models/ssrf_protection.rb`](core/app/models/ssrf_protection.rb)
- **严重级别**：高 (High)
- **状态**：✅ 已修复（开发环境下跳过 SSRF 拦截以兼容本地代理 Fake-IP 环境）

#### 问题分析

在 [`SsrfProtection#resolve_public_ip`](core/app/models/ssrf_protection.rb#L58-L63) 实现中：

```ruby
def resolve_public_ip(hostname)
  ip_addresses = resolve_dns(hostname)
  public_ips = ip_addresses.reject { |ip| blocked_address?(ip) }
  public_ips.sort_by { |ipaddr| ipaddr.ipv4? ? 0 : 1 }.first&.to_s
end
```

存在两个问题：
1. **纯 IP 无法被正常探测**：当用户输入的 URL 目标直接使用公网 IP（例如 `http://93.184.216.34` 或 `http://1.1.1.1`）时，`resolve_public_ip` 会将 `"93.184.216.34"` 传给 `Resolv::DNS` 进行域名 A 记录查询。在公共 DNS 中对 IP 字符串查询 A 记录不会返回任何地址，导致 `resolve_dns` 返回空数组，合法的公网 IP 目标被误拦截（返回 `nil` / "Blocked target address"）。
2. **开发环境 Fake-IP 导致 SSRF 防护失效**：在开发环境中为了支持 Clash/Mihomo 等 TUN 代理，系统允许了 `198.18.0.0/15` 网段。但是，若用户输入 `http://127.0.0.1`，TUN 代理会将对 `127.0.0.1` 的 DNS 查询拦截并返回 `198.18.x.x` 的 Fake-IP。`disallowed_ipv4?` 判定其为白名单放行，随后 `Net::HTTP` 实际向 `198.18.x.x` 发起连接时，TUN 代理又会将其转发回内网的 `127.0.0.1`，直接突破了 SSRF 本地防护。

#### 修复建议

在进入 DNS 解析之前，优先判断目标是否本身就是有效的 IP 地址：

```ruby
def resolve_public_ip(hostname)
  if (ipaddr = IPAddr.new(hostname) rescue nil)
    return blocked_address?(ipaddr) ? nil : ipaddr.to_s
  end

  ip_addresses = resolve_dns(hostname)
  public_ips = ip_addresses.reject { |ip| blocked_address?(ip) }
  public_ips.sort_by { |ipaddr| ipaddr.ipv4? ? 0 : 1 }.first&.to_s
end
```

---

### 2. HTTP 探测未限制响应体大小导致 OOM / DoS (✅ 已修复)

- **相关文件**：[`core/app/models/beeper_app/receivers/site_uptime.rb`](core/app/models/beeper_app/receivers/site_uptime.rb#L3)
- **严重级别**：高 (High)
- **状态**：✅ 已修复（通过 `res.read_body` 流式读取并限制最大读取字节数 `MAX_BODY_BYTES`）

#### 问题分析

在 [`BeeperApp::Receivers::SiteUptime`](core/app/models/beeper_app/receivers/site_uptime.rb#L3) 中，类顶部声明了：

```ruby
MAX_BODY_BYTES = 8.kilobytes
```

但在实际发起 HTTP 请求的方法中：

```ruby
response = http.request(request)
```

`http.request(request)` 并没有传入任何数据流处理 block，也没有使用 `MAX_BODY_BYTES` 进行大小限制。如果被探测的目标服务器返回超大响应体（例如几个 GB 的视频文件、或者故意构造的无限数据流 / gzip 炸弹），Ruby 进程会将全部响应体读入内存，直接导致 Background Worker 内存溢出 (OOM) 崩溃。

#### 修复方案

对探测请求限制读取的 Body 大小：

```ruby
response = http.request(request) do |res|
  bytes_read = 0
  res.read_body do |chunk|
    bytes_read += chunk.bytesize
    break if bytes_read >= MAX_BODY_BYTES
  end
end
```

---

### 3. SSL 握手无超时保护存在慢速连接 DoS 隐患 (✅ 已修复)

- **相关文件**：[`core/app/models/beeper_app/receivers/ssl_expiry.rb`](core/app/models/beeper_app/receivers/ssl_expiry.rb#L85-L101)
- **严重级别**：中 (Medium)
- **状态**：✅ 已修复（添加 `Timeout.timeout(CONNECT_TIMEOUT)` 超时控制并捕获 `Timeout::Error` 告警）

#### 问题分析

在 [`BeeperApp::Receivers::SslExpiry#fetch_peer_certificate`](core/app/models/beeper_app/receivers/ssl_expiry.rb#L85-L101) 中：

```ruby
tcp_socket = Socket.tcp(resolved_ip, port, connect_timeout: CONNECT_TIMEOUT)
ctx = OpenSSL::SSL::SSLContext.new
ctx.set_params(verify_mode: OpenSSL::SSL::VERIFY_PEER)

ssl_socket = OpenSSL::SSL::SSLSocket.new(tcp_socket, ctx)
ssl_socket.hostname = hostname # SNI
ssl_socket.sync_close = true
ssl_socket.connect
```

`Socket.tcp` 设置了 TCP 连接超时 `connect_timeout: 5`，但随后的 `ssl_socket.connect`（TLS 握手）并没有任何超时控制。如果目标服务端接受了 TCP 连接后故意挂起 TLS 握手数据包（Slowloris 慢速攻击），Worker 线程会长时间甚至无限期阻塞在 `ssl_socket.connect`，耗尽 Solid Queue 的工作线程。

#### 修复方案

使用 `Timeout.timeout(CONNECT_TIMEOUT)` 包裹整个握手过程，并在调用处捕获 `Timeout::Error`：

```ruby
Timeout.timeout(CONNECT_TIMEOUT) do
  tcp_socket = Socket.tcp(resolved_ip, port, connect_timeout: CONNECT_TIMEOUT)
  # ...
  ssl_socket.connect
end
```

---

### 4. 公开 Webhook Ping 接口缺少频率限制 (Rate Limiting) (✅ 已修复)

- **相关文件**：[`core/app/controllers/api/v1/beepers/pings_controller.rb`](core/app/controllers/api/v1/beepers/pings_controller.rb)
- **严重级别**：中 (Medium)
- **状态**：✅ 已修复（配置 Rails `rate_limit to: 60, within: 1.minute` 并在超限时返回 429）

#### 问题分析

`/api/v1/ping/:token` 是面向外部系统的公开端点（无需用户登录态）。在控制器实现中：

```ruby
def create
  token = params[:token].to_s.strip
  if token.blank?
    head :bad_request
  elsif (beeper = Beeper.find_by_ping_token(token))
    beeper.record_ping
    head :ok
  else
    head :not_found
  end
end
```

只要 token 正确，每次请求都会直接执行数据库写入：`update_columns(signal_metadata: meta, updated_at: Time.current)`。若外部程序配置失误（死循环 ping）或遭受恶意高频调用，会导致 SQLite 数据库遭遇高频写锁争用（SQLite busy locks），进而阻塞核心业务。

#### 修复方案

在控制器中引入 Rate Limiting：

```ruby
rate_limit to: 60, within: 1.minute, by: -> { params[:token].presence || request.remote_ip }, with: -> { head :too_many_requests }, store: CACHE_STORE
```

---

## 二、业务逻辑与可靠性缺陷 (Logic & Reliability Bugs)

### 1. 暂停状态 (Paused) 下手动触发运行会导致监控被意外激活

- **相关文件**：[`core/app/models/beeper.rb`](core/app/models/beeper.rb#L55-L66), [`core/app/models/beeper.rb`](core/app/models/beeper.rb#L120-L131)
- **严重级别**：高 (High)

#### 问题分析

1. 当用户将某个 Beeper 设置为 `paused` 后，若在控制台点击了“Run Now”（手动触发一次运行），会调用：
   ```ruby
   def trigger_run!
     scheduled_for = Time.current
     update!(status: :firing, next_run_at: nil)
     ...
   end
   ```
   此时 `status` 从 `"paused"` 被强行改为了 `"firing"`。
2. 当后台任务执行完成并调用 `finish_firing` 时：
   ```ruby
   def finish_firing(last_run_at:)
     reload

     next_time = calculate_next_run_at(from: Time.current)
     if paused?
       update!(last_run_at: last_run_at, next_run_at: next_time)
     elsif firing?
       update!(status: :active, next_run_at: next_time, last_run_at: last_run_at)
     else
       update!(last_run_at: last_run_at)
     end
   end
   ```
   由于此时状态已是 `"firing"`，`paused?` 判定为 `false`，代码命中了 `elsif firing?` 分支，直接将监控状态重置为 `status: :active` 并重新计算了下一次运行时间。
3. **结果**：原本被用户主动暂停的监控在单次手动执行后被**静默恢复自动调度**。

#### 修复建议

在 `Beeper` 模型中区分“运行中”状态与“启用/暂停”主状态，或者在 `trigger_run!` 时不要破坏原始的主状态（或仅在 `active` 状态下完成自动转 `active`）。

---

### 2. 告警通知字段超长导致持久化失败 & 告警永久静默丢失

- **相关文件**：[`core/app/models/beeper.rb`](core/app/models/beeper.rb#L168-L178), [`core/app/models/beeper_run.rb`](core/app/models/beeper_run.rb#L27-L55)
- **严重级别**：高 (High)

#### 问题分析

在 [`Beeper#notify_from!`](core/app/models/beeper.rb#L168-L178) 中：

```ruby
def notify_from!(signal)
  channels = Array(notification_channels).presence || account.owner_user.notification_channels
  account.beeps.create!(
    kind: :once,
    title: signal.title.presence || title,
    body: signal.message,
    timezone: timezone,
    notification_channels: channels,
    beeper: self
  )
end
```

同时在 [`BeeperRun#execute_now`](core/app/models/beeper_run.rb#L37-L45) 中：

```ruby
beeper.update!(
  alert_state: decision.next_alert_state,
  consecutive_failures: decision.next_consecutive_failures
)

if decision.should_notify
  beeper.notify_from!(signal)
end
```

存在以下连锁隐患：
1. **长度校验失败异常**：`Beep` 模型定义了 `TITLE_MAX_LENGTH = 80` 和 `BODY_MAX_LENGTH = 2000`。当网络异常、SSL 异常或第三方响应返回较长的错误信息时（例如包含长堆栈或复杂 JSON），`account.beeps.create!` 会抛出 `ActiveRecord::RecordInvalid` 校验异常。
2. **告警永久静默丢失**：因为在调用 `notify_from!` 之前，`beeper.update!(alert_state: "alerting")` 已经将数据库状态改成了 `alerting`。抛出异常后虽然本次任务标记为 `failed`，但当下一次轮询到来时，状态机判定 `previous_state == "alerting"` 且 `signal_status == "alerting"`，根据规则 `should_notify: false`（无需重复报警）。因此，**该故障的报警通知将彻底丢失，用户永远收不到警报**。

#### 修复建议

1. 在 `notify_from!` 中主动对 `title` 和 `body` 进行安全截断：
   ```ruby
   title: (signal.title.presence || title).truncate(Beep::TITLE_MAX_LENGTH),
   body: signal.message&.truncate(Beep::BODY_MAX_LENGTH),
   ```
2. 将 `beeper` 状态更新与通知分发放入同一个事务，或在通知成功后才确认状态转移。

---

### 3. Heartbeat 监控新建时立即触发误报警 (False Positive)

- **相关文件**：[`core/app/models/beeper_app/receivers/heartbeat.rb`](core/app/models/beeper_app/receivers/heartbeat.rb#L8-L15)
- **严重级别**：中 (Medium)

#### 问题分析

在 [`BeeperApp::Receivers::Heartbeat#call`](core/app/models/beeper_app/receivers/heartbeat.rb#L8-L15) 中：

```ruby
if last_ping_at.nil?
  return BeeperApp::Signal.new(
    status: :alerting,
    title: "Heartbeat never received",
    message: "No ping has ever been received for this heartbeat monitor (grace period: #{grace_period_minutes}m)",
    metrics: { "minutes_since_last_ping" => nil }
  )
end
```

当用户刚刚在系统中创建一个 Heartbeat 监控实例时，`last_ping_at` 初始必然为 `nil`。如果用户的外部服务需要 10 分钟后才启动首次上报，但 Beeper 的调度周期为 5 分钟且 `failure_threshold = 2`，那么在创建 10 分钟内（轮询 2 次失败），系统就会直接触发报警，即使配置的宽限期（`grace_period_minutes`）为 60 分钟。

#### 修复建议

当 `last_ping_at.nil?` 时，应计算自该 Beeper 创建（`beeper.created_at`）以来的时长；若创建时间未超出 `grace_period_minutes`，应视为健康（`:ok` 或等待中），而非直接判定为 `:alerting`。

---

### 4. Hostname 清洗逻辑缺陷 (Basic Auth 与 IPv6 破坏)

- **相关文件**：[`core/app/models/beeper_app/receivers/ssl_expiry.rb`](core/app/models/beeper_app/receivers/ssl_expiry.rb#L78-L84)
- **严重级别**：低 (Low)

#### 问题分析

```ruby
def sanitize_hostname(value)
  cleaned = value.sub(%r{\A[a-zA-Z]+://}, "")
  cleaned = cleaned.split("/").first || ""
  cleaned.split(":").first || ""
end
```

1. 若用户输入 `https://user:pass@example.com`，`cleaned.split(":")` 会截取第 0 个元素得到 `"user"`，导致将用户名当作域名解析。
2. 若用户输入合法的 IPv6 目标如 `[2001:db8::1]`，冒号分割会直接破坏 IPv6 地址。

#### 修复建议

使用标准 URI 解析库进行提取：

```ruby
def sanitize_hostname(value)
  url = value.to_s.strip
  url = "https://#{url}" unless url.match?(%r{\A[a-zA-Z]+://})
  URI.parse(url).host || ""
rescue URI::InvalidURIError
  value.to_s.strip
end
```

---

## 三、架构与代码质量改进建议 (Improvements)

### 1. Config 参数缺乏基于 Manifest Inputs 的校验

- **相关文件**：[`core/app/controllers/api/v1/beepers_controller.rb`](core/app/controllers/api/v1/beepers_controller.rb#L79), [`core/app/models/beeper.rb`](core/app/models/beeper.rb)
- **分析**：
  控制器通过 `params[:config].to_unsafe_h` 允许任意键值存入数据库 `config` 字段。虽然在 Receiver 内部做了部分 fallback，但如果传入了负数 `timeout_ms: -100` 或非法的 `expected_status`，会导致探测器异常。
- **建议**：
  在 `Beeper` 模型中增加 `validate :validate_config_against_manifest`，根据关联的 `beeper_app.inputs` 自动验证字段类型、必填项、数值区间（`min`/`max`）及枚举（`options`）。

---

### 2. 前端 Heartbeat 详情页缺少 Ping Webhook URL 复制提示

- **相关文件**：[`apps/web/src/routes/$account_slug/beepers_.$beeperId.tsx`](apps/web/src/routes/$account_slug/beepers_.$beeperId.tsx)
- **分析**：
  对于 `webhook` 类型的 Beeper（如 Heartbeat），用户创建后最核心的操作是获取形如 `https://core.example.com/api/v1/ping/<token>` 的 Ping Webhook URL 并配置到自己的定时任务中。当前详情页仅展示了原始配置与执行历史，缺少直接生成并提供一键复制 Ping URL 的交互组件。
- **建议**：
  在前端识别到 `beeper.beeper_app.ingest.webhook == true` 时，在详情页顶部醒目展示专属 Ping URL 和 curl 示例命令。

---

## 四、问题清单与优先级矩阵

| 序号 | 类别   | 问题描述                                     | 严重级别 | 对应文件                                                   | 状态           |
| :--- | :----- | :------------------------------------------- | :------- | :--------------------------------------------------------- | :------------- |
| 1    | 安全   | SSRF 对纯 IP 失效 & 本地 Fake-IP 绕过        | 🔴 高    | `core/app/models/ssrf_protection.rb`                       | ✅ 已修复       |
| 2    | 安全   | HTTP 探测未限制响应体大小导致 OOM 隐患       | 🔴 高    | `core/app/models/beeper_app/receivers/site_uptime.rb`       | ✅ 已修复       |
| 3    | 安全   | SSL 握手无超时控制存在慢速挂起风险           | 🟡 中    | `core/app/models/beeper_app/receivers/ssl_expiry.rb`        | ✅ 已修复       |
| 4    | 安全   | Webhook Ping 缺少频率限制                    | 🟡 中    | `core/app/controllers/api/v1/beepers/pings_controller.rb` | ✅ 已修复       |
| 5    | 逻辑   | Paused 状态下手动触发运行导致监控被意外激活   | 🔴 高    | `core/app/models/beeper.rb`                                | 待修复         |
| 6    | 可靠性 | 告警字段未截断导致创建失败 & 告警永久丢失    | 🔴 高    | `core/app/models/beeper.rb`, `beeper_run.rb`               | 待修复         |
| 7    | 逻辑   | Heartbeat 监控新建未满宽限期即报假警         | 🟡 中    | `core/app/models/beeper_app/receivers/heartbeat.rb`        | 待修复         |
| 8    | 边界   | Hostname 清洗破坏 Basic Auth 与 IPv6         | 🟢 低    | `core/app/models/beeper_app/receivers/ssl_expiry.rb`        | 待修复         |
| 9    | 优化   | Config 参数缺乏基于 Manifest 的格式校验      | 🟢 低    | `core/app/models/beeper.rb`                                | 待修复         |
| 10   | 体验   | 详情页缺少 Ping Webhook URL 一键复制         | 🟢 低    | `apps/web/src/routes/$account_slug/beepers_.$beeperId.tsx`  | 待修复         |
