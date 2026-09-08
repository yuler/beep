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
- [ ] **S5. Duplicated Code — process-group kill (Go)** — `exec/procgroup_unix.go` `killProcessGroup` and `daemon/kill_unix.go` `killProcess` share the `pid > 1 → syscall.Kill(-pid, SIGKILL)` shape; a shared helper would unify it.
- [ ] **S6. Shotgun Surgery (mild)** — Windows support scatters paired `_unix`/`_windows` files across `cmd/`, `internal/daemon/`, `internal/exec/`; idiomatic Go, but the detach/kill concern now lives in three packages.
- [ ] **S7. Speculative Generality (mild)** — `bump.sh`'s `bump_semver` `*)` fallthrough and interactive/non-interactive dual paths in a 154-line maintainer script.

## Spec

### Missing / partial

- [ ] **P1. Windows process detachment is stub-only** — spec: "Abstract process detachment and process group signal handling to support cross-compiling on Windows alongside Unix systems." `setProcessDetach` is a no-op (`apps/cli/cmd/detach_windows.go:14`) so `beep runner up` doesn't detach on Windows, and `checkChildExited` always returns `false`, silently disabling the early-exit check. OK for *cross-compiling*, but degraded runtime behavior isn't flagged in the Verification claims.
- [ ] **P2. `install.sh` is not POSIX** — spec: "Provide canonical **POSIX** installer script". It's bash-only (`#!/usr/bin/env bash`, `[[`, `EUID`, `local`). Works via `| bash`, but not POSIX sh as claimed.

### Scope creep

- [ ] **P3. `scripts/bump.sh` (+ mise tasks)** — monorepo version tooling, unmentioned in the spec.
- [ ] **P4. Web changes** — `cli-install-snippet.tsx`, restructured `runner-token-modal.tsx`, install card in `runners.tsx`, new i18n keys in `en.json`/`zh.json` — no web/UI work was asked for (the snippet URL does match the spec's curl command).

### Implemented but looks wrong

- [ ] **P6. Archives ship without a LICENSE** — `.goreleaser.yaml` `archives.files: [README.md, LICENSE*]` — no LICENSE file exists at repo root, so the glob matches nothing and archives lack a license file.
- [ ] **P7. Windows `terminateProcess` likely fails at runtime** — uses `proc.Signal(os.Interrupt)` (`daemon/kill_windows.go:12`), which typically fails for processes outside the caller's console — graceful `StopDaemon` will likely error on Windows then fall through to timeout.
- [ ] **P8. install.sh checksum lookup fragile** — `grep -E "${archive_name}\$"` uses unescaped dots, and verification is silently skipped if the hash line is absent.
