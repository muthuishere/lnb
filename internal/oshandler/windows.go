package oshandler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type windowsHandler struct{}

func (h *windowsHandler) Handle(absPath, action string) error {
	binDir := filepath.Join(os.Getenv("USERPROFILE"), "bin")
	linkName := filepath.Base(absPath)
	linkNameWithoutExt := strings.TrimSuffix(linkName, filepath.Ext(linkName))
	cmdPath := filepath.Join(binDir, linkNameWithoutExt+".cmd")

	switch action {
	case "install":
		err := os.MkdirAll(binDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating bin dir: %v", err)
		}

		cmdContents := fmt.Sprintf(`@echo off
"%s" %%*
`, absPath)

		err = os.WriteFile(cmdPath, []byte(cmdContents), 0755)
		if err != nil {
			return fmt.Errorf("failed to write wrapper: %v", err)
		}
		fmt.Printf("Installed: %s\n", cmdPath)
		fmt.Println("⚠️ Make sure", binDir, "is in your PATH")

	case "remove":
		err := os.Remove(cmdPath)
		if err != nil {
			return fmt.Errorf("failed to remove: %v", err)
		}
		fmt.Printf("Removed: %s\n", cmdPath)
	}
	return nil
}
