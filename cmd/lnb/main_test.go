package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const (
	testBinaryContent = `#!/usr/bin/env node
console.log("Hello from Node.js test binary");
console.log("Args:", process.argv.slice(2));
`
	testJavaAppContent = `#!/usr/bin/env node
console.log("Java-like app called with args:", process.argv.slice(2));
console.log("Working directory:", process.cwd());
`
)

// getProjectRoot finds the project root directory by looking for go.mod
func getProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up the directory tree looking for go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find project root (go.mod not found)")
		}
		dir = parent
	}
}

// setupTestEnvironment creates the test environment and returns paths
func setupTestEnvironment(t *testing.T) (projectRoot, testLnbPath, testAssetsDir string, cleanup func()) {
	// Get project root
	root, err := getProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	// Set up paths
	testLnbPath = filepath.Join(root, "dist", "test-lnb")
	testAssetsDir = filepath.Join(root, "dist", "testassets")

	// Check if test-lnb binary exists, if not build it
	if _, err := os.Stat(testLnbPath); os.IsNotExist(err) {
		t.Logf("test-lnb binary not found at %s, building it...", testLnbPath)

		// Create dist directory if it doesn't exist
		if err := os.MkdirAll(filepath.Join(root, "dist"), 0755); err != nil {
			t.Fatalf("Failed to create dist directory: %v", err)
		}

		// Build the test binary
		cmd := exec.Command("go", "build", "-o", "dist/test-lnb", "./cmd/lnb")
		cmd.Dir = root
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to build test-lnb binary: %v\nOutput: %s", err, string(output))
		}

		t.Logf("Successfully built test-lnb binary at %s", testLnbPath)
	}

	// Create test assets directory
	if err := os.MkdirAll(testAssetsDir, 0755); err != nil {
		t.Fatalf("Failed to create test assets directory: %v", err)
	}

	// Return cleanup function
	cleanup = func() {
		os.RemoveAll(testAssetsDir)
		cleanupConfig()
	}

	return root, testLnbPath, testAssetsDir, cleanup
}

// TestLnbIntegration tests all LNB functionality end-to-end
func TestLnbIntegration(t *testing.T) {
	// Set up test environment
	projectRoot, testLnbPath, testAssetsDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create test executables in testassetsdir
	testBinary := filepath.Join(testAssetsDir, "testapp")
	testJavaApp := filepath.Join(testAssetsDir, "javaapp.jar")
	nonExecutable := filepath.Join(testAssetsDir, "notexec.txt")

	// Create executable test binary
	if err := os.WriteFile(testBinary, []byte(testBinaryContent), 0755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	// Create fake Java app (executable script for testing)
	if err := os.WriteFile(testJavaApp, []byte(testJavaAppContent), 0755); err != nil {
		t.Fatalf("Failed to create test Java app: %v", err)
	}

	// Create non-executable file
	if err := os.WriteFile(nonExecutable, []byte("not executable"), 0644); err != nil {
		t.Fatalf("Failed to create non-executable file: %v", err)
	}

	t.Logf("Using project root: %s", projectRoot)
	t.Logf("Using test binary: %s", testLnbPath)
	t.Logf("Using test assets: %s", testAssetsDir)

	// Test cases
	tests := []struct {
		name           string
		args           []string
		expectError    bool
		expectExitCode int
		checkOutput    func(string) bool
	}{
		{
			name:        "help command",
			args:        []string{"help"},
			expectError: false,
			checkOutput: func(output string) bool {
				return strings.Contains(output, "LNB vdev - Link Binary")
			},
		},
		{
			name:        "version command",
			args:        []string{"version"},
			expectError: false,
			checkOutput: func(output string) bool {
				return strings.Contains(output, "LNB vdev")
			},
		},
		{
			name:        "list empty",
			args:        []string{"list"},
			expectError: false,
			checkOutput: func(output string) bool {
				return strings.Contains(output, "No binaries or aliases installed by LNB")
			},
		},
		{
			name:           "install non-existent file",
			args:           []string{"install", "/nonexistent/file"},
			expectError:    true,
			expectExitCode: 1,
		},
		{
			name:           "install non-executable file",
			args:           []string{"install", nonExecutable},
			expectError:    true,
			expectExitCode: 1,
		},
		{
			name:        "install valid executable",
			args:        []string{"install", testBinary},
			expectError: false,
			checkOutput: func(output string) bool {
				return strings.Contains(output, "Successfully installed")
			},
		},
		{
			name:           "install duplicate binary",
			args:           []string{"install", testBinary},
			expectError:    true,
			expectExitCode: 1,
		},
		{
			name:        "list installed binary",
			args:        []string{"list"},
			expectError: false,
			checkOutput: func(output string) bool {
				return strings.Contains(output, "testapp") && strings.Contains(output, "Type:      binary")
			},
		},
		{
			name:        "create alias with relative path",
			args:        []string{"alias", "myapp", fmt.Sprintf("java -jar %s", testJavaApp)},
			expectError: false,
			checkOutput: func(output string) bool {
				return strings.Contains(output, "Created alias: myapp")
			},
		},
		{
			name:           "create duplicate alias",
			args:           []string{"alias", "myapp", "echo hello"},
			expectError:    true,
			expectExitCode: 1,
		},
		{
			name:        "list with binary and alias",
			args:        []string{"list"},
			expectError: false,
			checkOutput: func(output string) bool {
				return strings.Contains(output, "testapp") &&
					strings.Contains(output, "myapp") &&
					strings.Contains(output, "Type:      binary") &&
					strings.Contains(output, "Type:      alias")
			},
		},
		{
			name:           "create alias with invalid command",
			args:           []string{"alias", "badcmd", "/nonexistent/command"},
			expectError:    true,
			expectExitCode: 1,
		},
		{
			name:        "remove alias",
			args:        []string{"unalias", "myapp"},
			expectError: false,
			checkOutput: func(output string) bool {
				return strings.Contains(output, "Removed alias: myapp")
			},
		},
		{
			name:           "remove non-existent alias",
			args:           []string{"unalias", "nonexistent"},
			expectError:    true,
			expectExitCode: 1,
		},
		{
			name:        "remove binary",
			args:        []string{"remove", "testapp"},
			expectError: false,
			checkOutput: func(output string) bool {
				return strings.Contains(output, "Removed:")
			},
		},
		{
			name:           "remove non-existent binary",
			args:           []string{"remove", "testapp"},
			expectError:    true,
			expectExitCode: 1,
		},
		{
			name:        "list empty after cleanup",
			args:        []string{"list"},
			expectError: false,
			checkOutput: func(output string) bool {
				return strings.Contains(output, "No binaries or aliases installed by LNB")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up config only before the first install test to start fresh
			if tt.name == "install valid executable" {
				cleanupConfig()
			}

			cmd := exec.Command(testLnbPath, tt.args...)
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but command succeeded. Output: %s", outputStr)
				} else if exitErr, ok := err.(*exec.ExitError); ok && tt.expectExitCode != 0 {
					if exitErr.ExitCode() != tt.expectExitCode {
						t.Errorf("Expected exit code %d, got %d. Output: %s", tt.expectExitCode, exitErr.ExitCode(), outputStr)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected success but got error: %v. Output: %s", err, outputStr)
				}
			}

			if tt.checkOutput != nil {
				if !tt.checkOutput(outputStr) {
					t.Errorf("Output check failed. Output: %s", outputStr)
				}
			}

			t.Logf("Command: %v\nOutput: %s\n", tt.args, outputStr)
		})
	}
}

// TestLnbAliasWithJavaApp tests creating and using a Java app alias
func TestLnbAliasWithJavaApp(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Java app test on Windows for now")
	}

	// Set up test environment
	_, testLnbPath, testAssetsDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a fake Java app (actually a Node.js script for testing)
	javaApp := filepath.Join(testAssetsDir, "myapp.jar")
	javaAppContent := `#!/usr/bin/env node
console.log("Java-like app executed with args:", process.argv.slice(2));
console.log("Working directory:", process.cwd());
`
	if err := os.WriteFile(javaApp, []byte(javaAppContent), 0755); err != nil {
		t.Fatalf("Failed to create Java app: %v", err)
	}

	// Test creating alias with relative path
	cmd := exec.Command(testLnbPath, "alias", "javatest", fmt.Sprintf("node %s", javaApp))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to create alias: %v\nOutput: %s", err, string(output))
	}

	if !strings.Contains(string(output), "Created alias: javatest") {
		t.Errorf("Expected alias creation message, got: %s", string(output))
	}

	// Verify alias appears in list
	cmd = exec.Command(testLnbPath, "list")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to list: %v\nOutput: %s", err, string(output))
	}

	if !strings.Contains(string(output), "javatest") || !strings.Contains(string(output), "Type:      alias") {
		t.Errorf("Alias not found in list: %s", string(output))
	}

	// Clean up
	cmd = exec.Command(testLnbPath, "unalias", "javatest")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to remove alias: %v\nOutput: %s", err, string(output))
	}
}

// TestLnbPathConversion tests that relative paths are converted to absolute paths
func TestLnbPathConversion(t *testing.T) {
	// Set up test environment
	_, testLnbPath, testAssetsDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a script in the test assets directory
	script := filepath.Join(testAssetsDir, "myscript.js")
	scriptContent := `#!/usr/bin/env node
console.log("Script executed from:", __dirname);
console.log("Args:", process.argv.slice(2));
`
	if err := os.WriteFile(script, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to create script: %v", err)
	}

	// Change to test assets directory to test relative paths
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(testAssetsDir)

	// Create alias with relative path
	cmd := exec.Command(testLnbPath, "alias", "relscript", "./myscript.js hello world")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to create alias with relative path: %v\nOutput: %s", err, string(output))
	}

	// Verify the alias was created and shows absolute path
	cmd = exec.Command(testLnbPath, "list")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to list: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "relscript") {
		t.Errorf("Alias not found in list: %s", outputStr)
	}

	// The list output should show the original command
	if !strings.Contains(outputStr, "./myscript.js") {
		t.Errorf("Expected to see original relative path in list output: %s", outputStr)
	}

	// But the actual script file should contain the absolute path
	// Let's verify by reading the script file
	scriptPath := "/usr/local/bin/relscript"
	if scriptContent, err := os.ReadFile(scriptPath); err == nil {
		scriptStr := string(scriptContent)
		if !strings.Contains(scriptStr, testAssetsDir) {
			t.Errorf("Expected script file to contain absolute path, but got: %s", scriptStr)
		}
		if strings.Contains(scriptStr, "./myscript.js") {
			t.Errorf("Expected script file to NOT contain relative path, but got: %s", scriptStr)
		}
	} else {
		t.Errorf("Failed to read script file: %v", err)
	}

	// Clean up
	cmd = exec.Command(testLnbPath, "unalias", "relscript")
	cmd.Run()
}

// TestLnbInstallCommand tests the install command explicitly
func TestLnbInstallCommand(t *testing.T) {
	// Set up test environment
	_, testLnbPath, testAssetsDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create test executable
	testBinary := filepath.Join(testAssetsDir, "installtest")
	if err := os.WriteFile(testBinary, []byte("#!/usr/bin/env node\nconsole.log('hello');\n"), 0755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	// Test install command
	cmd := exec.Command(testLnbPath, "install", testBinary)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run install command: %v\nOutput: %s", err, string(output))
	}

	if !strings.Contains(string(output), "Successfully installed") {
		t.Errorf("Expected install message, got: %s", string(output))
	}

	// Verify it's in the list
	cmd = exec.Command(testLnbPath, "list")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to list: %v\nOutput: %s", err, string(output))
	}

	if !strings.Contains(string(output), "installtest") {
		t.Errorf("Binary not found in list after install: %s", string(output))
	}

	// Clean up
	cmd = exec.Command(testLnbPath, "remove", "installtest")
	cmd.Run()
}

// cleanupConfig removes the LNB config file to start fresh
func cleanupConfig() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	configPath := filepath.Join(homeDir, ".lnb", "config.json")
	os.Remove(configPath)
}

// TestLnbNoArgs tests the help is shown when no arguments are provided
func TestLnbNoArgs(t *testing.T) {
	// Set up test environment
	_, testLnbPath, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	cmd := exec.Command(testLnbPath)
	output, err := cmd.CombinedOutput()

	// Should not error, should just show help
	if err != nil {
		t.Errorf("Expected no error when running without args, got: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "LNB") && !strings.Contains(outputStr, "Link Binary") {
		t.Errorf("Expected help message when no args provided, got: %s", outputStr)
	}
}

// TestLnbSmartInstall tests that file paths are automatically treated as install commands
func TestLnbSmartInstall(t *testing.T) {
	// Set up test environment
	_, testLnbPath, testAssetsDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Clean up any existing config
	cleanupConfig()

	// Create a test binary
	smartBinary := filepath.Join(testAssetsDir, "smartbinary")
	if err := os.WriteFile(smartBinary, []byte(testBinaryContent), 0755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	testCases := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "absolute_path_install",
			args:     []string{smartBinary},
			wantErr:  false,
			contains: "Successfully installed",
		},
		{
			name:     "relative_path_install",
			args:     []string{"./testassets/smartbinary"},
			wantErr:  false,
			contains: "Successfully installed",
		},
		{
			name:     "explicit_install_still_works",
			args:     []string{"install", smartBinary},
			wantErr:  false,
			contains: "Successfully installed",
		},
		{
			name:     "non_existent_file_errors",
			args:     []string{"/nonexistent/file"},
			wantErr:  true,
			contains: "does not exist",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clean up before each test
			cleanupConfig()

			// Also remove any existing binary files to ensure clean state
			exec.Command(testLnbPath, "remove", "smartbinary").Run()

			cmd := exec.Command(testLnbPath, tc.args...)
			cmd.Dir = filepath.Dir(testAssetsDir) // Set working directory for relative paths
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none. Output: %s", outputStr)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v. Output: %s", err, outputStr)
				}
			}

			if !strings.Contains(outputStr, tc.contains) {
				t.Errorf("Expected output to contain '%s', got: %s", tc.contains, outputStr)
			}

			t.Logf("%s - Command: %v\n\tOutput: %s", tc.name, tc.args, outputStr)

			// Clean up after each test too
			exec.Command(testLnbPath, "remove", "smartbinary").Run()
		})
	}
}
