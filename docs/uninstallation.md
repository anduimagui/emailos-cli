# EmailOS Complete Uninstallation Guide

This guide provides comprehensive instructions for completely removing EmailOS from your system, including all configuration files, email data, and system traces.

## TL;DR - Quick Complete Removal

```bash
# One-command complete uninstallation with backup
mailos uninstall --backup

# Or force removal without confirmation
mailos uninstall --force
```

## Understanding EmailOS Data

EmailOS stores data in several locations:

### Primary Data Location
- **`~/.email/`** - Main directory containing:
  - `config.json` - Account configuration and credentials
  - `sent/` - Copies of sent emails
  - `received/` - Synced received emails
  - `drafts/` - Draft emails
  - `.license` - License information

### Secondary Locations
- **`.email/`** - Local project configurations (optional)
- **`EMAILOS.md`** - AI integration files in project directories

## Complete Uninstallation Methods

### Method 1: Built-in Uninstall Command (Recommended)

EmailOS includes a comprehensive uninstall command that handles all cleanup automatically:

#### Basic Uninstallation
```bash
# Interactive uninstallation with confirmation
mailos uninstall

# Silent uninstallation without prompts
mailos uninstall --force

# See what would be removed without doing it
mailos uninstall --dry-run
```

#### Advanced Options
```bash
# Create backup before removal
mailos uninstall --backup

# Specify custom backup location
mailos uninstall --backup --backup-path ~/Desktop/emailos-backup

# Keep email data, only remove configuration
mailos uninstall --keep-emails

# Keep configuration, only remove email data
mailos uninstall --keep-config

# Minimal output
mailos uninstall --quiet
```

### Method 2: Package Manager + Cleanup

If you prefer using your package manager:

#### npm
```bash
# Uninstall binary (will prompt for data removal)
npm uninstall -g mailos

# If prompted, choose to remove data or clean up manually later
```

#### Homebrew
```bash
# Remove binary
brew uninstall mailos

# Clean up remaining data
mailos cleanup  # if command still works
# OR manually: rm -rf ~/.email
```

#### Manual Binary Removal
```bash
# Find and remove binary
sudo rm $(which mailos)

# Clean up data
rm -rf ~/.email
```

## Handling Orphaned Data

Sometimes EmailOS data remains after external uninstallation. EmailOS automatically detects this situation:

### Automatic Detection
- EmailOS shows hints when orphaned data is detected
- Runs periodic checks for orphaned configurations
- Provides guidance for cleanup

### Manual Cleanup
```bash
# Check for orphaned data
ls -la ~/.email

# If EmailOS binary is still available
mailos cleanup

# If binary is gone, download cleanup tool
curl -L https://github.com/anduimagui/emailos-cli/releases/latest/download/mailos-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m).tar.gz | tar xz
./mailos cleanup
rm ./mailos

# Or manual removal
rm -rf ~/.email
find ~ -name ".email" -type d -exec rm -rf {} + 2>/dev/null
```

## Backup and Recovery

### Creating Backups
```bash
# Manual backup before removal
cp -r ~/.email ~/emailos-backup-$(date +%Y%m%d)

# Automatic backup during uninstall
mailos uninstall --backup

# Custom backup location
mailos uninstall --backup --backup-path /path/to/backup
```

### Recovering Data
```bash
# Check if data still exists
ls -la ~/.email

# Restore from backup
cp -r ~/emailos-backup-YYYYMMDD ~/.email

# Reinstall EmailOS to access recovered data
npm install -g mailos
```

## Verification and Troubleshooting

### Verify Complete Removal
```bash
# Check for EmailOS binary
which mailos 2>/dev/null && echo "Binary still exists" || echo "✓ Binary removed"

# Check for configuration data
ls -la ~/.email 2>/dev/null && echo "Data still exists" || echo "✓ Data removed"

# Check for local project configs
find ~ -name ".email" -type d 2>/dev/null | head -10

# Check for AI integration files
find ~ -name "EMAILOS.md" 2>/dev/null | head -10
```

### Common Issues

#### Permission Denied
```bash
# If you can't remove ~/.email
sudo rm -rf ~/.email

# If you can't remove the binary
sudo rm $(which mailos)
```

#### Partial Uninstallation
```bash
# If uninstallation was interrupted, reinstall temporarily
npm install -g mailos

# Complete the uninstallation
mailos uninstall --force

# Remove binary again
npm uninstall -g mailos
```

#### Hidden Files
```bash
# Check for hidden EmailOS files
find ~ -name ".*email*" -o -name ".*mailos*" 2>/dev/null

# Remove if found
rm -rf ~/.email ~/.mailos ~/.emailos
```

#### Process Still Running
```bash
# Check for running EmailOS processes
ps aux | grep mailos

# Kill if necessary
pkill -f mailos
```

## Platform-Specific Instructions

### macOS
```bash
# Standard uninstallation
mailos uninstall --backup

# Manual cleanup
rm -rf ~/.email
rm -f /usr/local/bin/mailos
rm -f /opt/homebrew/bin/mailos

# Check LaunchAgents (if any)
ls ~/Library/LaunchAgents/*mailos* 2>/dev/null
```

### Linux
```bash
# Standard uninstallation
mailos uninstall --backup

# Manual cleanup
rm -rf ~/.email
rm -f /usr/local/bin/mailos
rm -f ~/.local/bin/mailos

# Check systemd services (if any)
systemctl --user list-units | grep mailos
```

### Windows
```powershell
# Standard uninstallation
mailos uninstall --backup

# Manual cleanup
Remove-Item -Recurse -Force "$env:USERPROFILE\.email"
Remove-Item "C:\Program Files\mailos\mailos.exe"

# Check scheduled tasks (if any)
Get-ScheduledTask | Where-Object {$_.TaskName -like "*mailos*"}
```

## Security Considerations

### Sensitive Data Removal
EmailOS configuration contains sensitive information:
- Email app passwords
- License keys
- Account credentials

Ensure complete removal:
```bash
# Secure deletion (Linux/macOS)
rm -P ~/.email/config.json  # macOS
shred -u ~/.email/config.json  # Linux

# Or use the built-in secure cleanup
mailos uninstall --force
```

### Backup Security
If creating backups:
- Store in secure location
- Encrypt if necessary
- Delete when no longer needed

```bash
# Encrypted backup
tar czf - ~/.email | gpg -c > emailos-backup-$(date +%Y%m%d).tar.gz.gpg

# Secure backup deletion
rm -P emailos-backup-*.tar.gz.gpg  # macOS
shred -u emailos-backup-*.tar.gz.gpg  # Linux
```

## Reinstallation

If you need to reinstall EmailOS later:

```bash
# Reinstall via npm
npm install -g mailos

# Restore from backup if needed
cp -r ~/emailos-backup-YYYYMMDD ~/.email

# Run setup if starting fresh
mailos setup
```

## Support

If you encounter issues during uninstallation:

1. **Try the automated cleanup**: `mailos cleanup`
2. **Check the troubleshooting section** above
3. **Report issues**: https://github.com/anduimagui/emailos-cli/issues
4. **Manual cleanup**: Remove files manually as documented

## Complete Uninstallation Checklist

- [ ] Run `mailos uninstall --backup`
- [ ] Verify binary removal: `which mailos`
- [ ] Verify data removal: `ls ~/.email`
- [ ] Check for local configs: `find ~ -name ".email" -type d`
- [ ] Check for EMAILOS.md files: `find ~ -name "EMAILOS.md"`
- [ ] Secure backup if needed
- [ ] Update package manager if required

✅ **Complete!** EmailOS has been fully removed from your system.