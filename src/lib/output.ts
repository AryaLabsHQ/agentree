import type { ManagedWorkspace } from './types.js'

export function printJson(payload: unknown): void {
  console.log(JSON.stringify(payload, null, 2))
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
