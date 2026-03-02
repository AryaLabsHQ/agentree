import { afterEach, beforeEach, describe, expect, test } from 'bun:test'
import { mkdir, rm, writeFile } from 'node:fs/promises'
import { dirname, join } from 'node:path'

const tempBase = '/tmp/agentree-new-rollback-test'

async function run(cmd: string[], cwd: string) {
  return Bun.spawnSync({ cmd, cwd, stdout: 'pipe', stderr: 'pipe' })
}

async function setupRepo(repo: string) {
  await mkdir(repo, { recursive: true })
  await run(['git', 'init'], repo)
  await run(['git', 'config', 'user.name', 'Test'], repo)
  await run(['git', 'config', 'user.email', 'test@example.com'], repo)
  await writeFile(join(repo, '.gitignore'), '.env\n')
  await writeFile(join(repo, 'README.md'), 'hello\n')
  await run(['git', 'add', '.'], repo)
  await run(['git', 'commit', '-m', 'init'], repo)
  await writeFile(join(repo, '.env'), 'SECRET=1\n')
}

describe('new command rollback', () => {
  beforeEach(async () => {
    await rm(tempBase, { recursive: true, force: true })
    await mkdir(tempBase, { recursive: true })
  })

  afterEach(async () => {
    await rm(tempBase, { recursive: true, force: true })
  })

  test('rolls back branch and worktree when push fails', async () => {
    const repo = join(tempBase, 'repo')
    await setupRepo(repo)

    const result = await run(['bun', 'src/cli.ts', 'new', 'rollback-case', '--push=true'], repo)
    expect(result.exitCode).not.toBe(0)

    const worktreePath = join(dirname(repo), 'repo-worktrees', 'rollback-case')
    const worktreeExists = await Bun.file(worktreePath).exists()
    expect(worktreeExists).toBe(false)

    const branchCheck = await run(
      ['git', 'show-ref', '--verify', '--quiet', 'refs/heads/agent/rollback-case'],
      repo
    )
    expect(branchCheck.exitCode).not.toBe(0)
  })
})
