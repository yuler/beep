# Self-hosted Runner (Phase 1 & 2) Code Review 指南

本指南用于 Review **Self-hosted Runner 完整功能**（PR [#38](https://github.com/yuler/beep/pull/38)）的代码改动。

* **架构设计文档**：[`docs/architecture/runner.md`](docs/architecture/runner.md)
* **对应 PR**：[yuler/beep#38](https://github.com/yuler/beep/pull/38)
* **分支**：`feat/self-hosted-runner`

---

## 1. 核心改动概览 (Change Scope)

PR 涵盖了 Self-hosted Runner 的 **阶段 1（后端核心 + Go 客户端）** 与 **阶段 2（Web 管理 UI + Beeper 路由选择 + 容器打包）**：

```
beep monorepo
├── core/                               # Rails 8.1 后端核心
│   ├── app/models/runner.rb            # Runner 模型、Token 鉴权、心跳与离线判定
│   ├── app/controllers/api/v1/
│   │   ├── runners_controller.rb       # 用户控制台 Runner CRUD 与 Token 重新生成
│   │   ├── beepers_controller.rb       # Beeper runner_id / runner_tag 参数支持
│   │   └── runner/                     # Runner Agent 通信 API (ping, poll, result)
│   ├── app/views/api/v1/runner*/       # Jbuilder 视图契约 (runner & beeper)
│   └── test/                           # 单元与集成测试 (Runner 与 Beeper 关联)
├── apps/web/                           # Web 前端控制台 (TanStack Start / React 19)
│   ├── src/routes/$account_slug/
│   │   ├── runners.tsx                 # Runner 节点管理主页面
│   │   ├── beepers.tsx                 # Beeper 安装表单 (集成 Runner 路由选择)
│   │   └── beepers_.$beeperId.tsx      # Beeper 详情页 (展示执行节点路由)
│   ├── src/components/runners/         # Runner 列表、添加/编辑弹窗、Token 引导弹窗
│   ├── src/components/beepers/         # Beeper 路由选择器 (RunnerRoutingPicker)
│   └── src/lib/api/runners.ts          # TypeScript API 客户端
├── apps/runner/                        # Go 客户端 (beep-runner 单二进制与 Dockerfile)
│   ├── Dockerfile                      # 多架构轻量级镜像打包 (Go 多阶段静态编译)
│   ├── main.go                         # CLI 命令 (run, ping, test, version)
│   ├── internal/client/                # HTTP 长轮询与结果上报客户端
│   ├── internal/daemon/                # 常驻 Worker Pool 与并发调度器
│   ├── internal/probe/                 # HTTP/HTTPS, TLS, TCP, DNS 探活引擎
│   └── internal/exec/                  # 本地 Shell 脚本安全执行引擎
└── project.inlang/messages/            # 中英文国际化翻译词条
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
  - [ ] `online?` 辅助方法依据状态与 `OFFLINE_TIMEOUT`（60s）精准判定节点可用性。
  - [ ] `Runner.mark_stale_offline!` 超过 60 秒无心跳自动标记为 `offline`（在后台 `BeeperPollerJob` 中周期自动执行）。
- [ ] [`core/app/models/beeper.rb`](core/app/models/beeper.rb) 与 [`beeper_run.rb`](core/app/models/beeper_run.rb)
  - [ ] `has_online_runner?` 正确识别单个节点或 Tag 匹配的在线 Runner 可用性。
  - [ ] **离线防卡死机制**：当 `requires_runner?` 且 Runner 处于离线状态时，调度或手动触发立即记录 `status: :error` 信号（"Runner offline"），累加连续失败并更新告警状态机，自动恢复 `active` 并排期下次调度，防止 Beeper 永久卡死在 `firing` 状态。
  - [ ] **Pending 超时回收**：`reclaim_stale_firing` 对排队超时未被认领或 Runner 失联的任务自动产出离线错误信号并驱动告警。
  - [ ] **Running 执行超时回收**：对 Runner 领单后崩溃或超时的任务（`RUNNING_STALE_AFTER`）记录执行超时错误信号并驱动告警。
  - [ ] `BeeperRun#record_signal_result!` 统一接入 `Beeper::AlertPolicy` 状态机。

### 2.2 Core API 端点与安全性 (Security & Core API)
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
  - [ ] 支持 Runner CRUD 与 `regenerate_token`。
- [ ] [`core/app/controllers/api/v1/beepers_controller.rb`](core/app/controllers/api/v1/beepers_controller.rb) & [`core/app/views/api/v1/beepers/_beeper.json.jbuilder`](core/app/views/api/v1/beepers/_beeper.json.jbuilder)
  - [ ] `beeper_params` 与 `update_params` 正确放行 `:runner_id` 与 `:runner_tag`。
  - [ ] 视图序列化输出 `has_online_runner` 以及 `runner.is_online`、`runner.last_seen_at` 供前端判定。

### 2.3 Web 管理 UI 与路由交互 (`apps/web`)
- [ ] **工作区侧边栏导航** ([`apps/web/src/components/dashboard/dashboard-sidebar.tsx`](apps/web/src/components/dashboard/dashboard-sidebar.tsx))
  - [ ] 在 Workspace 分组下新增 `Runners` 一级入口，高亮与路由匹配正确。
- [ ] **Runner 控制台与列表** ([`apps/web/src/routes/$account_slug/runners.tsx`](apps/web/src/routes/$account_slug/runners.tsx) & [`runner-list.tsx`](apps/web/src/components/runners/runner-list.tsx))
  - [ ] 状态 Badge 区分在线（绿色脉冲）、空闲（蓝色）、离线（灰色）。
  - [ ] 展示版本号、系统架构、主机名、IP 地址、关联 Beeper 数量、最后活跃时间与 exec 权限标签。
  - [ ] 支持编辑 Runner、重新生成 Token 以及删除 Runner。
- [ ] **Token 接入与运行引导弹窗** ([`apps/web/src/components/runners/runner-token-modal.tsx`](apps/web/src/components/runners/runner-token-modal.tsx))
  - [ ] 醒目展示高熵 Token 并提示不可再次查看。
  - [ ] 提供一键复制 Docker Run 命令、Docker Compose 配置片段、原生 CLI 命令行。
- [ ] **Beeper 执行节点路由选择与详情页** ([`apps/web/src/components/beepers/runner-routing-picker.tsx`](apps/web/src/components/beepers/runner-routing-picker.tsx) & [`beepers_.$beeperId.tsx`](apps/web/src/routes/$account_slug/beepers_.$beeperId.tsx))
  - [ ] 三选一模式：云端 Core（默认）、指定特定 Runner 节点（下拉选择）、按 Tag 标签动态调度（输入 Tag）。
  - [ ] 安装弹窗与编辑弹窗均支持配置路由。
  - [ ] Beeper 详情页展示执行节点归属及实时在线/离线状态 Badge。
  - [ ] 当绑定的 Runner 离线或 Tag 无在线节点时，顶部醒目展示离线警告横幅（Warning Banner）。

### 2.4 Go Runner 客户端与 Docker 打包 (`apps/runner`)
- [ ] **CLI 体验与命令行入口** ([`main.go`](apps/runner/main.go))
  - [ ] 默认或 `beep-runner run`：启动后台守护进程。
  - [ ] `beep-runner ping`：连通性与认证诊断。
  - [ ] `beep-runner test <http|tls|tcp|dns|exec> <target>`：单次本地探活测试。
- [ ] **探活引擎** ([`apps/runner/internal/probe/`](apps/runner/internal/probe/))
  - [ ] `http.go`：支持状态码断言、响应耗时、重定向限制 (5次)、响应体关键词包含检测。
  - [ ] `tls.go`：支持证书剩余天数计算与阈值告警。
  - [ ] `tcp.go` / `dns.go`：支持端口连通性与 DNS 延时检测。
- [ ] **脚本执行安全边界** ([`apps/runner/internal/exec/exec.go`](apps/runner/internal/exec/exec.go))
  - [ ] 默认关闭脚本执行，必须传入 `--allow-exec` 或 `BEEP_ALLOW_EXEC=true` 显式开启。
  - [ ] 命令输出强制截断为 8KB，超时强制中断进程。
- [ ] **Worker Pool 调度器** ([`apps/runner/internal/daemon/daemon.go`](apps/runner/internal/daemon/daemon.go))
  - [ ] 基于 channel semaphore 实现并发控制 (`--concurrency`)。
  - [ ] 优雅退出处理（监听 `SIGINT` / `SIGTERM`，等待正在执行的任务完成后再退出）。
- [ ] **Docker 容器构建** ([`apps/runner/Dockerfile`](apps/runner/Dockerfile))
  - [ ] 多阶段构建（`CGO_ENABLED=0` 静态编译），基础镜像精简为 Alpine。
  - [ ] 包含必要探活工具（`ca-certificates`, `tzdata`, `curl`, `bind-tools`），以非 root 用户运行。

---

## 3. 本地验证步骤 (Verification Steps)

### Step 1: 运行后端、前端与客户端自动化测试
```bash
# 1. 运行 Core 后端全量测试 (包含 Runner 与 Beeper 关联及离线容错测试)
cd core && bin/rails test

# 2. 运行 Web 前端类型与 Biome 检查
pnpm --filter web check
pnpm --filter web build

# 3. 运行 Go Runner 客户端全量单元测试
cd apps/runner && go test -v ./...
```

### Step 2: 编译与测试 Go Runner CLI
```bash
# 在项目根目录编译
cd apps/runner && go build -o ../../bin/beep-runner ./main.go && cd ../..

# 1. 测试版本输出
./bin/beep-runner version

# 2. 测试本地命令探活
./bin/beep-runner test exec "echo 'Hello from Beep Runner'"

# 3. 测试本地 HTTP 探活
./bin/beep-runner test http https://example.com
```

### Step 3: Web 端到端管理与执行联调
1. 启动全栈开发环境：`mise dev`
2. 打开浏览器访问 `http://web.beep.localhost:3000` 并登录。
3. 进入左侧侧边栏 **Runners**（`/$slug/runners`）：
   - 点击 **Add Runner** 创建一个节点（例如 `Office-Mac`，标签 `intranet`，开启 allow exec）。
   - 弹窗中复制生成的 Token（`beep_rt_xxx`）及 Docker/CLI 启动命令。
4. 在本地终端启动 Runner 守护进程：
   ```bash
   ./bin/beep-runner run --server http://core.beep.localhost:3000 --token <YOUR_TOKEN> --allow-exec
   ```
5. 刷新 Web 页面：Runner 状态实时变为 `Online`（绿色脉冲），系统架构、版本号、主机名正确上报。
6. 进入 **Beepers** 页面，安装或编辑一个 Beeper（如 Site Uptime）：
   - 在 **Execution Target** 选择 **Specific Runner Node**（选择刚刚启动的 Runner）。
   - 安装完成后点击 **Trigger Run**。
7. 查看终端：Runner 成功领取任务、探活执行并上报结果。Web 端详情页即时展示执行成功的 Run 记录。
8. **验证离线容错**：终止本地 Runner 进程（或将其状态置为 offline），返回 Beeper 详情页可看到离线警告横幅与 Badge；再次触发任务，系统立即记录 `Runner offline` 错误信号，Beeper 恢复为 Active 并正确计算下一次调度时间。
