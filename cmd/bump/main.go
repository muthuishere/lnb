package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: bump <major|minor|patch>")
		fmt.Println("Examples:")
		fmt.Println("  bump patch   # 0.2.4 -> 0.2.5")
		fmt.Println("  bump minor   # 0.2.4 -> 0.3.0")
		fmt.Println("  bump major   # 0.2.4 -> 1.0.0")
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
	fmt.Printf("New version: %s\n", newVersion)

	// Write new version to versions.txt
	err = ioutil.WriteFile("versions.txt", []byte(newVersion), 0644)
	if err != nil {
		fmt.Printf("Error writing versions.txt: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Successfully bumped version from %s to %s\n", currentVersion, newVersion)
	fmt.Println("📝 Run 'task setup-release' to prepare the release")
}
