#!/usr/bin/env bash
# Unlink a Beep workspace symlink from the repository root.
#
# Usage:
#   ./scripts/cli-workspace-unlink.sh              # unlinks .beep
#   ./scripts/cli-workspace-unlink.sh .beep        # unlinks .beep
#   ./scripts/cli-workspace-unlink.sh .beep.local  # unlinks .beep.local

set -euo pipefail

if [ "$#" -gt 1 ]; then
  echo "Error: Only a single parameter is accepted" >&2
  echo "Usage: $0 [name]" >&2
  exit 1
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

target="${1:-.beep}"

case "$target" in
  ~/*|/*)
    link_name="$(basename "$target")"
    ;;
  *)
    link_name="$target"
    ;;
esac

if [ -L "$link_name" ]; then
  rm "$link_name"
  echo "Unlinked $link_name"
elif [ -e "$link_name" ]; then
  echo "Error: $link_name exists but is not a symlink" >&2
  exit 1
else
  echo "$link_name is not linked"
fi
