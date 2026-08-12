import { describe, it, expect } from 'vitest'
import { isMyRelayTurn } from './missionTurn'

const relayActive = {
  Mission_type: 'RELAY',
  status: 'ACTIVE',
  active_challenge_assigned_to: 'u1',
}

describe('isMyRelayTurn', () => {
  it('returns true when an active relay Mission is assigned to me', () => {
    expect(isMyRelayTurn(relayActive, 'u1')).toBe(true)
  })

  it('returns false when assigned to someone else', () => {
    expect(isMyRelayTurn(relayActive, 'u2')).toBe(false)
  })

  it('returns false when no session uid is available', () => {
    expect(isMyRelayTurn(relayActive, null)).toBe(false)
    expect(isMyRelayTurn(relayActive, undefined)).toBe(false)
    expect(isMyRelayTurn(relayActive, '')).toBe(false)
  })

  it('returns false for non-relay Mission types', () => {
    expect(isMyRelayTurn({ ...relayActive, Mission_type: 'SOLO' }, 'u1')).toBe(false)
    expect(isMyRelayTurn({ ...relayActive, Mission_type: 'CREATIVE' }, 'u1')).toBe(false)
  })

  it('returns false for non-active missions', () => {
    expect(isMyRelayTurn({ ...relayActive, status: 'PENDING' }, 'u1')).toBe(false)
    expect(isMyRelayTurn({ ...relayActive, status: 'DONE' }, 'u1')).toBe(false)
  })

  it('returns false when no assignment is present', () => {
    expect(isMyRelayTurn({ ...relayActive, active_challenge_assigned_to: undefined }, 'u1')).toBe(false)
  })

  it('handles null and undefined Mission', () => {
    expect(isMyRelayTurn(null, 'u1')).toBe(false)
    expect(isMyRelayTurn(undefined, 'u1')).toBe(false)
  })
})
