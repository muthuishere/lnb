# LNB – Cross-Platform Command Manager

## The Problem

You build tools. Scripts. Binaries. Commands you want to run from anywhere.

But every platform has different rules:

- **Mac/Linux**: Edit `.bashrc`, source it, hope it works across shells
- **Windows**: Create `.bat` files, mess with PATH, pray PowerShell finds them

You waste time on platform differences instead of building things.

## The Solution

One command. All platforms. No configuration.

```bash
npm install -g lnb
```

## Why You Need This

**Stop typing long commands:**
```bash
# Instead of this every time:
docker run --rm -v $(pwd):/workspace -w /workspace node:18 npm run build

# Do this once:
lnb alias build "docker run --rm -v $(pwd):/workspace -w /workspace node:18 npm run build"

# Then just:
build
```

**Make your tools globally available:**
```bash
# You built a great tool, but it only works from its directory
./my-awesome-tool --help

# Make it work from anywhere:
lnb ./my-awesome-tool

# Now this works from any directory:
my-awesome-tool --help
```

**Share commands across your team:**
```bash
# Everyone can run the same commands, same way, all platforms:
lnb alias deploy "docker-compose -f prod.yml up -d"
lnb alias logs "kubectl logs -f deployment/app"
lnb alias test "npm test -- --coverage"
```

## Usage

**Create shortcuts:**
```bash
lnb alias deploy "docker-compose up -d"
lnb alias serve "python -m http.server 8080"
lnb alias logs "tail -f /var/log/app.log"
```

**Make binaries global:**
```bash
lnb ./mybinary          # Now 'mybinary' works everywhere
lnb ~/tools/deploy.sh   # Now 'deploy.sh' works everywhere
```

**Manage everything:**
```bash
lnb list                # See what you've installed
lnb remove mybinary     # Remove a binary
lnb unalias deploy      # Remove an alias
```

## How It Works

Same command. All platforms. Zero configuration.

Your commands work everywhere:
- ✅ Any terminal
- ✅ Windows Start menu search  
- ✅ Command prompt, PowerShell, Bash
- ✅ Anywhere you'd normally type commands

LNB handles the platform differences so you don't have to.

## That's It

LNB does one thing: makes your commands and tools globally available.

No config files. No learning curve. It just works.

---

**Source**: [github.com/muthuishere/lnb](https://github.com/muthuishere/lnb) • **License**: MIT
