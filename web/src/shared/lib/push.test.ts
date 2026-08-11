import { describe, it, expect } from 'vitest'
import { urlBase64ToUint8Array, arrayBufferToBase64, isPushSupported } from './push'

describe('push lib utilities', () => {
  it('converts base64url string to Uint8Array', () => {
    // Simple ASCII string encoded as URL-safe base64: btoa("hello") = "aGVsbG8="
    const b64url = 'aGVsbG8' // "hello" without padding
    const arr = urlBase64ToUint8Array(b64url)
    expect(arr).toBeInstanceOf(Uint8Array)
    expect(arr.length).toBe(5) // "hello" is 5 bytes
    expect(arr[0]).toBe(0x68) // 'h'
    expect(arr[1]).toBe(0x65) // 'e'
  })

  it('converts ArrayBuffer to URL-safe base64 string', () => {
    // Encode [0x68, 0x65, 0x6c, 0x6c, 0x6f] = "hello"
    const buffer = new Uint8Array([0x68, 0x65, 0x6c, 0x6c, 0x6f]).buffer
    const b64 = arrayBufferToBase64(buffer)
    expect(typeof b64).toBe('string')
    expect(b64.length).toBeGreaterThan(0)
    // Should not contain standard base64 chars that were replaced
    expect(b64).not.toContain('+')
    expect(b64).not.toContain('/')
    expect(b64).not.toMatch(/=+$/)
  })

  it('handles null or empty buffer safely', () => {
    expect(arrayBufferToBase64(null)).toBe('')
  })

  it('returns false for isPushSupported when environment is incomplete', () => {
    // Vitest jsdom environment lacks PushManager
    const supported = isPushSupported()
    expect(typeof supported).toBe('boolean')
    // In jsdom, PushManager is not available, so this should be false
    expect(supported).toBe(false)
  })
})
