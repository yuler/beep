#!/usr/bin/env bash
set -euo pipefail

# Colorized logging helpers that use gum when available, otherwise plain echo.
log()   { gum style --foreground 212 "✦ $*"; }
ok()    { gum style --foreground 78  "✔ $*"; }
warn()  { gum style --foreground 227 "⚠ $*"; }
step()  { gum style --foreground 99  "▶ $*"; }

if ! command -v gum >/dev/null 2>&1; then
  warn "gum not found; falling back to plain output"
  log()  { echo "✦ $*"; }
  ok()   { echo "✔ $*"; }
  warn() { echo "⚠ $*"; }
  step() { echo "▶ $*"; }
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

gum spin --title "Installing toolchain (mise)…" -- mise install || true

step "Installing Ruby gems"
cd core
if command -v bundle >/dev/null 2>&1; then
  gum spin --title "bundle install…" -- bundle install
else
  warn "bundler missing, run 'mise install' first"
fi

cd "$ROOT_DIR"
step "Installing Node packages"
if command -v pnpm >/dev/null 2>&1; then
  gum spin --title "pnpm install…" -- pnpm install
else
  warn "pnpm missing, skipping"
fi

if [ ! -f .env ]; then
  step "Creating .env from .env.example"
  cp .env.example .env
  ok "Created .env"
else
  ok ".env already exists"
fi

cd core
step "Preparing database"
gum spin --title "rails db:prepare…" -- ./bin/rails db:prepare

if gum confirm "Reset database (rails db:reset)?"; then
  gum spin --title "rails db:reset…" -- ./bin/rails db:reset
fi

cd "$ROOT_DIR"
ok "Setup complete! Run 'mise run dev' to start."
