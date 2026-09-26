// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import { describe, it, expect } from 'vitest'
import { rfc3339ToLocalInput, localInputToRfc3339 } from './useAdminConfig'

describe('Gate 2-D — timezone-aware announcement schedule conversion', () => {
  it('round-trips stored RFC3339 through the system timezone without hardcoding +07:00', () => {
    // Asia/Jakarta wall time, offset derived from the zone — never a literal.
    expect(rfc3339ToLocalInput('2026-09-12T00:00:00+07:00', 'Asia/Jakarta')).toBe('2026-09-12T00:00')
    expect(localInputToRfc3339('2026-09-12T00:00', 'Asia/Jakarta')).toBe('2026-09-12T00:00:00+07:00')
    expect(localInputToRfc3339('2026-09-30T23:59', 'Asia/Jakarta')).toBe('2026-09-30T23:59:00+07:00')
  })

  it('resolves the offset from the configured zone, not a constant', () => {
    // Same wall time in UTC carries a different offset — proves no hardcoded +07:00.
    expect(localInputToRfc3339('2026-09-12T00:00', 'UTC')).toBe('2026-09-12T00:00:00+00:00')
    expect(rfc3339ToLocalInput('2026-09-12T00:00:00Z', 'UTC')).toBe('2026-09-12T00:00')
    // Cross-zone instant preservation: 00:00 Jakarta == 17:00 UTC previous day.
    expect(rfc3339ToLocalInput('2026-09-12T00:00:00+07:00', 'UTC')).toBe('2026-09-11T17:00')
  })

  it('treats empty as unbounded (clear affordance)', () => {
    expect(rfc3339ToLocalInput('', 'Asia/Jakarta')).toBe('')
    expect(localInputToRfc3339('', 'Asia/Jakarta')).toBe('')
    expect(localInputToRfc3339('   ', 'Asia/Jakarta')).toBe('')
  })

  it('passes unparseable input through so RFC3339 validation can report it', () => {
    expect(localInputToRfc3339('not-a-date', 'Asia/Jakarta')).toBe('not-a-date')
    expect(rfc3339ToLocalInput('not-a-date', 'Asia/Jakarta')).toBe('')
  })

  it('handles a DST zone: summer/winter offsets and round-trips', () => {
    // America/New_York: EDT (-04:00) in July, EST (-05:00) in January.
    expect(localInputToRfc3339('2026-07-01T12:00', 'America/New_York')).toBe('2026-07-01T12:00:00-04:00')
    expect(localInputToRfc3339('2026-01-01T12:00', 'America/New_York')).toBe('2026-01-01T12:00:00-05:00')
    // Stored instant renders as zone wall time in both seasons.
    expect(rfc3339ToLocalInput('2026-07-01T16:00:00Z', 'America/New_York')).toBe('2026-07-01T12:00')
    expect(rfc3339ToLocalInput('2026-01-01T17:00:00Z', 'America/New_York')).toBe('2026-01-01T12:00')
    // Round-trip across a DST boundary keeps the intended instant.
    const stored = '2026-03-10T12:00:00-04:00' // after US spring-forward (Mar 8 2026)
    const local = rfc3339ToLocalInput(stored, 'America/New_York')
    expect(local).toBe('2026-03-10T12:00')
    expect(localInputToRfc3339(local, 'America/New_York')).toBe(stored)
  })

  it('renders non-zero-offset stored values as zone wall time', () => {
    // Same instant written with a +00:00 offset still renders Jakarta wall time.
    expect(rfc3339ToLocalInput('2026-09-11T17:00:00Z', 'Asia/Jakarta')).toBe('2026-09-12T00:00')
  })
})
