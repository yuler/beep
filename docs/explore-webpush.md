# Explore: Web Push (Rails)

调研来源：

- [basecamp/once-campfire](https://github.com/basecamp/once-campfire) — 同域 Rails + Stimulus，实现更直、更适合对照「订阅 + 发送」
- [basecamp/fizzy](https://github.com/basecamp/fizzy) — 同一套投递内核，加上租户、SSRF、失效订阅清理、站内通知 / 原生推送分流
- Beep：`core`（Rails 8.1 API）+ `apps/web`（TanStack Router，另一 origin）

结论先说：**发送侧几乎可以原样搬 `web-push` + `WebPush::Pool`；订阅侧必须拆。** Service Worker 只能挂在 `apps/web` 的 origin 上，`core` 只负责 VAPID、存订阅、在 `DeliverBeepRunJob` 里 POST 到浏览器厂商的 push service。

---

## 1. Web Push 在做什么

浏览器 **Push API** 不是「服务端直连用户电脑」。链路是：

```
apps/web (SW + PushManager)
    → POST 订阅 { endpoint, p256dh, auth } 到 core
core
    → VAPID 签名 + RFC 8291 加密 payload
    → HTTPS POST 到 FCM / Mozilla / Apple / WNS
push service
    → 唤醒该 origin 上的 Service Worker
SW
    → showNotification / setAppBadge
```

三块密钥：

| 字段        | 谁生成                         | 作用                                      |
| ----------- | ------------------------------ | ----------------------------------------- |
| VAPID 密钥对 | 应用服务器（一次，长期固定）   | 证明「是这个应用在推」                    |
| `p256dh`    | 浏览器每次订阅                 | 加密 payload 的客户端公钥                 |
| `auth`      | 浏览器每次订阅                 | 认证密钥                                  |
| `endpoint`  | 浏览器每次订阅                 | 该设备在 push service 上的投递 URL        |

VAPID **公钥**给前端 `pushManager.subscribe({ applicationServerKey })`；**私钥**只留在 `core`。换密钥会让已有订阅全部失效。

浏览器约束（产品上必须面对）：

- 必须 HTTPS（本地 `localhost` 例外）
- 必须先 `Notification.requestPermission()`，且 `userVisibleOnly: true`（不能静默推）
- iOS Safari：用户把站点加到主屏幕（standalone PWA）后才有 Web Push
- 用户关掉权限 / 清站点数据后，下次发送会得到过期订阅，服务端要删记录

---

## 2. 两仓库共用的投递内核

Campfire 和 Fizzy 的发送层几乎同一份代码。Gem：

```ruby
gem "web-push"
gem "net-http-persistent"
```

`concurrent-ruby` 随 Rails 提供，用来跑投递线程池。

### 2.1 VAPID

Campfire 允许 ENV 或 `credentials`：

```ruby
# once-campfire config/initializers/vapid.rb
config.x.vapid.private_key = ENV.fetch("VAPID_PRIVATE_KEY", Rails.application.credentials.dig(:vapid, :private_key))
config.x.vapid.public_key  = ENV.fetch("VAPID_PUBLIC_KEY",  Rails.application.credentials.dig(:vapid, :public_key))
```

Fizzy 只用 ENV。生成：

```ruby
# once-campfire script/admin/create-vapid-key
vapid_key = WebPush.generate_key
# private_key / public_key 写入 ENV 或 credentials
```

前端读 `<meta name="vapid-public-key">`（Beep 应改成 API 返回公钥，见 §5）。

### 2.2 `WebPush::Notification`

把业务文案编成 SW 约定的 JSON，再调 `WebPush.payload_send`：

```json
{
  "title": "…",
  "options": {
    "body": "…",
    "icon": "/…",
    "data": { "path": "/rooms/1", "badge": 3 }
  }
}
```

Fizzy 用 `data.url`（绝对或站点内路径）；Campfire 用 `data.path`。两边 SW 的 `notificationclick` 都是：关通知 → 已有 focused window 则 `navigate`，否则 `openWindow`。同时 `navigator.setAppBadge?.(badge)`。

`urgency: "high"` 提高投递优先级。VAPID `subject` 是 `mailto:` 联系地址（协议要求，可写成产品支持邮箱）。

### 2.3 `WebPush::Pool`（不要在 ActiveJob 里同步打 HTTP）

设计意图：一条业务事件可能打到 **很多设备**（Campfire 一个房间、多名离线用户）。Job 里只做 AR 查询，HTTP 丢给进程内线程池：

| 组件                | 配置                                      | 职责                                      |
| ------------------- | ----------------------------------------- | ----------------------------------------- |
| `delivery_pool`     | `max_threads: 50`, `queue_size: 10000`    | 真正 `payload_send`                       |
| `invalidation_pool` | 单线程                                    | 删过期订阅（要回 Rails executor）          |
| `connection`        | `Net::HTTP::Persistent` `pool_size: 150`  | 复用到 FCM 等的 TLS                       |

流程：

1. 在调用线程里 `subscription.notification(**payload)`（读 AR、算 badge）
2. `delivery_pool.post` 里 `notification.deliver(connection:)`
3. 捕获 `WebPush::ExpiredSubscription` / `OpenSSL::OpenSSLError` → 失效池里 `Push::Subscription.find_by(id:)&.destroy`
4. 队列满则 `Concurrent::RejectedExecutionError`，静默丢弃（不拖垮请求）
5. `at_exit` shutdown persistent HTTP + 线程池

`lib/web_push/*` 刻意不经过 Rails autoload 的热路径，方便在 executor 外跑。失效回调必须 `Rails.application.executor.wrap`。

### 2.4 给 `web-push` gem 打补丁

`config/initializers/web_push.rb` `prepend` 了 `WebPush::PersistentRequest`：`WebPush::Request#perform` 在传入 `connection:`（`Net::HTTP::Persistent`）时走持久连接，否则退回普通 `Net::HTTP`。

Fizzy 额外支持 `endpoint_ip:`，设置 `http.ipaddr = endpoint_ip`，把 TLS 钉在订阅时解析出的公网 IP 上，避免发送时 DNS 被换成内网地址（SSRF）。Campfire 没有这一步。

---

## 3. 订阅模型

表名 `push_subscriptions`（`Push` 模块 `table_name_prefix`）。

### Campfire（简单）

| 列           | 说明                    |
| ------------ | ----------------------- |
| `user_id`    | 所属用户                |
| `endpoint`   | string                  |
| `p256dh_key` |                         |
| `auth_key`   |                         |
| `user_agent` | 设置页展示浏览器 / 系统 |
| timestamps   |                         |

索引：`(endpoint, p256dh_key, auth_key)`（非 unique 声明在 schema 里是普通 index）。创建逻辑：三元组已存在则 `touch`，否则 `create!`。无 HTTPS / host allowlist。

`Push::Subscription#notification` 把 `badge: user.memberships.unread.count` 填进 payload。

测试推送：`Users::PushSubscriptions::TestNotificationsController` 对单条订阅同步 `.deliver`（不走 pool）。

### Fizzy（更接近 Beep 的租户与安全）

| 列           | 说明                                      |
| ------------ | ----------------------------------------- |
| `id`         | uuid                                      |
| `account_id` | 租户                                      |
| `user_id`    | 该租户下的 User                           |
| `endpoint`   | **text**（FCM URL 可能很长）              |
| `p256dh_key` |                                           |
| `auth_key`   |                                           |
| `user_agent` | limit 4096                                |

唯一索引：`(user_id, endpoint)`，`endpoint` 前缀长度 255（MySQL text 索引限制）。迁移 `ChangeEndpointToTextInPushSubscriptions` 就是因为 string 装不下真实 endpoint。

创建：`create_with(user_agent:).create_or_find_by!(endpoint, p256dh_key, auth_key)`，HTML `204` / JSON `201`。

**订阅校验（Beep 应抄）：**

1. 必须是合法 HTTPS URL
2. host 必须落在 allowlist（`end_with?` 匹配）：
   - `fcm.googleapis.com` / `jmt17.google.com`（Chrome / Android）
   - `updates.push.services.mozilla.com`（Firefox）
   - `web.push.apple.com`（Safari）
   - `notify.windows.com`（Edge / WNS）
3. `SsrfProtection.resolve_public_ip(host)`：用 `1.1.1.1` / `8.8.8.8` 解析，丢掉 RFC1918、loopback、link-local（含 `169.254.169.254`）、CGNAT、组播、文档网段、NAT64 嵌套内网 IPv4 等；解析不到公网 IP 则拒绝
4. 发送时把该 IP 钉在 HTTP 连接上

`User` 通过 `User::Configurable`：`has_many :push_subscriptions, class_name: "Push::Subscription", dependent: :delete_all`。

---

## 4. 谁在什么时候推

### Campfire：消息 → 离线成员

`Room::PushMessageJob` → `Room::MessagePusher`：

- payload：私聊用发送者名当 title；房间用房间名 + `Name: body`
- 只推 **可见、当前未连上 Action Cable、且不是发送者** 的 membership
- `involvement: everything` 的人每条都推；`mentions` 只在被 @ 时推
- `web_push_pool.queue(payload, subscriptions)`

这是「聊天在线抑制」模型，Beep **用不上**（提醒本来就是用户不在场时才有价值）。

### Fizzy：站内 Notification 记录 → 多 target

`Notification` include `Pushable`：`after_save_commit` 若 `source_id` 变了则 `Notification::PushJob`。

`pushable?`：非系统账号、user/account 都 active。然后 `register_push_target(:web)`（SaaS 再注册 `:native`）。

`PushTarget::Web`：`user.push_subscriptions` 非空则 `web_push_pool.queue(payload.to_h, subscriptions)`。Payload 按 source 类型（Event / Mention / Default）生成 `title` / `body` / `url`。

SaaS 的 `PushTarget::Native` 走 `action_push_native`（APNs / FCM 设备表），与 Web Push **并列**，不是替换。Beep MVP 只做 Web Push。

站内通知托盘、邮件 bundling、unsubscribe token 是另一条产品线，不要和「Beep 到点投递」绑在一起。

---

## 5. 前端：SW + 订阅

两边都是 Stimulus `notifications_controller.js`，逻辑可直接翻译成 TanStack 里的 hook。

1. 能力检测：`navigator.serviceWorker && window.Notification`
2. `getRegistration`；没有则 `register("/service-worker.js", { scope: "/" })`
3. `Notification.permission`：`denied` 展示系统/浏览器设置说明；`granted` 直接 subscribe；`default` 先 `requestPermission`
4. `pushManager.subscribe({ userVisibleOnly: true, applicationServerKey: Uint8Array(vapidPublicKey) })`
5. `subscription.toJSON()` → `{ endpoint, keys: { p256dh, auth } }` POST 到后端
6. POST 失败则 `subscription.unsubscribe()`，避免浏览器有订阅、服务器没有

VAPID 公钥是 URL-safe Base64，必须 pad + 换成标准 Base64 再 `atob` 成 `Uint8Array`。

Campfire 的 SW **只处理 push**（另加 PWA install / 系统设置文案）。Fizzy 的 SW 还用 Turbo Offline 做离线缓存；Beep 不必抄缓存层。

**关键约束：** Service Worker 的脚本 URL 必须与页面 **同源**。Fizzy/Campfire 因此用稳定路径，而不是 fingerprint 后的 asset：

```ruby
get "service-worker" => "pwa#service_worker"
# skip_forgery_protection；Campfire 还 allow_unauthenticated_access
```

Beep `core` 里已有 Rails 默认 stub：

- `GET /service-worker` → `rails/pwa#service_worker`（注释掉的 push 示例）
- `GET /manifest`
- `allow_browser versions: :modern`（含 web push 能力）

这些挂在 **core origin**（`core.beep.localhost:3001`）。用户实际打开的是 **web origin**（`web.beep.localhost:3000`）。在 core 上注册 SW **无法**给 web 收推送。

生产里 TanStack Start 把 `/api` 反代到 core（same-origin），SW 仍应属于 `apps/web` 的 `scope: "/"`，不要用 core 的 PWA 路由冒充。

---

## 6. 对照 Beep：建议怎么接

产品定义见 [`TERMS.md`](TERMS.md)：Beep 到点由 Poller → `DeliverBeepRunJob` 走 Channel（MVP：`email` | `web_push`）。这是 **定时投递**，不是 Fizzy 那种「协作事件的站内通知」。更接近 Campfire 的「把一段 title/body/url 推到设备」，但接收人是 **Beep 所属 account 里该提醒的目标用户**，通常是创建者自己。

### 6.1 职责切分

| 层        | 做什么                                                                 | 不做什么                         |
| --------- | ---------------------------------------------------------------------- | -------------------------------- |
| `apps/web` | 注册 SW、权限、`PushManager.subscribe`、展示通知点击后的路由         | 不持有 VAPID 私钥、不发 push HTTP |
| `core`    | VAPID、CRUD 订阅、校验 endpoint、Pool 发送、过期删除、写入 `beep_runs.result` | 不托管给用户用的 SW               |

本地开发：web 与 core 不同 origin，CORS 已有 [`development_cors.rb`](../core/config/initializers/development_cors.rb)。`pushManager.subscribe` 仍在 web origin；把订阅 JSON `fetch` 到 `core /api/v1` 即可。

### 6.2 `core` 建议形状（对齐 Fizzy，API 化）

Gem：`web-push`、`net-http-persistent`。Initializer 可几乎照抄 Campfire 的 `web_push.rb` + Fizzy 的 `endpoint_ip` 补丁。VAPID 用 ENV（与 Fizzy 一致），文档里写清 `WebPush.generate_key`。

表 `push_subscriptions`：

- `identity_id`（人）或 `user_id`（某租户成员）—— **设备属于人**。Identity 可跨 personal/team account；提醒却挂在 Account 上。建议：**订阅挂 Identity**，投递时用 Beep 的 account 找到对应 `User` → `identity.push_subscriptions`。若 MVP 只有「给自己的提醒」，先 `user_id` + `account_id` 也够，和 Fizzy 一样。
- `endpoint` **text**、`p256dh_key`、`auth_key`、`user_agent`
- unique `(user_id, endpoint)` 或 `(identity_id, endpoint)`
- Fizzy 那套 HTTPS + host allowlist + `SsrfProtection` + IP pin。`SsrfProtection` 可整文件搬进 `core/app/models/`（它不依赖 Fizzy 业务）

API（jbuilder，不要在 controller 里拼 hash）：

| 方法   | 路径                                      | 说明                                      |
| ------ | ----------------------------------------- | ----------------------------------------- |
| GET    | `/api/v1/web_push` 或挂在 `/api/v1/me`    | `{ vapid_public_key }`                    |
| POST   | `/api/v1/push_subscriptions`              | `endpoint`, `p256dh_key`, `auth_key`      |
| DELETE | `/api/v1/push_subscriptions/:id`          | 当前 identity/user 范围                   |
| POST   | `/api/v1/push_subscriptions/:id/test`     | 可选；对齐 Campfire 测通                  |

`wrap_parameters` / strong params 与 Fizzy 相同。JSON 创建用 `create_or_find_by!`，避免重复点「开启」产生重复行。

投递：`DeliverBeepRunJob` 对 `web_push` channel 调用 `web_push_pool.queue({ title:, body:, url: }, subscriptions)`。Beep 单用户设备数很少，池仍值得用（过期删除、连接复用、不阻塞 Solid Queue worker）。payload 建议：

```json
{
  "title": "Beep",
  "options": {
    "body": "<beep.message>",
    "data": { "url": "https://web…/beeps/<id>", "badge": 1 }
  }
}
```

`url` 用 **web 的绝对 URL**（SW 在 web origin）。`result` 写入 run：订阅数为 0 应记失败/跳过（用户没开权限），和 email 通道并列。

### 6.3 `apps/web` 建议形状

- 静态或路由提供 **稳定 URL** `/service-worker.js`（不要带 content hash；更新靠 SW 自身 `skipWaiting` / 版本注释）
- `push`：`event.data.json()` → `showNotification(title, options)` + optional badge
- `notificationclick`：用 `data.url` 打开/聚焦（Fizzy 风格绝对 URL 更适合拆 origin）
- 设置页：请求权限、同步订阅、说明 Chrome / Firefox / Safari / iOS「加到主屏幕」
- 从 API 拉 VAPID 公钥，不要写死进前端包（否则轮换密钥要发版）

### 6.4 明确不要抄的

- Fizzy 整套 `Notification` / tray / email bundle / Turbo 广播
- Campfire 的「已连接房间不推」
- 把 SW 放在 `core` 的 `rails/pwa#service_worker` 上给 web 用
- SaaS native APNs（可列为以后的 Channel，不是 MVP）
- Fizzy Turbo Offline 缓存规则

### 6.5 和 Channel 模型的关系

[`TERMS.md`](TERMS.md) 的 Channel 是「这个 Beep 走哪些投递方式」。Web Push 还需要 **另一张设备表**（订阅）。关系：

- Channel `web_push` = 用户是否允许这条提醒走浏览器推送
- `push_subscriptions` = 实际能打到的设备

没有订阅时，channel 开着也应在 run result 里标明 `no_subscriptions`，并引导设置页开权限。Email 作为兜底仍然成立。

---

## 7. 实现顺序（建议）

1. `core`：VAPID ENV、`web-push` gem、`Push::Subscription` + SSRF 校验、API CRUD、测试里 stub DNS（照 Fizzy `DnsTestHelper`）
2. `apps/web`：SW + 订阅设置页 + 从 API 同步
3. `DeliverBeepRunJob` 接 `web_push`：Pool 发送、过期销毁、写 `beep_runs.result`
4. 可选：测试推送 endpoint、App Badge、iOS PWA 安装引导

验证清单：

- Chrome 开权限后，到点（或 test endpoint）弹出系统通知；点击回到对应 Beep
- 关权限或清站点数据后再发，订阅行被删除
- 伪造 `https://attacker.example.com` 或解析到 `127.0.0.1` / `169.254.169.254` 的 endpoint 被 422
- 超长 FCM endpoint 能入库（text 列）
- 本地 `web.*` / `core.*` 分 origin 仍能订阅；生产反代 same-origin 也能订阅

---

## 8. 源码索引

Campfire：

- `config/initializers/vapid.rb`, `config/initializers/web_push.rb`
- `lib/web_push/notification.rb`, `lib/web_push/pool.rb`
- `app/models/push/subscription.rb`, `app/models/room/message_pusher.rb`
- `app/controllers/users/push_subscriptions_controller.rb`
- `app/javascript/controllers/notifications_controller.js`
- `app/views/pwa/service_worker.js`
- `script/admin/create-vapid-key`

Fizzy（在 Campfire 之上）：

- `app/models/ssrf_protection.rb`, `app/models/push/subscription.rb`
- `app/models/notification/pushable.rb`, `push_target/web.rb`
- `config/initializers/push_notifications.rb`
- `app/views/pwa/service_worker.js.erb`（push + badge + click；其余是 offline cache）
- `test/models/push/subscription_test.rb`, `test/lib/web_push/persistent_request_test.rb`
- `saas/` 下 native push（非 MVP）
