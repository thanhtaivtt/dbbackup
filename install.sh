#!/bin/sh
set -e

REPO="thanhtaivtt/dbbackup"
INSTALL_DIR="/usr/local/bin"
TMP_DIR=$(mktemp -d)

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

echo "Detecting system: ${OS}/${ARCH}"

# Get latest version
VERSION=$(curl -sI "https://github.com/${REPO}/releases/latest" \
  | grep -i "^location:" | sed 's|.*/tag/||' | tr -d '\r')

if [ -z "$VERSION" ]; then
  echo "Error: Could not determine latest version"
  exit 1
fi

echo "Latest version: ${VERSION}"

# Download
FILE="dbbackup_${VERSION#v}_${OS}_${ARCH}"
if [ "$OS" = "windows" ]; then
  FILE="${FILE}.zip"
else
  FILE="${FILE}.tar.gz"
fi

URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILE}"
echo "Downloading ${URL}..."
curl -sL "$URL" -o "${TMP_DIR}/${FILE}"

# Extract
cd "$TMP_DIR"
if [ "$OS" = "windows" ]; then
  unzip -q "$FILE"
else
  tar -xzf "$FILE"
fi

# Install
if [ -w "$INSTALL_DIR" ]; then
  mv dbbackup "$INSTALL_DIR/"
else
  sudo mv dbbackup "$INSTALL_DIR/"
fi

# Cleanup
rm -rf "$TMP_DIR"

echo "✅ dbbackup ${VERSION} installed to ${INSTALL_DIR}/dbbackup"
echo ""
echo "Get started:"
echo "  dbbackup --version"
echo "  cp config.example.toml config.toml"
echo "  dbbackup backup --config config.toml"
