import { afterEach, beforeEach, describe, expect, test } from 'bun:test'
import { mkdir, rm, writeFile } from 'node:fs/promises'
import { join } from 'node:path'
import { resolveConfig } from '../src/lib/config.js'

const tempBase = '/tmp/agentree-config-test'

async function createRepo(): Promise<string> {
  await rm(tempBase, { recursive: true, force: true })
  const repo = join(tempBase, 'repo')
  const home = join(tempBase, 'home')
  await mkdir(join(repo, '.agentree'), { recursive: true })
  await mkdir(join(home, '.config', 'agentree'), { recursive: true })

  await writeFile(
    join(home, '.config', 'agentree', 'config.json'),
    JSON.stringify({ branchPrefix: 'agent/', strictDefault: true, extraExcludes: ['.cache/'] }, null, 2)
  )

  await writeFile(
    join(repo, '.agentree', 'config.json'),
    JSON.stringify({ branchPrefix: 'agent/', strictDefault: false, copyIgnoredEnabled: true }, null, 2)
  )

  process.env.HOME = home
  return repo
}

describe('config resolution', () => {
  beforeEach(async () => {
    await createRepo()
  })

  afterEach(async () => {
    await rm(tempBase, { recursive: true, force: true })
  })

  test('project overrides global, flags override project', async () => {
    const repoRoot = join(tempBase, 'repo')
    const resolved = await resolveConfig(repoRoot, {
      strict: true,
      include: '.env',
      exclude: '.tmp'
    })

    expect(resolved.config.branchPrefix).toBe('agent/')
    expect(resolved.strict).toBe(true)
    expect(resolved.config.extraIncludes).toContain('.env')
    expect(resolved.config.extraExcludes).toContain('.tmp')
  })
})
