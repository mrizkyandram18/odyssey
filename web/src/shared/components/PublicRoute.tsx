import { Navigate, Outlet } from 'react-router-dom'
import { useSession } from '../hooks/useSession'

// PublicRoute guards routes that should only be visible to unauthenticated
// users (e.g. the login page). While loading, it renders a full-screen
// loading indicator. If a session already exists, it redirects to /.
export function PublicRoute() {
  const { session, loading } = useSession()

  if (loading) {
    return (
      <div className="flex h-screen w-full items-center justify-center">
        <div className="text-sm text-muted-foreground">Loading...</div>
      </div>
    )
  }

  if (session) {
    return <Navigate to="/" replace />
  }

  return <Outlet />
}
