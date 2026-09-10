#!/bin/sh
# Beep CLI Installer Script
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/yuler/beep/main/install.sh | sh
# Options (via environment variables):
#   VERSION:     Version to install (e.g. v0.1.0 or latest). Default: latest
#   INSTALL_DIR: Target installation directory. Default: /usr/local/bin or ~/.local/bin
#   REPO:        GitHub repository (owner/name). Default: yuler/beep

set -eu

REPO="${REPO:-yuler/beep}"
BINARY_NAME="beep"
REQUESTED_VERSION="${VERSION:-latest}"

# Color output helpers
setup_colors() {
  if [ -t 1 ] && [ -z "${NO_COLOR:-}" ]; then
    BOLD="$(printf '\033[1m')"
    DIM="$(printf '\033[2m')"
    GREEN="$(printf '\033[0;32m')"
    CYAN="$(printf '\033[0;36m')"
    RED="$(printf '\033[0;31m')"
    YELLOW="$(printf '\033[0;33m')"
    RESET="$(printf '\033[0m')"
  else
    BOLD=""
    DIM=""
    GREEN=""
    CYAN=""
    RED=""
    YELLOW=""
    RESET=""
  fi
}

info() {
  printf "%sinfo%s %s\n" "${CYAN}" "${RESET}" "$1" >&2
}

success() {
  printf "%s✓%s %s%s%s\n" "${GREEN}" "${RESET}" "${BOLD}" "$1" "${RESET}"
}

warn() {
  printf "%swarning%s %s\n" "${YELLOW}" "${RESET}" "$1" >&2
}

error() {
  # %b expands backslash escapes (\n) in the message argument
  printf "%serror%s %b\n" "${RED}" "${RESET}" "$1" >&2
  exit 1
}

# HTTP fetch helper using curl (guaranteed present — the installer itself is piped from curl)
http_get() {
  curl -fsSL "$1"
}

# Download file to destination
http_download() {
  curl -fSL --progress-bar "$1" -o "$2"
}

detect_os() {
  _os="$(uname -s)"
  case "$_os" in
    Linux|linux)
      echo "linux"
      ;;
    Darwin|darwin)
      echo "darwin"
      ;;
    MINGW*|MSYS*|CYGWIN*|Windows_NT)
      echo "windows"
      ;;
    *)
      error "Unsupported operating system: $_os"
      ;;
  esac
}

detect_arch() {
  _arch="$(uname -m)"
  case "$_arch" in
    x86_64|amd64)
      echo "amd64"
      ;;
    aarch64|arm64|armv8*)
      echo "arm64"
      ;;
    *)
      error "Unsupported CPU architecture: $_arch"
      ;;
  esac
}

resolve_version() {
  _target="$1"
  if [ "$_target" != "latest" ]; then
    case "$_target" in
      v*)
        echo "$_target"
        ;;
      *)
        echo "v$_target"
        ;;
    esac
    return
  fi

  info "Fetching latest release version for ${REPO}..."
  _api_url="https://api.github.com/repos/${REPO}/releases/latest"

  # Try GitHub API first
  _tag_response="$(http_get "$_api_url" 2>/dev/null || true)"
  if [ -n "$_tag_response" ]; then
    _tag="$(printf '%s\n' "$_tag_response" | grep '"tag_name":' | sed -e 's/.*"tag_name": *"//' -e 's/".*//')"
    if [ -n "$_tag" ]; then
      echo "$_tag"
      return
    fi
  fi

  # Fallback to redirect resolution if API was rate-limited
  _redirect_url="$(curl -fsSL -o /dev/null -w "%{url_effective}" "https://github.com/${REPO}/releases/latest" 2>/dev/null || true)"
  _tag="${_redirect_url##*/}"
  if [ -n "$_tag" ] && [ "$_tag" != "latest" ]; then
    echo "$_tag"
    return
  fi

  error "Failed to resolve latest version from GitHub releases. Please specify VERSION=vX.Y.Z manually."
}

compute_sha256() {
  _file="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$_file" | awk '{print $1}'
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$_file" | awk '{print $1}'
  elif command -v openssl >/dev/null 2>&1; then
    openssl dgst -sha256 "$_file" | awk '{print $NF}'
  else
    echo ""
  fi
}

determine_install_dir() {
  if [ -n "${INSTALL_DIR:-}" ]; then
    echo "$INSTALL_DIR"
    return
  fi

  if [ -w "/usr/local/bin" ]; then
    echo "/usr/local/bin"
  elif [ "$(id -u)" -eq 0 ]; then
    echo "/usr/local/bin"
  else
    echo "$HOME/.local/bin"
  fi
}

main() {
  setup_colors

  os="$(detect_os)"
  arch="$(detect_arch)"

  raw_version="$(resolve_version "$REQUESTED_VERSION")"

  # Strip leading 'v' for asset archive filenames (e.g. v0.1.0 -> 0.1.0)
  version="${raw_version#v}"

  ext="tar.gz"
  bin_file="${BINARY_NAME}"
  if [ "$os" = "windows" ]; then
    ext="zip"
    bin_file="${BINARY_NAME}.exe"
  fi

  archive_name="${BINARY_NAME}_${version}_${os}_${arch}.${ext}"
  download_url="https://github.com/${REPO}/releases/download/${raw_version}/${archive_name}"
  checksum_url="https://github.com/${REPO}/releases/download/${raw_version}/checksums.txt"

  install_dir="$(determine_install_dir)"

  printf "\n"
  printf "  %s %s\n" "${BOLD}beep${RESET}" "${DIM}installer${RESET}"
  printf "  Version:      %s%s%s\n" "${CYAN}" "${raw_version}" "${RESET}"
  printf "  Platform:     %s%s/%s%s\n" "${CYAN}" "${os}" "${arch}" "${RESET}"
  printf "  Destination:  %s%s%s\n\n" "${CYAN}" "${install_dir}" "${RESET}"

  tmpdir="$(mktemp -d 2>/dev/null || mktemp -d -t 'beep-install')"
  cleanup() {
    rm -rf "$tmpdir"
  }
  trap cleanup EXIT
  trap 'exit 1' INT TERM

  info "Downloading ${archive_name}..."
  http_download "$download_url" "$tmpdir/$archive_name"

  info "Downloading checksums.txt..."
  if ! http_get "$checksum_url" > "$tmpdir/checksums.txt" 2>/dev/null; then
    error "Could not fetch checksums.txt — refusing to install an unverified binary."
  fi

  # Look up hash by exact archive filename in checksums.txt (matches '<hash>  <filename>' or '<hash> *<filename>')
  expected_hash="$(awk -v name="$archive_name" '$2 == name || $2 == "*"name { print $1; exit }' "$tmpdir/checksums.txt")"
  if [ -z "$expected_hash" ]; then
    error "Could not find checksum for ${archive_name} in checksums.txt"
  fi

  actual_hash="$(compute_sha256 "$tmpdir/$archive_name")"
  if [ -z "$actual_hash" ]; then
    error "No sha256 tool found (sha256sum, shasum, openssl) — refusing to install an unverified binary."
  fi

  if [ "$expected_hash" != "$actual_hash" ]; then
    error "Checksum verification failed!\nExpected: $expected_hash\nActual:   $actual_hash"
  fi
  short_hash="$(printf "%.12s" "$actual_hash")"
  info "Checksum verified (${short_hash}...)"

  info "Extracting archive..."
  if [ "$ext" = "tar.gz" ]; then
    tar -xzf "$tmpdir/$archive_name" -C "$tmpdir"
  elif [ "$ext" = "zip" ]; then
    if command -v unzip >/dev/null 2>&1; then
      unzip -q -o "$tmpdir/$archive_name" -d "$tmpdir"
    else
      error "unzip command is required to extract Windows zip archives."
    fi
  fi

  if [ ! -f "$tmpdir/$bin_file" ]; then
    error "Extracted archive did not contain expected binary: $bin_file"
  fi

  chmod +x "$tmpdir/$bin_file"

  # Ensure destination directory exists
  if [ ! -d "$install_dir" ]; then
    parent_dir="$(dirname "$install_dir")"
    if [ ! -w "$parent_dir" ] && command -v sudo >/dev/null 2>&1 && [ "$(id -u)" -ne 0 ]; then
      sudo mkdir -p "$install_dir"
    else
      mkdir -p "$install_dir"
    fi
  fi

  info "Installing ${bin_file} to ${install_dir}..."
  if [ -w "$install_dir" ]; then
    mv "$tmpdir/$bin_file" "$install_dir/$bin_file"
  elif command -v sudo >/dev/null 2>&1 && [ "$(id -u)" -ne 0 ]; then
    sudo mv "$tmpdir/$bin_file" "$install_dir/$bin_file"
  else
    error "Cannot write to ${install_dir}. Please run with sudo or set INSTALL_DIR to a writable directory."
  fi

  success "Successfully installed ${BINARY_NAME} to ${install_dir}/${bin_file}"

  # Check if installed directory is in PATH
  case ":$PATH:" in
    *":$install_dir:"*) ;;
    *)
      printf "\n"
      warn "${install_dir} is not in your \$PATH."
      printf "  Add it to your profile by running:\n"
      printf "    %sexport PATH=\"%s:\$PATH\"%s\n\n" "${BOLD}" "${install_dir}" "${RESET}"
      ;;
  esac

  if command -v "$install_dir/$bin_file" >/dev/null 2>&1; then
    printf "  Installed version: "
    "$install_dir/$bin_file" version || true
  fi
  printf "\n"
}

main "$@"
