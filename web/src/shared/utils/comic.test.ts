import { describe, expect, it } from 'vitest'
import {
  buildComicPayload,
  isComicReady,
  parseComicPayload,
  MAX_COMIC_CAPTION,
  MIN_COMIC_PANELS,
} from './comic'

describe('parseComicPayload', () => {
  it('parses a valid multi-panel comic', () => {
    const raw = JSON.stringify({
      v: 1,
      panels: [{ caption: 'A' }, { caption: 'B', svg: '<svg></svg>' }],
    })
    const parsed = parseComicPayload(raw)
    expect(parsed?.panels).toHaveLength(2)
    expect(parsed?.panels[0].caption).toBe('A')
    expect(parsed?.panels[1].svg).toBe('<svg></svg>')
  })

  it('returns null for plain story text', () => {
    expect(parseComicPayload('Once upon a time')).toBeNull()
  })

  it('returns null for invalid JSON', () => {
    expect(parseComicPayload('{broken')).toBeNull()
  })
})

describe('buildComicPayload / isComicReady', () => {
  it('builds a versioned payload and trims captions', () => {
    const json = buildComicPayload([
      { caption: '  hello  ' },
      { caption: 'world', svg: ' <svg/> ' },
    ])
    expect(JSON.parse(json)).toEqual({
      v: 1,
      panels: [{ caption: 'hello' }, { caption: 'world', svg: '<svg/>' }],
    })
  })

  it('requires at least two filled panels', () => {
    expect(isComicReady([{ caption: 'only one' }])).toBe(false)
    expect(
      isComicReady([
        { caption: 'one' },
        { caption: 'two' },
      ]),
    ).toBe(true)
    expect(MIN_COMIC_PANELS).toBe(2)
  })

  it('rejects empty panels and overlong captions', () => {
    expect(
      isComicReady([
        { caption: 'ok' },
        { caption: '   ' },
      ]),
    ).toBe(false)
    expect(
      isComicReady([
        { caption: 'a'.repeat(MAX_COMIC_CAPTION + 1) },
        { caption: 'b' },
      ]),
    ).toBe(false)
  })

  it('accepts svg-only panels', () => {
    expect(
      isComicReady([
        { caption: '', svg: '<svg/>' },
        { caption: '', svg: '<svg/>' },
      ]),
    ).toBe(true)
  })
})
