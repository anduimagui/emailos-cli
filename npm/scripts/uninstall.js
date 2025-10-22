#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const os = require('os');
const readline = require('readline');

const binPath = path.join(__dirname, '..', 'bin', 'mailos');
const binPathWin = path.join(__dirname, '..', 'bin', 'mailos.exe');

console.log('🗑️  EmailOS Uninstallation');
console.log('========================');
console.log();
console.log('Removing EmailOS binary...');

try {
  if (fs.existsSync(binPath)) {
    fs.unlinkSync(binPath);
  }
  if (fs.existsSync(binPathWin)) {
    fs.unlinkSync(binPathWin);
  }
  console.log('✓ Binary removed');
} catch (error) {
  console.error('Binary cleanup failed:', error.message);
}

// Check for EmailOS data
const homeDir = os.homedir();
const emailDir = path.join(homeDir, '.email');

if (fs.existsSync(emailDir)) {
  console.log();
  console.log('⚠️  EmailOS Data Detected');
  console.log('========================');
  console.log();
  console.log('EmailOS configuration and email data was found at:');
  console.log(`   ${emailDir}`);
  console.log();
  console.log('This directory contains:');
  console.log('• Your email account configuration (including app passwords)');
  console.log('• All synced email data (sent, received, drafts)');
  console.log('• License information');
  console.log();
  console.log('⚠️  This data will NOT be automatically removed!');
  console.log();
  console.log('To completely remove EmailOS and all data:');
  console.log('   1. Keep the data and remove it manually later:');
  console.log(`      rm -rf "${emailDir}"`);
  console.log();
  console.log('   2. Or use the EmailOS cleanup command (if still available):');
  console.log('      mailos uninstall --backup');
  console.log();
  
  // Try to run cleanup if mailos is still available
  if (process.platform !== 'win32') {
    try {
      const { execSync } = require('child_process');
      console.log('Attempting automatic cleanup...');
      
      // Try to run mailos cleanup
      try {
        execSync('mailos cleanup 2>/dev/null', { stdio: 'ignore' });
        console.log('✓ EmailOS cleanup completed automatically.');
      } catch (cleanupError) {
        // mailos command not found or failed, show manual instructions
        promptManualCleanup();
      }
    } catch (error) {
      promptManualCleanup();
    }
  } else {
    promptManualCleanup();
  }
} else {
  console.log('✓ No EmailOS data found to clean up.');
}

console.log();
console.log('📦 npm uninstall completed.');
console.log();
console.log('Thank you for using EmailOS! 👋');

function promptManualCleanup() {
  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout
  });

  console.log();
  rl.question('Would you like to remove EmailOS data now? (y/N): ', (answer) => {
    const response = answer.trim().toLowerCase();
    
    if (response === 'y' || response === 'yes') {
      try {
        // Create backup first
        const backupDir = path.join(homeDir, 'Downloads', `emailos-backup-${new Date().toISOString().split('T')[0]}`);
        console.log(`Creating backup at: ${backupDir}`);
        
        if (!fs.existsSync(backupDir)) {
          fs.mkdirSync(backupDir, { recursive: true });
        }
        
        // Copy .email directory to backup
        copyDir(emailDir, path.join(backupDir, '.email'));
        console.log('✓ Backup created');
        
        // Remove original directory
        fs.rmSync(emailDir, { recursive: true, force: true });
        console.log('✓ EmailOS data removed');
        console.log(`💾 Backup available at: ${backupDir}`);
      } catch (error) {
        console.error(`❌ Failed to remove EmailOS data: ${error.message}`);
        console.log();
        console.log('You can manually remove the data with:');
        console.log(`   rm -rf "${emailDir}"`);
      }
    } else {
      console.log('EmailOS data preserved.');
      console.log('You can remove it later with:');
      console.log(`   rm -rf "${emailDir}"`);
    }
    
    rl.close();
  });
}

function copyDir(src, dest) {
  if (!fs.existsSync(dest)) {
    fs.mkdirSync(dest, { recursive: true });
  }
  
  const entries = fs.readdirSync(src, { withFileTypes: true });
  
  for (const entry of entries) {
    const srcPath = path.join(src, entry.name);
    const destPath = path.join(dest, entry.name);
    
    if (entry.isDirectory()) {
      copyDir(srcPath, destPath);
    } else {
      fs.copyFileSync(srcPath, destPath);
    }
  }
}