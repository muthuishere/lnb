package oshandler

import (
	"fmt"
	"os"
	"path/filepath"
)

type linuxHandler struct{}

func (h *linuxHandler) Handle(absPath, action string) error {
	linkName := filepath.Base(absPath)
	linkPath := "/usr/local/bin/" + linkName

	switch action {
	case "install":
		err := os.Symlink(absPath, linkPath)
		if err != nil {
			return fmt.Errorf("failed to install: %v", err)
		}
		fmt.Printf("Installed: %s -> %s\n", linkPath, absPath)

	case "remove":
		err := os.Remove(linkPath)
		if err != nil {
			return fmt.Errorf("failed to remove: %v", err)
		}
		fmt.Printf("Removed: %s\n", linkPath)
	}
	return nil
}
