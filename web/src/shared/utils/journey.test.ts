import { describe, it, expect } from 'vitest'
import {
  KNOWN_REALMS,
  getRealmMetadata,
  formatRealmName,
  getRealmForMission,
  isRealmUnlocked,
  getMergedJourneyProgress,
} from './journey'

describe('journey utilities', () => {
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

  it('formats journey names properly', () => {
    expect(formatRealmName('whispering-woods')).toBe('Whispering Woods')
    expect(formatRealmName('clockwork-city')).toBe('Clockwork City')
    expect(formatRealmName('custom-journey-name')).toBe('Custom Journey Name')
  })

  it('maps Mission template slugs to realms', () => {
    expect(getRealmForMission({ template_slug: 'morning-light' })).toBe('whispering-woods')
    expect(getRealmForMission({ template_slug: 'clockwork-intro' })).toBe('clockwork-city')
    expect(getRealmForMission({ template_slug: 'star-observation' })).toBe('starlit-library')
    expect(getRealmForMission({ template_slug: 'unknown-Mission' })).toBe('whispering-woods')
    expect(getRealmForMission({ template_slug: 'any', journey: 'clockwork-city' })).toBe('clockwork-city')
  })

  it('identifies unlocked status correctly', () => {
    expect(isRealmUnlocked('ACTIVE')).toBe(true)
    expect(isRealmUnlocked('COMPLETE')).toBe(true)
    expect(isRealmUnlocked('LOCKED')).toBe(false)
  })

  it('merges server progress with default realms', () => {
    const merged = getMergedJourneyProgress([
      { family_id: 'c1', journey: 'whispering-woods', status: 'COMPLETE', progress: 100, updated_at: '' },
      { family_id: 'c1', journey: 'clockwork-city', status: 'ACTIVE', progress: 25, updated_at: '' },
    ])

    expect(merged[0].status).toBe('COMPLETE')
    expect(merged[0].progress).toBe(100)
    expect(merged[1].status).toBe('ACTIVE')
    expect(merged[1].progress).toBe(25)
    expect(merged[2].status).toBe('LOCKED')
    expect(merged[2].progress).toBe(0)
  })
})
