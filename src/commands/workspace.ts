import { defineGroup } from '@bunli/core'
import createWorkspaceCommand from './workspace/create.js'
import listWorkspacesCommand from './workspace/list.js'
import removeWorkspaceCommand from './workspace/remove.js'

export default defineGroup({
  name: 'workspace',
  description: 'Create, list, and remove workspaces',
  commands: [createWorkspaceCommand, listWorkspacesCommand, removeWorkspaceCommand]
})
