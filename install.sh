#!/bin/bash
set -euo pipefail

REPO="hsk-kr/licokit"
BINARY="licokit-darwin-arm64"
INSTALL_DIR="$HOME/.local/bin"
INSTALL_PATH="$INSTALL_DIR/licokit"

if [[ "$(uname -s)" != "Darwin" || "$(uname -m)" != "arm64" ]]; then
  echo "LicoKit currently supports Apple Silicon macOS only." >&2
  exit 1
fi

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

release_json="$tmp_dir/release.json"
curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" -o "$release_json"

asset_url="$(sed -n 's/.*"browser_download_url": "\([^"]*\/'"$BINARY"'\)".*/\1/p' "$release_json" | head -1)"
checksum_url="$(sed -n 's/.*"browser_download_url": "\([^"]*\/'"$BINARY"'.sha256\)".*/\1/p' "$release_json" | head -1)"

if [[ -z "$asset_url" || -z "$checksum_url" ]]; then
  echo "Could not find the ${BINARY} release and checksum." >&2
  exit 1
fi

echo "Downloading the latest LicoKit release..."
curl -fsSL "$asset_url" -o "$tmp_dir/$BINARY"
curl -fsSL "$checksum_url" -o "$tmp_dir/$BINARY.sha256"

(
  cd "$tmp_dir"
  shasum -a 256 -c "$BINARY.sha256"
)

mkdir -p "$INSTALL_DIR"
install -m 0755 "$tmp_dir/$BINARY" "$INSTALL_PATH"
xattr -d com.apple.quarantine "$INSTALL_PATH" 2>/dev/null || true

echo "Installed to $INSTALL_PATH"
exec "$INSTALL_PATH" "$@"
