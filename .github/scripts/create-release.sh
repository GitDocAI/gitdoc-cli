#!/bin/bash
set -e

# Download all artifacts
gh run download --name release-* --dir release-artifacts

# Create GitHub release
echo "Creating GitHub release $VERSION..."

# Find all archives
RELEASE_FILES=$(find release-artifacts -name "*.tar.gz" -o -name "*.zip" | tr '\n' ' ')

gh release create "$VERSION" \
    --title "Release $VERSION" \
    --notes "Release $VERSION of DockitAI CLI" \
    $RELEASE_FILES

echo "Release created successfully!"