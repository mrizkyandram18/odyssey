import { describe, it, expect } from 'vitest'
import { toSvgDataUri } from './svg'

describe('toSvgDataUri', () => {
  it('encodes svg correctly', () => {
    const svg = '<svg><circle cx="50" /></svg>'
    const uri = toSvgDataUri(svg)
    expect(uri).toBe('data:image/svg+xml;utf8,%3Csvg%3E%3Ccircle%20cx%3D%2250%22%20%2F%3E%3C%2Fsvg%3E')
  })
})
