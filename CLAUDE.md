# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working in this repository.

## Commands

Build and generation:

- `bun run generate` - Generate Bunli command metadata in `.bunli/commands.gen.ts`
- `bun run build` - Generate metadata and build JS output in `dist/`
- `bun run build:binaries` - Build standalone binaries for multiple targets

Testing and checks:

- `bun run typecheck` - TypeScript checks (`tsc --noEmit`)
- `bun test` - Run test suite
- `bun run check` - Typecheck + tests

## Architecture

`agentree` is a Bun + TypeScript CLI built with Bunli.

Entry point:

- `src/cli.ts`

Command groups:

- `src/commands/workspace.ts` - Workspace lifecycle group
- `src/commands/git.ts` - Git actions for workspaces

Workspace commands:

- `src/commands/workspace/create.ts`
- `src/commands/workspace/list.ts`
- `src/commands/workspace/remove.ts`

Git commands:

- `src/commands/git/push.ts`
- `src/commands/git/pr.ts`

Top-level aliases:

- `src/commands/new.ts`
- `src/commands/ls.ts`
- `src/commands/rm.ts`

Core libraries:

- `src/lib/git.ts` - Git and GitHub CLI command wrappers
- `src/lib/config.ts` - Global + project JSON config resolution with flag overrides
- `src/lib/copyIgnored.ts` - Gitignored artifact copy engine with include/exclude logic
- `src/lib/setup.ts` - Setup command auto-detection and execution
- `src/lib/worktree.ts` - Workspace name/path/branch helpers
- `src/lib/output.ts` - Versioned JSON envelope and table output
- `src/lib/errors.ts` - Structured command and rollback errors
- `src/lib/types.ts` - Shared types

## Testing Approach

- Unit tests in `test/*.test.ts`
- Integration tests for workspace command behavior and rollback semantics
- Tests use temporary git repositories and Bun test runner

## Current CLI Surface

- `agentree workspace create|list|remove`
- `agentree git push|pr`
- `agentree new|ls|rm` (top-level aliases)
- `agentree completions`
