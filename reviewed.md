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

## 已核对、未发现问题的部分

- **认证与租户隔离**：runner agent API 仅凭 `X-Runner-Token` 认证并按 runner 作用域查询；Web 端所有 runners/jobs/runs 控制器均经 `Current.account` 作用域。
- **Token 安全**：`blockedJobEnvKeys` 严格清洗子进程环境变量，Runner Token 不会泄露给脚本进程。
- **回调 URL 校验**：CLI 对 log/result URL 做白名单校验，Core 侧回调 base URL 优先采用配置域名。
- **Slug 防路径穿越**：`ValidateSlug` 严格校验，防止目录遍历。
- **并发正确性**：状态流转使用原子条件更新，唯一索引保证 poller 实例安全调度。

---

## 结论

所有高、中、低风险项与细节问题均已修复并通过单元测试与全量测试套件验证。
