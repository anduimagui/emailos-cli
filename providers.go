package mailos

import (
	"fmt"
)

type Provider struct {
	Name            string
	SMTPHost        string
	SMTPPort        int
	SMTPUseTLS      bool
	SMTPUseSSL      bool
	IMAPHost        string
	IMAPPort        int
	AppPasswordURL  string
	AppPasswordHelp string
}

var Providers = map[string]Provider{
	"gmail": {
		Name:            "Gmail",
		SMTPHost:        "smtp.gmail.com",
		SMTPPort:        587,
		SMTPUseTLS:      true,
		IMAPHost:        "imap.gmail.com",
		IMAPPort:        993,
		AppPasswordURL:  "https://myaccount.google.com/apppasswords",
		AppPasswordHelp: "You need to enable 2-factor authentication and create an app password",
	},
	"fastmail": {
		Name:            "Fastmail",
		SMTPHost:        "smtp.fastmail.com",
		SMTPPort:        465,
		SMTPUseSSL:      true,
		IMAPHost:        "imap.fastmail.com",
		IMAPPort:        993,
		AppPasswordURL:  "https://app.fastmail.com/settings/security/apps/new",
		AppPasswordHelp: "Create an app-specific password in Settings > Security > Device Passwords",
	},
	"zoho": {
		Name:            "Zoho Mail",
		SMTPHost:        "smtp.zoho.com",
		SMTPPort:        465,
		SMTPUseSSL:      true,
		IMAPHost:        "imap.zoho.com",
		IMAPPort:        993,
		AppPasswordURL:  "https://accounts.zoho.eu/home#security/app_password",
		AppPasswordHelp: "Generate an application-specific password in Security settings",
	},
	"outlook": {
		Name:            "Outlook/Hotmail",
		SMTPHost:        "smtp-mail.outlook.com",
		SMTPPort:        587,
		SMTPUseTLS:      true,
		IMAPHost:        "outlook.office365.com",
		IMAPPort:        993,
		AppPasswordURL:  "https://account.microsoft.com/security",
		AppPasswordHelp: "Enable two-step verification and create an app password",
	},
	"yahoo": {
		Name:            "Yahoo Mail",
		SMTPHost:        "smtp.mail.yahoo.com",
		SMTPPort:        587,
		SMTPUseTLS:      true,
		IMAPHost:        "imap.mail.yahoo.com",
		IMAPPort:        993,
		AppPasswordURL:  "https://login.yahoo.com/account/security",
		AppPasswordHelp: "Generate an app password in Account Security settings",
	},
}

func GetProviderNames() []string {
	names := make([]string, 0, len(Providers))
	for key, provider := range Providers {
		names = append(names, fmt.Sprintf("%s (%s)", provider.Name, key))
	}
	return names
}

func GetProviderKeys() []string {
	// Return providers in preferred order: Gmail, Outlook, Fastmail first
	preferredOrder := []string{"gmail", "outlook", "fastmail"}
	
	// Add remaining providers
	otherProviders := []string{}
	for key := range Providers {
		isPreferred := false
		for _, preferred := range preferredOrder {
			if key == preferred {
				isPreferred = true
				break
			}
		}
		if !isPreferred {
			otherProviders = append(otherProviders, key)
		}
	}
	
	// Combine preferred and other providers
	allKeys := append(preferredOrder, otherProviders...)
	return allKeys
}
