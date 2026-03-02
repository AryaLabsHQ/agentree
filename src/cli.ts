#!/usr/bin/env bun
import { createCLI } from '@bunli/core'
import { completionsPlugin } from '@bunli/plugin-completions'
import { existsSync, mkdtempSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import newCommand from './commands/new.js'
import lsCommand from './commands/ls.js'
import rmCommand from './commands/rm.js'

async function createAgentreeCLI() {
  const startupCwd = process.cwd()
  const isolatedCwd = mkdtempSync(join(tmpdir(), 'agentree-bunli-'))
  const moduleDir = dirname(fileURLToPath(import.meta.url))
  const generatedMetadataPath = (() => {
    const candidates = [
      join(moduleDir, '.bunli', 'commands.gen.ts'),
      join(dirname(process.execPath), '.bunli', 'commands.gen.ts'),
      join(dirname(process.execPath), '..', '.bunli', 'commands.gen.ts'),
      join(moduleDir, '..', '.bunli', 'commands.gen.ts'),
      join(moduleDir, '..', '..', '.bunli', 'commands.gen.ts'),
      join(process.cwd(), '.bunli', 'commands.gen.ts')
    ]
    return candidates.find((path) => existsSync(path)) ?? candidates[0]
  })()

  try {
    // Isolate Bunli config discovery from the user's current repository.
    process.chdir(isolatedCwd)
    return await createCLI({
      name: 'agentree',
      version: '1.0.0',
      description: 'Create and manage isolated git worktrees for AI coding agents',
      plugins: [
        completionsPlugin({
          generatedPath: generatedMetadataPath,
          commandName: 'agentree',
          executable: 'agentree',
          includeAliases: true,
          includeGlobalFlags: true
        })
      ]
    })
  } finally {
    process.chdir(startupCwd)
    rmSync(isolatedCwd, { recursive: true, force: true })
  }
}

const cli = await createAgentreeCLI()

cli.command(newCommand)
cli.command(lsCommand)
cli.command(rmCommand)

await cli.run()
