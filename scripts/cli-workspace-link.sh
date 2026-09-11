#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

workspace="${1:-}"
link_name="${2:-.beep}"

if [ -z "$workspace" ]; then
  if [ "$link_name" = ".beep.local" ]; then
    workspace="$HOME/.beep.local"
  else
    workspace="$HOME/.beep"
  fi
fi

case "$workspace" in
  ~/*) workspace="$HOME/${workspace#~/}" ;;
  ~) workspace="$HOME" ;;
esac

mkdir -p "$workspace"
ln -sfn "$workspace" "$link_name"
echo "Linked $workspace -> $link_name"
