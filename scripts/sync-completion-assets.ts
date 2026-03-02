import { mkdir, readFile, writeFile } from 'node:fs/promises'
import { join } from 'node:path'

const root = process.cwd()
const generatedSrc = join(root, '.bunli', 'commands.gen.ts')
const generatedDist = join(root, 'dist', '.bunli', 'commands.gen.ts')

await mkdir(join(root, 'dist', '.bunli'), { recursive: true })
await mkdir(join(root, 'dist', 'src', 'commands'), { recursive: true })

const generated = await readFile(generatedSrc, 'utf8')
await writeFile(generatedDist, generated)

const result = await Bun.build({
  entrypoints: [
    join(root, 'src', 'commands', 'new.ts'),
    join(root, 'src', 'commands', 'ls.ts'),
    join(root, 'src', 'commands', 'rm.ts')
  ],
  outdir: join(root, 'dist', 'src', 'commands'),
  target: 'bun',
  format: 'esm',
  bundle: true,
  sourcemap: 'none',
  minify: false
})

if (!result.success) {
  for (const log of result.logs) {
    console.error(log)
  }
  throw new Error('Failed to build completion command modules')
}
