# agentree Roadmap 🌳🤖

## Vision

agentree aims to be the definitive tool for managing isolated development environments optimized for AI coding agents. By treating each worktree as a complete, context-aware development environment, we enable developers to work seamlessly with multiple AI assistants on concurrent tasks without conflicts or context pollution.

## Core Value Proposition

While Git worktrees provide branch isolation, they lack critical features for modern AI-assisted development:
- Environment configuration (.env files) aren't copied
- No setup automation for dependencies
- Missing context management for AI agents
- No workflow optimization for agent handoffs

agentree bridges this gap by providing environment-aware worktree management specifically designed for AI coding workflows.

## Key Integration: Claude Code + GitHub Actions

A primary focus of agentree is enabling powerful automation workflows with Claude Code through GitHub Actions. This combination allows:

- **Automated Code Reviews**: Trigger Claude Code to review PRs with full project context
- **Test Generation**: Automatically generate tests for new features
- **Documentation Updates**: Keep docs in sync with code changes
- **Refactoring Tasks**: Schedule large-scale refactoring with proper isolation
- **Bug Fixes**: Auto-create worktrees for issue resolution

Example workflow:
```yaml
on:
  issues:
    types: [opened, labeled]

jobs:
  auto-fix:
    if: contains(github.event.label.name, 'claude-fix')
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: aryalabs/agentree-action@v1
        with:
          agent: claude
          task: fix-issue
          issue: ${{ github.event.issue.number }}
```

## Roadmap

### Phase 1: Enhanced Environment Management

#### Expanded File Support
- Support additional environment files:
  - `.env.local`
  - `.env.development` 
  - `.env.test`
  - `.env.production`
  - `docker-compose.env`
  - `.env.*.local` patterns
- Configuration for custom env file patterns
- Selective copying based on worktree purpose

#### Template Variables
- Support placeholders in environment files:
  - `${AGENT_NAME}` - Current AI agent (claude, cursor, copilot)
  - `${TASK_ID}` - Issue/ticket number
  - `${BRANCH_NAME}` - Current branch
  - `${TIMESTAMP}` - Creation timestamp
- Template resolution during worktree creation
- Custom variable definitions

#### Secrets Integration
- Integrate with popular secret managers:
  - 1Password CLI
  - AWS Secrets Manager
  - HashiCorp Vault
  - Doppler
  - Azure Key Vault
- Secure secret injection into env files
- Temporary credential management

#### Environment Validation
- Verify required environment variables are set
- Validate API keys and tokens
- Check database connectivity
- Test external service availability
- Pre-flight checks before agent starts work

#### Smart Environment Diffing
- Show environment variable differences from main branch
- Highlight new, modified, or removed variables
- Environment migration suggestions
- Conflict detection and resolution

#### CLI Autocomplete
- Shell completion support for:
  - Bash
  - Zsh
  - Fish
  - PowerShell
- Dynamic completions for:
  - Branch names from current repo
  - Existing worktree names
  - Available base branches
  - Configuration keys
  - Agent types
- Installation helpers for each shell type
- Context-aware suggestions based on current command

### Phase 2: AI Agent-Specific Features

#### Native Claude Code Integration (Primary Focus)
- First-class support for Claude Code:
  - Auto-generate `.claude.md` with project context
  - GitHub Actions integration for automated workflows
  - Claude-optimized environment setup
  - Context preservation between sessions
  - Direct API integration for headless operations
- GitHub Actions recipes:
  - PR review automation
  - Test generation
  - Documentation updates
  - Code refactoring tasks

#### Cursor Support (Secondary)
- Built-in support for Cursor:
  - `.cursorrules` generation
  - Cursor-specific environment setup
  - Context file management
- Seamless handoff between Claude Code and Cursor

#### Plugin Architecture for Other Agents
- Generic adapter interface for community contributions:
  ```go
  type AgentAdapter interface {
    GenerateContext(worktree *Worktree) error
    SetupEnvironment(worktree *Worktree) error
    GetBranchPrefix() string
    ValidateSetup() error
  }
  ```
- Plugin discovery and loading
- Documentation for adapter development
- Example adapters as reference implementations

#### Context Provisioning
- Auto-generate context files based on agent type
- Pull context from:
  - README files
  - Documentation
  - Recent commits
  - Issue descriptions
- Dynamic context updates
- Context inheritance from main branch

#### Smart Branch Naming
- Agent-aware branch prefixes:
  - `agent/claude/feature-name`
  - `agent/cursor/bugfix-123`
  - Custom prefixes via plugins
- Automatic ticket ID extraction
- Customizable naming patterns

#### Agent Handoffs
- Export current context and progress
- Import context when switching agents
- Progress tracking between agents
- Handoff notes and summaries
- Maintain conversation history

#### Usage Analytics
- Track per-agent metrics:
  - Time spent
  - Files modified
  - Commits created
  - Commands run
- Generate reports for:
  - Productivity analysis
  - Cost optimization
  - Agent effectiveness

### Phase 3: Advanced Setup & Dependencies

#### Shared Resource Management
- Symlink shared directories:
  - `node_modules`
  - `vendor`
  - `.venv`
  - `target/`
  - Build caches
- Copy-on-write for modifications
- Space usage optimization
- Faster worktree creation

#### Parallel Setup Execution
- Analyze dependency graph
- Run independent steps concurrently
- Progress visualization
- Intelligent retry logic
- Setup time optimization

#### Health Checks
- Post-setup verification:
  - Database connections
  - API endpoint availability
  - Service dependencies
  - File permissions
  - Tool versions
- Automated fixes for common issues
- Health status dashboard

#### Setup Profiles
- Pre-defined setup modes:
  - **Minimal**: Just essentials
  - **Full**: Complete environment
  - **Test**: Testing dependencies only
  - **Development**: Dev tools included
  - **Production**: Prod-like environment
- Custom profile creation
- Profile inheritance

#### Container Integration
- Docker Compose support per worktree
- Isolated database instances
- Service isolation
- Container lifecycle management
- Resource limits per worktree

### Phase 4: Workflow Automation

#### Task Templates
- Pre-configured templates for:
  - Bug fixes
  - Feature development
  - Refactoring
  - Documentation
  - Testing
- Template marketplace
- Organization-specific templates

#### Issue Integration
- Auto-link to:
  - GitHub Issues
  - Jira tickets
  - Linear issues
  - Asana tasks
- Pull issue context into branch
- Update issue status automatically

#### Session Recording
- Track development sessions:
  - File changes
  - Commands executed
  - Test results
  - Build outputs
- Generate session summaries
- Knowledge base creation
- Replay capabilities

#### Smart Cleanup
- Auto-remove worktrees after:
  - PR merge
  - Inactivity timeout
  - Failed builds
  - Abandoned work
- Archive before deletion
- Cleanup policies

### Phase 5: Developer Experience

#### TUI Dashboard
- Real-time worktree overview:
  - Active worktrees
  - Resource usage
  - Agent activity
  - Health status
- Quick actions:
  - Create new worktree
  - Switch between worktrees
  - View logs
  - Cleanup

#### Plugin System for IDE Integration
- Core CLI remains IDE-agnostic
- Plugin interface for IDE extensions:
  - Standardized API for worktree operations
  - Event hooks for IDE notifications
  - Context file generation
- Community-driven integrations:
  - Reference implementation for VS Code
  - Documentation and examples
  - Plugin registry/marketplace

#### Worktree Sync
- Keep files synchronized:
  - Configuration changes
  - Documentation updates
  - Schema changes
- Selective sync rules
- Conflict resolution
- Sync history

#### Time Tracking
- Built-in time tracking:
  - Per worktree
  - Per task
  - Per agent
- Integration with:
  - Toggl
  - Harvest
  - Clockify
- Automatic tracking
- Time reports

## Implementation Priority

1. **High Priority** (Next 3 months)
   - Native Claude Code integration
   - GitHub Actions recipes for Claude Code
   - Expanded environment file support
   - CLI autocomplete
   - Plugin architecture foundation
   - Shared node_modules linking
   - Health checks

2. **Medium Priority** (3-6 months)
   - Cursor support
   - Template variables
   - Context provisioning system
   - Plugin development documentation
   - Setup profiles
   - Basic TUI dashboard

3. **Low Priority** (6-12 months)
   - Community plugin marketplace
   - Container integration
   - Session recording
   - Advanced analytics
   - Complex workflow automation

## Success Metrics

- Worktree creation time < 30 seconds
- 90% reduction in environment setup issues
- 50% improvement in agent context quality
- 80% of users using agent-specific features
- 95% successful environment validations

## Community Involvement

We welcome community contributions in:
- Agent profile definitions
- Template creation
- Integration development
- Documentation
- Testing and feedback

## Conclusion

agentree will evolve from a simple worktree wrapper to a comprehensive AI development environment manager, making it indispensable for teams working with AI coding assistants. By focusing on the unique needs of AI-assisted development, we'll create a tool that significantly improves developer productivity and AI agent effectiveness.