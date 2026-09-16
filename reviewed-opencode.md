# PR #49 Review 记录

分支：`feat/cli-beep-and-beeper-commands` → `main`（约 5200 行新增）
结论：Core 改动干净，可合；CLI 建议合入前修 delete 无确认，其余可跟进。

已与 `reviewed-gemini.md` 去重：UTF-8 按字节截断、`parseAPIError` 无界回显、Ctrl+C/`cancelled` 误回退表单、`beeper create` Ping Token 明文/脱敏见 Gemini。

## Core（没问题）

- `core/app/controllers/concerns/api/v1/responses.rb` + beeps/beepers controller 透出 `errors: full_messages`，与 CLI `parseAPIError` 对齐，无敏感泄露。
- `core/app/models/beep/run.rb:7` 加 `dependent: :destroy` 修孤儿 `Channel::Delivery`，有测试覆盖。需确认 `Beep` 对 `runs` 本来就有 `dependent: :destroy`（级联依赖它）。

## 安全问题

1. `delete` 无二次确认 — `apps/cli/cmd/beep/delete.go`、`apps/cli/cmd/beeper/delete.go` 选完 ID 直接删。建议 interactive 非 `--json` 下加 `PromptConfirm`。
2. Ping URL 未转义 — `apps/cli/cmd/beeper/show.go:73`、`apps/cli/cmd/beeper/create.go:191` 用 `%s` 拼 token，应统一 `url.PathEscape`（其他 Get/Delete 已用）。
3. `EnsureLoggedIn` fail-open — `apps/cli/internal/cmdutil/cmdutil.go:57-61` 非 401 错误直接放行，且 401 靠字符串匹配，建议区分网络错误与鉴权错误。

## BUG

4. 负延迟可通过 — `apps/cli/internal/client/beep.go:330-344` `ParseInDuration` 接受 `-5m` / `-1d`，`ToRequest` 未校验 `d > 0`，`run_at` 落到过去。应拒绝 `<= 0`。
5. `--config` 写错静默丢弃 — `apps/cli/cmd/beeper/create.go:55-60` 无 `=` 的项直接跳过，应报错。
6. 时区校验缺口 — `CreateBeeperParams.ToRequest`（`apps/cli/internal/client/beeper.go:241`）不校验 Timezone；`handleNaturalCreate`（`apps/cli/cmd/beep/create.go:279`）直接信任 AI 的 `proposal.Timezone`；`CreateBeepParams.ToRequest`（`apps/cli/internal/client/beep.go:309`）`LoadLocation` 失败静默 fallback 却仍发送非法值。AI 路径应先 `ValidIANATimezone`。
7. 小项 — `--channels` 未去重（`client/beep.go:281`、`client/beeper.go:255`）；`ProposalErrors` 兜底分支吞未知类型（`client/beep.go:156`）；interactive 有 title 也先调 AI，离线时多一次往返。

验证：`go test ./internal/client/ ./cmd/` 通过（apps/cli 下）。
