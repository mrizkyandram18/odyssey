import { describe, it, expect } from 'vitest'
import { createAvatar } from '@dicebear/core'
import { adventurer } from '@dicebear/collection'

describe('Avatar deterministic rendering', () => {
  it('generates the exact same SVG data uri for a given seed', () => {
    const avatar1 = createAvatar(adventurer, {
      seed: 'deterministic-test-seed',
      backgroundColor: ['b6e3f4', 'c0aede', 'd1d4f9', 'ffd5dc', 'ffdfbf'],
      radius: 50,
    })
    
    const avatar2 = createAvatar(adventurer, {
      seed: 'deterministic-test-seed',
      backgroundColor: ['b6e3f4', 'c0aede', 'd1d4f9', 'ffd5dc', 'ffdfbf'],
      radius: 50,
    })

    const uri1 = avatar1.toDataUri()
    const uri2 = avatar2.toDataUri()

    expect(uri1).toEqual(uri2)
    expect(uri1).toContain('data:image/svg+xml;utf8,')
  })

  it('generates a different SVG for a different seed', () => {
    const avatar1 = createAvatar(adventurer, { seed: 'seed-A' }).toDataUri()
    const avatar2 = createAvatar(adventurer, { seed: 'seed-B' }).toDataUri()

    expect(avatar1).not.toEqual(avatar2)
  })
})
