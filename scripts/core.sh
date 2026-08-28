#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RAILS_BIN="$ROOT_DIR/core/bin/rails"
cd "$ROOT_DIR/core"

OPTIONS=(
  "db:reset + db:seed"
  "log:clear + tmp:clear"
  "create-vapid-key"
)

choices="$(gum choose --header="Select core maintenance · Space to toggle · Enter to run" --no-limit "${OPTIONS[@]}")"

if [ -z "${choices:-}" ]; then
  echo "Nothing selected. Aborted."
  exit 0
fi

if ! gum confirm "Run selected core maintenance tasks now?"; then
  echo "Aborted."
  exit 0
fi

while IFS= read -r line; do
  [ -z "$line" ] && continue
  case "$line" in
    "db:reset + db:seed")
      # db:reset loads checked-in schemas (primary + queue + cable); do not use db:migrate here —
      # it dumps schemas and can overwrite queue/cable/cache_schema.rb when those DBs are empty.
      ruby "$RAILS_BIN" db:reset
      ;;
    "log:clear + tmp:clear")
      ruby "$RAILS_BIN" log:clear
      ruby "$RAILS_BIN" tmp:clear
      ;;
    "create-vapid-key")
      ruby script/create-vapid-key.rb
      ;;
    *)
      echo "Unknown selection: $line"
      exit 2
      ;;
  esac
done <<< "$choices"

