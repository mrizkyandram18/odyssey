// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from 'vitest'
import { uploadTaskProof } from './compress'

describe('uploadTaskProof', () => {
  afterEach(() => {
    vi.restoreAllMocks()
    vi.unstubAllGlobals()
  })

  it('translates Vercel FUNCTION_PAYLOAD_TOO_LARGE 413 error into friendly message', async () => {
    const fakeFile = new File(['fake-content'], 'test.mp4', { type: 'video/mp4' })
    const mockFetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 413,
      text: () => Promise.resolve('FUNCTION_PAYLOAD_TOO_LARGE sin1::zrzsg-1788929608204-01d40f265bb6'),
    })
    vi.stubGlobal('fetch', mockFetch)

    await expect(uploadTaskProof(fakeFile)).rejects.toThrow(
      'Ukuran file melebihi batas server (maksimal 4 MB). Silakan perkecil atau rekam ulang file bukti.'
    )
  })

  it('translates Request Entity Too Large into friendly message even if status is not 413', async () => {
    const fakeFile = new File(['fake-content'], 'test.mp4', { type: 'video/mp4' })
    const mockFetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
      text: () => Promise.resolve('Request Entity Too Large'),
    })
    vi.stubGlobal('fetch', mockFetch)

    await expect(uploadTaskProof(fakeFile)).rejects.toThrow(
      'Ukuran file melebihi batas server (maksimal 4 MB). Silakan perkecil atau rekam ulang file bukti.'
    )
  })

  it('returns JSON response on successful upload', async () => {
    const fakeFile = new File(['fake-content'], 'test.mp4', { type: 'video/mp4' })
    const expected = { file_url: 'https://storage/vid.mp4', file_name: 'test.mp4', file_size: 100 }
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(expected),
    })
    vi.stubGlobal('fetch', mockFetch)

    const res = await uploadTaskProof(fakeFile)
    expect(res).toEqual(expected)
  })
})
