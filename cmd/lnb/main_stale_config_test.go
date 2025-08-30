package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"lnb/internal/config"
)

// TestStaleConfigCleanup tests that LNB cleans up stale config entries when target files don't exist
func TestStaleConfigCleanup(t *testing.T) {
	// Create a temporary test binary
	tempDir := t.TempDir()
	testBinaryPath := filepath.Join(tempDir, "stale-test-binary")

	// Create a simple test script
	testContent := `#!/bin/bash
echo "This is a test binary for stale config cleanup"
`

	err := os.WriteFile(testBinaryPath, []byte(testContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	binaryName := filepath.Base(testBinaryPath)
	targetPath := "/usr/local/bin/" + binaryName

	// Cleanup function
	cleanup := func() {
		os.Remove(targetPath)
		// Also clean up any config entries
		if cfg, err := config.Load(); err == nil {
			cfg.RemoveEntry(binaryName)
			cfg.Save()
		}
	}
	defer cleanup()

	// Step 1: Install the binary normally
	cmd := exec.Command("go", "run", ".", "install", testBinaryPath)
	cmd.Dir = "/Users/muthuishere/muthu/gitworkspace/lnb-workspace/lnb/cmd/lnb"
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Skipf("Cannot install test binary (need sudo?): %v\nOutput: %s", err, string(output))
		return
	}

	// Verify it was installed
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Errorf("Binary was not installed at expected location: %s", targetPath)
		return
	}

	// Step 2: Manually remove the target file (simulating external deletion)
	err = os.Remove(targetPath)
	if err != nil {
		t.Fatalf("Failed to manually remove target file: %v", err)
	}

	// Verify target file is gone but config entry still exists
	if _, err := os.Stat(targetPath); err == nil {
		t.Errorf("Target file should have been removed")
		return
	}

	// Verify config entry still exists
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if _, exists := cfg.GetEntry(binaryName); !exists {
		t.Errorf("Config entry should still exist after manual file removal")
		return
	}

	// Step 3: Try to install the same binary again
	// This should detect the stale config entry and clean it up automatically
	cmd = exec.Command("go", "run", ".", "install", testBinaryPath)
	cmd.Dir = "/Users/muthuishere/muthu/gitworkspace/lnb-workspace/lnb/cmd/lnb"
	output, err = cmd.CombinedOutput()

	if err != nil {
		t.Errorf("Failed to reinstall binary after stale config cleanup: %v\nOutput: %s", err, string(output))
		return
	}

	// Verify the warning message about cleaning up stale config
	outputStr := string(output)
	if !strings.Contains(outputStr, "Warning: Config shows") || !strings.Contains(outputStr, "doesn't exist. Cleaning up config entry") {
		t.Errorf("Expected warning message about cleaning up stale config entry, got: %s", outputStr)
	}

	// Verify binary was successfully installed
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Errorf("Binary was not reinstalled after stale config cleanup")
	}
}

// TestStaleAliasConfigCleanup tests stale config cleanup for aliases
func TestStaleAliasConfigCleanup(t *testing.T) {
	aliasName := "stale-test-alias"
	aliasCommand := "echo 'test alias command'"
	aliasPath := "/usr/local/bin/" + aliasName

	// Cleanup function
	cleanup := func() {
		os.Remove(aliasPath)
		// Also clean up any config entries
		if cfg, err := config.Load(); err == nil {
			cfg.RemoveEntry(aliasName)
			cfg.Save()
		}
	}
	defer cleanup()

	// Step 1: Create alias normally
	cmd := exec.Command("go", "run", ".", "alias", aliasName, aliasCommand)
	cmd.Dir = "/Users/muthuishere/muthu/gitworkspace/lnb-workspace/lnb/cmd/lnb"
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Skipf("Cannot create test alias (need sudo?): %v\nOutput: %s", err, string(output))
		return
	}

	// Verify it was created
	if _, err := os.Stat(aliasPath); os.IsNotExist(err) {
		t.Errorf("Alias was not created at expected location: %s", aliasPath)
		return
	}

	// Step 2: Manually remove the target alias file (simulating external deletion)
	err = os.Remove(aliasPath)
	if err != nil {
		t.Fatalf("Failed to manually remove alias file: %v", err)
	}

	// Verify target file is gone but config entry still exists
	if _, err := os.Stat(aliasPath); err == nil {
		t.Errorf("Alias file should have been removed")
		return
	}

	// Verify config entry still exists
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if _, exists := cfg.GetEntry(aliasName); !exists {
		t.Errorf("Config entry should still exist after manual alias removal")
		return
	}

	// Step 3: Try to create the same alias again
	// This should detect the stale config entry and clean it up automatically
	cmd = exec.Command("go", "run", ".", "alias", aliasName, aliasCommand)
	cmd.Dir = "/Users/muthuishere/muthu/gitworkspace/lnb-workspace/lnb/cmd/lnb"
	output, err = cmd.CombinedOutput()

	if err != nil {
		t.Errorf("Failed to recreate alias after stale config cleanup: %v\nOutput: %s", err, string(output))
		return
	}

	// Verify the warning message about cleaning up stale config
	outputStr := string(output)
	if !strings.Contains(outputStr, "Warning: Config shows") || !strings.Contains(outputStr, "doesn't exist. Cleaning up config entry") {
		t.Errorf("Expected warning message about cleaning up stale config entry, got: %s", outputStr)
	}

	// Verify alias was successfully recreated
	if _, err := os.Stat(aliasPath); os.IsNotExist(err) {
		t.Errorf("Alias was not recreated after stale config cleanup")
	}
}

// TestConfigIntegrity verifies that the config remains consistent after cleanup operations
func TestConfigIntegrity(t *testing.T) {
	// Create temporary config for testing
	tempDir := t.TempDir()
	tempBinaryPath := filepath.Join(tempDir, "integrity-test")

	// Create test binary
	err := os.WriteFile(tempBinaryPath, []byte("#!/bin/bash\necho 'test'\n"), 0755)
	if err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	binaryName := filepath.Base(tempBinaryPath)
	targetPath := "/usr/local/bin/" + binaryName

	// Cleanup
	defer func() {
		os.Remove(targetPath)
		if cfg, err := config.Load(); err == nil {
			cfg.RemoveEntry(binaryName)
			cfg.Save()
		}
	}()

	// Install binary
	cmd := exec.Command("go", "run", ".", "install", tempBinaryPath)
	cmd.Dir = "/Users/muthuishere/muthu/gitworkspace/lnb-workspace/lnb/cmd/lnb"
	_, err = cmd.CombinedOutput()

	if err != nil {
		t.Skipf("Cannot install test binary (need sudo?): %v", err)
		return
	}

	// Load config and verify entry
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	entry, exists := cfg.GetEntry(binaryName)
	if !exists {
		t.Errorf("Config entry should exist after installation")
		return
	}

	// Verify entry details
	if entry.Name != binaryName {
		t.Errorf("Expected entry name %s, got %s", binaryName, entry.Name)
	}

	if entry.TargetPath != targetPath {
		t.Errorf("Expected target path %s, got %s", targetPath, entry.TargetPath)
	}

	if entry.SourcePath == "" {
		t.Errorf("Source path should not be empty")
	}

	// Verify timestamp is recent
	if time.Since(entry.InstalledAt) > time.Minute {
		t.Errorf("Install timestamp seems too old: %v", entry.InstalledAt)
	}
}
