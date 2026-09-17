# Beep CLI

The `beep` command-line interface and background daemon runner for Beep.

## Commands

- **`service`**: Manage local background daemon services (`runner` and `channel`)
  - `start` / `up`: Start background daemon services
  - `stop` / `down`: Stop running daemon services
  - `restart`: Restart daemon services (supports interactive multi-select and non-interactive CI fallback)
  - `status`: Show daemon process state, PID, socket, and logs
- **`channel`**: Connect this CLI as an alert delivery channel
  - `connect`, `disconnect`, `status`, `start` / `up`, `stop` / `down`
- **`runner`**: Manage Beep self-hosted runner daemon and workspace
  - `connect`, `disconnect`, `status`, `start` / `up`, `stop` / `down`, `job` (`list`, `create`, `delete`, `pull`, `push`)
- **`beep`**: Trigger one-off alerts, fire webhooks, or test notifications
- **`beeper`**: Manage server-side monitor probes
- **`auth`**: Login and session management (`login`, `logout`, `status`, `switch`)
- **`config`**: Manage CLI configuration (`show`, `list`, `get`, `set`, `path`)

## Local Development (`bin/beep-local`)

During local development, use `bin/beep-local` (or `beep-local` when `mise` has added `./bin` to `PATH`):

```bash
# Display help
bin/beep-local --help

# Check service status
bin/beep-local service status

# Restart services
bin/beep-local service restart
```

### Why `bin/beep-local`?

`bin/beep-local` is a development launcher script at the monorepo root that executes `apps/cli` on the fly via `go run`. It provides:

1. **Working Directory Preservation**: Retains the caller's working directory so relative paths work as expected.
2. **Local Core Server Resolution**: Automatically sets the default server URL to `http://core.${APP_HOST}:${CORE_PORT}` (from the monorepo `.env`).
3. **Workspace Isolation**: Defaults to `~/.beep.local` instead of `~/.beep`, preventing dev configs from colliding with real CLI environments.
4. **Contextual Binary Naming**: Sets `BEEP_BIN_NAME="beep-local"`, so help messages, examples, and prompts accurately reflect the `beep-local` command.

## Testing & Building

```bash
# Run unit tests
mise run cli:test     # or: cd apps/cli && go test -v ./...

# Build production binary (outputs to bin/beep)
mise run cli:build
```
