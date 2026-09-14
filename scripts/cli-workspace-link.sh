#!/usr/bin/env bash
# Symlink a Beep workspace directory to the repository root.
#
# Usage:
#   ./scripts/cli-workspace-link.sh              # links ~/.beep -> .beep
#   ./scripts/cli-workspace-link.sh .beep        # links ~/.beep -> .beep
#   ./scripts/cli-workspace-link.sh .beep.local  # links ~/.beep.local -> .beep.local

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
  ~/*)
    workspace="$HOME/${target#~/}"
    link_name="$(basename "$workspace")"
    ;;
  ~)
    workspace="$HOME"
    link_name=".beep"
    ;;
  /*)
    workspace="$target"
    link_name="$(basename "$workspace")"
    ;;
  *)
    link_name="$target"
    workspace="$HOME/$target"
    ;;
esac

mkdir -p "$workspace"
ln -sfn "$workspace" "$link_name"
echo "Linked $workspace -> $link_name"
