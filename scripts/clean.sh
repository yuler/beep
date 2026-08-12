#!/usr/bin/env bash
set -euo pipefail

# Colorized logging helpers that use gum when available, otherwise plain echo.
log()   { gum style --foreground 212 "✦ $*"; }
ok()    { gum style --foreground 78  "✔ $*"; }
warn()  { gum style --foreground 227 "⚠ $*"; }
step()  { gum style --foreground 99  "▶ $*"; }

if ! command -v gum >/dev/null 2>&1; then
  warn "gum not found; falling back to plain output"
  log()  { echo "✦ $*"; }
  ok()   { echo "✔ $*"; }
  warn() { echo "⚠ $*"; }
  step() { echo "▶ $*"; }
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

LOG_LABEL="core: logs & tmp (rails log:clear tmp:clear)"
DB_LABEL="core: storage (rm core/storage/*.sqlite3)"
NODE_LABEL="core: node_modules (rm node_modules)"
ENV_LABEL="root: .env (rm .env)"
FE_LABEL="core: frontend build (propshaft assets)"
ALL_LABEL="All of the above"

choices=$(gum choose --no-limit \
  --selected="$FE_LABEL" \
  --header="Space to select · Enter to confirm" \
  "$ALL_LABEL" \
  "$LOG_LABEL" \
  "$DB_LABEL" \
  "$NODE_LABEL" \
  "$ENV_LABEL" \
  "$FE_LABEL")

if [ -z "$choices" ]; then
  ok "Nothing selected. Aborted."
  exit 0
fi

if ! gum confirm "This will remove the selected items. Continue?"; then
  ok "Aborted."
  exit 0
fi

clean_logs() {
  step "Clearing Rails logs and tmp"
  (cd core && ./bin/rails log:clear tmp:clear 2>/dev/null) || rm -rf core/log/* core/tmp/*
  ok "Rails logs & tmp cleared"
}

clean_storage() {
  step "Removing SQLite databases"
  rm -f core/storage/*.sqlite3 core/storage/*.sqlite3-shm core/storage/*.sqlite3-wal
  ok "Storage cleared"
}

clean_node() {
  step "Removing node_modules"
  rm -rf node_modules core/node_modules apps/*/node_modules packages/*/node_modules
  ok "node_modules removed"
}

clean_env() {
  step "Removing local .env files"
  rm -f .env
  ok ".env removed"
}

clean_frontend() {
  step "Removing Rails frontend build output"
  rm -rf core/app/assets/builds/application.css \
         core/app/assets/builds/.manifest.json \
         core/public/assets \
         core/tmp/cache/assets
  ok "Rails frontend build output cleared"
}

while IFS= read -r line; do
  [ -z "$line" ] && continue
  case "$line" in
    "$ALL_LABEL")   clean_logs; clean_storage; clean_node; clean_env; clean_frontend; exit 0 ;;
    "$LOG_LABEL")  clean_logs ;;
    "$DB_LABEL")   clean_storage ;;
    "$NODE_LABEL") clean_node ;;
    "$ENV_LABEL")  clean_env ;;
    "$FE_LABEL")   clean_frontend ;;
  esac
done <<< "$choices"

ok "Clean complete!"
