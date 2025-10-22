# Installation Guide

MailOS can be installed through multiple methods depending on your platform and preferences.

## Quick Install

### npm (Recommended - All Platforms)
```bash
npm install -g mailos
```

The npm package automatically downloads the correct binary for your platform (macOS, Linux, Windows) and architecture (x64, ARM64).

### Homebrew (macOS/Linux)
```bash
brew tap anduimagui/mailos
brew install mailos
```

## Platform-Specific Installation

### macOS

#### Apple Silicon (M1/M2/M3/M4)
```bash
curl -L https://github.com/anduimagui/emailos-cli/releases/latest/download/mailos-darwin-arm64.tar.gz | tar xz
sudo mv mailos /usr/local/bin/
```

#### Intel Macs
```bash
curl -L https://github.com/anduimagui/emailos-cli/releases/latest/download/mailos-darwin-amd64.tar.gz | tar xz
sudo mv mailos /usr/local/bin/
```

### Linux

#### x64/AMD64
```bash
curl -L https://github.com/anduimagui/emailos-cli/releases/latest/download/mailos-linux-amd64.tar.gz | tar xz
sudo mv mailos /usr/local/bin/
```

#### ARM64
```bash
curl -L https://github.com/anduimagui/emailos-cli/releases/latest/download/mailos-linux-arm64.tar.gz | tar xz
sudo mv mailos /usr/local/bin/
```

### Windows

#### Using PowerShell
```powershell
# Download the binary
Invoke-WebRequest -Uri "https://github.com/anduimagui/emailos-cli/releases/latest/download/mailos-windows-amd64.tar.gz" -OutFile "mailos.tar.gz"

# Extract (requires tar in Windows 10+)
tar -xzf mailos.tar.gz

# Move to a directory in your PATH
Move-Item mailos.exe "C:\Program Files\mailos\mailos.exe"

# Add to PATH if needed
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\Program Files\mailos", [EnvironmentVariableTarget]::User)
```

#### Using WSL
```bash
# Install the Linux version in WSL
curl -L https://github.com/anduimagui/emailos-cli/releases/latest/download/mailos-linux-amd64.tar.gz | tar xz
sudo mv mailos /usr/local/bin/
```

## Build from Source

### Prerequisites
- Go 1.24 or higher
- Git

### Build Steps
```bash
# Clone the repository
git clone https://github.com/anduimagui/emailos-cli.git
cd emailos-cli

# Build the binary
go build -ldflags="-s -w" -o mailos ./cmd/mailos

# Install globally (Unix-like systems)
sudo mv mailos /usr/local/bin/

# Or add to PATH on Windows
# Move mailos.exe to a directory in your PATH
```

## Verify Installation

After installation, verify MailOS is working:

```bash
# Check version
mailos --version

# Run initial setup
mailos setup
```

## Updating MailOS

### Using npm
```bash
npm update -g mailos
```

### Using Direct Download
Re-run the installation commands above to get the latest version.

### Check Current Version
```bash
mailos --version
```

## Complete Uninstallation

⚠️ **Important**: Standard uninstallation methods only remove the EmailOS binary, but leave your configuration and email data intact in the `~/.email` directory.

### Quick Complete Uninstallation (Recommended)

Use EmailOS's built-in uninstall command for complete removal:

```bash
# Complete uninstallation with backup
mailos uninstall --backup

# Force uninstallation without prompts
mailos uninstall --force

# See what would be removed (dry run)
mailos uninstall --dry-run

# Keep emails but remove configuration
mailos uninstall --keep-emails

# Keep configuration but remove emails
mailos uninstall --keep-config
```

### Package Manager Uninstallation

#### npm
```bash
npm uninstall -g mailos
```
📝 **Note**: The npm uninstall script will detect your EmailOS data and offer to remove it interactively.

#### Homebrew
```bash
brew uninstall mailos
```
⚠️ **Important**: After Homebrew uninstallation, run `mailos cleanup` to remove remaining data.

#### Manual Binary Removal
```bash
# Unix-like systems
sudo rm /usr/local/bin/mailos

# Windows
# Remove mailos.exe from wherever you installed it
```
⚠️ **Important**: After manual removal, your data remains in `~/.email`. See cleanup instructions below.

### Cleaning Up After Package Manager Uninstallation

If you uninstalled EmailOS using a package manager but want to remove all data:

#### Automatic Cleanup (Recommended)
```bash
# If mailos command is still available
mailos cleanup

# If you get "command not found", manually download and run:
curl -L https://github.com/anduimagui/emailos-cli/releases/latest/download/mailos-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m).tar.gz | tar xz
./mailos cleanup
rm ./mailos
```

#### Manual Cleanup
```bash
# ⚠️ This will permanently delete all EmailOS data!
# Create backup first (optional):
cp -r ~/.email ~/Downloads/emailos-backup-$(date +%Y%m%d)

# Remove EmailOS data directory
rm -rf ~/.email

# Remove any local project configurations
find ~/projects -name ".email" -type d -exec rm -rf {} + 2>/dev/null
```

### What Gets Removed

When performing complete uninstallation, the following are removed:

📁 **Configuration Files**:
- `~/.email/config.json` - Account settings and credentials
- `~/.email/.license` - License information

📧 **Email Data**:
- `~/.email/sent/` - All sent email copies
- `~/.email/received/` - All synced received emails  
- `~/.email/drafts/` - All draft emails

🗂️ **Local Project Configs**:
- `.email/` directories in project folders (if any)

🔧 **System Integration**:
- `EMAILOS.md` files in project directories

### Data Recovery

If you uninstalled EmailOS but want to recover your data:

```bash
# Check if data still exists
ls -la ~/.email

# If data exists, reinstall EmailOS to access it
npm install -g mailos

# Or manually backup the data
cp -r ~/.email ~/emailos-backup-$(date +%Y%m%d)
```

### Verification

To verify complete removal:

```bash
# Check for remaining files
ls -la ~/.email 2>/dev/null || echo "✓ No EmailOS data found"

# Check for remaining binary
which mailos 2>/dev/null || echo "✓ No EmailOS binary found"

# Check for local configs (optional)
find ~ -name ".email" -type d 2>/dev/null
```

### Troubleshooting Uninstallation

#### "Permission Denied" Errors
```bash
# If you get permission errors removing ~/.email
sudo rm -rf ~/.email

# If you get permission errors removing the binary
sudo rm $(which mailos)
```

#### Partial Uninstallation Recovery
If uninstallation was interrupted:

```bash
# Reinstall EmailOS temporarily
npm install -g mailos

# Complete the uninstallation
mailos uninstall --force

# Remove the binary again
npm uninstall -g mailos
```

#### Orphaned Data Detection
EmailOS automatically detects orphaned data and will show hints when you run other commands. To manually check:

```bash
# If EmailOS is installed
mailos cleanup

# If EmailOS is not installed but you suspect data remains
ls -la ~/.email
```

## System Requirements

### Minimum Requirements
- **Operating System**: macOS 10.15+, Linux (glibc 2.17+), Windows 10+
- **Memory**: 512MB RAM
- **Storage**: 50MB free space
- **Network**: Internet connection for email operations

### Supported Email Providers
- Gmail
- Fastmail
- Zoho Mail
- Outlook/Hotmail
- Yahoo Mail
- Any IMAP/SMTP compatible provider

## Troubleshooting

### Permission Denied (Unix-like systems)
If you get a permission error when moving to `/usr/local/bin/`:
```bash
# Create the directory if it doesn't exist
sudo mkdir -p /usr/local/bin

# Move with sudo
sudo mv mailos /usr/local/bin/

# Make executable
sudo chmod +x /usr/local/bin/mailos
```

### Command Not Found
If `mailos` is not found after installation:

1. **Check if it's in PATH**:
   ```bash
   echo $PATH
   which mailos
   ```

2. **Add to PATH** (bash/zsh):
   ```bash
   echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
   source ~/.bashrc
   ```

3. **For Windows**, ensure the installation directory is in your PATH environment variable.

### npm Installation Issues

The npm package automatically installs the correct binary for your platform during the postinstall phase.

If npm installation fails:

1. **Update npm and Node.js**:
   ```bash
   npm install -g npm@latest
   # Ensure Node.js 14.0.0 or higher is installed
   node --version
   ```

2. **Clear npm cache**:
   ```bash
   npm cache clean --force
   ```

3. **Use sudo (Unix-like systems)**:
   ```bash
   sudo npm install -g mailos
   ```

4. **Check npm prefix**:
   ```bash
   npm config get prefix
   ```

5. **Platform-specific binary not found**:
   If the postinstall script fails to find a binary for your platform:
   ```bash
   # Check your platform details
   node -e "console.log(process.platform, process.arch)"
   
   # Supported platforms: darwin-x64, darwin-arm64, linux-x64, linux-arm64, win32-x64
   ```

6. **Manual binary installation**:
   If npm installation continues to fail, use the direct download method above.

### License Key Issues

If you encounter license validation errors:

1. Ensure you have an active internet connection
2. Check if your license key is valid
3. Visit https://email-os.com/checkout to purchase or renew
4. Run `mailos setup` to re-enter your license key

## Getting Help

- **Documentation**: https://github.com/anduimagui/emailos-cli/tree/main/docs
- **Issues**: https://github.com/anduimagui/emailos-cli/issues
- **Website**: https://email-os.com

## Next Steps

After installation, proceed to:
1. [Initial Setup](setup.md) - Configure your email account
2. [Usage Guide](usage.md) - Learn how to use MailOS
3. [AI Integration](ai-integration.md) - Set up AI features