import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useSession } from '../hooks/useSession'
import { AppLayout } from './layout/AppLayout'

// ProtectedRoute guards routes that require an authenticated session.
// While loading, it renders a full-screen loading indicator.
// If no session is present, it redirects to /login.
export function ProtectedRoute() {
  const { session, loading } = useSession()
  const location = useLocation()

  if (loading) {
    return (
      <div className="flex h-screen w-full items-center justify-center bg-bg-app">
        <div className="flex flex-col items-center gap-4">
          <div className="flex h-16 w-16 items-center justify-center rounded-full bg-surface-elevated border border-accent-magic/30 shadow-[0_0_20px_rgba(6,182,222,0.2)] animate-pulse">
            <span className="text-3xl">🧭</span>
          </div>
          <div className="text-sm text-text-secondary">Memuat aplikasi...</div>
        </div>
      </div>
    )
  }

  if (!session) {
    return <Navigate to="/login" state={{ from: location.pathname }} replace />
  }

  return (
    <AppLayout>
      <Outlet />
    </AppLayout>
  )
}
