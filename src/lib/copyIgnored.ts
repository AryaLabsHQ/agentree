import { copyFile, lstat, mkdir, readlink, symlink } from 'node:fs/promises'
import { dirname, join } from 'node:path'
import { minimatch } from 'minimatch'
import { CommandExecutionError } from './errors.js'

export interface CopyIgnoredOptions {
  repoRoot: string
  destinationRoot: string
  excludes: string[]
  includes: string[]
  skipExisting: boolean
}

export interface CopyIgnoredSummary {
  totalCandidates: number
  copied: number
  skippedExisting: number
  excluded: number
}

function runGitIgnoredList(repoRoot: string): string[] {
  const proc = Bun.spawnSync({
    cmd: ['git', 'ls-files', '--others', '-i', '--exclude-standard', '-z'],
    cwd: repoRoot,
    stdout: 'pipe',
    stderr: 'pipe'
  })

  if ((proc.exitCode ?? 1) !== 0) {
    throw new CommandExecutionError(
      'git',
      ['ls-files', '--others', '-i', '--exclude-standard', '-z'],
      proc.exitCode ?? 1,
      Buffer.from(proc.stdout).toString('utf8'),
      Buffer.from(proc.stderr).toString('utf8')
    )
  }

  const raw = Buffer.from(proc.stdout).toString('utf8')
  return raw
    .split('\0')
    .map((s) => s.trim())
    .filter(Boolean)
}

function normalizePattern(pattern: string): string {
  return pattern.replaceAll('\\', '/').trim()
}

function matchesPattern(path: string, pattern: string): boolean {
  const normalized = normalizePattern(pattern)
  if (!normalized) return false

  if (normalized.endsWith('/')) {
    const dir = normalized.slice(0, -1)
    return path === dir || path.startsWith(`${dir}/`)
  }

  return (
    minimatch(path, normalized, { dot: true }) ||
    path === normalized ||
    path.startsWith(`${normalized}/`)
  )
}

function isExcluded(path: string, excludes: string[], includes: string[]): boolean {
  const excluded = excludes.some((pattern) => matchesPattern(path, pattern))
  if (!excluded) return false

  const explicitlyIncluded = includes.some((pattern) => matchesPattern(path, pattern))
  return !explicitlyIncluded
}

async function copyEntry(source: string, destination: string): Promise<void> {
  const info = await lstat(source)
  if (info.isSymbolicLink()) {
    const target = await readlink(source)
    await mkdir(dirname(destination), { recursive: true })
    await symlink(target, destination)
    return
  }

  if (info.isDirectory()) {
    await mkdir(destination, { recursive: true })
    return
  }

  await mkdir(dirname(destination), { recursive: true })
  await copyFile(source, destination)
}

export async function copyIgnoredArtifacts(options: CopyIgnoredOptions): Promise<CopyIgnoredSummary> {
  const candidates = runGitIgnoredList(options.repoRoot)
  let copied = 0
  let skippedExisting = 0
  let excluded = 0

  for (const relPath of candidates) {
    const normalizedRel = relPath.replaceAll('\\', '/')

    if (isExcluded(normalizedRel, options.excludes, options.includes)) {
      excluded++
      continue
    }

    const sourcePath = join(options.repoRoot, relPath)
    const destinationPath = join(options.destinationRoot, relPath)

    const destinationInfo = await lstat(destinationPath).catch(() => null)
    if (destinationInfo && options.skipExisting) {
      skippedExisting++
      continue
    }

    await copyEntry(sourcePath, destinationPath)
    copied++
  }

  return {
    totalCandidates: candidates.length,
    copied,
    skippedExisting,
    excluded
  }
}
