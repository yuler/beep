# review

PR: [#38](https://github.com/yuler/beep/pull/38) `feat/self-hosted-runner` vs `main`  
Spec: `docs/architecture/runner.md`  
Scope: bugs + design defects (not style nits)

---

## Fixed (this session)

- [x] #1 Token modal undefined `token`
- [x] #2 `PullJob` rename collision wipe
- [x] #3 Job env inherits `BEEP_RUNNER_TOKEN`
- [x] #4 Reclaim can fail just-claimed run (`from_statuses: pending`)
- [x] #5 Atomic `record_result!`
- [x] #6 Pause orphans reclaim
- [x] #7 Log/result 422 treated as success
- [x] #8 Process group kill on job timeout/stop
- [x] #9 `up -d` orphan child on socket timeout
- [x] #10 In-flight jobs ignore daemon cancel ctx
- [x] #11 `job remove` after rename (delete by `@id`)
- [x] #13 Core ignores per-job `timeout_seconds` for reclaim (fixed: dynamic timeout based on `claimed_at + timeout_seconds`)
- [x] #14 Timezone edits refresh `next_run_at`
- [x] #15 Callback URLs from configured host (not request Host)
- [x] #16 Spec: scripts must not use runner token for callbacks
- [x] #17 Spec pairing order → `@id` then slug
- [x] #18 Web job create/edit hints & slug rename warning
- [x] #19 Offline banner / warning in runner detail UI
- [x] #20 Surface `serverErr` in `push` & check `@id` write errors
- [x] #22 Stale online badge on detail page (`mark_stale_offline` in show action)
- [x] N+1 query in `RunnersController#index` eliminated via aggregate `jobs_count`
- [x] CLI `daemon.go` log upload improved to non-blocking asynchronous buffer
- [x] CLI `workspace.go` added script extension fallback for job resolution (`.sh`, `.py`, `.js`, etc.)
- [x] Blank cron on sync no longer silently defaults when explicitly provided empty
- [x] Corrupt `config.json` returns explicit parsing error instead of silently defaulting
- [x] Delete job automatically selects next remaining job instead of clearing selection

---

## Still open

### High

### 12. Runner tokens stored plaintext (PR claim vs code)
**Where:** `core/app/models/runner.rb` `has_secure_token` + `find_by(token:)`

PR summary claims SHA-256 hashed `beep_rt_*` tokens; DB stores and matches raw strings. Backup / DB leak = full runner impersonation.

**Fix:** Store only a hash; return raw token once on create / regenerate. Align PR/docs with reality until then.

---

### Medium / Future Features

### 21. `alerting` results notification integration
**Where:** `Runner::Run` stores `alerting` / `error`; no Beep/notify hook

Phase 1 establishes runner infrastructure; alert notification channels can be hooked up in Phase 2.
