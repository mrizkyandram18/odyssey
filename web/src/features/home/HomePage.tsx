import { Navigate } from 'react-router-dom'
import { LinearPath } from '../stepper/LinearPath'
import { OnboardingModal } from './OnboardingModal'
import { useSession } from '../../shared/hooks/useSession'

export function HomePage() {
  const { profile } = useSession()
  const isAdmin = profile?.role === 'ADMIN' || profile?.role === 'GUIDE' || profile?.role === 'BUILDER'

  if (isAdmin) {
    return <Navigate to="/admin" replace />
  }

  return (
    <>
      <OnboardingModal />
      <LinearPath />
    </>
  )
}
