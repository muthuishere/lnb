package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"lnb/internal/config"
	"lnb/internal/oshandler"
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
    install <file>              Install a binary to your PATH (default action)
    remove <file>               Remove a binary from your PATH
    alias <name> <command>      Create an alias that runs a command
    unalias <name>              Remove an alias
    list                        List all binaries and aliases installed by LNB
    help                        Show this help message
    version                     Show version information

EXAMPLES:
    lnb ./mybinary                           # Install mybinary (default action)
    lnb ./mybinary install                   # Install mybinary explicitly
    lnb ./mybinary remove                    # Remove mybinary
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

func listInstalled() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	entries := cfg.List()
	if len(entries) == 0 {
		fmt.Println("No binaries installed by LNB.")
		return nil
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

	return nil
}

func main() {
	// If no arguments, show help
	if len(os.Args) < 2 {
		showHelp()
		return
	}

	command := strings.ToLower(os.Args[1])

	// Handle commands that don't require a file argument
	switch command {
	case "help", "-h", "--help":
		showHelp()
		return
	case "version", "-v", "--version":
		showVersion()
		return
	case "list", "ls":
		if err := listInstalled(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	case "alias":
		if len(os.Args) < 4 {
			fmt.Println("Error: alias command requires both name and command.")
			fmt.Println("Usage: lnb alias <name> <command>")
			fmt.Println("Example: lnb alias myapp \"java -jar ./app.jar\"")
			os.Exit(1)
		}
		aliasName := os.Args[2]
		aliasCommand := os.Args[3]

		handler := oshandler.New()
		if handler == nil {
			fmt.Println("Error: Unsupported operating system")
			os.Exit(1)
		}

		if err := handler.HandleAlias(aliasName, aliasCommand, "install"); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	case "unalias":
		if len(os.Args) < 3 {
			fmt.Println("Error: unalias command requires an alias name.")
			fmt.Println("Usage: lnb unalias <name>")
			os.Exit(1)
		}
		aliasName := os.Args[2]

		handler := oshandler.New()
		if handler == nil {
			fmt.Println("Error: Unsupported operating system")
			os.Exit(1)
		}

		if err := handler.HandleAlias(aliasName, "", "remove"); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// For install/remove operations, we need at least a filename
	if len(os.Args) < 2 {
		fmt.Println("Error: Please specify a file to install or remove.")
		fmt.Println("Use 'lnb help' for usage information.")
		os.Exit(1)
	}

	var filename, action string

	// Parse arguments for install/remove operations
	if command == "install" || command == "remove" {
		if len(os.Args) < 3 {
			fmt.Printf("Error: Please specify a file to %s.\n", command)
			fmt.Println("Use 'lnb help' for usage information.")
			os.Exit(1)
		}
		action = command
		filename = os.Args[2]
	} else {
		// Default behavior: first arg is filename, second is action (defaulting to install)
		filename = os.Args[1]
		action = "install"
		if len(os.Args) > 2 {
			possibleAction := strings.ToLower(os.Args[2])
			if possibleAction == "install" || possibleAction == "remove" {
				action = possibleAction
			}
		}
	}

	// Validate that the file exists for install operations
	if action == "install" {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			fmt.Printf("Error: File '%s' does not exist.\n", filename)
			os.Exit(1)
		}
	}

	absPath, err := filepath.Abs(filename)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	handler := oshandler.New()
	if handler == nil {
		fmt.Println("Error: Unsupported operating system")
		os.Exit(1)
	}

	if err := handler.Handle(absPath, action); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
