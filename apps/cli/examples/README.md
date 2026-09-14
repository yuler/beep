# Runner workspace examples

Put scripts on the machine that runs `beep runner`. Core only stores the job slug, cron, and optional config. The runner looks up `slug` in this workspace.

Default workspace: `~/.beep`

```
~/.beep/
  jobs/
    intranet-http           # extensionless executable named after the job slug
```

Create a job in the web UI with the same slug, then:

```bash
cp examples/intranet-http ~/.beep/jobs/intranet-http
chmod +x ~/.beep/jobs/intranet-http
beep runner config set --server https://core.example.com --token beep_rt_...
beep runner up
```

The process injects `BEEP_RUNNER_LOG_URL`, `BEEP_RUNNER_RESULT_URL`, and `BEEP_RUNNER_CONFIG_*` from the job config. Stdout/stderr is uploaded as the run log. Exit `0` is `ok`; any other exit is `alerting`. Scripts may also `POST` JSON to `$BEEP_RUNNER_RESULT_URL`:

```json
{ "status": "ok", "title": "healthy", "message": "...", "metrics": { "latency_ms": 12 } }
```

## Channel hooks

`beep up` pulls CLI channel deliveries and execs `$WORKSPACE/hooks/on_channel` for every event (`beep.fired`, `channel.test`, …). Copy the example, drop the `.example` suffix:

```bash
cp examples/hooks/on_channel.example ~/.beep/hooks/on_channel
chmod 755 ~/.beep/hooks/on_channel
```

Branch on `BEEP_EVENT` inside the script (see `docs/architecture/channel.md` §5). Lookup also checks `$WORKSPACE/.beep/hooks/on_channel`, then legacy `on_beep` / `on_beep_fired`. The hook must be owned by you, executable, and not group/world-writable, or the CLI refuses to run it.
