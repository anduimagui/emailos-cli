package mailos

import (
	"fmt"
)

// DetectEmailProvider is a utility function to detect provider for a given email
func DetectEmailProvider(email string) {
	if email == "" {
		fmt.Println("Error: No email provided")
		return
	}

	fmt.Printf("Detecting provider for: %s\n", email)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	provider, detected, method := detectProviderFromEmail(email)
	
	if detected {
		fmt.Printf("✓ Detected provider: %s (via %s)\n", GetProviderName(provider), method)
		fmt.Printf("Provider key: %s\n", provider)
		
		// Show provider details
		if providerInfo, exists := Providers[provider]; exists {
			fmt.Printf("SMTP Server: %s:%d\n", providerInfo.SMTPHost, providerInfo.SMTPPort)
			fmt.Printf("IMAP Server: %s:%d\n", providerInfo.IMAPHost, providerInfo.IMAPPort)
			fmt.Printf("App Password URL: %s\n", providerInfo.AppPasswordURL)
		}
	} else {
		fmt.Printf("? Using default provider: %s (unable to detect from domain/MX records)\n", GetProviderName(provider))
		fmt.Printf("Provider key: %s\n", provider)
		fmt.Println("Note: This is a fallback provider that works well with custom domains")
	}
}