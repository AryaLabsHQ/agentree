#!/bin/bash
# Demo script for agentree multiplexer

set -e

echo "🌳 Agentree Multiplexer Demo"
echo "============================"
echo ""

# Check if agentree is built
if [ ! -f "agentree" ]; then
    echo "Building agentree..."
    make build
fi

# Create demo worktrees if they don't exist
echo "Setting up demo worktrees..."

# Check if we're in a git repo
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "Error: Not in a git repository"
    exit 1
fi

# Create demo branches and worktrees
demo_branches=("demo-feat-ui" "demo-feat-api" "demo-feat-tests")
demo_dir="../agentree-demo-worktrees"

for branch in "${demo_branches[@]}"; do
    if ! git show-ref --verify --quiet "refs/heads/$branch"; then
        echo "Creating branch: $branch"
        git branch "$branch" HEAD
    fi
    
    worktree_path="$demo_dir/$branch"
    if [ ! -d "$worktree_path" ]; then
        echo "Creating worktree: $worktree_path"
        git worktree add "$worktree_path" "$branch"
    fi
done

echo ""
echo "Demo worktrees created:"
git worktree list | grep demo-

# Create a demo configuration
cat > .agentree-multiplex-demo.yml << 'EOF'
# Demo multiplexer configuration
auto_start: true
token_limit: 10000
theme: dark

instances:
  - worktree: ../agentree-demo-worktrees/demo-feat-ui
    auto_start: true
    token_limit: 3000
    environment:
      - MOCK_CLAUDE_PATH=./mock-claude
  
  - worktree: ../agentree-demo-worktrees/demo-feat-api
    auto_start: true
    token_limit: 3000
    environment:
      - MOCK_CLAUDE_PATH=./mock-claude
  
  - worktree: ../agentree-demo-worktrees/demo-feat-tests
    auto_start: false
    token_limit: 4000
    environment:
      - MOCK_CLAUDE_PATH=./mock-claude

ui:
  sidebar_width: 30
  show_token_usage: true
  show_timestamps: false
EOF

echo ""
echo "Demo configuration created: .agentree-multiplex-demo.yml"
echo ""
echo "Launching multiplexer with demo worktrees..."
echo "(Using mock-claude for demonstration)"
echo ""
echo "Press 'q' to quit, '?' for help"
echo ""
sleep 2

# Run the multiplexer with demo config
./agentree multiplex --config .agentree-multiplex-demo.yml

# Cleanup
echo ""
echo "Demo completed!"
echo ""
echo "To clean up demo worktrees, run:"
echo "  git worktree remove ../agentree-demo-worktrees/demo-feat-ui"
echo "  git worktree remove ../agentree-demo-worktrees/demo-feat-api"
echo "  git worktree remove ../agentree-demo-worktrees/demo-feat-tests"
echo "  git branch -D demo-feat-ui demo-feat-api demo-feat-tests"
echo "  rm .agentree-multiplex-demo.yml"