import { afterEach, describe, expect, test } from 'bun:test'
import { printJsonError, printJsonSuccess } from '../src/lib/output.js'

const originalLog = console.log

afterEach(() => {
  console.log = originalLog
})

describe('json output envelope', () => {
  test('success envelope shape', () => {
    let captured = ''
    console.log = (value?: unknown) => {
      captured = String(value ?? '')
    }

    printJsonSuccess('workspace list', { workspaces: [] })
    const parsed = JSON.parse(captured)

    expect(parsed.schemaVersion).toBe('1.0.0')
    expect(parsed.ok).toBe(true)
    expect(parsed.command).toBe('workspace list')
    expect(parsed.data).toEqual({ workspaces: [] })
    expect(parsed.warnings).toEqual([])
    expect(parsed.error).toBeUndefined()
  })

  test('error envelope shape', () => {
    let captured = ''
    console.log = (value?: unknown) => {
      captured = String(value ?? '')
    }

    printJsonError('workspace create', 'boom', { reason: 'test' })
    const parsed = JSON.parse(captured)

    expect(parsed.schemaVersion).toBe('1.0.0')
    expect(parsed.ok).toBe(false)
    expect(parsed.command).toBe('workspace create')
    expect(parsed.data).toBeNull()
    expect(parsed.warnings).toEqual([])
    expect(parsed.error.message).toBe('boom')
    expect(parsed.error.details).toEqual({ reason: 'test' })
  })
})
