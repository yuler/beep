# Code Review: PR #38 — ✨ [cli] Add Self-hosted Runner architecture, unified beep CLI, Web UI and Core APIs

- 分支：`feat/self-hosted-runner`（基线 `main`），审查时 HEAD 为 `d6c0f4b`
- 范围：125 个文件，约 +10.6k / -0.6k。涵盖 Rails 后端（Runner / RunnerJob / RunnerRun 模型、runner agent API、poller）、Go CLI（daemon、workspace、client）、TanStack Web 控制台
- 已执行的验证：
  - `apps/cli`：`go build ./...`、`go vet ./...`、`go test ./...` 全部通过
  - `core`：`runner_test` / `runner/job_test` / runner 控制器测试共 31 runs, 132 assertions 全过
  - `apps/web`：`biome check` 通过（140 files）

---

## 🟢 已修复：🔴 高 — 长任务（> 15s）丢失执行结果和后续日志

- **位置**：`apps/cli/internal/daemon/daemon.go`
- **问题**：`reportCtx` 在任务开始前创建并在整个 `execute()` 生命周期共享，超过 15s 后所有日志与结果上报均因 `context deadline exceeded` 失败。
- **修复**：
  - 移除全局共享的 `reportCtx`；
  - 在 `flushChunk`、`ReportResult` 以及缺失脚本的错误上报中，统一为每次 HTTP 调用创建独立的短超时上下文（`context.WithTimeout(context.Background(), 15*time.Second)`）；
  - 新增 `TestExecuteReportsLogsAndResult` 单元测试。

---

## 🟢 已修复：🟠 中高 — 排队中的 pending run 在 2 分钟后被强制判死及饱和丢任务

- **位置**：`core/app/models/runner/job.rb` & `apps/cli/internal/daemon/daemon.go`
- **问题**：
  1. `job.rb:98` 中的 `reclaim_stale` 在 `run.created_at < STALE_FIRING_AFTER.ago` 时将排队未认领的 pending run 强制标记为 "Runner offline" 失败并丢弃；
  2. Runner 饱和（`len(d.sem) >= cap(d.sem)`）时跳过轮询，执行静默长任务时因 60s 无心跳被 Core 误置为 offline。
- **修复**：
  - `job.rb`：移除 2 分钟强制失败判定，在线 Runner 的排队任务保持 `pending` 状态直至被认领或达到 `EXPIRED_AFTER`（1小时）；仅在 Runner 真正离线或 Job 被暂停时进行回收；
  - `daemon.go`：并发满载时在轮询周期内主动发送心跳 Ping（`/api/v1/runner/ping`），维持 Runner 在线活跃度；
  - `core/test/models/runner/job_test.rb` & `apps/cli/internal/daemon/daemon_test.go` 新增对应的测试用例。

---

## 🟢 已修复：🟡 低（边缘场景 / 一致性）

1. **setsid 孙进程导致 worker 槽阻塞** — `apps/cli/internal/exec/exec.go`
   - **修复**：`cmd.Wait()` 与进程组 SIGKILL 后，为 `streamLines` 的 `wg.Wait()` 增加 2 秒 drain 超时；超时后主动关闭 `stdout`/`stderr` 管道，彻底防止孤儿孙进程阻塞 worker。

2. **"超时/取消" 状态语义两端对齐** — `apps/cli/internal/exec/exec.go`
   - **修复**：超时与主动取消统一返回 `task.Error`（Core 记录为 `error/failed`）；脚本非零退出继续保留为 `task.Alerting`（Core 记录为 `alerting/succeeded` 业务告警）。

3. **日志缓冲写入顺序性** — `apps/cli/internal/daemon/daemon.go`
   - **修复**：扩大 `logChan` 缓冲容量至 1000，移除满载时并发开 goroutine 抢写的逻辑，改为受 task 上下文约束的同步入队，保证日志顺序严格递增。

4. **`RemoveJob` 与 `ListJobs` 扩展名处理** — `apps/cli/internal/workspace/workspace.go`
   - **修复**：`RemoveJob` 改用 `FindScriptFile` 支持脚本扩展名回退删除；`ListJobs` 自动剥离常见扩展名以生成规范 slug，与服务端配对保持一致。

5. **暂停 Job 的手动触发与回收语义** — `core/app/models/runner/job.rb`
   - **修复**：明确 paused job 的手动触发为合法的即时试运行；`reclaim_stale` 针对 paused job 未认领 run 进行超时优雅清理。

6. **日志 chunk 请求体大小保护** — `core/app/controllers/api/v1/runner/tasks/logs_controller.rb`
   - **修复**：增加 `MAX_CHUNK_BYTES = 512.kilobytes` 上限截断保护。

7. **push 接口空 slug 严格校验** — `core/app/controllers/api/v1/runner/jobs/pushes_controller.rb`
   - **修复**：检测到 `attrs[:slug].blank?` 时直接返回 422 `VALIDATION_ERROR`，避免静默跳过。

---

## 🟢 已优化：⚪ Nit & 文档一致性

- **Go 工具链版本对齐**：`apps/cli/go.mod` 与 `apps/cli/Dockerfile` 统一对齐到 Go 1.25，与 `mise.toml` 及本地环境保持一致。
- **Runner 活跃状态动态显示**：`tasks_controller.rb` / `results_controller.rb` / `pings_controller.rb` 检查当前是否有运行中的任务，动态呈现 `online`（执行中）与 `idle`（空闲）。
- **Jbuilder N+1 查询优化**：`runners/jobs_controller.rb` 和 `runner/jobs_controller.rb` 中加载 `@jobs` 时追加 `.includes(:runner)`。
- **示例脚本规范**：`apps/cli/examples/intranet-http` 注释统一为无扩展名路径 `~/.beep/jobs/intranet-http`。

---

---

## 🟢 第二轮深度审查与修复 (BUG、安全与边缘缺陷)

### 1. 🟢 [已修复] 🔴 BUG（高）— `Api::V1::Runners::JobsController#create` 时区解析参数倒置
- **位置**：[`core/app/controllers/api/v1/runners/jobs_controller.rb:78`](file:///home/yule/Projects/beep/core/app/controllers/api/v1/runners/jobs_controller.rb#L77-L80)
- **问题**：`IanaTimezone.resolve` 误将 `Current.user.timezone` 放首位，导致创建 Job 时用户显式选择的时区被强制覆盖为个人默认时区。
- **修复**：调整为 `IanaTimezone.resolve(params[:timezone], Current.user&.timezone)`，并在控制器测试中加测。

### 2. 🟢 [已修复] 🟠 性能 / BUG（中高）— `reclaim_stale` 运行中任务未超时分支缺少 `touch`
- **位置**：[`core/app/models/runner/job.rb:117-128`](file:///home/yule/Projects/beep/core/app/models/runner/job.rb#L117-L128)
- **问题**：running 状态的任务在 2 分钟进入 stale 扫描范围后若未超时，分支直接结束未更新 `updated_at`，导致后续每个 10s tick 反复查询该 Job。
- **修复**：在 `elsif run.running?` 的未超时分支增加 `else touch`。

### 3. 🟢 [已修复] 🟡 体验 / 一致性（中）— `Workspace#ListJobs` 返回顺序非确定性
- **位置**：[`apps/cli/internal/workspace/workspace.go:657-715`](file:///home/yule/Projects/beep/apps/cli/internal/workspace/workspace.go#L657-L715)
- **问题**：从 map 导出切片时迭代顺序随机，导致 CLI 输出顺序跳动。
- **修复**：在 `ListJobs()` 返回切片前按 `Slug` 稳定排序。

### 4. 🟢 [已修复] 🟡 缺陷（低）— `JobExecutor.Run` 在流式读取完成前提前调用 `cmd.Wait()`
- **位置**：[`apps/cli/internal/exec/exec.go:54-73`](file:///home/yule/Projects/beep/apps/cli/internal/exec/exec.go#L54-L73)
- **问题**：提前调用 `cmd.Wait()` 会过早关闭 pipe 读端，可能截断丢失进程末尾输出的几行日志。
- **修复**：等待 stdout/stderr 流读取完毕后再收割进程退出。

### 5. 🟢 [已修复] 🟡 安全兼容（低）— `StopDaemon` 强制 KILL `-pid` 对 PID 1 的影响
- **位置**：[`apps/cli/internal/daemon/socket.go:108`](file:///home/yule/Projects/beep/apps/cli/internal/daemon/socket.go#L108) & [`apps/cli/internal/exec/exec.go:57`](file:///home/yule/Projects/beep/apps/cli/internal/exec/exec.go#L57)
- **问题**：容器环境中以 PID 1 运行时，向 `-1` 广播 `SIGKILL` 会杀掉容器内其它所有进程。
- **修复**：限制只有当 `pid > 1` 时才执行负进程组 kill。

### 6. 🟢 [已修复] 🟡 数据截断（低）— 日志超长边界 UTF-8 字符截断
- **位置**：[`core/app/models/runner/run.rb:35`](file:///home/yule/Projects/beep/core/app/models/runner/run.rb#L35) & [`core/app/controllers/api/v1/runner/tasks/logs_controller.rb:22`](file:///home/yule/Projects/beep/core/app/controllers/api/v1/runner/tasks/logs_controller.rb#L22)
- **修复**：使用 `truncate_bytes(MAX_CHUNK_BYTES, omission: "")` 与 `byteslice(overflow..-1)&.scrub("")`，防止 UTF-8 多字节字符在截断边界损坏。

### 7. 🟢 [已修复] ⚪ 前端竞态（建议）— Web 详情页轮询在快速切换 Job 时的异步状态覆写
- **位置**：[`apps/web/src/routes/$account_slug/runners_.$runnerId.tsx:135-156`](file:///home/yule/Projects/beep/apps/web/src/routes/$account_slug/runners_.$runnerId.tsx#L135-L156)
- **修复**：在 `useEffect` cleanup 中增加 active 标志，忽略已失效旧请求的响应。

### 8. 🟢 [已修复] 🔴 BUG（高）— `job remove` 把过期 `@id` 的 404 当成删除成功
- **位置**：[`apps/cli/internal/client/client.go:120-136`](file:///home/yule/Projects/beep/apps/cli/internal/client/client.go#L120-L136)、[`apps/cli/cmd/job.go:222-247`](file:///home/yule/Projects/beep/apps/cli/cmd/job.go#L222-L247)
- **修复**：`DeleteJob` 对 404 正常返回错误；CLI 删除时优先尝试 `@id`，若失败则自动回退按 `slug` 尝试删除并准确报告结果。

### 9. 🟢 [已修复] 🔴 BUG（高）— `beep runner up -d` 在握手成功前就报启动成功
- **位置**：[`apps/cli/cmd/up.go:70-144`](file:///home/yule/Projects/beep/apps/cli/cmd/up.go#L70-L144)、[`apps/cli/internal/daemon/daemon.go:39-50`](file:///home/yule/Projects/beep/apps/cli/internal/daemon/daemon.go#L39-L50)
- **修复**：Socket 初始置为 `starting` 状态，待首次 `Ping` 握手成功后置为 `running`；后台启动等待 `CheckReady`，并在握手失败或子进程过早退出时返回非零错误码并输出日志路径。

### 10. 🟢 [已修复] 🟠 安全（中）— HTTP 跨主机跳转会带上 `X-Runner-Token`
- **位置**：[`apps/cli/internal/client/client.go:25-31`](file:///home/yule/Projects/beep/apps/cli/internal/client/client.go#L25-L31)
- **修复**：在 `http.Client.CheckRedirect` 中检测跨 host 跳转并主动剔除 `X-Runner-Token`，并添加回归测试用例。

### 11. 🟢 [已修复] 🟠 安全（中）— `beep runner stop` 信任 socket 里未认证的 PID
- **位置**：[`apps/cli/internal/daemon/socket.go:40-116`](file:///home/yule/Projects/beep/apps/cli/internal/daemon/socket.go#L40-L116)
- **修复**：Workspace 目录权限收敛为 `0700`；通过 `SO_PEERCRED` 校验 Unix domain socket 对端的 UID 与 PID 是否一致。

### 12. 🟢 [已修复] 🟡 BUG（中低）— `PairJobs` 用 slug 兜底会吞掉过期 `@id`
- **位置**：[`apps/cli/internal/ui/interactive.go:337-414`](file:///home/yule/Projects/beep/apps/cli/internal/ui/interactive.go#L337-L414)、[`apps/cli/internal/ui/interactive.go:462-478`](file:///home/yule/Projects/beep/apps/cli/internal/ui/interactive.go#L462-L478)
- **修复**：`CompareJob` 增加 ID 差异比较；Pass 2 严格跳过 `ID != ""` 的项，并添加回归测试用例。

---

## 结论与验证

所有两轮审查发现的问题均已修复并通过自动化测试套件验证：

- **Go CLI** (`apps/cli`): `go test ./...`、`go vet ./...`、`go build ./...` 全部通过（包含新增测试用例）。
- **Rails Core** (`core`): `bin/rails test` 381 runs, 1329 assertions 全部通过（0 failures, 0 errors）。
- **Web** (`apps/web`): `biome check` 140 files 全部通过。
