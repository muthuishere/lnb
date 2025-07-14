# LNB Scoop Bucket

This directory contains the Scoop bucket manifest for LNB, which is automatically generated and updated by GoReleaser during the release process.

## Files

- `bucket/lnb.json` - The Scoop manifest (automatically generated during releases)
- `README.md` - This documentation file

## Installation

To install LNB using Scoop directly from this repository:

```powershell
# Add this repository as a bucket
scoop bucket add lnb https://github.com/muthuishere/lnb

# Install LNB
scoop install lnb
```

Or install directly without adding the bucket:

```powershell
scoop install https://raw.githubusercontent.com/muthuishere/lnb/main/scoop/bucket/lnb.json
```

## Usage

After installation, the `lnb` command will be available in your PATH:

```cmd
lnb help
lnb version
lnb install ./mybinary.exe
lnb list
```

## Updating

To update to the latest version:

```powershell
scoop update lnb
```

## Uninstalling

To uninstall:

```powershell
scoop uninstall lnb
```

## Development

The manifest (`bucket/lnb.json`) in this directory is automatically generated during the release process. Do not edit it manually as your changes will be overwritten during releases.

### Manifest Structure

The Scoop manifest includes:
- Version information (auto-filled by GoReleaser)
- Download URL pointing to GitHub releases
- Hash verification for security
- Binary installation configuration
- License and description metadata
