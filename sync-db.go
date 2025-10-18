package mailos

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DatabaseManager struct {
	db          *sql.DB
	accountEmail string
	dbPath      string
}

func NewDatabaseManager(accountEmail string) (*DatabaseManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	accountDir := filepath.Join(homeDir, ".email", accountEmail)
	if err := os.MkdirAll(accountDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create account directory: %v", err)
	}

	dbPath := filepath.Join(accountDir, "archive.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	dm := &DatabaseManager{
		db:          db,
		accountEmail: accountEmail,
		dbPath:      dbPath,
	}

	if err := dm.createTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %v", err)
	}

	return dm, nil
}

func (dm *DatabaseManager) createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS emails (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		message_id TEXT UNIQUE NOT NULL,
		from_address TEXT NOT NULL,
		to_addresses TEXT NOT NULL,
		subject TEXT NOT NULL,
		date_sent DATETIME NOT NULL,
		body_text TEXT,
		body_html TEXT,
		attachments TEXT,
		attachment_data BLOB,
		in_reply_to TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_emails_message_id ON emails(message_id);
	CREATE INDEX IF NOT EXISTS idx_emails_from ON emails(from_address);
	CREATE INDEX IF NOT EXISTS idx_emails_date ON emails(date_sent);
	CREATE INDEX IF NOT EXISTS idx_emails_subject ON emails(subject);

	CREATE TABLE IF NOT EXISTS sync_metadata (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		account_email TEXT NOT NULL,
		last_sync_time DATETIME NOT NULL,
		total_emails INTEGER DEFAULT 0,
		last_email_date DATETIME,
		sync_version INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_sync_metadata_account ON sync_metadata(account_email);
	`

	_, err := dm.db.Exec(schema)
	return err
}

func (dm *DatabaseManager) Close() error {
	return dm.db.Close()
}

func (dm *DatabaseManager) SyncEmailsFromInbox() error {
	inboxData, err := LoadGlobalInbox(dm.accountEmail)
	if err != nil {
		return fmt.Errorf("failed to load inbox data: %v", err)
	}

	tx, err := dm.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO emails (
			message_id, from_address, to_addresses, subject, date_sent,
			body_text, body_html, attachments, attachment_data, in_reply_to
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	syncedCount := 0
	for _, email := range inboxData.Emails {
		toAddresses, _ := json.Marshal(email.To)
		attachments, _ := json.Marshal(email.Attachments)
		attachmentData, _ := json.Marshal(email.AttachmentData)

		_, err := stmt.Exec(
			email.MessageID,
			email.From,
			string(toAddresses),
			email.Subject,
			email.Date,
			email.Body,
			email.BodyHTML,
			string(attachments),
			attachmentData,
			email.InReplyTo,
		)
		if err != nil {
			fmt.Printf("Warning: failed to insert email %s: %v\n", email.MessageID, err)
			continue
		}
		syncedCount++
	}

	if err := dm.updateSyncMetadata(tx, inboxData); err != nil {
		return fmt.Errorf("failed to update sync metadata: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	fmt.Printf("✓ Synced %d emails to database for %s\n", syncedCount, dm.accountEmail)
	fmt.Printf("✓ Database location: %s\n", dm.dbPath)

	return nil
}

func (dm *DatabaseManager) updateSyncMetadata(tx *sql.Tx, inboxData *InboxData) error {
	_, err := tx.Exec(`
		INSERT OR REPLACE INTO sync_metadata (
			account_email, last_sync_time, total_emails, last_email_date, sync_version
		) VALUES (?, ?, ?, ?, ?)
	`, dm.accountEmail, time.Now(), len(inboxData.Emails), inboxData.LastEmailDate, inboxData.LastSyncVersion)
	
	return err
}

func (dm *DatabaseManager) GetEmailsFromDB(opts ReadOptions) ([]*Email, error) {
	query := `
		SELECT message_id, from_address, to_addresses, subject, date_sent,
			   body_text, body_html, attachments, attachment_data, in_reply_to
		FROM emails
		WHERE 1=1
	`
	args := []interface{}{}

	if opts.FromAddress != "" {
		query += " AND from_address LIKE ?"
		args = append(args, "%"+opts.FromAddress+"%")
	}

	if opts.Subject != "" {
		query += " AND subject LIKE ?"
		args = append(args, "%"+opts.Subject+"%")
	}

	if !opts.Since.IsZero() {
		query += " AND date_sent >= ?"
		args = append(args, opts.Since)
	}

	query += " ORDER BY date_sent DESC"

	if opts.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, opts.Limit)
	}

	rows, err := dm.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query emails: %v", err)
	}
	defer rows.Close()

	var emails []*Email
	for rows.Next() {
		var email Email
		var toAddressesJSON, attachmentsJSON, attachmentDataJSON string

		err := rows.Scan(
			&email.MessageID,
			&email.From,
			&toAddressesJSON,
			&email.Subject,
			&email.Date,
			&email.Body,
			&email.BodyHTML,
			&attachmentsJSON,
			&attachmentDataJSON,
			&email.InReplyTo,
		)
		if err != nil {
			continue
		}

		json.Unmarshal([]byte(toAddressesJSON), &email.To)
		json.Unmarshal([]byte(attachmentsJSON), &email.Attachments)
		json.Unmarshal([]byte(attachmentDataJSON), &email.AttachmentData)

		emails = append(emails, &email)
	}

	return emails, nil
}

func (dm *DatabaseManager) GetDatabaseStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalEmails int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM emails").Scan(&totalEmails)
	if err != nil {
		return nil, err
	}
	stats["total_emails"] = totalEmails

	var oldestEmail, newestEmail time.Time
	err = dm.db.QueryRow("SELECT MIN(date_sent), MAX(date_sent) FROM emails").Scan(&oldestEmail, &newestEmail)
	if err == nil {
		stats["oldest_email"] = oldestEmail
		stats["newest_email"] = newestEmail
	}

	var dbSize int64
	if fileInfo, err := os.Stat(dm.dbPath); err == nil {
		dbSize = fileInfo.Size()
	}
	stats["database_size_bytes"] = dbSize
	stats["database_path"] = dm.dbPath

	return stats, nil
}

func SyncEmailsToDB(accountEmail string) error {
	dm, err := NewDatabaseManager(accountEmail)
	if err != nil {
		return err
	}
	defer dm.Close()

	return dm.SyncEmailsFromInbox()
}

func SyncAllAccountsToDB() error {
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	accounts := GetAllAccounts(config)
	if len(accounts) == 0 {
		return fmt.Errorf("no accounts configured")
	}

	fmt.Printf("Syncing emails to database for %d accounts...\n", len(accounts))

	for _, account := range accounts {
		fmt.Printf("\n--- Syncing %s to database ---\n", account.Email)
		
		if err := SyncEmailsToDB(account.Email); err != nil {
			fmt.Printf("Error syncing %s to database: %v\n", account.Email, err)
			continue
		}
	}

	fmt.Printf("\n✓ Finished syncing all accounts to database\n")
	return nil
}

func QueryEmailsFromDB(accountEmail string, opts ReadOptions) ([]*Email, error) {
	dm, err := NewDatabaseManager(accountEmail)
	if err != nil {
		return nil, err
	}
	defer dm.Close()

	return dm.GetEmailsFromDB(opts)
}

func GetDBStats(accountEmail string) (map[string]interface{}, error) {
	dm, err := NewDatabaseManager(accountEmail)
	if err != nil {
		return nil, err
	}
	defer dm.Close()

	return dm.GetDatabaseStats()
}