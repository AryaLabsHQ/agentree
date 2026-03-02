import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import { CommandExecutionError } from './errors.js'

export interface SetupSummary {
  commands: string[]
  completed: number
}

function hasBuildScript(dir: string): boolean {
  const packageJsonPath = join(dir, 'package.json')
  if (!existsSync(packageJsonPath)) return false

  try {
    const parsed = JSON.parse(readFileSync(packageJsonPath, 'utf8')) as {
      scripts?: Record<string, string>
    }
    return Boolean(parsed.scripts?.build)
  } catch {
    return false
  }
}

export function detectSetupCommands(dir: string): string[] {
  if (existsSync(join(dir, 'bun.lock')) || existsSync(join(dir, 'bun.lockb'))) {
    return ['bun install']
  }

  if (existsSync(join(dir, 'pnpm-lock.yaml'))) {
    return ['pnpm install']
  }

  if (existsSync(join(dir, 'package-lock.json'))) {
    return ['npm install']
  }

  if (existsSync(join(dir, 'yarn.lock'))) {
    return ['yarn install']
  }

  if (existsSync(join(dir, 'Cargo.lock'))) {
    return ['cargo build']
  }

  if (existsSync(join(dir, 'go.mod'))) {
    return ['go mod download']
  }

  if (existsSync(join(dir, 'requirements.txt'))) {
    return ['pip install -r requirements.txt']
  }

  if (existsSync(join(dir, 'Gemfile.lock'))) {
    return ['bundle install']
  }

  return []
}

function runScript(cwd: string, script: string): void {
  const proc = Bun.spawnSync({
    cmd: ['sh', '-lc', script],
    cwd,
    stdout: 'inherit',
    stderr: 'inherit'
  })

  if ((proc.exitCode ?? 1) !== 0) {
    throw new CommandExecutionError('sh', ['-lc', script], proc.exitCode ?? 1, '', '')
  }
}

export function runSetup(cwd: string, commands: string[]): SetupSummary {
  if (commands.length === 0) {
    return { commands: [], completed: 0 }
  }

  let completed = 0
  for (const command of commands) {
    runScript(cwd, command)
    completed++
  }

  return {
    commands,
    completed
  }
}

export function resolveSetupCommands(
  cwd: string,
  mode: 'auto-install' | 'custom',
  configuredScripts: string[]
): string[] {
  if (mode === 'custom') {
    return configuredScripts
  }

  const detected = detectSetupCommands(cwd)

  if (detected.length > 0 && hasBuildScript(cwd) && configuredScripts.includes('npm run build')) {
    return [...detected, 'npm run build']
  }

  return detected
}
