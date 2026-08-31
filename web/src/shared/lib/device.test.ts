// @vitest-environment jsdom
import { describe, it, expect, beforeEach } from 'vitest'
import { getOrCreateDeviceId } from './device'

describe('device.ts', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('generates and persists a stable device ID in localStorage', () => {
    const id1 = getOrCreateDeviceId()
    expect(id1).toBeTruthy()
    expect(id1.startsWith('web_')).toBe(true)

    // Second call should return the exact same device ID
    const id2 = getOrCreateDeviceId()
    expect(id2).toBe(id1)
  })
})
