package mailos

// Application information
const (
	AppName        = "EmailOS"
	AppDisplayName = "EmailOS"
	AppSite        = "email-os.com"
	AppVersion     = "0.1.11" // Update this when releasing new versions
	AppDescription = "AI-powered email management system"
	GitHubRepo     = "https://github.com/corp-os/emailos"
)

// Configuration file paths
const (
	ConfigDir      = ".email"
	ConfigFileName = "config.json"
	SlashConfigFileName = ".slash_config.json"
	LicenseFileName = ".license"
	GitIgnoreEntry = ".email"
)

// Provider URLs - Gmail
const (
	GmailWebURL      = "https://mail.google.com/mail/u/0/"
	GmailInboxURL    = "https://mail.google.com/mail/u/0/#inbox"
	GmailSentURL     = "https://mail.google.com/mail/u/0/#sent"
	GmailDraftsURL   = "https://mail.google.com/mail/u/0/#drafts"
	GmailAllMailURL  = "https://mail.google.com/mail/u/0/#all"
	GmailAppPasswordURL = "https://myaccount.google.com/apppasswords"
)

// Provider URLs - Fastmail
const (
	FastmailWebURL      = "https://app.fastmail.com/mail/"
	FastmailInboxURL    = "https://app.fastmail.com/mail/Inbox"
	FastmailSentURL     = "https://app.fastmail.com/mail/Sent"
	FastmailDraftsURL   = "https://app.fastmail.com/mail/Drafts"
	FastmailAllMailURL  = "https://app.fastmail.com/mail/All"
	FastmailAppPasswordURL = "https://app.fastmail.com/settings/security/apps/new"
)

// Provider URLs - Outlook
const (
	OutlookWebURL      = "https://outlook.live.com/mail/0/"
	OutlookInboxURL    = "https://outlook.live.com/mail/0/inbox"
	OutlookSentURL     = "https://outlook.live.com/mail/0/sentitems"
	OutlookDraftsURL   = "https://outlook.live.com/mail/0/drafts"
	OutlookAppPasswordURL = "https://account.microsoft.com/security"
)

// Provider URLs - Yahoo
const (
	YahooWebURL      = "https://mail.yahoo.com/d/"
	YahooInboxURL    = "https://mail.yahoo.com/d/folders/1"
	YahooSentURL     = "https://mail.yahoo.com/d/folders/2"
	YahooDraftsURL   = "https://mail.yahoo.com/d/folders/3"
	YahooAppPasswordURL = "https://login.yahoo.com/account/security"
)

// Provider URLs - Zoho
const (
	ZohoWebURL      = "https://mail.zoho.com/zm/"
	ZohoInboxURL    = "https://mail.zoho.com/zm/#mail/folder/inbox"
	ZohoSentURL     = "https://mail.zoho.com/zm/#mail/folder/sent"
	ZohoDraftsURL   = "https://mail.zoho.com/zm/#mail/folder/drafts"
	ZohoAppPasswordURL = "https://accounts.zoho.eu/home#security/app_password"
)

// Provider keys
const (
	ProviderGmail    = "gmail"
	ProviderFastmail = "fastmail"
	ProviderOutlook  = "outlook"
	ProviderYahoo    = "yahoo"
	ProviderZoho     = "zoho"
)

// SMTP/IMAP Ports
const (
	SMTPPortTLS = 587
	SMTPPortSSL = 465
	IMAPPortSSL = 993
)

// AI Provider keys
const (
	AIProviderClaudeCode     = "claude-code"
	AIProviderClaudeCodeYolo = "claude-code-yolo"
	AIProviderOpenAI         = "openai-codex"
	AIProviderGemini         = "gemini-cli"
	AIProviderOpenCode       = "opencode"
	AIProviderNone           = "none"
)

// AI Provider display names
const (
	AIDisplayClaudeCode     = "Claude Code"
	AIDisplayClaudeCodeYolo = "Claude Code YOLO Mode"
	AIDisplayOpenAI         = "OpenAI"
	AIDisplayGemini         = "Gemini"
	AIDisplayOpenCode       = "OpenCode"
	AIDisplayNone           = "None"
)

// UI Mode environment variables
const (
	EnvUseBubbleTea     = "MAILOS_USE_BUBBLETEA"
	EnvUseInk           = "MAILOS_USE_INK"
	EnvSuggestionMode   = "MAILOS_SUGGESTION_MODE"
	EnvShowLogo         = "MAILOS_SHOW_LOGO"
	EnvNoLogo           = "MAILOS_NO_LOGO"
)

// Suggestion modes
const (
	SuggestionModeDynamic = "dynamic"
	SuggestionModeSimple  = "simple"
	SuggestionModeLive    = "live"
	SuggestionModeClean   = "clean"
)

// Default values
const (
	DefaultEmailLimit = 10
	DefaultReportLimit = 1000
	DefaultDeleteLimit = 100
	DefaultSyncInterval = 300 // seconds
)

// Terminal colors and formatting
const (
	ColorReset   = "\033[0m"
	ColorBold    = "\033[1m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"
)

// Icons for UI
const (
	IconEmail       = "📧"
	IconSend        = "✉️"
	IconInbox       = "📥"
	IconSent        = "📤"
	IconDrafts      = "📝"
	IconReport      = "📊"
	IconUnsubscribe = "🔗"
	IconDelete      = "🗑️"
	IconCheck       = "✓"
	IconSettings    = "⚙️"
	IconAI          = "🤖"
	IconInfo        = "ℹ️"
	IconHelp        = "❓"
	IconExit        = "👋"
	IconFolder      = "📁"
	IconTemplate    = "📄"
	IconSuccess     = "✅"
	IconError       = "❌"
	IconWarning     = "⚠️"
	IconAccount     = "📬"
	IconFromEmail   = "📮"
)

// Time formats
const (
	TimeFormatShort = "Jan 2, 2006"
	TimeFormatLong  = "January 2, 2006 at 3:04 PM"
	TimeFormatFile  = "2006-01-02-15-04-05"
	TimeFormatISO   = "2006-01-02T15:04:05Z07:00"
)

// File extensions
const (
	ExtMarkdown = ".md"
	ExtJSON     = ".json"
	ExtHTML     = ".html"
	ExtText     = ".txt"
)

// Error messages
const (
	ErrNoConfig         = "no configuration found - please run setup first"
	ErrInvalidProvider  = "invalid email provider specified"
	ErrAuthFailed       = "authentication failed - please check your credentials"
	ErrConnectionFailed = "failed to connect to email server"
	ErrNoEmails         = "no emails found"
)

// Success messages
const (
	MsgConfigSaved     = "Configuration saved successfully"
	MsgEmailSent       = "Email sent successfully"
	MsgEmailsSaved     = "Emails saved successfully"
	MsgEmailsDeleted   = "Emails deleted successfully"
	MsgEmailsMarkedRead = "Emails marked as read successfully"
)