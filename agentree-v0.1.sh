#!/usr/bin/env bash
# ----------------------------------------------------------------------------
# agentree — create *or* remove isolated Git worktrees for agentic workflows
# ----------------------------------------------------------------------------
#   agentree -b <name>   → create worktree on agent/<name>
#   agentree rm <branch|path> [-y] [-R]  → remove worktree (and optionally branch)
# ----------------------------------------------------------------------------
# CREATION FLAGS
#   -b <name|branch>   Name after agent/ OR full branch path if it contains '/'
#   -f <base>          Base branch to fork from (default: current HEAD)
#   -p                 Push new branch to origin after creation
#   -r                 Create a GitHub PR via gh CLI after push (implies -p)
#   -d <dest>          Override destination directory
#   -e                 Copy .env and .dev.vars files from source worktree
#   -s                 Run setup scripts (auto-detect or from config)
#   -S <script>        Run custom post-create script (can use multiple times)
#
# REMOVAL FLAGS (after `rm` sub‑command)
#   -y                 No‑prompt force removal (implies --force to Git)
#   -R                 Also delete the local branch after removing worktree
# ----------------------------------------------------------------------------
# INSTALL:
#   chmod +x ~/bin/agentree && echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc
# ----------------------------------------------------------------------------

set -euo pipefail

# ─── Helper ------------------------------------------------------------------
# Load configuration from file
load_config() {
  local config_file="$1"
  [[ -f "$config_file" ]] && source "$config_file"
}

# Auto-detect package manager and return setup commands
auto_detect_setup() {
  local dir="$1"
  local scripts=()
  
  # Node.js package managers
  if [[ -f "$dir/pnpm-lock.yaml" ]]; then
    scripts+=("pnpm install")
    [[ -f "$dir/package.json" ]] && grep -q '"build"' "$dir/package.json" && scripts+=("pnpm build")
  elif [[ -f "$dir/package-lock.json" ]]; then
    scripts+=("npm install")
    [[ -f "$dir/package.json" ]] && grep -q '"build"' "$dir/package.json" && scripts+=("npm run build")
  elif [[ -f "$dir/yarn.lock" ]]; then
    scripts+=("yarn install")
    [[ -f "$dir/package.json" ]] && grep -q '"build"' "$dir/package.json" && scripts+=("yarn build")
  # Other languages
  elif [[ -f "$dir/Cargo.lock" ]]; then
    scripts+=("cargo build")
  elif [[ -f "$dir/go.mod" ]]; then
    scripts+=("go mod download")
  elif [[ -f "$dir/requirements.txt" ]]; then
    scripts+=("pip install -r requirements.txt")
  elif [[ -f "$dir/Gemfile.lock" ]]; then
    scripts+=("bundle install")
  fi
  
  printf '%s\n' "${scripts[@]}"
}

# Run post-create scripts
run_post_scripts() {
  local dest="$1"
  shift
  local scripts=("$@")
  
  if [[ ${#scripts[@]} -eq 0 ]]; then
    return
  fi
  
  echo "🚀  Running post-create scripts..."
  cd "$dest"
  
  for script in "${scripts[@]}"; do
    echo "   → $script"
    if eval "$script"; then
      echo "   ✓ Success"
    else
      echo "   ✗ Failed: $script" >&2
    fi
  done
}

usage() {
  cat <<EOF
Usage:  agentree -b <name> [options]        # create worktree
        agentree rm <branch|path> [flags]  # remove worktree

Creation options:
  -b <name>        required; branch name (prefixes agent/ if no slash)
  -f <base>        base branch (default: current HEAD)
  -p               push branch to origin
  -r               create GitHub PR (implies -p)
  -d <dest>        custom destination path
  -e               copy .env and .dev.vars files from source
  -s               run setup scripts (auto-detect or from config)
  -S <script>      run custom post-create script (repeatable)

Removal flags (after 'rm'):
  -y               no confirmation; force if dirty
  -R               delete the local branch too
EOF
  exit 1
}

confirm() {
  # $1 prompt
  read -r -p "$1 [y/N] " ans; [[ $ans =~ ^[Yy]$ ]];
}

# ─── Sub‑command dispatch -----------------------------------------------------
subcmd="create"
if [[ "${1:-}" == "rm" || "${1:-}" == "remove" ]]; then
  subcmd="remove"; shift; fi

# ──────────────────────────────────────────────────────────────────────────────
if [[ $subcmd == "remove" ]]; then
  # Default removal flags
  force=""; del_branch=false
  # Parse flags: -y (force) -R (delete branch)
  while getopts "yR" opt; do
    case $opt in
      y) force="--force" ;;
      R) del_branch=true ;;
      *) usage ;;
    esac
  done
  shift $((OPTIND-1))
  target="${1:-}"
  [[ -z $target ]] && { echo "❌  agentree rm <branch|path> required" >&2; usage; }

  if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    echo "❌  Not inside a Git repo" >&2; exit 1; fi

  # Determine path and branch
  path=""; branch=""
  if [[ -d $target ]]; then
    path=$(cd "$target" && pwd)
    branch=$(git worktree list --porcelain |
      awk -v p="$path" '$1=="worktree" && $2==p {getline; if($1=="branch") {sub("refs/heads/","",$2); print $2}}')
  else
    branch="$target"
    path=$(git worktree list --porcelain |
      awk -v b="$branch" '$1=="worktree" {p=$2} $1=="branch" {sub("refs/heads/","",$2); if($2==b) print p}')
  fi

  [[ -z $path ]] && { echo "❌  Worktree not found for $target" >&2; exit 1; }

  # Confirm if not forced
  if [[ -z $force ]] && ! confirm "Remove worktree at $path?"; then exit 0; fi

  git worktree remove $force "$path"
  echo "✅  Removed worktree $path"

  if $del_branch && [[ -n $branch ]]; then
    git branch -D "$branch" || true
    echo "🗑️   Deleted branch $branch"
  fi
  exit 0
fi

# ─── Creation path -----------------------------------------------------------
# Defaults
branch=""; base=""; push=false; pr=false; custom_dest=""; copy_env=false
run_setup=false; custom_scripts=()

# Load global config
load_config "$HOME/.config/agentree/config"

while getopts "b:f:prd:esS:h" opt; do
  case $opt in
    b) branch="$OPTARG" ;;
    f) base="$OPTARG" ;;
    p) push=true ;;
    r) pr=true; push=true ;;
    d) custom_dest="$OPTARG" ;;
    e) copy_env=true ;;
    s) run_setup=true ;;
    S) custom_scripts+=("$OPTARG") ;;
    h|*) usage ;;
  esac
done

[[ -z $branch ]] && { echo "❌  -b <name> is required" >&2; usage; }

# Prefix agent/ if no slash
if [[ $branch != */* ]]; then
  branch="agent/$branch"
fi

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "❌  Run this inside a Git repository" >&2; exit 1; fi

git fetch --prune

root=$(git rev-parse --show-toplevel)
repo=$(basename "$root")
parent=$(dirname "$root")
workdir="${parent}/${repo}-worktrees"
mkdir -p "$workdir"

# Resolve base
if [[ -z $base ]]; then
  base=$(git symbolic-ref --quiet --short HEAD || git rev-parse --short HEAD)
fi

san_branch="${branch//\//-}"
dest="${custom_dest:-${workdir}/${san_branch}}"

if [[ -d $dest ]]; then
  echo "❌  Destination $dest exists" >&2; exit 1; fi
if git show-ref --verify --quiet "refs/heads/$branch"; then
  echo "❌  Branch $branch already exists" >&2; exit 1; fi

# Create branch + worktree

git branch "$branch" "$base"
git worktree add "$dest" "$branch"

echo "✅  Worktree ready:"
echo "    path   $dest"
echo "    branch $branch (from $base)"

# Copy env files if requested
if $copy_env; then
  for env_file in .env .dev.vars; do
    if [[ -f "$root/$env_file" ]]; then
      cp "$root/$env_file" "$dest/$env_file"
      echo "📋  Copied $env_file"
    fi
  done
fi

# Run post-create scripts
if [[ ${#custom_scripts[@]} -gt 0 ]]; then
  # Custom scripts take precedence
  run_post_scripts "$dest" "${custom_scripts[@]}"
elif $run_setup; then
  # Load project config if exists
  POST_CREATE_SCRIPTS=()
  load_config "$root/.agentreerc"
  
  if [[ ${#POST_CREATE_SCRIPTS[@]} -gt 0 ]]; then
    # Use project-specific scripts
    run_post_scripts "$dest" "${POST_CREATE_SCRIPTS[@]}"
  else
    # Auto-detect setup commands
    mapfile -t detected_scripts < <(auto_detect_setup "$dest")
    
    # Check for user overrides in global config
    if [[ ${#detected_scripts[@]} -gt 0 ]]; then
      # Check for package manager specific overrides
      if [[ "${detected_scripts[0]}" == "pnpm install" ]] && [[ -n "${PNPM_SETUP:-}" ]]; then
        run_post_scripts "$dest" "$PNPM_SETUP"
      elif [[ "${detected_scripts[0]}" == "npm install" ]] && [[ -n "${NPM_SETUP:-}" ]]; then
        run_post_scripts "$dest" "$NPM_SETUP"
      elif [[ "${detected_scripts[0]}" == "yarn install" ]] && [[ -n "${YARN_SETUP:-}" ]]; then
        run_post_scripts "$dest" "$YARN_SETUP"
      elif [[ -n "${DEFAULT_POST_CREATE:-}" ]]; then
        run_post_scripts "$dest" "$DEFAULT_POST_CREATE"
      else
        run_post_scripts "$dest" "${detected_scripts[@]}"
      fi
    elif [[ -n "${DEFAULT_POST_CREATE:-}" ]]; then
      run_post_scripts "$dest" "$DEFAULT_POST_CREATE"
    fi
  fi
fi

if $push; then
  git -C "$dest" push -u origin "$branch"
fi
if $pr; then
  if command -v gh >/dev/null 2>&1; then
    gh -C "$dest" pr create --fill --web
  else
    echo "⚠️  gh CLI not found; skipping PR creation" >&2
  fi
fi

