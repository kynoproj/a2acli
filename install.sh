#!/usr/bin/env bash
#
# a2acli installer
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/kynoproj/a2acli/main/install.sh | bash
#
# Environment variables:
#   A2ACLI_VERSION   Specific version to install (e.g. v0.1.0). Defaults to the
#                    latest published release.
#   INSTALL_DIR      Directory to install the binary into. Defaults to
#                    /usr/local/bin (falls back to $HOME/.local/bin if not
#                    writable).
#   BINARY_NAME      Name of the installed binary. Defaults to a2acli.
#
set -euo pipefail

REPO="kynoproj/a2acli"
BINARY_NAME="${BINARY_NAME:-a2acli}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

log()  { printf '\033[1;34m==>\033[0m %s\n' "$*"; }
warn() { printf '\033[1;33m==>\033[0m %s\n' "$*" >&2; }
err()  { printf '\033[1;31m==>\033[0m %s\n' "$*" >&2; exit 1; }

require() {
  command -v "$1" >/dev/null 2>&1 || err "required command not found: $1"
}

require uname
require gzip
require install
require mktemp

if command -v curl >/dev/null 2>&1; then
  DOWNLOADER="curl"
elif command -v wget >/dev/null 2>&1; then
  DOWNLOADER="wget"
else
  err "neither curl nor wget is installed"
fi

fetch() {
  local url="$1" out="$2"
  if [ "$DOWNLOADER" = "curl" ]; then
    curl -fsSL "$url" -o "$out"
  else
    wget -qO "$out" "$url"
  fi
}

fetch_stdout() {
  local url="$1"
  if [ "$DOWNLOADER" = "curl" ]; then
    curl -fsSL "$url"
  else
    wget -qO- "$url"
  fi
}

detect_os() {
  local os
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "$os" in
    linux)  echo "linux" ;;
    darwin) echo "darwin" ;;
    *) err "unsupported OS: $os" ;;
  esac
}

detect_arch() {
  local arch
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64)   echo "amd64" ;;
    aarch64|arm64)  echo "arm64" ;;
    armv7l|armv6l)  echo "arm" ;;
    ppc64le)        echo "ppc64le" ;;
    s390x)          echo "s390x" ;;
    *) err "unsupported architecture: $arch" ;;
  esac
}

resolve_version() {
  if [ -n "${A2ACLI_VERSION:-}" ]; then
    echo "$A2ACLI_VERSION"
    return
  fi
  local api_url="https://api.github.com/repos/${REPO}/releases/latest"
  local tag
  tag="$(fetch_stdout "$api_url" \
    | grep -E '"tag_name"' \
    | head -n 1 \
    | sed -E 's/.*"tag_name"[[:space:]]*:[[:space:]]*"([^"]+)".*/\1/')"
  [ -n "$tag" ] || err "failed to detect latest version from $api_url"
  echo "$tag"
}

choose_install_dir() {
  local dir="$INSTALL_DIR"
  if [ -d "$dir" ] && [ -w "$dir" ]; then
    echo "$dir"
    return
  fi
  if [ ! -d "$dir" ] && mkdir -p "$dir" 2>/dev/null; then
    echo "$dir"
    return
  fi
  if command -v sudo >/dev/null 2>&1; then
    echo "$dir"
    return
  fi
  warn "$dir is not writable and sudo is unavailable; falling back to \$HOME/.local/bin"
  mkdir -p "$HOME/.local/bin"
  echo "$HOME/.local/bin"
}

install_binary() {
  local src="$1" dest_dir="$2" dest="$2/$BINARY_NAME"
  if [ -w "$dest_dir" ]; then
    install -m 0755 "$src" "$dest"
  else
    log "Installing to $dest (requires sudo)"
    sudo install -m 0755 "$src" "$dest"
  fi
}

main() {
  local os arch version asset_url tmpdir gz_path bin_path dest_dir
  os="$(detect_os)"
  arch="$(detect_arch)"
  version="$(resolve_version)"
  asset_url="https://github.com/${REPO}/releases/download/${version}/a2acli-${os}-${arch}.gz"

  log "Installing a2acli ${version} for ${os}/${arch}"
  log "Asset: ${asset_url}"

  tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir"' EXIT

  gz_path="${tmpdir}/a2acli.gz"
  bin_path="${tmpdir}/a2acli"

  fetch "$asset_url" "$gz_path" || err "failed to download $asset_url"
  gzip -d -c "$gz_path" > "$bin_path"
  chmod +x "$bin_path"

  dest_dir="$(choose_install_dir)"
  install_binary "$bin_path" "$dest_dir"

  log "Installed ${BINARY_NAME} to ${dest_dir}/${BINARY_NAME}"

  case ":$PATH:" in
    *":${dest_dir}:"*) ;;
    *) warn "${dest_dir} is not in your PATH; add it to use ${BINARY_NAME} directly" ;;
  esac

  if command -v "$BINARY_NAME" >/dev/null 2>&1; then
    "$BINARY_NAME" version || true
  fi
}

main "$@"
