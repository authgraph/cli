#!/bin/sh
# Authgraph CLI installer
# Usage: curl -fsSL https://get.authgraph.dev | sh
set -e

REPO="authgraph/cli"
BINARY="authgraph"
INSTALL_DIR="/usr/local/bin"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    linux|darwin) ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Get latest version
VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"v([^"]+)".*/\1/')
if [ -z "$VERSION" ]; then
    echo "Failed to determine latest version"
    exit 1
fi

echo "Installing authgraph v${VERSION} (${OS}/${ARCH})..."

# Download
URL="https://github.com/$REPO/releases/download/v${VERSION}/${BINARY}_${OS}_${ARCH}.tar.gz"
TMP_DIR=$(mktemp -d)
curl -fsSL "$URL" -o "$TMP_DIR/authgraph.tar.gz"

# Extract
tar -xzf "$TMP_DIR/authgraph.tar.gz" -C "$TMP_DIR"

# Install
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
else
    sudo mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
fi
chmod +x "$INSTALL_DIR/$BINARY"

# Cleanup
rm -rf "$TMP_DIR"

echo "✓ authgraph v${VERSION} installed to $INSTALL_DIR/$BINARY"
echo ""
echo "Get started:"
echo "  authgraph login"
echo "  authgraph --help"
