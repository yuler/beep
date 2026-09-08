#!/usr/bin/env bash
# Beep CLI Installer Script
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/yuler/beep/main/install.sh | bash
# Options (via environment variables):
#   VERSION:     Version to install (e.g. v0.1.0 or latest). Default: latest
#   INSTALL_DIR: Target installation directory. Default: /usr/local/bin or ~/.local/bin
#   REPO:        GitHub repository (owner/name). Default: yuler/beep

set -euo pipefail

REPO="${REPO:-yuler/beep}"
BINARY_NAME="beep"
REQUESTED_VERSION="${VERSION:-latest}"

# Color output helpers
setup_colors() {
  if [[ -t 1 ]] && [[ -z "${NO_COLOR:-}" ]]; then
    BOLD="\033[1m"
    DIM="\033[2m"
    GREEN="\033[0;32m"
    CYAN="\033[0;36m"
    RED="\033[0;31m"
    YELLOW="\033[0;33m"
    RESET="\033[0m"
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
  printf "${CYAN}info${RESET} %s\n" "$1"
}

success() {
  printf "${GREEN}✓${RESET} ${BOLD}%s${RESET}\n" "$1"
}

warn() {
  printf "${YELLOW}warning${RESET} %s\n" "$1"
}

error() {
  printf "${RED}error${RESET} %s\n" "$1" >&2
  exit 1
}

# HTTP fetch helper using curl or wget
http_get() {
  local url="$1"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url"
  elif command -v wget >/dev/null 2>&1; then
    wget -qO- "$url"
  else
    error "Neither curl nor wget was found on your system. Please install one to proceed."
  fi
}

# Download file to destination
http_download() {
  local url="$1"
  local dest="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fSL --progress-bar "$url" -o "$dest"
  elif command -v wget >/dev/null 2>&1; then
    wget --show-progress -qO "$dest" "$url"
  else
    error "Neither curl nor wget was found on your system."
  fi
}

detect_os() {
  local os
  os="$(uname -s)"
  case "$os" in
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
      error "Unsupported operating system: $os"
      ;;
  esac
}

detect_arch() {
  local arch
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64)
      echo "amd64"
      ;;
    aarch64|arm64|armv8*)
      echo "arm64"
      ;;
    *)
      error "Unsupported CPU architecture: $arch"
      ;;
  esac
}

resolve_version() {
  local target="$1"
  if [[ "$target" != "latest" ]]; then
    # Return tag as-is, ensuring format
    echo "$target"
    return
  fi

  info "Fetching latest release version for ${REPO}..."
  local api_url="https://api.github.com/repos/${REPO}/releases/latest"
  local tag=""

  # Try GitHub API first
  if tag=$(http_get "$api_url" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/'); then
    if [[ -n "$tag" ]]; then
      echo "$tag"
      return
    fi
  fi

  # Fallback to redirect resolution if API was rate-limited
  if command -v curl >/dev/null 2>&1; then
    local redirect_url
    redirect_url=$(curl -fsSL -o /dev/null -w "%{url_effective}" "https://github.com/${REPO}/releases/latest" 2>/dev/null || true)
    tag="${redirect_url##*/}"
    if [[ -n "$tag" && "$tag" != "latest" ]]; then
      echo "$tag"
      return
    fi
  fi

  error "Failed to resolve latest version from GitHub releases. Please specify VERSION=vX.Y.Z manually."
}

compute_sha256() {
  local file="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$file" | awk '{print $1}'
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$file" | awk '{print $1}'
  elif command -v openssl >/dev/null 2>&1; then
    openssl dgst -sha256 "$file" | awk '{print $NF}'
  else
    echo ""
  fi
}

determine_install_dir() {
  if [[ -n "${INSTALL_DIR:-}" ]]; then
    echo "$INSTALL_DIR"
    return
  fi

  if [[ -w "/usr/local/bin" ]]; then
    echo "/usr/local/bin"
  elif [[ "$EUID" -eq 0 ]]; then
    echo "/usr/local/bin"
  else
    echo "$HOME/.local/bin"
  fi
}

main() {
  setup_colors

  local os
  local arch
  os="$(detect_os)"
  arch="$(detect_arch)"

  local raw_version
  raw_version="$(resolve_version "$REQUESTED_VERSION")"

  # Strip leading 'v' for asset archive filenames (e.g. v0.1.0 -> 0.1.0)
  local version="${raw_version#v}"

  local ext="tar.gz"
  local bin_file="${BINARY_NAME}"
  if [[ "$os" == "windows" ]]; then
    ext="zip"
    bin_file="${BINARY_NAME}.exe"
  fi

  local archive_name="${BINARY_NAME}_${version}_${os}_${arch}.${ext}"
  local download_url="https://github.com/${REPO}/releases/download/${raw_version}/${archive_name}"
  local checksum_url="https://github.com/${REPO}/releases/download/${raw_version}/checksums.txt"

  local install_dir
  install_dir="$(determine_install_dir)"

  printf "\n"
  printf "  ${BOLD}%s${RESET} ${DIM}installer${RESET}\n" "beep"
  printf "  Version:      ${CYAN}%s${RESET}\n" "${raw_version}"
  printf "  Platform:     ${CYAN}%s/%s${RESET}\n" "${os}" "${arch}"
  printf "  Destination:  ${CYAN}%s${RESET}\n\n" "${install_dir}"

  local tmpdir
  tmpdir="$(mktemp -d 2>/dev/null || mktemp -d -t 'beep-install')"
  cleanup() {
    rm -rf "$tmpdir"
  }
  trap cleanup EXIT

  info "Downloading ${archive_name}..."
  http_download "$download_url" "$tmpdir/$archive_name"

  info "Downloading checksums.txt..."
  if http_get "$checksum_url" > "$tmpdir/checksums.txt" 2>/dev/null; then
    local expected_hash
    expected_hash=$(grep -E "${archive_name}\$" "$tmpdir/checksums.txt" | awk '{print $1}' || true)
    if [[ -n "$expected_hash" ]]; then
      local actual_hash
      actual_hash="$(compute_sha256 "$tmpdir/$archive_name")"
      if [[ -n "$actual_hash" ]]; then
        if [[ "$expected_hash" != "$actual_hash" ]]; then
          error "Checksum verification failed!\nExpected: $expected_hash\nActual:   $actual_hash"
        fi
        info "Checksum verified (${actual_hash:0:12}...)"
      fi
    fi
  else
    warn "Could not fetch checksums.txt, skipping verification."
  fi

  info "Extracting archive..."
  if [[ "$ext" == "tar.gz" ]]; then
    tar -xzf "$tmpdir/$archive_name" -C "$tmpdir"
  elif [[ "$ext" == "zip" ]]; then
    if command -v unzip >/dev/null 2>&1; then
      unzip -q -o "$tmpdir/$archive_name" -d "$tmpdir"
    else
      error "unzip command is required to extract Windows zip archives."
    fi
  fi

  if [[ ! -f "$tmpdir/$bin_file" ]]; then
    error "Extracted archive did not contain expected binary: $bin_file"
  fi

  chmod +x "$tmpdir/$bin_file"

  # Ensure destination directory exists
  if [[ ! -d "$install_dir" ]]; then
    if [[ ! -w "$(dirname "$install_dir")" ]] && command -v sudo >/dev/null 2>&1 && [[ "$EUID" -ne 0 ]]; then
      sudo mkdir -p "$install_dir"
    else
      mkdir -p "$install_dir"
    fi
  fi

  info "Installing ${bin_file} to ${install_dir}..."
  if [[ -w "$install_dir" ]]; then
    mv "$tmpdir/$bin_file" "$install_dir/$bin_file"
  elif command -v sudo >/dev/null 2>&1 && [[ "$EUID" -ne 0 ]]; then
    sudo mv "$tmpdir/$bin_file" "$install_dir/$bin_file"
  else
    error "Cannot write to ${install_dir}. Please run with sudo or set INSTALL_DIR to a writable directory."
  fi

  success "Successfully installed ${BINARY_NAME} to ${install_dir}/${bin_file}"

  # Check if installed directory is in PATH
  if [[ ":$PATH:" != *":$install_dir:"* ]]; then
    printf "\n"
    warn "${install_dir} is not in your \$PATH."
    printf "  Add it to your profile by running:\n"
    printf "    ${BOLD}export PATH=\"%s:\$PATH\"${RESET}\n\n" "${install_dir}"
  fi

  if command -v "$install_dir/$bin_file" >/dev/null 2>&1; then
    printf "  Installed version: "
    "$install_dir/$bin_file" version || true
  fi
  printf "\n"
}

main "$@"
