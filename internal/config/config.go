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
	Entries []LnbEntry `json:"entries"`
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %v", err)
	}

	configDir := filepath.Join(homeDir, ".lnb")
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
		return &Config{Entries: []LnbEntry{}}, nil
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
	// Remove existing entry if it exists
	c.RemoveEntry(name)

	// Add new entry
	entry := LnbEntry{
		Name:        name,
		SourcePath:  sourcePath,
		TargetPath:  targetPath,
		InstalledAt: time.Now(),
	}
	c.Entries = append(c.Entries, entry)
}

// RemoveEntry removes an entry from the config
func (c *Config) RemoveEntry(name string) {
	for i, entry := range c.Entries {
		if entry.Name == name {
			c.Entries = append(c.Entries[:i], c.Entries[i+1:]...)
			return
		}
	}
}

// GetEntry finds an entry by name
func (c *Config) GetEntry(name string) (*LnbEntry, bool) {
	for _, entry := range c.Entries {
		if entry.Name == name {
			return &entry, true
		}
	}
	return nil, false
}

// List returns all entries
func (c *Config) List() []LnbEntry {
	return c.Entries
}
