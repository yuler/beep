#!/usr/bin/env bash
# Show versions for all Beep components.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

gum_header() {
  if command -v gum >/dev/null 2>&1; then
    gum style --border double --padding "0 2" --border-foreground 212 "$1"
  else
    printf "\n=== %s ===\n" "$1"
  fi
}

gum_info() {
  if command -v gum >/dev/null 2>&1; then
    gum log --level info "$@"
  else
    echo "[info] $*"
  fi
}

gum_warn() {
  if command -v gum >/dev/null 2>&1; then
    gum log --level warn "$@"
  else
    echo "[warn] $*"
  fi
}

gum_err() {
  if command -v gum >/dev/null 2>&1; then
    gum log --level error "$@"
  else
    echo "[error] $*" >&2
  fi
}

# ── Canonical version from VERSION file ──
VERSION_FILE="$ROOT/VERSION"
if [[ ! -f "$VERSION_FILE" ]]; then
  gum_err "VERSION file not found at $VERSION_FILE"
  exit 1
fi
canonical=$(tr -d '[:space:]' < "$VERSION_FILE")

gum_header "Beep Monorepo Versions"
if command -v gum >/dev/null 2>&1; then
  gum style --margin "1 0 0 0" --foreground 212 --bold "Canonical version: $canonical"
else
  printf "\nCanonical version: %s\n" "$canonical"
fi

# ── Read component versions ──

declare -A versions
MISSING="N/A"

# CLI (Go internal/version/version.go)
cli_ver=$(sed -n 's/^[[:space:]]*Version[[:space:]]*=[[:space:]]*"\(.*\)"/\1/p' apps/cli/internal/version/version.go) || true
versions["CLI"]="${cli_ver:-$MISSING}"

# Web (apps/web/package.json)
if [[ -f "apps/web/package.json" ]]; then
  web_ver=$(node -e "console.log(JSON.parse(require('fs').readFileSync('apps/web/package.json')).version || '')" 2>/dev/null) || true
  versions["Web"]="${web_ver:-$MISSING}"
else
  versions["Web"]="$MISSING"
fi

# Web Version File (apps/web/public/version.json)
if [[ -f "apps/web/public/version.json" ]]; then
  web_pub_ver=$(node -e "console.log(JSON.parse(require('fs').readFileSync('apps/web/public/version.json')).version || '')" 2>/dev/null) || true
  versions["Web Public"]="${web_pub_ver:-$MISSING}"
else
  versions["Web Public"]="$MISSING"
fi

# Root package.json
if [[ -f "package.json" ]]; then
  root_pkg_ver=$(node -e "console.log(JSON.parse(require('fs').readFileSync('package.json')).version || '')" 2>/dev/null) || true
  versions["Root Package"]="${root_pkg_ver:-$MISSING}"
else
  versions["Root Package"]="$MISSING"
fi

# ── Print table ──

printf '\n'
printf '  \033[1;35m%-16s %-12s %s\033[0m\n' "Component" "Version" "Status"
printf '  %-16s %-12s %s\n' "────────────────" "────────────" "──────"
for component in "CLI" "Web" "Web Public" "Root Package"; do
  ver="${versions[$component]}"
  if [[ "$ver" == "$MISSING" ]]; then
    printf '  \033[0;90m%-16s %-12s missing\033[0m\n' "$component" "$ver"
  elif [[ "$ver" == "$canonical" ]]; then
    printf '  \033[1;37m%-16s %-12s \033[1;32mOK\033[0m\n' "$component" "$ver"
  else
    printf '  \033[1;33m%-16s %-12s \033[1;31mmismatch\033[0m\n' "$component" "$ver"
  fi
done
printf '\n'

# ── Git tag ──

if git describe --tags --abbrev=0 >/dev/null 2>&1; then
  latest_tag=$(git describe --tags --abbrev=0)
  if [[ "$latest_tag" == "v$canonical" ]]; then
    gum_info "Latest git tag: $latest_tag (matches VERSION)"
  else
    gum_warn "Latest git tag: $latest_tag (VERSION is $canonical)"
  fi
else
  gum_info "No git tags found."
fi
