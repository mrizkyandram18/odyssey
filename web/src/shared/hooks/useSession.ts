import { useSessionContext } from '../../app/SessionProvider'

export function useSession() {
  const { session, profile, loading, error, login, logout, refreshProfile } = useSessionContext()

  return {
    session,
    profile,
    loading,
    error,
    login,
    logout,
    refreshProfile,
  }
}
