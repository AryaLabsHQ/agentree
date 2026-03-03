import { defineCommand } from '@bunli/core'
import { workspaceRemoveOptions, runWorkspaceRemove } from './workspace/remove.js'

export default defineCommand({
  name: 'rm',
  description: 'Alias for workspace remove',
  options: workspaceRemoveOptions,
  handler: async (args) => {
    await runWorkspaceRemove(args, 'workspace remove')
  }
})
