import { existsSync, mkdirSync, writeFileSync } from 'node:fs'
import { basename, join, relative, resolve } from 'node:path'
import { writeProject } from './config'
import { UserError, slugify } from './util'

const AGENTS_MD = `# Documentation site (GitDocAI)

This folder is a GitDocAI documentation site: \`gitdocai.json\` (manifest + navigation),
one \`.mdx\` file per page, and \`assets/\` for images.

Commands:
- \`gitdoc dev\` — live preview with auto-reload (run it in the background).
- \`gitdoc publish --dry-run\` — show what would be published and where.
- \`gitdoc publish\` — publish. **Ask the user to confirm first.**
- \`gitdoc sites\` — list the documentations on the connected server.

Rules:
- Every page must be listed in \`gitdocai.json\` navigation (path without extension), or it won't appear.
- No ids in files; never invent UUIDs. Keep \`x-site-key\` unchanged.
- Don't add a \`# Title\` heading to a page: the site prints the page's \`title\` above the content.
  Start with a short intro and use \`##\` for sections.
- Components go in block form: the tags on their own lines, content indented between them.
  A one-line \`<Note>text</Note>\` loses the component.
- Never ask for, print or store tokens or passwords. Connecting to a server is done by the user.
- If the \`gitdoc\` skill is available, follow it: it has the full format reference.
`

const INTRO = (name: string) => `---
title: Introduction
description: Welcome to ${name}
---

Welcome to **${name}**. Replace this page with your own content.

<Note>
  Run \`gitdoc dev\` to see this site locally. It reloads every time you save a file.
</Note>

## Next steps

- Add pages under \`v1/\` and list them in \`gitdocai.json\`.
- Put images in \`assets/\`.
`

export function init(pos: string[], flags: Record<string, string | boolean>): void {
  const dir = resolve(pos[0] ?? '.')
  if (existsSync(join(dir, 'gitdocai.json')) && !flags.force) {
    throw new UserError(`${dir} already has a gitdocai.json. Nothing was changed.`)
  }
  const name =
    (typeof flags.name === 'string' && flags.name.trim()) ||
    basename(dir)
      .replace(/[-_]+/g, ' ')
      .replace(/\b\w/g, (c) => c.toUpperCase()) ||
    'Documentation'
  const siteKey = slugify(name) || 'docs'

  const manifest = {
    format: 1,
    name,
    description: '',
    'x-site-key': siteKey,
    theme: { preset: 'mist', default: 'dark', modes: ['light', 'dark'] },
    versions: [
      {
        name: 'v1',
        slug: 'v1',
        default: true,
        latest: true,
        navigation: [
          {
            section: 'Documentation',
            icon: 'book-open',
            pages: ['v1/getting-started/introduction'],
          },
        ],
      },
    ],
  }

  mkdirSync(join(dir, 'v1', 'getting-started'), { recursive: true })
  mkdirSync(join(dir, 'assets'), { recursive: true })
  writeFileSync(join(dir, 'gitdocai.json'), JSON.stringify(manifest, null, 2) + '\n')
  const intro = join(dir, 'v1', 'getting-started', 'introduction.mdx')
  if (!existsSync(intro)) writeFileSync(intro, INTRO(name))

  const written = ['gitdocai.json', 'v1/getting-started/introduction.mdx', 'assets/']
  const agents = join(dir, 'AGENTS.md')
  if (!existsSync(agents)) {
    writeFileSync(agents, AGENTS_MD)
    written.push('AGENTS.md')
  }
  const claude = join(dir, 'CLAUDE.md')
  if (!existsSync(claude)) {
    writeFileSync(claude, '@AGENTS.md\n')
    written.push('CLAUDE.md')
  }
  if (typeof flags.hostname === 'string' && flags.hostname) {
    writeProject(dir, { hostname: flags.hostname })
    written.push('.gitdoc/project.json')
  }

  const rel = relative(process.cwd(), dir) || '.'
  console.log(`Created "${name}" in ${dir}`)
  for (const w of written) console.log(`  + ${w}`)
  console.log(`\nNext:\n  cd ${rel}\n  gitdoc dev        # live preview\n  gitdoc publish    # when it's ready (needs: gitdoc connect <server> --token …)`)
}
