package oshandler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"lnb/internal/config"
)

type macHandler struct{}

func (h *macHandler) Handle(absPath, action string) error {
	linkName := filepath.Base(absPath)
	linkPath := "/usr/local/bin/" + linkName

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	switch action {
	case "install":
		// Check if file exists
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return fmt.Errorf("file '%s' does not exist", absPath)
		}

		// Check if file is executable
		if err := h.checkExecutable(absPath); err != nil {
			return fmt.Errorf("file '%s' is not executable: %v", absPath, err)
		}

		// Check if this binary is already installed
		if _, exists := cfg.GetEntry(linkName); exists {
			return fmt.Errorf("binary '%s' is already installed. Use 'lnb remove %s' first to reinstall", linkName, linkName)
		}

		// Check if the target path already exists
		if _, err := os.Stat(linkPath); err == nil {
			return fmt.Errorf("file already exists at %s. Please remove it manually or use 'lnb remove %s' if it was installed by LNB", linkPath, linkName)
		}

		err := os.Symlink(absPath, linkPath)
		if err != nil {
			return fmt.Errorf("failed to install: %v", err)
		}
		fmt.Printf("Installed: %s -> %s\n", linkPath, absPath)

		// Add to config
		cfg.AddEntry(linkName, absPath, linkPath)
		if err := cfg.Save(); err != nil {
			fmt.Printf("Warning: failed to update config: %v\n", err)
		}

	case "remove":
		err := os.Remove(linkPath)
		if err != nil {
			return fmt.Errorf("failed to remove: %v", err)
		}
		fmt.Printf("Removed: %s\n", linkPath)

		// Remove from config
		cfg.RemoveEntry(linkName)
		if err := cfg.Save(); err != nil {
			fmt.Printf("Warning: failed to update config: %v\n", err)
		}
	}
	return nil
}

func (h *macHandler) HandleAlias(aliasName, command, action string) error {
	scriptPath := "/usr/local/bin/" + aliasName

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	switch action {
	case "install":
		// Validate the command
		if err := h.validateCommand(command); err != nil {
			return fmt.Errorf("invalid command '%s': %v", command, err)
		}

		// Check if this alias is already installed
		if _, exists := cfg.GetEntry(aliasName); exists {
			return fmt.Errorf("alias '%s' is already installed. Use 'lnb unalias %s' first to reinstall", aliasName, aliasName)
		}

		// Check if the target path already exists
		if _, err := os.Stat(scriptPath); err == nil {
			return fmt.Errorf("file already exists at %s. Please remove it manually or use 'lnb unalias %s' if it was installed by LNB", scriptPath, aliasName)
		}

		// Convert relative paths to absolute paths in the command
		convertedCommand := h.convertRelativePaths(command)

		// Create the shell script content
		scriptContent := fmt.Sprintf(`#!/bin/bash
%s "$@"
`, convertedCommand)

		// Write the script file
		err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
		if err != nil {
			return fmt.Errorf("failed to create alias script: %v", err)
		}

		fmt.Printf("Created alias: %s -> %s\n", aliasName, convertedCommand)

		// Add to config with special marker for aliases
		cfg.AddEntry(aliasName, "alias:"+command, scriptPath)
		if err := cfg.Save(); err != nil {
			fmt.Printf("Warning: failed to update config: %v\n", err)
		}

	case "remove":
		err := os.Remove(scriptPath)
		if err != nil {
			return fmt.Errorf("failed to remove alias: %v", err)
		}
		fmt.Printf("Removed alias: %s\n", aliasName)

		// Remove from config
		cfg.RemoveEntry(aliasName)
		if err := cfg.Save(); err != nil {
			fmt.Printf("Warning: failed to update config: %v\n", err)
		}
	}
	return nil
}

// convertRelativePaths converts relative paths in command to absolute paths
func (h *macHandler) convertRelativePaths(command string) string {
	words := strings.Fields(command)
	for i, word := range words {
		// Check if this looks like a relative path (contains ./ or ../ or just a filename with extension)
		if strings.HasPrefix(word, "./") || strings.HasPrefix(word, "../") ||
			(strings.Contains(word, ".") && !strings.HasPrefix(word, "/") && !strings.Contains(word, "://")) {
			if absPath, err := filepath.Abs(word); err == nil {
				// Verify the file exists before converting
				if _, err := os.Stat(absPath); err == nil {
					words[i] = absPath
				}
			}
		}
	}
	return strings.Join(words, " ")
}

// checkExecutable verifies if a file is executable
func (h *macHandler) checkExecutable(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}

	// Check if file has execute permissions
	if fileInfo.Mode()&0111 == 0 {
		return fmt.Errorf("file does not have execute permissions")
	}

	return nil
}

// validateCommand checks if the command can be executed (basic validation)
func (h *macHandler) validateCommand(command string) error {
	words := strings.Fields(command)
	if len(words) == 0 {
		return fmt.Errorf("empty command")
	}

	// Get the first word (the actual command)
	cmdName := words[0]

	// If it's a relative or absolute path, check if it exists and is executable
	if strings.Contains(cmdName, "/") {
		if absPath, err := filepath.Abs(cmdName); err == nil {
			if _, err := os.Stat(absPath); err != nil {
				return fmt.Errorf("command '%s' not found", cmdName)
			}
			return h.checkExecutable(absPath)
		}
	} else {
		// For commands in PATH, try to find them using 'which'
		// This is a simple check - we don't want to be too strict
		// Just verify it's not obviously wrong
		if strings.ContainsAny(cmdName, "{}[]()<>|&;") {
			return fmt.Errorf("command '%s' contains invalid characters", cmdName)
		}
	}

	return nil
}
