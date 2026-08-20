# Todo — 所有 API 调用改走 TanStack Start Node 服务端 (Mode B only)

- **Branch**: `feat/all-api-via-node-server`
- **Date**: 2026-08-20
- **Goal**: 浏览器不再直连 Rails core；读/写都从 TanStack Start 的 Node 服务端发往 core，并开启真 SSR。部署只支持 Mode B（web 单源），废弃 Mode A（spin 域名）。

## 背景

原始实现是 Mode A + 客户端直连 core：浏览器 `fetch("/api/v1/...")` 带 `credentials: include` 跨域打 core；session cookie 常驻 core 域；页面用 `ssr: false` 客户端渲染（因为 SSR 读不到 cookie，硬做会把 guest 脱水给鉴权锁死）。

目标：所有 `/api/v1` 调用改经 Node 发起（读走 SSR loader、写走 `createServerFn`），页面真服务端渲染。

## 已完成

- [x] 新增服务端 HTTP 层 `src/server/core.ts`
  - 从请求转发 `Cookie` 头（`getRequestHeader`），服务端到服务端，无 CORS。
  - 暴露 `coreFetch` / `coreFetchWithHeaders`；非 2xx 抛 `ApiError`（复用 `lib/api/client.ts` 的定义）。
- [x] 写操作与读操作全部包成 `createServerFn`
  - `server/session.ts`：`startSession` / `verifyMagicLink` / `destroySession` / `rememberLastAccount` / `fetchMe`；会话接口用 `forwardCookies`（`setCookie`）回传 core 的 `Set-Cookie`，让 cookie 落在 web 域。
  - `server/beeps.ts`、`settings.ts`、`push.ts`、`dev.ts`、`admin.ts`
- [x] `lib/api/*` 收口为薄封装，改调 server fn；`lib/api/client.ts` 瘦身为只剩 `ApiError`。
- [x] 鉴权改服务端可用
  - `__root.tsx` beforeLoad：SSR 也拉 `me`（不再硬编码 `null`）。
  - `guards.ts`：`requireSession`/`requireGuest`/`probeSession` 在服务端直接跑，未登录 302 `/sign`。
  - 移除 `$account_slug` / `accounts` / `admin` / `dev` / `sign` 的 `ssr: false` → 真 SSR。
- [x] 修复：`fetchMe` 恢复 SSR 分支跳过模块级缓存，避免多租户身份串号（`lib/api/session.ts`）。
- [x] 移除 `vite.config.ts` 里过时的 `importProtection`（client.ts 不再动态 import server）。
- [x] 文档：`docs/core/DEVELOP.md` 鉴权/部署段更新为 Mode B-only。
- [x] 验证：`tsc --noEmit` ✅、`biome check` ✅、`pnpm build` ✅；内置产物实测 `/` SSR 200，标题在服务端 HTML 中。

## 待办

### 高优先（架构清理）

- [ ] 合并两层重复结构：`lib/api/*` 与 `src/server/*` 现在类型/转发逻辑重复。收敛为单层，避免"同名转发"双份维护
  - 方案 A：删 `lib/api/*` 转发，组件/路由直接用 `server/*` 导出（改动组件导入较多）
  - 方案 B：删 `server/` 目录，直接在 `lib/api/*` 内写 `createServerFn`（推荐，改动集中）
- [ ] 复核 `Nitro /api` 代理是否移除：浏览器已不调用 `/api/v1`，代理属死配置；保留仅作 `/up` 健康检查与兜底。若移除需确认 compose/deploy 健康检查不受影响。

### 中优先（部署/文档一致性）

- [ ] core 侧 `development_cors.rb`：浏览器直连 CORS 路径已不再被使用，评估移除
- [ ] 清理其余文档的 Mode A/split 描述：`README.md`、`docs/architecture/web-push.md`、`docs/deploy-to-dokploy.md`、`docs/core/STYLE.md`、`.agents/web.md`
- [ ] 复核 `compose.yml` / `compose.dokploy.yml`：`SESSION_COOKIE_DOMAIN` 需匹配 web 源，保证 SSR/服务端函数能读到 cookie；`VITE_CORE_URL` 仅剩外链用途

### 低优先 / 待确认

- [ ] 分支中混入无关改动 `core/db/schema.rb`（admin `notification_channels` 列顺序 diff，非本次工作产生）——提交时排除或单独确认
- [ ] 提交：确认 commit message 遵循仓库格式 `emoji [scope] The main change`，无 agent trailer
