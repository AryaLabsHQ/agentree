import { defineCommand } from '@bunli/core'
import { workspaceCreateOptions, runWorkspaceCreate } from './workspace/create.js'

export default defineCommand({
  name: 'new',
  description: 'Alias for workspace create',
  options: workspaceCreateOptions,
  handler: async (args) => {
    await runWorkspaceCreate(args, 'workspace create')
  }
})
