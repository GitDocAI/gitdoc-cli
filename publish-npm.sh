#!/bin/bash
set -e

# Arguments passed by GoReleaser
OS=$1
ARCH=$2
VERSION=$3
ARTIFACT_PATH=$4

PACKAGE_DIR="npm-package-${OS}-${ARCH}"
rm -rf "$PACKAGE_DIR"
mkdir -p "$PACKAGE_DIR/bin"

if [ "$OS" = "windows" ]; then
  unzip -j "$ARTIFACT_PATH" -d "$PACKAGE_DIR/bin/"
  BINARY_NAME="dockitai-cli.exe"
else
  tar -xzf "$ARTIFACT_PATH" -C "$PACKAGE_DIR/bin/" --strip-components=0
  BINARY_NAME="dockitai-cli"
fi

chmod +x "$PACKAGE_DIR/bin/"*

cat > "$PACKAGE_DIR/package.json" << EOF
{
  "name": "dockitai-cli-${OS}-${ARCH}",
  "version": "${VERSION}",
  "description": "DockitAI CLI tool for ${OS} ${ARCH}",
  "bin": {
    "dockitai-cli": "./bin/$BINARY_NAME"
  },
  "files": ["bin/"],
  "os": ["${OS}"],
  "cpu": ["${ARCH}"],
  "repository": {
    "type": "git",
    "url": "https://github.com/${GITHUB_REPOSITORY_OWNER}/${GITHUB_REPOSITORY}.git"
  },
  "license": "Apache-2.0"
}
EOF

cd "$PACKAGE_DIR"
echo "//registry.npmjs.org/:_authToken=${NPM_TOKEN}" > .npmrc
npm publish --access public