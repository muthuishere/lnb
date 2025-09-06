package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LNBEntry represents an installed binary or alias
type LNBEntry struct {
	Name        string    `json:"name"`
	SourcePath  string    `json:"source_path"` // source path for binaries, or "alias:command" for aliases
	TargetPath  string    `json:"target_path"` // target path (symlink/script location)
	InstalledAt time.Time `json:"installed_at"`
}

// LNBConfig represents the configuration file
type LNBConfig struct {
	Entries map[string]*LNBEntry `json:"entries"`
	Version string               `json:"version"`
}

// getConfigPath returns the path to the LNB configuration file
func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error: Unable to get home directory: %v\n", err)
		os.Exit(1)
	}

	configDir := filepath.Join(homeDir, ".lnb")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Printf("Error: Unable to create config directory: %v\n", err)
		os.Exit(1)
	}

	return filepath.Join(configDir, "config.json")
}

// loadConfig loads the LNB configuration from file
func loadConfig() *LNBConfig {
	configPath := getConfigPath()

	// If config doesn't exist, create new one
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &LNBConfig{
			Entries: make(map[string]*LNBEntry),
			Version: "1.0",
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Error: Unable to read config file: %v\n", err)
		os.Exit(1)
	}

	var config LNBConfig
	if err := json.Unmarshal(data, &config); err != nil {
		// If parsing fails, it might be the old array format - create a new config
		fmt.Printf("Warning: Config file format is outdated, creating new config: %v\n", err)
		return &LNBConfig{
			Entries: make(map[string]*LNBEntry),
			Version: "1.0",
		}
	}

	// Initialize entries map if nil (for backwards compatibility)
	if config.Entries == nil {
		config.Entries = make(map[string]*LNBEntry)
	}

	return &config
}

// saveConfig saves the LNB configuration to file
func saveConfig(config *LNBConfig) {
	configPath := getConfigPath()

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Printf("Error: Unable to marshal config: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		fmt.Printf("Error: Unable to write config file: %v\n", err)
		os.Exit(1)
	}
}

// addEntry adds an entry to the configuration
func addEntry(name, sourcePath, targetPath string) {
	config := loadConfig()

	// Check if entry already exists
	if _, exists := config.Entries[name]; exists {
		entryType := "binary"
		if strings.HasPrefix(sourcePath, "alias:") {
			entryType = "alias"
		}
		fmt.Printf("Error: %s '%s' is already installed. Use 'lnb remove %s' first to reinstall\n", entryType, name, name)
		os.Exit(1)
	}

	config.Entries[name] = &LNBEntry{
		Name:        name,
		SourcePath:  sourcePath,
		TargetPath:  targetPath,
		InstalledAt: time.Now(),
	}

	saveConfig(config)
}

// removeEntry removes an entry from the configuration
func removeEntry(name string) *LNBEntry {
	config := loadConfig()

	entry, exists := config.Entries[name]
	if !exists {
		fmt.Printf("Error: '%s' was not installed by LNB\n", name)
		os.Exit(1)
	}

	delete(config.Entries, name)
	saveConfig(config)

	return entry
}

// getEntry retrieves an entry from the configuration
func getEntry(name string) *LNBEntry {
	config := loadConfig()
	return config.Entries[name]
}

// listEntries returns all entries from the configuration
func listEntries() map[string]*LNBEntry {
	config := loadConfig()
	return config.Entries
}
