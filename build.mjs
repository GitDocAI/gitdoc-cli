// Bundle the CLI into a single file: dist/gitdoc.js
import { build } from 'esbuild'
import { chmodSync } from 'node:fs'

await build({
  entryPoints: ['src/gitdoc.ts'],
  bundle: true,
  platform: 'node',
  target: 'node20',
  format: 'esm',
  outfile: 'dist/gitdoc.js',
  banner: { js: '#!/usr/bin/env node' },
  loader: { '.md': 'text' },
  logLevel: 'warning',
})
chmodSync('dist/gitdoc.js', 0o755)
console.log('built dist/gitdoc.js')
