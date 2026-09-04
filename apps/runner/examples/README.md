# Runner workspace examples

Put scripts on the machine that runs `beep-runner`. Core only stores the job slug, cron, and optional config. The runner looks up `slug` in this workspace.

Default workspace: `~/.beep-runner`

```
~/.beep-runner/
  jobs.json                 # optional command map
  jobs/
    intranet-http           # extensionless executable named after the job slug
```

Create a job in the web UI with the same slug, then:

```bash
cp examples/intranet-http.sh ~/.beep-runner/jobs/intranet-http
chmod +x ~/.beep-runner/jobs/intranet-http
beep-runner config set --server https://core.example.com --token beep_runner_...
beep-runner up
```

The process injects `BEEP_LOG_URL`, `BEEP_RESULT_URL`, and `BEEP_CONFIG_*` from the job config. Stdout/stderr is uploaded as the run log. Exit `0` is `ok`; any other exit is `alerting`. Scripts may also `POST` JSON to `$BEEP_RESULT_URL`:

```json
{ "status": "ok", "title": "healthy", "message": "...", "metrics": { "latency_ms": 12 } }
```
