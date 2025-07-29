# LNB – Link Binary

## The Problem

Have you ever spent too much time tweaking your PATH, creating symlinks, or writing wrapper scripts just so your custom CLI or downloaded binary works from anywhere? Yeah, it sucks. You shouldn't have to jump through hoops every time you build or grab a new tool.

**Why this matters**: Pointless setup steps slow you down. You want to focus on coding or using the tool, not messing with config files.

## The Solution

**LNB** fixes that by giving you a one-liner to make any binary globally accessible. On Linux/macOS, it creates symlinks in `/usr/local/bin`. On Windows, it creates wrapper scripts in `%USERPROFILE%\bin`. Plus, it can create command aliases for complex commands. No extra config, no guessing which folder to use—LNB handles it all and tracks what you've installed.

**Why I like it**: It's dead simple. You don't need to learn some complex packaging system or remember different commands per OS.

## Features

- **Smart command detection**: Just pass a file path - LNB automatically installs it
- **Cross-platform**: Same command on Linux, macOS, and Windows  
- **Binary installation**: Install any executable to your PATH with one command
- **Command aliases**: Create shortcuts for complex commands (e.g., `java -jar myapp.jar`)
- **Installation tracking**: Keep track of what you've installed with `lnb list`
- **Path conversion**: Automatically converts relative paths to absolute paths
- **Zero config**: LNB auto-detects your OS and does the right thing  
- **Fast**: No dependencies beyond the Go binary itself

## Installation

### From Package Managers

I recommend using a package manager if you can—setup is instant.

- **NPM (Cross-platform)**  
  ```bash
  # Install globally via NPM
  npm install -g lnb
  ```

  *Why NPM?* Available everywhere Node.js is installed, excellent global reach, and automatic platform detection.

- **Homebrew (macOS/Linux)**  
  ```bash
  # Add the tap (separate repository)
  brew tap muthuishere/lnb https://github.com/muthuishere/homebrew-lnb
  
  # Install LNB
  brew install --cask lnb
  ```

  *Why Homebrew?* Everyone on macOS/Linux already has it, and updates are a breeze.

- **Scoop (Windows)**  
  ```powershell
  # Add the bucket (separate repository)
  scoop bucket add lnb https://github.com/muthuishere/scoop-lnb
  
  # Install LNB
  scoop install lnb
  ```

  *Why Scoop?* If you're already using Scoop, it feels right to keep things consistent.

### Manual Installation

If you prefer to download directly:

1. Grab the latest binary from the [releases page](https://github.com/muthuishere/lnb/releases).
2. Run LNB to install itself globally:

   ```bash
   ./lnb-[platform] ./lnb-[platform]
   ```

   This makes `lnb` available system-wide so you can use LNB to manage other tools.

## Usage

### Smart Installation (Recommended)

LNB automatically detects when you pass a file path and installs it:

```bash
# All of these work the same way - LNB detects the file path
lnb ./my-binary                    # Relative path
lnb /path/to/my-binary             # Absolute path  
lnb ~/Downloads/some-tool          # Home directory path
```

**Why it's neat**: No need to remember the `install` command - just point LNB at your binary and it works.

### Explicit Installation (Also Works)

```bash
# Traditional explicit syntax still works
lnb install ./my-binary
lnb install /path/to/my-binary
```

### What Happens During Installation

- On **Linux/macOS**, this creates:
  ```
  /usr/local/bin/<binary-name> -> /absolute/path/to/binary
  ```
- On **Windows**, this creates:
  ```
  %USERPROFILE%\bin\<binary-name>.cmd
  ```
  which wraps your `.exe` or any executable. *(Make sure `%USERPROFILE%\bin` is in your PATH.)*

**Path Intelligence**: LNB automatically converts relative paths (like `./mybinary`) to absolute paths so your installations work from anywhere.

### Remove a Binary

```bash
lnb remove binary-name
```

Deletes the symlink (or `.cmd` wrapper on Windows), but leaves your original file untouched.

### Create Command Aliases

```bash
# Create an alias for a complex command
lnb alias myapp "java -jar ./myapp.jar"
lnb alias serve "python -m http.server 8080"  
lnb alias deploy "docker run --rm -v $(pwd):/workspace deploy-tool"
lnb alias build "npm run build && npm run test"
```

- **Linux/macOS**: Creates a shell script at `/usr/local/bin/<alias-name>`
- **Windows**: Creates a batch file at `%USERPROFILE%\bin\<alias-name>.bat`

**Smart path handling**: When creating aliases, LNB converts relative paths in your commands to absolute paths so they work from anywhere.

### Remove Aliases

```bash
lnb unalias myapp
```

### List Installed Items

```bash
lnb list
```

Shows all binaries and aliases installed by LNB, including:
- Installation date
- Type (binary or alias)
- Source path (for binaries) or command (for aliases)
- Target location

### Help and Version

```bash
lnb help        # Show detailed help
lnb version     # Show version information
lnb             # Show help (when no arguments provided)
```

## Examples

### Quick Binary Installation

```bash
# Build your Go project
go build -o mytool ./cmd/mytool

# Install it instantly - LNB detects it's a file path
lnb ./mytool

# Now "mytool" runs from anywhere
mytool --help
```

### Java Application Alias

```bash
# Instead of remembering the full command
java -jar /long/path/to/myapp.jar --config production

# Create a simple alias
lnb alias myapp "java -jar /long/path/to/myapp.jar --config production"

# Now just run
myapp
```

### Development Workflow

```bash
# Install your dev tool
lnb ./bin/dev-server

# Create shortcuts for common tasks  
lnb alias dev "npm run dev"
lnb alias test "npm test -- --watch"
lnb alias deploy "rsync -av dist/ server:/var/www/"

# Later, clean up
lnb remove dev-server
lnb unalias dev test deploy
```

### Working with Downloaded Tools

```bash
# Download a binary
curl -L https://github.com/user/tool/releases/download/v1.0/tool-linux > ~/Downloads/tool
chmod +x ~/Downloads/tool

# Install it (LNB detects the path automatically)
lnb ~/Downloads/tool

# Use it anywhere
tool --version
```

## How It Works

### Smart Command Detection

LNB uses intelligent parsing to determine what you want to do:

1. **File path detection**: If the argument looks like a file path (absolute, relative, or contains `/` or `\`), LNB treats it as an install command
2. **Known command detection**: If the first argument is a known command (`install`, `remove`, `alias`, `unalias`, `list`, `help`, `version`), LNB executes that command
3. **Fallback**: If uncertain, LNB shows help

```bash
lnb ./mybinary          # Detected as file path → install
lnb /usr/bin/tool       # Detected as file path → install  
lnb list                # Detected as known command → list
lnb remove mytool       # Detected as known command → remove
lnb help                # Detected as known command → help
```

### Binary Installation

- **Linux/macOS**  
  Uses `os.Symlink()` to create `/usr/local/bin/<name>` pointing to your binary.  
  *Why it's simple*: Symlinks are built-in, reliable, and everyone's used to binaries living in `/usr/local/bin`.

- **Windows**  
  1. Ensures `%USERPROFILE%\bin\` exists
  2. Creates a `.cmd` wrapper:
     ```bat
     @echo off
     "C:\full\path\to\binary.exe" %*
     ```
  3. Reminds you to add `%USERPROFILE%\bin` to your PATH if needed

  *Why a wrapper?* Windows doesn't handle symlinks to arbitrary files as gracefully—this is more consistent.

### Command Aliases

- **Linux/macOS**  
  Creates executable shell scripts that run your command with arguments passed through.

- **Windows**  
  Creates `.bat` files that execute your command with arguments.

### Configuration Tracking

LNB maintains a configuration file at `~/.lnb/config.json` that tracks:
- What you've installed (binaries and aliases)
- Source paths and target locations  
- Installation timestamps
- Entry types (binary or alias)

This enables the `lnb list` command and helps prevent conflicts.

## Error Handling

LNB provides clear error messages for common issues:

- **File doesn't exist**: "File /path/to/file does not exist"
- **Not executable**: "File is not executable" 
- **Already installed**: "Binary 'name' is already installed"
- **Not found for removal**: "Binary 'name' not found"
- **Invalid alias**: Validates that alias commands are executable

## Building from Source

**Requirements**:
- Go 1.23+
- [Task](https://taskfile.dev) (optional, for automation)

### Available Tasks

```bash
task build              # Build for current platform
task build:all          # Build for all platforms using GoReleaser
task test:unit          # Run unit tests
task test:integration   # Run integration tests  
task test:all           # Run all tests
task install            # Install LNB locally
task remove             # Remove LNB from local system
task clean              # Clean build artifacts
task version            # Show current version

```

### Quick Start

1. Clone the repo:
   ```bash
   git clone https://github.com/muthuishere/lnb.git
   cd lnb
   ```

2. Build and install:
   ```bash
   task build
   task install
   ```

3. Test it works:
   ```bash
   lnb version
   ```


## Package Management

### Separate Installer Repositories

Package manager configurations are maintained in separate repositories to avoid merge conflicts during releases:

- **Homebrew Cask**: https://github.com/muthuishere/homebrew-lnb
- **Scoop Bucket**: https://github.com/muthuishere/scoop-lnb  
- **NPM Package**: Published directly to npmjs.org as `lnb`

These are automatically updated by GoReleaser when new versions are released.

## Command Reference

```bash
# Smart installation (recommended)
lnb <file-path>                     # Auto-detect and install binary

# Explicit commands
lnb install <file-path>             # Install binary explicitly  
lnb remove <binary-name>            # Remove installed binary
lnb alias <name> "<command>"        # Create command alias
lnb unalias <name>                  # Remove alias
lnb list                            # List all installed items
lnb help                            # Show help
lnb version                         # Show version information
lnb                                 # Show help (no arguments)
```

### File Path Detection

LNB automatically detects file paths and treats them as install commands when:
- Path starts with `/` (absolute path)
- Path starts with `./` or `../` (relative path)  
- Path starts with `~/` (home directory)
- Path contains `/` (Unix-style path separator)
- Path contains `\` (Windows-style path separator)
- Path exists as a file on the filesystem

## License

MIT
