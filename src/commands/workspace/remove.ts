import { defineCommand, option } from '@bunli/core'
import type { HandlerArgs } from '@bunli/core'
import { z } from 'zod'
import { deleteBranch, findWorktreeByTarget, removeWorktree, resolveRepoRoot } from '../../lib/git.js'
import { resolveConfig } from '../../lib/config.js'
import { printJsonSuccess } from '../../lib/output.js'
import type { WorkspaceRemoveFlags } from '../../lib/types.js'

export const workspaceRemoveOptions = {
  yes: option(z.coerce.boolean().default(false), {
    short: 'y',
    description: 'Skip confirmation'
  }),
  deleteBranch: option(z.coerce.boolean().default(false), {
    short: 'R',
    description: 'Also delete local branch'
  }),
  json: option(z.coerce.boolean().default(false), {
    description: 'Emit JSON output'
  })
}

export async function runWorkspaceRemove(
  { positional, flags, prompt, colors, cwd }: HandlerArgs<WorkspaceRemoveFlags>,
  commandName = 'workspace remove'
): Promise<void> {
  const parsedFlags = flags as WorkspaceRemoveFlags
  const target = positional[0]
  if (!target) {
    throw new Error('Usage: agentree workspace remove <name|branch|path>')
  }

  const repoRoot = resolveRepoRoot(cwd)
  const { config } = await resolveConfig(repoRoot, {})
  const worktree = findWorktreeByTarget(repoRoot, target, config.branchPrefix)
  if (!worktree) {
    throw new Error(`Worktree not found for target: ${target}`)
  }

  if (!parsedFlags.yes) {
    const confirmed = await prompt.confirm(`Remove worktree at ${worktree.path}?`, {
      default: false
    })
    if (!confirmed) {
      if (parsedFlags.json) {
        printJsonSuccess(commandName, { cancelled: true })
      } else {
        console.log(colors.yellow('Cancelled'))
      }
      return
    }
  }

  removeWorktree(repoRoot, worktree.path, true)

  let deletedBranch: string | null = null
  if (parsedFlags.deleteBranch && worktree.branch) {
    deleteBranch(repoRoot, worktree.branch)
    deletedBranch = worktree.branch
  }

  const data = {
    removedPath: worktree.path,
    deletedBranch
  }

  if (parsedFlags.json) {
    printJsonSuccess(commandName, data)
  } else {
    console.log(colors.green(`Removed ${worktree.path}`))
    if (deletedBranch) {
      console.log(colors.green(`Deleted branch ${deletedBranch}`))
    }
  }
}

export default defineCommand({
  name: 'remove',
  description: 'Remove a workspace by name, branch, or path',
  options: workspaceRemoveOptions,
  handler: async (args) => {
    await runWorkspaceRemove(args, 'workspace remove')
  }
})
