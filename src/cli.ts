#!/usr/bin/env bun
import { createCLI } from '@bunli/core'
import { completionsPlugin } from '@bunli/plugin-completions'
import gitCommand from './commands/git.js'
import lsCommand from './commands/ls.js'
import newCommand from './commands/new.js'
import rmCommand from './commands/rm.js'
import workspaceCommand from './commands/workspace.js'

const cli = await createCLI({
  name: 'agentree',
  version: '1.0.0',
  description: 'Create and manage isolated git worktrees for AI coding agents',
  plugins: [
    completionsPlugin({
      commandName: 'agentree',
      executable: 'agentree',
      includeAliases: true,
      includeGlobalFlags: true
    })
  ]
})

cli.command(workspaceCommand)
cli.command(gitCommand)

// Official top-level aliases for workspace lifecycle.
cli.command(newCommand)
cli.command(lsCommand)
cli.command(rmCommand)

await cli.run()
