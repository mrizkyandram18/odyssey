import { Routes, Route, Navigate, useParams } from 'react-router-dom'
import { ErrorBoundary } from '../shared/components/ErrorBoundary'
import { useSession } from '../shared/hooks/useSession'
import { ProtectedRoute } from '../shared/components/ProtectedRoute'
import { PublicRoute } from '../shared/components/PublicRoute'
import { LoginPage } from '../features/login/LoginPage'
import { HomePage } from '../features/home/HomePage'
import { MissionsPage } from '../features/mission/MissionsPage'
import { MissionView } from '../features/mission/MissionView'
import { FamilyTimeline } from '../features/creative/FamilyTimeline'
import { GalleryPage } from '../features/creative/GalleryPage'
import { ComicReaderPage } from '../features/creative/ComicReaderPage'
import { StoryPage } from '../features/creative/StoryPage'
import { JournalPage } from '../features/journal/JournalPage'
import { ProfilePage } from '../features/profile/ProfilePage'
import { GiftPage } from '../features/gifts/GiftPage'
import { GiftOpeningPage } from '../features/gifts/GiftOpeningPage'
import { CollectionInventoryPage } from '../features/collections/CollectionInventoryPage'
import { CollectionDetailPage } from '../features/collections/CollectionDetailPage'
import { AdminPage } from '../features/admin/AdminPage'

export function App() {
  return (
    <ErrorBoundary>
      <Routes>
        <Route element={<PublicRoute />}>
          <Route path="/login" element={<LoginPage />} />
        </Route>

        <Route element={<ProtectedRoute />}>
          <Route index element={<HomePage />} />
          <Route path="/missions" element={<MissionsPage />} />
          <Route path="/missions/:missionId" element={<MissionRoute />} />
        <Route path="/creative" element={<FamilyTimeline />} />
        <Route path="/gallery" element={<GalleryPage />} />
        <Route path="/comics/:id" element={<ComicReaderPage />} />
        <Route path="/stories/:id" element={<StoryPage />} />
        <Route path="/journal" element={<JournalPage />} />
          <Route path="/profile" element={<ProfilePage />} />
          <Route path="/gifts" element={<GiftPage />} />
          <Route path="/gifts/open/:chestId" element={<GiftOpeningPage />} />
          <Route path="/collections" element={<RelicRoute><CollectionInventoryPage /></RelicRoute>} />
          <Route path="/collections/:slug" element={<RelicRoute><CollectionDetailPage /></RelicRoute>} />
          <Route path="/admin" element={<AdminPage />} />
        </Route>

        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </ErrorBoundary>
  )
}

function MissionRoute() {
  const params = useParams()
  return <MissionView missionId={Number(params?.missionId)} />
}

function RelicRoute({ children }: { children: React.ReactNode }) {
  const { profile } = useSession()
  if (profile?.role !== 'GUIDE') return <Navigate to="/profile" replace />
  return <>{children}</>
}
