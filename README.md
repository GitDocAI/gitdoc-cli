# GitDoc CLI

Write documentation as files, preview it live, and publish it to your own GitDoc server.

A documentation site is a folder: a `gitdocai.json` manifest, one `.mdx` file per page and an `assets/` directory. It lives in git, you review changes in pull requests, and an AI assistant can edit it like any other code. `gitdoc` is the command line tool around that folder.

```bash
gitdoc init my-docs --name "My Docs"   # create a site
cd my-docs
gitdoc dev                             # live preview, reloads on save
gitdoc connect https://docs.example.com --token <token>
gitdoc publish                         # package the folder and publish it
```

## Install

```bash
npm install -g @gitdocai/cli
gitdoc --help
```

Requires Node.js 20 or later. The npm package is `@gitdocai/cli`; the command is `gitdoc`. To try it without installing: `npx @gitdocai/cli init my-docs`.

> **Status: alpha.** Commands and options may still change.

## Commands

| Command | What it does |
|---|---|
| `gitdoc init [dir] [--name "My Docs"] [--hostname docs.acme.com]` | Create a site: manifest, a first page, `AGENTS.md` and `CLAUDE.md` for AI assistants |
| `gitdoc dev [dir] [--port 3000]` | Live preview with reload on every save, including search. If the port is taken it tries the next ones |
| `gitdoc connect <url> [--token <token>] [--dir <docs>]` | Save a server's credentials and remember the server for this folder. The token can also come from `GITDOC_TOKEN` or stdin |
| `gitdoc publish [dir] [--server <url>] [--hostname <host>] [--dry-run]` | Package the folder and publish it. `--dry-run` shows what would be sent and sends nothing |
| `gitdoc sites [--server <url>]` | List the documentations a server hosts |
| `gitdoc setup [--print] [--uninstall]` | Install the skill that teaches Claude Code the format and the commands. `--print` shows it, `--uninstall` removes it |

### What gets published

`gitdocai.json`, every `.md` / `.mdx` page and `assets/`. Dotfiles, `node_modules` and the root-level `AGENTS.md`, `CLAUDE.md` and `README.md` are not part of the site.

The server renders the pages and indexes them for search as part of the publish, and the CLI tells you how it went:

```
Published (ok): 7 pages in 113 ms
  site: http://acme-docs.localhost:8080
  search: indexed 7 pages (opensearch, with vectors)
```

If the search backend is down the site still goes live, and you are told to publish again once it is back.

### Which site is which

A server hosts any number of documentations. Each one answers on its own **hostname**: `<name>.localhost` on a local server, your domain in production (`gitdoc publish --hostname docs.acme.com`). After the first publish the folder remembers it in `.gitdoc/project.json`, so later publishes always go to the same site. That file holds no secret and can be committed.

### Pages

A page is an `.mdx` file with a `title` in its front matter. The site shows that title as the page heading, so do not repeat it as a `# Title` in the body.

```mdx
---
title: Installation
---

Run `npm install` to get started.
```

## Credentials

The token is stored in your user config directory (`~/.config/gitdoc/credentials.json`, owner-only) and never inside the project. In CI use `GITDOC_TOKEN` and `GITDOC_SERVER`.

## Working with an AI assistant

```bash
gitdoc setup
```

installs a skill for Claude Code. From then on you can say "I want to write documentation for this project" and the assistant knows the file format, the components, how to preview and when to ask before publishing. `gitdoc init` also writes an `AGENTS.md` in each site for other assistants (and a `CLAUDE.md` that imports it).

## The server

`gitdoc` talks to a GitDoc server over HTTP (`POST /api/publish`, `GET /api/sites`). The server is a separate program with its own repository and license ([AGPL-3.0](https://github.com/GitDocAI/GitDoc)); this CLI never bundles or imports it.

To run one:

```bash
export PUBLISH_TOKEN=$(openssl rand -hex 24)   # keep it: the CLI and the admin panel ask for it
docker run -d -p 8080:8080 -p 127.0.0.1:8081:8081 \
  -e PUBLISH_TOKEN -v gitdoc-data:/data ghcr.io/gitdocai/gitdoc-server
```

Its admin panel (`http://127.0.0.1:8081`) shows the exact commands to connect this CLI. See the [GitDoc README](https://github.com/GitDocAI/GitDoc#readme) for search, the AI assistant, private sites and everything else the server does.

### Preview

`gitdoc dev` runs the same engine locally as a child process, so what you preview is what gets served. The first time, the CLI downloads the prebuilt engine, [`@gitdocai/engine`](https://www.npmjs.com/package/@gitdocai/engine) (about 4 MB), from the npm registry. It checks the checksum, unpacks it into `~/.cache/gitdoc/engine/<version>` (using the system `tar`) and reuses it from then on without touching the network. Delete that folder to fetch a newer one. An engine of a different major version than the CLI expects is refused, with a message telling you to update.

The preview inherits your environment: set `AI_URL` and `AI_EMBEDDING_MODEL` / `AI_CHAT_MODEL` before `gitdoc dev` to try search by meaning and the assistant locally. Publishing to a server does not need the engine at all.

To use your own build of the engine instead (`npm run build` in the GitDoc repository produces an `engine/` directory), point the CLI at it. This takes precedence over the cache:

```bash
export GITDOC_ENGINE_SERVER=/path/to/GitDoc/engine/server/entry.mjs
export GITDOC_ENGINE_RENDER=/path/to/GitDoc/engine/render.js
```

or save the same two paths as JSON (`{"server": "...", "renderCli": "..."}`) in `~/.config/gitdoc/engine.json`.

## Environment

| Variable | Meaning |
|---|---|
| `GITDOC_TOKEN` | Token to use instead of the saved one (CI) |
| `GITDOC_SERVER` | Server to use when none is configured for the folder |
| `GITDOC_CONFIG_DIR` | Where credentials are kept (default `~/.config/gitdoc`, or `$XDG_CONFIG_HOME/gitdoc`) |
| `GITDOC_CACHE_DIR` | Where the downloaded engine is cached (default `~/.cache/gitdoc`) |
| `GITDOC_ENGINE_SERVER`, `GITDOC_ENGINE_RENDER` | Use this engine build instead of downloading one |
| `GITDOC_NO_DOWNLOAD` | `1` forbids downloading the engine (air-gapped machines) |
| `GITDOC_REGISTRY` | npm registry to fetch the engine from (default `https://registry.npmjs.org`) |
| `GITDOC_ENGINE_PACKAGE` | Engine package name (default `@gitdocai/engine`) |

## Development

```bash
npm install
npm test        # builds, then runs the contract tests against a fake server
npm run build   # dist/gitdoc.js
npm run typecheck
```

The tests need no server and no network: a fake server records what the CLI sends, and a fake registry serves an engine package.

## Releasing

Publish a GitHub Release (tag `v<version>`). The release workflow is the only pipeline: it typechecks and tests on Node 20 and 22, then publishes `@gitdocai/cli` to npm with the version of the tag. A release marked as pre-release goes to the `next` dist-tag instead of `latest`. It needs the `NPM_TOKEN` secret.

## License

[Apache-2.0](LICENSE).
