import { defineCommand, option } from '@bunli/core'
import { z } from 'zod'
import { deleteBranch, findWorktreeByTarget, removeWorktree, resolveRepoRoot } from '../lib/git.js'
import { printJson } from '../lib/output.js'

const DEFAULT_BRANCH_PREFIX = 'agent/'

export default defineCommand({
  name: 'rm',
  description: 'Remove a workspace by name, branch, or path',
  options: {
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
  },
  handler: async ({ positional, flags, prompt, colors, cwd }) => {
    const target = positional[0]
    if (!target) {
      throw new Error('Usage: agentree rm <name|branch|path>')
    }

    const repoRoot = resolveRepoRoot(cwd)
    const worktree = findWorktreeByTarget(repoRoot, target, DEFAULT_BRANCH_PREFIX)
    if (!worktree) {
      throw new Error(`Worktree not found for target: ${target}`)
    }

    if (!flags.yes) {
      const confirmed = await prompt.confirm(`Remove worktree at ${worktree.path}?`, {
        default: false
      })
      if (!confirmed) {
        if (flags.json) {
          printJson({ ok: true, cancelled: true })
        } else {
          console.log(colors.yellow('Cancelled'))
        }
        return
      }
    }

    removeWorktree(repoRoot, worktree.path, true)

    let deletedBranch: string | null = null
    if (flags.deleteBranch && worktree.branch) {
      deleteBranch(repoRoot, worktree.branch)
      deletedBranch = worktree.branch
    }

    const result = {
      ok: true,
      removedPath: worktree.path,
      deletedBranch
    }

    if (flags.json) {
      printJson(result)
    } else {
      console.log(colors.green(`Removed ${worktree.path}`))
      if (deletedBranch) {
        console.log(colors.green(`Deleted branch ${deletedBranch}`))
      }
    }
  }
})
