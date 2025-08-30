# LNB – Cross-Platform Alias Manager

*"The best software is written to solve problems the author actually has."* — DHH

I got tired of doing different things on every platform for simple commands:

**Unix/Mac:**
```bash
echo 'alias deploy="docker run --rm -v $(pwd):/app deploy"' >> ~/.zshrc  
source ~/.zshrc
```

**Windows:**
```powershell
# Create .bat files manually? 
# Edit PowerShell profile? 
# Add to System PATH?
# Good luck remembering how.
```

Different rules. Different commands. Different headaches.

I wanted one command that works everywhere. So I built it.

## Installation

```bash
npm install -g lnb
```

## Usage

**Create shortcuts that work from anywhere:**
```bash
# Now you can run 'deploy' from any terminal or Start menu search (Windows)
lnb alias deploy "docker run --rm -v $(pwd):/app deploy-image"

# Run 'logs' from anywhere - terminal, Run dialog (Win+R), or Start menu
lnb alias logs "tail -f /var/log/nginx/access.log"  

# Type 'serve' in any terminal or Windows Start bar search
lnb alias serve "python -m http.server 8080"
```

**Make a binary globally accessible:**
```bash
# Now 'mybinary' works from anywhere - any terminal or Windows Start menu
lnb ./mybinary
```

**List everything:**
```bash
lnb list
```

**Remove stuff:**
```bash
lnb remove mybinary
lnb unalias deploy
```

## How it works

**Same command. All platforms.**

- **Unix/Mac**: Creates shell scripts in `/usr/local/bin`
- **Windows**: Creates .bat/.cmd files in `%USERPROFILE%\bin` and adds to your PATH automatically

You don't need to know or care about these details.

After installation, your commands work from:
- ✅ **Any terminal** (PowerShell, Command Prompt, Bash, etc.)
- ✅ **Windows Start menu search** (just type the command name)
- ✅ **Run dialog** (Win+R on Windows)
- ✅ **Anywhere you'd normally type commands**

## That's it

LNB does one thing: manages aliases and binaries consistently across platforms.

No configuration files. No plugins. No complexity.

It just works.

---

**Source**: [github.com/muthuishere/lnb](https://github.com/muthuishere/lnb) • **License**: MIT
