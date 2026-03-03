import { describe, expect, test } from 'bun:test'
import { defaultWorktreePath, deriveBranchName, validateWorkspaceName } from '../src/lib/worktree.js'

describe('worktree utils', () => {
  test('derive branch prefix', () => {
    expect(deriveBranchName('agent/', 'feature-x')).toBe('agent/feature-x')
    expect(deriveBranchName('agent/', 'agent/already')).toBe('agent/already')
  })

  test('default worktree path', () => {
    const dest = defaultWorktreePath('/tmp/repo', 'alpha')
    expect(dest).toContain('/tmp/repo-worktrees/alpha')
  })

  test('workspace name validation', () => {
    expect(() => validateWorkspaceName('ok-name')).not.toThrow()
    expect(() => validateWorkspaceName('')).toThrow()
    expect(() => validateWorkspaceName('a/b')).toThrow()
    expect(() => validateWorkspaceName('.')).toThrow()
  })
})
