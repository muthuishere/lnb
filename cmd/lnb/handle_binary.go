package main

import (
	"fmt"
	"os"
	"path/filepath"

	"lnb/internal/oshandler"
)

// getBinaryPath prompts for or gets the binary path from arguments
func getBinaryPath(command string, args []string) string {
	if len(args) < 1 {
		if command == "install" {
			// Interactive prompt for install
			fmt.Print("Enter path to binary: ")
			var filename string
			fmt.Scanln(&filename)
			if filename == "" {
				fmt.Println("Error: File path cannot be empty.")
				os.Exit(1)
			}
			return filename
		} else {
			// Error for remove command
			fmt.Printf("Error: Please specify a file to %s.\n", command)
			fmt.Println("Use 'lnb help' for usage information.")
			os.Exit(1)
		}
	}
	return args[0]
}

// validateBinaryExists checks if the binary file exists
func validateBinaryExists(filename string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("Error: File '%s' does not exist.\n", filename)
		os.Exit(1)
	}
}

// getAbsolutePath converts relative path to absolute path
func getAbsolutePath(filename string) string {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	return absPath
}

// getOSHandler creates and returns the appropriate OS handler
func getOSHandler() oshandler.Handler {
	handler := oshandler.New()
	if handler == nil {
		fmt.Println("Error: Unsupported operating system")
		os.Exit(1)
	}
	return handler
}

// handleInstallBinary handles the installation of a binary
func handleInstallBinary(filename string) {
	validateBinaryExists(filename)
	absPath := getAbsolutePath(filename)
	handler := getOSHandler()

	if err := handler.Handle(absPath, "install"); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Successfully installed '%s'\n", filepath.Base(filename))
}

// handleRemoveBinary handles the removal of a binary
func handleRemoveBinary(filename string) {
	absPath := getAbsolutePath(filename)
	handler := getOSHandler()

	if err := handler.Handle(absPath, "remove"); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Successfully removed '%s'\n", filepath.Base(filename))
}

// handleBinaryCommand handles install and remove commands for binaries
func handleBinaryCommand(command string, args []string) {
	filename := getBinaryPath(command, args)

	switch command {
	case "install":
		handleInstallBinary(filename)
	case "remove":
		handleRemoveBinary(filename)
	default:
		fmt.Printf("Error: Unknown binary command '%s'\n", command)
		os.Exit(1)
	}
}
