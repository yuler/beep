#!/usr/bin/env bash
# Export production SQLite database as SQL statements and import into local development.
#
# IMPORTANT / SECURITY NOTICE:
# Dump files contain production data (including potentially PII/secrets).
# Dumps MUST stay local under core/storage/ (gitignored) and must NEVER be committed.
#
# Connects to the Dokploy server over SSH, finds the running core container by
# its <project>-<service> names, dumps SQLite SQL from production, saves it locally,
# and imports it into core/storage/development.sqlite3.
#
# Usage:
#   bash scripts/db-pull-production.sh               # Export from production and import into local DB
#   bash scripts/db-pull-production.sh -y            # Non-interactive (assume yes)
#   bash scripts/db-pull-production.sh --export-only # Only export SQL to local file
#   bash scripts/db-pull-production.sh --import-only <file.sql> # Only import an existing SQL file
#   bash scripts/db-pull-production.sh [output.sql]  # Custom dump file path
set -euo pipefail

# Output helpers
log()   { gum style --foreground 212 "✦ $*"; }
ok()    { gum style --foreground 78  "✔ $*"; }
warn()  { gum style --foreground 227 "⚠ $*"; }
step()  { gum style --foreground 99  "▶ $*"; }
err()   { gum style --foreground 196 "✖ $*"; }

if ! command -v gum >/dev/null 2>&1; then
  log()  { echo "✦ $*"; }
  ok()   { echo "✔ $*"; }
  warn() { echo "⚠ $*"; }
  step() { echo "▶ $*"; }
  err()  { echo "✖ $*"; }
fi

print_help() {
  cat << 'EOF'
Usage:
  scripts/db-pull-production.sh [options] [dump_file]

Options:
  --export-only             Only export SQLite SQL from production to local file
  --import-only <file.sql>  Only import specified SQL file into local database
  -y, --yes                 Skip confirmation prompts
  -h, --help                Show this help message

Environment Variables:
  PRODUCTION_HOST           SSH host (default: beep.yuler.cc)
  PRODUCTION_USER           SSH user (default: root)
  PRODUCTION_PROJECT        Dokploy project name (default: beep)
  PRODUCTION_SERVICE        Dokploy service name (default: core)
  PRODUCTION_SSH_OPTS       Extra options passed to ssh
  TARGET_DB                 Local SQLite database path (default: core/storage/development.sqlite3)
EOF
}

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TIMESTAMP="$(date +%Y%m%d_%H%M%S)"
DEFAULT_DUMP_FILE="$ROOT_DIR/core/storage/production_dump_${TIMESTAMP}.sql"
LATEST_SYMLINK="$ROOT_DIR/core/storage/production_dump.sql"
LOCAL_DB="${TARGET_DB:-$ROOT_DIR/core/storage/development.sqlite3}"
RAILS_DIR="$ROOT_DIR/core"

HOST="${PRODUCTION_HOST:-beep.yuler.cc}"
USER="${PRODUCTION_USER:-root}"
PROJECT_NAME="${PRODUCTION_PROJECT:-beep}"
SERVICE_NAME="${PRODUCTION_SERVICE:-core}"

EXPORT_ONLY=false
IMPORT_ONLY=false
ASSUME_YES=false
DUMP_FILE=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --export-only)
      EXPORT_ONLY=true
      shift
      ;;
    --import-only)
      IMPORT_ONLY=true
      shift
      if [[ $# -gt 0 && ! "$1" =~ ^- ]]; then
        DUMP_FILE="$1"
        shift
      fi
      ;;
    -y|--yes)
      ASSUME_YES=true
      shift
      ;;
    -h|--help)
      print_help
      exit 0
      ;;
    -*)
      err "Unknown option: $1"
      print_help
      exit 1
      ;;
    *)
      if [ -z "$DUMP_FILE" ]; then
        DUMP_FILE="$1"
      else
        err "Unexpected argument: $1"
        exit 1
      fi
      shift
      ;;
  esac
done

if [ "$IMPORT_ONLY" = true ] && [ -z "$DUMP_FILE" ]; then
  if [ -e "$LATEST_SYMLINK" ]; then
    DUMP_FILE="$LATEST_SYMLINK"
  else
    LATEST_FOUND=$(find "$ROOT_DIR/core/storage" -maxdepth 1 -name "production_dump_*.sql" -type f 2>/dev/null | sort -r | head -n 1 || true)
    if [ -n "$LATEST_FOUND" ]; then
      DUMP_FILE="$LATEST_FOUND"
    else
      err "No SQL dump found to import. Please specify a file: --import-only <file.sql>"
      exit 1
    fi
  fi
else
  DUMP_FILE="${DUMP_FILE:-$DEFAULT_DUMP_FILE}"
fi

confirm_prompt() {
  local prompt="$1"
  if [ "$ASSUME_YES" = true ] || [ ! -t 0 ]; then
    return 0
  fi
  if command -v gum >/dev/null 2>&1; then
    gum confirm "$prompt"
  else
    read -rp "$prompt [y/N] " -r reply
    [[ "$reply" =~ ^[Yy]$ ]]
  fi
}

if ! command -v sqlite3 >/dev/null 2>&1; then
  err "sqlite3 CLI is not installed locally. Please install sqlite3 first."
  exit 1
fi

# ── 1. Export from production ────────────────────────────────────────────────
if [ "$IMPORT_ONLY" = false ]; then
  mkdir -p "$(dirname "$DUMP_FILE")"
  TEMP_DUMP_FILE="${DUMP_FILE}.tmp.$$"

  SSH_OPTS=(-o StrictHostKeyChecking=accept-new)
  if [ -n "${PRODUCTION_SSH_OPTS:-}" ]; then
    read -ra extra_opts <<< "$PRODUCTION_SSH_OPTS"
    SSH_OPTS+=("${extra_opts[@]}")
  fi

  step "Locating '$PROJECT_NAME-$SERVICE_NAME' container on $HOST..."
  CONTAINER_ID=$(
    ssh "${SSH_OPTS[@]}" "$USER@$HOST" \
      "docker ps --format '{{.ID}}\t{{.Names}}' | awk -F'\t' -v proj=\"${PROJECT_NAME}\" -v svc=\"${SERVICE_NAME}\" '\$2 ~ proj \"-.*-\" svc \"-[0-9]+\$\" {print \$1; exit}'"
  )

  if [ -z "$CONTAINER_ID" ]; then
    err "No container found matching '${PROJECT_NAME}-*-${SERVICE_NAME}' on $HOST"
    echo "Available containers:"
    ssh "${SSH_OPTS[@]}" "$USER@$HOST" "docker ps --format '{{.Names}}'"
    exit 1
  fi
  ok "Found core container: $CONTAINER_ID"

  step "Resolving remote SQLite database path..."
  REMOTE_DB_PATH=$(
    ssh "${SSH_OPTS[@]}" "$USER@$HOST" \
      "docker exec $CONTAINER_ID sh -c 'echo \"\${DB_NAME:-storage/production.sqlite3}\"'"
  )
  log "Remote database: $REMOTE_DB_PATH"

  step "Exporting SQL statements from production container..."
  ssh "${SSH_OPTS[@]}" "$USER@$HOST" \
    "docker exec $CONTAINER_ID sqlite3 \"$REMOTE_DB_PATH\" .dump" > "$TEMP_DUMP_FILE"

  if [ ! -s "$TEMP_DUMP_FILE" ]; then
    err "Exported SQL dump is empty."
    rm -f "$TEMP_DUMP_FILE"
    exit 1
  fi

  if ! grep -q "COMMIT;" "$TEMP_DUMP_FILE"; then
    err "Exported SQL dump seems incomplete (no COMMIT statement found)."
    rm -f "$TEMP_DUMP_FILE"
    exit 1
  fi

  mv "$TEMP_DUMP_FILE" "$DUMP_FILE"
  ln -sfn "$(basename "$DUMP_FILE")" "$LATEST_SYMLINK"
  ok "Saved production SQL to $DUMP_FILE ($(du -h "$DUMP_FILE" | cut -f1))"
  warn "Dump contains production data. Keep local under core/storage/ and NEVER commit to git!"
  log "Updated symlink: $LATEST_SYMLINK -> $(basename "$DUMP_FILE")"
else
  if [ ! -f "$DUMP_FILE" ]; then
    err "Dump file does not exist: $DUMP_FILE"
    exit 1
  fi
  ok "Using existing SQL file: $DUMP_FILE ($(du -h "$DUMP_FILE" | cut -f1))"
fi

# ── 2. Import into local database ───────────────────────────────────────────
if [ "$EXPORT_ONLY" = true ]; then
  ok "Export finished (--export-only specified). Skipped local import."
  exit 0
fi

if ! confirm_prompt "Overwrite local database ($(basename "$LOCAL_DB")) with imported data?"; then
  warn "Import aborted by user. SQL dump remains saved at $DUMP_FILE"
  exit 0
fi

mkdir -p "$(dirname "$LOCAL_DB")"

# Backup existing local DB if non-empty
if [ -f "$LOCAL_DB" ] && [ -s "$LOCAL_DB" ]; then
  BACKUP_FILE="${LOCAL_DB}.bak.$(date +%Y%m%d_%H%M%S)"
  step "Backing up current local database to $(basename "$BACKUP_FILE")..."
  cp "$LOCAL_DB" "$BACKUP_FILE"
fi

step "Importing SQL into local database ($LOCAL_DB)..."
rm -f "$LOCAL_DB" "${LOCAL_DB}-wal" "${LOCAL_DB}-shm"

sqlite3 "$LOCAL_DB" < "$DUMP_FILE"
ok "Imported SQL into $LOCAL_DB"

step "Preparing local Rails database (schemas & secondary DBs)..."
(cd "$RAILS_DIR" && bin/rails db:prepare)
ok "Database sync complete!"
