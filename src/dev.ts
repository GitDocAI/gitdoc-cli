import { spawn, type ChildProcess } from 'node:child_process'
import { randomBytes } from 'node:crypto'
import { existsSync, mkdtempSync, rmSync, watch } from 'node:fs'
import http from 'node:http'
import net from 'node:net'
import { tmpdir } from 'node:os'
import { join, resolve } from 'node:path'
import { loadEngine } from './engine'
import { tarGz } from './tar'
import { UserError, collect } from './util'

const CLIENT = `(function(){if(window.__gdDev)return;window.__gdDev=1;
var bar=null;function show(t){if(!bar){bar=document.createElement('div');
bar.style.cssText='position:fixed;z-index:2147483647;left:0;right:0;top:0;padding:10px 16px;background:#7f1d1d;color:#fff;font:13px/1.45 ui-monospace,monospace;white-space:pre-wrap';
document.documentElement.appendChild(bar)}bar.textContent=t}
var es=new EventSource('/__gitdoc/events');
es.onmessage=function(e){var m=JSON.parse(e.data);
if(m.type==='reload')location.reload();
else if(m.type==='problems')show('gitdoc: '+m.problems.join('\\n'));
else if(m.type==='ok'&&bar){bar.remove();bar=null}}})();`

async function freePort(): Promise<number> {
  return new Promise((res, rej) => {
    const s = net.createServer()
    s.once('error', rej)
    s.listen(0, '127.0.0.1', () => {
      const p = (s.address() as net.AddressInfo).port
      s.close(() => res(p))
    })
  })
}

async function waitFor(port: number, ms = 20_000): Promise<void> {
  const end = Date.now() + ms
  while (Date.now() < end) {
    const ok = await new Promise<boolean>((r) => {
      const q = http.get({ host: '127.0.0.1', port, path: '/login', timeout: 800 }, (res) => {
        res.resume()
        r(true)
      })
      q.on('error', () => r(false))
      q.on('timeout', () => {
        q.destroy()
        r(false)
      })
    })
    if (ok) return
    await new Promise((r) => setTimeout(r, 150))
  }
  throw new UserError('the preview server did not start in time')
}

export async function dev(pos: string[], flags: Record<string, string | boolean>): Promise<void> {
  const dir = resolve(pos[0] ?? '.')
  if (!existsSync(join(dir, 'gitdocai.json'))) {
    throw new UserError(`no gitdocai.json in ${dir}. Create one with: gitdoc init ${pos[0] ?? '.'}`)
  }
  const engine = loadEngine()
  const wanted = Number(typeof flags.port === 'string' ? flags.port : 3000)
  const innerPort = await freePort()
  const token = randomBytes(16).toString('hex')
  const dataDir = mkdtempSync(join(tmpdir(), 'gitdoc-dev-'))

  const server: ChildProcess = spawn(process.execPath, [engine.server], {
    stdio: ['ignore', 'ignore', 'pipe'],
    env: {
      ...process.env,
      NODE_ENV: 'production',
      HOST: '127.0.0.1',
      PORT: String(innerPort),
      DATA_DIR: dataDir,
      SINGLE_SITE: 'true',
      PUBLISH_TOKEN: token,
      PUBLISH_HOSTNAME: 'localhost',
      PUBLISH_MAX_MB: '500',
      RENDER_CLI_PATH: engine.renderCli,
    },
  })
  let serverErr = ''
  server.stderr?.on('data', (d) => (serverErr += d))

  const cleanup = () => {
    try {
      server.kill()
    } catch {
      /* already gone */
    }
    rmSync(dataDir, { recursive: true, force: true })
  }
  process.on('SIGINT', () => (cleanup(), process.exit(0)))
  process.on('SIGTERM', () => (cleanup(), process.exit(0)))
  server.on('exit', (code) => {
    if (code) {
      console.error(`preview server stopped (${code})\n${serverErr.slice(-600)}`)
      cleanup()
      process.exit(1)
    }
  })

  try {
    await waitFor(innerPort)
  } catch (e) {
    cleanup()
    throw serverErr ? new UserError(`${(e as Error).message}\n${serverErr.slice(-600)}`) : e
  }

  const clients = new Set<http.ServerResponse>()
  const send = (msg: unknown) => {
    for (const c of clients) c.write(`data: ${JSON.stringify(msg)}\n\n`)
  }

  /** Publish the folder to the preview server; returns the problems found. */
  const publish = async (): Promise<{ pages: number; ms: number; problems: string[] }> => {
    const body = tarGz(collect(dir))
    const res = await fetch(`http://127.0.0.1:${innerPort}/api/publish?hostname=localhost`, {
      method: 'POST',
      headers: { authorization: `Bearer ${token}`, 'content-type': 'application/gzip' },
      body,
    })
    const data = (await res.json().catch(() => ({}))) as any
    if (!res.ok) return { pages: 0, ms: 0, problems: [data.error ?? `publish failed (${res.status})`] }
    const errs = ((data.result?.errors ?? []) as { where: string; message: string }[]).filter(
      (e) => e.where !== 'r2-cleanup',
    )
    return { pages: data.result?.rendered_pages ?? 0, ms: data.result?.duration_ms ?? 0, problems: errs.map((e) => e.message) }
  }

  let first
  try {
    first = await publish()
  } catch (e) {
    cleanup()
    throw e
  }

  // The proxy: injects the reload client into HTML and serves the SSE channel.
  const proxy = http.createServer((req, res) => {
    if (req.url === '/__gitdoc/events') {
      res.writeHead(200, { 'content-type': 'text/event-stream', 'cache-control': 'no-store', connection: 'keep-alive' })
      res.write('\n')
      clients.add(res)
      req.on('close', () => clients.delete(res))
      return
    }
    const up = http.request(
      {
        host: '127.0.0.1',
        port: innerPort,
        method: req.method,
        path: req.url,
        headers: { ...req.headers, host: 'localhost', 'accept-encoding': 'identity' },
      },
      (r) => {
        const type = String(r.headers['content-type'] ?? '')
        if (!type.includes('text/html')) {
          res.writeHead(r.statusCode ?? 502, r.headers)
          r.pipe(res)
          return
        }
        const chunks: Buffer[] = []
        r.on('data', (c) => chunks.push(c))
        r.on('end', () => {
          let html = Buffer.concat(chunks).toString('utf-8')
          const tag = `<script>${CLIENT}</script>`
          html = html.includes('</body>') ? html.replace('</body>', `${tag}</body>`) : html + tag
          const headers = { ...r.headers }
          delete headers['content-length']
          delete headers['content-encoding']
          res.writeHead(r.statusCode ?? 200, { ...headers, 'cache-control': 'no-store' })
          res.end(html)
        })
      },
    )
    up.on('error', () => {
      res.writeHead(502).end('preview server unavailable')
    })
    req.pipe(up)
  })

  const port = await new Promise<number>((ok, fail) => {
    const listen = (p: number) => {
      proxy.once('error', (e: NodeJS.ErrnoException) => {
        if (e.code === 'EADDRINUSE' && p < wanted + 20) listen(p + 1)
        else fail(e)
      })
      proxy.listen(p, '127.0.0.1', () => ok(p))
    }
    listen(wanted)
  })

  const report = (r: { pages: number; ms: number; problems: string[] }) => {
    if (r.problems.length) {
      console.log(`  ${r.problems.length} problem(s):`)
      for (const p of r.problems.slice(0, 15)) console.log(`   - ${p}`)
      send({ type: 'problems', problems: r.problems.slice(0, 8) })
    } else {
      send({ type: 'ok' })
    }
  }

  console.log(`\n  gitdoc dev — ${dir}`)
  console.log(`  Local:   http://localhost:${port}`)
  console.log(`  ${first.pages} pages rendered in ${first.ms} ms · reloads on save · Ctrl+C to stop\n`)
  report(first)

  let timer: NodeJS.Timeout | undefined
  let running = false
  let again = false
  const sync = async () => {
    if (running) return void (again = true)
    running = true
    try {
      do {
        again = false
        const r = await publish()
        console.log(`  republished: ${r.pages} pages in ${r.ms} ms`)
        report(r)
        send({ type: 'reload' })
      } while (again)
    } catch (e) {
      console.log(`  publish failed: ${(e as Error).message}`)
      send({ type: 'problems', problems: [(e as Error).message] })
    } finally {
      running = false
    }
  }
  watch(dir, { recursive: true }, (_e, name) => {
    if (!name) return
    if (/(^|[\\/])(\.[^\\/]+|node_modules)([\\/]|$)/.test(name)) return
    console.log(`  changed: ${name}`)
    clearTimeout(timer)
    timer = setTimeout(sync, 200)
  })

  await new Promise(() => {}) // until Ctrl+C
}
