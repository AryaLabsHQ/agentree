# Multiplexer Demo Scenarios

## Scenario 1: Parallel Feature Development

```bash
# Start multiplexer with specific worktrees
$ agentree mx auth ui api

# UI shows:
┌─ agentree multiplexer ─────────────────────────────────────────┐
│ Worktrees          │ agent/feat-authentication                  │
│ ──────────────     │ Starting Claude Code instance...           │
│ ● auth       0.0k  │ Initializing in /path/to/auth              │
│ ● ui         0.0k  │ Ready.                                     │
│ ● api        0.0k  │                                            │
└────────────────────┴────────────────────────────────────────────┘
```

### Developer Workflow:
1. **Task Assignment**: "Auth agent, implement login system. UI agent, create login form. API agent, set up JWT endpoints."
2. **Agents work in parallel**, each in their worktree
3. **Real-time monitoring** of progress and token usage
4. **Coordination**: Agents can see each other's changes

## Scenario 2: Bug Hunt Across Codebase

```bash
# Launch with all worktrees
$ agentree mx --all

# Assign investigation task to all agents
> broadcast "Find all instances of user authentication and check for security issues"

# Agents work simultaneously:
- Agent 1: Scanning frontend code...
- Agent 2: Reviewing API endpoints...
- Agent 3: Checking database queries...
- Agent 4: Analyzing middleware...
```

## Scenario 3: Refactoring with Coordination

```bash
# Start specific agents for refactoring
$ agentree mx main refactor-branch

# Use coordination commands:
> agent main "Review the current user service implementation"
> agent refactor-branch "Wait for main's analysis, then refactor the user service"
> sync-context  # Share findings between agents
```

## Interactive Commands

While multiplexer is running:

```
:task auth "Implement OAuth2 flow"          # Assign task to specific agent
:broadcast "Update all imports"             # Send to all agents
:pause ui                                   # Pause specific agent
:limit auth 10000                          # Set token limit
:sync                                      # Sync context between agents
:export-usage                              # Export token usage report
```

## Advanced Features Demo

### 1. Load Balancing
```
# Multiplexer automatically distributes tasks based on token usage
> smart-assign "Implement test suite for all components"

Auth agent (2.3k tokens used): Writing auth tests...
UI agent (5.1k tokens used): [Waiting - higher usage]
API agent (1.2k tokens used): Writing API tests...
```

### 2. Shared Context
```
# Agents share discoveries
Auth agent: "Found security issue in login endpoint"
> share-finding auth all

UI agent: "I see the security issue. Updating form validation..."
API agent: "I'll add rate limiting to address this..."
```

### 3. Progress Tracking
```
┌─ Task Progress ─────────────────────────────────────┐
│ Overall: ████████████░░░░░░ 65%                    │
│                                                     │
│ auth:    ██████████████████ 90% (Login complete)   │
│ ui:      ████████░░░░░░░░░░ 40% (Working on form)  │
│ api:     ██████████████░░░░ 70% (JWT implemented)  │
└─────────────────────────────────────────────────────┘
```

## Benefits

1. **10x Faster Development**: Multiple features developed simultaneously
2. **Better Resource Usage**: Distribute token usage across instances
3. **Reduced Context Switching**: Each agent maintains its own context
4. **Real-time Collaboration**: Agents can coordinate on complex tasks
5. **Comprehensive Testing**: Run tests in parallel across different scenarios

## Example Output

```
┌─ agentree multiplexer ──────────────────────────────────────────────┐
│ Worktrees          │ agent/feat-authentication [23m 45s]            │
│ ──────────────     │ ╭─ Claude ────────────────────────────────────╮ │
│ ● auth      23.5k  │ │ I've implemented the login system with:     │ │
│ ● ui        18.2k  │ │ - JWT token generation                      │ │
│ ● api       15.7k  │ │ - Secure password hashing (bcrypt)          │ │
│ ○ main       2.1k  │ │ - Rate limiting on login attempts           │ │
│                    │ │ - Session management                        │ │
│ Summary            │ ╰──────────────────────────────────────────────╯ │
│ ──────────────     │                                                  │
│ Total:      59.5k  │ > npm test -- --coverage                        │
│ Cost:       $2.38  │ PASS  src/auth/login.test.js                    │
│ Time:       45min  │ PASS  src/auth/jwt.test.js                      │
│ Status:   3 active │ Coverage: 94.2% (statements)                    │
└────────────────────┴──────────────────────────────────────────────────┘
 [AUTH] j/k:nav ↵:focus TAB:next q:quit r:restart ?:help │ 3 agents
```

## Configuration Example

```yaml
# ~/.config/agentree/multiplex.yml
defaults:
  auto_start: true
  token_limit: 100000
  theme: dark
  
coordination:
  shared_context: true
  sync_interval: 30s
  
instances:
  - worktree: main
    role: "coordinator"
    auto_start: true
    
  - worktree: "feat-*"
    role: "developer"
    token_limit: 25000
    
shortcuts:
  next_agent: "TAB"
  broadcast: "b"
  sync: "s"
  
smart_features:
  load_balancing: true
  auto_pause_on_limit: true
  crash_recovery: true
```

This multiplexer will revolutionize how developers work with AI agents, enabling true parallel development!