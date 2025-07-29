# LNB – Link Binary

## The Problem

Have you ever spent too much time tweaking your PATH, creating symlinks, or writing wrapper scripts just so your custom CLI or downloaded binary works from anywhere? Yeah, it sucks. You shouldn't have to jump through hoops every time you build or grab a new tool.

**Why this matters**: Pointless setup steps slow you down. You want to focus on coding or using the tool, not messing with config files.

## The Solution

**LNB** fixes that by giving you a one-liner to make any binary globally accessible. On Linux/macOS, it creates symlinks in `/usr/local/bin`. On Windows, it creates wrapper scripts in `%USERPROFILE%\bin`. Plus, it can create command aliases for complex commands. No extra config, no guessing which folder to use—LNB handles it all and tracks what you've installed.

**Why I like it**: It's dead simple. You don't need to learn some complex packaging system or remember different commands per OS.

## Features

- **Cross-platform**: Same command on Linux, macOS, and Windows  
- **Binary installation**: Install any executable to your PATH with one command
- **Command aliases**: Create shortcuts for complex commands (e.g., `java -jar myapp.jar`)
- **Installation tracking**: Keep track of what you've installed with `lnb list`
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
   ./lnb-[platform] ./lnb-[platform] install
   ```

   This makes `lnb` available system-wide so you can use LNB to manage other tools.

## Usage

### Install a Binary

```bash
# Default action is "install"
lnb path/to/binary
# or explicitly:
lnb path/to/binary install
```

- On **Linux/macOS**, this creates:
  ```
  /usr/local/bin/<binary> -> /absolute/path/to/binary
  ```
- On **Windows**, this creates:
  ```
  %USERPROFILE%\bin\<binary>.cmd
  ```
  which wraps your `.exe` or any executable. *(Make sure `%USERPROFILE%\bin` is in your PATH.)*

**Why it's neat**: You don't have to think about where "bin" directories live on each OS—LNB does it.

### Remove a Binary

```bash
lnb path/to/binary remove
```

Deletes the symlink (or `.cmd` wrapper on Windows), but leaves your original file untouched.

### Create Command Aliases

```bash
# Create an alias for a complex command
lnb alias myapp "java -jar ./myapp.jar"
lnb alias serve "python -m http.server 8080"
lnb alias deploy "docker run --rm -v $(pwd):/workspace deploy-tool"
```

- **Linux/macOS**: Creates a shell script at `/usr/local/bin/<alias-name>`
- **Windows**: Creates a batch file at `%USERPROFILE%\bin\<alias-name>.bat`

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
- Source path (for binaries) or command (for aliases)
- Target location

## Examples

- **Make a dev tool available everywhere**

  ```bash
  cd ~/projects/mytool
  lnb mytool         # now "mytool" runs anywhere
  ```

  *Opinion*: I love not having to copy binaries around. One command and it "just works."

- **Create an alias for a Java application**

  ```bash
  lnb alias myapp "java -jar ./target/myapp.jar"
  # Now you can run "myapp" from anywhere
  ```

- **Uninstall when you're done**

  ```bash
  lnb ~/projects/mytool remove
  lnb unalias myapp
  ```

  *Opinion*: Removing is just as easy—no more hunting for old symlinks.

## How It Works

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

This enables the `lnb list` command and helps prevent conflicts.

## Building from Source

**Requirements**:
- Go 1.23+
- [Task](https://taskfile.dev) (optional, for automation)

1. Clone the repo:
   ```bash
   git clone https://github.com/muthuishere/lnb.git
   cd lnb
   ```

2. Build for your current platform:
   ```bash
   task build
   ```

   Or build for all platforms:
   ```bash
   task build:all
   ```

3. Install the built binary:
   ```bash
   task install
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
lnb <file>                          # Install binary (default action)
lnb <file> install                  # Install binary explicitly  
lnb <file> remove                   # Remove binary
lnb alias <name> <command>          # Create command alias
lnb unalias <name>                  # Remove alias
lnb list                            # List all installed items
lnb help                            # Show help
lnb version                         # Show version information
```

## License

MIT
