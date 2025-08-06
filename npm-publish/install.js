const fs = require('fs');
const path = require('path');

const BINARY_NAME = 'gitdocai-cli';

function getPlatform() {
  const type = process.platform;
  const arch = process.arch;
  
  if (type === 'darwin') {
    return arch === 'x64' ? 'darwin-x64' : 'darwin-arm64';
  }
  
  if (type === 'win32') {
    return arch === 'x64' ? 'win32-x64' : 'win32-arm64';
  }
  
  if (type === 'linux') {
    return arch === 'x64' ? 'linux-x64' : 'linux-arm64';
  }
  
  throw new Error(`Unsupported platform: ${type} ${arch}`);
}

const platform = getPlatform();
const ext = process.platform === 'win32' ? '.exe' : '';
const sourcePath = path.join(__dirname, 'dist', `${BINARY_NAME}-${platform}${ext}`);
const targetPath = path.join(__dirname, `${BINARY_NAME}${ext}`);

if (fs.existsSync(sourcePath)) {
  fs.copyFileSync(sourcePath, targetPath);
  if (process.platform !== 'win32') {
    fs.chmodSync(targetPath, 0o755);
  }
  console.log(`✅ ${BINARY_NAME} binary prepared for ${platform}`);
} else {
  console.error(`❌ Binary not found for ${platform}`);
  process.exit(1);
}