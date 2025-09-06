package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// LnbEntry represents a single installed binary
type LnbEntry struct {
	Name        string    `json:"name"`
	SourcePath  string    `json:"source_path"`
	TargetPath  string    `json:"target_path"`
	InstalledAt time.Time `json:"installed_at"`
}

// Config represents the LNB configuration
type Config struct {
	Entries map[string]*LnbEntry `json:"entries"`
	Version string               `json:"version"`
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	var configDir string

	// Check if we're in test mode
	if testConfigDir := os.Getenv("LNB_TEST_CONFIG_DIR"); testConfigDir != "" {
		configDir = testConfigDir
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %v", err)
		}
		configDir = filepath.Join(homeDir, ".lnb")
	}

	configFile := filepath.Join(configDir, "config.json")

	// Ensure the config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %v", err)
	}

	return configFile, nil
}

// Load reads the config file
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{
			Entries: make(map[string]*LnbEntry),
			Version: "1.0",
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}

// Save writes the config file
func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// AddEntry adds a new entry to the config
func (c *Config) AddEntry(name, sourcePath, targetPath string) {
	// Initialize entries map if nil
	if c.Entries == nil {
		c.Entries = make(map[string]*LnbEntry)
	}

	// Add new entry
	c.Entries[name] = &LnbEntry{
		Name:        name,
		SourcePath:  sourcePath,
		TargetPath:  targetPath,
		InstalledAt: time.Now(),
	}
}

// RemoveEntry removes an entry from the config
func (c *Config) RemoveEntry(name string) {
	if c.Entries != nil {
		delete(c.Entries, name)
	}
}

// GetEntry finds an entry by name
func (c *Config) GetEntry(name string) (*LnbEntry, bool) {
	if c.Entries == nil {
		return nil, false
	}
	entry, exists := c.Entries[name]
	return entry, exists
}

// List returns all entries as a slice
func (c *Config) List() []*LnbEntry {
	if c.Entries == nil {
		return []*LnbEntry{}
	}

	entries := make([]*LnbEntry, 0, len(c.Entries))
	for _, entry := range c.Entries {
		entries = append(entries, entry)
	}
	return entries
}
