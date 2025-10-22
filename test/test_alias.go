package main

import (
	"fmt"
	"strings"
	"testing"
)

var aliasArray = []string{
	"support-test@raggle.co",
	"hello@raggle.co",
	"support@email-os.com",
	"test@email-os.com",
	"any-prefix@speedrunners.dev",
}

// TestWildcardDomainMatching tests wildcard domain alias functionality
func TestWildcardDomainMatching(t *testing.T) {
	testCases := []struct {
		email           string
		configuredEmail string
		domain          string
		shouldMatch     bool
		description     string
	}{
		{
			email:           "support-test@raggle.co",
			configuredEmail: "andrew@raggle.co",
			domain:          "raggle.co",
			shouldMatch:     true,
			description:     "Should match wildcard for raggle.co domain",
		},
		{
			email:           "hello@raggle.co",
			configuredEmail: "andrew@raggle.co",
			domain:          "raggle.co",
			shouldMatch:     true,
			description:     "Should match any prefix for raggle.co domain",
		},
		{
			email:           "test@email-os.com",
			configuredEmail: "andrew@email-os.com",
			domain:          "email-os.com",
			shouldMatch:     true,
			description:     "Should match wildcard for email-os.com domain",
		},
		{
			email:           "support@email-os.com",
			configuredEmail: "andrew@email-os.com",
			domain:          "email-os.com",
			shouldMatch:     true,
			description:     "Should match support@ for email-os.com domain",
		},
		{
			email:           "test@unknown.com",
			configuredEmail: "andrew@raggle.co",
			domain:          "raggle.co",
			shouldMatch:     false,
			description:     "Should not match different domain",
		},
		{
			email:           "andrew@speedrunners.dev",
			configuredEmail: "andrew@speedrunners.dev",
			domain:          "speedrunners.dev",
			shouldMatch:     true,
			description:     "Should match exact configured email",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := matchesWildcardDomain(tc.email, tc.configuredEmail, tc.domain)
			if result != tc.shouldMatch {
				t.Errorf("Expected %v for %s matching %s (domain: %s), got %v",
					tc.shouldMatch, tc.email, tc.configuredEmail, tc.domain, result)
			}
		})
	}
}

// matchesWildcardDomain checks if an email matches a wildcard domain pattern
func matchesWildcardDomain(email, configuredEmail, domain string) bool {
	emailParts := strings.Split(email, "@")
	configParts := strings.Split(configuredEmail, "@")

	if len(emailParts) != 2 || len(configParts) != 2 {
		return false
	}

	emailDomain := emailParts[1]
	configDomain := configParts[1]

	// Exact match
	if email == configuredEmail {
		return true
	}

	// Domain wildcard match
	if emailDomain == configDomain && emailDomain == domain {
		return true
	}

	return false
}

// TestDomainWildcards tests the specific domains mentioned in the configuration
func TestDomainWildcards(t *testing.T) {
	configuredAccounts := map[string]string{
		"raggle.co":        "andrew@raggle.co",
		"email-os.com":     "andrew@email-os.com",
		"speedrunners.dev": "andrew@speedrunners.dev",
	}

	testEmails := []string{
		"support-test@raggle.co",
		"hello@raggle.co",
		"support@email-os.com",
		"test@email-os.com",
		"any-prefix@speedrunners.dev",
	}

	for _, email := range testEmails {
		parts := strings.Split(email, "@")
		if len(parts) != 2 {
			continue
		}

		domain := parts[1]
		if configuredAccount, exists := configuredAccounts[domain]; exists {
			t.Logf("✅ %s should work with configured account %s", email, configuredAccount)

			if !matchesWildcardDomain(email, configuredAccount, domain) {
				t.Errorf("❌ %s should match wildcard for domain %s", email, domain)
			}
		} else {
			t.Logf("❌ %s has no configured account for domain %s", email, domain)
		}
	}
}

// TestAliasArrayIteration tests all aliases with different email scenarios
func TestAliasArrayIteration(t *testing.T) {
	fmt.Println("=== Testing Alias Array Iteration ===")

	// Configured accounts for each domain
	configuredAccounts := map[string]string{
		"raggle.co":        "andrew@raggle.co",
		"email-os.com":     "andrew@email-os.com",
		"speedrunners.dev": "andrew@speedrunners.dev",
	}

	for _, alias := range aliasArray {
		t.Run(fmt.Sprintf("Alias_%s", alias), func(t *testing.T) {
			parts := strings.Split(alias, "@")
			if len(parts) != 2 {
				t.Errorf("Invalid alias format: %s", alias)
				return
			}

			originalPrefix := parts[0]
			domain := parts[1]

			// Check if we have a configured account for this domain
			configuredAccount, exists := configuredAccounts[domain]
			if !exists {
				t.Logf("⚠️  No configured account for domain %s", domain)
				return
			}

			t.Logf("Testing alias: %s with domain: %s", alias, domain)

			// Test the original alias
			matches := matchesWildcardDomain(alias, configuredAccount, domain)
			if !matches {
				t.Errorf("❌ Original alias %s should match configured account %s", alias, configuredAccount)
			} else {
				t.Logf("✅ Original alias %s matches %s", alias, configuredAccount)
			}

			// Test different prefixes for the same domain
			testPrefixes := []string{
				"admin",
				"info",
				"contact",
				"help",
				"sales",
				"noreply",
			}

			for _, prefix := range testPrefixes {
				testEmail := fmt.Sprintf("%s@%s", prefix, domain)
				matches := matchesWildcardDomain(testEmail, configuredAccount, domain)
				if !matches {
					t.Errorf("❌ Test email %s should match configured account %s for domain %s",
						testEmail, configuredAccount, domain)
				} else {
					t.Logf("✅ Test email %s matches %s", testEmail, configuredAccount)
				}
			}

			// Test that emails from different domains don't match
			wrongDomainEmail := fmt.Sprintf("%s@wrongdomain.com", originalPrefix)
			matches = matchesWildcardDomain(wrongDomainEmail, configuredAccount, domain)
			if matches {
				t.Errorf("❌ Email %s should NOT match configured account %s (wrong domain)",
					wrongDomainEmail, configuredAccount)
			} else {
				t.Logf("✅ Correctly rejected %s (wrong domain)", wrongDomainEmail)
			}
		})
	}
}

func TestSpecificScenarios(t *testing.T) {
	fmt.Println("=== Testing Specific Wildcard Scenarios ===")

	scenarios := map[string]struct {
		email      string
		configured string
		domain     string
	}{
		"Raggle support test": {
			email:      "support-test@raggle.co",
			configured: "andrew@raggle.co",
			domain:     "raggle.co",
		},
		"EmailOS support": {
			email:      "support@email-os.com",
			configured: "andrew@email-os.com",
			domain:     "email-os.com",
		},
		"Speedrunners dev": {
			email:      "dev@speedrunners.dev",
			configured: "andrew@speedrunners.dev",
			domain:     "speedrunners.dev",
		},
	}

	for name, scenario := range scenarios {
		t.Run(name, func(t *testing.T) {
			matches := matchesWildcardDomain(scenario.email, scenario.configured, scenario.domain)
			if !matches {
				t.Errorf("Scenario '%s' failed: %s should match %s for domain %s",
					name, scenario.email, scenario.configured, scenario.domain)
			} else {
				t.Logf("✅ Scenario '%s' passed: %s matches %s for domain %s",
					name, scenario.email, scenario.configured, scenario.domain)
			}
		})
	}
}
