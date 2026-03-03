import { existsSync, realpathSync } from 'node:fs'
import { resolve } from 'node:path'
import { CommandExecutionError } from './errors.js'
import type { WorktreeInfo } from './types.js'

interface RunOptions {
  cwd: string
  allowFailure?: boolean
}

interface RunResult {
  stdout: string
  stderr: string
  exitCode: number
}

function run(command: string, args: string[], options: RunOptions): RunResult {
  const proc = Bun.spawnSync({
    cmd: [command, ...args],
    cwd: options.cwd,
    stdout: 'pipe',
    stderr: 'pipe'
  })

  const stdout = Buffer.from(proc.stdout).toString('utf8')
  const stderr = Buffer.from(proc.stderr).toString('utf8')
  const exitCode = proc.exitCode ?? 1

  if (exitCode !== 0 && !options.allowFailure) {
    throw new CommandExecutionError(command, args, exitCode, stdout, stderr)
  }

  return { stdout, stderr, exitCode }
}

export function resolveRepoRoot(cwd: string): string {
  const { stdout } = run('git', ['rev-parse', '--show-toplevel'], { cwd })
  return stdout.trim()
}

export function currentBaseRef(repoRoot: string): string {
  const symbolic = run('git', ['symbolic-ref', '--quiet', '--short', 'HEAD'], {
    cwd: repoRoot,
    allowFailure: true
  })

  if (symbolic.exitCode === 0 && symbolic.stdout.trim()) {
    return symbolic.stdout.trim()
  }

  const { stdout } = run('git', ['rev-parse', '--short', 'HEAD'], { cwd: repoRoot })
  return stdout.trim()
}

export function branchExists(repoRoot: string, branch: string): boolean {
  const result = run('git', ['show-ref', '--verify', '--quiet', `refs/heads/${branch}`], {
    cwd: repoRoot,
    allowFailure: true
  })
  return result.exitCode === 0
}

export function createWorktree(repoRoot: string, branch: string, destination: string, baseRef: string): void {
  run('git', ['worktree', 'add', '-b', branch, destination, baseRef], { cwd: repoRoot })
}

export function removeWorktree(repoRoot: string, destination: string, force = false): void {
  const args = ['worktree', 'remove']
  if (force) args.push('--force')
  args.push(destination)
  run('git', args, { cwd: repoRoot })
}

export function deleteBranch(repoRoot: string, branch: string): void {
  run('git', ['branch', '-D', branch], { cwd: repoRoot })
}

export function pushBranch(repoRoot: string, remote: string, branch: string): void {
  run('git', ['push', '-u', remote, branch], { cwd: repoRoot })
}

export function ghExists(repoRoot: string): boolean {
  const result = run('gh', ['--version'], { cwd: repoRoot, allowFailure: true })
  return result.exitCode === 0
}

export function createPullRequest(repoRoot: string): void {
  run('gh', ['pr', 'create', '--fill', '--web'], { cwd: repoRoot })
}

export function listWorktrees(repoRoot: string): WorktreeInfo[] {
  const { stdout } = run('git', ['worktree', 'list', '--porcelain'], { cwd: repoRoot })
  const blocks = stdout
    .split(/\n\n+/)
    .map((b) => b.trim())
    .filter(Boolean)

  const worktrees: WorktreeInfo[] = []

  for (const block of blocks) {
    const lines = block.split('\n')
    let path = ''
    let head: string | undefined
    let branch: string | undefined
    let detached = false

    for (const line of lines) {
      if (line.startsWith('worktree ')) {
        path = line.slice('worktree '.length)
      } else if (line.startsWith('HEAD ')) {
        head = line.slice('HEAD '.length)
      } else if (line.startsWith('branch ')) {
        const ref = line.slice('branch '.length)
        branch = ref.startsWith('refs/heads/') ? ref.slice('refs/heads/'.length) : ref
      } else if (line === 'detached') {
        detached = true
      }
    }

    if (path) {
      worktrees.push({ path, head, branch, detached })
    }
  }

  return worktrees
}

function normalizePath(path: string): string {
  const absolute = resolve(path)
  if (existsSync(absolute)) {
    return realpathSync(absolute)
  }
  return absolute
}

function branchForWorkspaceName(branchPrefix: string, workspaceName: string): string {
  return `${branchPrefix}${workspaceName}`
}

function parseWorkspaceName(target: string, branchPrefix: string): string | null {
  if (target.startsWith(branchPrefix)) {
    const name = target.slice(branchPrefix.length)
    return name.length > 0 ? name : null
  }

  if (!target.includes('/')) {
    return target
  }

  return null
}

// Deterministic resolution order: name -> branch -> path.
export function findWorktreeByTarget(
  repoRoot: string,
  target: string,
  branchPrefix: string
): WorktreeInfo | undefined {
  const worktrees = listWorktrees(repoRoot)

  const workspaceName = parseWorkspaceName(target, branchPrefix)
  if (workspaceName) {
    const nameBranch = branchForWorkspaceName(branchPrefix, workspaceName)
    const byName = worktrees.find((wt) => wt.branch === nameBranch)
    if (byName) {
      return byName
    }
  }

  const byBranch = worktrees.find((wt) => wt.branch === target)
  if (byBranch) {
    return byBranch
  }

  if (existsSync(target)) {
    const normalizedTarget = normalizePath(target)
    return worktrees.find((wt) => normalizePath(wt.path) === normalizedTarget)
  }

  return undefined
}
