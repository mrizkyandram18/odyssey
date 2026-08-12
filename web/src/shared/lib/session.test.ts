import { describe, it, expect, beforeEach, vi } from 'vitest'
import { saveSession, getSession, clearSession, isSessionExpired } from './session'
import type { Session } from '../types'

const localStorageMock = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: (key: string): string | null => store[key] || null,
    setItem: (key: string, value: string): void => {
      store[key] = value
    },
    removeItem: (key: string): void => {
      delete store[key]
    },
    clear: (): void => {
      store = {}
    },
  }
})()

beforeEach(() => {
  localStorageMock.clear()
  vi.stubGlobal('localStorage', localStorageMock)
})

function makeSession(overrides: Partial<Session> = {}): Session {
  return {
    uid: 'alice',
    family_id: 'crew-1',
    kind: 'user',
    role: 'SEEKER',
    expires: Math.floor(Date.now() / 1000) + 3600,
    token: 'token-abc',
    ...overrides,
  }
}

describe('saveSession / getSession', () => {
  it('round-trips a session', () => {
    const session = makeSession()
    saveSession(session)
    expect(getSession()).toEqual(session)
  })

  it('returns null when no session is stored', () => {
    expect(getSession()).toBeNull()
  })

  it('returns null for invalid JSON in storage', () => {
    localStorageMock.setItem('odyssey_session', 'not-json')
    expect(getSession()).toBeNull()
  })
})

describe('clearSession', () => {
  it('removes the session from localStorage', () => {
    localStorageMock.setItem('odyssey_session', '{"uid":"alice"}')
    clearSession()
    expect(getSession()).toBeNull()
  })
})

describe('isSessionExpired', () => {
  it('returns true for null session', () => {
    expect(isSessionExpired(null)).toBe(true)
  })

  it('returns true for an expired session', () => {
    expect(isSessionExpired(makeSession({ expires: Math.floor(Date.now() / 1000) - 1 }))).toBe(true)
  })

  it('returns false for a valid session', () => {
    expect(isSessionExpired(makeSession())).toBe(false)
  })

  it('returns false at the exact expiry boundary minus one', () => {
    const now = Math.floor(Date.now() / 1000)
    expect(isSessionExpired(makeSession({ expires: now + 1 }))).toBe(false)
  })
})
