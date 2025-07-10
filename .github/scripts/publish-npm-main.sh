#!/bin/bash
set -e

echo "Creating main NPM package..."

mkdir -p npm-main-package

# Create package.json
cat > npm-main-package/package.json << EOF
{
  "name": "dockitai-cli",
  "version": "$VERSION",
  "description": "DockitAI CLI tool",
  "bin": {
    "dockitai-cli": "./install.js"
  },
  "files": ["install.js", "README.md"],
  "scripts": {
    "postinstall": "node install.js"
  },
  "repository": {
    "type": "git",
    "url": "https://github.com/${GITHUB_REPOSITORY}.git"
  },
  "license": "Apache-2.0",
  "optionalDependencies": {
    "dockitai-cli-linux-amd64": "$VERSION",
    "dockitai-cli-linux-arm64": "$VERSION",
    "dockitai-cli-darwin-amd64": "$VERSION",
    "dockitai-cli-darwin-arm64": "$VERSION",
    "dockitai-cli-windows-amd64": "$VERSION",
    "dockitai-cli-windows-arm64": "$VERSION"
  }
}
EOF

# Create install.js
cat > npm-main-package/install.js << 'EOF'
#!/usr/bin/env node

const os = require('os');
const fs = require('fs');
const path = require('path');

function getPackageName() {
  const platform = os.platform();
  const arch = os.arch();
  
  const platformMap = {
    'darwin': 'darwin',
    'linux': 'linux',
    'win32': 'windows'
  };
  
  const archMap = {
    'x64': 'amd64',
    'arm64': 'arm64'
  };
  
  const mappedPlatform = platformMap[platform];
  const mappedArch = archMap[arch];
  
  if (!mappedPlatform || !mappedArch) {
    console.error(`Unsupported platform: ${platform} ${arch}`);
    process.exit(1);
  }
  
  return `dockitai-cli-${mappedPlatform}-${mappedArch}`;
}

const packageName = getPackageName();

try {
  const packagePath = require.resolve(packageName);
  const binPath = path.join(path.dirname(packagePath), 'bin');
  const binaryName = os.platform() === 'win32' ? 'dockitai-cli.exe' : 'dockitai-cli';
  const sourcePath = path.join(binPath, binaryName);
  const targetPath = path.join(__dirname, binaryName);
  
  fs.copyFileSync(sourcePath, targetPath);
  fs.chmodSync(targetPath, 0o755);
  
  console.log(`DockitAI CLI installed successfully!`);
} catch (error) {
  console.error(`Failed to install platform-specific package: ${packageName}`);
  console.error('Please install it manually:', `npm install -g ${packageName}`);
  process.exit(1);
}
EOF

# Copy README if exists
if [ -f "README.md" ]; then
  cp README.md npm-main-package/
fi

chmod +x npm-main-package/install.js

# Publish to NPM
cd npm-main-package
npm publish --access public

echo "Published main package successfully!"