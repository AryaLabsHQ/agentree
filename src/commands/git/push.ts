import { defineCommand, option } from '@bunli/core'
import type { HandlerArgs } from '@bunli/core'
import { z } from 'zod'
import { findWorktreeByTarget, pushBranch, resolveRepoRoot } from '../../lib/git.js'
import { resolveConfig } from '../../lib/config.js'
import { printJsonSuccess } from '../../lib/output.js'
import type { GitPushFlags } from '../../lib/types.js'

export default defineCommand({
  name: 'push',
  description: 'Push a workspace branch to remote',
  options: {
    remote: option(z.string().default('origin'), {
      short: 'r',
      description: 'Remote name to push to'
    }),
    json: option(z.coerce.boolean().default(false), {
      description: 'Emit JSON output'
    })
  },
  handler: async ({ positional, flags, colors, cwd }: HandlerArgs<GitPushFlags>) => {
    const parsedFlags = flags as GitPushFlags
    const target = positional[0]
    if (!target) {
      throw new Error('Usage: agentree git push <name|branch|path>')
    }

    const repoRoot = resolveRepoRoot(cwd)
    const { config } = await resolveConfig(repoRoot, {})
    const worktree = findWorktreeByTarget(repoRoot, target, config.branchPrefix)
    if (!worktree) {
      throw new Error(`Worktree not found for target: ${target}`)
    }

    if (!worktree.branch) {
      throw new Error('Cannot push detached worktree (no branch associated)')
    }

    pushBranch(worktree.path, parsedFlags.remote, worktree.branch)

    const data = {
      workspace: {
        path: worktree.path,
        branch: worktree.branch
      },
      remote: parsedFlags.remote
    }

    if (parsedFlags.json) {
      printJsonSuccess('git push', data)
      return
    }

    console.log(colors.green(`Pushed ${worktree.branch} to ${parsedFlags.remote}`))
    console.log(`path: ${worktree.path}`)
  }
})
