#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');
const fetch = require('node-fetch');
const tar = require('tar');
const rimraf = require('rimraf');

const BINARY_NAME = 'mailos';
const GITHUB_REPO = 'emailos/mailos';
const VERSION = require('../package.json').version;
const USE_LOCAL_BUILD = process.env.MAILOS_BUILD_LOCAL === 'true';

// Platform mapping
const PLATFORM_MAPPING = {
  'darwin-x64': 'darwin-amd64',
  'darwin-arm64': 'darwin-arm64',
  'linux-x64': 'linux-amd64',
  'linux-arm64': 'linux-arm64',
  'win32-x64': 'windows-amd64',
  'win32-arm64': 'windows-arm64'
};

async function getDownloadUrl() {
  const platform = process.platform;
  const arch = process.arch;
  const platformKey = `${platform}-${arch}`;
  
  const goPlatform = PLATFORM_MAPPING[platformKey];
  if (!goPlatform) {
    throw new Error(`Unsupported platform: ${platformKey}`);
  }
  
  // Check if we should build locally
  if (USE_LOCAL_BUILD) {
    return { goPlatform, needsBuild: true };
  }
  
  // Try to get release URL
  try {
    const releaseUrl = `https://api.github.com/repos/${GITHUB_REPO}/releases/tags/v${VERSION}`;
    const response = await fetch(releaseUrl);
    
    if (response.ok) {
      const release = await response.json();
      const assetName = `mailos-${goPlatform}.tar.gz`;
      const asset = release.assets.find(a => a.name === assetName);
      
      if (asset) {
        return { 
          goPlatform, 
          needsBuild: false, 
          downloadUrl: asset.browser_download_url 
        };
      }
    }
  } catch (error) {
    console.warn('Could not fetch release info, will build locally');
  }
  
  // Fall back to local build
  return { goPlatform, needsBuild: true };
}

async function downloadBinary(url, dest) {
  console.log(`Downloading ${BINARY_NAME} from ${url}...`);
  
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`Failed to download: ${response.statusText}`);
  }
  
  const buffer = await response.buffer();
  fs.writeFileSync(dest, buffer);
  
  if (process.platform !== 'win32') {
    fs.chmodSync(dest, '755');
  }
  
  console.log(`Downloaded ${BINARY_NAME} successfully!`);
}

async function buildBinary() {
  console.log(`Building ${BINARY_NAME} from source...`);
  
  const binDir = path.join(__dirname, '..', 'bin');
  const binPath = path.join(binDir, process.platform === 'win32' ? `${BINARY_NAME}.exe` : BINARY_NAME);
  
  // Ensure bin directory exists
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }
  
  // Build the Go binary
  const goPath = path.join(__dirname, '..', '..', 'cmd', 'mailos', 'main.go');
  const goBuildCmd = `go build -o "${binPath}" "${goPath}"`;
  
  try {
    console.log('Running go build...');
    execSync(goBuildCmd, { 
      stdio: 'inherit',
      cwd: path.join(__dirname, '..', '..')
    });
    console.log(`Built ${BINARY_NAME} successfully!`);
  } catch (error) {
    console.error('Failed to build from source. Make sure Go is installed.');
    throw error;
  }
}

async function downloadAndExtract(url, destDir) {
  console.log(`Downloading from ${url}...`);
  
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`Failed to download: ${response.statusText}`);
  }
  
  const buffer = await response.buffer();
  
  // Extract tar.gz
  await tar.x({
    file: buffer,
    cwd: destDir,
    strip: 0
  });
  
  console.log('Extracted successfully!');
}

async function install() {
  try {
    const { goPlatform, needsBuild, downloadUrl } = await getDownloadUrl();
    
    const binDir = path.join(__dirname, '..', 'bin');
    if (!fs.existsSync(binDir)) {
      fs.mkdirSync(binDir, { recursive: true });
    }
    
    if (needsBuild) {
      // Check if Go is installed
      try {
        execSync('go version', { stdio: 'ignore' });
      } catch (error) {
        console.error('Go is not installed. Please install Go to build mailos.');
        console.error('Visit: https://golang.org/dl/');
        process.exit(1);
      }
      
      await buildBinary();
    } else {
      // Download pre-built binary
      await downloadAndExtract(downloadUrl, binDir);
      
      // Make binary executable
      const binPath = path.join(binDir, process.platform === 'win32' ? `${BINARY_NAME}.exe` : BINARY_NAME);
      if (process.platform !== 'win32') {
        fs.chmodSync(binPath, '755');
      }
    }
    
    console.log('\n✓ mailos installed successfully!');
    console.log('Run "mailos help" to get started.');
    
  } catch (error) {
    console.error('Installation failed:', error.message);
    process.exit(1);
  }
}

// Run installation
install();