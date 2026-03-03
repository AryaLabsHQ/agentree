import { afterEach, beforeEach, describe, expect, test } from 'bun:test'
import { mkdir, rm, writeFile } from 'node:fs/promises'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const tempBase = '/tmp/agentree-workspace-commands-test'
const projectRoot = join(dirname(fileURLToPath(import.meta.url)), '..')
const cliEntry = join(projectRoot, 'src', 'cli.ts')

function runCli(args: string[], cwd: string) {
  return Bun.spawnSync({ cmd: ['bun', cliEntry, ...args], cwd, stdout: 'pipe', stderr: 'pipe' })
}

function run(cmd: string[], cwd: string) {
  return Bun.spawnSync({ cmd, cwd, stdout: 'pipe', stderr: 'pipe' })
}

function parseJsonFromStdout(proc: Bun.SyncSubprocess): any {
  const stdout = Buffer.from(proc.stdout ?? new Uint8Array()).toString('utf8').trim()
  const firstBrace = stdout.indexOf('{')
  if (firstBrace === -1) {
    throw new Error(`Expected JSON in stdout, got: ${stdout}`)
  }
  return JSON.parse(stdout.slice(firstBrace))
}

async function setupRepo(repo: string, branchPrefix = 'workspace/') {
  await mkdir(repo, { recursive: true })
  run(['git', 'init'], repo)
  run(['git', 'config', 'user.name', 'Test'], repo)
  run(['git', 'config', 'user.email', 'test@example.com'], repo)
  await writeFile(join(repo, 'README.md'), 'hello\n')
  run(['git', 'add', '.'], repo)
  run(['git', 'commit', '-m', 'init'], repo)

  await mkdir(join(repo, '.agentree'), { recursive: true })
  await writeFile(
    join(repo, '.agentree', 'config.json'),
    JSON.stringify(
      {
        branchPrefix,
        copyIgnoredEnabled: false,
        setupEnabled: false
      },
      null,
      2
    )
  )
}

describe('workspace command surface', () => {
  beforeEach(async () => {
    await rm(tempBase, { recursive: true, force: true })
    await mkdir(tempBase, { recursive: true })
  })

  afterEach(async () => {
    await rm(tempBase, { recursive: true, force: true })
  })

  test('grouped commands work and ls/rm aliases remain valid', async () => {
    const repo = join(tempBase, 'repo-aliases')
    await setupRepo(repo)

    const createGrouped = runCli(['workspace', 'create', 'alpha', '--json'], repo)
    expect(createGrouped.exitCode).toBe(0)

    const createAlias = runCli(['new', 'beta', '--json'], repo)
    expect(createAlias.exitCode).toBe(0)

    const groupedList = parseJsonFromStdout(runCli(['workspace', 'list', '--json'], repo))
    expect(groupedList.schemaVersion).toBe('1.0.0')
    expect(groupedList.ok).toBe(true)
    expect(groupedList.command).toBe('workspace list')
    expect(Array.isArray(groupedList.data.workspaces)).toBe(true)

    const aliasList = parseJsonFromStdout(runCli(['ls', '--json'], repo))
    expect(aliasList.ok).toBe(true)
    expect(aliasList.data.workspaces.length).toBe(groupedList.data.workspaces.length)

    const removeGrouped = runCli(['workspace', 'remove', 'alpha', '--yes', '--deleteBranch', '--json'], repo)
    expect(removeGrouped.exitCode).toBe(0)

    const removeAlias = runCli(['rm', 'beta', '--yes', '--deleteBranch', '--json'], repo)
    expect(removeAlias.exitCode).toBe(0)
  })

  test('list/remove honor configured branchPrefix and rm resolves name before path', async () => {
    const repo = join(tempBase, 'repo-prefix')
    await setupRepo(repo, 'workspace/')

    const createResult = runCli(['workspace', 'create', 'alpha', '--json'], repo)
    expect(createResult.exitCode).toBe(0)

    const listResult = parseJsonFromStdout(runCli(['workspace', 'list', '--json'], repo))
    const alpha = listResult.data.workspaces.find(
      (workspace: { name: string; branch?: string }) => workspace.name === 'alpha'
    )
    expect(alpha).toBeTruthy()
    expect(alpha.branch).toBe('workspace/alpha')

    await mkdir(join(repo, 'alpha'), { recursive: true })

    const removeResult = runCli(
      ['workspace', 'remove', 'alpha', '--yes', '--deleteBranch', '--json'],
      repo
    )
    expect(removeResult.exitCode).toBe(0)

    const branchCheck = run(
      ['git', 'show-ref', '--verify', '--quiet', 'refs/heads/workspace/alpha'],
      repo
    )
    expect(branchCheck.exitCode).not.toBe(0)
  })
})
