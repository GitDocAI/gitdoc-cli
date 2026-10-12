import { existsSync } from 'node:fs'
import { join, resolve } from 'node:path'
import {
  connectedServers, normalizeServer, readProject, saveToken, tokenFor, writeProject,
} from './config'
import { dev } from './dev'
import { init } from './init'
import { setup } from './setup'
import { tarGz } from './tar'
import { UserError, collect, mb, parseArgs, siteUrl } from './util'

const USAGE = `gitdoc — write, preview and publish documentation with GitDocAI

Usage:
  gitdoc init [dir] [--name "My Docs"] [--hostname docs.acme.com]
      Create a docs site in a folder (manifest, a first page, AGENTS.md).
  gitdoc dev [dir] [--port 3000]
      Live preview: renders the folder and reloads the browser on every save.
  gitdoc connect <server-url> [--token <token>] [--dir <docs>]
      Save the credentials for a server and remember it for this folder.
      The token can also come from GITDOC_TOKEN or stdin.
  gitdoc publish [dir] [--server <url>] [--hostname <host>] [--dry-run]
      Package the folder (gitdocai.json, .md/.mdx pages, assets/) and publish it.
  gitdoc sites [--server <url>]
      List the documentations the connected server hosts.
  gitdoc setup [--print] [--uninstall]
      Install the GitDocAI skill for Claude Code (~/.claude/skills).

Which site is which: a server hosts many documentations, each answering on its
own hostname (<name>.localhost locally, your domain in production). The folder
remembers it in .gitdoc/project.json after the first publish.

Environment:
  GITDOC_TOKEN    token to use instead of the saved one (CI)
  GITDOC_SERVER   server to use when none is configured for the folder
`

function resolveServer(dir: string | null, flags: Record<string, string | boolean>): string {
  const project = dir ? readProject(dir) : {}
  const raw =
    (typeof flags.server === 'string' && flags.server) ||
    project.server ||
    process.env.GITDOC_SERVER ||
    (connectedServers().length === 1 ? connectedServers()[0] : '')
  if (!raw) {
    throw new UserError(
      connectedServers().length > 1
        ? 'several servers are connected; pick one with --server <url>:\n  ' + connectedServers().join('\n  ')
        : 'no server configured. Run: gitdoc connect <server-url> --token <token>',
    )
  }
  return normalizeServer(raw)
}

async function connect(pos: string[], flags: Record<string, string | boolean>) {
  const urlArg = pos[0]
  if (!urlArg) throw new UserError('usage: gitdoc connect <server-url> [--token <token>]')
  let server: string
  try {
    server = normalizeServer(urlArg)
  } catch {
    throw new UserError(`not a valid URL: ${urlArg}`)
  }
  const token = (typeof flags.token === 'string' && flags.token) || process.env.GITDOC_TOKEN || (await readStdin())
  if (!token) {
    throw new UserError(
      'no token given. Pass --token, set GITDOC_TOKEN, or pipe it in:\n  echo "$TOKEN" | gitdoc connect ' + server,
    )
  }
  const file = saveToken(server, token)
  const dir = resolve(typeof flags.dir === 'string' ? flags.dir : '.')
  let projectFile = ''
  if (existsSync(join(dir, 'gitdocai.json'))) projectFile = writeProject(dir, { server })
  console.log(`Connected to ${server}`)
  console.log(`  token saved in ${file} (owner-only)`)
  if (projectFile) console.log(`  this folder will publish to it (.gitdoc/project.json)`)
}

async function readStdin(): Promise<string> {
  if (process.stdin.isTTY) return ''
  const chunks: Buffer[] = []
  for await (const c of process.stdin) chunks.push(c as Buffer)
  return Buffer.concat(chunks).toString('utf-8').trim()
}

async function publish(pos: string[], flags: Record<string, string | boolean>) {
  const dir = resolve(pos[0] ?? '.')
  if (!existsSync(join(dir, 'gitdocai.json'))) {
    throw new UserError(`no gitdocai.json in ${dir} — is this a docs folder?`)
  }
  const project = readProject(dir)
  const server = resolveServer(dir, flags)
  const token = tokenFor(server)
  if (!token) throw new UserError(`not connected to ${server}. Run: gitdoc connect ${server} --token <token>`)

  const files = collect(dir)
  const pages = files.filter((f) => /\.mdx?$/i.test(f.name)).length
  const assets = files.filter((f) => f.name.startsWith('assets/')).length
  const body = tarGz(files)
  const hostname = (typeof flags.hostname === 'string' && flags.hostname) || project.hostname
  console.log(`${pages} pages, ${assets} asset files → ${mb(body.length)} compressed`)
  console.log(`server:   ${server}`)
  console.log(`hostname: ${hostname ?? '(derived from the site name on first publish)'}`)
  if (flags['dry-run']) {
    console.log('dry run: nothing was sent.')
    return
  }

  const url = new URL('/api/publish', server)
  if (hostname) url.searchParams.set('hostname', hostname)

  let res: Response
  try {
    res = await fetch(url, {
      method: 'POST',
      headers: { authorization: `Bearer ${token}`, 'content-type': 'application/gzip' },
      body,
      signal: AbortSignal.timeout(10 * 60_000),
    })
  } catch (e) {
    throw new UserError(`could not reach ${server}: ${(e as Error).message}`)
  }

  const text = await res.text()
  let data: any = null
  try {
    data = JSON.parse(text)
  } catch {
    /* non-JSON error body */
  }
  if (!res.ok) {
    const why: Record<number, string> = {
      401: `the server rejected the token. Reconnect: gitdoc connect ${server} --token <token>`,
      404: 'the server has publishing disabled (PUBLISH_TOKEN is not set there).',
      409: 'another publish is already running; try again in a moment.',
      413: 'the archive is larger than the server allows (PUBLISH_MAX_MB).',
    }
    throw new UserError(`publish failed (${res.status}): ${why[res.status] ?? data?.error ?? text.slice(0, 300)}`)
  }

  const r = data?.result ?? {}
  console.log(`\nPublished (${r.status ?? 'ok'}): ${r.rendered_pages ?? '?'} pages in ${r.duration_ms ?? '?'} ms`)
  // Ingestion happens on the server as part of the publish: say what it did.
  const ix = data?.search
  if (ix?.error) {
    console.log(`  search: NOT indexed (${ix.error}). The site is live; fix the search backend and publish again.`)
  } else if (ix) {
    const how = ix.semantic ? 'with vectors' : 'text only'
    console.log(`  search: indexed ${ix.indexed} pages (${ix.backend}, ${how})`)
    if (ix.warning) console.log(`  search: ${ix.warning}`)
  }
  if (data?.hostname) {
    console.log(`  site: ${siteUrl(server, data.hostname)}`)
    // Pin the folder to its site, so later publishes can't drift to another one.
    if (!project.hostname) {
      try {
        writeProject(dir, { hostname: data.hostname })
      } catch {
        /* read-only folder: publish still worked */
      }
    }
  }
  const errs = (r.errors ?? []) as { where: string; message: string }[]
  const real = errs.filter((e) => e.where !== 'r2-cleanup')
  if (real.length) {
    console.log(`\n${real.length} problem(s):`)
    for (const e of real.slice(0, 20)) console.log(`  - ${e.message}`)
    if (real.length > 20) console.log(`  … and ${real.length - 20} more`)
  }
}

async function sites(flags: Record<string, string | boolean>) {
  const server = resolveServer(existsSync('gitdocai.json') ? resolve('.') : null, flags)
  const token = tokenFor(server)
  if (!token) throw new UserError(`not connected to ${server}. Run: gitdoc connect ${server} --token <token>`)
  let res: Response
  try {
    res = await fetch(new URL('/api/sites', server), { headers: { authorization: `Bearer ${token}` } })
  } catch (e) {
    throw new UserError(`could not reach ${server}: ${(e as Error).message}`)
  }
  if (res.status === 401) throw new UserError(`the server rejected the token. Reconnect: gitdoc connect ${server} --token <token>`)
  if (!res.ok) throw new UserError(`could not list sites (${res.status}); is this server a GitDocAI self-hosted install?`)
  const { sites } = (await res.json()) as {
    sites: { hostname: string; name: string; pages?: number; isPrivate: boolean }[]
  }
  if (!sites.length) {
    console.log(`${server} has no documentation published yet.`)
    return
  }
  const here = readProject(resolve('.')).hostname
  console.log(`${server} hosts ${sites.length} documentation${sites.length === 1 ? '' : 's'}:\n`)
  for (const s of sites) {
    const mark = s.hostname === here ? '*' : ' '
    console.log(
      `${mark} ${siteUrl(server, s.hostname).padEnd(44)} ${(s.name || '(unnamed)').padEnd(24)} ` +
        `${s.pages ?? '?'} pages${s.isPrivate ? '  private' : ''}`,
    )
  }
  if (here) console.log('\n* = the site this folder publishes to')
}

async function main(): Promise<number> {
  const [cmd, ...rest] = process.argv.slice(2)
  const { flags, pos } = parseArgs(rest)
  if (!cmd || cmd === 'help' || cmd === '--help' || cmd === '-h' || flags.help) {
    process.stdout.write(USAGE)
    return cmd ? 0 : 2
  }
  try {
    if (cmd === 'init') init(pos, flags)
    else if (cmd === 'dev') await dev(pos, flags)
    else if (cmd === 'connect') await connect(pos, flags)
    else if (cmd === 'publish') await publish(pos, flags)
    else if (cmd === 'sites') await sites(flags)
    else if (cmd === 'setup') setup(flags)
    else {
      process.stderr.write(`unknown command: ${cmd}\n\n${USAGE}`)
      return 2
    }
    return 0
  } catch (e) {
    if (e instanceof UserError) {
      process.stderr.write(`error: ${e.message}\n`)
      return 1
    }
    throw e
  }
}

main().then(
  (c) => process.exit(c),
  (e) => {
    process.stderr.write(`gitdoc crashed: ${e instanceof Error ? e.stack : e}\n`)
    process.exit(1)
  },
)
