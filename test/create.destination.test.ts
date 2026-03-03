import { describe, expect, test } from 'bun:test'
import {
  defaultDestinationForName,
  resolveInteractiveDestinationDefault
} from '../src/commands/workspace/create.js'

describe('interactive destination defaults', () => {
  test('recomputes destination default from selected workspace name', () => {
    const repoRoot = '/tmp/repo'
    const recomputed = defaultDestinationForName(repoRoot, 'alpha')

    const resolved = resolveInteractiveDestinationDefault(false, '/tmp/repo-worktrees', recomputed)
    expect(resolved).toContain('/tmp/repo-worktrees/alpha')
  })

  test('keeps explicitly provided destination', () => {
    const resolved = resolveInteractiveDestinationDefault(
      true,
      '/custom/worktrees/dest',
      '/tmp/repo-worktrees/alpha'
    )
    expect(resolved).toBe('/custom/worktrees/dest')
  })
})
