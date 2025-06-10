# Environment Configuration Guide

agentree now provides intelligent environment file management that automatically discovers and copies environment files based on your `.gitignore` patterns, with support for monorepos and AI tool configurations.

## Overview

The enhanced environment configuration system:
- Uses `.gitignore` as the source of truth for environment files
- Automatically discovers files like `.env`, `.env.local`, `.env.production`, etc.
- Supports nested environment files in monorepos
- Copies AI tool configurations (`.claude/settings.local.json`, `.cursorrules`, etc.)
- Provides flexible configuration through `.agentreerc` and global config

## How It Works

When creating a new worktree, agentree:
1. Parses all `.gitignore` files in your repository
2. Identifies patterns that likely represent environment/configuration files
3. Finds actual files matching those patterns
4. Adds default AI tool configuration patterns
5. Applies any custom include/exclude patterns from configuration
6. Copies all discovered files to the new worktree, preserving directory structure

## Configuration

### Project Configuration (.agentreerc)

Create an `.agentreerc` file in your project root:

```bash
# Enable/disable environment file copying (default: true)
ENV_COPY_ENABLED=true

# Enable recursive search for env files in monorepos (default: true)
ENV_RECURSIVE=true

# Use .gitignore patterns as source of truth (default: true)
ENV_USE_GITIGNORE=true

# Additional patterns to include
ENV_INCLUDE_PATTERNS=(
  "*.env.example"
  "config/*.sample"
  ".vscode/settings.local.json"
)

# Patterns to exclude
ENV_EXCLUDE_PATTERNS=(
  "*.test.env"
  "node_modules/**/.env"
)
```

### Global Configuration

Create `~/.config/agentree/config`:

```bash
# Global environment settings
ENV_COPY_ENABLED=true
ENV_RECURSIVE=true
ENV_USE_GITIGNORE=true

# Comma-separated patterns for global config
ENV_INCLUDE_PATTERNS=.env.global,.company-secrets
ENV_EXCLUDE_PATTERNS=*.backup,*.tmp
```

## Supported File Types

### Automatically Detected (via .gitignore)

The system intelligently detects files based on keywords in your `.gitignore`:
- Files containing: `.env`, `.vars`, `local`, `secret`, `config`, `settings`, `credentials`
- Common patterns: `.env`, `.env.*`, `*.local`, `.dev.vars`

### Default AI Tool Configurations

Always included regardless of `.gitignore`:
- `.claude/settings.local.json` - Claude AI local settings
- `.cursorrules` - Cursor AI rules
- `.github/copilot/config.json` - GitHub Copilot configuration
- `.aider.conf` - Aider configuration
- `.codeium/config.json` - Codeium configuration

### Monorepo Support

Recursive patterns are automatically handled:
- `**/.env` - All .env files at any depth
- `**/.env.local` - All .env.local files
- `packages/*/.env` - Environment files in package directories
- `apps/*/.env.*` - All env variants in apps directories

## Examples

### Basic Usage

```bash
# Uses intelligent discovery based on .gitignore
agentree create -b feature/new-api

# Output:
# ✅ Worktree ready:
#     path /Users/you/project-worktrees/agent-feature-new-api
#     branch agent/feature/new-api (from main)
# Discovering environment files...
# 📋 Copied .env
# 📋 Copied .env.local
# 📋 Copied .claude/settings.local.json
# 📋 Copied packages/api/.env
# 📋 Copied packages/web/.env.local
```

### Disable Environment Copying

```bash
# Via command line
agentree create -b feature/test --no-env

# Via configuration
echo "ENV_COPY_ENABLED=false" >> .agentreerc
```

### Custom Patterns

Add patterns to `.agentreerc`:

```bash
ENV_INCLUDE_PATTERNS=(
  "*.env.example"           # Include example files
  "config/*.sample"         # Include sample configs
  "secrets/*.template"      # Include secret templates
  "**/*.local.json"        # All local JSON configs
)

ENV_EXCLUDE_PATTERNS=(
  "test/**/*"              # Exclude test files
  "*.backup"               # Exclude backups
  "node_modules/**/*"      # Exclude deps
)
```

## Configuration Precedence

Configuration is merged in this order (later overrides earlier):
1. Built-in defaults
2. Global config (`~/.config/agentree/config`)
3. Project config (`.agentreerc`)
4. Command-line flags

## Troubleshooting

### No Files Copied

1. Check if `.gitignore` exists and contains environment patterns
2. Verify files actually exist in the repository
3. Check configuration with `ENV_COPY_ENABLED=true`
4. Look for exclude patterns that might be filtering files

### Too Many Files Copied

1. Add exclude patterns to filter unwanted files
2. Disable gitignore parsing: `ENV_USE_GITIGNORE=false`
3. Use custom patterns instead of automatic discovery

### Debugging

Run with verbose output to see discovery process:
```bash
# This will show which files are being discovered and why
agentree create -b feature/debug -v
```

## Best Practices

1. **Use `.gitignore`** - Add all environment files to `.gitignore` for automatic discovery
2. **Local Overrides** - Use `.env.local` or `*.local` patterns for machine-specific settings
3. **Monorepo Organization** - Place env files next to their packages for better organization
4. **AI Tool Configs** - Keep `.claude/settings.local.json` for project-specific AI context
5. **Templates** - Include `.env.example` files to document required variables

## Migration from Old Version

The new system is backward compatible. Old behavior (copying only `.env` and `.dev.vars`) is preserved as a fallback if:
- Gitignore parsing fails
- No gitignore files exist
- Discovery returns no results

To migrate:
1. Ensure environment files are in `.gitignore`
2. Remove any custom scripts that copy env files
3. Configure patterns as needed in `.agentreerc`