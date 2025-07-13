package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

type FormulaData struct {
	Version     string
	Description string
	Homepage    string
	License     string
}

const formulaTemplate = `class Lnb < Formula
  desc "{{.Description}}"
  homepage "{{.Homepage}}"
  version "{{.Version}}"
  license "{{.License}}"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/muthuishere/lnb/releases/download/v{{.Version}}/lnb_{{.Version}}_Darwin_arm64.zip"
      sha256 "sha256-will-be-updated-by-goreleaser"
    else
      url "https://github.com/muthuishere/lnb/releases/download/v{{.Version}}/lnb_{{.Version}}_Darwin_x86_64.zip"
      sha256 "sha256-will-be-updated-by-goreleaser"
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/muthuishere/lnb/releases/download/v{{.Version}}/lnb_{{.Version}}_Linux_arm64.zip"
      sha256 "sha256-will-be-updated-by-goreleaser"
    else
      url "https://github.com/muthuishere/lnb/releases/download/v{{.Version}}/lnb_{{.Version}}_Linux_x86_64.zip"
      sha256 "sha256-will-be-updated-by-goreleaser"
    end
  end

  def install
    bin.install "lnb"
  end

  test do
    system "#{bin}/lnb", "--version"
    assert_match "LNB v", shell_output("#{bin}/lnb --version")
  end
end
`

func main() {
	var (
		version     = flag.String("version", "", "Version to update the formula to")
		description = flag.String("description", "A cross-platform utility that makes command-line tools accessible from anywhere", "Description for the formula")
		homepage    = flag.String("homepage", "https://github.com/muthuishere/lnb", "Homepage URL")
		license     = flag.String("license", "MIT", "License")
		formulaPath = flag.String("formula-path", "", "Path to the formula file")
	)
	flag.Parse()

	if *version == "" {
		log.Fatal("Version is required. Use -version flag.")
	}

	if *formulaPath == "" {
		log.Fatal("Formula path is required. Use -formula-path flag.")
	}

	// Ensure the directory exists
	formulaDir := filepath.Dir(*formulaPath)
	if err := os.MkdirAll(formulaDir, 0755); err != nil {
		log.Fatalf("Failed to create formula directory: %v", err)
	}

	// Prepare template data
	data := FormulaData{
		Version:     *version,
		Description: *description,
		Homepage:    *homepage,
		License:     *license,
	}

	// Parse and execute template
	tmpl, err := template.New("formula").Parse(formulaTemplate)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	// Create or overwrite the formula file
	file, err := os.Create(*formulaPath)
	if err != nil {
		log.Fatalf("Failed to create formula file: %v", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	fmt.Printf("Successfully updated formula at %s with version %s\n", *formulaPath, *version)
}
