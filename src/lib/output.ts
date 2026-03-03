import type { ManagedWorkspace } from './types.js'

export const JSON_SCHEMA_VERSION = '1.0.0'

interface JsonEnvelope<T> {
  schemaVersion: string
  ok: boolean
  command: string
  data: T | null
  warnings: string[]
  error?: {
    message: string
    details?: unknown
  }
}

function printEnvelope<T>(envelope: JsonEnvelope<T>): void {
  console.log(JSON.stringify(envelope, null, 2))
}

export function printJsonSuccess<T>(command: string, data: T, warnings: string[] = []): void {
  printEnvelope({
    schemaVersion: JSON_SCHEMA_VERSION,
    ok: true,
    command,
    data,
    warnings
  })
}

export function printJsonError(
  command: string,
  message: string,
  details?: unknown,
  warnings: string[] = []
): void {
  printEnvelope({
    schemaVersion: JSON_SCHEMA_VERSION,
    ok: false,
    command,
    data: null,
    warnings,
    error: {
      message,
      ...(details === undefined ? {} : { details })
    }
  })
}

export function printWorkspacesTable(workspaces: ManagedWorkspace[]): void {
  if (workspaces.length === 0) {
    console.log('No workspaces found.')
    return
  }

  console.log('Name                 Branch                      Path')
  console.log('-------------------- --------------------------- --------------------------------')
  for (const workspace of workspaces) {
    const name = workspace.name.padEnd(20)
    const branch = (workspace.branch ?? '(detached)').padEnd(27)
    console.log(`${name} ${branch} ${workspace.path}`)
  }
}
