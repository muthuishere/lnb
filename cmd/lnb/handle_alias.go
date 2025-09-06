package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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

// ensureQuotedIfNeeded adds quotes to an argument if it contains spaces and isn't already quoted
// This function is smart about not double-quoting complex commands
func ensureQuotedIfNeeded(arg string) string {
	// If the command already contains quotes or is a complex command, don't modify it
	if strings.Contains(arg, `"`) || strings.Contains(arg, `'`) {
		return arg
	}

	// If it's a simple path with spaces, add quotes
	if strings.Contains(arg, " ") && !strings.ContainsAny(arg, "|&;<>(){}[]") {
		return `"` + arg + `"`
	}

	return arg
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
func validateAliasInputs(aliasName, aliasCommand string) {
	if strings.TrimSpace(aliasName) == "" {
		fmt.Println("Error: Alias name cannot be empty.")
		os.Exit(1)
	}
	if strings.TrimSpace(aliasCommand) == "" {
		fmt.Println("Error: Command cannot be empty.")
		os.Exit(1)
	}
}

// handleCreateAlias handles alias creation
func handleCreateAlias(aliasName, aliasCommand string) {
	validateAliasInputs(aliasName, aliasCommand)

	// Ensure the command is properly quoted for shell scripts
	quotedCommand := ensureQuotedIfNeeded(aliasCommand)

	handler := getOSHandler()

	if err := handler.HandleAlias(aliasName, quotedCommand, "install"); err != nil {
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
