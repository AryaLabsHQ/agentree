# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Build and Development
- `make build` - Build the binary
- `make run` - Build and run the application
- `make install` - Install to GOPATH/bin

### Testing
- `make test` - Run all tests
- `make test-coverage` - Run tests with coverage report
- `make test-integration` - Run integration tests
- `go test -v ./cmd -run TestCreateCommand` - Run a specific test

### Code Quality
- `make lint` - Run golangci-lint
- `make fmt` - Format code with go fmt

### Dependencies
- `make deps` - Update and download dependencies

## Architecture

agentree is a Go CLI tool that creates isolated Git worktrees for AI coding agents. Key architectural components:

### Command Structure (Cobra)
- `cmd/root.go` - Base command configuration
- `cmd/create.go` - Worktree creation logic with flags for branch, base, push, PR, env copying, and setup
- `cmd/remove.go` - Worktree removal with optional branch deletion

### Core Packages
- `internal/git/` - Git operations wrapper around worktree management
- `internal/detector/` - Auto-detects package managers (pnpm, npm, yarn, cargo, go, pip, bundler)
- `internal/tui/` - Interactive branch selector using Bubble Tea framework
- `internal/config/` - Configuration management for project (.agentreerc) and global settings
- `internal/scripts/` - Script execution engine for post-create commands
- `internal/env/` - Environment file operations (.env, .dev.vars copying)

### Key Patterns
- All Git operations go through `git.Repository` type for consistency
- Package manager detection uses file existence checks (package.json, Cargo.toml, etc.)
- TUI components use the Bubble Tea architecture (Model, Update, View)
- Configuration precedence: CLI flags > project config > global config > defaults

### Testing Approach
- Unit tests alongside implementation files (*_test.go)
- Integration tests in `cmd/integration_test.go` using build tags
- Mock Git repository for testing Git operations without real repos