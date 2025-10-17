// +build !opentui

package mailos

import "fmt"

// InteractiveModeWithOpenTUI is a stub when OpenTUI is not available
func InteractiveModeWithOpenTUI() error {
	return fmt.Errorf("OpenTUI support not compiled in. To enable OpenTUI, build with: go build -tags opentui")
}