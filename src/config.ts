import { chmodSync, existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs'
import { homedir } from 'node:os'
import { dirname, join } from 'node:path'

/** Secrets live in the user's config dir, never in the project. */
export function configDir(): string {
  if (process.env.GITDOC_CONFIG_DIR) return process.env.GITDOC_CONFIG_DIR
  const base = process.env.XDG_CONFIG_HOME || join(homedir(), '.config')
  return join(base, 'gitdoc')
}

/** Where earlier versions kept it (`~/.config/gitdocai`); still read, never written. */
function legacyConfigDir(): string | null {
  if (process.env.GITDOC_CONFIG_DIR) return null
  const base = process.env.XDG_CONFIG_HOME || join(homedir(), '.config')
  return join(base, 'gitdocai')
}

const credsFile = () => join(configDir(), 'credentials.json')

type Creds = { servers: Record<string, { token: string }> }

function readCredsFile(file: string): Creds {
  try {
    return JSON.parse(readFileSync(file, 'utf-8')) as Creds
  } catch {
    return { servers: {} }
  }
}

export function normalizeServer(url: string): string {
  const u = new URL(/^https?:\/\//.test(url) ? url : `https://${url}`)
  return u.origin
}

/** Current credentials, plus any saved by an earlier version (the current ones win). */
function readCreds(): Creds {
  const legacy = legacyConfigDir()
  const old = legacy ? readCredsFile(join(legacy, 'credentials.json')) : { servers: {} }
  const now = readCredsFile(credsFile())
  return { servers: { ...old.servers, ...now.servers } }
}

export function saveToken(server: string, token: string): string {
  const c = readCreds()
  c.servers[server] = { token }
  const file = credsFile()
  mkdirSync(dirname(file), { recursive: true, mode: 0o700 })
  writeFileSync(file, JSON.stringify(c, null, 2) + '\n', { mode: 0o600 })
  chmodSync(file, 0o600)
  return file
}

export function tokenFor(server: string): string | undefined {
  return process.env.GITDOC_TOKEN || readCreds().servers[server]?.token
}

export function connectedServers(): string[] {
  return Object.keys(readCreds().servers)
}

/** Non-secret, committed with the docs: where this folder publishes to. */
export interface Project {
  server?: string
  hostname?: string
}

export function readProject(dir: string): Project {
  const f = join(dir, '.gitdoc', 'project.json')
  if (!existsSync(f)) return {}
  try {
    return JSON.parse(readFileSync(f, 'utf-8')) as Project
  } catch {
    return {}
  }
}

export function writeProject(dir: string, p: Project): string {
  const f = join(dir, '.gitdoc', 'project.json')
  mkdirSync(dirname(f), { recursive: true })
  writeFileSync(f, JSON.stringify({ ...readProject(dir), ...p }, null, 2) + '\n')
  return f
}
