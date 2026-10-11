import { readFileSync, readdirSync, statSync } from 'node:fs'
import { join, relative, sep } from 'node:path'
import type { TarFile } from './tar'

export class UserError extends Error {}

export const mb = (n: number) => (n / 1024 / 1024).toFixed(1) + ' MB'

/** `Acme Docs` → `acme-docs`. */
export function slugify(s: string): string {
  return s
    .normalize('NFKD')
    .replace(/[̀-ͯ]/g, '')
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

/** gitdocai.json + pages + assets/, skipping dotfiles and node_modules. */
export function collect(dir: string): TarFile[] {
  const out: TarFile[] = []
  const walk = (d: string) => {
    for (const name of readdirSync(d)) {
      if (name.startsWith('.') || name === 'node_modules') continue
      const p = join(d, name)
      const st = statSync(p)
      if (st.isDirectory()) walk(p)
      else if (st.isFile()) {
        const rel = relative(dir, p).split(sep).join('/')
        // Root-level agent/readme files are instructions for tools and people,
        // not pages of the site.
        const notPage = /^(AGENTS|CLAUDE|README)\.md$/i.test(rel)
        const keep = !notPage && (rel === 'gitdocai.json' || rel.startsWith('assets/') || /\.mdx?$/i.test(rel))
        if (keep) out.push({ name: rel, data: readFileSync(p), mtime: st.mtimeMs })
      }
    }
  }
  walk(dir)
  return out
}

/** The address a hostname is reachable at, given the server it was published to. */
export function siteUrl(server: string, hostname: string): string {
  const s = new URL(server)
  const local = hostname === 'localhost' || hostname.endsWith('.localhost')
  return `${s.protocol}//${hostname}${local && s.port ? ':' + s.port : ''}`
}

export function parseArgs(argv: string[]) {
  const flags: Record<string, string | boolean> = {}
  const pos: string[] = []
  const bool = new Set(['dry-run', 'help', 'print', 'uninstall', 'force', 'open'])
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i]
    if (a.startsWith('--')) {
      const [k, v] = a.slice(2).split('=')
      if (v !== undefined) flags[k] = v
      else if (bool.has(k)) flags[k] = true
      else flags[k] = argv[++i] ?? ''
    } else pos.push(a)
  }
  return { flags, pos }
}
