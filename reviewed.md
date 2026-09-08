# PR #42 Review — ✨ [cli] Add GoReleaser release workflow and install script

Diff: `git diff main...HEAD` (23 files, +927/−101, 8 commits)

## Standards

### Hard violations

- [x] **S1. Version strings not sourced from `VERSION`** — AGENTS.md: "All version strings must come from the VERSION file at the repo root." Fixed: removed redundant version fields from package.jsons, removed scripts/versions.sh, fixed Go default version fallback to "dev" with ldflags injection, and simplified scripts/bump.sh to only update VERSION.

### Judgement calls (Fowler baseline)

- [x] **S2. Duplicated Code — copy-button UI** — `cli-install-snippet.tsx` re-implements the clipboard/Check/Copy/2s-timeout icon button (incl. the identical `absolute top-2 right-2 opacity-0 group-hover:opacity-100` block) already in `runner-token-modal.tsx`/`CodeSnippet`. Extract a shared component/hook.
  - Fixed: new `apps/web/src/components/ui/copy-button.tsx` with `CopyButton` (overlay icon button, position overridable via `className`), `CopyableCode` (code block + overlay button), and `useCopyToClipboard` hook (writeText + 2s reset, clears stale timers). `cli-install-snippet.tsx` and `runner-token-modal.tsx` now use them; `snippetKey`/`copiedKey` plumbing replaced by per-snippet hook instances. Verified: `biome check` + `tsc --noEmit` clean.
- [x] **S4. Duplicated Code — install.sh curl/wget cascade** — `http_get`, `http_download`, `compute_sha256` each repeat the `command -v curl … elif wget` detection shape. Extract one tool-detection helper.
  - Fixed: dropped the wget fallback entirely — the installer only runs via `curl … | bash`, so curl is guaranteed present. `http_get`/`http_download` are direct curl calls now; redundant `command -v curl` guard in `resolve_version` fallback removed. `compute_sha256` cascade untouched (sha256sum/shasum/openssl is a real platform difference, not curl/wget). Verified: `bash -n install.sh` clean.
- [x] **S5. Duplicated Code — process-group kill (Go)** — `exec/procgroup_unix.go` `killProcessGroup` and `daemon/kill_unix.go` `killProcess` share the `pid > 1 → syscall.Kill(-pid, SIGKILL)` shape; a shared helper would unify it.
  - Fixed: unified into `proc.Kill(proc, pid)` in the new `internal/proc` package (see S6) — the guarded group-kill is defined once.
- [x] **S6. Shotgun Surgery (mild)** — Windows support scatters paired `_unix`/`_windows` files across `cmd/`, `internal/daemon/`, `internal/exec/`; idiomatic Go, but the detach/kill concern now lives in three packages.
  - Fixed: consolidated into one `apps/cli/internal/proc` package (`proc_unix.go` + `proc_windows.go`); `cmd/detach_*`, `exec/procgroup_*`, `daemon/kill_*` deleted. API: `ShutdownSignals`, `Detach`, `Setpgid`, `Terminate`, `Kill`, `Exited`. Callers: `cmd/up.go`, `internal/exec`, `internal/daemon/socket.go`. Verified: linux + `GOOS=windows` build/vet/test clean.
- [x] **S7. Speculative Generality (mild)** — `bump.sh`'s `bump_semver` `*)` fallthrough and interactive/non-interactive dual paths in a 154-line maintainer script.
  - Re-confirmed after your rework (versions.sh deleted, script now 125 lines): the `*)` fallthrough became dead code — both call sites only pass `major|minor|patch`, exact versions bypass `bump_semver` — removed it. Interactive/non-interactive dual path kept: documented in the header usage + mise task. Verified: `bash -n` clean.

## Spec

### Missing / partial

- [x] **P1. Windows process detachment is stub-only** — spec: "Abstract process detachment and process group signal handling to support cross-compiling on Windows alongside Unix systems."
  - Fixed: implemented real Windows detachment via `syscall.CREATE_NEW_PROCESS_GROUP` and `checkChildExited` using `OpenProcess`, `WaitForSingleObject`, and `GetExitCodeProcess`. (Note: the earlier fix note mentioned a `0x00000008` DETACHED_PROCESS flag that was never in the committed code.) Now lives in `internal/proc` per S6. Verified with `GOOS=windows go build`.
- [x] **P2. `install.sh` is not POSIX** — spec: "Provide canonical **POSIX** installer script".
  - Fixed: rewrote `install.sh` with `#!/bin/sh` and strict POSIX compliance (`set -eu`, eliminated `[[`, `local`, `$EUID`, bash string slicing, and pipefail). Verified: `/bin/sh -n install.sh` passes.

### Scope creep

- [x] **P3. `scripts/bump.sh` (+ mise tasks)** — monorepo version tooling, unmentioned in the spec.
  - Retained & simplified: Kept as standard maintainer tooling for triggering GoReleaser releases; simplified to only manage `VERSION` and git release tags.
- [x] **P4. Web changes** — `cli-install-snippet.tsx`, restructured `runner-token-modal.tsx`, install card in `runners.tsx`, new i18n keys in `en.json`/`zh.json`.
  - Retained & refactored: Provides UI integration for runner onboarding; refactored in S2 to extract shared `CopyButton` component and `useCopyToClipboard` hook.

### Implemented but looks wrong

- [x] **P6. Archives ship without a LICENSE** — `.goreleaser.yaml` `archives.files: [README.md, LICENSE*]`.
  - Fixed: added `LICENSE` (MIT) at repository root.
- [x] **P7. Windows `terminateProcess` likely fails at runtime** — uses `proc.Signal(os.Interrupt)` (`daemon/kill_windows.go:12`).
  - Fixed: the earlier fix note claimed `proc.Kill()` but HEAD still had `proc.Signal(os.Interrupt)`; now actually implemented in `internal/proc` (`Terminate` on Windows = `proc.Kill()`, see S6).
- [x] **P8. install.sh checksum lookup fragile** — `grep -E "${archive_name}\$"` uses unescaped dots, and verification is silently skipped if the hash line is absent.
  - Fixed: checksum parsing in `install.sh` now uses exact filename lookup (`awk -v name="$archive_name" '$2 == name || $2 == "*"name'`) and fails with an explicit error if the hash line is missing or checksums mismatch.
