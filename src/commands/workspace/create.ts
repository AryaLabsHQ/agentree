import { defineCommand, option } from '@bunli/core'
import type { HandlerArgs } from '@bunli/core'
import { z } from 'zod'
import { mkdir, stat } from 'node:fs/promises'
import { dirname, isAbsolute, join, resolve } from 'node:path'
import {
  branchExists,
  createWorktree,
  currentBaseRef,
  deleteBranch,
  removeWorktree,
  resolveRepoRoot
} from '../../lib/git.js'
import { copyIgnoredArtifacts } from '../../lib/copyIgnored.js'
import { resolveConfig } from '../../lib/config.js'
import { PostStepError, RollbackError } from '../../lib/errors.js'
import { printJsonError, printJsonSuccess } from '../../lib/output.js'
import { resolveSetupCommands, runSetup } from '../../lib/setup.js'
import { defaultWorktreePath, deriveBranchName, validateWorkspaceName } from '../../lib/worktree.js'
import type { CreateStepSummary, WorkspaceCreateFlags } from '../../lib/types.js'

interface InteractiveCreateInput {
  name: string
  from: string
  dest: string
  copyIgnored: boolean
  destinationFromFlag: boolean
}

interface SpinnerLike {
  update(message: string): void
  succeed(message: string): void
  fail(message: string): void
}

function createSilentSpinner(): SpinnerLike {
  return {
    update() {},
    succeed() {},
    fail() {}
  }
}

async function askInteractive(
  prompt: {
    confirm(message: string, options?: { default?: boolean }): Promise<boolean>
    text(message: string, options?: { default?: string }): Promise<string>
  },
  initial: InteractiveCreateInput,
  destinationForName: (name: string) => string
): Promise<Omit<InteractiveCreateInput, 'destinationFromFlag'>> {
  const name = await prompt.text('Workspace name', { default: initial.name })
  const from = await prompt.text('Base branch/ref', { default: initial.from })

  const defaultDestination = resolveInteractiveDestinationDefault(
    initial.destinationFromFlag,
    initial.dest,
    destinationForName(name)
  )
  const dest = await prompt.text('Destination path', { default: defaultDestination })
  const copyIgnored = await prompt.confirm('Copy ignored artifacts?', { default: initial.copyIgnored })

  return { name, from, dest, copyIgnored }
}

export function defaultDestinationForName(repoRoot: string, workspaceName: string): string {
  return defaultWorktreePath(repoRoot, workspaceName)
}

export function resolveInteractiveDestinationDefault(
  destinationFromFlag: boolean,
  currentDestination: string,
  destinationForSelectedName: string
): string {
  if (destinationFromFlag) {
    return currentDestination
  }

  return destinationForSelectedName
}

async function pathExists(path: string): Promise<boolean> {
  const info = await stat(path).catch(() => null)
  return Boolean(info)
}

export const workspaceCreateOptions = {
  from: option(z.string().optional(), {
    short: 'f',
    description: 'Base branch or commit (default: current branch)'
  }),
  dest: option(z.string().optional(), {
    short: 'd',
    description: 'Destination path (default: ../<repo>-worktrees/<name>)'
  }),
  interactive: option(z.coerce.boolean().default(false), {
    short: 'i',
    description: 'Prompt-driven interactive flow'
  }),
  json: option(z.coerce.boolean().default(false), {
    description: 'Emit JSON output'
  }),
  copyIgnored: option(z.coerce.boolean().optional(), {
    description: 'Override copy ignored artifacts behavior'
  }),
  strict: option(z.coerce.boolean().optional(), {
    description: 'Fail on post-step errors (default: true)'
  }),
  rollbackOnFail: option(z.coerce.boolean().optional(), {
    description: 'Rollback branch/worktree on post-step failure (default: true)'
  }),
  include: option(z.string().optional(), {
    description: 'Comma-separated include patterns overriding excludes'
  }),
  exclude: option(z.string().optional(), {
    description: 'Comma-separated extra exclude patterns'
  })
}

export async function runWorkspaceCreate(
  { flags, positional, prompt, spinner, colors, cwd }: HandlerArgs<WorkspaceCreateFlags>,
  commandName = 'workspace create'
): Promise<void> {
  const parsedFlags = flags as WorkspaceCreateFlags
  const spin: SpinnerLike = parsedFlags.json ? createSilentSpinner() : spinner('Preparing workspace...')

  let createdWorktreePath: string | null = null
  let createdBranch: string | null = null
  let rollbackOnFailSetting = parsedFlags.rollbackOnFail ?? true

  try {
    const repoRoot = resolveRepoRoot(cwd)
    const configResult = await resolveConfig(repoRoot, parsedFlags)
    const config = configResult.config
    const strict = configResult.strict
    rollbackOnFailSetting = configResult.rollbackOnFail

    const baseRefDefault = parsedFlags.from ?? currentBaseRef(repoRoot)

    let workspaceName = positional[0] ?? ''
    let baseRef = baseRefDefault

    const destinationFromFlag = Boolean(parsedFlags.dest)
    const resolvedDestinationFromFlag = parsedFlags.dest
      ? isAbsolute(parsedFlags.dest)
        ? parsedFlags.dest
        : resolve(cwd, parsedFlags.dest)
      : undefined

    let destination = resolvedDestinationFromFlag ?? defaultDestinationForName(repoRoot, workspaceName)
    let copyIgnored = configResult.copyIgnored

    if (parsedFlags.interactive) {
      const chosen = await askInteractive(
        prompt,
        {
          name: workspaceName,
          from: baseRef,
          dest: destination,
          copyIgnored,
          destinationFromFlag
        },
        (name) => defaultDestinationForName(repoRoot, name)
      )
      workspaceName = chosen.name
      baseRef = chosen.from
      destination = chosen.dest
      copyIgnored = chosen.copyIgnored
    }

    validateWorkspaceName(workspaceName)

    const branch = deriveBranchName(config.branchPrefix, workspaceName)
    const destinationPath = isAbsolute(destination) ? destination : join(repoRoot, destination)

    if (branchExists(repoRoot, branch)) {
      throw new Error(`Branch already exists: ${branch}`)
    }

    if (await pathExists(destinationPath)) {
      throw new Error(`Destination already exists: ${destinationPath}`)
    }

    await mkdir(dirname(destinationPath), { recursive: true })

    spin.update(`Creating worktree ${branch}...`)
    createWorktree(repoRoot, branch, destinationPath, baseRef)
    createdWorktreePath = destinationPath
    createdBranch = branch

    const summary: CreateStepSummary = {}
    const warnings: string[] = []

    const runPostStep = async (step: string, fn: () => Promise<void> | void) => {
      try {
        await fn()
      } catch (error) {
        if (strict) {
          throw new PostStepError(step, error)
        }
        warnings.push(`${step}: ${error instanceof Error ? error.message : String(error)}`)
      }
    }

    if (copyIgnored) {
      spin.update('Copying ignored artifacts...')
      await runPostStep('copy-ignored', async () => {
        const copySummary = await copyIgnoredArtifacts({
          repoRoot,
          destinationRoot: destinationPath,
          excludes: [...config.defaultExcludes, ...config.extraExcludes],
          includes: config.extraIncludes,
          skipExisting: true
        })
        summary.copied = copySummary
      })
    }

    if (config.setupEnabled) {
      spin.update('Running setup...')
      await runPostStep('setup', async () => {
        const setupCommands = resolveSetupCommands(destinationPath, config.setupMode, config.setupScripts)
        const setupSummary = runSetup(destinationPath, setupCommands)
        summary.setup = setupSummary
      })
    }

    spin.succeed(`Created workspace ${workspaceName}`)

    const data = {
      workspace: {
        name: workspaceName,
        branch,
        path: destinationPath,
        from: baseRef
      },
      summary
    }

    if (parsedFlags.json) {
      printJsonSuccess(commandName, data, warnings)
    } else {
      console.log(colors.bold('Workspace ready'))
      console.log(`name:   ${workspaceName}`)
      console.log(`branch: ${branch}`)
      console.log(`path:   ${destinationPath}`)
      if (summary.copied) {
        console.log(
          `copied: ${summary.copied.copied}/${summary.copied.totalCandidates} (skipped existing: ${summary.copied.skippedExisting}, excluded: ${summary.copied.excluded})`
        )
      }
      if (summary.setup) {
        console.log(`setup:  ${summary.setup.completed} command(s)`) 
      }
      if (warnings.length > 0) {
        console.log(colors.yellow('Warnings:'))
        for (const warning of warnings) {
          console.log(`- ${warning}`)
        }
      }
    }
  } catch (error) {
    const repoRoot = (() => {
      try {
        return resolveRepoRoot(cwd)
      } catch {
        return null
      }
    })()
    const rollbackErrors: Error[] = []

    const shouldRollback =
      rollbackOnFailSetting &&
      createdWorktreePath !== null &&
      createdBranch !== null &&
      repoRoot !== null

    if (shouldRollback) {
      spin.update('Rolling back failed create...')
      try {
        removeWorktree(repoRoot!, createdWorktreePath!, true)
      } catch (rollbackError) {
        rollbackErrors.push(
          rollbackError instanceof Error ? rollbackError : new Error(String(rollbackError))
        )
      }

      try {
        deleteBranch(repoRoot!, createdBranch!)
      } catch (rollbackError) {
        rollbackErrors.push(
          rollbackError instanceof Error ? rollbackError : new Error(String(rollbackError))
        )
      }
    }

    if (rollbackErrors.length > 0 && error instanceof Error) {
      const combined = new RollbackError(error, rollbackErrors)
      spin.fail(combined.message)
      if (parsedFlags.json) {
        printJsonError(commandName, combined.message, {
          originalError: combined.original.message,
          rollbackErrors: combined.rollbackFailures.map((rollbackError) => rollbackError.message)
        })
      }
      throw combined
    }

    const message = error instanceof Error ? error.message : String(error)
    spin.fail(message)
    if (parsedFlags.json) {
      printJsonError(commandName, message)
    }
    throw error
  }
}

export default defineCommand({
  name: 'create',
  description: 'Create a new isolated worktree workspace',
  options: workspaceCreateOptions,
  handler: async (args) => {
    await runWorkspaceCreate(args, 'workspace create')
  }
})
