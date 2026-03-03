import { basename, dirname, join } from 'node:path'

export function validateWorkspaceName(name: string): void {
  if (!name || name.trim().length === 0) {
    throw new Error('Workspace name is required')
  }

  if (name === '.' || name === '..') {
    throw new Error('Workspace name cannot be . or ..')
  }

  if (name.includes('/') || name.includes('\\')) {
    throw new Error('Workspace name must not contain path separators')
  }
}

export function deriveBranchName(prefix: string, workspaceName: string): string {
  if (workspaceName.startsWith(prefix)) {
    return workspaceName
  }
  return `${prefix}${workspaceName}`
}

export function defaultWorktreePath(repoRoot: string, workspaceName: string): string {
  const parent = dirname(repoRoot)
  const repoName = basename(repoRoot)
  return join(parent, `${repoName}-worktrees`, workspaceName)
}

export function normalizeWorkspaceNameFromBranch(branch: string, prefix: string): string {
  if (branch.startsWith(prefix)) {
    return branch.slice(prefix.length)
  }
  return branch
}
