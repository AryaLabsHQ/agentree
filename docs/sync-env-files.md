# Syncing Environment Files Back to Main Worktree

agentree now supports syncing environment files back to the main worktree when removing a worktree. This ensures that any changes made to environment files in isolated worktrees can be preserved.

## How It Works

When you remove a worktree, agentree can automatically sync modified environment files back to the main worktree. Only files that:
1. Exist in both the worktree and main worktree
2. Have been modified (newer timestamp in the worktree AND different content)

Will be synced back. This dual check prevents unnecessary syncs when only timestamps differ but content remains the same. New files created in the worktree are not synced to avoid accidentally committing sensitive files.

## Usage

### Command Line Flag

Use the `-S` or `--sync-env` flag when removing a worktree:

```bash
agentree rm agent/feature-x --sync-env
```

### Configuration

You can enable automatic sync-back in your configuration:

#### Project Configuration (.agentreerc)

```bash
# Enable automatic sync on remove
ENV_SYNC_BACK_ON_REMOVE=true

# Optionally specify patterns for files to sync
# If not specified, uses the same patterns as ENV_INCLUDE_PATTERNS
ENV_SYNC_PATTERNS=(
  ".env"
  ".env.local"
  ".claude/settings.local.json"
)
```

#### Global Configuration (~/.config/agentree/config)

```bash
# Enable for all projects
ENV_SYNC_BACK_ON_REMOVE=true

# Comma-separated patterns
ENV_SYNC_PATTERNS=.env,.env.local,.claude/settings.local.json
```

## Safety Features

1. **Temporary Backups**: A temporary backup is created during sync and automatically removed after successful sync
2. **Rollback on Failure**: If sync fails, the original file is restored from the temporary backup
3. **Skip New Files**: Files that exist only in the worktree (not in main) are skipped
4. **Confirmation**: The remove command still asks for confirmation before removing the worktree
5. **Verbose Mode**: Use `-v` flag to see detailed information about what's being synced

## Example Workflow

```bash
# Create a new worktree
agentree create -b feature-x

# Work in the worktree, modify .env files
cd ../myrepo-worktrees/agent-feature-x
echo "NEW_API_KEY=xyz" >> .env

# When done, sync environment changes back and remove
agentree rm agent/feature-x --sync-env -v

# Output:
# 🔍 Checking 3 files for modifications...
# 📝 File .env modified (worktree: 2024-01-15T10:30:00Z, main: 2024-01-15T09:00:00Z)
# ✅ Synced 1 files:
#     📋 .env
# Remove worktree at /path/to/worktree? [y/N]
```

## Configuration Precedence

The sync behavior follows this precedence:
1. Command line flag (`--sync-env`)
2. Project configuration (`.agentreerc`)
3. Global configuration (`~/.config/agentree/config`)
4. Default (sync disabled)

## AI Tool Configuration Files

The sync feature works particularly well with AI tool configuration files that are often git-ignored:

- `.claude/settings.local.json` - Claude's local settings
- `.cursorrules` - Cursor AI rules
- `.aider.conf` - Aider configuration
- `.github/copilot/config.json` - GitHub Copilot settings

These files are automatically discovered and synced if they've been modified.