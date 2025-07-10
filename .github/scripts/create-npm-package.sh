#!/bin/bash
set -e

PACKAGE_NAME="dockitai-cli-${GOOS}-${GOARCH}"
PACKAGE_DIR="npm-packages/$PACKAGE_NAME"

echo "Creating NPM package: $PACKAGE_NAME"

mkdir -p "$PACKAGE_DIR/bin"

# Copy binary
if [ "$GOOS" = "windows" ]; then
    cp dockitai-cli.exe "$PACKAGE_DIR/bin/"
    BINARY_NAME="dockitai-cli.exe"
else
    cp dockitai-cli "$PACKAGE_DIR/bin/"
    BINARY_NAME="dockitai-cli"
fi

chmod +x "$PACKAGE_DIR/bin/"*

# Create package.json
cat > "$PACKAGE_DIR/package.json" << EOF
{
  "name": "$PACKAGE_NAME",
  "version": "$VERSION",
  "description": "DockitAI CLI tool for $GOOS $GOARCH",
  "bin": {
    "dockitai-cli": "./bin/$BINARY_NAME"
  },
  "files": ["bin/"],
  "os": ["$GOOS"],
  "cpu": ["$GOARCH"],
  "repository": {
    "type": "git",
    "url": "https://github.com/${GITHUB_REPOSITORY}.git"
  },
  "license": "Apache-2.0",
  "publishConfig": {
    "access": "public"
  }
}
EOF

echo "NPM package created at: $PACKAGE_DIR"