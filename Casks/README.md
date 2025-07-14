# LNB Homebrew Cask

This directory contains the Homebrew Cask for LNB, which is automatically generated and updated by GoReleaser during the release process.

## Files

- `lnb.rb` - The Homebrew Cask definition (automatically generated during releases)
- `README.md` - This documentation file

## Installation

To install LNB using Homebrew Cask directly from this repository:

```bash
# Add this repository as a tap
brew tap muthuishere/lnb https://github.com/muthuishere/lnb

# Install LNB
brew install --cask lnb
```

Or install directly without adding the tap:

```bash
brew install --cask muthuishere/lnb/lnb
```

## Usage

After installation, the `lnb` command will be available in your PATH:

```bash
lnb help
lnb version
lnb install ./mybinary
lnb list
```

## Updating

To update to the latest version:

```bash
brew upgrade --cask lnb
```

## Uninstalling

To uninstall:

```bash
brew uninstall --cask lnb
```

## Development

The cask definition (`lnb.rb`) in this directory is automatically generated during the release process. The current file serves as a template to show the structure. Do not edit it manually as your changes will be overwritten during releases.

### Template Structure

The cask includes:
- Version and checksum information (auto-filled by GoReleaser)
- Download URL pointing to GitHub releases
- Binary installation
- Post-install quarantine removal for macOS security
- Zap directive for clean uninstallation
