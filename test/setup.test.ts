import { describe, expect, test } from 'bun:test'
import { resolveScriptShell } from '../src/lib/setup.js'

describe('setup shell selection', () => {
  test('uses cmd.exe on win32', () => {
    const shell = resolveScriptShell('win32')
    expect(shell.command).toBe('cmd.exe')
    expect(shell.args).toEqual(['/d', '/s', '/c'])
  })

  test('uses sh on non-windows', () => {
    const shell = resolveScriptShell('darwin')
    expect(shell.command).toBe('sh')
    expect(shell.args).toEqual(['-lc'])
  })
})
