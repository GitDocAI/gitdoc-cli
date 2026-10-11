import { existsSync, mkdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import { homedir } from 'node:os'
import { dirname, join } from 'node:path'
import skill from '../skill/SKILL.md'

function claudeDir(): string {
  return process.env.CLAUDE_CONFIG_DIR || join(homedir(), '.claude')
}

export function setup(flags: Record<string, string | boolean>): void {
  const file = join(claudeDir(), 'skills', 'gitdoc', 'SKILL.md')
  if (flags.print) {
    process.stdout.write(skill)
    return
  }
  if (flags.uninstall) {
    rmSync(dirname(file), { recursive: true, force: true })
    console.log(`Removed ${dirname(file)}`)
    return
  }
  // Earlier versions installed the skill as `gitdocai`: replace it, don't duplicate it.
  const legacy = join(claudeDir(), 'skills', 'gitdocai', 'SKILL.md')
  let migrated = false
  try {
    if (/^name:\s*gitdocai\s*$/m.test(readFileSync(legacy, 'utf-8'))) {
      rmSync(dirname(legacy), { recursive: true, force: true })
      migrated = true
    }
  } catch {
    /* no previous install */
  }

  const had = existsSync(file)
  const same = had && readFileSync(file, 'utf-8') === skill
  mkdirSync(dirname(file), { recursive: true })
  writeFileSync(file, skill)
  console.log(same ? 'Claude Code skill already up to date:' : had ? 'Claude Code skill updated:' : 'Claude Code skill installed:')
  console.log(`  ${file}`)
  if (migrated) console.log('  (replaced the earlier "gitdocai" skill)')
  console.log('\nOpen a new Claude Code session and ask it to create a documentation — it will know the format and the gitdoc commands.')
  console.log('Other agents: `gitdoc init` writes an AGENTS.md into each docs folder.')
}
