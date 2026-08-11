/** @vitest-environment jsdom */
import { describe, expect, it, vi } from 'vitest'
import {
  buildPhotoPayload,
  isImageFile,
  isImageDataURL,
  parsePhotoPayload,
  dataURLToBlob,
  fileToDataURL,
  MAX_PHOTO_FILE_BYTES,
} from './media'

function dataURL(mime: string, bytes: Uint8Array) {
  let binary = ''
  const chunk = 0x8000
  for (let i = 0; i < bytes.length; i += chunk) {
    binary += String.fromCharCode(...bytes.slice(i, i + chunk))
  }
  return `data:${mime};base64,${btoa(binary)}`
}

function makeFile(name: string, type: string, bytes: number): File {
  const arr = new Uint8Array(bytes)
  return new File([arr], name, { type })
}

describe('parsePhotoPayload', () => {
  it('parses a valid photo payload', () => {
    const content = JSON.stringify({ v: 1, photo: 'data:image/png;base64,abc', caption: 'hi' })
    const parsed = parsePhotoPayload(content)
    expect(parsed?.photo).toBe('data:image/png;base64,abc')
    expect(parsed?.caption).toBe('hi')
  })

  it('returns null for non-JSON text', () => {
    expect(parsePhotoPayload('a sunny day')).toBeNull()
  })

  it('returns null for missing photo field', () => {
    expect(parsePhotoPayload(JSON.stringify({ caption: 'no photo' }))).toBeNull()
  })

  it('returns null for invalid JSON', () => {
    expect(parsePhotoPayload('{broken')).toBeNull()
  })

  it('normalizes caption to undefined when not a string', () => {
    const parsed = parsePhotoPayload(JSON.stringify({ photo: 'data:image/png;base64,x', caption: 123 }))
    expect(parsed?.caption).toBeUndefined()
  })
})

describe('buildPhotoPayload', () => {
  it('builds a versioned payload and trims fields', () => {
    const json = buildPhotoPayload('  data:image/png;base64,abc  ', '  caption  ')
    const parsed = JSON.parse(json) as { v: number; photo: string; caption: string }
    expect(parsed.v).toBe(1)
    expect(parsed.photo).toBe('data:image/png;base64,abc')
    expect(parsed.caption).toBe('caption')
  })
})

describe('isImageDataURL', () => {
  it('accepts image data URIs', () => {
    expect(isImageDataURL('data:image/png;base64,abc')).toBe(true)
    expect(isImageDataURL('data:image/jpeg;base64,/9j/4AAQ')).toBe(true)
  })

  it('rejects non-image and non-base64 URIs', () => {
    expect(isImageDataURL('data:text/plain;base64,aGk=')).toBe(false)
    expect(isImageDataURL('http://example.com/photo.png')).toBe(false)
    expect(isImageDataURL('data:image/png,rawbytes')).toBe(false)
    expect(isImageDataURL('')).toBe(false)
  })
})

describe('isImageFile', () => {
  it('accepts image/* files', () => {
    expect(isImageFile(makeFile('a.png', 'image/png', 4))).toBe(true)
  })

  it('rejects non-image files', () => {
    expect(isImageFile(makeFile('a.txt', 'text/plain', 4))).toBe(false)
  })
})

describe('dataURLToBlob', () => {
  it('decodes a base64 image data URL into a Blob', () => {
    const bytes = new Uint8Array([0x89, 0x50, 0x4e, 0x47])
    const url = dataURL('image/png', bytes)
    const blob = dataURLToBlob(url)
    expect(blob.type).toBe('image/png')
    expect(blob.size).toBe(4)
  })

  it('throws for a non-data URL', () => {
    expect(() => dataURLToBlob('not-a-data-url')).toThrow()
  })
})

describe('fileToDataURL', () => {
  it('resolves with a data URL for a valid image file', async () => {
    const file = makeFile('photo.png', 'image/png', 8)
    const url = await fileToDataURL(file)
    expect(url).toMatch(/^data:image\/png;base64,/)
  })

  it('rejects non-image files before reading', async () => {
    const file = makeFile('doc.txt', 'text/plain', 4)
    await expect(fileToDataURL(file)).rejects.toThrow(/only image/)
  })

  it('rejects files exceeding the size cap', async () => {
    const file = makeFile('big.png', 'image/png', MAX_PHOTO_FILE_BYTES + 1)
    await expect(fileToDataURL(file)).rejects.toThrow(/exceeds maximum size/)
  })

  it('does not call FileReader when validation fails', async () => {
    const spy = vi.spyOn(global, 'FileReader')
    const file = makeFile('doc.txt', 'text/plain', 4)
    await expect(fileToDataURL(file)).rejects.toThrow()
    expect(spy).not.toHaveBeenCalled()
  })
})
