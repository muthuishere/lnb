package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// parseShellArgs parses a command string into arguments while respecting quotes
func parseShellArgs(command string) []string {
	// First check if the entire command is a valid file path
	// This handles cases like "/Applications/Visual Studio Code.app"
	if _, err := os.Stat(command); err == nil {
		return []string{command}
	}
	
	// Also check if it's an absolute path that might exist (common on macOS)
	if filepath.IsAbs(command) {
		if _, err := os.Stat(command); err == nil {
			return []string{command}
		}
	}

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

// getAliasInputs prompts for or gets alias name and command from arguments
func getAliasInputs(args []string) (string, string) {
	var aliasName, aliasCommand string

	if len(args) < 2 {
		// Interactive prompt for alias
		aliasName = promptForAliasName()
		aliasCommand = promptForAliasCommand()
	} else {
		aliasName = args[0]
		// Join all remaining arguments to handle commands with spaces
		aliasCommand = strings.Join(args[1:], " ")
	}

	return aliasName, aliasCommand
}

// promptForAliasName prompts user for alias name
func promptForAliasName() string {
	fmt.Print("Enter alias name: ")
	var aliasName string
	fmt.Scanln(&aliasName)
	if aliasName == "" {
		fmt.Println("Error: Alias name cannot be empty.")
		os.Exit(1)
	}
	return aliasName
}

// promptForAliasCommand prompts user for alias command
func promptForAliasCommand() string {
	fmt.Print("Enter command: ")
	// Read the full line to handle commands with spaces
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		aliasCommand := strings.TrimSpace(scanner.Text())
		if aliasCommand == "" {
			fmt.Println("Error: Command cannot be empty.")
			os.Exit(1)
		}
		return aliasCommand
	}
	fmt.Println("Error: Failed to read command.")
	os.Exit(1)
	return ""
}

// validateAliasInputs validates alias name and command
func validateAliasInputs(aliasName string, aliasCommand *string) {
	if strings.TrimSpace(aliasName) == "" {
		fmt.Println("Error: Alias name cannot be empty.")
		os.Exit(1)
	}
	if strings.TrimSpace(*aliasCommand) == "" {
		fmt.Println("Error: Command cannot be empty.")
		os.Exit(1)
	}

	// Validate and normalize the command
	if err := validateAndNormalizeCommand(aliasCommand); err != nil {
		fmt.Printf("Error: invalid command '%s': %v\n", *aliasCommand, err)
		os.Exit(1)
	}
}

// validateAndNormalizeCommand validates a command and converts relative paths to absolute paths
func validateAndNormalizeCommand(command *string) error {
	if command == nil || *command == "" {
		return fmt.Errorf("command cannot be empty")
	}

	// Parse the command to extract the main executable
	args := parseShellArgs(*command)
	if len(args) == 0 {
		return fmt.Errorf("could not parse command")
	}

	// Get the first argument (the command/executable)
	cmdName := args[0]
	fmt.Printf("DEBUG: Original cmdName: '%s'\n", cmdName)

	// Remove quotes if present to check the actual path
	originalCmdName := cmdName
	if (strings.HasPrefix(cmdName, `"`) && strings.HasSuffix(cmdName, `"`)) ||
		(strings.HasPrefix(cmdName, `'`) && strings.HasSuffix(cmdName, `'`)) {
		cmdName = cmdName[1 : len(cmdName)-1]
		fmt.Printf("DEBUG: After removing quotes: '%s'\n", cmdName)
	}

	// Handle tilde expansion
	if strings.HasPrefix(cmdName, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not get home directory: %v", err)
		}
		cmdName = filepath.Join(homeDir, cmdName[2:])
	}

	// Check if this looks like a path (contains path separators or file extensions)
	isPath := strings.Contains(cmdName, "/") || strings.Contains(cmdName, "\\") ||
		strings.HasPrefix(cmdName, "./") || strings.HasPrefix(cmdName, "../")

	if isPath {
		// This appears to be a path, validate it exists and convert to absolute
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
			return fmt.Errorf("file not found: %s", absPath)
		}

		// Update the command with the absolute path
		if originalCmdName != cmdName {
			// Had quotes, preserve them
			if strings.HasPrefix(originalCmdName, `"`) {
				args[0] = `"` + absPath + `"`
			} else {
				args[0] = `'` + absPath + `'`
			}
		} else {
			// No quotes originally
			if strings.Contains(absPath, " ") {
				args[0] = `"` + absPath + `"`
			} else {
				args[0] = absPath
			}
		}

		// Reconstruct the command
		*command = strings.Join(args, " ")
		fmt.Printf("📁 Validated file path: %s\n", absPath)
		return nil
	} else {
		// This appears to be a command name, not a file path
		fmt.Printf("💻 Command '%s' will be executed as-is (assuming it's available in PATH or installed)\n", cmdName)

		// Just do basic sanity checks
		if strings.ContainsAny(cmdName, "{}[]()<>|&;") {
			return fmt.Errorf("command '%s' contains potentially dangerous characters", cmdName)
		}

		return nil // Allow the command - let the system handle execution
	}
}

// handleCreateAlias handles alias creation
func handleCreateAlias(aliasName, aliasCommand string) {
	validateAliasInputs(aliasName, &aliasCommand)

	// Command is already validated and normalized, pass it directly to the handler
	handler := getOSHandler()

	if err := handler.HandleAlias(aliasName, aliasCommand, "install"); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Successfully created alias '%s' for command '%s'\n", aliasName, aliasCommand)
}

// handleRemoveAlias handles alias removal
func handleRemoveAlias(aliasName string) {
	if strings.TrimSpace(aliasName) == "" {
		fmt.Println("Error: Alias name cannot be empty.")
		os.Exit(1)
	}

	handler := getOSHandler()

	if err := handler.HandleAlias(aliasName, "", "remove"); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Successfully removed alias '%s'\n", aliasName)
}

// displayEntries displays the list of installed binaries and aliases
func displayEntries(entries map[string]*LNBEntry) {
	if len(entries) == 0 {
		fmt.Println("No binaries or aliases installed by LNB.")
		return
	}

	fmt.Printf("Binaries and aliases installed by LNB (%d):\n\n", len(entries))
	for _, entry := range entries {
		fmt.Printf("  %s\n", entry.Name)
		if strings.HasPrefix(entry.SourcePath, "alias:") {
			fmt.Printf("    Type:      alias\n")
			fmt.Printf("    Command:   %s\n", strings.TrimPrefix(entry.SourcePath, "alias:"))
		} else {
			fmt.Printf("    Type:      binary\n")
			fmt.Printf("    Source:    %s\n", entry.SourcePath)
		}
		fmt.Printf("    Target:    %s\n", entry.TargetPath)
		fmt.Printf("    Installed: %s\n", entry.InstalledAt.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}
}

// handleAliasCommand handles alias creation
func handleAliasCommand(args []string) {
	aliasName, aliasCommand := getAliasInputs(args)
	handleCreateAlias(aliasName, aliasCommand)
}

// handleUnaliasCommand handles alias removal
func handleUnaliasCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: unalias command requires an alias name.")
		fmt.Println("Usage: lnb unalias <name>")
		os.Exit(1)
	}

	aliasName := args[0]
	handleRemoveAlias(aliasName)
}

// handleListCommand lists all installed binaries and aliases
func handleListCommand() {
	entries := listEntries()
	displayEntries(entries)
}
