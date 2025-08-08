package mailos

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	// UI version to download
	uiVersion = "0.1.0"
	// URL to download the UI package from
	uiPackageURL = "https://registry.npmjs.org/@mailos/ui/-/ui-" + uiVersion + ".tgz"
)

// LaunchReactInkUI launches the React Ink interactive interface
func LaunchReactInkUI() error {
	// Find or install the UI
	uiPath, err := ensureUIInstalled()
	if err != nil {
		// If UI installation fails, fall back to classic UI
		fmt.Printf("React Ink UI not available: %v\n", err)
		fmt.Println("Falling back to classic interface...")
		return showEnhancedInteractiveMenu(false)
	}
	
	distPath := filepath.Join(uiPath, "dist", "index.js")
	
	// Check if built, if not build it
	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		fmt.Println("Building React Ink UI...")
		if err := buildReactInkUI(uiPath); err != nil {
			fmt.Printf("Failed to build UI: %v\n", err)
			fmt.Println("Falling back to classic interface...")
			return showEnhancedInteractiveMenu(false)
		}
	}

	// Start the API server in the background
	go startAPIServer()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Launch the React Ink UI
	cmd := exec.Command("node", distPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	// Run the command and check exit code
	err = cmd.Run()
	
	// Check if the UI exited with code 42 (AI query)
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() == 42 {
			// Read the query from temp file
			tempFile := filepath.Join(os.TempDir(), "mailos_query.txt")
			queryBytes, err := os.ReadFile(tempFile)
			if err == nil {
				query := strings.TrimSpace(string(queryBytes))
				os.Remove(tempFile) // Clean up temp file
				
				// Get the AI provider command
				config, _ := LoadConfig()
				aiCommand := config.DefaultAICLI
				
				// Map to actual command names
				switch aiCommand {
				case "claude-code-yolo":
					aiCommand = "claude-code --yolo"
				case "claude-code":
					aiCommand = "claude-code"
				case "openai-codex":
					aiCommand = "openai"
				case "gemini-cli":
					aiCommand = "gemini"
				}
				
				if aiCommand != "" && aiCommand != "none" {
					// Execute the AI provider directly
					fmt.Printf("\nExecuting: %s \"%s\"\n\n", aiCommand, query)
					
					// Build the command
					var cmd *exec.Cmd
					if strings.Contains(aiCommand, " ") {
						parts := strings.Split(aiCommand, " ")
						args := append(parts[1:], query)
						cmd = exec.Command(parts[0], args...)
					} else {
						cmd = exec.Command(aiCommand, query)
					}
					
					cmd.Stdin = os.Stdin
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					
					// Run the AI command
					return cmd.Run()
				} else {
					fmt.Println("No AI provider configured. Use 'mailos provider' to set one up.")
				}
				return nil
			}
		}
	}
	
	return err
}

// ensureUIInstalled ensures the UI is installed and returns its path
func ensureUIInstalled() (string, error) {
	// Check if UI is already installed
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	
	uiPath := filepath.Join(homeDir, ".email", "ui")
	
	// Check if package.json exists
	if _, err := os.Stat(filepath.Join(uiPath, "package.json")); err == nil {
		return uiPath, nil
	}
	
	// UI not installed, install it now
	fmt.Println("Installing React Ink UI for first-time use...")
	if err := installUI(uiPath); err != nil {
		return "", err
	}
	
	return uiPath, nil
}

// installUI downloads and installs the UI package
func installUI(targetDir string) error {
	// Create directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create UI directory: %w", err)
	}
	
	// For production, we'll embed the UI files or download from CDN
	// For now, create a minimal package.json and the necessary files
	if err := createMinimalUI(targetDir); err != nil {
		return fmt.Errorf("failed to create UI files: %w", err)
	}
	
	return nil
}

// createMinimalUI creates a minimal UI installation
func createMinimalUI(targetDir string) error {
	// Create dist directory
	distDir := filepath.Join(targetDir, "dist")
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return err
	}
	
	// Create a simple Node.js CLI without React dependencies
	indexJS := `#!/usr/bin/env node

const readline = require('readline');
const { exec } = require('child_process');
const fs = require('fs');
const path = require('path');
const os = require('os');

// ANSI color codes
const colors = {
  reset: '\x1b[0m',
  cyan: '\x1b[36m',
  yellow: '\x1b[33m',
  gray: '\x1b[90m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  blue: '\x1b[34m',
  bold: '\x1b[1m'
};

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
  prompt: colors.cyan + '> ' + colors.reset
});

// Load config to get email and AI provider
let config = {};
try {
  const configPath = path.join(os.homedir(), '.email', 'config.json');
  if (fs.existsSync(configPath)) {
    config = JSON.parse(fs.readFileSync(configPath, 'utf8'));
  }
} catch (err) {
  // Config not found, will use defaults
}

// Clear screen and show header
console.clear();
console.log(colors.cyan + colors.bold + '📧 MailOS - Interactive Mode' + colors.reset);
console.log(colors.gray + '━'.repeat(80) + colors.reset);
console.log(colors.green + '📬 Account: ' + colors.reset + (config.email || 'Not configured'));
console.log(colors.blue + '🤖 AI Provider: ' + colors.reset + (config.default_ai_cli || 'Not configured'));
console.log(colors.gray + '━'.repeat(80) + colors.reset);
console.log('');
console.log(colors.yellow + '💡 Enter a query for AI, or use:' + colors.reset);
console.log(colors.gray + '   /  - Show available commands' + colors.reset);
console.log(colors.gray + '   @  - Select email template' + colors.reset);
console.log(colors.gray + '   q  - Quit' + colors.reset);
console.log('');

// Show commands
function showCommands() {
  console.log('');
  console.log(colors.cyan + '📋 Available Commands:' + colors.reset);
  console.log(colors.gray + '─'.repeat(50) + colors.reset);
  console.log(colors.yellow + '  /read' + colors.reset + '        - Read recent emails');
  console.log(colors.yellow + '  /send' + colors.reset + '        - Send an email');
  console.log(colors.yellow + '  /search' + colors.reset + '      - Search emails');
  console.log(colors.yellow + '  /stats' + colors.reset + '       - Show email statistics');
  console.log(colors.yellow + '  /report' + colors.reset + '      - Generate email report');
  console.log(colors.yellow + '  /delete' + colors.reset + '      - Delete emails');
  console.log(colors.yellow + '  /unsubscribe' + colors.reset + ' - Find unsubscribe links');
  console.log(colors.yellow + '  /template' + colors.reset + '    - Manage templates');
  console.log(colors.yellow + '  /config' + colors.reset + '      - Configure settings');
  console.log(colors.yellow + '  /provider' + colors.reset + '    - Set AI provider');
  console.log(colors.yellow + '  /help' + colors.reset + '        - Show this help');
  console.log(colors.yellow + '  /exit' + colors.reset + '        - Exit program');
  console.log(colors.gray + '─'.repeat(50) + colors.reset);
  console.log('');
}

// Show templates
function showTemplates() {
  console.log('');
  console.log(colors.cyan + '📝 Email Templates:' + colors.reset);
  console.log(colors.gray + '─'.repeat(50) + colors.reset);
  console.log(colors.yellow + '  @meeting' + colors.reset + '   - Schedule a meeting');
  console.log(colors.yellow + '  @followup' + colors.reset + '  - Follow up email');
  console.log(colors.yellow + '  @thank' + colors.reset + '     - Thank you email');
  console.log(colors.yellow + '  @intro' + colors.reset + '     - Introduction email');
  console.log(colors.yellow + '  @request' + colors.reset + '   - Request information');
  console.log(colors.yellow + '  @reminder' + colors.reset + '  - Send a reminder');
  console.log(colors.yellow + '  @apologize' + colors.reset + ' - Apology email');
  console.log(colors.yellow + '  @decline' + colors.reset + '   - Decline politely');
  console.log(colors.gray + '─'.repeat(50) + colors.reset);
  console.log(colors.gray + 'Example: @meeting John tomorrow at 3pm' + colors.reset);
  console.log('');
}

// Execute command
function executeCommand(cmd, args) {
  const fullCmd = 'mailos ' + cmd + ' ' + args;
  console.log(colors.gray + 'Executing: ' + fullCmd + colors.reset);
  
  exec(fullCmd, (error, stdout, stderr) => {
    if (error) {
      console.error(colors.red + 'Error: ' + error.message + colors.reset);
    } else {
      console.log(stdout);
    }
    rl.prompt();
  });
}

// Send query to AI
function sendToAI(query) {
  if (!config.default_ai_cli || config.default_ai_cli === 'none') {
    console.log(colors.yellow + '⚠️  No AI provider configured.' + colors.reset);
    console.log(colors.gray + 'Use /provider to set up an AI provider.' + colors.reset);
    rl.prompt();
    return;
  }
  
  console.log(colors.gray + 'Launching AI provider with query: ' + query + colors.reset);
  
  // Close readline interface
  rl.close();
  
  // Exit and let the parent process handle the AI query
  // Write query to temp file for parent to pick up
  const tempFile = path.join(os.tmpdir(), 'mailos_query.txt');
  fs.writeFileSync(tempFile, query);
  
  // Exit with special code to indicate AI query
  process.exit(42);
}

// Process template
function processTemplate(input) {
  const parts = input.substring(1).split(' ');
  const template = parts[0];
  const args = parts.slice(1).join(' ');
  
  const templates = {
    'meeting': 'compose an email to schedule a meeting about ',
    'followup': 'write a follow-up email regarding ',
    'thank': 'write a thank you email for ',
    'intro': 'write an introduction email to ',
    'request': 'write an email requesting ',
    'reminder': 'write a reminder email about ',
    'apologize': 'write an apology email for ',
    'decline': 'write a polite email declining '
  };
  
  if (templates[template]) {
    const query = templates[template] + args;
    console.log(colors.green + '📝 Using template: ' + template + colors.reset);
    sendToAI(query);
  } else {
    console.log(colors.red + 'Unknown template: @' + template + colors.reset);
    showTemplates();
    rl.prompt();
  }
}

// Main prompt loop
rl.prompt();

rl.on('line', (input) => {
  input = input.trim();
  
  if (input.toLowerCase() === 'q' || input.toLowerCase() === 'quit') {
    console.log(colors.cyan + '👋 Goodbye!' + colors.reset);
    process.exit(0);
  }
  
  // Handle slash commands
  if (input.startsWith('/')) {
    // Special handling for just "/" - show commands
    if (input === '/') {
      showCommands();
      rl.prompt();
      return;
    }
    
    const parts = input.substring(1).split(' ');
    const cmd = parts[0].toLowerCase();
    const args = parts.slice(1).join(' ');
    
    switch(cmd) {
      case 'help':
        showCommands();
        rl.prompt();
        break;
      case 'exit':
        console.log(colors.cyan + '👋 Goodbye!' + colors.reset);
        process.exit(0);
        break;
      case 'read':
      case 'send':
      case 'search':
      case 'stats':
      case 'report':
      case 'delete':
      case 'unsubscribe':
      case 'template':
      case 'config':
      case 'provider':
        executeCommand(cmd, args);
        break;
      default:
        console.log(colors.red + 'Unknown command: /' + cmd + colors.reset);
        showCommands();
        rl.prompt();
    }
  }
  // Handle template shortcuts
  else if (input.startsWith('@')) {
    if (input === '@') {
      showTemplates();
      rl.prompt();
    } else {
      processTemplate(input);
    }
  }
  // Send to AI as query
  else if (input.length > 0) {
    sendToAI(input);
  } else {
    rl.prompt();
  }
});

rl.on('close', () => {
  console.log(colors.cyan + '\n👋 Goodbye!' + colors.reset);
  process.exit(0);
});
`
	
	// Also create a package.json for compatibility
	packageJSON := map[string]interface{}{
		"name":    "mailos-ui",
		"version": uiVersion,
		"main":    "dist/index.js",
	}
	
	packageData, err := json.MarshalIndent(packageJSON, "", "  ")
	if err != nil {
		return err
	}
	
	if err := os.WriteFile(filepath.Join(targetDir, "package.json"), packageData, 0644); err != nil {
		return err
	}
	
	return os.WriteFile(filepath.Join(distDir, "index.js"), []byte(indexJS), 0755)
}

// buildReactInkUI builds the React Ink application
func buildReactInkUI(uiPath string) error {
	// Check if npm is installed
	if _, err := exec.LookPath("npm"); err != nil {
		// Try to use the pre-built version
		return nil
	}

	// Install dependencies
	fmt.Println("Installing dependencies...")
	installCmd := exec.Command("npm", "install", "--production")
	installCmd.Dir = uiPath
	if output, err := installCmd.CombinedOutput(); err != nil {
		// If npm install fails, we can still try to use the pre-built version
		fmt.Printf("Note: npm install had issues: %s\n", output)
		return nil
	}

	return nil
}

// EmailAPI represents the API response for emails
type EmailAPI struct {
	ID             string    `json:"id"`
	From           string    `json:"from"`
	To             string    `json:"to"`
	Subject        string    `json:"subject"`
	Date           time.Time `json:"date"`
	Body           string    `json:"body"`
	IsRead         bool      `json:"isRead"`
	HasAttachments bool      `json:"hasAttachments"`
	Tags           []string  `json:"tags"`
}

// startAPIServer starts a local HTTP server for the React app to communicate with
func startAPIServer() {
	mux := http.NewServeMux()

	// GET /api/emails - Fetch all emails
	mux.HandleFunc("/api/emails", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		config, err := LoadConfig()
		if err != nil {
			http.Error(w, "Failed to load config", http.StatusInternalServerError)
			return
		}

		// Create client
		client, err := NewClient()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create client: %v", err), http.StatusInternalServerError)
			return
		}

		// Read emails
		emails, err := client.ReadEmails(ReadOptions{
			Limit:      100,
			UnreadOnly: false,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read emails: %v", err), http.StatusInternalServerError)
			return
		}

		// Convert to API format
		apiEmails := make([]EmailAPI, len(emails))
		for i, email := range emails {
			apiEmails[i] = EmailAPI{
				ID:             fmt.Sprintf("%d", email.ID),
				From:           email.From,
				To:             config.Email,
				Subject:        email.Subject,
				Date:           email.Date,
				Body:           email.Body,
				IsRead:         false, // TODO: Implement read flag tracking
				HasAttachments: len(email.Attachments) > 0,
				Tags:           []string{}, // Initialize with empty tags
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(apiEmails)
	})

	// POST /api/emails/:id/read - Mark email as read
	mux.HandleFunc("/api/emails/read", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request struct {
			IDs []string `json:"ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Convert string IDs to uint32
		var ids []uint32
		for _, idStr := range request.IDs {
			var id uint32
			fmt.Sscanf(idStr, "%d", &id)
			ids = append(ids, id)
		}

		client, err := NewClient()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create client: %v", err), http.StatusInternalServerError)
			return
		}

		if err := client.MarkEmailsAsRead(ids); err != nil {
			http.Error(w, fmt.Sprintf("Failed to mark as read: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	})

	// DELETE /api/emails - Delete emails
	mux.HandleFunc("/api/emails/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request struct {
			IDs []string `json:"ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Convert string IDs to uint32
		var ids []uint32
		for _, idStr := range request.IDs {
			var id uint32
			fmt.Sscanf(idStr, "%d", &id)
			ids = append(ids, id)
		}

		client, err := NewClient()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create client: %v", err), http.StatusInternalServerError)
			return
		}

		if err := client.DeleteEmails(ids); err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete emails: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	})

	// POST /api/emails/send - Send a new email
	mux.HandleFunc("/api/emails/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request struct {
			To      string `json:"to"`
			Subject string `json:"subject"`
			Body    string `json:"body"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Create client
		client, err := NewClient()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create client: %v", err), http.StatusInternalServerError)
			return
		}

		if err := client.SendEmail([]string{request.To}, request.Subject, request.Body, nil, nil); err != nil {
			http.Error(w, fmt.Sprintf("Failed to send email: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	})

	// Start server
	port := "8080"
	if runtime.GOOS == "windows" {
		port = "8081" // Use different port on Windows to avoid conflicts
	}

	// Start server silently
	http.ListenAndServe(":"+port, mux)
}

// InteractiveModeWithReactInk launches the new React Ink based interactive mode
func InteractiveModeWithReactInk() error {
	return LaunchReactInkUI()
}