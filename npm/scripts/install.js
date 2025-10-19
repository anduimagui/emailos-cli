#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

const BINARY_NAME = 'mailos';

// Platform mapping
const PLATFORM_MAPPING = {
  'darwin-x64': 'darwin-amd64',
  'darwin-arm64': 'darwin-arm64',
  'linux-x64': 'linux-amd64',
  'linux-arm64': 'linux-arm64',
  'win32-x64': 'windows-amd64',
  'win32-arm64': 'windows-amd64' // Fallback to x64 for Windows ARM
};

function install() {
  try {
    const platform = process.platform;
    const arch = process.arch;
    const platformKey = `${platform}-${arch}`;
    
    const goPlatform = PLATFORM_MAPPING[platformKey];
    if (!goPlatform) {
      throw new Error(`Unsupported platform: ${platformKey}`);
    }
    
    console.log(`Installing ${BINARY_NAME} for ${platformKey} (${goPlatform})`);
    
    // Determine source and target binary paths
    const isWindows = platform === 'win32';
    const sourceExt = isWindows ? '.exe' : '';
    const sourceBinary = path.join(__dirname, '..', 'bin', `${BINARY_NAME}-${goPlatform}${sourceExt}`);
    const targetBinary = path.join(__dirname, '..', 'bin', `${BINARY_NAME}${sourceExt}`);
    
    // Check if source binary exists
    if (!fs.existsSync(sourceBinary)) {
      throw new Error(`Binary not found for platform ${goPlatform}: ${sourceBinary}`);
    }
    
    // Copy platform-specific binary to generic name
    console.log(`Copying ${sourceBinary} to ${targetBinary}`);
    fs.copyFileSync(sourceBinary, targetBinary);
    
    // Make executable on Unix systems
    if (!isWindows) {
      fs.chmodSync(targetBinary, '755');
    }
    
    console.log(`✓ ${BINARY_NAME} installed successfully!`);
    console.log('Run "mailos help" to get started.');
    
  } catch (error) {
    console.error('Installation failed:', error.message);
    console.error('\nTroubleshooting:');
    console.error('- Check if your platform is supported');
    console.error('- Try reinstalling the package');
    console.error('- Report issues at: https://github.com/anduimagui/emailos-cli/issues');
    process.exit(1);
  }
}

// Run installation
install();