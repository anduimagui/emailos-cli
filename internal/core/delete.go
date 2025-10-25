package core

import (
	"crypto/tls"
	"fmt"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

// DeleteEmails deletes the given email IDs from the INBOX folder
func DeleteEmails(ids []uint32, config ConfigInterface) error {
	return DeleteEmailsFromFolder(ids, "INBOX", config)
}

// DeleteDrafts deletes the given draft IDs from the Drafts folder
func DeleteDrafts(ids []uint32, config ConfigInterface) error {
	return DeleteEmailsFromFolder(ids, "Drafts", config)
}

// DeleteEmailsFromFolder deletes emails from a specific folder
func DeleteEmailsFromFolder(ids []uint32, folder string, config ConfigInterface) error {
	// Get IMAP settings from provider
	imapHost, imapPort, err := config.GetIMAPSettings()
	if err != nil {
		return fmt.Errorf("failed to get IMAP settings: %v", err)
	}

	// Connect to IMAP server
	var c *client.Client
	if imapPort == 993 {
		tlsConfig := &tls.Config{ServerName: imapHost}
		c, err = client.DialTLS(fmt.Sprintf("%s:%d", imapHost, imapPort), tlsConfig)
	} else {
		c, err = client.Dial(fmt.Sprintf("%s:%d", imapHost, imapPort))
	}
	if err != nil {
		return fmt.Errorf("failed to connect to IMAP server: %v", err)
	}
	defer c.Logout()

	// Login
	if err := c.Login(config.GetEmail(), config.GetPassword()); err != nil {
		return fmt.Errorf("failed to login: %v", err)
	}

	// Select the specified folder
	_, err = c.Select(folder, false)
	if err != nil {
		// Try with [Gmail]/Drafts for Gmail
		if folder == "Drafts" && config.GetProvider() == "gmail" {
			_, err = c.Select("[Gmail]/Drafts", false)
		}
		if err != nil {
			return fmt.Errorf("failed to select %s folder: %v", folder, err)
		}
	}

	// Create sequence set
	seqSet := new(imap.SeqSet)
	for _, id := range ids {
		seqSet.AddNum(id)
	}

	// Mark as deleted
	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.DeletedFlag}
	if err := c.Store(seqSet, item, flags, nil); err != nil {
		return fmt.Errorf("failed to mark messages for deletion: %v", err)
	}

	// Expunge to permanently delete
	if err := c.Expunge(nil); err != nil {
		return fmt.Errorf("failed to expunge deleted messages: %v", err)
	}

	return nil
}

// ConfigInterface defines the interface for configuration access needed by delete operations
type ConfigInterface interface {
	GetIMAPSettings() (string, int, error)
	GetEmail() string
	GetPassword() string
	GetProvider() string
}

// ConfigWrapper wraps the main Config to implement ConfigInterface
type ConfigWrapper struct {
	Email    string
	Password string
	Provider string
	GetIMAPSettingsFunc func() (string, int, error)
}

func (cw *ConfigWrapper) GetIMAPSettings() (string, int, error) {
	return cw.GetIMAPSettingsFunc()
}

func (cw *ConfigWrapper) GetEmail() string {
	return cw.Email
}

func (cw *ConfigWrapper) GetPassword() string {
	return cw.Password
}

func (cw *ConfigWrapper) GetProvider() string {
	return cw.Provider
}