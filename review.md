# review

PR: [#38](https://github.com/yuler/beep/pull/38) `feat/self-hosted-runner` vs `main`  
Spec: `docs/architecture/runner.md`  
Scope: bugs + design defects (not style nits)

---

## Fixed (this session)

- [x] #1 **Token modal undefined `token`**：修复前端创建 Runner 弹窗中 `token` 字段未解构导致复制和展示为 `undefined` 的问题。
- [x] #2 **`PullJob` rename collision wipe**：修复 CLI 拉取任务时重命名冲突导致本地非目标脚本被误清空覆盖的问题。
- [x] #3 **Job env inherits `BEEP_RUNNER_TOKEN`**：修复执行用户自定义脚本时子进程意外继承 runner 敏感 Token 的环境变量泄漏问题。
- [x] #4 **Reclaim can fail just-claimed run (`from_statuses: pending`)**：修复超时回收任务时并发状态判断缺陷导致刚领取的任务被误标记失败的问题。
- [x] #5 **Atomic `record_result!`**：优化任务执行结果写入为数据库原子事务与乐观锁更新，避免并发冲突覆盖结果。
- [x] #6 **Pause orphans reclaim**：修复 Runner 或任务暂停时未正确释放或清理后台回收调度导致的孤儿重试任务。
- [x] #7 **Log/result 422 treated as success**：修复 CLI 客户端在服务端返回 422 校验失败时仍误判为成功并继续执行的问题。
- [x] #8 **Process group kill on job timeout/stop**：修复超时或取消任务时仅杀死主进程而残留子进程/孙进程的问题（改用进程组信号 `SIGKILL`）。
- [x] #9 **`up -d` orphan child on socket timeout**：修复守护进程后台启动等待 socket 响应超时导致生成孤儿进程的问题。
- [x] #10 **In-flight jobs ignore daemon cancel ctx**：修复守护进程收到退出信号时正在执行的后台任务忽略上下文取消导致挂起的问题。
- [x] #11 **`job remove` after rename (delete by `@id`)**：修复本地重命名任务后执行 `job remove` 因找不到旧 slug 导致删除失败的问题（改为优先按 `@id` 唯一标识删除）。
- [x] #13 **Core ignores per-job `timeout_seconds` for reclaim**：修复服务端统一写死 300s 超时的问题，改为根据每个任务自定义的 `claimed_at + timeout_seconds` 动态计算超时回收时间。
- [x] #14 **Timezone edits refresh `next_run_at`**：修复在 Web 端修改 Runner 任务的时区设置后未重新计算并刷新下一次运行时间 (`next_run_at`) 的问题。
- [x] #15 **Callback URLs from configured host**：修复 Webhook/Callback 回调地址从请求上下文中的 `Host` 获取导致内网或反代场景 URL 错误的问题（统一走配置的标准 Host）。
- [x] #16 **Spec: scripts must not use runner token for callbacks**：完善架构规范与权限隔离，明确禁止脚本执行环境复用 runner 认证 Token 进行回调。
- [x] #17 **Spec pairing order → `@id` then slug**：统一 CLI 与 Core 架构规范中的任务配对解析顺序为优先匹配 `@id`、后降级匹配 `slug`。
- [x] #18 **Web job create/edit hints & slug rename warning**：在前端任务创建/编辑界面增加表单输入提示及重命名 slug 潜在影响的告警说明。
- [x] #19 **Offline banner / warning in runner detail UI**：在前端 Runner 详情页增加离线状态 Banner 提示与排查引导。
- [x] #20 **Surface `serverErr` in `push` & check `@id` write errors**：改进 CLI `push` 命令，捕获并向终端显式透出服务端具体错误，增加 `@id` 本地回写校验。
- [x] #22 **Stale online badge on detail page**：修复 Runner 详情页在节点长时间未心跳时仍展示绿色 Online 徽章的问题（在 Show Action 增加即时 `mark_stale_offline` 检测）。
- [x] **N+1 query in `RunnersController#index` eliminated**：通过聚合 `jobs_count` 预加载消除 Runner 列表页因统计任务数量引发的 N+1 查询。
- [x] **CLI `daemon.go` log upload improved to non-blocking async buffer**：将 CLI 执行日志上传优化为非阻塞异步缓冲队列，避免网络抖动卡顿任务执行。
- [x] **CLI `workspace.go` added script extension fallback**：完善 CLI 本地脚本查找逻辑，找不到同名命令时自动匹配 `.sh`、`.py`、`.js` 等常见扩展名。
- [x] **Blank cron on sync no longer silently defaults**：修复同步时若显式传入空 cron 会被静默填充默认值的问题，明确报错或保持空。
- [x] **Purge dead CLI probe package (`internal/probe`)**：彻底移除早期遗留的未使用的 probe 原型包（7 个文件，~567 行代码与测试）。
- [x] **Deduplicate Core runner controllers and helpers**：抽取 `find_or_build_job`、`job_attrs_from` 与 `set_run` 到 `Api::V1::Runner::BaseController`，消除多处重复；精简 `RunnersController` 与 `RunnerJobsController` 中冗余的 `update_params`。
- [x] **Unify runner job Jbuilder partials & web client types**：复用 `runner_jobs/_job.json.jbuilder` 视图局部模板，简化 Web 端 `runner-jobs.ts` 的响应类型。

