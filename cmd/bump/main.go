package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("🔧 LNB Version Bump & Release Tool")
		fmt.Println()
		fmt.Println("Usage: bump <major|minor|patch>")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  bump patch   # 0.2.7 -> 0.2.8")
		fmt.Println("  bump minor   # 0.2.7 -> 0.3.0")
		fmt.Println("  bump major   # 0.2.7 -> 1.0.0")
		fmt.Println()
		fmt.Println("This command will:")
		fmt.Println("  1. Update versions.txt")
		fmt.Println("  2. Commit changes")
		fmt.Println("  3. Create git tag (local)")
		fmt.Println()
		fmt.Println("After bump, choose release method:")
		fmt.Println("  • task release-via-git   (push to GitHub Actions)")
		fmt.Println("  • task release-via-local (release locally)")
		os.Exit(1)
	}

	bumpType := strings.ToLower(os.Args[1])
	if bumpType != "major" && bumpType != "minor" && bumpType != "patch" {
		fmt.Printf("Error: Invalid bump type '%s'. Use major, minor, or patch.\n", bumpType)
		os.Exit(1)
	}

	// Read current version from versions.txt
	versionBytes, err := ioutil.ReadFile("versions.txt")
	if err != nil {
		fmt.Printf("Error reading versions.txt: %v\n", err)
		os.Exit(1)
	}

	currentVersion := strings.TrimSpace(string(versionBytes))
	fmt.Printf("Current version: %s\n", currentVersion)

	// Parse version (expecting format: major.minor.patch)
	parts := strings.Split(currentVersion, ".")
	if len(parts) != 3 {
		fmt.Printf("Error: Invalid version format '%s'. Expected format: major.minor.patch\n", currentVersion)
		os.Exit(1)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		fmt.Printf("Error parsing major version '%s': %v\n", parts[0], err)
		os.Exit(1)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		fmt.Printf("Error parsing minor version '%s': %v\n", parts[1], err)
		os.Exit(1)
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		fmt.Printf("Error parsing patch version '%s': %v\n", parts[2], err)
		os.Exit(1)
	}

	// Bump version based on type
	switch bumpType {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	}

	newVersion := fmt.Sprintf("%d.%d.%d", major, minor, patch)
	fmt.Printf("🎯 Bumping version from %s to %s\n", currentVersion, newVersion)
	fmt.Println()

	// Check git status first
	if !isGitClean() {
		fmt.Println("❌ Error: Git working directory is not clean")
		fmt.Println("   Please commit or stash your changes first (except versions.txt)")
		os.Exit(1)
	}

	// Step 1: Write new version to versions.txt
	fmt.Println("📝 Step 1: Updating versions.txt...")
	err = ioutil.WriteFile("versions.txt", []byte(newVersion), 0644)
	if err != nil {
		fmt.Printf("❌ Error writing versions.txt: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Updated versions.txt to %s\n", newVersion)

	// Step 2: Commit changes
	fmt.Println("💾 Step 2: Committing changes...")
	err = commitChanges(newVersion)
	if err != nil {
		fmt.Printf("❌ Error committing changes: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Committed version bump to %s\n", newVersion)

	// Step 3: Create git tag
	fmt.Println("🏷️  Step 3: Creating git tag...")
	tagName := "v" + newVersion
	err = createGitTag(tagName, newVersion)
	if err != nil {
		fmt.Printf("❌ Error creating git tag: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Created git tag %s\n", tagName)

	fmt.Println()
	fmt.Println("🎉 Version bump complete!")
	fmt.Printf("📦 Version: %s -> %s\n", currentVersion, newVersion)
	fmt.Printf("🏷️  Tag: %s (local)\n", tagName)
	fmt.Println()
	fmt.Println("🚀 Next Steps:")
	fmt.Println("   • Push & Release via GitHub Actions: task release-via-git")
	fmt.Println("   • Release locally (faster):          task release-via-local")
	fmt.Println()
	fmt.Printf("   Git status: ✅ Ready to push tag %s\n", tagName)
}

// isGitClean checks if the git working directory is clean (no uncommitted changes)
// except for versions.txt which is expected to change during bump
func isGitClean() bool {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Warning: Failed to check git status: %v\n", err)
		return false
	}

	// Parse git status output - each line represents a changed file
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Extract filename from git status line (format: "?? filename" or "M  filename")
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		filename := parts[1]
		// Allow only versions.txt to be modified during bump
		if filename != "versions.txt" {
			fmt.Printf("   Uncommitted changes in: %s\n", filename)
			return false
		}
	}

	return true
}

// commitChanges commits the version bump changes to git
func commitChanges(version string) error {
	// Add versions.txt to staging
	cmd := exec.Command("git", "add", "versions.txt")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	// Commit with version bump message
	commitMsg := fmt.Sprintf("Bump version to %s", version)
	cmd = exec.Command("git", "commit", "-m", commitMsg)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	return nil
}

// createGitTag creates an annotated git tag for the version
func createGitTag(tagName string, version string) error {
	tagMsg := fmt.Sprintf("Release %s", version)
	cmd := exec.Command("git", "tag", "-a", tagName, "-m", tagMsg)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create git tag: %w", err)
	}

	return nil
}

// pushToRemote pushes commits and tags to the remote repository
func pushToRemote(tagName string) error {
	// Push commits first
	cmd := exec.Command("git", "push")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push commits: %w", err)
	}

	// Push the tag
	cmd = exec.Command("git", "push", "origin", tagName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push tag: %w", err)
	}

	return nil
}
