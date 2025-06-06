# agentree 🌳🤖

Create and manage isolated Git worktrees for AI coding agents.

## Overview

`agentree` simplifies working with multiple AI coding agents (like Cursor, Claude Code, and GitHub Copilot) by creating isolated Git worktrees. Each agent gets its own branch and directory, allowing concurrent work without conflicts.

## Why agentree?

When working with AI coding assistants, you often need to:
- Create isolated environments for different tasks
- Quickly switch between multiple concurrent experiments
- Maintain clean separation between AI-generated changes
- Easily clean up after experiments

`agentree` automates this workflow with a single command.

## Installation

### Homebrew (Recommended)

```bash
brew tap AryaLabsHQ/tap
brew install agentree
```

### Download Pre-built Binary

Download the latest release for your platform from the [releases page](https://github.com/AryaLabsHQ/agentree/releases).

### Build from Source

```bash
git clone https://github.com/AryaLabsHQ/agentree.git
cd agentree
make build
make install
```

## Usage

### Create a worktree

```bash
# Interactive mode - select base branch from list
agentree -i

# Create worktree with specific branch name
agentree -b feature-x

# Create from specific base branch
agentree -b feature-x -f main

# Push to origin after creation
agentree -b feature-x -p

# Create PR after push (requires gh CLI)
agentree -b feature-x -r

# Custom destination
agentree -b feature-x -d ~/projects/custom-dir

# Copy .env and .dev.vars files from source
agentree -b feature-x -e

# Auto-detect and run setup (pnpm install, npm install, etc.)
agentree -b feature-x -s

# Run custom post-create scripts
agentree -b feature-x -S "pnpm install --frozen-lockfile" -S "pnpm test"

# Combine: copy env files and run setup
agentree -b feature-x -e -s
```

### Remove a worktree

```bash
# Remove by branch name
agentree rm agent/feature-x

# Remove by path
agentree rm ../myrepo-worktrees/agent-feature-x

# Force removal (no confirmation)
agentree rm agent/feature-x -y

# Also delete the branch
agentree rm agent/feature-x -R
```

## Features

- **Interactive branch selection**: Use `-i` flag for a TUI branch picker
- **Quick worktree creation**: Automatically prefixes branches with `agent/` for organization
- **Easy cleanup**: Remove worktrees and optionally delete branches
- **GitHub integration**: Push branches and create PRs directly
- **Flexible paths**: Custom destination directories or automatic organization
- **Environment copying**: Optionally copy `.env` and `.dev.vars` files to new worktrees
- **Auto-setup**: Automatically detect and run package manager install/build commands
- **Configurable**: Project and global configuration for custom post-create scripts
- **Cross-platform**: Works on macOS, Linux, and Windows

## Configuration

### Project Configuration (`.agentreerc`)

Create a `.agentreerc` file in your project root to define custom post-create scripts:

```bash
# .agentreerc
POST_CREATE_SCRIPTS=(
  "pnpm install"
  "pnpm build"
  "cp .env.example .env"
)
```

### Global Configuration (`~/.config/agentree/config`)

Create a global config for user-wide defaults:

```bash
# ~/.config/agentree/config
# Override auto-detected scripts
PNPM_SETUP="pnpm install --frozen-lockfile && pnpm build"
NPM_SETUP="npm ci && npm run build"

# Default when no package manager detected
DEFAULT_POST_CREATE="echo 'Ready to work!'"
```

### Auto-Detection

When using `-s` flag, agentree automatically detects and runs appropriate setup commands:

- **pnpm**: `pnpm install` + `pnpm build` (if build script exists)
- **npm**: `npm install` + `npm run build` (if build script exists)
- **yarn**: `yarn install` + `yarn build` (if build script exists)
- **cargo**: `cargo build`
- **go**: `go mod download`
- **pip**: `pip install -r requirements.txt`
- **bundler**: `bundle install`

## Examples

```bash
# AI agent working on authentication
agentree -b auth-system

# Interactive mode to choose base branch
agentree -i

# AI agent fixing bugs, push when ready
agentree -b bugfix-123 -p

# Quick experiment, create PR immediately
agentree -b experiment-ml -r

# Cleanup after work is merged
agentree rm agent/auth-system -R
```

## Development

```bash
# Clone the repository
git clone https://github.com/AryaLabsHQ/agentree.git
cd agentree

# Run tests
make test

# Build binary
make build

# Run locally
make run
```

## Version History

- **v1.0+**: Complete rewrite in Go with interactive mode, better performance, and cross-platform support
- **v0.1**: Original bash implementation (available as `agentree-v0.1.sh` for reference)

### Why the rewrite?

The original bash script served well but had limitations:
- Platform-specific issues (especially on Windows)
- Limited testing capabilities
- Difficult to add complex features like interactive mode
- No proper dependency management

The Go version provides:
- Cross-platform compatibility
- Better performance
- Comprehensive test coverage
- Interactive TUI with branch selection
- Easier distribution via homebrew
- Type safety and better error handling

## Requirements

- Git 2.5+ (for worktree support)
- Optional: `gh` CLI for GitHub PR creation

## License

MIT