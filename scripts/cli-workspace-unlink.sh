#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

if [ -L .beep ]; then
  rm .beep
  echo "Unlinked .beep"
elif [ -e .beep ]; then
  echo "Error: .beep exists but is not a symlink" >&2
  exit 1
else
  echo ".beep is not linked"
fi
