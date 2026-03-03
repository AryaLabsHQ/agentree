# agentree

`agentree` creates and manages isolated git worktrees for parallel AI-agent coding workflows.

This branch is a full rewrite on Bun + TypeScript + Bunli. Backward compatibility with the legacy Go/bash CLI is intentionally removed.

## Install

```bash
npm install -g agentree
```

Or from source:

```bash
git clone https://github.com/AryaLabsHQ/agentree.git
cd agentree
bun install
bun run build
```

## Quick Start

Create a workspace:

```bash
agentree new fix-auth
```

List workspaces:

```bash
agentree ls
```

Remove a workspace:

```bash
agentree rm fix-auth --yes --deleteBranch
```

## Commands

Create workspace:

```bash
agentree new <name> [--from <ref>] [--dest <path>] [--push=true] [--pr=true]
```

List workspaces:

```bash
agentree ls [--all] [--json]
```

Remove workspace:

```bash
agentree rm <name|branch|path> [--yes] [--deleteBranch] [--json]
```

Shell completions:

```bash
agentree completions <bash|zsh|fish|powershell>
```

Interactive creation:

```bash
agentree new --interactive
```

JSON output:

```bash
agentree new feature-a --json
agentree ls --json
agentree rm feature-a --yes --json
```

## Configuration

Project config path:

```text
.agentree/config.json
```

Global config path:

```text
~/.config/agentree/config.json
```

Example:

```json
{
  "branchPrefix": "agent/",
  "copyIgnoredEnabled": true,
  "defaultExcludes": [
    ".git/",
    "node_modules/",
    ".venv/"
  ],
  "extraIncludes": [".env"],
  "extraExcludes": [".cache/"],
  "setupEnabled": true,
  "setupMode": "auto-install",
  "setupScripts": [],
  "strictDefault": true,
  "rollbackOnFailDefault": true
}
```

Config precedence:

1. CLI flags
2. Project config (`.agentree/config.json`)
3. Global config (`~/.config/agentree/config.json`)
4. Built-in defaults

## Setup Detection

When `setupEnabled` is `true` and `setupMode` is `"auto-install"`, agentree detects and runs:

- `bun install` for `bun.lock` or `bun.lockb`
- `pnpm install` for `pnpm-lock.yaml`
- `npm install` for `package-lock.json`
- `yarn install` for `yarn.lock`
- `cargo build` for `Cargo.lock`
- `go mod download` for `go.mod`
- `pip install -r requirements.txt` for `requirements.txt`
- `bundle install` for `Gemfile.lock`

## Breaking Changes

Legacy interface replacements:

- `agentree -b <name>` -> `agentree new <name>`
- `agentree -i` -> `agentree new --interactive`
- `agentree rm agent/<name> -R` -> `agentree rm <name> --deleteBranch`
- `agentree completion <shell>` -> `agentree completions <shell>`
- `.agentreerc` shell config -> `.agentree/config.json`

## Development

```bash
bun run generate
bun run typecheck
bun test
bun run build
bun run build:binaries
```

## License

MIT
