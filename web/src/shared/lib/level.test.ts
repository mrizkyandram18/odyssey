import { describe, it, expect } from 'vitest'
import { levelProgress } from './level'

describe('levelProgress (backend curve: level = floor(sqrt(xp/100)) + 1)', () => {
  it.each([
    // xp, level, have, required
    [99, 1, 99, 100],
    [100, 2, 0, 300],
    [399, 2, 299, 300],
    [400, 3, 0, 500],
    [1599, 4, 699, 700],
    [1600, 5, 0, 900],
    [0, 1, 0, 100],
  ])('xp=%i level=%i → %i/%i', (xp, level, have, required) => {
    const p = levelProgress(xp, level)
    expect(p.have).toBe(have)
    expect(p.required).toBe(required)
  })

  it('never divides by zero and clamps percent', () => {
    const p = levelProgress(100000, 32)
    expect(p.percent).toBeLessThanOrEqual(100)
    expect(p.required).toBeGreaterThan(0)
  })
})
