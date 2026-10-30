#!/usr/bin/env bash
# Symlink apps/omarchy-plugin to ~/.config/omarchy/plugins/beep
#
# Usage:
#   ./scripts/omarchy-plugin-link.sh

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PLUGIN_DIR="$ROOT_DIR/apps/omarchy-plugin"
TARGET="$HOME/.config/omarchy/plugins/beep"

if [ -L "$TARGET" ]; then
  current_target="$(readlink "$TARGET")"
  if [ "$current_target" = "$PLUGIN_DIR" ] || [ "$(readlink -f "$TARGET" 2>/dev/null)" = "$(readlink -f "$PLUGIN_DIR" 2>/dev/null)" ]; then
    echo "Symlink already points to $PLUGIN_DIR"
    exit 0
  fi
fi

if [ -e "$TARGET" ] || [ -L "$TARGET" ]; then
  echo "Error: $TARGET already exists and does not point to $PLUGIN_DIR" >&2
  exit 1
fi

mkdir -p "$(dirname "$TARGET")"
ln -s "$PLUGIN_DIR" "$TARGET"
echo "Linked $TARGET -> $PLUGIN_DIR"
