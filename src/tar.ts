import { gzipSync } from 'node:zlib'

/**
 * Minimal ustar writer: regular files only, which is all publish needs.
 * Pure JS so the CLI doesn't depend on a system `tar` (Windows, minimal
 * containers). Names over 100 bytes use the ustar `prefix` field.
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

function header(name: string, size: number, mtime: number): Buffer {
  const h = Buffer.alloc(512)
  let prefix = ''
  let base = name
  if (Buffer.byteLength(name) > 100) {
    const cut = name.lastIndexOf('/', 155 + 1)
    if (cut <= 0 || Buffer.byteLength(name.slice(0, cut)) > 155 || Buffer.byteLength(name.slice(cut + 1)) > 100) {
      throw new Error(`path too long for the archive: ${name}`)
    }
    prefix = name.slice(0, cut)
    base = name.slice(cut + 1)
  }
  h.write(base, 0, 100, 'utf-8')
  h.write(octal(0o644, 8), 100)
  h.write(octal(0, 8), 108)
  h.write(octal(0, 8), 116)
  h.write(octal(size, 12), 124)
  h.write(octal(mtime, 12), 136)
  h.write('        ', 148) // checksum placeholder: 8 spaces
  h.write('0', 156)
  h.write('ustar\0', 257)
  h.write('00', 263)
  h.write(prefix, 345, 155, 'utf-8')
  let sum = 0
  for (const b of h) sum += b
  h.write(sum.toString(8).padStart(6, '0') + '\0 ', 148)
  return h
}

export function tarGz(files: TarFile[]): Buffer {
  const parts: Buffer[] = []
  for (const f of files) {
    parts.push(header(f.name, f.data.length, Math.floor((f.mtime ?? Date.now()) / 1000)))
    parts.push(f.data)
    const pad = (512 - (f.data.length % 512)) % 512
    if (pad) parts.push(Buffer.alloc(pad))
  }
  parts.push(Buffer.alloc(1024)) // end-of-archive
  return gzipSync(Buffer.concat(parts), { level: 6 })
}
