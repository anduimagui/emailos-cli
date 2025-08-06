#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const rimraf = require('rimraf');

const binDir = path.join(__dirname, '..', 'bin');

console.log('Cleaning up mailos binary...');

try {
  if (fs.existsSync(binDir)) {
    rimraf.sync(binDir);
    console.log('✓ Cleanup complete');
  }
} catch (error) {
  console.error('Cleanup failed:', error.message);
}