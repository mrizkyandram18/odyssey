import { describe, expect, it } from 'vitest'
import { deriveRelayLegs, memberName } from './relayRotation'
import type { Challenge, CrewMember } from '../types'

const doneLeg = (id: number, by: string): Challenge => ({
  id,
  quest_id: 1,
  slug: `leg-${id}`,
  description: `Leg ${id}`,
  status: 'DONE',
  completed_by: by,
  created_at: '2026-01-01T00:00:00Z',
})

const pendingLeg = (id: number, assignedTo?: string): Challenge => ({
  id,
  quest_id: 1,
  slug: `leg-${id}`,
  description: `Leg ${id}`,
  status: 'PENDING',
  assigned_to: assignedTo ?? null,
  created_at: '2026-01-01T00:00:00Z',
})

const members: CrewMember[] = [
  { uid: 'u1', explorer_name: 'Leo', role: 'SEEKER' },
  { uid: 'u2', explorer_name: 'Maya', role: 'GUIDE' },
  { uid: 'u3', explorer_name: 'Sam', role: 'BUILDER' },
]

describe('deriveRelayLegs', () => {
  it('returns [] for an empty challenge list', () => {
    expect(deriveRelayLegs([], undefined, 'u1')).toEqual([])
  })

  it('marks the assigned pending leg as active', () => {
    const legs = deriveRelayLegs([doneLeg(1, 'u1'), pendingLeg(2, 'u2')], 'u2', 'u2')
    expect(legs[0].state).toBe('done')
    expect(legs[1].state).toBe('active')
    expect(legs[1].isMyTurn).toBe(true)
  })

  it('active turn belongs only to the assigned uid', () => {
    const legs = deriveRelayLegs([doneLeg(1, 'u1'), pendingLeg(2, 'u2')], 'u2', 'u3')
    expect(legs[1].state).toBe('active')
    expect(legs[1].isMyTurn).toBe(false)
  })

  it('marks the leg after the active one as next', () => {
    const legs = deriveRelayLegs([doneLeg(1, 'u1'), pendingLeg(2, 'u2'), pendingLeg(3)], 'u2', 'u2')
    expect(legs.map((l) => l.state)).toEqual(['done', 'active', 'next'])
  })

  it('falls back to the first pending leg when nothing is assigned yet', () => {
    const legs = deriveRelayLegs([pendingLeg(1), pendingLeg(2)], undefined, 'u1')
    expect(legs[0].state).toBe('active')
    expect(legs[0].isMyTurn).toBe(false)
    expect(legs[1].state).toBe('next')
  })

  it('unassigned active leg is not marked as my turn', () => {
    const legs = deriveRelayLegs([pendingLeg(1)], undefined, 'u1')
    expect(legs[0].state).toBe('active')
    expect(legs[0].isMyTurn).toBe(false)
  })

  it('marks all remaining pending legs as open', () => {
    const legs = deriveRelayLegs([doneLeg(1, 'u1'), pendingLeg(2, 'u2'), pendingLeg(3), pendingLeg(4)], 'u2', 'u3')
    expect(legs.map((l) => l.state)).toEqual(['done', 'active', 'next', 'open'])
  })

  it('classifies a fully completed relay as all done', () => {
    const legs = deriveRelayLegs([doneLeg(1, 'u1'), doneLeg(2, 'u2')], undefined, 'u1')
    expect(legs.every((l) => l.state === 'done')).toBe(true)
  })
})

describe('memberName', () => {
  it('resolves the explorer name from the roster', () => {
    expect(memberName(members, 'u2')).toBe('Maya')
  })

  it('falls back to the raw uid when not in the roster', () => {
    expect(memberName(members, 'unknown-uid')).toBe('unknown-uid')
  })

  it('handles missing roster and empty uid', () => {
    expect(memberName(undefined, 'u1')).toBe('u1')
    expect(memberName(members, null)).toBe('')
    expect(memberName(members, undefined)).toBe('')
  })
})
