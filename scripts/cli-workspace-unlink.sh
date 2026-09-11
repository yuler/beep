#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

link_name="${1:-.beep}"

if [ -L "$link_name" ]; then
  rm "$link_name"
  echo "Unlinked $link_name"
elif [ -e "$link_name" ]; then
  echo "Error: $link_name exists but is not a symlink" >&2
  exit 1
else
  echo "$link_name is not linked"
fi
