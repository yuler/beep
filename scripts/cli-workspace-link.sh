#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

workspace="${1:-}"
if [ -z "$workspace" ]; then
  workspace="$HOME/.beep"
fi
case "$workspace" in
  ~/*) workspace="$HOME/${workspace#~/}" ;;
  ~) workspace="$HOME" ;;
esac

mkdir -p "$workspace"
ln -sfn "$workspace" .beep
echo "Linked $workspace -> .beep"
