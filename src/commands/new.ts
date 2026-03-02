import { defineCommand, option } from '@bunli/core'
import { z } from 'zod'
import { mkdir, stat } from 'node:fs/promises'
import { dirname, isAbsolute, join, resolve } from 'node:path'
import {
  branchExists,
  createPullRequest,
  createWorktree,
  currentBaseRef,
  deleteBranch,
  ghExists,
  pushBranch,
  removeWorktree,
  resolveRepoRoot
} from '../lib/git.js'
import { copyIgnoredArtifacts } from '../lib/copyIgnored.js'
import { resolveConfig } from '../lib/config.js'
import { RollbackError, PostStepError } from '../lib/errors.js'
import { printJson } from '../lib/output.js'
import { resolveSetupCommands, runSetup } from '../lib/setup.js'
import { defaultWorktreePath, deriveBranchName, validateWorkspaceName } from '../lib/worktree.js'
import type { CreateStepSummary, NewCommandFlags } from '../lib/types.js'

async function askInteractive(
  prompt: { confirm(message: string, options?: { default?: boolean }): Promise<boolean>; text(message: string, options?: { default?: string }): Promise<string> },
  initial: {
    name: string
    from: string
    dest: string
    push: boolean
    pr: boolean
    copyIgnored: boolean
  }
): Promise<typeof initial> {
  const name = await prompt.text('Workspace name', { default: initial.name })
  const from = await prompt.text('Base branch/ref', { default: initial.from })
  const dest = await prompt.text('Destination path', { default: initial.dest })
  const copyIgnored = await prompt.confirm('Copy ignored artifacts?', { default: initial.copyIgnored })
  const push = await prompt.confirm('Push after create?', { default: initial.push })
  const pr = await prompt.confirm('Create PR after push?', { default: initial.pr })

  return { name, from, dest, push, pr, copyIgnored }
}

async function pathExists(path: string): Promise<boolean> {
  const info = await stat(path).catch(() => null)
  return Boolean(info)
}

export default defineCommand({
  name: 'new',
  description: 'Create a new isolated worktree workspace',
  options: {
    from: option(z.string().optional(), {
      short: 'f',
      description: 'Base branch or commit (default: current branch)'
    }),
    dest: option(z.string().optional(), {
      short: 'd',
      description: 'Destination path (default: ../<repo>-worktrees/<name>)'
    }),
    push: option(z.coerce.boolean().default(false), {
      short: 'p',
      description: 'Push branch after creation'
    }),
    pr: option(z.coerce.boolean().default(false), {
      short: 'r',
      description: 'Create PR (implies --push)'
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
  },
  handler: async ({ flags, positional, prompt, spinner, colors, cwd }) => {
    const parsedFlags = flags as unknown as NewCommandFlags
    const spin = spinner('Preparing workspace...')

    let createdWorktreePath: string | null = null
    let createdBranch: string | null = null

    try {
      const repoRoot = resolveRepoRoot(cwd)
      const configResult = await resolveConfig(repoRoot, parsedFlags)
      const config = configResult.config
      const strict = configResult.strict
      const rollbackOnFail = configResult.rollbackOnFail

      const baseRefDefault = parsedFlags.from ?? currentBaseRef(repoRoot)

      let workspaceName = positional[0] ?? ''
      let baseRef = baseRefDefault
      let destination = parsedFlags.dest
        ? isAbsolute(parsedFlags.dest)
          ? parsedFlags.dest
          : resolve(cwd, parsedFlags.dest)
        : defaultWorktreePath(repoRoot, workspaceName)
      let push = parsedFlags.push || parsedFlags.pr
      let pr = parsedFlags.pr
      let copyIgnored = configResult.copyIgnored

      if (parsedFlags.interactive) {
        const chosen = await askInteractive(prompt, {
          name: workspaceName,
          from: baseRef,
          dest: destination,
          push,
          pr,
          copyIgnored
        })
        workspaceName = chosen.name
        baseRef = chosen.from
        destination = chosen.dest
        push = chosen.push || chosen.pr
        pr = chosen.pr
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

      if (push) {
        spin.update('Pushing branch...')
        await runPostStep('push', () => {
          pushBranch(destinationPath, 'origin', branch)
          summary.push = { remote: 'origin', branch }
        })
      }

      if (pr) {
        spin.update('Creating pull request...')
        await runPostStep('pr', () => {
          if (!ghExists(destinationPath)) {
            throw new Error('GitHub CLI (gh) is not installed')
          }
          createPullRequest(destinationPath)
          summary.pr = { created: true }
        })
      }

      spin.succeed(`Created workspace ${workspaceName}`)

      const result = {
        ok: true,
        workspace: {
          name: workspaceName,
          branch,
          path: destinationPath,
          from: baseRef
        },
        summary,
        warnings
      }

      if (parsedFlags.json) {
        printJson(result)
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
      const parsedFlags = flags as unknown as NewCommandFlags
      const repoRoot = (() => {
        try {
          return resolveRepoRoot(cwd)
        } catch {
          return null
        }
      })()
      const rollbackErrors: Error[] = []

      const shouldRollback =
        (parsedFlags.rollbackOnFail ?? true) &&
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
        if ((flags as unknown as NewCommandFlags).json) {
          printJson({
            ok: false,
            error: combined.message,
            originalError: combined.original.message,
            rollbackErrors: combined.rollbackFailures.map((e) => e.message)
          })
        }
        throw combined
      }

      const message = error instanceof Error ? error.message : String(error)
      spin.fail(message)
      if ((flags as unknown as NewCommandFlags).json) {
        printJson({ ok: false, error: message })
      }
      throw error
    }
  }
})
