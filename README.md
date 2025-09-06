# LNB – Cross-Platform Alias Manager

*"The best software is written to solve problems the author actually has."* — DHH

I got tired of doing different things on every platform for simple aliases:

**Unix:**
```bash
echo 'alias deploy="docker run --rm -v $(pwd):/app deploy"' >> ~/.zshrc  
source ~/.zshrc
```

**Windows:**
```powershell
# Create .bat files manually? 
# Edit PowerShell profile? 
# Good luck remembering where that is.
```

Different rules. Different commands. Different headaches.

I wanted one command that works everywhere. So I built it.

## Installation

```bash
npm install -g lnb
```

## Usage

**Create an alias:**
```bash
lnb alias deploy "docker run --rm -v $(pwd):/app deploy-image"
lnb alias logs "tail -f /var/log/nginx/access.log"
lnb alias serve "python -m http.server 8080"
```

**Make a binary globally accessible:**
```bash
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

- **Unix**: Creates shell scripts in `/usr/local/bin`
- **Windows**: Creates .bat/.cmd files in `%USERPROFILE%\bin`

You don't need to know or care about these details.

## That's it

LNB does one thing: manages aliases and binaries consistently across platforms.

No configuration files. No plugins. No complexity.

It just works.

---

**Source**: [github.com/muthuishere/lnb](https://github.com/muthuishere/lnb) • **License**: MIT
