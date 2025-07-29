package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		showHelp()
		return
	}

	// Check for --local flag
	localOnly := false
	if len(os.Args) > 1 && os.Args[1] == "--local" {
		localOnly = true
	}

	fmt.Println("🚀 LNB Release Manager")
	if localOnly {
		fmt.Println("   📍 Local mode: will not push to remote")
	}
	fmt.Println()

	// Check if we're in a git repository
	if !isGitRepo() {
		fmt.Println("❌ Error: Not in a git repository")
		os.Exit(1)
	}

	// Check git status - if dirty, ask user to commit
	if isGitDirty() {
		fmt.Println("⚠️  Git working directory is dirty (uncommitted changes)")
		if !askYesNo("Do you want to commit all changes first? (y/N)") {
			fmt.Println("❌ Please commit your changes before releasing")
			os.Exit(1)
		}

		if err := commitAllChanges(); err != nil {
			fmt.Printf("❌ Error committing changes: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Changes committed")
	}

	// Read current version from versions.txt
	currentVersion, err := readVersionFile()
	if err != nil {
		fmt.Printf("❌ Error reading versions.txt: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📦 Current version: %s\n", currentVersion)

	// Get the last tagged version from git
	lastTaggedVersion, err := getLastTaggedVersion()
	if err != nil {
		fmt.Printf("⚠️  Could not get last tagged version: %v\n", err)
		fmt.Println("   This might be the first release")
	} else {
		fmt.Printf("🏷️  Last tagged version: %s\n", lastTaggedVersion)

		// Check if version has changed
		if currentVersion == lastTaggedVersion {
			fmt.Printf("❌ Version %s is the same as the last tagged version\n", currentVersion)
			fmt.Println("   Use 'go run ./cmd/bump patch|minor|major' to bump the version first")
			os.Exit(1)
		}
	}
	fmt.Println()
	fmt.Printf("🎯 Ready to release version %s\n", currentVersion)

	var confirmMessage string
	if localOnly {
		confirmMessage = "Create local git tag (no push)? (y/N)"
	} else {
		confirmMessage = "Create git tag and trigger release? (y/N)"
	}

	if !askYesNo(confirmMessage) {
		fmt.Println("❌ Release cancelled")
		os.Exit(1)
	}

	// Create git tag
	tagName := "v" + currentVersion
	if err := createGitTag(tagName, currentVersion); err != nil {
		fmt.Printf("❌ Error creating git tag: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Created git tag: %s\n", tagName)

	if localOnly {
		fmt.Println()
		fmt.Println("✅ Local release complete!")
		fmt.Printf("📦 Git tag %s created locally\n", tagName)
		fmt.Println("💡 To trigger the release workflow later, push the tag:")
		fmt.Printf("   git push origin %s\n", tagName)
	} else {
		// Push the tag to trigger release workflow
		if err := pushGitTag(tagName); err != nil {
			fmt.Printf("❌ Error pushing git tag: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Pushed git tag: %s\n", tagName)
		fmt.Println()
		fmt.Println("🎉 Release triggered!")
		fmt.Println("📋 Check GitHub Actions for the release build:")

		// Get repository URL for convenience
		if repoURL, err := getRepositoryURL(); err == nil {
			fmt.Printf("🔗 %s/actions\n", repoURL)
		}
	}
}

func showHelp() {
	fmt.Println("LNB Release Manager")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  go run ./cmd/release [--local]")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  --local    Create git tag locally without pushing to remote")
	fmt.Println()
	fmt.Println("DESCRIPTION:")
	fmt.Println("  Manages the release process for LNB:")
	fmt.Println("  1. Checks if git working directory is clean")
	fmt.Println("  2. Reads version from versions.txt")
	fmt.Println("  3. Compares with last tagged version")
	fmt.Println("  4. Creates git tag (and optionally pushes to trigger release workflow)")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  go run ./cmd/release          # Create tag and push to trigger release")
	fmt.Println("  go run ./cmd/release --local  # Create tag locally only")
	fmt.Println()
	fmt.Println("PREREQUISITES:")
	fmt.Println("  - Clean git working directory (or will prompt to commit)")
	fmt.Println("  - Version in versions.txt must be different from last tag")
	fmt.Println("  - Use 'go run ./cmd/bump patch|minor|major' to bump version")
}

func isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

func isGitDirty() bool {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return true // Assume dirty if we can't check
	}
	return len(strings.TrimSpace(string(output))) > 0
}

func askYesNo(question string) bool {
	fmt.Print(question + " ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func commitAllChanges() error {
	// Add all changes
	if err := exec.Command("git", "add", ".").Run(); err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}

	// Read current version for commit message
	version, err := readVersionFile()
	if err != nil {
		return fmt.Errorf("failed to read version: %w", err)
	}

	// Commit with version message
	commitMsg := fmt.Sprintf("Prepare release v%s\n\n- Update version to %s", version, version)
	if err := exec.Command("git", "commit", "-m", commitMsg).Run(); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	return nil
}

func readVersionFile() (string, error) {
	versionBytes, err := ioutil.ReadFile("versions.txt")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(versionBytes)), nil
}

func getLastTaggedVersion() (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	tag := strings.TrimSpace(string(output))
	// Remove 'v' prefix if present
	if strings.HasPrefix(tag, "v") {
		return tag[1:], nil
	}
	return tag, nil
}

func createGitTag(tagName, version string) error {
	message := fmt.Sprintf("Release %s", tagName)
	return exec.Command("git", "tag", "-a", tagName, "-m", message).Run()
}

func pushGitTag(tagName string) error {
	return exec.Command("git", "push", "origin", tagName).Run()
}

func getRepositoryURL() (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	url := strings.TrimSpace(string(output))
	// Convert SSH URL to HTTPS for browser
	if strings.HasPrefix(url, "git@github.com:") {
		url = strings.Replace(url, "git@github.com:", "https://github.com/", 1)
		url = strings.TrimSuffix(url, ".git")
	}

	return url, nil
}
