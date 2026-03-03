import { defineCommand } from '@bunli/core'
import { workspaceListOptions, runWorkspaceList } from './workspace/list.js'

export default defineCommand({
  name: 'ls',
  description: 'Alias for workspace list',
  options: workspaceListOptions,
  handler: async (args) => {
    await runWorkspaceList(args, 'workspace list')
  }
})
