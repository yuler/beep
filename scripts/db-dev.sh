#!/usr/bin/env bash
# Reverse-introspect the local Rails SQLite DB into Prisma, then start Studio.
# Schema lives under .prisma/ (gitignored) and never touches Rails code.
set -euo pipefail

ok()   { gum style --foreground 78  "✔ $*"; }
warn() { gum style --foreground 227 "⚠ $*"; }
step() { gum style --foreground 99  "▶ $*"; }

if ! command -v gum >/dev/null 2>&1; then
  ok()   { echo "✔ $*"; }
  warn() { echo "⚠ $*"; }
  step() { echo "▶ $*"; }
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PRISMA_DIR="$ROOT_DIR/.prisma"
SCHEMA="$PRISMA_DIR/schema.prisma"
DB_FILE="${TARGET_DB:-$ROOT_DIR/core/storage/development.sqlite3}"
PORT="${PRISMA_STUDIO_PORT:-5555}"
# Pin Prisma 6: classic schema.prisma datasource (Prisma 7 switched to prisma.config.ts).
PRISMA=(pnpm dlx prisma@6)

if [ ! -f "$DB_FILE" ]; then
  warn "Missing $DB_FILE — run mise setup first."
  exit 1
fi

mkdir -p "$PRISMA_DIR"

# Relative to schema.prisma so the file: URL stays portable.
rel_db="$(realpath --relative-to="$PRISMA_DIR" "$DB_FILE")"

cat > "$SCHEMA" <<EOF
generator client {
  provider = "prisma-client-js"
}

datasource db {
  provider = "sqlite"
  url      = "file:${rel_db}"
}
EOF

cd "$PRISMA_DIR"
step "Pulling schema from ${DB_FILE}"
"${PRISMA[@]}" db pull --schema "$SCHEMA" --force

# Rails SQLite stores UUID PKs as blob(16), timestamps as datetime(6), and
# json columns — Prisma marks those Unsupported and @@ignore's most tables.
# Remap so Studio can browse; types are display-only and never written to Rails.
sed -i \
  -e 's/Unsupported("blob(16)")/Bytes/g' \
  -e 's/Unsupported("datetime(6)")/DateTime/g' \
  -e 's/Unsupported("json")/Json/g' \
  -e '/The underlying table does not contain a valid unique identifier/d' \
  -e 's/ @ignore//g' \
  -e '/^  @@ignore$/d' \
  "$SCHEMA"

ok "Wrote $SCHEMA (Rails SQLite types remapped for Studio)"

step "Prisma Studio → http://127.0.0.1:${PORT}"
exec "${PRISMA[@]}" studio --schema "$SCHEMA" --hostname 127.0.0.1 --port "$PORT"
