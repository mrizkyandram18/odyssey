import { describe, it, expect } from 'vitest'
import { getRoleMastery } from './roleMastery'

describe('roleMastery utility', () => {
  it('returns Novice Seeker for level 1-4', () => {
    expect(getRoleMastery('SEEKER', 1).title).toBe('Novice Seeker')
    expect(getRoleMastery('SEEKER', 4).title).toBe('Novice Seeker')
  })

  it('returns Adept Seeker for level 5-9', () => {
    expect(getRoleMastery('SEEKER', 5).title).toBe('Adept Seeker')
    expect(getRoleMastery('SEEKER', 9).title).toBe('Adept Seeker')
  })

  it('returns Master Seeker for level 10+', () => {
    expect(getRoleMastery('SEEKER', 10).title).toBe('Master Seeker')
    expect(getRoleMastery('SEEKER', 99).title).toBe('Master Seeker')
  })

  it('returns Novice Builder for level 1-4', () => {
    expect(getRoleMastery('BUILDER', 1).title).toBe('Novice Builder')
    expect(getRoleMastery('BUILDER', 4).title).toBe('Novice Builder')
  })

  it('returns Adept Builder for level 5-9', () => {
    expect(getRoleMastery('BUILDER', 5).title).toBe('Adept Builder')
    expect(getRoleMastery('BUILDER', 9).title).toBe('Adept Builder')
  })

  it('returns Master Builder for level 10+', () => {
    expect(getRoleMastery('BUILDER', 10).title).toBe('Master Builder')
  })

  it('returns Novice Guide for level 1-4', () => {
    expect(getRoleMastery('GUIDE', 1).title).toBe('Novice Guide')
    expect(getRoleMastery('GUIDE', 4).title).toBe('Novice Guide')
  })

  it('returns Adept Guide for level 5-9', () => {
    expect(getRoleMastery('GUIDE', 5).title).toBe('Adept Guide')
    expect(getRoleMastery('GUIDE', 9).title).toBe('Adept Guide')
  })

  it('returns Master Guide for level 10+', () => {
    expect(getRoleMastery('GUIDE', 10).title).toBe('Master Guide')
  })

  it('handles lowercase roles', () => {
    expect(getRoleMastery('seeker', 5).title).toBe('Adept Seeker')
  })

  it('provides a safe fallback for unknown roles', () => {
    const fallback = getRoleMastery('UNKNOWN_ROLE', 5)
    expect(fallback.title).toBe('Level 5 Explorer')
    expect(fallback.flavor).toBe('An intrepid explorer ready for any adventure.')
  })

  it('handles level 0 or negative gracefully by treating as level 1', () => {
    expect(getRoleMastery('SEEKER', 0).title).toBe('Novice Seeker')
    expect(getRoleMastery('SEEKER', -5).title).toBe('Novice Seeker')
    expect(getRoleMastery('UNKNOWN_ROLE', 0).title).toBe('Level 1 Explorer')
  })

  it('handles empty role gracefully', () => {
    expect(getRoleMastery('', 5).title).toBe('Adept Seeker')
  })

  it('always returns an object with title and flavor', () => {
    const res = getRoleMastery('GUIDE', 12)
    expect(res).toHaveProperty('title')
    expect(res).toHaveProperty('flavor')
    expect(typeof res.title).toBe('string')
    expect(typeof res.flavor).toBe('string')
  })
})
