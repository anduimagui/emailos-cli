# EmailOS Configure Command Documentation

The `mailos configure` command manages email account configuration with support for both global and local (project-specific) settings.

## Basic Usage

```bash
mailos configure
```

Interactive configuration wizard for setting up or modifying email settings.

## Command-Line Flags

### Configuration Scope

| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `--local` | Create/modify project-specific configuration | false | `mailos configure --local` |
| `--quick` | Quick configuration menu | false | `mailos configure --quick` |

### Direct Configuration

| Flag | Description | Example |
|------|-------------|---------|
| `--email` | Email address to configure | `--email john@example.com` |
| `--provider` | Email provider | `--provider gmail` |
| `--name` | Display name for emails | `--name "John Smith"` |
| `--from` | From email address (sender) | `--from noreply@company.com` |
| `--image` | Profile image path | `--image /path/to/profile.jpg` |
| `--ai` | AI CLI provider | `--ai claude-code` |

## Configuration Types

### Global Configuration
Located at `~/.email/config.json`
- Used by default for all EmailOS commands
- Shared across all projects
- Contains full email credentials

### Local Configuration
Located at `./.email/config.json`
- Project-specific settings
- Inherits from global configuration
- Override specific fields like display name or from address
- Automatically added to `.gitignore`

## Configuration Inheritance

Local configs inherit from global settings:

```
Global Config (~/. email/):
  - provider: gmail
  - email: john@example.com
  - password: ****
  - from_name: John Smith

Local Config (./.email/):
  - from_name: Project Bot     # Overrides global
  - from_email: bot@project.com # Project-specific sender
```

## Supported Providers

| Provider | Configuration Key | Notes |
|----------|------------------|-------|
| Gmail | `gmail` | Requires app-specific password |
| Outlook | `outlook` | Includes Hotmail, Live.com |
| Yahoo | `yahoo` | Requires app password |
| iCloud | `icloud` | Apple Mail |
| ProtonMail | `proton` | Bridge required |
| Fastmail | `fastmail` | App password recommended |
| Custom | `custom` | Manual SMTP/IMAP settings |

## AI CLI Providers

| Provider | Key | Description |
|----------|-----|-------------|
| Claude Code | `claude-code` | Claude via VS Code |
| Claude Yolo | `claude-code-yolo` | Claude without confirmations |
| OpenAI | `openai` | ChatGPT integration |
| Gemini | `gemini` | Google AI |
| OpenCode | `opencode` | Open source alternative |
| None | `none` | Disable AI features |

## Interactive Mode

Running `mailos configure` without flags starts interactive setup:

1. **Provider Selection**: Choose your email service
2. **Email Address**: Enter your email
3. **App Password**: Enter app-specific password
4. **Display Name**: Optional sender name
5. **Profile Image**: Optional profile image path
6. **From Email**: Optional custom sender address
7. **AI Provider**: Select AI integration

## Quick Configuration

```bash
mailos configure --quick
```

Provides a menu for quick changes:
- Update password
- Change display name
- Switch AI provider
- Modify from address

## Examples

### Initial Setup
```bash
mailos configure
# Follow interactive prompts
```

### Create Local Configuration
```bash
mailos configure --local
# Creates .email/config.json in current directory
```

### Direct Configuration
```bash
mailos configure \
  --email john@gmail.com \
  --provider gmail \
  --name "John Smith" \
  --ai claude-code
```

### Project-Specific Sender
```bash
# In project directory
mailos configure --local \
  --from noreply@project.com \
  --name "Project Notifications"
```

### Update Password Only
```bash
mailos configure --quick
# Select "Update Password" from menu
```

## App Passwords

Most providers require app-specific passwords:

### Gmail
1. Enable 2-factor authentication
2. Go to Google Account settings
3. Security → 2-Step Verification → App passwords
4. Generate password for "Mail"

### Outlook/Hotmail
1. Enable two-step verification
2. Go to Security settings
3. Advanced security → Create app password

### Yahoo
1. Enable two-step verification
2. Account Security → Generate app password
3. Select "Other App" and name it

### iCloud
1. Enable two-factor authentication
2. Sign in to Apple ID
3. Security → App-Specific Passwords
4. Generate password

## Configuration File Structure

```json
{
  "provider": "gmail",
  "email": "user@gmail.com",
  "password": "app-specific-password",
  "from_name": "Display Name",
  "from_email": "sender@example.com",
  "profile_image": "/absolute/path/to/image.jpg",
  "default_ai_cli": "claude-code",
  "license_key": "optional-license"
}
```

### Profile Image Configuration

The `profile_image` field allows you to include a profile picture in your emails:

- **Supported Formats**: PNG, JPG/JPEG, GIF, WebP
- **Path Type**: Absolute path to the image file
- **Email Embedding**: Image is embedded as base64 in HTML emails
- **Template Support**: Use `{{PROFILE_IMAGE}}` placeholder in custom templates
- **Display**: Appears as a circular profile photo (150px max width)

Example usage:
```bash
# During setup
mailos configure
# When prompted, enter: /Users/john/Pictures/profile.jpg

# Or directly via command line
mailos configure --image /Users/john/Pictures/profile.jpg

# For project-specific profile image
mailos configure --local --image ./assets/team-logo.png
```

## Security Best Practices

1. **Never commit credentials**: Local configs are auto-added to `.gitignore`
2. **Use app passwords**: Never use your main account password
3. **File permissions**: Config files are created with 600 permissions
4. **Encrypted storage**: Consider using encrypted filesystems
5. **Regular rotation**: Update app passwords periodically

## Multiple Accounts

To manage multiple email accounts:

1. **Global Default**: Set primary account globally
   ```bash
   mailos configure --email primary@example.com
   ```

2. **Project Override**: Use different account per project
   ```bash
   cd project1
   mailos configure --local --email project1@example.com
   
   cd ../project2
   mailos configure --local --email project2@example.com
   ```

## Troubleshooting

### Authentication Failed
- Verify app password is correct
- Check 2FA is enabled
- Ensure IMAP/SMTP access is enabled
- Try regenerating app password

### Local Config Not Working
- Ensure you're in the correct directory
- Check `.email/config.json` exists
- Verify JSON syntax is valid
- Check file permissions (should be 600)

### Provider Not Working
- Verify IMAP/SMTP settings are correct
- Check firewall/antivirus settings
- Some providers need "less secure apps" enabled
- Corporate accounts may have restrictions

### Inheritance Issues
- Local config only overrides specified fields
- Password cannot be inherited for security
- Run `mailos info` to see active configuration

## Advanced Configuration

### Custom Email Templates with Profile Images

When using custom email templates (via `mailos template`), you can include your profile image:

```html
<!-- Example template with profile image -->
<html>
  <body style="font-family: Arial, sans-serif;">
    <div style="text-align: center; padding: 20px;">
      {{PROFILE_IMAGE}}
      <h2>{{FROM_NAME}}</h2>
    </div>
    <div style="padding: 20px;">
      {{BODY}}
    </div>
    <div style="border-top: 1px solid #ccc; padding-top: 10px; margin-top: 20px;">
      <small>Sent with EmailOS</small>
    </div>
  </body>
</html>
```

The `{{PROFILE_IMAGE}}` placeholder will be replaced with your profile image embedded as base64 data.

### Custom SMTP/IMAP Settings
For providers not listed, use custom configuration:

```bash
mailos configure --provider custom
# You'll be prompted for:
# - SMTP Host
# - SMTP Port
# - IMAP Host  
# - IMAP Port
# - TLS/SSL settings
```

### Environment Variables
Override configuration with environment variables:
- `MAILOS_EMAIL`: Override email address
- `MAILOS_PROVIDER`: Override provider
- `MAILOS_AI_CLI`: Override AI provider

### Validation
Configuration is validated for:
- Valid email format
- Known provider settings
- Required fields present
- Password strength (warning only)

## See Also

- `mailos setup` - Initial setup wizard
- `mailos info` - Display current configuration
- `mailos local` - Create local configuration
- `mailos provider` - Configure AI provider