# 死代码 / 冗余扫描报告（2026-09-17）

> 全仓 grep / read 验证过引用关系。误报 1 个已剔除：`.env.dokploy:38` 拼写正确，无 typo。
> 约定：每条为 `文件:行号 — 原因 — 建议`。

## 高优先级：可直接删

### core（Rails）

- `core/config/routes.rb:4` — `resource :landing` 要找 `LandingController`，实际只有 `LandingsController`，且 `root landings#show` 已覆盖 — 删这行，`account/join/show.html.erb:20` 的 `landing_path` 一并改掉
- `core/config/routes.rb:5` — `resources :home` 生成 7 路由，`home_controller.rb:5` 只有 `index` — 改为 `only: %i[index]`
- `core/config/routes.rb:17` — `my/accounts` 默认 7 路由，controller 只有 `new/create/index` — 加 `only: %i[new create index]`
- `core/config/routes.rb:35` — 空 `resource :subscription` 无对应 controller，块内只有注释 + TODO — 删
- `core/lib/github_client.rb:1` — `GithubClient` 全仓零引用 — 删
- `core/app/models/account/payable/wechat.rb:1` — 空类，零引用 — 删
- `core/app/models/account/payable/stripe.rb:1` — 全文件只有注释 + TODO，零引用 — 删
- `core/app/models/account/payable/creem.rb:28` — `create_subscription(body)` 空方法体，零调用 — 删
- `core/app/helpers/api/v1/users_helper.rb:1` — 空 module，零引用 — 删
- `core/app/helpers/landings_helper.rb:1` — 空 module，零引用 — 删
- `core/app/controllers/concerns/authentication.rb:26` — `alias_method :skip_authentication` 零调用 — 删
- `core/app/controllers/concerns/api/v1/responses.rb:8` — `render_json_created` 零调用 — 删
- `core/app/controllers/concerns/api/v1/responses.rb:30` — 通用 `render_json(json:, status:)` 零调用 — 删
- `core/app/views/api/v1/cli/authorizations/device_flow_error.json.jbuilder:1` — 与 `channels/cli/...`、`runners/...` 三份同名文件内容完全重复 — 合并为 `shared/device_flow_error` partial

### web（TanStack Router）

- `apps/web/src/components/beautiful-ui/chat-widget.tsx:57` — `ChatWidget` 无路由挂载 — 整套删
- `apps/web/src/components/beautiful-ui/chat-composer.tsx:47` — 唯一调用者是已死的 `ChatWidget` — 随上一起删
- `apps/web/src/styles/beautiful-ui.css:1` — 整个 `.bui` 主题仅服务上面两个死组件，却被 `styles.css:9` 全局 import — 随上一起删
- `apps/web/src/components/beeps/create-beep-form.tsx:34` — 零 import，与 `beep-quick-create.tsx:48 BeepQuickCreate` 重复实现（`TITLE/BODY_MAX_LENGTH`、`CRON_PRESETS`、`defaultRunAt`） — 删前者
- `apps/web/src/lib/mock/ai-chat.ts:8,10,12` — `mockChatTabs/mockStarterMessage/mockChatReplies` 零 import，`chat-composer` 自己内联了 mock — 删文件或让 composer 复用它
- `apps/web/src/lib/theme.ts:11,15,21` — `getThemePreference/getSystemTheme/getResolvedTheme` 仅文件内自引用，外部只用 `getStoredTheme + toggleTheme` — 收敛为内部函数或删
- `apps/web/src/lib/theme.ts:34,40,45` — `applyTheme/setTheme/setThemePreference` 外部零引用 — 若不做主题设置页则删
- `apps/web/src/lib/locale.ts:40,48,52` — `parseMessageJson/getMessageList/getCanonicalPathname` 零引用 — 删
- `apps/web/src/lib/run-success-rate.ts:9` — `runSuccessRate` 零外部引用 — 删
- `apps/web/src/lib/api/beeps.ts:42` vs `src/lib/api/beepers.ts:99` — `PaginationMeta` 逐字段完全重复 — 抽到 `lib/api/pagination.ts`
- `apps/web/package.json:36` — `shadcn@^4.16.1` 放 `dependencies`，实际只用到 `styles.css:7` 的 CSS import — 移到 `devDependencies`

### cli / 根目录

- `apps/cli/internal/ui/color.go:103` — `StatusBadge` 零调用 — 删，或在 beep/beeper/runner 状态输出中统一使用
- `apps/cli/internal/ui/color.go:67` — `White` 零调用（`Gray` 有真实调用） — 删
- `apps/cli/internal/ui/table.go:51` — `Table.SetPadding` 零调用（调用方只用 `SetIndent`） — 删
- `apps/cli/internal/exec/hook.go:19` — `FindHook` 零外部调用，只是 `FindHookForEvent(ws, "")` 薄包装 — 删
- `apps/cli/internal/exec/hook.go:23` — `FindHookForEvent(ws, _ string)` 第二参数直接丢弃，`DispatchHook:55` 传 `eventName` 也被忽略 — 删参数或实现按事件分发，同步修 `hook_test.go:62-76`
- `apps/cli/cmd/auth.go:29` — `ensureLoggedIn` 零生产调用（都直调 `cmdutil.EnsureLoggedIn`） — 删
- `pnpm-workspace.yaml:2` — `packages/*` 通配但仓库无 `packages/` 目录（`glob packages/*` 为空，`AGENTS.md` 的共享包不存在） — 删这行
- `TODO.md:22` — `官方 seed apps/plugins/*/manifest.json` 路径已死，实际为 `beeper_apps/*/manifest.json` — 改路径
- `TODO.md:52` — `Go Runner 客户端核心（apps/runner）` 路径已死，实际为 `apps/cli` 单二进制 — 改路径
- `TODO.md:73` — `server 运维：up/stop/status daemon` 描述已死，实际为 `service/channel/runner` 三组命令 — 重写该条
- `TODO.md:78` — `当前 root.go 未注册 completion` 已过时，`root.go:150` 已调 `initCompletionCmd()` — 勾选完成或改写

## 中优先级：合并 / 确认后删

- `core/app/models/membership.rb:1` — 生产代码无表无关联引用（`User` 已取代其角色），仅 `test/models/membership_test.rb:3` 孤立存在 — 确认后删
- `core/app/models/invite_code.rb:1` — 生产代码零引用（仅 test 自测），实际邀请走 `Account::JoinCode + Invitation`，疑似被取代的 legacy — 确认后删
- `core/app/helpers/forms_helper.rb:2` — `auto_submit_form_with` 零调用 — 确认后删
- `core/app/controllers/concerns/authentication.rb:49` vs `api_authentication.rb` — 约 40 行 session/cookie 逻辑双份实现 — 抽共享 concern
- `core/app/controllers/api/v1/cli/authorizations_controller.rb:1` — 与 `channels/cli/authorizations`、`runners/authorizations` 三套 device-flow + token 控制器同构重复 — 抽基类/concern
- `core/config/routes.rb:64` — `put "me/last_account"` 非 RESTful，违反 `docs/core/STYLE.md` — 改为新 resource
- `core/config/routes.rb:181` — `delete :clear, on: :collection` 非 RESTful 自定义动作 — 改为新 resource
- `core/app/models/account/payable/plan.rb:25` — `Plan.free/free?/paid?` 中 `free` 引用的 `:free` key 在 `PLANS` 中被注释，全仓零调用（仅 `Plan.paid` 被 payments 用作 fallback） — 确认后删
- `apps/web/src/components/beeps/beep-list.tsx:1` vs `beepers/beeper-list.tsx:1` — 各约 500-600 行 `DataTable/makeSelectColumn/SortableHeader/StatusPill+ProgressBar/分页` 模式重复 — 抽公共 `PaginatedDataTable` / `RunProgress`
- `apps/web/src/lib/api/client.ts:114` — `apiFetchWithHeaders` 唯一调用在 `session.ts:42` — 合并为 `apiFetch` 的 `includeHeaders` 选项
- `apps/web/src/lib/beep-datetime.ts:24` vs `beep-stats.ts:3` — `Intl.DateTimeFormat + try/catch 回退 UTC` 重复 — 抽 `formatInTimezone`
- `apps/web/src/components/beeps/beep-status.ts:17` vs `lib/i18n-labels.ts:17` — 状态图标/颜色与文案双源并存，新增状态需改两处 — 合并为单一 `beepStatusMeta`
- `apps/web/src/lib/build-info.ts:25` — `logBuildInfo()` 内 `console.log` 常驻生产 bundle — 加 `import.meta.env.DEV` 守卫
- `apps/web/src/lib/auth/slugs.ts:43` — `RESERVED_SLUGS` 从未被外部 import，仅内部 `isAccountSlug` 使用 — 去掉 `export`
- `apps/web/src/routes/device/channel.tsx:1` vs `runner.tsx/cli.tsx` — 三件套 `requireSession + resolveShellAccount` 重定向逻辑雷同 — 抽公共 `beforeLoad`
- `apps/cli/cmd/channel.go:7` / `apps/cli/cmd/config.go:7` — 7 行纯 `var xCmd = subpkg.NewCmdX()` 重导出 — 内联到 `root.go` 或删中转文件
- `apps/cli/cmd/root.go:109-131` — `beep list` 与 `beep beep list` 双注册，`cmd/beep/*` 与 `cmd/beeper/*` 大面积重复 — 选一为主，另一标 `Deprecated` + 隐藏
- `compose.yml` 全文件 — 被 `.gitignore:71` 标记为本地生成副本却被提交，且 `SESSION_COOKIE_DOMAIN` 硬编码、`image::main` 写死 — 删追踪或与 `compose.example.yml` 重同步
- `beeper_apps/heartbeat|site-uptime|ssl-expiry/manifest.json:12` — `alerting + author + manifest_version` 三份逐字重复 — 抽 schema defaults 或模板生成

## 低优先级：保留 / 仅记录

- `core/config/routes.rb:186` — `namespace :test`（`public/private`）常驻生产路由表 — 确认后约束为 `Rails.env.local?/test?`
- `core/config/routes.rb:34`、`Gemfile:50` — `# TODO: subscription / stripe` 注释残留，与已选 `Creem` 路线冲突 — 确认后删注释
- `apps/web/package.json:34` — `react-dom` 零直接 import，属 `@tanstack/react-start` 间接 peer，保留勿删
- `apps/web/src/lib/auth/guards.ts:12,28` — `redirectForTarget/redirectToSign` 仅内部使用，属正常分层，保留
- `apps/web/src` 全仓 `TODO/FIXME/HACK/XXX` 及大段注释掉的代码零命中 — 卫生良好，无需处理
- `DESIGN.md:1` — 通篇 Rails vanilla CSS 设计系统，与当前 TanStack 主 UI 无链接 — 顶部加适用范围声明或归档
- `docs/superpowers/plans/2026-08-27-beeper-split-from-beep.md` — 已落地计划残留，引用 `apps/plugins` / `Plugin::*` / `apps/beepers` 皆与现状 `beeper_apps/BeeperApp` 不符，且 `.gitignore:87` 已忽略该目录 — 归档或移出仓库

## 验证方式

- `grep GithubClient|Receivers::Echo|Payable::Wechat|Payable::Stripe|skip_authentication|render_json_created|ChatWidget|CreateBeepForm|mockChatTabs|StatusBadge|FindHook|ensureLoggedIn` 全仓查零引用
- `read core/config/routes.rb`、`landings_controller.rb`、`home_controller.rb`、`responses.rb`、`payable/stripe.rb`、`payable/wechat.rb`、`github_client.rb`、`color.go`、`pnpm-workspace.yaml`、`TODO.md`、`.env.dokploy` 逐项核对
- `glob packages/*` 为空；`glob core/app/controllers/*landing*` 仅复数；`grep DEEPSEEK` 确认无拼写 typo
