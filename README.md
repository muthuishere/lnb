# LNB - Link Binary

A simple cross-platform utility to link binaries to your PATH.

## Overview

LNB makes command-line tools accessible from anywhere by creating symbolic links or wrapper scripts in your system's PATH. It works across Linux, macOS, and Windows with a consistent interface.

## Features

- **Cross-platform**: Works on Linux, macOS, and Windows
- **Simple interface**: Just two commands - install and remove
- **No configuration needed**: Automatically detects your OS and does the right thing

## Installation

### From Package Managers

#### Homebrew (macOS/Linux)
```bash
brew install muthuishere/tap/lnb
```

#### Chocolatey (Windows)
```powershell
choco install lnb
```

#### Scoop (Windows)
```powershell
scoop bucket add muthuishere https://github.com/muthuishere/scoop-bucket-lnb.git
scoop install lnb
```

#### Snap (Linux)
```bash
snap install lnb
```

### Manual Installation

1. Download the binary for your platform from the [releases page](https://github.com/muthuishere/lnb/releases)
2. Use LNB to install itself:
```bash
./lnb-[platform] ./lnb-[platform] install
```

## Usage

### Installing a binary
```bash
lnb path/to/binary install
```

This will make `binary` available in your PATH. On Linux and macOS, it creates a symbolic link in `/usr/local/bin`. On Windows, it creates a wrapper script in `%USERPROFILE%\bin`.

### Removing a binary
```bash
lnb path/to/binary remove
```

This will remove the binary from your PATH.

## Examples

### Make a development tool available system-wide
```bash
lnb ~/tools/awesome-cli install
# Now you can use awesome-cli from anywhere
```

### Remove a tool when you no longer need it
```bash
lnb ~/tools/awesome-cli remove
```

## How it works

- **Linux/macOS**: Creates symbolic links in `/usr/local/bin`
- **Windows**: Creates `.cmd` wrapper scripts in `%USERPROFILE%\bin`

## Building from source

Requirements:
- Go 1.16+
- [Task](https://taskfile.dev) (optional, for running tasks)

```bash
# Clone the repository
git clone https://github.com/muthuishere/lnb.git
cd lnb

# Build for your platform
task build:local

# Or build for all platforms
task build
```

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

## License

MIT
