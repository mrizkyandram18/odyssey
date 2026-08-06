import type { Session } from '../types'

const SESSION_KEY = 'odyssey_session'

export function saveSession(session: Session): void {
  try {
    localStorage.setItem(SESSION_KEY, JSON.stringify(session))
  } catch {
    // Storage may be unavailable (e.g. private browsing mode)
  }
}

export function getSession(): Session | null {
  const raw = localStorage.getItem(SESSION_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw) as Session
  } catch {
    return null
  }
}

export function clearSession(): void {
  localStorage.removeItem(SESSION_KEY)
}

// isSessionExpired returns true when the session is null or its
// expiry timestamp (in Unix seconds) has passed.
export function isSessionExpired(session: Session | null): boolean {
  if (!session) return true
  return Date.now() / 1000 >= session.expires
}
