# hatch

A bash script for creating and managing isolated Git worktrees for agentic workflows.

## Overview

`hatch` simplifies working with multiple AI coding agents (like Codex and Claude Code) by creating isolated Git worktrees. Each agent gets its own branch and directory, allowing concurrent work without conflicts.

## Features

- **Quick worktree creation**: Automatically prefixes branches with `agent/` for organization
- **Easy cleanup**: Remove worktrees and optionally delete branches
- **GitHub integration**: Push branches and create PRs directly
- **Flexible paths**: Custom destination directories or automatic organization
- **Environment copying**: Optionally copy `.env` and `.dev.vars` files to new worktrees

## Installation

```bash
chmod +x hatch
cp hatch ~/bin/  # or any directory in your PATH
```

Or add to your PATH:
```bash
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc
```

## Usage

### Create a worktree

```bash
# Create worktree on branch agent/feature-x
hatch -b feature-x

# Create from specific base branch
hatch -b feature-x -f main

# Push to origin after creation
hatch -b feature-x -p

# Create PR after push (requires gh CLI)
hatch -b feature-x -r

# Custom destination
hatch -b feature-x -d ~/projects/custom-dir

# Copy .env and .dev.vars files from source
hatch -b feature-x -e
```

### Remove a worktree

```bash
# Remove by branch name
hatch rm agent/feature-x

# Remove by path
hatch rm ../myrepo-worktrees/agent-feature-x

# Force removal (no confirmation)
hatch rm agent/feature-x -y

# Also delete the branch
hatch rm agent/feature-x -R
```

## How it works

1. **Worktree organization**: Creates worktrees in `../repo-worktrees/` by default
2. **Branch naming**: Automatically prefixes with `agent/` unless branch contains `/`
3. **Safe operations**: Validates paths and branches before creation
4. **Git integration**: Uses native Git worktree commands for reliability

## Examples

```bash
# AI agent working on authentication
hatch -b auth-system

# AI agent fixing bugs, push when ready
hatch -b bugfix-123 -p

# Quick experiment, create PR immediately
hatch -b experiment-ml -r

# Cleanup after work is merged
hatch rm agent/auth-system -R
```

## Requirements

- Git 2.5+ (for worktree support)
- Bash 4.0+
- Optional: `gh` CLI for GitHub PR creation

## License

MIT