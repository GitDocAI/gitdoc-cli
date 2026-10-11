# gitdoc

Write documentation as files, preview it live, publish it to a GitDocAI server.

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
npm install -g gitdoc-cli
gitdoc --help
```

Requires Node.js 20 or later. The npm package is `gitdoc-cli`; the command is `gitdoc`. To try it without installing: `npx gitdoc-cli init my-docs`.

> **Status: alpha.** The package has not been published yet; until it is, build it from source (below).

## Commands

| Command | What it does |
|---|---|
| `gitdoc init [dir]` | Create a site: manifest, a first page, and an `AGENTS.md` for AI assistants |
| `gitdoc dev [dir]` | Live preview with reload on every save |
| `gitdoc connect <url>` | Save a server's credentials and remember it for this folder |
| `gitdoc publish [dir]` | Package the folder and publish it (`--dry-run` to see what would be sent) |
| `gitdoc sites` | List the documentations a server hosts |
| `gitdoc setup` | Install the skill that teaches Claude Code the format and the commands |

### Which site is which

A server hosts any number of documentations. Each one answers on its own **hostname**: `<name>.localhost` on a local server, your domain in production (`gitdoc publish --hostname docs.acme.com`). After the first publish the folder remembers it in `.gitdoc/project.json`, so later publishes always go to the same site.

### Credentials

The token is stored in your user config directory (`~/.config/gitdoc/credentials.json`, owner-only) and never inside the project. In CI use `GITDOC_TOKEN`.

## Working with an AI assistant

```bash
gitdoc setup
```

installs a skill for Claude Code. From then on you can say "I want to write documentation for this project" and the assistant knows the file format, the components, how to preview and when to ask before publishing. `gitdoc init` also writes an `AGENTS.md` in each site for other assistants.

## The server

`gitdoc` talks to a GitDocAI server over HTTP (`POST /api/publish`, `GET /api/sites`). The server is a separate program with its own repository and license: [GitDocAI/GitDoc](https://github.com/GitDocAI/GitDoc). This CLI never bundles or imports it.

`gitdoc dev` runs that server locally as a child process. Build it once (`npm run build` in that repository produces an `engine/` directory) and point the CLI at it:

```bash
export GITDOC_ENGINE_SERVER=/path/to/GitDoc/engine/server/entry.mjs
export GITDOC_ENGINE_RENDER=/path/to/GitDoc/engine/render.js
```

or save the same two paths as JSON (`{"server": "...", "renderCli": "..."}`) in `~/.config/gitdoc/engine.json`. The engine is self-contained: it can be moved to any directory. Publishing to a server does not need it.

## Development

```bash
npm install
npm test        # builds, then runs the contract tests against a fake server
npm run build   # dist/gitdoc.js
```

## License

[Apache-2.0](LICENSE).
