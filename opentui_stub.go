package mailos

import "fmt"

// InteractiveModeWithOpenTUI is a stub implementation for when OpenTUI is not available
func InteractiveModeWithOpenTUI() error {
	return fmt.Errorf("OpenTUI functionality is not available in this build")
}