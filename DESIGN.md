# Theme

> Beep 页面映射：`/`（介绍首页）、`/sign`（邮箱魔法链接与验证码登录）、`/$account_slug`（控制台概览）、`/$account_slug/beeps`（任务监控）、`/$account_slug/beepers`（探测探针）、`/$account_slug/runners`（执行器）、`/$account_slug/settings`（通用 / 渠道 / 通知）。顶栏左边 SVG 铃铛 Logo（`LogoMark`）+ `Beep`，右边按登录态与当前区域组织。

界面采用移动端优先（Mobile First）。支持中英双语（保留 `LocaleSwitcher`）。颜色、字号和组件的名字用英文。

所有边角都是直角，`border-radius: 0`。

分割线是 1px 直线。层次只留黑白灰，不使用彩色。黑、白两套用同一批名字，只对调底和字。状态完全依靠文字、图标与实心/空心/灰色边框表达，不引入红、绿、黄。

## Color

| Token | 黑 (默认) | 白 | 用在 |
| --- | --- | --- | --- |
| `ground` | `#000000` | `#ffffff` | 所有页面的底、弹窗底色 |
| `ink` | `#ffffff` | `#000000` | 主文字、激活高亮、出错竖线/边框 |
| `gray` | `#888888` | `#888888` | Meta、次要说明、已填/次要边框 |
| `line` | `#333333` | `#e5e5e5` | 分割线、幽灵按钮边框、输入框底边/边框 |
| `wash` | `#111111` | `#f2f2f2` | 浅底块、按钮/菜单按下态反馈底色 |
| `solid` | `#ffffff` | `#000000` | 实心按钮的底、选中态背景 |
| `solid-ink` | `#000000` | `#ffffff` | 实心按钮的字、选中态文字 |

默认是黑。顶栏右上始终挂 `ThemeSwitch`；仅在黑白之间切换（无跟随系统选项），选择保存在 `localStorage`（键名 `theme`）。

## Type

全站一种无衬线，中文优先：`"Noto Sans CJK SC"`, `"Source Han Sans SC"`, `"Noto Sans SC"`, `sans-serif`。

字重只用两档：正文 400，标题 500。

| Role | Size | Weight | Line height | 用在 |
| --- | --- | --- | --- | --- |
| `Kicker` | 13px | 400 | 1.4 | 导航、标签、App 名。字距 `0.18em` |
| `Title` | 28px | 500 | 1.2 | 页面主标题 |
| `AppTitle` | 22px | 500 | 1.2 | 分区标题 |
| `Meta` | 14px | 400 | 1.45 | 说明文字、表格次要信息 |
| `ButtonLabel` | 继承所在表面的正文 | 400 | 1.2 | 按钮字 |

## Components

### `Button`

`ButtonSolid`：底 `solid`，字 `solid-ink`，边框 1px `solid`，直角，内边距 12px 14px。

`ButtonGhost`：透明底，1px `line` 边框，直角，内边距 8px 12px。悬停边框变 `ink`，按下 `wash` 底。选中态改为 `solid` 底、`solid-ink` 字。

### `TextField`

无底色，只有底边 1px `line`（复杂表单为 1px `line` 四周边框），直角。上下内边距 8px。聚焦时 1px `ink` 外偏 2px 轮廓。

### `AppBar`

顶栏极简。
- 公开页（首页、登录）：左侧 `LogoMark`（Beep 铃铛图标，颜色随 `ink`）+ `Beep` 字标；右侧为 `LocaleSwitcher` + 登录/控制台入口 + `ThemeSwitch`。
- 控制台内：左侧 `LogoMark` + `Beep` 字标 + 工作区选择器（`AccountSwitcher`）；右侧为 `LocaleSwitcher` + `ThemeSwitch` + 用户菜单（个人偏好、API Tokens、Admin、Dev、退出）。

### `Subnav`

紧贴 `AppBar` 下方。采用 `Kicker` 风格字号与字距。
- 账号工作区：`概览`、`Beeps`、`Beepers`、`Runners`、`设置`。
- 管理后台：`← 工作区`、`任务`、`统计`。
- 个人中心：`← 工作区`、`Tokens`、`偏好`。
手机端支持横向滑动（`overflow-x-auto`），吸顶展示。当前激活项使用 `ink` 底部高亮边框。

### `Timeline`

时间线与紧凑行列表。每行上边框 1px `line`，上下内边距 10px。角色/标签前置，时间当天显示 `HH:mm`，跨天带日期。重要变动左侧加 1px `line` 竖线。

### `Notice`

一行状态或提示。无背景底色，左侧 1px `line` 竖线，前面小号 `gray` 标签。异常/错误状态将竖线替换为 `ink`、标签显示「错误」或「失败」，绝不使用红色。

### `Panel`

分区容器，不填底色。
- 手机端（<768px）：仅保留上边 1px `line` 分隔。
- 平板与电脑端（≥768px）：四边 1px `line` 空心方框。
`wash` 只留给悬停或按下态反馈。

### `OtpField`

6 个等宽方格，1px `line` 边，当前聚焦格 `ink` 边，已填格 `gray` 边。底层为透明的真实输入框，覆盖在方格上。自动填充、粘贴、退格走原生行为。填满 6 位自动提交；出错时格子变 `ink` 边重置。

### `Segmented`

视图与选项切换（如设置页的 通用 / 渠道 / 通知）。多枚等宽 `ButtonGhost`，选中项为 `solid` 底、`solid-ink` 字。

### `Dialog / Drawer`

弹层与模态操作：
- 手机端（<768px）：底部抽屉，直角，仅顶部 1px `line` 边框，无圆角拉手。
- 平板与电脑端（≥768px）：居中方框，四边 1px `line`，直角，无投影。
遮罩层采用 `ground` 配合 80% 半透明。

### `Table / List`

- 手机端（<1024px）：采用 `Timeline` 风格的多行列表，主次分明，减少横向滚动。
- 电脑端（≥1024px）：采用 1px `line` 网格的表格。

## Layout

移动优先。断点：`md` 768px、`lg` 1024px。

| | 手机 (<768) | 平板 (768–1023) | 电脑 (≥1024) |
| --- | --- | --- | --- |
| 页边距 | 16px | 24px | 32px |
| 顶栏高 | 56px | 64px | 64px |
| `Panel` | 无底色，只留上边 1px `line` | 四边 1px `line` 空心方框 | 同左 |
| 表单主按钮 | 通栏 | 自然宽度 | 自然宽度 |
| 任务与资源列表 | 行列表展示 | 行列表展示 | 数据表格展示 |

内容最大宽：首页、控制台 1080px，登录 440px。顶栏通栏，内容与内容区对齐。适配 `safe-area-inset-*`。

## Interaction

- 点击区至少 44px（电脑端 40px）。
- 悬停效果仅在 `(hover: hover)` 设备上生效；触屏用 `:active` 给按下反馈。
- 键盘焦点统一 1px `ink` 描边、外偏 2px。
- 有异步动作的按钮进入 `is-loading`：方形转圈 + 「…中」文案，并禁用。
- 不可逆操作（断开绑定、删除探针、撤销 Token 等）采用确认弹窗（直角风格）。
- 输入框按语义设 `inputMode` / `enterKeyHint` / `autoComplete`；手机端字号 ≥16px 防缩放。
- 系统 `theme-color` 跟随 `ground`。
- `prefers-reduced-motion` 时关掉过渡与动画。
