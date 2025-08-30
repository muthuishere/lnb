package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestRemoveSafetyCheck tests that LNB refuses to remove files it didn't install
func TestRemoveSafetyCheck(t *testing.T) {
	// Create a temporary dummy binary in /usr/local/bin (or equivalent)
	dummyName := "lnb-test-dummy"
	dummyPath := "/usr/local/bin/" + dummyName

	// Create a simple dummy script
	dummyContent := `#!/bin/bash
echo "This is a dummy file not installed by LNB"
`

	// Write the dummy file
	err := os.WriteFile(dummyPath, []byte(dummyContent), 0755)
	if err != nil {
		t.Skipf("Cannot create test file in /usr/local/bin (need sudo?): %v", err)
		return
	}

	// Ensure cleanup even if test fails
	defer func() {
		os.Remove(dummyPath)
	}()

	// Try to remove it with lnb remove - this should fail
	cmd := exec.Command("go", "run", ".", "remove", dummyName)
	cmd.Dir = "/Users/muthuishere/muthu/gitworkspace/lnb-workspace/lnb/cmd/lnb"
	output, err := cmd.CombinedOutput()

	// This should fail because LNB didn't install it
	if err == nil {
		t.Errorf("Expected lnb remove to fail for file not installed by LNB, but it succeeded")
		t.Errorf("Output: %s", string(output))
	}

	// Check that the error message is appropriate
	outputStr := string(output)
	if !strings.Contains(outputStr, "was not installed by LNB") {
		t.Errorf("Expected error message about file not installed by LNB, got: %s", outputStr)
	}

	// Verify the dummy file still exists (wasn't removed)
	if _, err := os.Stat(dummyPath); os.IsNotExist(err) {
		t.Errorf("Dummy file was incorrectly removed by LNB")
	}
}

// TestInstallThenRemoveSafety tests the complete install/remove cycle works correctly
func TestInstallThenRemoveSafety(t *testing.T) {
	// Create a temporary test binary
	tempDir := t.TempDir()
	testBinaryPath := filepath.Join(tempDir, "test-binary")

	// Create a simple test script
	testContent := `#!/bin/bash
echo "This is a test binary"
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
	}
	defer cleanup()

	// Install the binary with LNB
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

	// Now remove it with LNB - this should work
	cmd = exec.Command("go", "run", ".", "remove", binaryName)
	cmd.Dir = "/Users/muthuishere/muthu/gitworkspace/lnb-workspace/lnb/cmd/lnb"
	output, err = cmd.CombinedOutput()

	if err != nil {
		t.Errorf("Failed to remove binary that was installed by LNB: %v\nOutput: %s", err, string(output))
	}

	// Verify it was removed
	if _, err := os.Stat(targetPath); err == nil {
		t.Errorf("Binary still exists after LNB remove: %s", targetPath)
	}
}

// TestAliasSafetyCheck tests that LNB refuses to remove aliases it didn't create
func TestAliasSafetyCheck(t *testing.T) {
	// Create a dummy alias script manually (not via LNB)
	aliasName := "lnb-test-dummy-alias"
	aliasPath := "/usr/local/bin/" + aliasName

	// Create a simple dummy alias script
	aliasContent := `#!/bin/bash
echo "This is a dummy alias not created by LNB"
`

	// Write the dummy alias
	err := os.WriteFile(aliasPath, []byte(aliasContent), 0755)
	if err != nil {
		t.Skipf("Cannot create test alias in /usr/local/bin (need sudo?): %v", err)
		return
	}

	// Ensure cleanup even if test fails
	defer func() {
		os.Remove(aliasPath)
	}()

	// Try to remove it with lnb unalias - this should fail
	cmd := exec.Command("go", "run", ".", "unalias", aliasName)
	cmd.Dir = "/Users/muthuishere/muthu/gitworkspace/lnb-workspace/lnb/cmd/lnb"
	output, err := cmd.CombinedOutput()

	// This should fail because LNB didn't create it
	if err == nil {
		t.Errorf("Expected lnb unalias to fail for alias not created by LNB, but it succeeded")
		t.Errorf("Output: %s", string(output))
	}

	// Check that the error message is appropriate
	outputStr := string(output)
	if !strings.Contains(outputStr, "was not installed by LNB") {
		t.Errorf("Expected error message about alias not installed by LNB, got: %s", outputStr)
	}

	// Verify the dummy alias still exists (wasn't removed)
	if _, err := os.Stat(aliasPath); os.IsNotExist(err) {
		t.Errorf("Dummy alias was incorrectly removed by LNB")
	}
}
