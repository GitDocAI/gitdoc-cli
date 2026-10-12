import { gzipSync } from 'node:zlib'

/**
 * Minimal ustar writer: regular files only, which is all publish needs.
 * Pure JS so the CLI doesn't depend on a system `tar` (Windows, minimal
 * containers). A path that fits ustar's name (100) + prefix (155) fields is
 * stored as such; anything longer gets a PAX extended header (`path=`), which
 * GNU tar and bsdtar both read.
 */
export interface TarFile {
  /** Path inside the archive, forward slashes, no leading slash. */
  name: string
  data: Buffer
  mtime?: number
}

function octal(n: number, len: number): string {
  return n.toString(8).padStart(len - 1, '0') + '\0'
}

/** Split a path into ustar `prefix` + `name`, or null if it does not fit. */
function splitUstar(name: string): { prefix: string; base: string } | null {
  if (Buffer.byteLength(name) <= 100) return { prefix: '', base: name }
  for (let cut = name.lastIndexOf('/'); cut > 0; cut = name.lastIndexOf('/', cut - 1)) {
    const prefix = name.slice(0, cut)
    const base = name.slice(cut + 1)
    if (Buffer.byteLength(prefix) <= 155 && Buffer.byteLength(base) <= 100) return { prefix, base }
  }
  return null
}

function header(name: string, size: number, mtime: number, type = '0', prefix = ''): Buffer {
  const h = Buffer.alloc(512)
  h.write(name, 0, 100, 'utf-8')
  h.write(octal(0o644, 8), 100)
  h.write(octal(0, 8), 108)
  h.write(octal(0, 8), 116)
  h.write(octal(size, 12), 124)
  h.write(octal(mtime, 12), 136)
  h.write('        ', 148) // checksum placeholder: 8 spaces
  h.write(type, 156)
  h.write('ustar\0', 257)
  h.write('00', 263)
  h.write(prefix, 345, 155, 'utf-8')
  let sum = 0
  for (const b of h) sum += b
  h.write(sum.toString(8).padStart(6, '0') + '\0 ', 148)
  return h
}

/** One PAX record: "<length> <key>=<value>\n", where length counts itself. */
function paxRecord(key: string, value: string): Buffer {
  const body = Buffer.from(` ${key}=${value}\n`, 'utf-8')
  let len = body.length + 1
  while (String(len).length + body.length !== len) len = String(len).length + body.length
  return Buffer.concat([Buffer.from(String(len)), body])
}

function entry(f: TarFile): Buffer[] {
  const mtime = Math.floor((f.mtime ?? Date.now()) / 1000)
  const parts: Buffer[] = []
  let name = f.name
  let prefix = ''
  const split = splitUstar(f.name)
  if (split) {
    name = split.base
    prefix = split.prefix
  } else {
    // Too long for ustar: carry the real path in a PAX extended header and
    // leave a truncated, unique placeholder in the regular header.
    const pax = paxRecord('path', f.name)
    parts.push(header('PaxHeader', pax.length, mtime, 'x'), pax)
    const pad = (512 - (pax.length % 512)) % 512
    if (pad) parts.push(Buffer.alloc(pad))
    name = Buffer.from(f.name).subarray(0, 100).toString('utf-8').replace(/\uFFFD/g, '_')
  }
  parts.push(header(name, f.data.length, mtime, '0', prefix), f.data)
  const pad = (512 - (f.data.length % 512)) % 512
  if (pad) parts.push(Buffer.alloc(pad))
  return parts
}

export function tarGz(files: TarFile[]): Buffer {
  const parts: Buffer[] = []
  for (const f of files) parts.push(...entry(f))
  parts.push(Buffer.alloc(1024)) // end-of-archive
  return gzipSync(Buffer.concat(parts), { level: 6 })
}
