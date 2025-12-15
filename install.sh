#!/bin/sh
set -e

# envdiff installer
# Usage: curl -fsSL https://raw.githubusercontent.com/GBerghoff/envdiff/main/install.sh | sh

REPO="GBerghoff/envdiff"
BINARY="envdiff"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
    darwin) OS="darwin" ;;
    linux) OS="linux" ;;
    mingw*|msys*|cygwin*) OS="windows" ;;
    *) echo "Unsupported OS: $OS" && exit 1 ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH" && exit 1 ;;
esac

# Get latest version
VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$VERSION" ]; then
    echo "Failed to get latest version"
    exit 1
fi

echo "Installing ${BINARY} ${VERSION} (${OS}/${ARCH})..."

# Download
EXT="tar.gz"
[ "$OS" = "windows" ] && EXT="zip"

URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY}_${VERSION#v}_${OS}_${ARCH}.${EXT}"
TMPDIR=$(mktemp -d)
ARCHIVE="${TMPDIR}/${BINARY}.${EXT}"

echo "Downloading ${URL}..."
curl -fsSL "$URL" -o "$ARCHIVE"

# Extract
cd "$TMPDIR"
if [ "$EXT" = "zip" ]; then
    unzip -q "$ARCHIVE"
else
    tar -xzf "$ARCHIVE"
fi

# Install
if [ -w "$INSTALL_DIR" ]; then
    mv "${BINARY}" "${INSTALL_DIR}/"
else
    echo "Installing to ${INSTALL_DIR} (requires sudo)..."
    sudo mv "${BINARY}" "${INSTALL_DIR}/"
fi

# Cleanup
rm -rf "$TMPDIR"

echo "Successfully installed ${BINARY} to ${INSTALL_DIR}/${BINARY}"
echo "Run '${BINARY} --help' to get started"
