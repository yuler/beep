# PR Review Report: `feat/cli-beep-and-beeper-commands`

分支：`feat/cli-beep-and-beeper-commands` → `main`
结论：CLI 建议合入前修 2 个 P0（UTF-8 截断、交互带参默认 UTC），其余可跟进。

## 安全问题

1. `beeper create --json` 拿不到 Ping Token — `apps/cli/cmd/beeper/create.go:169-195` 终端模式明文打印完整 `PingToken` + Ping URL，`--json` 走 `Redacted(false)` 无条件打码，且无 `--show-token`。脚本 `jq -r .ping_token` 只能拿到掩码。建议 `create --json` 与终端一致返回明文，或加 `--show-token` 控制。

## BUG

2. 按 byte 截断标题 — `apps/cli/cmd/beep/beep.go:61`、`apps/cli/cmd/beep/list.go:65`、`apps/cli/cmd/beeper/apps.go:55`、`apps/cli/cmd/beeper/beeper.go:64`、`apps/cli/cmd/beeper/list.go:78` 的 `title[:n]` 会切断中文/Emoji UTF-8。应按 `[]rune` 截断。
3. 交互终端带参创建默认 UTC — `apps/cli/cmd/beep/create.go:67-77`、`apps/cli/cmd/beeper/create.go:63-73` 只在 `!IsInteractive` 时 `DetectTimezoneOK`。TTY 下 `beep create "喝水" --at "16:30"` 不进表单、时区为空，`ToRequest` 兜底 `UTC`，东八区会偏 8 小时。未传 `--timezone` 时应始终探测本机时区。
4. Ctrl+C 误回退表单 — `apps/cli/cmd/beep/create.go:150-160` 只匹配 `"cancelled"`。`huh.ErrUserAborted` 是 `"user aborted"`，会当成 AI 失败并弹出 `PromptBeepCreate`。应识别 `huh.ErrUserAborted` / `context.Canceled` / `"user aborted"`。
5. `--json` 仍走交互 — `apps/cli/internal/cmdutil/cmdutil.go:90-97` 的 `IsInteractive` 未排除 `IsJSON`。TTY 上 `beep create --json` 会弹出向导、混入 ANSI。应在 `IsJSON(cmd)` 时返回 `false`。
6. 时区提示晚于 Schedule 校验 — `apps/cli/internal/ui/prompt_beep.go:133-145` 顺序是 Title → Body → Schedule → Timezone。`16:30` 在时区未选时按 UTC 校验。应将时区提前，或先填本机默认时区。
7. 错误 body 无界回显 — `apps/cli/internal/client/client.go:449` 的 `parseAPIError` 把 502 HTML 整页拼进 error，建议截断（如 300 字符）。
8. 校验错误 map 乱序 — `apps/cli/internal/client/client.go:420-435` 迭代 `map[string][]string` 无序，重试/测试断言不稳定。应对 key 排序后再拼接。
9. `/api/v1/me` 预检开销 — `apps/cli/internal/cmdutil/cmdutil.go:47-63` 对 list/show 等只读命令也先 `GET /me`，多一次 50–200ms。目标 API 的 401 已能被 `parseAPIError` 处理。
