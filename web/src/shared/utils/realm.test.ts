import { describe, it, expect } from 'vitest'
import {
  KNOWN_REALMS,
  getRealmMetadata,
  formatRealmName,
  getRealmForQuest,
  isRealmUnlocked,
  getMergedRealmProgress,
} from './realm'

describe('realm utilities', () => {
  it('defines known realms in order', () => {
    expect(KNOWN_REALMS.map((r) => r.slug)).toEqual([
      'whispering-woods',
      'clockwork-city',
      'starlit-library',
    ])
  })

  it('gets metadata by slug', () => {
    expect(getRealmMetadata('clockwork-city')?.name).toBe('Clockwork City')
    expect(getRealmMetadata('UNKNOWN')).toBeUndefined()
  })

  it('formats realm names properly', () => {
    expect(formatRealmName('whispering-woods')).toBe('Whispering Woods')
    expect(formatRealmName('clockwork-city')).toBe('Clockwork City')
    expect(formatRealmName('custom-realm-name')).toBe('Custom Realm Name')
  })

  it('maps quest template slugs to realms', () => {
    expect(getRealmForQuest({ template_slug: 'morning-light' })).toBe('whispering-woods')
    expect(getRealmForQuest({ template_slug: 'clockwork-intro' })).toBe('clockwork-city')
    expect(getRealmForQuest({ template_slug: 'star-observation' })).toBe('starlit-library')
    expect(getRealmForQuest({ template_slug: 'unknown-quest' })).toBe('whispering-woods')
    expect(getRealmForQuest({ template_slug: 'any', realm: 'clockwork-city' })).toBe('clockwork-city')
  })

  it('identifies unlocked status correctly', () => {
    expect(isRealmUnlocked('ACTIVE')).toBe(true)
    expect(isRealmUnlocked('COMPLETE')).toBe(true)
    expect(isRealmUnlocked('LOCKED')).toBe(false)
  })

  it('merges server progress with default realms', () => {
    const merged = getMergedRealmProgress([
      { crew_id: 'c1', realm: 'whispering-woods', status: 'COMPLETE', progress: 100, updated_at: '' },
      { crew_id: 'c1', realm: 'clockwork-city', status: 'ACTIVE', progress: 25, updated_at: '' },
    ])

    expect(merged[0].status).toBe('COMPLETE')
    expect(merged[0].progress).toBe(100)
    expect(merged[1].status).toBe('ACTIVE')
    expect(merged[1].progress).toBe(25)
    expect(merged[2].status).toBe('LOCKED')
    expect(merged[2].progress).toBe(0)
  })
})
