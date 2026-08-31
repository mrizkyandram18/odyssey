import { Routes, Route, Navigate } from 'react-router-dom'
import { ErrorBoundary } from '../shared/components/ErrorBoundary'
import { ProtectedRoute } from '../shared/components/ProtectedRoute'
import { PublicRoute } from '../shared/components/PublicRoute'
import { LoginPage } from '../features/login/LoginPage'
import { HomePage } from '../features/home/HomePage'
import { RewardShopPage } from '../features/shop/RewardShopPage'
import { ProfilePage } from '../features/profile/ProfilePage'
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
          <Route path="/shop" element={<RewardShopPage />} />
          <Route path="/profile" element={<ProfilePage />} />
          <Route path="/admin" element={<AdminPage />} />
        </Route>

        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </ErrorBoundary>
  )
}

