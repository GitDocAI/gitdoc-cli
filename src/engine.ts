import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import { configDir } from './config'
import { UserError } from './util'

export interface Engine {
  /** The server's entry point: `engine/server/entry.mjs` of a gitdoc build. */
  server: string
  /** The renderer bundle the server spawns when publishing: `engine/render.js`. */
  renderCli: string
}

/**
 * `gitdoc dev` runs the rendering engine (the gitdoc server, a separate
 * program with its own license) as a child process. This finds it:
 *
 *   1. GITDOC_ENGINE_SERVER and GITDOC_ENGINE_RENDER
 *   2. `engine.json` in the config dir: { "server": "...", "renderCli": "..." }
 *
 * The CLI never bundles or imports the engine.
 */
export function loadEngine(): Engine {
  let e: Partial<Engine> = {}
  if (process.env.GITDOC_ENGINE_SERVER && process.env.GITDOC_ENGINE_RENDER) {
    e = { server: process.env.GITDOC_ENGINE_SERVER, renderCli: process.env.GITDOC_ENGINE_RENDER }
  } else {
    try {
      e = JSON.parse(readFileSync(join(configDir(), 'engine.json'), 'utf-8')) as Engine
    } catch {
      /* handled below */
    }
  }
  if (!e.server || !e.renderCli || !existsSync(e.server) || !existsSync(e.renderCli)) {
    throw new UserError(
      'the rendering engine was not found, and `gitdoc dev` needs it.\n' +
        '  Point the CLI at a build of the gitdoc server:\n' +
        '    export GITDOC_ENGINE_SERVER=/path/to/engine/server/entry.mjs\n' +
        '    export GITDOC_ENGINE_RENDER=/path/to/engine/render.js\n' +
        `  or save them as JSON in ${join(configDir(), 'engine.json')}.\n` +
        '  (Publishing to a server does not need the engine.)',
    )
  }
  return e as Engine
}
