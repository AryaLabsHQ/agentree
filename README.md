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
agentree workspace create fix-auth
# alias
agentree new fix-auth
```

List workspaces:

```bash
agentree workspace list
# alias
agentree ls
```

Remove a workspace:

```bash
agentree workspace remove fix-auth --yes --deleteBranch
# alias
agentree rm fix-auth --yes --deleteBranch
```

Push and open PR explicitly:

```bash
agentree git push fix-auth
agentree git pr fix-auth
```

## Commands

Primary grouped commands:

```bash
agentree workspace create <name> [--from <ref>] [--dest <path>] [--interactive] [--json]
agentree workspace list [--all] [--json]
agentree workspace remove <name|branch|path> [--yes] [--deleteBranch] [--json]
agentree git push <name|branch|path> [--remote <name>] [--json]
agentree git pr <name|branch|path> [--json]
```

Top-level aliases:

```bash
agentree new <name>
agentree ls
agentree rm <name|branch|path>
```

Shell completions:

```bash
agentree completions <bash|zsh|fish|powershell>
```

## JSON Output Contract

Every `--json` response uses a versioned envelope:

- `schemaVersion`
- `ok`
- `command`
- `data`
- `warnings`
- `error` (only when `ok=false`)

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

`workspace list` and `workspace remove` always use merged `branchPrefix` from this config chain.

## Remove Target Resolution

`workspace remove <target>` resolves in deterministic order:

1. Workspace name
2. Full branch name
3. Filesystem path

This prevents path collisions from overriding a workspace-name match.

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

Setup command execution is platform-aware (`sh -lc` on Unix-like platforms, `cmd.exe /d /s /c` on Windows).

## Breaking Changes

- `workspace create` no longer supports `--push` / `--pr` side effects.
- Push and PR are explicit commands: `git push` and `git pr`.
- Grouped command model is primary (`workspace *`, `git *`).
- Top-level `new|ls|rm` remain aliases for workspace lifecycle.

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
