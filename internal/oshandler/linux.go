package oshandler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"lnb/internal/config"
)

// parseShellArgs parses a command string into arguments while respecting quotes
func parseShellArgsLinux(command string) []string {
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

// reconstructCommandLinux rebuilds a command string from parsed arguments
func reconstructCommandLinux(args []string) string {
	return strings.Join(args, " ")
}

type linuxHandler struct{}

func (h *linuxHandler) Handle(absPath, action string) error {
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

func (h *linuxHandler) HandleAlias(aliasName, command, action string) error {
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
func (h *linuxHandler) convertRelativePaths(command string) string {
	args := parseShellArgsLinux(command)

	for i, arg := range args {
		// Remove quotes temporarily to check the path, but preserve them in the final result
		unquotedArg := arg
		hasQuotes := false
		var quoteChar string

		if (strings.HasPrefix(arg, `"`) && strings.HasSuffix(arg, `"`)) ||
			(strings.HasPrefix(arg, `'`) && strings.HasSuffix(arg, `'`)) {
			hasQuotes = true
			quoteChar = string(arg[0])
			unquotedArg = arg[1 : len(arg)-1]
		}

		// Check if this looks like a relative path
		if strings.HasPrefix(unquotedArg, "./") || strings.HasPrefix(unquotedArg, "../") ||
			(strings.Contains(unquotedArg, ".") && !strings.HasPrefix(unquotedArg, "/") && !strings.Contains(unquotedArg, "://")) {
			if absPath, err := filepath.Abs(unquotedArg); err == nil {
				// Verify the file exists before converting
				if _, err := os.Stat(absPath); err == nil {
					if hasQuotes {
						args[i] = quoteChar + absPath + quoteChar
					} else {
						args[i] = absPath
					}
				}
			}
		}
	}

	return reconstructCommandLinux(args)
}

// checkExecutable verifies if a file is executable
func (h *linuxHandler) checkExecutable(path string) error {
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
func (h *linuxHandler) validateCommand(command string) error {
	if strings.TrimSpace(command) == "" {
		return fmt.Errorf("empty command")
	}

	// Parse the command properly respecting quotes
	args := parseShellArgsLinux(command)
	if len(args) == 0 {
		return fmt.Errorf("could not parse command")
	}

	// Get the first argument (the command/executable)
	cmdName := args[0]

	// Remove quotes if present to check the actual path
	if (strings.HasPrefix(cmdName, `"`) && strings.HasSuffix(cmdName, `"`)) ||
		(strings.HasPrefix(cmdName, `'`) && strings.HasSuffix(cmdName, `'`)) {
		cmdName = cmdName[1 : len(cmdName)-1]
	}

	// If it's a path (absolute or relative), check if it exists
	if strings.Contains(cmdName, "/") {
		var absPath string
		var err error

		if filepath.IsAbs(cmdName) {
			absPath = cmdName
		} else {
			absPath, err = filepath.Abs(cmdName)
			if err != nil {
				return fmt.Errorf("could not resolve path '%s': %v", cmdName, err)
			}
		}

		// Check if the path exists
		if _, err := os.Stat(absPath); err != nil {
			return fmt.Errorf("command '%s' not found", cmdName)
		}

		// If it's a regular file, check if it's executable
		if fileInfo, err := os.Stat(absPath); err == nil && fileInfo.Mode().IsRegular() {
			return h.checkExecutable(absPath)
		}

		return nil
	} else {
		// For commands in PATH, do basic validation
		if strings.ContainsAny(cmdName, "{}[]()<>|&;") {
			return fmt.Errorf("command '%s' contains invalid characters", cmdName)
		}
	}

	return nil
}
