import { existsSync } from 'node:fs'
import { mkdir, readFile } from 'node:fs/promises'
import { join } from 'node:path'
import type { AgentreeConfig, ConfigOverrideFlags } from './types.js'

export const DEFAULT_DEPENDENCY_EXCLUDES = [
  '.git/',
  'node_modules/',
  '.venv/',
  'venv/',
  '.tox/',
  '.gradle/',
  'Pods/',
  '.build/',
  'vendor/bundle/',
  '.m2/',
  '.nuget/',
  '.dart_tool/'
]

export const DEFAULT_CONFIG: AgentreeConfig = {
  branchPrefix: 'agent/',
  copyIgnoredEnabled: true,
  copyIgnoredSource: 'gitignored',
  copyConflictPolicy: 'skip-existing',
  defaultExcludes: DEFAULT_DEPENDENCY_EXCLUDES,
  extraIncludes: [],
  extraExcludes: [],
  setupEnabled: true,
  setupMode: 'auto-install',
  setupScripts: [],
  strictDefault: true,
  rollbackOnFailDefault: true,
  completionEnabled: true
}

type PartialConfig = Partial<AgentreeConfig>

async function loadJsonConfig(path: string): Promise<PartialConfig> {
  if (!existsSync(path)) {
    return {}
  }

  const raw = await readFile(path, 'utf8')
  const parsed = JSON.parse(raw)
  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
    throw new Error(`Config at ${path} must be a JSON object`)
  }

  return parsed as PartialConfig
}

function mergeConfig(base: AgentreeConfig, override: PartialConfig): AgentreeConfig {
  return {
    ...base,
    ...override,
    defaultExcludes: override.defaultExcludes ?? base.defaultExcludes,
    extraIncludes: override.extraIncludes ?? base.extraIncludes,
    extraExcludes: override.extraExcludes ?? base.extraExcludes,
    setupScripts: override.setupScripts ?? base.setupScripts
  }
}

function parseCsv(value?: string): string[] {
  if (!value) return []
  return value
    .split(',')
    .map((v) => v.trim())
    .filter(Boolean)
}

export async function resolveConfig(
  repoRoot: string,
  flags: ConfigOverrideFlags
): Promise<{
  config: AgentreeConfig
  strict: boolean
  rollbackOnFail: boolean
  copyIgnored: boolean
}> {
  const home = process.env.HOME || process.env.USERPROFILE
  const globalPath = home ? join(home, '.config', 'agentree', 'config.json') : undefined
  const projectPath = join(repoRoot, '.agentree', 'config.json')

  let merged = DEFAULT_CONFIG

  if (globalPath) {
    merged = mergeConfig(merged, await loadJsonConfig(globalPath))
  }

  merged = mergeConfig(merged, await loadJsonConfig(projectPath))

  const cliExtraIncludes = parseCsv(flags.include)
  const cliExtraExcludes = parseCsv(flags.exclude)

  merged = {
    ...merged,
    extraIncludes: [...merged.extraIncludes, ...cliExtraIncludes],
    extraExcludes: [...merged.extraExcludes, ...cliExtraExcludes]
  }

  const strict = flags.strict ?? merged.strictDefault
  const rollbackOnFail = flags.rollbackOnFail ?? merged.rollbackOnFailDefault
  const copyIgnored = flags.copyIgnored ?? merged.copyIgnoredEnabled

  return { config: merged, strict, rollbackOnFail, copyIgnored }
}

export async function ensureProjectConfigDirectory(repoRoot: string): Promise<void> {
  await mkdir(join(repoRoot, '.agentree'), { recursive: true })
}
