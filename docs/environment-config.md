# Environment Configuration

agentree intelligently discovers and copies environment files based on `.gitignore` patterns, with support for monorepos and AI tool configurations.

## How It Works

1. Parses all `.gitignore` files in your repository
2. Identifies environment-related patterns (`.env`, `local`, `secret`, etc.)
3. Finds matching files and copies them to the new worktree
4. Includes AI tool configs (`.claude/settings.local.json`, `.cursorrules`, etc.)

## Configuration

### Project Config (.agentreerc)

```bash
# Additional patterns to include
ENV_INCLUDE_PATTERNS=(
  "*.env.example"
  "config/*.sample"
)

# Patterns to exclude
ENV_EXCLUDE_PATTERNS=(
  "*.test.env"
  "node_modules/**/.env"
)
```

### Global Config (~/.config/agentree/config)

```bash
# Comma-separated patterns
ENV_INCLUDE_PATTERNS=.env.global,.company-secrets
ENV_EXCLUDE_PATTERNS=*.backup,*.tmp
```

## Examples

```bash
# Basic usage - auto-discovers from .gitignore
agentree create -b feature/new-api

# Disable environment copying
agentree create -b feature/test -e=false

# Debug discovery process
agentree create -b feature/debug -v
```

## Supported Files

- **Auto-detected**: Files in `.gitignore` containing keywords like `.env`, `local`, `secret`
- **AI configs**: `.claude/settings.local.json`, `.cursorrules`, `.github/copilot/config.json`
- **Monorepo**: Recursive patterns like `**/.env`, `packages/*/.env`

## Troubleshooting

- **No files copied?** Check if files exist and are in `.gitignore`
- **Too many files?** Add exclude patterns to `.agentreerc`
- **Debug mode**: Use `-v` flag to see discovery details

## Best Practices

1. Add environment files to `.gitignore` for automatic discovery
2. Use `.env.local` for machine-specific settings
3. Keep AI configs like `.claude/settings.local.json` for project context