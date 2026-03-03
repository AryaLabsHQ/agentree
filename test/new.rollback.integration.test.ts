import { afterEach, beforeEach, describe, expect, test } from 'bun:test'
import { mkdir, realpath, rm, writeFile } from 'node:fs/promises'
import { basename, dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const tempBase = '/tmp/agentree-new-rollback-test'
const projectRoot = join(dirname(fileURLToPath(import.meta.url)), '..')
const cliEntry = join(projectRoot, 'src', 'cli.ts')

function run(cmd: string[], cwd: string) {
  return Bun.spawnSync({ cmd, cwd, stdout: 'pipe', stderr: 'pipe' })
}

async function setupRepo(repo: string, rollbackOnFailDefault: boolean) {
  await mkdir(repo, { recursive: true })
  await run(['git', 'init'], repo)
  await run(['git', 'config', 'user.name', 'Test'], repo)
  await run(['git', 'config', 'user.email', 'test@example.com'], repo)
  await writeFile(join(repo, '.gitignore'), '.env\n')
  await writeFile(join(repo, 'README.md'), 'hello\n')
  await run(['git', 'add', '.'], repo)
  await run(['git', 'commit', '-m', 'init'], repo)
  await writeFile(join(repo, '.env'), 'SECRET=1\n')

  await mkdir(join(repo, '.agentree'), { recursive: true })
  await writeFile(
    join(repo, '.agentree', 'config.json'),
    JSON.stringify(
      {
        rollbackOnFailDefault,
        setupEnabled: true,
        setupMode: 'custom',
        setupScripts: ['exit 13']
      },
      null,
      2
    )
  )
}

function branchExists(repo: string, branch: string): boolean {
  const result = run(['git', 'show-ref', '--verify', '--quiet', `refs/heads/${branch}`], repo)
  return result.exitCode === 0
}

async function expectedWorktreePath(repo: string, workspaceName: string): Promise<string> {
  const resolvedRepo = await realpath(repo)
  return join(dirname(resolvedRepo), `${basename(resolvedRepo)}-worktrees`, workspaceName)
}

describe('workspace create rollback behavior', () => {
  beforeEach(async () => {
    await rm(tempBase, { recursive: true, force: true })
    await mkdir(tempBase, { recursive: true })
  })

  afterEach(async () => {
    await rm(tempBase, { recursive: true, force: true })
  })

  test('does not rollback when rollbackOnFailDefault is false', async () => {
    const repo = join(tempBase, 'repo-no-rollback')
    await setupRepo(repo, false)

    const result = run(
      ['bun', cliEntry, 'workspace', 'create', 'no-rollback', '--copyIgnored=false', '--json'],
      repo
    )
    expect(result.exitCode).not.toBe(0)

    const worktreePath = await expectedWorktreePath(repo, 'no-rollback')
    expect(await Bun.file(join(worktreePath, '.git')).exists()).toBe(true)
    expect(branchExists(repo, 'agent/no-rollback')).toBe(true)
  })

  test('CLI override rollbackOnFail=true forces rollback', async () => {
    const repo = join(tempBase, 'repo-force-rollback')
    await setupRepo(repo, false)

    const result = run(
      [
        'bun',
        cliEntry,
        'workspace',
        'create',
        'force-rollback',
        '--copyIgnored=false',
        '--rollbackOnFail=true',
        '--json'
      ],
      repo
    )
    expect(result.exitCode).not.toBe(0)

    const worktreePath = await expectedWorktreePath(repo, 'force-rollback')
    expect(await Bun.file(join(worktreePath, '.git')).exists()).toBe(false)
    expect(branchExists(repo, 'agent/force-rollback')).toBe(false)
  })
})
