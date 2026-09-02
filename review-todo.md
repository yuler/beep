# Self-hosted Runner (Phase 1) Code Review 指南

本指南用于 Review **Self-hosted Runner 阶段 1**（PR [#38](https://github.com/yuler/beep/pull/38)）的代码改动。

* **架构设计文档**：[`docs/architecture/runner.md`](docs/architecture/runner.md)
* **对应 PR**：[yuler/beep#38](https://github.com/yuler/beep/pull/38)
* **分支**：`feat/self-hosted-runner`

---

## 1. 核心改动概览 (Change Scope)

阶段 1 实现了 Self-hosted Runner 的**后端核心能力**与**Go 客户端最小闭环**：

```
beep monorepo
├── core/                               # Rails 8.1 后端核心
│   ├── app/models/runner.rb            # Runner 模型、Token 鉴权、心跳与离线判定
│   ├── app/controllers/api/v1/
│   │   ├── runners_controller.rb       # 用户控制台 Runner CRUD 与 Token 重新生成
│   │   └── runner/                     # Runner Agent 通信 API (ping, poll, result)
│   ├── app/views/api/v1/runner*/       # Jbuilder 视图契约
│   └── test/                           # 19 个单元与集成测试
└── apps/runner/                        # Go 客户端 (beep-runner 单二进制)
    ├── cmd / main.go                   # CLI 命令 (run, ping, test, version)
    ├── internal/client/                # HTTP 长轮询与结果上报客户端
    ├── internal/daemon/                # 常驻 Worker Pool 与并发调度器
    ├── internal/probe/                 # HTTP/HTTPS, TLS, TCP, DNS 探活引擎
    └── internal/exec/                  # 本地 Shell 脚本安全执行引擎
```

---

## 2. Review Checklist (检查清单)

### 2.1 数据库迁移与数据模型 (Data Models)
- [ ] [`core/db/migrate/20260901180000_create_runners_and_add_runner_to_beepers.rb`](core/db/migrate/20260901180000_create_runners_and_add_runner_to_beepers.rb)
  - [ ] `runners` 表使用 UUID 主键，`token_digest` 有唯一索引。
  - [ ] `[account_id, status]` 组合索引满足高频租户状态查询。
  - [ ] `beepers` 表新增可选 `runner_id` 与 `runner_tag` 字段。
  - [ ] `beeper_runs` 表新增可选 `runner_id` 与 `claimed_at` 字段。
- [ ] [`core/app/models/runner.rb`](core/app/models/runner.rb)
  - [ ] Token 生成采用 `beep_rt_` 前缀 + 24 字节高熵随机串，数据库仅保存 SHA256 哈希值 (`token_digest`)。
  - [ ] `raw_token` 仅在 `create` 或 `regenerate_token!` 时在内存中短暂暴露一次。
  - [ ] `matches_tag?` 正确匹配标签（支持空 tag 全匹配与具体 tag 精确匹配）。
  - [ ] `Runner.mark_stale_offline!` 超过 60 秒无心跳自动标记为 `offline`。
- [ ] [`core/app/models/beeper.rb`](core/app/models/beeper.rb) 与 [`beeper_run.rb`](core/app/models/beeper_run.rb)
  - [ ] 当 `requires_runner?` 为真时，`claim_run` 与 `trigger_run!` 不触发 Rails 进程内 Ruby 检查，而是保持 `pending` 状态等待 Runner 抢占。
  - [ ] `BeeperRun#record_signal_result!` 统一接入 `Beeper::AlertPolicy` 状态机。

### 2.2 API 端点与安全性 (Security & API)
- [ ] [`core/config/initializers/account_slug.rb`](core/config/initializers/account_slug.rb)
  - [ ] `runner` 与 `runners` 正确加入 `RESERVED_FROM_ROUTES`，防止路由中间件将其误识别为租户 slug。
- [ ] [`core/app/controllers/api/v1/runner/base_controller.rb`](core/app/controllers/api/v1/runner/base_controller.rb)
  - [ ] 支持 `X-Runner-Token` 请求头或 `Authorization: Bearer <token>` 认证。
  - [ ] 无效或缺失 Token 时统一返回 401 Unauthorized。
- [ ] [`core/app/controllers/api/v1/runner/tasks_controller.rb`](core/app/controllers/api/v1/runner/tasks_controller.rb)
  - [ ] `poll` 动作严格按租户 `account_id` 隔离，且只能领取分配给自身 `runner_id` 或匹配 `runner_tag` 的待执行任务。
  - [ ] 抢占任务通过 `update_all(status: 'running', ...)` 原子执行，防止多 Runner 并发重复领取。
  - [ ] `result` 动作验证任务所属权，安全构造 `BeeperApp::Signal` 并驱动告警状态机。
- [ ] [`core/app/controllers/api/v1/runners_controller.rb`](core/app/controllers/api/v1/runners_controller.rb)
  - [ ] 遵循 Jbuilder 规范（无 inline JSON hash），严格按 `Current.account` 权限边界过滤。

### 2.3 Go Runner 客户端 (`apps/runner`)
- [ ] **CLI 体验与命令行入口** ([`main.go`](apps/runner/main.go))
  - [ ] 默认或 `beep-runner run`：启动后台守护进程。
  - [ ] `beep-runner ping`：连通性与认证诊断。
  - [ ] `beep-runner test <http|tls|tcp|dns|exec> <target>`：单次本地探活测试。
- [ ] **探活引擎** ([`apps/runner/internal/probe/`](apps/runner/internal/probe/))
  - [ ] `http.go`：支持状态码断言、响应耗时、重定向限制 (5次)、响应体关键词包含检测。
  - [ ] `tls.go`：支持证书剩余天数计算与阈值告警。
  - [ ] `tcp.go` / `dns.go`：支持端口连通性与 DNS 延时检测。
- [ ] **脚本执行安全边界** ([`apps/runner/internal/exec/exec.go`](apps/runner/internal/exec/exec.go))
  - [ ] 默认关闭脚本执行，必须传入 `--allow-exec` 或 `BEEP_ALLOW_EXEC=1` 显式开启。
  - [ ] 命令输出强制截断为 8KB，超时强制中断进程。
- [ ] **Worker Pool 调度器** ([`apps/runner/internal/daemon/daemon.go`](apps/runner/internal/daemon/daemon.go))
  - [ ] 基于 channel semaphore 实现并发控制 (`--concurrency`)。
  - [ ] 优雅退出处理（监听 `SIGINT` / `SIGTERM`，等待正在执行的任务完成后再退出）。

---

## 3. 本地验证步骤 (Verification Steps)

### Step 1: 运行后端与客户端自动化测试
```bash
# 1. 运行 Core 后端全量测试 (包括 Runner 模型与 Controller 测试)
cd core && bin/rails test

# 2. 运行 Go Runner 客户端全量单元测试
cd ../apps/runner && go test -v ./...
```

### Step 2: 编译与测试 Go Runner CLI
```bash
# 在项目根目录使用 mise 任务编译
mise run runner:build

# 1. 测试版本输出
./bin/beep-runner version

# 2. 测试离线命令执行探活
./bin/beep-runner test exec "echo 'Hello from Beep Runner'"

# 3. 测试离线 DNS 探活
./bin/beep-runner test dns localhost

# 4. 测试离线 HTTP 探活
./bin/beep-runner test http https://example.com
```

### Step 3: 端到端联调测试 (Core API + Runner 守护进程)
1. 启动 Core 服务：`mise run core:dev`（默认监听 `http://core.localhost:3000`）。
2. 在 Rails Console 或通过 API 为账户创建一个 Runner，获取 Token（例如 `beep_rt_xxx`）：
   ```ruby
   # bin/rails console
   account = Account.first
   runner = account.runners.create!(name: "Local-Dev-Runner", allow_exec: true)
   puts runner.raw_token
   ```
3. 测试 Runner Ping 认证：
   ```bash
   ./bin/beep-runner ping --server http://core.localhost:3000 --token beep_rt_xxx
   ```
4. 启动 Runner 常驻监听：
   ```bash
   ./bin/beep-runner run --server http://core.localhost:3000 --token beep_rt_xxx --allow-exec
   ```
5. 在 Rails Console 中创建一个绑定该 Runner 的 Beeper 并触发运行：
   ```ruby
   app = BeeperApp.find_by!(slug: "site-uptime")
   beeper = account.beepers.create!(
     beeper_app: app,
     runner: runner,
     title: "Localhost HTTP Check",
     cron: "*/5 * * * *",
     timezone: "UTC",
     config: { "target_url" => "http://127.0.0.1:3000/up" }
   )
   beeper.trigger_run!
   ```
6. 观察 Runner 终端输出：成功拉取任务、执行探测并回传结果给 Core。
7. 在 Rails Console 确认 `BeeperRun` 状态变为 `succeeded`，`signal_status` 变为 `ok`。
