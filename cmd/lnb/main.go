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
	fmt.Printf(`LNB v%s - Link Binary

A cross-platform utility that makes command-line tools accessible from anywhere
by creating symbolic links or wrapper scripts in your system's PATH.

USAGE:
    lnb [COMMAND] [OPTIONS]

COMMANDS:
    install <file>              Install a binary to your PATH
    remove <file>               Remove a binary from your PATH
    alias <name> <command>      Create an alias that runs a command
    unalias <name>              Remove an alias
    list                        List all binaries and aliases installed by LNB
    help                        Show this help message
    version                     Show version information

EXAMPLES:
    lnb install ./mybinary                   # Install mybinary
    lnb remove ./mybinary                    # Remove mybinary
    lnb alias myapp "java -jar ./app.jar"    # Create alias for Java app
    lnb unalias myapp                        # Remove alias
    lnb list                                 # Show all installed binaries and aliases
    lnb help                                 # Show this help

For more information, visit: https://github.com/muthuishere/lnb
`, version)
}

func showVersion() {
	fmt.Printf("LNB v%s\n", version)
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
		"list", "ls",
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
	case "list", "ls":
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
