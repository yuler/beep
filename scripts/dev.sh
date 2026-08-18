#!/usr/bin/env bash
set -euo pipefail

ok()    { gum style --foreground 78  "✔ $*"; }
warn()  { gum style --foreground 227 "⚠ $*"; }
step()  { gum style --foreground 99  "▶ $*"; }

# PIDs listening on any of the given TCP ports (unique, space-separated).
listening_pids() {
  local port
  {
    for port in "$@"; do
      lsof -nP -iTCP:"$port" -sTCP:LISTEN -t 2>/dev/null || true
    done
  } | sort -u | tr '\n' ' '
}

stop_overmind() {
  if [ -S .overmind.sock ]; then
    overmind quit 2>/dev/null || overmind kill 2>/dev/null || true
    rm -f .overmind.sock
  fi
  # Overmind's tmux session defaults to the app directory basename
  tmux kill-session -t "$(basename "$PWD")" 2>/dev/null || true
}

ensure_ports_free() {
  local pids
  pids="$(listening_pids "$@")"
  [ -n "${pids// /}" ] || return 0

  warn "Port(s) in use: $*"
  local port
  for port in "$@"; do
    lsof -nP -iTCP:"$port" -sTCP:LISTEN 2>/dev/null || true
  done
  echo

  gum confirm "Force kill and restart?" || { warn "Aborting."; exit 1; }

  # Quit overmind first so it does not respawn children we kill
  stop_overmind
  # shellcheck disable=SC2086
  kill -9 $pids 2>/dev/null || true
  sleep 0.3
  # Re-query after the grace period: a child may have re-bound the port
  # (e.g. puma worker restart) between the first kill and now.
  pids="$(listening_pids "$@")"
  # shellcheck disable=SC2086
  [ -n "${pids// /}" ] && kill -9 $pids 2>/dev/null || true

  ok "Ports cleared — continuing startup"
}

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

# Env comes from mise (`_.file` loads `.env` / `.env.local` in mise.toml).
# Running dev.sh directly would skip env loading, so fail fast.
if [ -z "${MISE_TASK_NAME:-}" ]; then
  warn "dev.sh must be started via 'mise dev' (not directly)." >&2
  exit 1
fi

# Values come from `.env` / `.env.local` (copy `.env.example`). No hardcoded hosts/ports.
: "${APP_HOST:?APP_HOST is required. Copy .env.example to .env.}"
: "${CORE_PORT:?CORE_PORT is required. Copy .env.example to .env.}"
: "${WEB_PORT:?WEB_PORT is required. Copy .env.example to .env.}"

CORE_URL="${CORE_URL:-http://core.${APP_HOST}:${CORE_PORT}}"
WEB_URL="${WEB_URL:-http://web.${APP_HOST}:${WEB_PORT}}"

step "Local subdomain URLs (*.localhost → 127.0.0.1)"
ok "core  ${CORE_URL}"
ok "web   ${WEB_URL}"
echo

step "Checking ports ${WEB_PORT} (web) and ${CORE_PORT} (core)…"
ensure_ports_free "${WEB_PORT}" "${CORE_PORT}"

export RUBY_DEBUG_OPEN="${RUBY_DEBUG_OPEN:-true}"
export RUBY_DEBUG_LAZY="${RUBY_DEBUG_LAZY:-true}"

step "Starting via overmind (Procfile.dev)…"
# -N: we set ports ourselves via CORE_PORT / WEB_PORT
exec overmind start -f Procfile.dev -N "$@"
