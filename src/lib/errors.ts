export class CommandExecutionError extends Error {
  constructor(
    public readonly command: string,
    public readonly args: string[],
    public readonly exitCode: number,
    public readonly stdout: string,
    public readonly stderr: string
  ) {
    super(`Command failed: ${command} ${args.join(' ')}`)
    this.name = 'CommandExecutionError'
  }
}

export class PostStepError extends Error {
  constructor(public readonly step: string, cause: unknown) {
    super(`Post-step failed at ${step}: ${cause instanceof Error ? cause.message : String(cause)}`)
    this.name = 'PostStepError'
  }
}

export class RollbackError extends Error {
  constructor(
    public readonly original: Error,
    public readonly rollbackFailures: Error[]
  ) {
    super(
      `Create failed: ${original.message}. Rollback failures: ${rollbackFailures
        .map((e) => e.message)
        .join('; ')}`
    )
    this.name = 'RollbackError'
  }
}
