#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const https = require('https');
const { execSync } = require('child_process');

const packageJson = require('../package.json');
const version = packageJson.version;

// Detect platform and architecture
const platform = process.platform;
const arch = process.arch;

// Map Node.js platform/arch to GoReleaser naming
const platformMap = {
  'darwin': 'Darwin',
  'linux': 'Linux', 
  'win32': 'Windows'
};

const archMap = {
  'x64': 'x86_64',
  'arm64': 'arm64'
};

const mappedPlatform = platformMap[platform];
const mappedArch = archMap[arch];

if (!mappedPlatform || !mappedArch) {
  console.error(`Unsupported platform: ${platform}-${arch}`);
  process.exit(1);
}

// Skip Windows arm64 (not supported)
if (platform === 'win32' && arch === 'arm64') {
  console.error('Windows ARM64 is not supported');
  process.exit(1);
}

const binaryName = platform === 'win32' ? 'lnb.exe' : 'lnb';
const archiveName = `lnb_${version}_${mappedPlatform}_${mappedArch}.tar.gz`;
const downloadUrl = `https://github.com/muthuishere/lnb/releases/download/v${version}/${archiveName}`;

const binDir = path.join(__dirname, '..', 'bin');
const binaryPath = path.join(binDir, binaryName);

async function downloadBinary() {
  console.log(`Downloading LNB v${version} for ${platform}-${arch}...`);
  console.log(`URL: ${downloadUrl}`);

  // Create bin directory
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }

  try {
    // Download and extract the binary
    const tempDir = fs.mkdtempSync(path.join(require('os').tmpdir(), 'lnb-'));
    const archivePath = path.join(tempDir, archiveName);

    await downloadFile(downloadUrl, archivePath);
    
    // Extract the archive
    if (platform === 'win32') {
      // Use PowerShell on Windows
      execSync(`powershell -command "Expand-Archive -Path '${archivePath}' -DestinationPath '${tempDir}'"`, { stdio: 'inherit' });
    } else {
      // Use tar on Unix-like systems
      execSync(`tar -xzf "${archivePath}" -C "${tempDir}"`, { stdio: 'inherit' });
    }

    // Move the binary to bin directory
    const extractedBinary = path.join(tempDir, binaryName);
    fs.copyFileSync(extractedBinary, binaryPath);
    
    // Make executable on Unix-like systems
    if (platform !== 'win32') {
      fs.chmodSync(binaryPath, '755');
    }

    // Clean up
    fs.rmSync(tempDir, { recursive: true, force: true });

    console.log(`✅ LNB v${version} installed successfully!`);
    console.log(`Binary location: ${binaryPath}`);
    console.log(`Try running: lnb help`);

  } catch (error) {
    console.error('❌ Failed to install LNB:', error.message);
    process.exit(1);
  }
}

function downloadFile(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    
    https.get(url, (response) => {
      if (response.statusCode === 302 || response.statusCode === 301) {
        // Handle redirect
        return downloadFile(response.headers.location, dest).then(resolve).catch(reject);
      }
      
      if (response.statusCode !== 200) {
        reject(new Error(`Download failed with status ${response.statusCode}`));
        return;
      }

      response.pipe(file);
      
      file.on('finish', () => {
        file.close(resolve);
      });
      
      file.on('error', (err) => {
        fs.unlink(dest, () => {}); // Delete the file on error
        reject(err);
      });
    }).on('error', reject);
  });
}

downloadBinary();
