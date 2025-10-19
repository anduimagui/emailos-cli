#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

const binPath = path.join(__dirname, '..', 'bin', 'mailos');
const binPathWin = path.join(__dirname, '..', 'bin', 'mailos.exe');

console.log('Cleaning up mailos binary...');

try {
  if (fs.existsSync(binPath)) {
    fs.unlinkSync(binPath);
  }
  if (fs.existsSync(binPathWin)) {
    fs.unlinkSync(binPathWin);
  }
  console.log('✓ Cleanup complete');
} catch (error) {
  console.error('Cleanup failed:', error.message);
}