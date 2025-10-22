# EmailOS CLI Uninstallation Analysis

## Summary

**Does the current CLI command installation and configuration completely delete all required files if the user decides to uninstall?**

**Answer: NO** - The current uninstallation process does **NOT** completely remove all traces of the application from the user's system, specifically leaving the `~/.email` folder and all its contents intact.

## Installation Process Analysis

### Files Created During Installation

The EmailOS CLI creates several files and directories during installation and setup:

#### 1. Binary Installation Files
- **Location**: Various locations depending on installation method
- **npm**: Binary copied to npm's global packages directory
- **Homebrew**: Binary installed to Homebrew's bin directory  
- **Manual**: Binary copied to `/usr/local/bin/mailos` or custom location

#### 2. Configuration and Data Files (Primary Concern)
- **Location**: `~/.email/` directory
- **Files Created**:
  - `~/.email/config.json` - Main configuration file containing:
    - Email provider settings
    - Email credentials (app passwords)
    - License key
    - AI CLI preferences
    - Account configurations
  - `~/.email/sent/` - Directory for sent emails
  - `~/.email/received/` - Directory for received emails  
  - `~/.email/drafts/` - Directory for draft emails
  - `~/.email/README.md` - Documentation file

#### 3. Local Project Files (Optional)
- **Location**: `.email/` directory in project folders
- **Files**: Local configuration overrides when using project-specific settings

#### 4. AI Integration Files
- **Location**: Current working directory
- **Files**: `EMAILOS.md` - AI integration instructions file

## Uninstallation Process Analysis

### What Gets Removed

#### 1. npm Uninstallation
**File**: [npm/scripts/uninstall.js](npm/scripts/uninstall.js:1)
```javascript
const binPath = path.join(__dirname, '..', 'bin', 'mailos');
const binPathWin = path.join(__dirname, '..', 'bin', 'mailos.exe');

// Only removes the binary files from npm package
if (fs.existsSync(binPath)) {
    fs.unlinkSync(binPath);
}
if (fs.existsSync(binPathWin)) {
    fs.unlinkSync(binPathWin);
}
```

#### 2. Homebrew Uninstallation
**Command**: `brew uninstall mailos`
- Only removes the binary from Homebrew's installation directory

#### 3. Manual Uninstallation
**Commands from documentation**: [docs/installation.md:136](docs/installation.md:136)
```bash
# Unix-like systems
sudo rm /usr/local/bin/mailos

# Windows
# Remove mailos.exe from wherever you installed it
```

### What Does NOT Get Removed

#### Critical Gap: ~/.email Directory
**None of the uninstallation methods remove the `~/.email` directory**, which contains:

1. **Sensitive Configuration Data**:
   - Email credentials and app passwords ([setup.go:21](setup.go:21))
   - License keys
   - Provider authentication settings

2. **Email Storage**:
   - All synced emails in `received/`, `sent/`, and `drafts/` directories
   - Potentially large amounts of user data

3. **Configuration Files**:
   - Account settings and preferences
   - AI CLI configurations

#### Evidence from Code Analysis

**File**: [config.go:886](config.go:886) - `EnsureEmailDirectories()` function
```go
func EnsureEmailDirectories() error {
    baseDir, err := GetEmailStorageDir()
    // Creates ~/.email directory and subdirectories
    // These directories are NEVER cleaned up during uninstall
}
```

**File**: [setup.go:144](setup.go:144) - Security notice acknowledges local storage
```go
fmt.Println("• Stores your email configuration ONLY on your local machine (~/.email/)")
```

**File**: [constants.go:15](constants.go:15) - Configuration constants
```go
const (
    ConfigDir      = ".email"        // This directory persists after uninstall
    ConfigFileName = "config.json"   // Contains sensitive data
)
```

#### Limited Cleanup Options

**File**: [frontend.go:1055](frontend.go:1055) - Local config removal (not global)
```go
if err := os.RemoveAll(".email"); err != nil {
    return fmt.Errorf("failed to remove local configuration: %v", err)
}
```
This only removes local project `.email` directories, not the global `~/.email` directory.

## Security and Privacy Implications

### 1. Persistent Sensitive Data
- **App passwords remain on disk** in plaintext in `~/.email/config.json`
- **License keys persist** after uninstallation
- **Email content remains accessible** in local storage directories

### 2. Incomplete Uninstallation
Users who uninstall EmailOS believing all traces are removed will have:
- Authentication credentials still stored on their system
- Potentially large amounts of email data consuming disk space
- Configuration files that could be accessed by other applications

### 3. Documentation Gap
The [installation documentation](docs/installation.md:123) mentions uninstallation but doesn't warn users about persistent data:
```markdown
## Uninstallation

### npm
```bash
npm uninstall -g mailos
```
```
No mention of manually removing `~/.email` directory.

## Recommendations

### 1. Immediate Actions Needed

1. **Update Uninstallation Scripts**: Modify npm uninstall script to remove `~/.email` directory
2. **Update Documentation**: Add clear warnings about persistent data and manual cleanup steps
3. **Add Cleanup Command**: Implement `mailos cleanup` or `mailos uninstall` command

### 2. Proposed Uninstall Process

#### Enhanced npm Uninstallation Script
```javascript
// In npm/scripts/uninstall.js - add after binary cleanup:
const os = require('os');
const homeDir = os.homedir();
const emailDir = path.join(homeDir, '.email');

console.log('EmailOS uninstallation includes removing configuration and email data.');
console.log('This will delete: ' + emailDir);
console.log('This action cannot be undone.');

// Prompt user for confirmation before removing ~/.email
```

#### Documentation Updates
Add to installation.md:
```markdown
## Complete Uninstallation

⚠️ **Important**: Standard uninstallation only removes the binary.
Your email data and configuration remain in `~/.email/`

To completely remove EmailOS:
1. Uninstall the binary: `npm uninstall -g mailos`
2. Remove data directory: `rm -rf ~/.email`
```

### 3. Implementation Priority
- **HIGH**: Update documentation to warn users
- **HIGH**: Enhance npm uninstall script with user prompt
- **MEDIUM**: Add dedicated cleanup command to CLI
- **LOW**: Add automatic cleanup option during uninstall

## Conclusion

The current EmailOS CLI uninstallation process is **incomplete and potentially problematic** from both security and user experience perspectives. Users expect uninstallation to remove all traces of an application, but currently the `~/.email` directory with sensitive credentials and email data persists indefinitely.

This requires immediate attention to:
1. Protect user privacy and security
2. Meet user expectations for complete removal
3. Comply with good software practices for data lifecycle management