import { describe, it, expect } from 'vitest'
import { calculateTargetPace } from './targetPace'

describe('calculateTargetPace', () => {
  it('returns ON_TRACK for target <= 0', () => {
    const res = calculateTargetPace(100, 0)
    expect(res.status).toBe('ON_TRACK')
    expect(res.label).toBe('Tanpa Target')
  })

  it('returns ACHIEVED when earned >= target', () => {
    const res = calculateTargetPace(3500, 3320, new Date(2026, 8, 15))
    expect(res.status).toBe('ACHIEVED')
    expect(res.targetPercent).toBe(100)
    expect(res.label).toContain('Target Tercapai')
  })

  it('returns CRITICAL when day >= 20 and targetPercent < 50%', () => {
    // September has 30 days. Sep 25 = day 25.
    // 1000 earned out of 3320 is ~30%
    const res = calculateTargetPace(1000, 3320, new Date(2026, 8, 25))
    expect(res.status).toBe('CRITICAL')
    expect(res.label).toContain('Di Bawah Target')
  })

  it('returns BEHIND when day >= 12 and targetPercent < monthProgress * 0.6', () => {
    // Sep 15 is 50% through the month. 60% of 50% = 30%.
    // 600 out of 3320 is 18%, which is < 30%.
    const res = calculateTargetPace(600, 3320, new Date(2026, 8, 15))
    expect(res.status).toBe('BEHIND')
    expect(res.label).toContain('Tertinggal')
  })

  it('returns ON_TRACK when pace is good mid-month', () => {
    // Sep 15 is 50% through the month.
    // 1800 out of 3320 is ~54%, which is healthy.
    const res = calculateTargetPace(1800, 3320, new Date(2026, 8, 15))
    expect(res.status).toBe('ON_TRACK')
    expect(res.label).toContain('Sesuai Target')
  })
})
