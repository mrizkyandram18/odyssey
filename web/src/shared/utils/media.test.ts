/** @vitest-environment jsdom */
import { describe, expect, it, vi, afterEach } from 'vitest'
import {
  buildPhotoPayload,
  isImageFile,
  isImageDataURL,
  parsePhotoPayload,
  dataURLToBlob,
  fileToDataURL,
  MAX_PHOTO_FILE_BYTES,
  buildVideoPayload,
  isVideoFile,
  isVideoDataURL,
  isVideoContainerSignature,
  parseVideoPayload,
  fileToVideoDataURL,
  MAX_VIDEO_FILE_BYTES,
} from './media'

afterEach(() => {
  vi.restoreAllMocks()
})

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

const minMP4 = new Uint8Array([0x00, 0x00, 0x00, 0x14, 0x66, 0x74, 0x79, 0x70, 0x69, 0x73, 0x6f, 0x6d, 0x00, 0x00, 0x00, 0x00, 0x69, 0x73, 0x6f, 0x6d])
const minWebM = new Uint8Array([0x1a, 0x45, 0xdf, 0xa3, 0x93, 0x42, 0x82, 0x88])
const minOgg = new Uint8Array([0x4f, 0x67, 0x67, 0x53, 0x00, 0x00, 0x00, 0x00])

describe('parseVideoPayload', () => {
  it('parses a valid video payload', () => {
    const content = JSON.stringify({ v: 1, video: 'data:video/mp4;base64,abc', caption: 'hi' })
    const parsed = parseVideoPayload(content)
    expect(parsed?.video).toBe('data:video/mp4;base64,abc')
    expect(parsed?.caption).toBe('hi')
  })

  it('returns null for non-JSON text', () => {
    expect(parseVideoPayload('a sunny clip')).toBeNull()
  })

  it('returns null for missing video field', () => {
    expect(parseVideoPayload(JSON.stringify({ caption: 'no video' }))).toBeNull()
  })

  it('returns null for invalid JSON', () => {
    expect(parseVideoPayload('{broken')).toBeNull()
  })

  it('normalizes caption to undefined when not a string', () => {
    const parsed = parseVideoPayload(JSON.stringify({ video: 'data:video/mp4;base64,x', caption: 123 }))
    expect(parsed?.caption).toBeUndefined()
  })
})

describe('buildVideoPayload', () => {
  it('builds a versioned payload and trims fields', () => {
    const json = buildVideoPayload('  data:video/mp4;base64,abc  ', '  caption  ')
    const parsed = JSON.parse(json) as { v: number; video: string; caption: string }
    expect(parsed.v).toBe(1)
    expect(parsed.video).toBe('data:video/mp4;base64,abc')
    expect(parsed.caption).toBe('caption')
  })
})

describe('isVideoDataURL', () => {
  it('accepts video data URIs', () => {
    expect(isVideoDataURL('data:video/mp4;base64,abc')).toBe(true)
    expect(isVideoDataURL('data:video/webm;base64,/9j/4AAQ')).toBe(true)
  })

  it('rejects non-video and non-base64 URIs', () => {
    expect(isVideoDataURL('data:image/png;base64,abc')).toBe(false)
    expect(isVideoDataURL('data:text/plain;base64,aGk=')).toBe(false)
    expect(isVideoDataURL('https://example.com/clip.mp4')).toBe(false)
    expect(isVideoDataURL('data:video/mp4,rawbytes')).toBe(false)
    expect(isVideoDataURL('')).toBe(false)
  })
})

describe('isVideoFile', () => {
  it('accepts video/* files', () => {
    expect(isVideoFile(makeFile('a.mp4', 'video/mp4', 4))).toBe(true)
    expect(isVideoFile(makeFile('a.webm', 'video/webm', 4))).toBe(true)
  })

  it('rejects non-video files', () => {
    expect(isVideoFile(makeFile('a.txt', 'text/plain', 4))).toBe(false)
    expect(isVideoFile(makeFile('a.png', 'image/png', 4))).toBe(false)
  })
})

describe('isVideoContainerSignature', () => {
  it('accepts valid mp4/webm/ogg signatures', () => {
    expect(isVideoContainerSignature('video/mp4', minMP4)).toBe(true)
    expect(isVideoContainerSignature('video/webm', minWebM)).toBe(true)
    expect(isVideoContainerSignature('video/ogg', minOgg)).toBe(true)
  })

  it('rejects mismatched container for a given mime', () => {
    expect(isVideoContainerSignature('video/mp4', minWebM)).toBe(false)
    expect(isVideoContainerSignature('video/webm', minMP4)).toBe(false)
    expect(isVideoContainerSignature('video/quicktime', minMP4)).toBe(false)
  })

  it('rejects too-short buffers', () => {
    expect(isVideoContainerSignature('video/mp4', new Uint8Array([0x1a, 0x45]))).toBe(false)
  })
})

describe('fileToVideoDataURL', () => {
  it('resolves with a data URL for a valid video file', async () => {
    const file = makeFile('clip.mp4', 'video/mp4', 8)
    const url = await fileToVideoDataURL(file)
    expect(url).toMatch(/^data:video\/mp4;base64,/)
  })

  it('rejects non-video files before reading', async () => {
    const file = makeFile('doc.txt', 'text/plain', 4)
    await expect(fileToVideoDataURL(file)).rejects.toThrow(/only video/)
  })

  it('rejects files exceeding the size cap', async () => {
    const file = makeFile('big.mp4', 'video/mp4', MAX_VIDEO_FILE_BYTES + 1)
    await expect(fileToVideoDataURL(file)).rejects.toThrow(/exceeds maximum size/)
  })

  it('does not call FileReader when validation fails', async () => {
    const spy = vi.spyOn(global, 'FileReader')
    const file = makeFile('doc.txt', 'text/plain', 4)
    await expect(fileToVideoDataURL(file)).rejects.toThrow()
    expect(spy).not.toHaveBeenCalled()
  })
})
