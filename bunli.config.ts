import { defineConfig } from '@bunli/core'
import { completionsPlugin } from '@bunli/plugin-completions'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const rootDir = dirname(fileURLToPath(import.meta.url))

export default defineConfig({
  name: 'agentree',
  version: '1.0.0',
  description: 'Create and manage isolated git worktrees for AI coding agents',
  plugins: [
    completionsPlugin({
      generatedPath: join(rootDir, '.bunli', 'commands.gen.ts'),
      commandName: 'agentree',
      executable: 'agentree',
      includeAliases: true,
      includeGlobalFlags: true
    })
  ],
  commands: {
    entry: './src/cli.ts',
    directory: './src/commands'
  },
  build: {
    entry: './src/cli.ts',
    outdir: './dist',
    targets: [],
    compress: false,
    minify: false,
    sourcemap: true
  },
  dev: {
    watch: true,
    inspect: false
  },
  test: {
    pattern: ['**/*.test.ts'],
    watch: false,
    coverage: false
  },
  release: {
    npm: true,
    github: true,
    tagFormat: 'v{{version}}'
  }
})
