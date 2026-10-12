import { spawn } from 'node:child_process'
import { createHash } from 'node:crypto'
import { createWriteStream, existsSync, mkdirSync, mkdtempSync, readdirSync, readFileSync, renameSync, rmSync } from 'node:fs'
import { homedir } from 'node:os'
import { join } from 'node:path'
import { Readable } from 'node:stream'
import { pipeline } from 'node:stream/promises'
import { configDir } from './config'
import { UserError, mb } from './util'

export interface Engine {
  /** The server's entry point: `engine/server/entry.mjs` of a gitdoc build. */
  server: string
  /** The renderer bundle the server spawns when publishing: `engine/render.js`. */
  renderCli: string
}

/** The npm package that carries the engine (a build of github.com/GitDocAI/GitDoc). */
const PACKAGE = process.env.GITDOC_ENGINE_PACKAGE || '@gitdocai/engine'
/** Engines of this major version speak the protocol this CLI expects. */
const ENGINE_MAJOR = 0
const MAX_DOWNLOAD = 150 * 1024 * 1024

function cacheDir(): string {
  if (process.env.GITDOC_CACHE_DIR) return process.env.GITDOC_CACHE_DIR
  return join(process.env.XDG_CACHE_HOME || join(homedir(), '.cache'), 'gitdoc')
}

function fromDir(dir: string): Engine | null {
  try {
    const m = JSON.parse(readFileSync(join(dir, 'engine.json'), 'utf-8')) as { server?: string; renderCli?: string }
    if (!m.server || !m.renderCli) return null
    const e = { server: join(dir, m.server), renderCli: join(dir, m.renderCli) }
    return existsSync(e.server) && existsSync(e.renderCli) ? e : null
  } catch {
    return null
  }
}

function run(cmd: string, args: string[]): Promise<{ code: number; out: string; err: string }> {
  return new Promise((res, rej) => {
    const p = spawn(cmd, args, { stdio: ['ignore', 'pipe', 'pipe'] })
    let out = ''
    let err = ''
    p.stdout.on('data', (d) => (out += d))
    p.stderr.on('data', (d) => (err += d))
    p.on('error', rej)
    p.on('close', (code) => res({ code: code ?? 1, out, err }))
  })
}

/** Fetch the engine from the npm registry into the cache; returns its directory. */
async function download(): Promise<string> {
  const registry = (process.env.GITDOC_REGISTRY || 'https://registry.npmjs.org').replace(/\/+$/, '')
  const meta = await fetch(`${registry}/${PACKAGE.replace('/', '%2F')}/latest`, {
    headers: { accept: 'application/json' },
    signal: AbortSignal.timeout(30_000),
  })
  if (!meta.ok) throw new Error(`the registry answered ${meta.status} for ${PACKAGE}`)
  const info = (await meta.json()) as {
    version: string
    dist?: { tarball?: string; integrity?: string; shasum?: string }
  }
  const major = Number(info.version.split('.')[0])
  if (major !== ENGINE_MAJOR) {
    throw new Error(
      `the latest ${PACKAGE} is ${info.version}, which this gitdoc does not support; update it with: npm i -g @gitdocai/cli`,
    )
  }
  const dest = join(cacheDir(), 'engine', info.version)
  if (fromDir(dest)) return dest

  const tarball = info.dist?.tarball
  const expected = info.dist?.integrity ?? (info.dist?.shasum ? `sha1-hex:${info.dist.shasum}` : '')
  if (!tarball || !expected) throw new Error('the registry gave no checksum for the engine; refusing to run it')

  mkdirSync(join(cacheDir(), 'engine'), { recursive: true })
  const work = mkdtempSync(join(cacheDir(), 'engine', '.dl-'))
  try {
    const res = await fetch(tarball, { signal: AbortSignal.timeout(300_000) })
    if (!res.ok || !res.body) throw new Error(`download failed (${res.status})`)
    const sri = expected.startsWith('sha1-hex:') ? 'sha1' : expected.split('-')[0]
    const hash = createHash(sri)
    const file = join(work, 'engine.tgz')
    let size = 0
    const src = Readable.fromWeb(res.body as never)
    src.on('data', (c: Buffer) => {
      size += c.length
      hash.update(c)
      if (size > MAX_DOWNLOAD) src.destroy(new Error('the engine download is larger than expected'))
    })
    await pipeline(src, createWriteStream(file))
    const got = expected.startsWith('sha1-hex:') ? hash.digest('hex') : hash.digest('base64')
    const want = expected.startsWith('sha1-hex:') ? expected.slice(9) : expected.split('-').slice(1).join('-')
    if (got !== want) throw new Error('the downloaded engine does not match its checksum; refusing to run it')

    // Only plain files and directories, nothing absolute or with `..`.
    const list = await run('tar', ['-tvzf', file])
    if (list.code !== 0) throw new Error('the engine archive could not be read')
    for (const line of list.out.split('\n').filter(Boolean)) {
      const name = line.replace(/^\S+\s+\S+\s+\S+\s+\S+\s+\S+\s+\S+\s+\S+\s+/, '')
      if ((line[0] !== '-' && line[0] !== 'd') || /(^|\/)\.\.(\/|$)/.test(name) || name.startsWith('/')) {
        throw new Error(`unexpected entry in the engine archive: ${name}`)
      }
    }
    const staging = join(work, 'out')
    mkdirSync(staging)
    // npm tarballs wrap everything in `package/`.
    const ex = await run('tar', ['-xzf', file, '-C', staging, '--strip-components=1', '--no-same-owner'])
    if (ex.code !== 0) throw new Error(`could not unpack the engine: ${ex.err.trim()}`)
    if (!fromDir(staging)) throw new Error('the downloaded package is not a gitdoc engine')

    console.error(`  engine ${info.version} ready (${mb(size)}), cached in ${dest}`)
    try {
      renameSync(staging, dest)
    } catch {
      if (!fromDir(dest)) throw new Error(`could not place the engine in ${dest}`)
    }
    return dest
  } finally {
    rmSync(work, { recursive: true, force: true })
  }
}

/** Newest engine already in the cache (so `gitdoc dev` works offline). */
function cached(): Engine | null {
  const root = join(cacheDir(), 'engine')
  let names: string[]
  try {
    names = readdirSync(root)
  } catch {
    return null
  }
  const versions = names
    .filter((n) => /^\d+\.\d+\.\d+/.test(n) && Number(n.split('.')[0]) === ENGINE_MAJOR)
    .sort((a, b) => b.localeCompare(a, undefined, { numeric: true }))
  for (const v of versions) {
    const e = fromDir(join(root, v))
    if (e) return e
  }
  return null
}

/**
 * `gitdoc dev` runs the rendering engine (the gitdoc server, a separate
 * program with its own license) as a child process. It is found, in order:
 *
 *   1. GITDOC_ENGINE_SERVER and GITDOC_ENGINE_RENDER
 *   2. `engine.json` in the config dir: { "server": "...", "renderCli": "..." }
 *   3. the cache (~/.cache/gitdoc/engine/<version>)
 *   4. downloaded once from npm (@gitdocai/engine), checksum-verified, then cached.
 *      Disable with GITDOC_NO_DOWNLOAD=1.
 *
 * The CLI never bundles or imports the engine.
 */
export async function loadEngine(): Promise<Engine> {
  let e: Partial<Engine> = {}
  if (process.env.GITDOC_ENGINE_SERVER && process.env.GITDOC_ENGINE_RENDER) {
    e = { server: process.env.GITDOC_ENGINE_SERVER, renderCli: process.env.GITDOC_ENGINE_RENDER }
  } else {
    try {
      e = JSON.parse(readFileSync(join(configDir(), 'engine.json'), 'utf-8')) as Engine
    } catch {
      /* fall through to the cache */
    }
  }
  if (e.server && e.renderCli && existsSync(e.server) && existsSync(e.renderCli)) return e as Engine

  const local = cached()
  if (local) return local // no network once it is cached

  let reason = ''
  if (process.env.GITDOC_NO_DOWNLOAD) {
    reason = 'automatic download is disabled (GITDOC_NO_DOWNLOAD)'
  } else {
    try {
      console.error(`  fetching the rendering engine (${PACKAGE}), once…`)
      const dir = await download()
      const got = fromDir(dir)
      if (got) return got
    } catch (err) {
      reason = (err as Error).message
    }
  }
  throw new UserError(
    'the rendering engine was not found, and `gitdoc dev` needs it.\n' +
      (reason ? `  ${reason}\n` : '') +
      '  Point the CLI at a build of the gitdoc server:\n' +
      '    export GITDOC_ENGINE_SERVER=/path/to/engine/server/entry.mjs\n' +
      '    export GITDOC_ENGINE_RENDER=/path/to/engine/render.js\n' +
      `  or save them as JSON in ${join(configDir(), 'engine.json')}.\n` +
      '  (Publishing to a server does not need the engine.)',
  )
}
