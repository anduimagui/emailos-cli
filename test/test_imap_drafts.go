package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"strings"
	"time"
	
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/anduimagui/emailos"
)

func main() {
	// Load config
	config, err := mailos.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}
	
	// Get IMAP settings
	imapHost, imapPort, err := config.GetIMAPSettings()
	if err != nil {
		log.Fatal("Failed to get IMAP settings:", err)
	}
	
	addr := fmt.Sprintf("%s:%d", imapHost, imapPort)
	
	// Connect with TLS
	tlsConfig := &tls.Config{ServerName: imapHost}
	c, err := client.DialTLS(addr, tlsConfig)
	if err != nil {
		// Try without TLS
		c, err = client.Dial(addr)
		if err != nil {
			log.Fatal("Failed to connect:", err)
		}
		
		// Start TLS if supported
		if ok, _ := c.SupportStartTLS(); ok {
			if err := c.StartTLS(tlsConfig); err != nil {
				log.Fatal("Failed to start TLS:", err)
			}
		}
	}
	defer c.Logout()
	
	// Login
	if err := c.Login(config.Email, config.Password); err != nil {
		log.Fatal("Failed to login:", err)
	}
	
	fmt.Println("✓ Connected to IMAP server")
	
	// List all folders to find Drafts
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()
	
	var draftFolder string
	fmt.Println("\n📁 Available folders:")
	for m := range mailboxes {
		fmt.Printf("  - %s", m.Name)
		if strings.Contains(strings.ToLower(m.Name), "draft") {
			fmt.Printf(" ← Drafts folder found!")
			if draftFolder == "" {
				draftFolder = m.Name
			}
		}
		fmt.Println()
	}
	
	if err := <-done; err != nil {
		log.Fatal("Failed to list folders:", err)
	}
	
	if draftFolder == "" {
		fmt.Println("\n⚠️  No Drafts folder found")
		return
	}
	
	// Select the Drafts folder
	mbox, err := c.Select(draftFolder, false)
	if err != nil {
		log.Fatal("Failed to select Drafts folder:", err)
	}
	
	fmt.Printf("\n📮 Drafts folder: %s (contains %d messages)\n", draftFolder, mbox.Messages)
	
	if mbox.Messages == 0 {
		fmt.Println("No drafts found in IMAP folder")
		return
	}
	
	// Fetch recent drafts
	seqset := new(imap.SeqSet)
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > 5 {
		from = mbox.Messages - 4
	}
	seqset.AddRange(from, to)
	
	messages := make(chan *imap.Message, 10)
	done = make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags}, messages)
	}()
	
	fmt.Println("\n📧 Recent drafts in IMAP folder:")
	fmt.Println(strings.Repeat("-", 80))
	
	count := 0
	for msg := range messages {
		count++
		env := msg.Envelope
		
		// Format date
		dateStr := "Unknown date"
		if !env.Date.IsZero() {
			dateStr = env.Date.Format("2006-01-02 15:04:05")
		}
		
		// Get To addresses
		toAddrs := []string{}
		for _, addr := range env.To {
			if addr.PersonalName != "" {
				toAddrs = append(toAddrs, fmt.Sprintf("%s <%s@%s>", addr.PersonalName, addr.MailboxName, addr.HostName))
			} else {
				toAddrs = append(toAddrs, fmt.Sprintf("%s@%s", addr.MailboxName, addr.HostName))
			}
		}
		
		fmt.Printf("%d. Subject: %s\n", count, env.Subject)
		fmt.Printf("   To: %s\n", strings.Join(toAddrs, ", "))
		fmt.Printf("   Date: %s\n", dateStr)
		
		// Check if it has the Draft flag
		hasDraftFlag := false
		for _, flag := range msg.Flags {
			if flag == imap.DraftFlag {
				hasDraftFlag = true
				break
			}
		}
		if hasDraftFlag {
			fmt.Printf("   Status: ✓ Marked as Draft\n")
		}
		fmt.Println()
	}
	
	if err := <-done; err != nil {
		log.Fatal("Failed to fetch messages:", err)
	}
	
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Total drafts shown: %d (out of %d in folder)\n", count, mbox.Messages)
	
	// Check if our test draft is there
	testTime := time.Now().Format("15:04")
	fmt.Printf("\n🔍 Looking for recent test drafts (created around %s)...\n", testTime)
	fmt.Println("If you just created a draft, it should appear above.")
}