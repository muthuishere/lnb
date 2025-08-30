package main

import (
	"fmt"
	"os"
	"strings"
)

// Build-time variables (set via ldflags)
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func showHelp() {
	fmt.Printf(`LNB v%s - Cross-Platform Alias Manager

USAGE:
    lnb <command> [options]

COMMANDS:
    alias <name> "<command>"    Create an alias for a command
    unalias <name>              Remove an alias
    <file-path>                 Make a binary globally accessible
    remove <name>               Remove a binary or alias
    list                        List everything
    help                        Show this help
    version                     Show version

EXAMPLES:
    lnb alias deploy "docker run --rm -v $(pwd):/app deploy-image"
    lnb alias logs "tail -f /var/log/nginx/access.log"  
    lnb ./mybinary              Make binary globally accessible
    lnb remove mybinary         Remove binary
    lnb unalias deploy          Remove alias
    lnb list                    Show everything

Same command. All platforms.
Source: https://github.com/muthuishere/lnb
`, version)
}

func showVersion() {
	fmt.Printf("lnb v%s\n", version)
	if commit != "unknown" && date != "unknown" {
		fmt.Printf("Built from %s on %s\n", commit, date)
	}
}

func isFilePath(arg string) bool {
	// Check if it's a file path
	if strings.Contains(arg, "/") || strings.Contains(arg, "\\") {
		return true
	}

	// Check if file exists in current directory
	if _, err := os.Stat(arg); err == nil {
		return true
	}

	return false
}

func isKnownCommand(cmd string) bool {
	knownCommands := []string{
		"help", "-h", "--help",
		"version", "-v", "--version",
		"list", "ls", "--ls",
		"alias", "unalias",
		"install", "remove",
	}

	for _, known := range knownCommands {
		if cmd == known {
			return true
		}
	}
	return false
}

func main() {
	// If no arguments, show help
	if len(os.Args) < 2 {
		showHelp()
		return
	}

	command := strings.ToLower(os.Args[1])
	args := os.Args[2:] // Remaining arguments after command

	// Check if first argument is a file path instead of a command
	if !isKnownCommand(command) && isFilePath(os.Args[1]) {
		// Treat as install command with the file path
		handleBinaryCommand("install", os.Args[1:])
		return
	}

	// Handle commands
	switch command {
	case "help", "-h", "--help":
		showHelp()
	case "version", "-v", "--version":
		showVersion()
	case "list", "ls", "--ls":
		handleListCommand()
	case "alias":
		handleAliasCommand(args)
	case "unalias":
		handleUnaliasCommand(args)
	case "install", "remove":
		handleBinaryCommand(command, args)
	default:
		fmt.Printf("Error: Unknown command '%s'\n", command)
		fmt.Println("Use 'lnb help' for usage information.")
		os.Exit(1)
	}
}
