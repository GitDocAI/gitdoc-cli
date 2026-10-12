// Contract tests: the CLI against a fake server. They need no rendering engine.
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { spawn, execFileSync } from 'node:child_process'
import { createServer } from 'node:http'
import { mkdtempSync, mkdirSync, writeFileSync, readFileSync, existsSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join, resolve } from 'node:path'

const BIN = resolve('dist/gitdoc.js')
const tmp = () => mkdtempSync(join(tmpdir(), 'gitdoc-test-'))

// Async on purpose: the fake server lives in this same process, so a blocking
// spawnSync would stop it from ever answering the CLI.
function run(args, { cwd, env = {} } = {}) {
  return new Promise((ok) => {
    const p = spawn(process.execPath, [BIN, ...args], {
      cwd,
      env: { PATH: process.env.PATH, ...env },
      stdio: ['ignore', 'pipe', 'pipe'],
    })
    let out = ''
    let err = ''
    p.stdout.on('data', (d) => (out += d))
    p.stderr.on('data', (d) => (err += d))
    p.on('close', (code) => ok({ code, out, err }))
  })
}

/** A stand-in for the gitdoc server: records what it receives. */
function fakeServer(handler) {
  const seen = []
  const srv = createServer(async (req, res) => {
    const chunks = []
    for await (const c of req) chunks.push(c)
    const body = Buffer.concat(chunks)
    seen.push({ method: req.method, url: req.url, headers: req.headers, body })
    handler(req, res, body)
  })
  return new Promise((ok) => srv.listen(0, '127.0.0.1', () => ok({ srv, seen, url: `http://127.0.0.1:${srv.address().port}` })))
}

test('init creates a site and refuses to overwrite one', async () => {
  const dir = join(tmp(), 'docs')
  const a = await run(['init', dir, '--name', 'Acme Docs'])
  assert.equal(a.code, 0)
  const m = JSON.parse(readFileSync(join(dir, 'gitdocai.json'), 'utf-8'))
  assert.equal(m.format, 1)
  assert.equal(m.name, 'Acme Docs')
  assert.equal(m['x-site-key'], 'acme-docs')
  assert.ok(existsSync(join(dir, 'v1/getting-started/introduction.mdx')))
  assert.ok(existsSync(join(dir, 'AGENTS.md')))
  assert.equal(readFileSync(join(dir, 'CLAUDE.md'), 'utf-8').trim(), '@AGENTS.md')

  const again = await run(['init', dir])
  assert.equal(again.code, 1)
  assert.match(again.err, /already has a gitdocai\.json/)
})

test('init does not put a title heading in the starter page', async () => {
  const dir = join(tmp(), 'docs')
  await run(['init', dir, '--name', 'X'])
  const page = readFileSync(join(dir, 'v1/getting-started/introduction.mdx'), 'utf-8')
  assert.doesNotMatch(page.replace(/^---[\s\S]*?---\n/, ''), /^# /m)
})

test('publish without a server explains how to connect', async () => {
  const home = tmp()
  const dir = join(tmp(), 'docs')
  await run(['init', dir, '--name', 'X'])
  const r = await run(['publish', dir], { env: { GITDOC_CONFIG_DIR: home } })
  assert.equal(r.code, 1)
  assert.match(r.err, /gitdoc connect/)
})

test('connect saves the token owner-only and publish sends a valid archive', async () => {
  const config = tmp()
  const dir = join(tmp(), 'docs')
  await run(['init', dir, '--name', 'Acme Docs'])
  mkdirSync(join(dir, 'assets'), { recursive: true })
  writeFileSync(join(dir, 'assets/logo.png'), 'png')
  writeFileSync(join(dir, 'README.md'), 'not a page')
  mkdirSync(join(dir, 'node_modules/x'), { recursive: true })
  writeFileSync(join(dir, 'node_modules/x/index.mdx'), 'ignore me')

  const { srv, seen, url } = await fakeServer((req, res) => {
    res.writeHead(200, { 'content-type': 'application/json' })
    res.end(JSON.stringify({ ok: true, hostname: 'acme-docs.localhost', result: { status: 'ok', rendered_pages: 1, duration_ms: 5, errors: [] }, search: { backend: 'opensearch', indexed: 1, semantic: true } }))
  })
  try {
    const env = { GITDOC_CONFIG_DIR: config }
    const c = await run(['connect', url, '--token', 's3cret', '--dir', dir], { env })
    assert.equal(c.code, 0, c.err)
    const mode = execFileSync('stat', process.platform === 'darwin' ? ['-f', '%Lp', join(config, 'credentials.json')] : ['-c', '%a', join(config, 'credentials.json')], { encoding: 'utf-8' }).trim()
    assert.equal(mode, '600')
    assert.ok(!readFileSync(join(dir, '.gitdoc/project.json'), 'utf-8').includes('s3cret'), 'token must not be written into the project')

    const p = await run(['publish', dir], { env })
    assert.equal(p.code, 0, p.err)
    assert.match(p.out, /acme-docs\.localhost/)
    assert.match(p.out, /search: indexed 1 pages \(opensearch, with vectors\)/, 'the publish reports its ingestion')

    const req = seen.find((s) => s.url.startsWith('/api/publish'))
    assert.ok(req, 'the server received a publish')
    assert.equal(req.method, 'POST')
    assert.equal(req.headers.authorization, 'Bearer s3cret')
    assert.equal(req.headers['content-type'], 'application/gzip')

    const file = join(tmp(), 'sent.tar.gz')
    writeFileSync(file, req.body)
    const names = execFileSync('tar', ['-tzf', file], { encoding: 'utf-8' }).split('\n').filter(Boolean)
    assert.ok(names.includes('gitdocai.json'))
    assert.ok(names.includes('v1/getting-started/introduction.mdx'))
    assert.ok(names.includes('assets/logo.png'))
    assert.ok(!names.some((n) => /AGENTS|CLAUDE|README|node_modules/.test(n)), `unexpected files sent: ${names}`)

    // the folder is now pinned to the site the server reported
    assert.equal(JSON.parse(readFileSync(join(dir, '.gitdoc/project.json'), 'utf-8')).hostname, 'acme-docs.localhost')
  } finally {
    srv.close()
  }
})

test('publish keeps very long file names (OpenAPI-style pages)', async () => {
  const config = tmp()
  const dir = join(tmp(), 'docs')
  await run(['init', dir, '--name', 'Long'])
  // A basename over 100 bytes cannot be split by ustar: it needs a PAX header.
  const long = 'delete-control-planeapiv1v1documentationorganization-iddocumentation-idversionversion-idsectionsection-identryentry-id'
  const nested = `v1/openapi/${'deep-folder-name-'.repeat(8)}`
  mkdirSync(join(dir, nested), { recursive: true })
  writeFileSync(join(dir, `v1/openapi/${long}.mdx`), '---\ntitle: Long\n---\n\nbody\n')
  writeFileSync(join(dir, nested, `${long}.mdx`), '---\ntitle: Deeper\n---\n\nbody\n')
  const { srv, seen, url } = await fakeServer((_q, res) => {
    res.writeHead(200, { 'content-type': 'application/json' })
    res.end(JSON.stringify({ ok: true, hostname: 'long.localhost', result: { status: 'ok', rendered_pages: 3, errors: [] } }))
  })
  try {
    const env = { GITDOC_CONFIG_DIR: config }
    await run(['connect', url, '--token', 't', '--dir', dir], { env })
    const p = await run(['publish', dir], { env })
    assert.equal(p.code, 0, p.err)
    const file = join(tmp(), 'long.tar.gz')
    writeFileSync(file, seen.find((x) => x.url.startsWith('/api/publish')).body)
    const names = execFileSync('tar', ['-tzf', file], { encoding: 'utf-8' }).split('\n').filter(Boolean)
    assert.ok(names.includes(`v1/openapi/${long}.mdx`), 'the long basename survives')
    assert.ok(names.includes(`${nested}/${long}.mdx`), 'and so does a long directory plus a long basename')
    // Extracting gives back exactly the file that was sent.
    const out = tmp()
    execFileSync('tar', ['-xzf', file, '-C', out])
    assert.match(readFileSync(join(out, nested, `${long}.mdx`), 'utf-8'), /title: Deeper/)
  } finally {
    srv.close()
  }
})

test('publish --dry-run sends nothing', async () => {
  const config = tmp()
  const dir = join(tmp(), 'docs')
  await run(['init', dir, '--name', 'X'])
  const { srv, seen, url } = await fakeServer((_q, res) => res.writeHead(500).end())
  try {
    await run(['connect', url, '--token', 't', '--dir', dir], { env: { GITDOC_CONFIG_DIR: config } })
    const r = await run(['publish', dir, '--dry-run'], { env: { GITDOC_CONFIG_DIR: config } })
    assert.equal(r.code, 0)
    assert.match(r.out, /nothing was sent/)
    assert.equal(seen.length, 0)
  } finally {
    srv.close()
  }
})

test('a rejected token is reported with the way out', async () => {
  const config = tmp()
  const dir = join(tmp(), 'docs')
  await run(['init', dir, '--name', 'X'])
  const { srv, url } = await fakeServer((_q, res) => {
    res.writeHead(401, { 'content-type': 'application/json' })
    res.end('{"error":"unauthorized"}')
  })
  try {
    await run(['connect', url, '--token', 'bad', '--dir', dir], { env: { GITDOC_CONFIG_DIR: config } })
    const r = await run(['publish', dir], { env: { GITDOC_CONFIG_DIR: config } })
    assert.equal(r.code, 1)
    assert.match(r.err, /rejected the token/)
  } finally {
    srv.close()
  }
})

test('credentials saved by earlier versions (~/.config/gitdocai) are still read', async () => {
  const xdg = tmp()
  mkdirSync(join(xdg, 'gitdocai'), { recursive: true })
  const dir = join(tmp(), 'docs')
  await run(['init', dir, '--name', 'X'])
  const { srv, seen, url } = await fakeServer((_q, res) => {
    res.writeHead(200, { 'content-type': 'application/json' })
    res.end(JSON.stringify({ ok: true, hostname: 'x.localhost', result: { status: 'ok', errors: [] } }))
  })
  try {
    writeFileSync(join(xdg, 'gitdocai/credentials.json'), JSON.stringify({ servers: { [new URL(url).origin]: { token: 'old-token' } } }))
    const r = await run(['publish', dir, '--server', url], { env: { XDG_CONFIG_HOME: xdg, HOME: xdg } })
    assert.equal(r.code, 0, r.err)
    assert.equal(seen[0].headers.authorization, 'Bearer old-token')
  } finally {
    srv.close()
  }
})

test('sites lists what the server hosts', async () => {
  const config = tmp()
  const { srv, url } = await fakeServer((_q, res) => {
    res.writeHead(200, { 'content-type': 'application/json' })
    res.end(JSON.stringify({ sites: [{ hostname: 'a.localhost', name: 'Site A', pages: 3, isPrivate: false }, { hostname: 'b.example.com', name: 'Site B', pages: 9, isPrivate: true }] }))
  })
  try {
    const env = { GITDOC_CONFIG_DIR: config }
    await run(['connect', url, '--token', 't'], { env, cwd: tmp() })
    const r = await run(['sites', '--server', url], { env, cwd: tmp() })
    assert.equal(r.code, 0, r.err)
    assert.match(r.out, /2 documentations/)
    assert.match(r.out, /Site A/)
    assert.match(r.out, /Site B.*private/)
  } finally {
    srv.close()
  }
})

test('setup installs the skill under the name gitdoc and replaces the old one', async () => {
  const claude = tmp()
  mkdirSync(join(claude, 'skills/gitdocai'), { recursive: true })
  writeFileSync(join(claude, 'skills/gitdocai/SKILL.md'), '---\nname: gitdocai\n---\nold')
  const r = await run(['setup'], { env: { CLAUDE_CONFIG_DIR: claude } })
  assert.equal(r.code, 0, r.err)
  const skill = readFileSync(join(claude, 'skills/gitdoc/SKILL.md'), 'utf-8')
  assert.match(skill, /^---\nname: gitdoc\n/)
  assert.ok(!existsSync(join(claude, 'skills/gitdocai')), 'the previous gitdocai skill is removed')
})

test('dev without the engine says how to point at it', async () => {
  const dir = join(tmp(), 'docs')
  await run(['init', dir, '--name', 'X'])
  const r = await run(['dev', dir], { env: { GITDOC_CONFIG_DIR: tmp(), GITDOC_CACHE_DIR: tmp(), GITDOC_NO_DOWNLOAD: '1' } })
  assert.equal(r.code, 1)
  assert.match(r.err, /GITDOC_ENGINE_SERVER/)
  assert.match(r.err, /download is disabled/)
})

/** A fake npm registry serving one engine package; counts what it is asked. */
async function fakeRegistry({ corrupt = false, version = '0.1.0' } = {}) {
  const { createHash } = await import('node:crypto')
  const pkg = tmp()
  mkdirSync(join(pkg, 'package/server'), { recursive: true })
  writeFileSync(join(pkg, 'package/package.json'), JSON.stringify({ name: '@gitdocai/engine', version }))
  writeFileSync(join(pkg, 'package/engine.json'), JSON.stringify({ server: 'server/entry.mjs', renderCli: 'render.js' }))
  writeFileSync(join(pkg, 'package/render.js'), '// renderer\n')
  // A stand-in server: answers the readiness probe and accepts a publish.
  writeFileSync(
    join(pkg, 'package/server/entry.mjs'),
    `import http from 'node:http'
http.createServer((q, r) => {
  r.setHeader('content-type', 'application/json')
  r.end(JSON.stringify({ ok: true, result: { rendered_pages: 1, duration_ms: 1, errors: [] } }))
}).listen(Number(process.env.PORT), process.env.HOST)\n`,
  )
  const tgz = join(tmp(), 'engine.tgz')
  execFileSync('tar', ['-czf', tgz, '-C', pkg, 'package'])
  const bytes = readFileSync(tgz)
  const integrity = 'sha512-' + createHash('sha512').update(corrupt ? 'something else' : bytes).digest('base64')
  const hits = []
  const f = await fakeServer((req, res) => {
    hits.push(req.url)
    if (req.url === '/@gitdocai%2Fengine/latest') {
      res.setHeader('content-type', 'application/json')
      res.end(JSON.stringify({ name: '@gitdocai/engine', version, dist: { tarball: `${f.url}/engine.tgz`, integrity } }))
    } else if (req.url === '/engine.tgz') {
      res.end(bytes)
    } else {
      res.writeHead(404).end()
    }
  })
  return { ...f, hits }
}

/** Start `gitdoc dev`, wait until it prints its URL, stop it; returns stdout/stderr. */
function devUntilReady(dir, env) {
  return new Promise((ok) => {
    const p = spawn(process.execPath, [BIN, 'dev', dir, '--port', '47311'], {
      env: { PATH: process.env.PATH, ...env },
      stdio: ['ignore', 'pipe', 'pipe'],
    })
    let out = ''
    let err = ''
    const stop = () => p.kill()
    p.stdout.on('data', (d) => {
      out += d
      if (out.includes('Local:')) stop()
    })
    p.stderr.on('data', (d) => (err += d))
    p.on('close', (code) => ok({ code, out, err }))
  })
}

test('dev downloads the engine once, verifies it, and reuses the cache', async () => {
  const reg = await fakeRegistry()
  const dir = join(tmp(), 'docs')
  await run(['init', dir, '--name', 'X'])
  const cache = tmp()
  const env = { GITDOC_CONFIG_DIR: tmp(), GITDOC_CACHE_DIR: cache, GITDOC_REGISTRY: reg.url }

  const first = await devUntilReady(dir, env)
  assert.match(first.out, /Local:/, first.err)
  assert.match(first.err, /engine 0\.1\.0 ready/)
  assert.ok(existsSync(join(cache, 'engine/0.1.0/server/entry.mjs')))
  assert.deepEqual(reg.hits, ['/@gitdocai%2Fengine/latest', '/engine.tgz'])

  const second = await devUntilReady(dir, env)
  assert.match(second.out, /Local:/, second.err)
  assert.equal(reg.hits.length, 2, 'the cached engine needs no network')
  reg.srv.close()
})

test('dev refuses an engine whose checksum does not match', async () => {
  const reg = await fakeRegistry({ corrupt: true })
  const dir = join(tmp(), 'docs')
  await run(['init', dir, '--name', 'X'])
  const cache = tmp()
  const r = await run(['dev', dir], { env: { GITDOC_CONFIG_DIR: tmp(), GITDOC_CACHE_DIR: cache, GITDOC_REGISTRY: reg.url } })
  assert.equal(r.code, 1)
  assert.match(r.err, /does not match its checksum/)
  assert.ok(!existsSync(join(cache, 'engine/0.1.0')))
  reg.srv.close()
})

test('dev refuses an engine of an unsupported major version', async () => {
  const reg = await fakeRegistry({ version: '3.0.0' })
  const dir = join(tmp(), 'docs')
  await run(['init', dir, '--name', 'X'])
  const r = await run(['dev', dir], { env: { GITDOC_CONFIG_DIR: tmp(), GITDOC_CACHE_DIR: tmp(), GITDOC_REGISTRY: reg.url } })
  assert.equal(r.code, 1)
  assert.match(r.err, /does not support/)
  reg.srv.close()
})

test('the package is gitdoc-cli and installs the gitdoc command', () => {
  const pkg = JSON.parse(readFileSync('package.json', 'utf-8'))
  assert.equal(pkg.name, 'gitdoc-cli')
  assert.deepEqual(Object.keys(pkg.bin), ['gitdoc'])
})
