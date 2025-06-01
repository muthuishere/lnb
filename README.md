````markdown
# LNB – Link Binary

## The Problem

Have you ever spent too much time tweaking your PATH, creating symlinks, or writing wrapper scripts just so your custom CLI or downloaded binary works from anywhere? Yeah, it sucks. You shouldn’t have to jump through hoops every time you build or grab a new tool.

**Why this matters**: Pointless setup steps slow you down. You want to focus on coding or using the tool, not messing with config files.

## The Solution

**LNB** fixes that by giving you a one-liner to make any binary globally accessible. On Linux/macOS, it just makes a symlink in your PATH. On Windows, it drops in a tiny wrapper script. No extra config, no guessing which folder to use—LNB handles it.

**Why I like it**: It’s dead simple. You don’t need to learn some complex packaging system or remember different commands per OS.

## Features

- **Cross-platform**: Same command on Linux, macOS, and Windows.  
- **Minimal interface**: Only two actions—`install` (default) and `remove`.  
- **Zero config**: LNB auto-detects your OS and does the right thing.  
- **Fast**: No dependencies beyond the Go binary itself.

## Installation

### From Package Managers

I recommend using a package manager if you can—setup is instant.

- **Homebrew (macOS/Linux)**  
  ```bash
  brew tap muthuishere/homebrew-tap
  brew install lnb
````

*Why Homebrew?* Everyone on macOS/Linux already has it, and updates are a breeze.

* **Chocolatey (Windows)**

  ```powershell
  choco install lnb
  ```

  *Why Chocolatey?* It’s the de facto on Windows—no extra hassle.

* **Scoop (Windows)**

  ```powershell
  scoop bucket add muthuishere https://github.com/muthuishere/scoop-bucket.git
  scoop install lnb
  ```

  *Why Scoop?* If you’re already using Scoop, it feels right to keep things consistent.



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

* On **Linux/macOS**, this creates:

  ```
  /usr/local/bin/<binary> -> /absolute/path/to/binary
  ```
* On **Windows**, this creates:

  ```
  %USERPROFILE%\bin\<binary>.cmd
  ```

  which wraps your `.exe` or any executable. *(Make sure `%USERPROFILE%\bin` is in your PATH.)*

**Why it’s neat**: You don’t have to think about where “bin” directories live on each OS—LNB does it.

### Remove a Binary

```bash
lnb path/to/binary remove
```

* Deletes the symlink (or `.cmd` wrapper on Windows), but leaves your original file untouched.

## Examples

* **Make a dev tool available everywhere**

  ```bash
  cd ~/projects/mytool
  lnb mytool         # now "mytool" runs anywhere
  ```

  *Opinion*: I love not having to copy binaries around. One command and it “just works.”

* **Uninstall when you’re done**

  ```bash
  lnb ~/projects/mytool remove
  ```

  *Opinion*: Removing is just as easy—no more hunting for old symlinks.

## How It Works

* **Linux/macOS (`linux.go` & `mac.go`)**
  Uses `os.Symlink()` to create `/usr/local/bin/<name>`. That’s it.
  *Why it’s simple*: Symlinks are built-in, reliable, and everyone’s used to binaries living in `/usr/local/bin`.

* **Windows (`windows.go`)**

  1. Ensures `~/bin/` exists.
  2. Writes a tiny `.cmd` wrapper that does:

     ```bat
     @echo off
     "C:\full\path\to\<binary>.exe" %*
     ```
  3. Reminds you to add `~/bin` to your PATH if it isn’t already.
     *Why a wrapper?* Windows doesn’t handle symlinks to arbitrary files as gracefully—this is more consistent.

## Building from Source

**Requirements**:

* Go 1.16+
* [Task](https://taskfile.dev) (optional, for automation)

1. Clone the repo:

   ```bash
   git clone https://github.com/muthuishere/lnb.git
   cd lnb
   ```

2. Build for your current platform:

   ```bash
   task build:local
   ```

   Or build everything:

   ```bash
   task build
   ```

   *Opinion*: Using `task` means no manual `GOOS/GOARCH` juggling. I like saving keystrokes.

3. Move the binary into your PATH:

   ```bash
   sudo mv lnb /usr/local/bin/
   chmod +x /usr/local/bin/lnb
   ```

   *(Windows users drop `lnb.exe` into `%USERPROFILE%\bin`.)*


## License

MIT

