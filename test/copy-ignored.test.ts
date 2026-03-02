import { afterEach, beforeEach, describe, expect, test } from 'bun:test'
import { mkdir, readFile, rm, writeFile } from 'node:fs/promises'
import { join } from 'node:path'
import { copyIgnoredArtifacts } from '../src/lib/copyIgnored.js'

const tempBase = '/tmp/agentree-copy-test'

async function run(cmd: string[], cwd: string) {
  const proc = Bun.spawnSync({ cmd, cwd, stdout: 'pipe', stderr: 'pipe' })
  if ((proc.exitCode ?? 1) !== 0) {
    throw new Error(Buffer.from(proc.stderr).toString('utf8'))
  }
}

beforeEach(async () => {
  await rm(tempBase, { recursive: true, force: true })
  await mkdir(tempBase, { recursive: true })
})

afterEach(async () => {
  await rm(tempBase, { recursive: true, force: true })
})

describe('copy ignored artifacts', () => {
  test('copies ignored files and excludes dependency dirs', async () => {
    const repo = join(tempBase, 'repo')
    const dest = join(tempBase, 'dest')
    await mkdir(repo, { recursive: true })
    await mkdir(dest, { recursive: true })

    await run(['git', 'init'], repo)
    await run(['git', 'config', 'user.name', 'Test'], repo)
    await run(['git', 'config', 'user.email', 'test@example.com'], repo)
    await writeFile(join(repo, '.gitignore'), '.env\nnode_modules/\n')
    await writeFile(join(repo, 'README.md'), 'hello\n')
    await run(['git', 'add', '.'], repo)
    await run(['git', 'commit', '-m', 'init'], repo)

    await writeFile(join(repo, '.env'), 'SECRET=1\n')
    await mkdir(join(repo, 'node_modules', 'left-pad'), { recursive: true })
    await writeFile(join(repo, 'node_modules', 'left-pad', 'index.js'), 'module.exports = 1\n')

    const summary = await copyIgnoredArtifacts({
      repoRoot: repo,
      destinationRoot: dest,
      excludes: ['.git/', 'node_modules/'],
      includes: [],
      skipExisting: true
    })

    expect(summary.totalCandidates).toBeGreaterThan(0)
    expect(summary.copied).toBeGreaterThan(0)

    const copiedEnv = await readFile(join(dest, '.env'), 'utf8')
    expect(copiedEnv).toContain('SECRET=1')

    const nodeModulesExists = await Bun.file(join(dest, 'node_modules', 'left-pad', 'index.js')).exists()
    expect(nodeModulesExists).toBe(false)
  })
})
