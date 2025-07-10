#!/bin/bash
set -e

PACKAGE_NAME="dockitai-cli-${GOOS}-${GOARCH}"

# Download the NPM package artifact
gh run download --name "npm-${GOOS}-${GOARCH}" --dir npm-packages

# Publish to NPM
echo "Publishing $PACKAGE_NAME to NPM..."
cd "npm-packages/$PACKAGE_NAME"
npm publish --access public

echo "Published $PACKAGE_NAME successfully!"