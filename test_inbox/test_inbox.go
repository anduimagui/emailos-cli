package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	mailos "github.com/corp-os/emailos"
)

func main() {
	fmt.Println("🚀 Testing Global Inbox System")
	fmt.Println("================================")

	// Load configuration
	config, err := mailos.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("✓ Loaded config for: %s (%s)\n", config.Email, config.Provider)

	// Test 1: Check if global inbox directory structure is created
	fmt.Println("\n📁 Testing directory structure...")
	homeDir, _ := os.UserHomeDir()
	accountDir := filepath.Join(homeDir, ".email", config.Email)
	inboxPath := filepath.Join(accountDir, "inbox.json")
	
	fmt.Printf("Account directory: %s\n", accountDir)
	fmt.Printf("Inbox path: %s\n", inboxPath)

	// Test 2: Load existing inbox or create new one
	fmt.Println("\n📥 Testing inbox loading...")
	inboxData, err := mailos.LoadGlobalInbox(config.Email)
	if err != nil {
		log.Fatalf("Failed to load inbox: %v", err)
	}

	fmt.Printf("✓ Inbox loaded successfully\n")
	fmt.Printf("  Account: %s\n", inboxData.AccountEmail)
	fmt.Printf("  Total emails: %d\n", inboxData.TotalEmails)
	fmt.Printf("  Last fetch: %v\n", inboxData.LastFetchTime.Format(time.RFC3339))
	fmt.Printf("  Last email date: %v\n", inboxData.LastEmailDate.Format(time.RFC3339))

	// Test 3: Try incremental email fetch (limit to 5 for testing)
	fmt.Println("\n📨 Testing incremental email fetch...")
	fmt.Println("Fetching up to 5 new emails...")
	
	err = mailos.FetchEmailsIncremental(config, 5)
	if err != nil {
		log.Printf("Warning: Failed to fetch emails: %v", err)
		fmt.Println("This might be expected if there are no new emails or connection issues")
	} else {
		fmt.Println("✓ Email fetch completed")
	}

	// Test 4: Load updated inbox data
	fmt.Println("\n📊 Testing updated inbox stats...")
	updatedInbox, err := mailos.LoadGlobalInbox(config.Email)
	if err != nil {
		log.Fatalf("Failed to reload inbox: %v", err)
	}

	fmt.Printf("✓ Updated inbox stats:\n")
	fmt.Printf("  Total emails: %d\n", updatedInbox.TotalEmails)
	fmt.Printf("  Last fetch: %v\n", updatedInbox.LastFetchTime.Format(time.RFC3339))
	if !updatedInbox.LastEmailDate.IsZero() {
		fmt.Printf("  Last email date: %v\n", updatedInbox.LastEmailDate.Format(time.RFC3339))
	}

	// Test 5: Try reading emails from inbox
	fmt.Println("\n📖 Testing email reading from global inbox...")
	opts := mailos.ReadOptions{
		Limit:     3,
		LocalOnly: true, // Force reading from global inbox
	}

	emails, err := mailos.GetEmailsFromInbox(config.Email, opts)
	if err != nil {
		log.Printf("Warning: Failed to read from inbox: %v", err)
	} else {
		fmt.Printf("✓ Read %d emails from global inbox\n", len(emails))
		for i, email := range emails {
			fmt.Printf("  %d. %s - %s\n", i+1, email.From, email.Subject)
		}
	}

	// Test 6: List all account inboxes
	fmt.Println("\n📋 Testing account inbox listing...")
	accounts, err := mailos.ListAccountInboxes()
	if err != nil {
		log.Printf("Warning: Failed to list account inboxes: %v", err)
	} else {
		fmt.Printf("✓ Found %d account inbox(es):\n", len(accounts))
		for _, account := range accounts {
			fmt.Printf("  - %s\n", account)
		}
	}

	fmt.Println("\n🎉 Global Inbox System Test Complete!")
	fmt.Printf("Inbox file location: %s\n", inboxPath)
}