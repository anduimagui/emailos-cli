# EmailOS Accounts Command Documentation

The `mailos accounts` command manages multiple email accounts, aliases, and sub-emails for use across your email workflow.

## Basic Usage

```bash
mailos accounts [flags]
```

## Command-Line Flags

- `--list` - List all available accounts (default behavior)
- `--set <email>` - Set session default account
- `--clear` - Clear session default account

## Account Types and Organization

EmailOS automatically organizes your accounts into three categories:

### Primary Accounts (🏠)
The main account for each email provider, used for authentication.

### Sub-emails (↳)  
Additional email addresses that use the same provider credentials as a primary account. These include:
- Email aliases configured at your provider
- Additional addresses from the same domain
- Forwarding addresses that share authentication

### Secondary Accounts
Completely separate accounts from different email providers.

## Managing Aliases and Sub-Emails

### Adding Aliases Manually

**Important**: EmailOS does not automatically sync aliases from email providers. You must add them manually.

To add an alias or sub-email:

```bash
mailos accounts --set your-alias@domain.com
```

When you specify an email that doesn't exist locally, EmailOS will:
1. Detect it's not configured
2. Prompt: "Account 'your-alias@domain.com' not found. Would you like to add this account? (y/N):"
3. If you answer "y", guide you through the setup process

### Provider-Specific Setup

#### Fastmail
```bash
# Add your Fastmail alias
mailos accounts --set alias@yourdomain.com

# Use the same app password as your main Fastmail account
# EmailOS will automatically label it as "Sub-email"
```

#### Gmail
```bash
# Add your Gmail alias (configured in Gmail Settings > Accounts)
mailos accounts --set alias@gmail.com

# Use the same app password as your main Gmail account
```

#### Microsoft/Office 365
```bash
# Add your Office 365 alias
mailos accounts --set alias@yourcompany.com

# Use the same credentials as your main Office 365 account
```

### Alias Syncing Limitations

**No Automatic Discovery**: EmailOS cannot automatically detect aliases from email providers because:
- Most providers don't expose alias information via IMAP/POP3
- API access varies significantly between providers
- Security policies often restrict alias enumeration

**Manual Configuration Required**: You must add each alias individually using the `mailos accounts --set` command.

**Same Provider = Same Credentials**: When adding an alias from the same provider as an existing account, use the same authentication credentials (app password, etc.).

## Account Workflow Examples

### Example 1: Adding a Fastmail Alias

```bash
# You have: john@example.com (primary Fastmail account)
# You want: support@example.com (Fastmail alias)

mailos accounts --set support@example.com
# Prompts to add account -> Yes
# Provider detected: fastmail
# Enter app password: [same as john@example.com]
# Result: support@example.com added as "Sub-email"
```

### Example 2: Listing All Accounts

```bash
mailos accounts --list
```

Output:
```
Available Accounts:
==================

Fastmail:
  🏠 john@example.com (Primary)
  ↳  support@example.com (Sub-email)

Gmail:
  🏠 john.doe@gmail.com (Primary)
```

### Example 3: Setting Session Default

```bash
# Set default account for current session
mailos accounts --set support@example.com

# Now all commands use support@example.com unless overridden
mailos send --to customer@company.com --subject "Support Response"
```

### Example 4: Using Specific Account for Sending

```bash
# Send from specific account regardless of default
mailos send --from support@example.com --to customer@company.com --subject "Help"
```

## Configuration Storage

### Global Configuration
- **Location**: `~/.email/config.json`
- **Scope**: Available across all projects and directories
- **Contains**: All account credentials and settings

### Local Configuration  
- **Location**: `./.email/config.json` (in current directory)
- **Scope**: Project-specific overrides
- **Contains**: Local account preferences and defaults

### Configuration Inheritance
Local configurations inherit missing settings from global configuration, allowing project-specific account preferences while maintaining global credentials.

## Troubleshooting

### "Account not found" Error
```bash
mailos send --from alias@domain.com --to recipient@example.com --subject "Test"
# Error: account 'alias@domain.com' not found. Available accounts: [list]
```

**Solution**: Add the alias using `mailos accounts --set alias@domain.com`

### "Authentication failed" for Sub-email
**Problem**: Using different credentials for an alias from the same provider.
**Solution**: Use the same app password/credentials as the primary account from that provider.

### Sub-email Not Labeled Correctly
**Problem**: Alias appears as separate account instead of sub-email.
**Solution**: Ensure you're using the exact same provider credentials. EmailOS automatically detects same-provider accounts and labels them as sub-emails.

### Cannot Send from Alias
**Problem**: Email sends from primary account instead of alias.
**Solution**: 
1. Verify alias is added: `mailos accounts --list`
2. Use explicit `--from` flag: `mailos send --from alias@domain.com`
3. Check that your email provider supports sending from that alias

## Best Practices

1. **Add aliases as soon as you configure them at your provider**
2. **Use the same credentials for all accounts from the same provider**
3. **Test sending from aliases before using them in production**
4. **Keep a list of your provider-configured aliases for reference**
5. **Use descriptive local defaults for different projects**

## Related Commands

- `mailos configure` - Initial account setup and configuration
- `mailos send --from <account>` - Send from specific account
- `mailos help configure` - Detailed configuration documentation
- `mailos help send` - Email sending documentation

## Integration with Other Commands

All EmailOS commands support the `--account` flag to specify which account to use:

```bash
mailos sync --account alias@domain.com
mailos read --account alias@domain.com  
mailos search --account alias@domain.com
```

For commands that don't specify an account, EmailOS uses this priority order:
1. Local directory default (`.email/config.json`)
2. Session default (set with `mailos accounts --set`)
3. Global default account
4. First available account