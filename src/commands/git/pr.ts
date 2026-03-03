import { defineCommand, option } from '@bunli/core'
import type { HandlerArgs } from '@bunli/core'
import { z } from 'zod'
import { createPullRequest, findWorktreeByTarget, ghExists, resolveRepoRoot } from '../../lib/git.js'
import { resolveConfig } from '../../lib/config.js'
import { printJsonSuccess } from '../../lib/output.js'
import type { GitPrFlags } from '../../lib/types.js'

export default defineCommand({
  name: 'pr',
  description: 'Create a pull request for a workspace',
  options: {
    json: option(z.coerce.boolean().default(false), {
      description: 'Emit JSON output'
    })
  },
  handler: async ({ positional, flags, colors, cwd }: HandlerArgs<GitPrFlags>) => {
    const parsedFlags = flags as GitPrFlags
    const target = positional[0]
    if (!target) {
      throw new Error('Usage: agentree git pr <name|branch|path>')
    }

    const repoRoot = resolveRepoRoot(cwd)
    const { config } = await resolveConfig(repoRoot, {})
    const worktree = findWorktreeByTarget(repoRoot, target, config.branchPrefix)
    if (!worktree) {
      throw new Error(`Worktree not found for target: ${target}`)
    }

    if (!ghExists(worktree.path)) {
      throw new Error('GitHub CLI (gh) is not installed')
    }

    createPullRequest(worktree.path)

    const data = {
      workspace: {
        path: worktree.path,
        branch: worktree.branch ?? null
      },
      created: true
    }

    if (parsedFlags.json) {
      printJsonSuccess('git pr', data)
      return
    }

    console.log(colors.green('Opened pull request flow via gh'))
    console.log(`path: ${worktree.path}`)
  }
})
