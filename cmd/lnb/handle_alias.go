package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"lnb/internal/config"
)

// getAliasInputs prompts for or gets alias name and command from arguments
func getAliasInputs(args []string) (string, string) {
	var aliasName, aliasCommand string

	if len(args) < 2 {
		// Interactive prompt for alias
		aliasName = promptForAliasName()
		aliasCommand = promptForAliasCommand()
	} else {
		aliasName = args[0]
		aliasCommand = args[1]
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

// loadConfig loads the LNB configuration
func loadConfig() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error: failed to load config: %v\n", err)
		os.Exit(1)
	}
	return cfg
}

// displayEntries displays the list of installed binaries and aliases
func displayEntries(entries []config.LnbEntry) {
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
	cfg := loadConfig()
	entries := cfg.List()
	displayEntries(entries)
}
