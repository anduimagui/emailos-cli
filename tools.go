package mailos

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MethodInfo represents information about a Go method/function
type MethodInfo struct {
	Name       string
	File       string
	Line       int
	Signature  string
	IsExported bool
	Package    string
}

// ToolsCommand displays all available methods/functions in the codebase
func ToolsCommand() error {
	fmt.Println("📚 EmailOS (mailos) - Complete Method Reference")
	fmt.Println("=" + strings.Repeat("=", 60))
	fmt.Println()

	methods, err := extractAllMethods()
	if err != nil {
		return fmt.Errorf("failed to extract methods: %v", err)
	}

	// Group methods by category
	categories := groupMethodsByCategory(methods)

	// Display by category
	for _, category := range []string{
		"Core Email Operations",
		"Configuration & Setup", 
		"Interactive & UI",
		"AI & Suggestions",
		"File & Storage",
		"Authentication & Security",
		"Utilities & Helpers",
		"Testing & Development",
	} {
		if methodList, exists := categories[category]; exists {
			fmt.Printf("## %s\n", category)
			fmt.Println()
			for _, method := range methodList {
				fmt.Printf("  %s\n", formatMethodForLLM(method))
			}
			fmt.Println()
		}
	}

	// Show CLI commands available
	fmt.Println("## Available CLI Commands")
	fmt.Println()
	cliCommands := []string{
		"mailos read              # Read and filter emails",
		"mailos send              # Send emails with attachments",
		"mailos draft             # Create and manage email drafts",
		"mailos reply             # Reply to emails with threading",
		"mailos interactive       # Interactive email management mode",
		"mailos setup             # Configure email accounts",
		"mailos provider          # Setup AI providers",
		"mailos sync              # Sync emails with local storage",
		"mailos search            # Advanced email search",
		"mailos stats             # Generate email statistics",
		"mailos report            # Create email reports",
		"mailos template          # Manage email templates",
		"mailos tools             # Show this method reference",
		"mailos help              # Get command-specific help",
	}

	for _, cmd := range cliCommands {
		fmt.Printf("  %s\n", cmd)
	}
	fmt.Println()

	// Show key structs and types
	fmt.Println("## Key Data Types")
	fmt.Println()
	keyTypes := []string{
		"Email                    # Core email message structure",
		"EmailMessage             # Outgoing email with metadata",
		"DraftEmail               # Draft email with send options",
		"ReadOptions              # Configuration for reading emails",
		"DraftsOptions            # Configuration for draft operations",
		"SendOptions              # Configuration for sending emails",
		"Config                   # Application configuration",
		"AccountConfig            # Individual account settings",
		"QueryOptions             # Search and filter parameters",
		"EmailStats               # Statistical analysis results",
	}

	for _, typ := range keyTypes {
		fmt.Printf("  %s\n", typ)
	}
	fmt.Println()

	fmt.Printf("Total Methods Found: %d\n", len(methods))
	fmt.Println("\nFor detailed documentation on any command, use: mailos help <command>")

	return nil
}

// extractAllMethods scans Go files and extracts all function/method signatures
func extractAllMethods() ([]MethodInfo, error) {
	var methods []MethodInfo
	
	// Get current working directory (should be the Go module root)
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip certain directories
		if info.IsDir() {
			name := info.Name()
			if name == "vendor" || name == ".git" || name == "node_modules" || 
			   name == "dist" || name == "test_bubbletea" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process .go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files for main listing (but include test helpers)
		if strings.HasSuffix(path, "_test.go") && !strings.Contains(path, "test_") {
			return nil
		}

		fileMethods, err := extractMethodsFromFile(path)
		if err != nil {
			// Log error but continue processing other files
			return nil
		}

		methods = append(methods, fileMethods...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort methods by name for consistent output
	sort.Slice(methods, func(i, j int) bool {
		return methods[i].Name < methods[j].Name
	})

	return methods, nil
}

// extractMethodsFromFile parses a single Go file and extracts function signatures
func extractMethodsFromFile(filename string) ([]MethodInfo, error) {
	var methods []MethodInfo

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// Get relative path for cleaner display
	relPath, _ := filepath.Rel(".", filename)

	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Name != nil {
				pos := fset.Position(x.Pos())
				signature := extractFunctionSignature(x)
				
				method := MethodInfo{
					Name:       x.Name.Name,
					File:       relPath,
					Line:       pos.Line,
					Signature:  signature,
					IsExported: x.Name.IsExported(),
					Package:    node.Name.Name,
				}
				methods = append(methods, method)
			}
		}
		return true
	})

	return methods, nil
}

// extractFunctionSignature creates a readable function signature
func extractFunctionSignature(fn *ast.FuncDecl) string {
	var parts []string
	
	// Add receiver if it's a method
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		recv := fn.Recv.List[0]
		recvType := formatType(recv.Type)
		parts = append(parts, fmt.Sprintf("(%s)", recvType))
	}

	// Add function name
	parts = append(parts, fn.Name.Name)

	// Add parameters
	params := formatFieldList(fn.Type.Params)
	parts = append(parts, fmt.Sprintf("(%s)", params))

	// Add return values
	if fn.Type.Results != nil {
		results := formatFieldList(fn.Type.Results)
		if results != "" {
			if len(fn.Type.Results.List) > 1 {
				parts = append(parts, fmt.Sprintf("(%s)", results))
			} else {
				parts = append(parts, results)
			}
		}
	}

	return strings.Join(parts, " ")
}

// formatFieldList formats function parameters or return values
func formatFieldList(fields *ast.FieldList) string {
	if fields == nil || len(fields.List) == 0 {
		return ""
	}

	var parts []string
	for _, field := range fields.List {
		fieldType := formatType(field.Type)
		if len(field.Names) > 0 {
			for _, name := range field.Names {
				parts = append(parts, fmt.Sprintf("%s %s", name.Name, fieldType))
			}
		} else {
			parts = append(parts, fieldType)
		}
	}

	return strings.Join(parts, ", ")
}

// formatType formats AST type expressions
func formatType(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.StarExpr:
		return "*" + formatType(x.X)
	case *ast.ArrayType:
		return "[]" + formatType(x.Elt)
	case *ast.SelectorExpr:
		return formatType(x.X) + "." + x.Sel.Name
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", formatType(x.Key), formatType(x.Value))
	case *ast.ChanType:
		return "chan " + formatType(x.Value)
	case *ast.FuncType:
		return "func"
	default:
		return "unknown"
	}
}

// groupMethodsByCategory organizes methods into logical categories
func groupMethodsByCategory(methods []MethodInfo) map[string][]MethodInfo {
	categories := make(map[string][]MethodInfo)

	for _, method := range methods {
		category := categorizeMethod(method)
		categories[category] = append(categories[category], method)
	}

	// Sort methods within each category
	for category := range categories {
		sort.Slice(categories[category], func(i, j int) bool {
			return categories[category][i].Name < categories[category][j].Name
		})
	}

	return categories
}

// categorizeMethod determines which category a method belongs to
func categorizeMethod(method MethodInfo) string {
	name := strings.ToLower(method.Name)
	file := strings.ToLower(method.File)

	// Core email operations
	if strings.Contains(name, "read") || strings.Contains(name, "send") || 
	   strings.Contains(name, "draft") || strings.Contains(name, "reply") ||
	   strings.Contains(name, "email") || strings.Contains(file, "read.go") ||
	   strings.Contains(file, "send.go") || strings.Contains(file, "drafts.go") ||
	   strings.Contains(file, "reply.go") {
		return "Core Email Operations"
	}

	// Configuration and setup
	if strings.Contains(name, "config") || strings.Contains(name, "setup") ||
	   strings.Contains(name, "provider") || strings.Contains(file, "setup.go") ||
	   strings.Contains(file, "config.go") || strings.Contains(file, "frontend.go") {
		return "Configuration & Setup"
	}

	// Interactive and UI
	if strings.Contains(name, "interactive") || strings.Contains(name, "menu") ||
	   strings.Contains(name, "prompt") || strings.Contains(file, "interactive") ||
	   strings.Contains(file, "input_") || strings.Contains(name, "ui") {
		return "Interactive & UI"
	}

	// AI and suggestions
	if strings.Contains(name, "ai") || strings.Contains(name, "suggestion") ||
	   strings.Contains(file, "ai_") || strings.Contains(file, "suggestion") {
		return "AI & Suggestions"
	}

	// File and storage operations
	if strings.Contains(name, "file") || strings.Contains(name, "save") ||
	   strings.Contains(name, "load") || strings.Contains(name, "sync") ||
	   strings.Contains(file, "save") || strings.Contains(file, "sync.go") ||
	   strings.Contains(file, "file_") {
		return "File & Storage"
	}

	// Authentication and security
	if strings.Contains(name, "auth") || strings.Contains(name, "validate") ||
	   strings.Contains(name, "license") || strings.Contains(file, "auth.go") ||
	   strings.Contains(file, "middleware.go") {
		return "Authentication & Security"
	}

	// Testing and development
	if strings.Contains(name, "test") || strings.Contains(file, "test") ||
	   strings.Contains(name, "debug") || method.Name == "main" {
		return "Testing & Development"
	}

	// Default to utilities
	return "Utilities & Helpers"
}

// formatMethodForLLM formats method info in a way that's useful for LLMs
func formatMethodForLLM(method MethodInfo) string {
	visibility := ""
	if method.IsExported {
		visibility = "[PUBLIC]  "
	} else {
		visibility = "[private] "
	}

	location := fmt.Sprintf("%s:%d", method.File, method.Line)
	
	return fmt.Sprintf("%s%-25s  %s", visibility, method.Name, location)
}