# Environment Configuration

`agentree` copies ignored artifacts from the source repository into a new workspace, with configurable include and exclude behavior.

## How It Works

1. Lists ignored files with `git ls-files --others -i --exclude-standard`.
2. Applies excludes from config (`defaultExcludes` + `extraExcludes`).
3. Re-includes matching paths from `extraIncludes`.
4. Copies remaining files to the destination worktree.
5. Skips destination collisions (`skip-existing` behavior).

## Config Files

Project config:

```text
.agentree/config.json
```

Global config:

```text
~/.config/agentree/config.json
```

## Config Keys

```json
{
  "copyIgnoredEnabled": true,
  "defaultExcludes": [
    ".git/",
    "node_modules/",
    ".venv/",
    "vendor/bundle/"
  ],
  "extraIncludes": [".env", ".dev.vars"],
  "extraExcludes": ["*.tmp", ".cache/"]
}
```

## CLI Overrides

Include extra patterns for a single run:

```bash
agentree new feature-a --include=.env,.dev.vars
```

Exclude extra patterns for a single run:

```bash
agentree new feature-a --exclude=.cache/,*.tmp
```

Disable copying for a single run:

```bash
agentree new feature-a --copyIgnored=false
```

## Examples

Basic create:

```bash
agentree new feature/new-api
```

Create with JSON output:

```bash
agentree new feature/new-api --json
```

## Troubleshooting

No files copied:

- Confirm files are ignored by git.
- Confirm `copyIgnoredEnabled` is true.
- Confirm excludes are not filtering your target file.

Unexpected files copied:

- Add patterns to `extraExcludes`.
- Remove broad patterns from `extraIncludes`.
