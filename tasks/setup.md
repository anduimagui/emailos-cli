# Setup Command Tasks

## Pending Tasks

### High Priority - Core Setup Flow
- [ ] **Add setup resume capability** - Allow users to resume interrupted setup process
- [ ] **Fix SMTP/IMAP config duplication** - Remove SMTP/IMAP from config.json since provider already determines these. Use provider definitions directly in commands.

### Medium Priority - Enhanced UX
- [ ] **Add quick setup for returning users** - Fast path for users who just need to re-enter credentials
- [ ] **Implement config migration** - Automatically upgrade old config formats to new structure
- [ ] **Add license status command** - `mailos license --status` to check license validity and expiration
- [ ] **Better offline mode feedback** - Clear messages when running in offline/grace period mode

### Low Priority - Polish
- [ ] **Add setup validation step** - Test email send/receive after setup completion
- [ ] **Implement setup profiles** - Save multiple email configurations and switch between them
- [ ] **Add setup import/export** - Backup and restore configuration (with encrypted passwords)

## Completed Tasks

### Initial Setup Features
- [x] Write MailOS in cool ASCII format to have a display for people
- [x] Confirm that MailOS is just a list of commands, and the user is responsible for managing their app password
- [x] After entering email/name, don't open the app password page directly - provide explanation first
- [x] Integrate Polar license key validation for commercial use
- [x] Add ability to configure the specific AI provider to use with the CLI

### Middleware Implementation (Completed 2025-08-06)
- [x] **Implement unified initialization check** - Created `EnsureInitialized()` middleware that runs before all commands
- [x] **Add auto-setup prompt** - When config/license is missing, automatically launches full setup flow with MailOS logo
- [x] **Centralize license validation** - All license checks now in middleware.go, removed scattered validation from individual commands

---

## Detailed Implementation Plans

### 1. Unified Initialization Check (High Priority)

**Problem**: Each command independently checks for config existence, leading to inconsistent behavior and poor UX when setup is needed.

**Solution**: Create middleware pattern that centralizes all initialization logic.

```go
// middleware.go
package mailos

func EnsureInitialized() error {
    // Check config exists
    if !ConfigExists() {
        fmt.Println("No configuration found.")
        fmt.Print("Would you like to run setup now? (Y/n): ")
        
        reader := bufio.NewReader(os.Stdin)
        response, _ := reader.ReadString('\n')
        response = strings.TrimSpace(strings.ToLower(response))
        
        if response != "n" && response != "no" {
            return Setup()
        }
        return fmt.Errorf("setup required to continue")
    }
    
    // Validate license
    config, err := LoadConfig()
    if err != nil {
        return err
    }
    
    if config.LicenseKey == "" {
        return triggerLicenseSetup()
    }
    
    // Quick validate with cache
    lm := GetLicenseManager()
    if err := lm.QuickValidate(config.LicenseKey); err != nil {
        return fmt.Errorf("license validation failed: %v", err)
    }
    
    return nil
}
```

**Integration**: Add `PreRunE` to all commands except setup:
```go
var sendCmd = &cobra.Command{
    Use:   "send",
    Short: "Send an email",
    PreRunE: func(cmd *cobra.Command, args []string) error {
        return mailos.EnsureInitialized()
    },
    RunE: existingRunE,
}
```

### 2. Auto-Setup Prompt (High Priority)

**Problem**: Users get error message but must manually run setup command.

**Solution**: Modify `NewClient()` to trigger setup automatically:
- Detect missing config
- Prompt user to run setup
- Continue with command after successful setup
- Cache setup completion to avoid repeated prompts

### 3. Quick Setup for Returning Users (Medium Priority)

**Problem**: Full setup wizard is tedious for users who just need to update credentials.

**Solution**: Add `--quick` flag to setup:
```bash
mailos setup --quick
```
- Skip ASCII art and explanations
- Pre-fill known values (email, provider)
- Only prompt for password/license
- Complete in under 30 seconds

### 4. Config Migration (Medium Priority)

**Problem**: Old config formats break when structure changes.

**Solution**: Version configs and auto-migrate:
```go
type Config struct {
    Version int `json:"version"`
    // ... other fields
}

func MigrateConfig(oldConfig map[string]interface{}) (*Config, error) {
    version := oldConfig["version"].(int)
    switch version {
    case 0, 1:
        return migrateV1ToV2(oldConfig)
    case 2:
        return migrateV2ToV3(oldConfig)
    }
}
```

---

## Current Implementation Status

### Middleware System ✅ NEW
- ✅ Created `middleware.go` with centralized validation
- ✅ `EnsureInitialized()` checks config and license before commands run
- ✅ Auto-launches setup flow when missing config/license
- ✅ Added `PreRunE` hooks to all main commands
- ✅ Removed duplicate validation from `send.go`, `read.go`, etc.
- ✅ Added `IsInGracePeriod()` for 7-day offline operation

### ASCII Art
✅ Cool ASCII banner displays "MailOS" at setup start

### License Validation
✅ Integrated with Polar billing system
✅ 24-hour cache for offline usage
✅ 7-day grace period for extended offline operation
✅ License stored in config.json
✅ Centralized validation in middleware

### Security Disclaimer
✅ Comprehensive security notice requiring acknowledgment
✅ Clear explanation of local-only storage

### App Password Education
✅ Detailed explanation comparing to API keys
✅ Direct links to provider password pages
✅ User confirmation before browser launch

### AI CLI Configuration
✅ Support for multiple AI providers:
  - Claude Code (with YOLO mode option)
  - OpenAI Codex
  - Gemini CLI
  - OpenCode
  - Manual only mode

---

## Architecture Analysis

### Current State

#### Package Structure
- **Entry point**: `cmd/mailos/main.go` - Cobra CLI commands
- **Core logic**: Root `mailos` package
  - `client.go`: Main client interface
  - `config.go`: Configuration management
  - `setup.go`: Interactive setup wizard
  - `license.go`: Polar license validation

#### Configuration Flow
1. Check local `.email/config.json`
2. Fallback to global `~/.email/config.json`
3. Error if neither exists

#### License Validation
- **Full validation**: Direct Polar API call during setup
- **Quick validation**: Uses cache, periodic checks during commands
- **Grace period**: 7 days offline operation after successful validation

### Issues Identified

1. ~~**No automatic setup trigger**~~ ✅ FIXED - Commands now auto-launch setup
2. ~~**Scattered validation**~~ ✅ FIXED - All validation centralized in middleware.go
3. ~~**Poor error recovery**~~ ✅ FIXED - Auto-setup provides clear path forward
4. ~~**Inconsistent UX**~~ ✅ FIXED - All commands use same middleware validation

### Proposed Solutions

#### Solution 1: Middleware Pattern ✨ RECOMMENDED
- Centralized initialization logic
- Consistent behavior across all commands
- Easy to maintain and extend
- Clean separation of concerns

#### Solution 2: Client Factory Pattern
- Encapsulates initialization in client creation
- Works but less flexible than middleware
- Harder to customize per-command behavior

### Implementation Roadmap

**Phase 1 - Foundation** ✅ COMPLETE
- [x] Implement `EnsureInitialized()` middleware
- [x] Add to all commands via `PreRunE`
- [x] Test missing config scenarios
- [x] Auto-launch setup when config/license missing

**Phase 2 - Enhancement** (In Progress)
- [ ] Quick setup mode for returning users
- [ ] Config migration system
- [ ] Setup resume capability
- [ ] Fix SMTP/IMAP config duplication

**Phase 3 - Polish** (Pending)
- [ ] License status command
- [ ] Better offline feedback messages
- [ ] Setup validation tests

### Benefits

1. **Consistent UX** - Same behavior across all commands
2. **Better onboarding** - Auto-setup reduces friction
3. **Maintainability** - Single source of truth for initialization
4. **Flexibility** - Easy to add new requirements
5. **Reliability** - Centralized error handling and recovery

### Migration Notes

- Preserve backward compatibility with existing configs
- Support gradual migration of old formats
- Maintain existing CLI interface
- Document changes clearly for users