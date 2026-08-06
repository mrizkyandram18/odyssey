import { createContext, useContext, useState, useEffect } from 'react'
import type { ReactNode } from 'react'
import { saveSession, getSession, clearSession, isSessionExpired } from '../shared/lib/session'
import type { Session, DevicePayload, Explorer } from '../shared/types'
import type { AuthClient, LoginOutcome } from '../shared/lib/auth'

interface SessionContextValue {
    session: Session | null
    profile: Explorer | null
    loading: boolean
    error: string | null
    login: (uid: string, credential: string, device: DevicePayload) => Promise<void>
    logout: () => void
}

export const SessionContext = createContext<SessionContextValue | undefined>(undefined)

export interface SessionProviderProps {
    children: ReactNode
    authClient: AuthClient
}

export function SessionProvider({ children, authClient }: SessionProviderProps) {
    const [session, setSession] = useState<Session | null>(null)
    const [profile, setProfile] = useState<Explorer | null>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        const bootstrap = async () => {
            const stored = getSession()
            if (!stored || isSessionExpired(stored)) {
                clearSession()
                setSession(null)
                setProfile(null)
                setLoading(false)
                return
            }

            const currentProfile = await authClient.current()
            if (currentProfile) {
                setSession(stored)
                setProfile(currentProfile)
            } else {
                clearSession()
                setSession(null)
                setProfile(null)
            }
            setLoading(false)
        }

        void bootstrap()
    }, [authClient])

    const handleLoginOutcome = (result: LoginOutcome): Session => {
        if (result.ok) return result.session
        if (result.variant === 'setup_needed') {
            throw new Error('Password setup is required. Please set a password first.')
        }
        throw new Error('Password is required.')
    }

    const login = async (uid: string, credential: string, device: DevicePayload): Promise<void> => {
        setError(null)
        const result = await authClient.login(uid, credential, device)
        const newSession = handleLoginOutcome(result)
        saveSession(newSession)
        setSession(newSession)
        const currentProfile = await authClient.current()
        setProfile(currentProfile)
    }

    const logout = (): void => {
        authClient.logout()
        setSession(null)
        setProfile(null)
        setError(null)
    }

    return (
        <SessionContext.Provider value={{ session, profile, loading, error, login, logout }}>
            {children}
        </SessionContext.Provider>
    )
}

export function useSessionContext() {
    const ctx = useContext(SessionContext)
    if (!ctx) throw new Error('useSessionContext must be used within SessionProvider')
    return ctx
}
