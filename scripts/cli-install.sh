#!/usr/bin/env bash
# Install the current tree's beep binary for local dogfooding (no GitHub release).
#
# Usage:
#   mise cli:install
#   INSTALL_DIR=$HOME/.local/bin ./scripts/cli-install.sh
#
# Destination (first match):
#   1. INSTALL_DIR (any writable dir; refuses Homebrew Cellar/bottle paths)
#   2. Directory of an existing `beep` on PATH (skip this repo's bin/; refuses Homebrew)
#   3. ~/.local/bin
#
# Does not use sudo. After install, restart daemons: beep service restart

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_BIN="$ROOT_DIR/bin"
SRC="$REPO_BIN/beep"

err() {
  printf 'error: %s\n' "$1" >&2
  exit 1
}

info() {
  printf '%s\n' "$1"
}

resolve_path() {
  local p="$1"
  if command -v realpath >/dev/null 2>&1; then
    if realpath "$p" 2>/dev/null; then
      return 0
    fi
  fi
  if command -v readlink >/dev/null 2>&1; then
    if readlink -f "$p" 2>/dev/null; then
      return 0
    fi
  fi
  if [ -d "$p" ]; then
    (cd "$p" && pwd)
    return 0
  fi
  if [ -e "$p" ]; then
    local dir base
    dir="$(cd "$(dirname "$p")" && pwd)"
    base="$(basename "$p")"
    printf '%s/%s\n' "$dir" "$base"
    return 0
  fi
  printf '%s\n' "$p"
}

is_homebrew_path() {
  local lower
  lower="$(printf '%s' "$1" | tr '[:upper:]' '[:lower:]')"
  case "$lower" in
    */cellar/*|*/homebrew/*|*/.linuxbrew/*) return 0 ;;
  esac
  return 1
}

# Standard Homebrew shim directories (automatic destination selection only).
is_brew_managed_bin() {
  local dir resolved
  dir="$1"
  resolved="$(resolve_path "$dir")"
  case "$resolved" in
    /opt/homebrew/bin|/usr/local/bin|/home/linuxbrew/.linuxbrew/bin) return 0 ;;
  esac
  return 1
}

# First `beep` on PATH that is not this repo's build output.
existing_beep() {
  local p dir
  while IFS= read -r p; do
    [ -n "$p" ] || continue
    dir="$(cd "$(dirname "$p")" && pwd)"
    if [ "$dir" = "$REPO_BIN" ]; then
      continue
    fi
    printf '%s\n' "$p"
    return 0
  done < <(type -ap beep 2>/dev/null || true)
  return 1
}

default_install_dir() {
  if [ -n "${INSTALL_DIR:-}" ]; then
    mkdir -p "$INSTALL_DIR" || err "cannot create $INSTALL_DIR"
    printf '%s\n' "$(cd "$INSTALL_DIR" && pwd)"
    return
  fi

  local existing resolved dest
  if existing="$(existing_beep)"; then
    resolved="$(resolve_path "$existing")"
    dest="$(cd "$(dirname "$existing")" && pwd)"
    if is_brew_managed_bin "$dest" || is_homebrew_path "$dest" || is_homebrew_path "$resolved"; then
      err "existing beep looks like a Homebrew install ($existing).
Refusing to overwrite. Use brew, or set INSTALL_DIR=~/.local/bin (or another writable directory)."
    fi
    printf '%s\n' "$dest"
    return
  fi

  if [ -d "$HOME/.local/bin" ] || mkdir -p "$HOME/.local/bin" 2>/dev/null; then
    if [ -w "$HOME/.local/bin" ]; then
      printf '%s\n' "$HOME/.local/bin"
      return
    fi
  fi

  err "no writable install directory.
Create ~/.local/bin and add it to PATH, or run:
  INSTALL_DIR=/path/to/dir mise cli:install"
}

[ -f "$SRC" ] || err "missing $SRC — run: mise cli:build"
[ -x "$SRC" ] || err "$SRC is not executable"

dest="$(default_install_dir)"
target="$dest/beep"

if is_homebrew_path "$dest" || is_homebrew_path "$target"; then
  err "refusing to install into a Homebrew bottle path ($dest).
Set INSTALL_DIR to a writable directory such as ~/.local/bin."
fi

mkdir -p "$dest" || err "cannot create $dest"
[ -w "$dest" ] || err "cannot write to $dest (no sudo). Set INSTALL_DIR to a writable directory."

tmp="$(mktemp "$dest/.beep-install.XXXXXX")"
trap 'rm -f "$tmp"' EXIT
cp "$SRC" "$tmp"
chmod 755 "$tmp"
mv -f "$tmp" "$target"
trap - EXIT

info "Installed $target"
"$target" version
info "Restart running daemons with: beep service restart"
