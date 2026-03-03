import { defineCommand, option } from '@bunli/core'
import type { HandlerArgs } from '@bunli/core'
import { z } from 'zod'
import { basename } from 'node:path'
import { listWorktrees, resolveRepoRoot } from '../../lib/git.js'
import { resolveConfig } from '../../lib/config.js'
import { printJsonSuccess, printWorkspacesTable } from '../../lib/output.js'
import type { ManagedWorkspace, WorkspaceListFlags } from '../../lib/types.js'

export const workspaceListOptions = {
  json: option(z.coerce.boolean().default(false), {
    description: 'Emit JSON output'
  }),
  all: option(z.coerce.boolean().default(false), {
    description: 'Include non-managed worktrees'
  })
}

export async function runWorkspaceList(
  { flags, cwd }: HandlerArgs<WorkspaceListFlags>,
  commandName = 'workspace list'
): Promise<void> {
  const parsedFlags = flags as WorkspaceListFlags

  const repoRoot = resolveRepoRoot(cwd)
  const { config } = await resolveConfig(repoRoot, {})
  const worktrees = listWorktrees(repoRoot)

  const mapped: ManagedWorkspace[] = worktrees.map((wt) => {
    const managed = Boolean(wt.branch?.startsWith(config.branchPrefix))
    const name = wt.branch?.startsWith(config.branchPrefix)
      ? wt.branch.slice(config.branchPrefix.length)
      : basename(wt.path)

    return {
      ...wt,
      name,
      managed
    }
  })

  const filtered = parsedFlags.all ? mapped : mapped.filter((workspace) => workspace.managed)

  if (parsedFlags.json) {
    printJsonSuccess(commandName, { workspaces: filtered })
    return
  }

  printWorkspacesTable(filtered)
}

export default defineCommand({
  name: 'list',
  description: 'List workspaces',
  options: workspaceListOptions,
  handler: async (args) => {
    await runWorkspaceList(args, 'workspace list')
  }
})
