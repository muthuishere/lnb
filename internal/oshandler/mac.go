package oshandler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"lnb/internal/config"
)

// parseShellArgs parses a command string into arguments while respecting quotes
func parseShellArgs(command string) []string {
	var args []string
	var current strings.Builder
	var inQuotes bool
	var quoteChar rune

	for _, char := range command {
		switch char {
		case '"', '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = char
				current.WriteRune(char) // Keep the quote in the argument
			} else if char == quoteChar {
				inQuotes = false
				current.WriteRune(char) // Keep the closing quote
				quoteChar = 0
			} else {
				current.WriteRune(char)
			}
		case ' ', '\t':
			if inQuotes {
				current.WriteRune(char)
			} else {
				if current.Len() > 0 {
					args = append(args, current.String())
					current.Reset()
				}
			}
		default:
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

// reconstructCommand rebuilds a command string from parsed arguments
func reconstructCommand(args []string) string {
	return strings.Join(args, " ")
}

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
		if entry, exists := cfg.GetEntry(linkName); exists {
			// Verify the target file actually exists
			if _, err := os.Stat(entry.TargetPath); err == nil {
				return fmt.Errorf("binary '%s' is already installed. Use 'lnb remove %s' first to reinstall", linkName, linkName)
			} else {
				// Config says it's installed but file doesn't exist - clean up the config
				fmt.Printf("Warning: Config shows '%s' as installed but target file '%s' doesn't exist. Cleaning up config entry.\n", linkName, entry.TargetPath)
				cfg.RemoveEntry(linkName)
				if err := cfg.Save(); err != nil {
					fmt.Printf("Warning: failed to clean up config: %v\n", err)
				}
			}
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
		// Check if this binary was installed by LNB
		entry, exists := cfg.GetEntry(linkName)
		if !exists {
			return fmt.Errorf("binary '%s' was not installed by LNB", linkName)
		}

		// Verify the target path matches what we expect
		if entry.TargetPath != linkPath {
			return fmt.Errorf("binary '%s' target path mismatch: expected %s, found %s", linkName, linkPath, entry.TargetPath)
		}

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
		if entry, exists := cfg.GetEntry(aliasName); exists {
			// Verify the target file actually exists
			if _, err := os.Stat(entry.TargetPath); err == nil {
				return fmt.Errorf("alias '%s' is already installed. Use 'lnb unalias %s' first to reinstall", aliasName, aliasName)
			} else {
				// Config says it's installed but file doesn't exist - clean up the config
				fmt.Printf("Warning: Config shows '%s' as installed but target file '%s' doesn't exist. Cleaning up config entry.\n", aliasName, entry.TargetPath)
				cfg.RemoveEntry(aliasName)
				if err := cfg.Save(); err != nil {
					fmt.Printf("Warning: failed to clean up config: %v\n", err)
				}
			}
		}

		// Check if the target path already exists
		if _, err := os.Stat(scriptPath); err == nil {
			return fmt.Errorf("file already exists at %s. Please remove it manually or use 'lnb unalias %s' if it was installed by LNB", scriptPath, aliasName)
		}

		// Convert relative paths to absolute paths in the command
		convertedCommand := h.convertRelativePaths(command)

		// Process .app bundles to use "open -a" automatically
		processedCommand := h.processAppBundle(convertedCommand)

		// Create the shell script content
		scriptContent := fmt.Sprintf(`#!/bin/bash
%s "$@"
`, processedCommand)

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
		// Check if this alias was installed by LNB
		entry, exists := cfg.GetEntry(aliasName)
		if !exists {
			return fmt.Errorf("alias '%s' was not installed by LNB", aliasName)
		}

		// Verify the target path matches what we expect
		if entry.TargetPath != scriptPath {
			return fmt.Errorf("alias '%s' target path mismatch: expected %s, found %s", aliasName, scriptPath, entry.TargetPath)
		}

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
	// Since quotes are now handled at the top level, we can work with the command as-is
	// If the command is quoted, preserve the quotes but convert the path inside

	trimmed := strings.TrimSpace(command)
	if (strings.HasPrefix(trimmed, `"`) && strings.HasSuffix(trimmed, `"`)) ||
		(strings.HasPrefix(trimmed, `'`) && strings.HasSuffix(trimmed, `'`)) {
		// Extract the quoted content
		quoteChar := string(trimmed[0])
		inner := trimmed[1 : len(trimmed)-1]

		// Convert relative paths in the quoted content
		converted := h.convertPathIfRelative(inner)

		// Return with quotes preserved
		return quoteChar + converted + quoteChar
	}

	// For unquoted commands, convert the first word if it's a relative path
	parts := strings.Fields(trimmed)
	if len(parts) > 0 {
		parts[0] = h.convertPathIfRelative(parts[0])
	}

	return strings.Join(parts, " ")
}

// convertPathIfRelative converts a single path if it's relative
func (h *macHandler) convertPathIfRelative(path string) string {
	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") ||
		(strings.Contains(path, ".") && !strings.HasPrefix(path, "/") && !strings.Contains(path, "://")) {
		if absPath, err := filepath.Abs(path); err == nil {
			// Verify the file exists before converting
			if _, err := os.Stat(absPath); err == nil {
				return absPath
			}
		}
	}
	return path
}

// processAppBundle automatically wraps .app bundles with "open -a"
func (h *macHandler) processAppBundle(command string) string {
	trimmed := strings.TrimSpace(command)

	// Handle quoted commands
	if (strings.HasPrefix(trimmed, `"`) && strings.HasSuffix(trimmed, `"`)) ||
		(strings.HasPrefix(trimmed, `'`) && strings.HasSuffix(trimmed, `'`)) {
		// Extract the quoted content
		quoteChar := string(trimmed[0])
		inner := trimmed[1 : len(trimmed)-1]

		// Check if the quoted path ends with .app
		if strings.HasSuffix(inner, ".app") {
			return fmt.Sprintf("open -a %s%s%s", quoteChar, inner, quoteChar)
		}

		return trimmed
	}

	// Handle unquoted commands - check first word
	parts := strings.Fields(trimmed)
	if len(parts) > 0 && strings.HasSuffix(parts[0], ".app") {
		// If the .app path contains spaces, we need to quote it
		appPath := parts[0]
		if strings.Contains(appPath, " ") {
			appPath = `"` + appPath + `"`
		}

		// Reconstruct with open -a and remaining arguments
		if len(parts) > 1 {
			return fmt.Sprintf("open -a %s %s", appPath, strings.Join(parts[1:], " "))
		}
		return fmt.Sprintf("open -a %s", appPath)
	}

	return trimmed
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
	if strings.TrimSpace(command) == "" {
		return fmt.Errorf("empty command")
	}

	// Simple validation - just check if the main command/path exists
	// Since the command string is already properly parsed from shell args,
	// we just need to extract the executable part

	var cmdPath string

	// Remove outer quotes if present (they were added by ensureQuotedIfNeeded)
	trimmed := strings.TrimSpace(command)
	if (strings.HasPrefix(trimmed, `"`) && strings.HasSuffix(trimmed, `"`)) ||
		(strings.HasPrefix(trimmed, `'`) && strings.HasSuffix(trimmed, `'`)) {
		// Extract content inside quotes
		inner := trimmed[1 : len(trimmed)-1]

		// If the quoted content contains spaces, it could be either:
		// 1. A path with spaces (like "/Applications/Visual Studio Code.app/bin/code")
		// 2. A command with arguments (like "java -jar file.jar" or "./script.js arg1 arg2")
		//
		// Strategy:
		// - If it starts with / and doesn't look like "command args", treat as full path
		// - If it starts with ./ or ../ and doesn't look like "script.ext args", treat as full path
		// - Otherwise, take the first word as the command

		if strings.HasPrefix(inner, "/") {
			// Absolute path - check if it looks like a command with arguments
			parts := strings.Fields(inner)
			if len(parts) == 1 {
				cmdPath = inner // Single absolute path
			} else {
				// Multiple parts - could be "/usr/bin/java -jar app.jar" or "/Applications/Visual Studio Code.app"
				// If the first part doesn't end with a typical executable extension, assume it's a path with spaces
				firstWord := parts[0]
				if strings.HasSuffix(firstWord, ".app") ||
					(!strings.Contains(firstWord, ".") && !strings.HasSuffix(firstWord, "/bin/java") && !strings.HasSuffix(firstWord, "/bin/node")) {
					cmdPath = inner // Treat whole thing as path (like "/Applications/Visual Studio Code.app")
				} else {
					cmdPath = firstWord // Treat as command with args (like "/usr/bin/java -jar app.jar")
				}
			}
		} else if strings.HasPrefix(inner, "./") || strings.HasPrefix(inner, "../") {
			// Relative path - check if it has arguments
			parts := strings.Fields(inner)
			firstWord := parts[0]
			if len(parts) == 1 {
				cmdPath = inner // Single relative path
			} else if strings.Contains(firstWord, ".") {
				cmdPath = firstWord // Relative script with args (like "./script.js args")
			} else {
				cmdPath = inner // Relative path with spaces
			}
		} else {
			// Not a path, treat as command with arguments
			parts := strings.Fields(inner)
			if len(parts) > 0 {
				cmdPath = parts[0]
			}
		}
	} else {
		// For unquoted commands, take the first word
		parts := strings.Fields(trimmed)
		if len(parts) > 0 {
			cmdPath = parts[0]
		}
	}

	if cmdPath == "" {
		return fmt.Errorf("could not determine command path")
	}

	// If it's a path (contains / or starts with ./ or ../), check if it exists
	if strings.Contains(cmdPath, "/") {
		if !filepath.IsAbs(cmdPath) {
			if absPath, err := filepath.Abs(cmdPath); err == nil {
				cmdPath = absPath
			}
		}

		// Check if the path exists
		if _, err := os.Stat(cmdPath); err != nil {
			return fmt.Errorf("command '%s' not found", cmdPath)
		}

		// For .app bundles, we don't need to check executable permissions
		if strings.HasSuffix(cmdPath, ".app") {
			return nil
		}

		// If it's a regular file, check if it's executable
		if fileInfo, err := os.Stat(cmdPath); err == nil && fileInfo.Mode().IsRegular() {
			return h.checkExecutable(cmdPath)
		}
	} else {
		// For commands without paths (like 'java', 'node', etc.), assume they're in PATH
		// We don't validate PATH commands as they may not be available during testing
		// or the user might install them later
		if strings.ContainsAny(cmdPath, "{}[]()<>|&;") {
			return fmt.Errorf("command '%s' contains invalid characters", cmdPath)
		}
	}

	return nil
}
