const DEVICE_ID_KEY = 'odyssey_device_id'

export function getOrCreateDeviceId(): string {
  try {
    const existing = localStorage.getItem(DEVICE_ID_KEY)
    if (existing && existing.trim() !== '') {
      return existing.trim()
    }
  } catch {
    // Storage may be unavailable in restricted environments
  }

  // Generate a random stable installation device ID
  const randomPart = typeof crypto !== 'undefined' && crypto.randomUUID
    ? crypto.randomUUID()
    : 'dev_' + Math.random().toString(36).substring(2, 15) + '_' + Date.now().toString(36)
  
  const deviceId = `web_${randomPart}`

  try {
    localStorage.setItem(DEVICE_ID_KEY, deviceId)
  } catch {
    // Ignore storage write error
  }

  return deviceId
}
