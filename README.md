# agentree 🌳🤖

Run multiple AI coding agents without them fighting over files.

## The Problem

Want to run Claude in multiple terminals? They'll overwrite each other's changes.

Git worktrees could help, but setting them up is painful:
- Create the worktree manually
- Copy over `.env` files
- Don't forget `.env.local`, `.dev.vars`...
- Install dependencies again
- Copy `.claude/settings.local.json`
- 10 minutes later, you're finally ready

## The Solution

```bash
agentree -b new-feature
```

One command gives you:
- ✓ New branch (`agent/new-feature`)
- ✓ Isolated worktree in `../myrepo-worktrees/`
- ✓ All env files copied automatically
- ✓ Dependencies installed
- ✓ Ready to code in seconds

## Install

```bash
brew tap AryaLabsHQ/tap
brew install AryaLabsHQ/tap/agentree
```

## Quick Start

Run multiple AI agents in parallel:

```bash
# Terminal 1: Claude working on auth
agentree -b fix-auth

# Terminal 2: Another Claude on UI bugs  
agentree -b ui-fixes

# Terminal 3: Cursor adding tests
agentree -b add-tests
```

Each agent works in isolation. No conflicts. Pure productivity.

## How It Works

1. **Interactive mode** (recommended):
   ```bash
   agentree -i
   ```
   Guides you through branch creation with all options.

2. **Quick mode**:
   ```bash
   agentree -b feature-name
   ```
   Creates worktree with smart defaults.

3. **Cleanup**:
   ```bash
   agentree rm agent/feature-name
   ```
   Removes worktree when you're done.

## Real World Example

Here's my actual workflow from last week:

```bash
# 9:00 AM - Start refactoring auth
agentree -b refactor-auth
# Let Claude work on this big task...

# 9:05 AM - Meanwhile, fix that urgent bug
agentree -b fix-login-bug
# Different Claude instance handles it

# 9:10 AM - Update docs while waiting
agentree -b update-api-docs

# 11:00 AM - Review all three PRs separately
```

<details>
<summary><strong>📚 Advanced Usage</strong></summary>

### Flags & Options

```bash
# Create from specific base branch
agentree -b feature-x -f main

# Skip environment file copying
agentree -b feature-x -e=false

# Skip auto-setup (dependency installation)
agentree -b feature-x -s=false

# Push to remote after creation
agentree -b feature-x -p

# Create and open PR
agentree -b feature-x -r

# Custom destination
agentree -b feature-x -d ~/custom-dir

# Remove worktree and delete branch
agentree rm agent/feature-x -R
```

### Configuration

Create `.agentreerc` in your project:

```bash
# .agentreerc
POST_CREATE_SCRIPTS=(
  "pnpm install"
  "pnpm build"
  "cp .env.example .env"
)
```

### Auto-Detection

Agentree automatically detects and runs the right setup:
- **pnpm/npm/yarn**: Installs dependencies + build
- **cargo**: Runs `cargo build`
- **pip**: Installs from requirements.txt
- **go**: Downloads modules

</details>

<details>
<summary><strong>🛠️ Installation Options</strong></summary>

### macOS/Linux Binary

```bash
# Download latest release
curl -L https://github.com/AryaLabsHQ/agentree/releases/latest/download/agentree-$(uname -s)-$(uname -m) -o agentree
chmod +x agentree
sudo mv agentree /usr/local/bin/
```

### From Source

```bash
git clone https://github.com/AryaLabsHQ/agentree.git
cd agentree
make build
make install
```

### Shell Completion

<details>
<summary>Bash</summary>

```bash
echo 'source <(agentree completion bash)' >> ~/.bashrc
```
</details>

<details>
<summary>Zsh</summary>

```bash
echo 'source <(agentree completion zsh)' >> ~/.zshrc
```
</details>

<details>
<summary>Fish</summary>

```bash
agentree completion fish > ~/.config/fish/completions/agentree.fish
```
</details>

</details>

<details>
<summary><strong>🤔 FAQ</strong></summary>

**Q: What's the difference between this and regular git worktrees?**

A: Agentree handles all the setup that git doesn't:
- Copies your environment files
- Installs dependencies
- Copies AI tool configurations
- Does it all in one command

**Q: Which AI tools does this work with?**

A: Any tool that edits code! Tested with:
- Claude Code (what I built it for)
- Cursor
- GitHub Copilot
- Cody
- Continue
- Any future AI coding tool

**Q: Can I use custom branch prefixes?**

A: Not yet, but it's on the roadmap. Currently uses `agent/` prefix.

**Q: Does it work with monorepos?**

A: Yes! Run agentree from any subdirectory.

</details>

## Why I Built This

I switched from Cursor to Claude Code to save money, but Claude is *slow*. Like, really slow. 10-minute waits for complex refactors.

So I started running multiple instances. But they kept overwriting each other's work. Git worktrees seemed perfect, but the manual setup was killing my productivity.

Agentree was born from frustration. Now I run 4-5 Claude instances in parallel, each on their own branch, with zero conflicts.

[Read the full story →](https://www.saatvikarya.com/agentree)

## Contributing

Found a bug? Have an idea? PRs welcome!

See something that could be better? [Open an issue](https://github.com/AryaLabsHQ/agentree/issues).

## License

MIT

---

Built with ❤️ and frustration by [@aryasaatvik](https://x.com/aryasaatvik)