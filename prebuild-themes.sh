#!/bin/bash
set -e

echo "🚀 Starting prebuild of all themes..."

# Configuration
THEMES_REPO="https://github.com/GitDocAI/themes-prototype"
# Use environment variable if set, otherwise use default based on context
# In Docker: /prebuilt-themes
# Local dev: ~/.gitdocai/prebuilt-themes
if [ -n "$PREBUILD_OUTPUT_PATH" ]; then
    PREBUILD_OUTPUT="$PREBUILD_OUTPUT_PATH"
elif [ -d "/.dockerenv" ] || [ -f "/.dockerenv" ]; then
    # Running in Docker
    PREBUILD_OUTPUT="/prebuilt-themes"
else
    # Running locally
    PREBUILD_OUTPUT="$HOME/.gitdocai/prebuilt-themes"
fi
WORKSPACE="/tmp/prebuild-workspace"

echo "📂 Prebuild output directory: $PREBUILD_OUTPUT"

# Check if GITDOCAI_THEMES_PATH is set, otherwise clone repo
if [ -n "$GITDOCAI_THEMES_PATH" ]; then
    echo "📁 Using themes from GITDOCAI_THEMES_PATH: $GITDOCAI_THEMES_PATH"
    THEMES_DIR="$GITDOCAI_THEMES_PATH"

    # Verify the path exists
    if [ ! -d "$THEMES_DIR" ]; then
        echo "❌ Error: GITDOCAI_THEMES_PATH directory does not exist: $THEMES_DIR"
        exit 1
    fi
else
    echo "📦 Cloning themes repository..."
    THEMES_DIR="/tmp/gitdocai-themes"
    rm -rf "$THEMES_DIR"
    git clone "$THEMES_REPO" "$THEMES_DIR"
fi

# Create prebuild output directory
mkdir -p "$PREBUILD_OUTPUT"

# Get list of themes (directories in themes repo, excluding hidden and non-theme dirs)
cd "$THEMES_DIR"
THEMES=$(find . -maxdepth 1 -type d ! -name "." ! -name ".git" ! -name ".*" -exec basename {} \;)

echo "📋 Found themes: $THEMES"

# Build each theme
for THEME in $THEMES; do
    echo ""
    echo "🔨 Building theme: $THEME"

    # Create workspace for this theme
    THEME_WORKSPACE="$WORKSPACE/$THEME"
    rm -rf "$THEME_WORKSPACE"
    mkdir -p "$THEME_WORKSPACE/output"

    # Copy theme directly to output (skip deploy, just copy theme files)
    # Exclude node_modules, dist, build artifacts, and git files
    echo "  📁 Copying theme files..."
    rsync -a \
        --exclude='node_modules' \
        --exclude='dist' \
        --exclude='build' \
        --exclude='.git' \
        --exclude='.DS_Store' \
        --exclude='*.log' \
        --exclude='content' \
        --exclude='public' \
        --exclude='gitdocai.config.json' \
        "$THEMES_DIR/$THEME/" "$THEME_WORKSPACE/output/"

    # Run npm install and build
    echo "  📦 Installing dependencies..."
    cd "$THEME_WORKSPACE/output"
    npm install --force

    echo "  🏗️  Building theme..."
    # Set environment variables for preview mode build
    # This tells the theme to skip embeddings generation
    export VITE_MODE=preview
    export VITE_BACKEND_URL=/api
    export VITE_USE_BACKEND=false
    export VITE_SOURCE=/docs
    npm run build

    # Copy built files to prebuild directory
    echo "  💾 Saving prebuild..."
    mkdir -p "$PREBUILD_OUTPUT/$THEME"

    # Copy the entire output directory (includes dist, public, node_modules, etc)
    cp -r "$THEME_WORKSPACE/output/"* "$PREBUILD_OUTPUT/$THEME/"

    echo "  ✅ Theme $THEME prebuilt successfully"
done

# Clean up
rm -rf "$WORKSPACE"

# Only remove THEMES_DIR if it was cloned (not from GITDOCAI_THEMES_PATH)
if [ -z "$GITDOCAI_THEMES_PATH" ]; then
    rm -rf "$THEMES_DIR"
fi

echo ""
echo "🎉 All themes prebuilt successfully!"
echo "📂 Prebuilt themes location: $PREBUILD_OUTPUT"
ls -la "$PREBUILD_OUTPUT"
