import { Routes, Route, Navigate, useParams } from 'react-router-dom'
import { ProtectedRoute } from '../shared/components/ProtectedRoute'
import { PublicRoute } from '../shared/components/PublicRoute'
import { LoginPage } from '../features/login/LoginPage'
import { HomePage } from '../features/home/HomePage'
import { QuestsPage } from '../features/quest/QuestsPage'
import { QuestView } from '../features/quest/QuestView'
import { FamilyTimeline } from '../features/creative/FamilyTimeline'
import { GalleryPage } from '../features/creative/GalleryPage'
import { ComicReaderPage } from '../features/creative/ComicReaderPage'
import { StoryPage } from '../features/creative/StoryPage'
import { JournalPage } from '../features/journal/JournalPage'
import { ProfilePage } from '../features/profile/ProfilePage'
import { ChestPage } from '../features/chests/ChestPage'
import { ChestOpeningPage } from '../features/chests/ChestOpeningPage'
import { RelicInventoryPage } from '../features/relics/RelicInventoryPage'
import { RelicDetailPage } from '../features/relics/RelicDetailPage'
import { AdminPage } from '../features/admin/AdminPage'

export function App() {
  return (
    <Routes>
      <Route element={<PublicRoute />}>
        <Route path="/login" element={<LoginPage />} />
      </Route>

      <Route element={<ProtectedRoute />}>
        <Route index element={<HomePage />} />
        <Route path="/quests" element={<QuestsPage />} />
        <Route path="/quests/:questId" element={<QuestRoute />} />
      <Route path="/creative" element={<FamilyTimeline />} />
      <Route path="/gallery" element={<GalleryPage />} />
      <Route path="/comics/:id" element={<ComicReaderPage />} />
      <Route path="/stories/:id" element={<StoryPage />} />
      <Route path="/journal" element={<JournalPage />} />
        <Route path="/profile" element={<ProfilePage />} />
        <Route path="/chests" element={<ChestPage />} />
        <Route path="/chests/open/:chestId" element={<ChestOpeningPage />} />
        <Route path="/relics" element={<RelicInventoryPage />} />
        <Route path="/relics/:slug" element={<RelicDetailPage />} />
        <Route path="/admin" element={<AdminPage />} />
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

function QuestRoute() {
  const params = useParams()
  return <QuestView questId={Number(params?.questId)} />
}
