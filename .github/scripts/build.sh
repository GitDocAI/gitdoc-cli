#!/bin/bash
set -e

# Build binary
BINARY_NAME="dockitai-cli"
if [ "$GOOS" = "windows" ]; then
    BINARY_NAME="${BINARY_NAME}.exe"
fi

echo "Building $BINARY_NAME for $GOOS/$GOARCH..."
go build -ldflags="-s -w" -o "$BINARY_NAME" .

# Create archive
mkdir -p dist
ARCHIVE_NAME="dockitai-cli_${VERSION}_${GOOS}_${GOARCH}"

if [ "$GOOS" = "windows" ]; then
    ARCHIVE_NAME="${ARCHIVE_NAME}.zip"
    zip "dist/$ARCHIVE_NAME" "$BINARY_NAME"
else
    ARCHIVE_NAME="${ARCHIVE_NAME}.tar.gz"
    tar -czf "dist/$ARCHIVE_NAME" "$BINARY_NAME"
fi

echo "Created archive: dist/$ARCHIVE_NAME"

# Create NPM package
./.github/scripts/create-npm-package.sh