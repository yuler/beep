# Theme

> Beep 页面映射：`/`（介绍首页）、`/sign`（邮箱魔法链接与验证码登录）、`/$account_slug`（控制台概览）、`/$account_slug/beeps`（任务监控）、`/$account_slug/beepers`（探测探针）、`/$account_slug/runners`（执行器）、`/$account_slug/settings`（通用 / 渠道 / 通知）。顶栏左边 SVG 铃铛 Logo（`LogoMark`）+ `Beep`，右边按登录态与当前区域组织。

界面采用移动端优先（Mobile First）。支持中英双语（保留 `LocaleSwitcher`）。颜色、字号和组件的名字用英文。

**圆角规范**：所有带边框或底色的盒子统一采用 **`2px` 小圆角**（`--radius: 2px` / `rounded-sm`）；状态指示点、头像与进度条使用正圆（`rounded-full`）；紧贴屏幕边缘的全宽结构（顶栏、手机底部抽屉、通栏底边输入框）保持直角（`0px`）。

**色彩哲学**：界面骨架保持克制的黑白灰单色；不设独立品牌强调色（主操作、导航激活、焦点描边仍使用 `ink` / `solid`）。彩色**仅用于表达结果与健康状态**：绿色 `success`（成功 / 在线 / 健康）、黄色 `warning`（警告 / 降级）、红色 `danger`（失败 / 告警 / 危险操作）。生命周期状态（`active`、`paused`、`completed`）保持单色，`firing`（执行中）使用单色脉冲圆点表达。

## Color

### Monochrome Skeleton

| Token         | 黑 (默认) | 白        | 用在                             |
| ------------- | --------- | --------- | -------------------------------- |
| `ground`      | `#0a0a0a` | `#ffffff` | 所有页面的底、弹窗底色           |
| `ink`         | `#fafafa` | `#0a0a0a` | 主文字、激活高亮、焦点描边       |
| `gray`        | `#a1a1a1` | `#666666` | Meta、次要说明、已填/次要边框    |
| `gray-subtle` | `#6b6b6b` | `#a3a3a3` | 占位符、离线圆点、弱化辅助信息   |
| `line`        | `#262626` | `#e5e5e5` | 分割线、幽灵按钮边框、输入框边框 |
| `wash`        | `#171717` | `#f5f5f5` | 浅底块、按钮/菜单悬停与按下反馈  |
| `solid`       | `#fafafa` | `#0a0a0a` | 实心按钮的底、选中态背景         |
| `solid-ink`   | `#0a0a0a` | `#ffffff` | 实心按钮的字、选中态文字         |

### Semantic Status

| Token          | 黑 (默认)              | 白                     | 用在                                               |
| -------------- | ---------------------- | ---------------------- | -------------------------------------------------- |
| `success`      | `#3dd68c`              | `#16a34a`              | 成功状态文字、在线/健康圆点、趋势图成功曲线        |
| `warning`      | `#f5c542`              | `#b7791f`              | 警告状态文字、降级提示、Notice 警告竖线            |
| `danger`       | `#ff5f57`              | `#dc2626`              | 失败/告警文字、表单校验错误、危险操作按钮、失败曲线 |
| `danger-ink`   | `#0a0a0a`              | `#ffffff`              | 实心危险按钮（`destructive`）上的文字              |
| `success-soft` | `success` 12% 透明混合 | `success` 12% 透明混合 | 成功 `StatusPill` / `Badge` 的浅底色               |
| `warning-soft` | `warning` 12% 透明混合 | `warning` 12% 透明混合 | 警告 `StatusPill` / `Badge` 的浅底色               |
| `danger-soft`  | `danger` 12% 透明混合  | `danger` 12% 透明混合  | 失败/告警 `StatusPill` / `Badge` 的浅底色          |

默认是黑（`dark`）。顶栏右上始终挂 `ThemeToggle` 下拉菜单，支持 **深色 (`dark`) / 浅色 (`light`) / 跟随系统 (`system`)** 三档切换，选择保存在 `localStorage`（键名 `theme`）。

## Type

全站一种无衬线，中文优先：`"Noto Sans CJK SC"`, `"Source Han Sans SC"`, `"Noto Sans SC"`, `sans-serif`。

字重只用两档：正文 400，标题 500。

| Role          | Size               | Weight | Line height | 用在                         |
| ------------- | ------------------ | ------ | ----------- | ---------------------------- |
| `Kicker`      | 13px               | 400    | 1.4         | 导航、标签、App 名。字距 `0.18em` |
| `Title`       | 28px               | 500    | 1.2         | 页面主标题                   |
| `AppTitle`    | 22px               | 500    | 1.2         | 分区标题                     |
| `Meta`        | 14px               | 400    | 1.45        | 说明文字、表格次要信息       |
| `ButtonLabel` | 继承所在表面的正文 | 400    | 1.2         | 按钮字                       |

## Components

### `Button`

- `ButtonSolid`：底 `solid`，字 `solid-ink`，边框 1px `solid`，2px 圆角，内边距 10px 14px。
- `ButtonGhost`：透明底，1px `line` 边框，2px 圆角，内边距 8px 12px。悬停边框变 `ink`，按下 `wash` 底。选中态改为 `solid` 底、`solid-ink` 字。
- `Button` (`variant="destructive"`)：底 `danger`，字 `danger-ink`，边框 1px `danger`，2px 圆角，用于删除、撤销等不可逆确认操作。

### `TextField`

无底色，只有底边 1px `line`（复杂表单为 1px `line` 四周边框 + 2px 圆角）。上下内边距 8px。聚焦时 1px `ink` 外偏 2px 轮廓。校验出错（`aria-invalid="true"`）时底边/边框与焦点轮廓变为 `danger`，下方错误文案使用 `danger`。

### `StatusPill` & `StatusDot`

- **结果与健康状态**：使用 `success` / `warning` / `danger` 三档，样式为 `*-soft` 浅底 + 同色文字，2px 圆角。
- **生命周期状态**：`active` 用 `outline`（1px `line` 边框 + `ink` 字），`paused` / `completed` 用 `muted`（1px `line` 边框 + `gray` 字），`firing`（执行中）用 `live`（1px `ink` 边框 + 脉冲圆点 `StatusDot`）。
- **在线指示点 (`StatusDot`)**：正圆 `6px`，在线为 `success` 绿点，离线为 `gray-subtle` 灰点。

### `AppBar`

顶栏极简。
- 公开页首页：不放顶栏，只在右上浮一层 `LocaleSwitcher` + `ThemeToggle`。登录页：左侧返回首页；右侧 `LocaleSwitcher` + `ThemeToggle`。
- 控制台内：左侧 `LogoMark` + `Beep` 字标 + 工作区选择器（`AccountSwitcher`）；右侧为 `LocaleSwitcher` + `ThemeToggle` + 用户菜单（个人偏好、API Tokens、Admin、Dev、退出）。

### `Subnav`

紧贴 `AppBar` 下方。采用 `Kicker` 风格字号与字距。
- 账号工作区：`概览`、`Beeps`、`Beepers`、`Runners`、`设置`。
- 管理后台：`← 工作区`、`任务`、`统计`。
- 个人中心：`← 工作区`、`Tokens`、`偏好`。
手机端支持横向滑动（`overflow-x-auto`），吸顶展示。当前激活项使用 `ink` 底部高亮边框。

### `Timeline`

时间线与紧凑行列表。每行上边框 1px `line`，上下内边距 10px。角色/标签前置，时间当天显示 `HH:mm`，跨天带日期。重要变动左侧加 1px `line` 竖线。

### `Notice`

一行状态或提示。无背景底色，左侧竖线 + 前置小号 `Kicker` 标签：
- `default`：1px `line` 竖线，`gray` 标签。
- `success` / `warning` / `danger`（或 `error`）：2px 对应语义色竖线与同色标签。

### `Panel` & `Card`

分区容器，不填底色（或 `ground` 底）。
- 手机端（<768px）：仅保留上边 1px `line` 分隔，直角。
- 平板与电脑端（≥768px）：四边 1px `line` 空心方框，2px 圆角。
`wash` 只留给悬停或按下态反馈。

### `OtpField`

6 个等宽方格，2px 圆角，1px `line` 边，当前聚焦格 `ink` 边，已填格 `gray` 边。底层为透明的真实输入框，覆盖在方格上。自动填充、粘贴、退格走原生行为。填满 6 位自动提交；出错（`aria-invalid`）时格子变 `danger` 红边与红字。

### `Segmented`

视图与选项切换（如设置页的 通用 / 渠道 / 通知）。多枚等宽 `ButtonGhost`，选中项为 `solid` 底、`solid-ink` 字。

### `Dialog / Drawer`

弹层与模态操作：
- 手机端（<768px）：底部抽屉，直角，仅顶部 1px `line` 边框。
- 平板与电脑端（≥768px）：居中方框，四边 1px `line`，2px 圆角，无投影。
遮罩层采用 `ground` 配合 80% 半透明。

### `Table / List`

- 手机端（<1024px）：采用 `Timeline` 风格的多行列表，主次分明，减少横向滚动。
- 电脑端（≥1024px）：采用 1px `line` 网格的表格，外框 2px 圆角。

### `Homepage Story Illustrations`

首页通过「服务异常告警与恢复（4 幕）」与「自然语言准时提醒（2 幕）」两组叙事插画传达产品价值：
- 采用 `1.5px` 单色线稿（`currentColor` 跟随 `ink` / `gray`）+ 2px 圆角卡片构件，自动适配深浅主题。
- 仅在叙事关键节点点缀 `danger`（宕机/告警）、`warning`（延迟/即将到期）、`success`（恢复/准时送达），让语义色成为视觉叙事焦点。

## Layout

移动优先。断点：`md` 768px、`lg` 1024px。

|                | 手机 (<768)                | 平板 (768–1023)              | 电脑 (≥1024)   |
| -------------- | -------------------------- | ---------------------------- | -------------- |
| 页边距         | 16px                       | 24px                         | 32px           |
| 顶栏高         | 56px                       | 64px                         | 64px           |
| `Panel`        | 无底色，只留上边 1px `line` | 四边 1px `line` 空心框 (2px) | 同左           |
| 表单主按钮     | 通栏                       | 自然宽度                     | 自然宽度       |
| 任务与资源列表 | 行列表展示                 | 行列表展示                   | 数据表格展示   |

内容最大宽：首页、控制台 1080px，登录 440px。顶栏通栏，内容与内容区对齐。适配 `safe-area-inset-*`。

## Interaction

- 点击区至少 44px（电脑端 40px）。
- 悬停效果仅在 `(hover: hover)` 设备上生效；触屏用 `:active` 给按下反馈。
- 键盘焦点统一 1px `ink` 描边、外偏 2px（校验错误时为 `danger`）。
- 有异步动作的按钮进入 `is-loading`：圆形转圈 + 「…中」文案，并禁用。
- 不可逆操作（断开绑定、删除探针、撤销 Token 等）采用确认弹窗，确认按钮使用 `destructive` 红色样式。
- 输入框按语义设 `inputMode` / `enterKeyHint` / `autoComplete`；手机端字号 ≥16px 防缩放。
- 系统 `theme-color` 跟随 `ground`（深色 `#0a0a0a`，浅色 `#ffffff`）。
- `prefers-reduced-motion` 时关掉过渡与动画。
