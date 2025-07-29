#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const os = require('os');

const platform = os.platform();
const arch = os.arch();

console.log(`Installing LNB for ${platform} ${arch}...`);

// Map Node.js platform/arch to Go's GOOS/GOARCH
const platformMap = {
  'darwin': 'darwin',
  'linux': 'linux', 
  'win32': 'windows'
};

const archMap = {
  'x64': 'amd64',
  'arm64': 'arm64'
};

const goos = platformMap[platform];
const goarch = archMap[arch];

if (!goos || !goarch) {
  console.error(`Unsupported platform: ${platform} ${arch}`);
  process.exit(1);
}

// Determine binary extension
const binaryExt = platform === 'win32' ? '.exe' : '';
const binaryName = 'lnb' + binaryExt;

// Expected paths from GoReleaser npm publisher
const possiblePaths = [
  // Try direct binary in package root (GoReleaser npm structure - this should be the main path)
  path.join(__dirname, '..', binaryName),
  // Try the bin directory (fallback)
  path.join(__dirname, '..', 'bin', binaryName)
];

// For GoReleaser npm publisher, each platform-specific package contains the binary directly in its root
// The install.js runs in each platform package (like lnb-darwin-amd-64-v-1) 
// and should just ensure the binary is accessible

let sourcePath = null;

// Find the binary
for (const possiblePath of possiblePaths) {
  if (fs.existsSync(possiblePath)) {
    sourcePath = possiblePath;
    console.log(`Found binary at: ${sourcePath}`);
    break;
  }
}

if (!sourcePath) {
  console.error('Could not find LNB binary for your platform');
  console.error('Searched in:');
  possiblePaths.forEach(p => console.error(`  ${p}`));
  process.exit(1);
}

// Ensure bin directory exists
const binDir = path.join(__dirname, '..', 'bin');
if (!fs.existsSync(binDir)) {
  fs.mkdirSync(binDir, { recursive: true });
}

// Copy binary to expected location
const targetPath = path.join(binDir, binaryName);

try {
  // Copy the binary
  fs.copyFileSync(sourcePath, targetPath);
  
  // Make it executable on Unix systems
  if (platform !== 'win32') {
    fs.chmodSync(targetPath, 0o755);
  }
  
  console.log(`✅ LNB installed successfully to ${targetPath}`);
  
  // Verify the binary works
  const { execSync } = require('child_process');
  try {
    const output = execSync(`"${targetPath}" --version`, { encoding: 'utf8' });
    console.log(`✅ Binary verification: ${output.trim()}`);
  } catch (err) {
    console.warn(`⚠️  Binary verification failed: ${err.message}`);
  }
  
} catch (err) {
  console.error(`❌ Failed to install LNB: ${err.message}`);
  process.exit(1);
}