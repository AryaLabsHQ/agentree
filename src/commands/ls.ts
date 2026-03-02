import { defineCommand, option } from '@bunli/core'
import { z } from 'zod'
import { basename } from 'node:path'
import { listWorktrees, resolveRepoRoot } from '../lib/git.js'
import { printJson, printWorkspacesTable } from '../lib/output.js'
import type { ManagedWorkspace } from '../lib/types.js'

const DEFAULT_BRANCH_PREFIX = 'agent/'

export default defineCommand({
  name: 'ls',
  description: 'List workspaces',
  options: {
    json: option(z.coerce.boolean().default(false), {
      description: 'Emit JSON output'
    }),
    all: option(z.coerce.boolean().default(false), {
      description: 'Include non-managed worktrees'
    })
  },
  handler: async ({ flags, cwd }) => {
    const repoRoot = resolveRepoRoot(cwd)
    const worktrees = listWorktrees(repoRoot)

    const mapped: ManagedWorkspace[] = worktrees.map((wt) => {
      const managed = Boolean(wt.branch?.startsWith(DEFAULT_BRANCH_PREFIX))
      const name = wt.branch?.startsWith(DEFAULT_BRANCH_PREFIX)
        ? wt.branch.slice(DEFAULT_BRANCH_PREFIX.length)
        : basename(wt.path)

      return {
        ...wt,
        name,
        managed
      }
    })

    const includeAll = flags.all as boolean
    const filtered = includeAll ? mapped : mapped.filter((w) => w.managed)

    if (flags.json) {
      printJson({ ok: true, workspaces: filtered })
      return
    }

    printWorkspacesTable(filtered)
  }
})
