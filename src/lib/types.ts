export type SetupMode = 'auto-install' | 'custom'

export interface AgentreeConfig {
  branchPrefix: string
  copyIgnoredEnabled: boolean
  copyIgnoredSource: 'gitignored'
  copyConflictPolicy: 'skip-existing'
  defaultExcludes: string[]
  extraIncludes: string[]
  extraExcludes: string[]
  setupEnabled: boolean
  setupMode: SetupMode
  setupScripts: string[]
  strictDefault: boolean
  rollbackOnFailDefault: boolean
  completionEnabled: boolean
}

export interface ConfigOverrideFlags {
  copyIgnored?: boolean
  strict?: boolean
  rollbackOnFail?: boolean
  include?: string
  exclude?: string
}

export interface WorkspaceCreateFlags extends ConfigOverrideFlags {
  from?: string
  dest?: string
  interactive: boolean
  json: boolean
}

export interface WorkspaceListFlags {
  json: boolean
  all: boolean
}

export interface WorkspaceRemoveFlags {
  yes: boolean
  deleteBranch: boolean
  json: boolean
}

export interface GitPushFlags {
  remote: string
  json: boolean
}

export interface GitPrFlags {
  json: boolean
}

export interface CreateStepSummary {
  copied?: {
    totalCandidates: number
    copied: number
    skippedExisting: number
    excluded: number
  }
  setup?: {
    commands: string[]
    completed: number
  }
}

export interface WorktreeInfo {
  path: string
  head?: string
  branch?: string
  detached: boolean
}

export interface ManagedWorkspace extends WorktreeInfo {
  name: string
  managed: boolean
}
