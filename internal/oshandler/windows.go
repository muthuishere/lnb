package oshandler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"lnb/internal/config"
)

// parseShellArgsWindows parses a command string into arguments while respecting quotes
func parseShellArgsWindows(command string) []string {
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

// reconstructCommandWindows rebuilds a command string from parsed arguments
func reconstructCommandWindows(args []string) string {
	return strings.Join(args, " ")
}

type windowsHandler struct{}

func (h *windowsHandler) Handle(absPath, action string) error {
	binDir := filepath.Join(os.Getenv("USERPROFILE"), "bin")
	linkName := filepath.Base(absPath)
	linkNameWithoutExt := strings.TrimSuffix(linkName, filepath.Ext(linkName))
	cmdPath := filepath.Join(binDir, linkNameWithoutExt+".cmd")

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

		// Check if this binary is already installed
		if entry, exists := cfg.GetEntry(linkNameWithoutExt); exists {
			// Verify the target file actually exists
			if _, err := os.Stat(entry.TargetPath); err == nil {
				return fmt.Errorf("binary '%s' is already installed. Use 'lnb remove %s' first to reinstall", linkNameWithoutExt, linkNameWithoutExt)
			} else {
				// Config says it's installed but file doesn't exist - clean up the config
				fmt.Printf("Warning: Config shows '%s' as installed but target file '%s' doesn't exist. Cleaning up config entry.\n", linkNameWithoutExt, entry.TargetPath)
				cfg.RemoveEntry(linkNameWithoutExt)
				if err := cfg.Save(); err != nil {
					fmt.Printf("Warning: failed to clean up config: %v\n", err)
				}
			}
		}

		// Check if the target path already exists
		if _, err := os.Stat(cmdPath); err == nil {
			return fmt.Errorf("file already exists at %s. Please remove it manually or use 'lnb remove %s' if it was installed by LNB", cmdPath, linkNameWithoutExt)
		}

		err := os.MkdirAll(binDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating bin dir: %v", err)
		}

		cmdContents := fmt.Sprintf(`@echo off
"%s" %%*
`, absPath)

		err = os.WriteFile(cmdPath, []byte(cmdContents), 0755)
		if err != nil {
			return fmt.Errorf("failed to write wrapper: %v", err)
		}
		fmt.Printf("Installed: %s\n", cmdPath)

		// Automatically ensure the bin directory is in PATH
		h.ensureInPath(binDir)

		// Add to config
		cfg.AddEntry(linkNameWithoutExt, absPath, cmdPath)
		if err := cfg.Save(); err != nil {
			fmt.Printf("Warning: failed to update config: %v\n", err)
		}

	case "remove":
		// Check if this binary was installed by LNB
		entry, exists := cfg.GetEntry(linkNameWithoutExt)
		if !exists {
			return fmt.Errorf("binary '%s' was not installed by LNB", linkNameWithoutExt)
		}

		// Verify the target path matches what we expect
		if entry.TargetPath != cmdPath {
			return fmt.Errorf("binary '%s' target path mismatch: expected %s, found %s", linkNameWithoutExt, cmdPath, entry.TargetPath)
		}

		err := os.Remove(cmdPath)
		if err != nil {
			return fmt.Errorf("failed to remove: %v", err)
		}
		fmt.Printf("Removed: %s\n", cmdPath)

		// Remove from config
		cfg.RemoveEntry(linkNameWithoutExt)
		if err := cfg.Save(); err != nil {
			fmt.Printf("Warning: failed to update config: %v\n", err)
		}
	}
	return nil
}

func (h *windowsHandler) HandleAlias(aliasName, command, action string) error {
	binDir := filepath.Join(os.Getenv("USERPROFILE"), "bin")
	batPath := filepath.Join(binDir, aliasName+".bat")

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
		if _, err := os.Stat(batPath); err == nil {
			return fmt.Errorf("file already exists at %s. Please remove it manually or use 'lnb unalias %s' if it was installed by LNB", batPath, aliasName)
		}

		err := os.MkdirAll(binDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating bin dir: %v", err)
		}

		// Convert relative paths to absolute paths in the command
		convertedCommand := h.convertRelativePaths(command)

		// Create the batch file content
		batContent := fmt.Sprintf(`@echo off
%s %%*
`, convertedCommand)

		// Write the batch file
		err = os.WriteFile(batPath, []byte(batContent), 0755)
		if err != nil {
			return fmt.Errorf("failed to create alias batch file: %v", err)
		}

		fmt.Printf("Created alias: %s -> %s\n", aliasName, convertedCommand)

		// Automatically ensure the bin directory is in PATH
		h.ensureInPath(binDir)

		// Add to config with special marker for aliases
		cfg.AddEntry(aliasName, "alias:"+command, batPath)
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
		if entry.TargetPath != batPath {
			return fmt.Errorf("alias '%s' target path mismatch: expected %s, found %s", aliasName, batPath, entry.TargetPath)
		}

		err := os.Remove(batPath)
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

// convertRelativePaths converts relative paths in command to absolute paths (Windows version)
func (h *windowsHandler) convertRelativePaths(command string) string {
	args := parseShellArgsWindows(command)

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
		// On Windows, relative paths might start with .\ or ..\ or be just filenames
		if strings.HasPrefix(unquotedArg, ".\\") || strings.HasPrefix(unquotedArg, "..\\") ||
			strings.HasPrefix(unquotedArg, "./") || strings.HasPrefix(unquotedArg, "../") ||
			(strings.Contains(unquotedArg, ".") && !strings.Contains(unquotedArg, ":") && !strings.Contains(unquotedArg, "://")) {
			if absPath, err := filepath.Abs(unquotedArg); err == nil {
				// Verify the file exists before converting
				if _, err := os.Stat(absPath); err == nil {
					// Always preserve quotes if they were there, or add them if path contains spaces
					if hasQuotes {
						args[i] = quoteChar + absPath + quoteChar
					} else if strings.Contains(absPath, " ") {
						args[i] = `"` + absPath + `"`
					} else {
						args[i] = absPath
					}
				}
			}
		}
	}

	return reconstructCommandWindows(args)
}

// validateCommand checks if the command can be executed (basic validation)
func (h *windowsHandler) validateCommand(command string) error {
	if strings.TrimSpace(command) == "" {
		return fmt.Errorf("empty command")
	}

	// Parse the command properly respecting quotes
	args := parseShellArgsWindows(command)
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
	if strings.Contains(cmdName, "/") || strings.Contains(cmdName, "\\") {
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

		return nil
	} else {
		// For commands in PATH, do basic validation
		if strings.ContainsAny(cmdName, "{}[]()<>|&;") {
			return fmt.Errorf("command '%s' contains invalid characters", cmdName)
		}
	}

	return nil
}

// isInUserPath checks if the given directory is in the user's PATH
func (h *windowsHandler) isInUserPath(dir string) bool {
	cmd := exec.Command("powershell", "-Command",
		"[Environment]::GetEnvironmentVariable('Path', 'User') -split ';' | Where-Object { $_ -eq '"+dir+"' }")
	output, err := cmd.Output()
	return err == nil && len(strings.TrimSpace(string(output))) > 0
}

// addToUserPath adds a directory to the user's PATH environment variable
func (h *windowsHandler) addToUserPath(dir string) error {
	// Use PowerShell to add to user PATH
	cmd := exec.Command("powershell", "-Command",
		"$currentPath = [Environment]::GetEnvironmentVariable('Path', 'User'); "+
			"if ($currentPath -notlike '*"+dir+"*') { "+
			"$newPath = if ($currentPath) { $currentPath + ';' + '"+dir+"' } else { '"+dir+"' }; "+
			"[Environment]::SetEnvironmentVariable('Path', $newPath, 'User'); "+
			"Write-Host 'Added to PATH' } else { Write-Host 'Already in PATH' }")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add directory to PATH: %v\nOutput: %s", err, string(output))
	}

	fmt.Printf("🔧 %s\n", strings.TrimSpace(string(output)))
	return nil
}

// ensureInPath ensures the bin directory is in the user's PATH
func (h *windowsHandler) ensureInPath(binDir string) {
	if !h.isInUserPath(binDir) {
		fmt.Printf("📍 Adding %s to your PATH...\n", binDir)
		if err := h.addToUserPath(binDir); err != nil {
			fmt.Printf("⚠️  Failed to automatically add to PATH: %v\n", err)
			fmt.Printf("⚠️  Please manually add %s to your PATH environment variable\n", binDir)
		} else {
			fmt.Println("✅ Successfully added to PATH! Restart your terminal to use the new PATH.")
		}
	} else {
		fmt.Printf("✅ %s is already in your PATH\n", binDir)
	}
}
