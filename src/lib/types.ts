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

export interface NewCommandFlags {
  from?: string
  dest?: string
  push: boolean
  pr: boolean
  interactive: boolean
  json: boolean
  copyIgnored?: boolean
  strict?: boolean
  rollbackOnFail?: boolean
  include?: string
  exclude?: string
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
  push?: {
    remote: string
    branch: string
  }
  pr?: {
    created: boolean
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
