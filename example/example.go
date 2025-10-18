package main

import (
	"fmt"
	"log"
	"time"

	"github.com/anduimagui/emailos"
)

func main() {
	// Check if configuration exists
	if !mailos.ConfigExists() {
		fmt.Println("No configuration found. Please run 'mailos setup' first.")
		return
	}

	// Create a new client
	client, err := mailos.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Example 1: Send an email
	fmt.Println("Sending test email...")
	err = client.SendEmail(
		[]string{"test@example.com"},
		"Test Email from EmailOS",
		`# Hello from EmailOS!

This is a **test email** sent using the EmailOS client.

## Features
- Markdown formatting
- Multiple providers
- Easy to use

Best regards,
EmailOS Team`,
		nil, // CC
		nil, // BCC
	)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
	} else {
		fmt.Println("✓ Email sent successfully!")
	}

	// Example 2: Read recent emails
	fmt.Println("\nReading recent emails...")
	emails, err := client.ReadEmails(mailos.ReadOptions{
		Limit: 5,
		Since: time.Now().AddDate(0, 0, -7), // Last 7 days
	})
	if err != nil {
		log.Printf("Failed to read emails: %v", err)
	} else {
		fmt.Printf("Found %d emails:\n", len(emails))
		for i, email := range emails {
			fmt.Printf("%d. From: %s - Subject: %s\n", i+1, email.From, email.Subject)
		}
	}

	// Example 3: Show configuration info
	config := client.GetConfig()
	fmt.Printf("\nCurrent configuration:\n")
	fmt.Printf("Provider: %s\n", client.GetProviderInfo())
	fmt.Printf("Email: %s\n", config.Email)
}