// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, cleanup, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { ProfilePage } from './ProfilePage'
import { useSession } from '../../shared/hooks/useSession'
import { apiClient } from '../../shared/lib/api'

vi.mock('../../shared/hooks/useSession', () => ({
  useSession: vi.fn(),
}))

vi.mock('../../shared/lib/api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../shared/lib/api')>()
  return {
    ...actual,
    apiClient: {
      get: vi.fn(),
      post: vi.fn(),
      patch: vi.fn(),
    },
  }
})

describe('ProfilePage', () => {
  const logoutMock = vi.fn()
  const refreshProfileMock = vi.fn()

  beforeEach(() => {
    vi.resetAllMocks()
    logoutMock.mockReset()
    refreshProfileMock.mockReset()
  })

  afterEach(() => {
    cleanup()
  })

  describe('Member / Seeker Role View', () => {
    beforeEach(() => {
      vi.mocked(useSession).mockReturnValue({
        session: { uid: 'u1', family_id: 'fam-1', role: 'SEEKER' } as any,
        profile: {
          uid: 'u1',
          family_id: 'fam-1',
          explorer_name: 'Budi Santoso',
          role: 'SEEKER',
          level: 3,
          xp: 250,
          coins: 1500,
          streak_days: 5,
          avatar_style: 'adventurer',
          avatar_seed: 'seed123',
          created_at: '',
          updated_at: '',
        } as any,
        loading: false,
        error: null,
        login: vi.fn(),
        logout: logoutMock,
        refreshProfile: refreshProfileMock,
      })
    })

    it('renders member profile details, XP, Level, Coins, and withdrawal CTA correctly', () => {
      render(
        <MemoryRouter>
          <ProfilePage />
        </MemoryRouter>
      )

      expect(screen.getByText('Budi Santoso')).toBeInTheDocument()
      expect(screen.getByText('Anggota')).toBeInTheDocument()
      expect(screen.getByText('Level 3')).toBeInTheDocument()
      expect(screen.getByText('250 XP')).toBeInTheDocument()
      expect(screen.getByText('1500')).toBeInTheDocument()
      expect(screen.getByText('5 Hari Streak')).toBeInTheDocument()
      expect(screen.getByText('Pencairan Koin')).toBeInTheDocument()

      const shopLink = screen.getByRole('link', { name: /Pencairan Koin/i })
      expect(shopLink).toHaveAttribute('href', '/shop')
    })

    it('switches to settings view when clicking settings button', async () => {
      render(
        <MemoryRouter>
          <ProfilePage />
        </MemoryRouter>
      )

      const settingsBtn = screen.getByText('Pengaturan')
      fireEvent.click(settingsBtn)

      expect(screen.getByText('Informasi Akun')).toBeInTheDocument()
      expect(screen.getByText('Keluar (Sign Out)')).toBeInTheDocument()
      expect(screen.getByText('Ubah Kata Sandi')).toBeInTheDocument()
    })

    it('calls randomize avatar API when clicking acak avatar', async () => {
      vi.mocked(apiClient.patch).mockResolvedValueOnce({})

      render(
        <MemoryRouter>
          <ProfilePage />
        </MemoryRouter>
      )

      fireEvent.click(screen.getByText('Pengaturan'))
      const randomizeBtn = screen.getAllByRole('button', { name: /Acak Avatar/i })[0]
      fireEvent.click(randomizeBtn)

      await waitFor(() => {
        expect(apiClient.patch).toHaveBeenCalledWith('/api/me/avatar', expect.objectContaining({
          avatar_style: 'adventurer',
        }))
      })
    })

    it('renders change password form in settings and submits successfully', async () => {
      vi.mocked(apiClient.post).mockResolvedValueOnce({
        status: 'success',
        message: 'Kata sandi berhasil diubah',
      })

      render(
        <MemoryRouter>
          <ProfilePage />
        </MemoryRouter>
      )

      fireEvent.click(screen.getByText('Pengaturan'))
      expect(screen.getByText('Ubah Kata Sandi')).toBeInTheDocument()

      const currentPassInput = screen.getByPlaceholderText('Masukkan kata sandi saat ini')
      const newPassInput = screen.getByPlaceholderText('Minimal 6 karakter')
      const confirmInput = screen.getByPlaceholderText('Ulangi kata sandi baru')
      const submitBtn = screen.getByText('Simpan Kata Sandi Baru')

      fireEvent.change(currentPassInput, { target: { value: 'oldPass123' } })
      fireEvent.change(newPassInput, { target: { value: 'newPass456' } })
      fireEvent.change(confirmInput, { target: { value: 'newPass456' } })
      fireEvent.click(submitBtn)

      await waitFor(() => {
        expect(apiClient.post).toHaveBeenCalledWith('/api/me/change-password', {
          current_password: 'oldPass123',
          new_password: 'newPass456',
          confirm_password: 'newPass456',
        })
        expect(screen.getByText('Kata sandi berhasil diubah')).toBeInTheDocument()
      })
    })

    it('validates password mismatch and current == new in profile settings', async () => {
      render(
        <MemoryRouter>
          <ProfilePage />
        </MemoryRouter>
      )

      fireEvent.click(screen.getByText('Pengaturan'))

      const currentPassInput = screen.getByPlaceholderText('Masukkan kata sandi saat ini')
      const newPassInput = screen.getByPlaceholderText('Minimal 6 karakter')
      const confirmInput = screen.getByPlaceholderText('Ulangi kata sandi baru')
      const submitBtn = screen.getByText('Simpan Kata Sandi Baru')

      // Same current and new
      fireEvent.change(currentPassInput, { target: { value: 'samePass123' } })
      fireEvent.change(newPassInput, { target: { value: 'samePass123' } })
      fireEvent.change(confirmInput, { target: { value: 'samePass123' } })
      fireEvent.click(submitBtn)

      expect(await screen.findByText('Kata sandi baru tidak boleh sama dengan kata sandi saat ini')).toBeInTheDocument()
      expect(apiClient.post).not.toHaveBeenCalled()
    })
  })

  describe('Admin / Guide / Builder Role View', () => {
    beforeEach(() => {
      vi.mocked(useSession).mockReturnValue({
        session: { uid: 'admin-1', family_id: 'fam-1', role: 'ADMIN' } as any,
        profile: {
          uid: 'admin-1',
          family_id: 'fam-1',
          explorer_name: 'Super Admin',
          role: 'ADMIN',
          level: 1,
          xp: 0,
          coins: 0,
          streak_days: 0,
          avatar_style: 'adventurer',
          avatar_seed: 'admin-seed',
          created_at: '',
          updated_at: '',
        } as any,
        loading: false,
        error: null,
        login: vi.fn(),
        logout: logoutMock,
        refreshProfile: refreshProfileMock,
      })
    })

    it('renders clean administrator account view with identity, account info, change password, and logout', () => {
      render(
        <MemoryRouter>
          <ProfilePage />
        </MemoryRouter>
      )

      // Title & Identity
      expect(screen.getByText('Akun Administrator')).toBeInTheDocument()
      expect(screen.getByText('Super Admin')).toBeInTheDocument()
      expect(screen.getByText('Administrator (ADMIN)')).toBeInTheDocument()

      // Account info & UID
      expect(screen.getByText('Informasi Akun')).toBeInTheDocument()
      expect(screen.getByText('admin-1')).toBeInTheDocument()

      // Essential forms & actions
      expect(screen.getByText('Ubah Kata Sandi')).toBeInTheDocument()
      expect(screen.getByPlaceholderText('Masukkan kata sandi saat ini')).toBeInTheDocument()
      expect(screen.getByPlaceholderText('Minimal 6 karakter')).toBeInTheDocument()
      expect(screen.getByPlaceholderText('Ulangi kata sandi baru')).toBeInTheDocument()
      expect(screen.getByText('Simpan Kata Sandi Baru')).toBeInTheDocument()
      expect(screen.getByText('Keluar (Sign Out)')).toBeInTheDocument()
    })

    it('does NOT render gamification stats, XP, Level, Coins, Streak, or withdrawal CTA for ADMIN', () => {
      render(
        <MemoryRouter>
          <ProfilePage />
        </MemoryRouter>
      )

      // Must NOT render player gamification elements
      expect(screen.queryByText(/Level/i)).toBeNull()
      expect(screen.queryByText(/XP/i)).toBeNull()
      expect(screen.queryByText(/Koin/i)).toBeNull()
      expect(screen.queryByText(/Streak/i)).toBeNull()
      expect(screen.queryByText('Pencairan Koin')).toBeNull()
      expect(screen.queryByText('Ringkasan')).toBeNull()
      expect(screen.queryByText('Pengaturan')).toBeNull()

      // Must NOT contain /shop navigation link
      const links = screen.queryAllByRole('link')
      const shopLinks = links.filter(l => l.getAttribute('href') === '/shop')
      expect(shopLinks.length).toBe(0)
    })

    it('allows Admin to submit change password successfully', async () => {
      vi.mocked(apiClient.post).mockResolvedValueOnce({
        status: 'success',
        message: 'Kata sandi berhasil diubah',
      })

      render(
        <MemoryRouter>
          <ProfilePage />
        </MemoryRouter>
      )

      const currentPassInput = screen.getByPlaceholderText('Masukkan kata sandi saat ini')
      const newPassInput = screen.getByPlaceholderText('Minimal 6 karakter')
      const confirmInput = screen.getByPlaceholderText('Ulangi kata sandi baru')
      const submitBtn = screen.getByText('Simpan Kata Sandi Baru')

      fireEvent.change(currentPassInput, { target: { value: 'adminCurrent123' } })
      fireEvent.change(newPassInput, { target: { value: 'adminNew456' } })
      fireEvent.change(confirmInput, { target: { value: 'adminNew456' } })
      fireEvent.click(submitBtn)

      await waitFor(() => {
        expect(apiClient.post).toHaveBeenCalledWith('/api/me/change-password', {
          current_password: 'adminCurrent123',
          new_password: 'adminNew456',
          confirm_password: 'adminNew456',
        })
        expect(screen.getByText('Kata sandi berhasil diubah')).toBeInTheDocument()
      })
    })

    it('calls logout when Admin clicks Keluar button', () => {
      render(
        <MemoryRouter>
          <ProfilePage />
        </MemoryRouter>
      )

      const logoutBtn = screen.getByText('Keluar (Sign Out)')
      fireEvent.click(logoutBtn)

      expect(logoutMock).toHaveBeenCalled()
    })

    it('also renders clean administrator view for GUIDE role', () => {
      vi.mocked(useSession).mockReturnValue({
        session: { uid: 'guide-1', family_id: 'fam-1', role: 'GUIDE' } as any,
        profile: {
          uid: 'guide-1',
          family_id: 'fam-1',
          explorer_name: 'Guide Parent',
          role: 'GUIDE',
          level: 1,
          xp: 0,
          coins: 0,
          streak_days: 0,
        } as any,
        loading: false,
        error: null,
        login: vi.fn(),
        logout: logoutMock,
        refreshProfile: refreshProfileMock,
      })

      render(
        <MemoryRouter>
          <ProfilePage />
        </MemoryRouter>
      )

      expect(screen.getByText('Akun Administrator')).toBeInTheDocument()
      expect(screen.getByText('Administrator (GUIDE)')).toBeInTheDocument()
      expect(screen.queryByText(/Level/i)).toBeNull()
      expect(screen.queryByText('Pencairan Koin')).toBeNull()
    })

    it('also renders clean administrator view for BUILDER role', () => {
      vi.mocked(useSession).mockReturnValue({
        session: { uid: 'builder-1', family_id: 'fam-1', role: 'BUILDER' } as any,
        profile: {
          uid: 'builder-1',
          family_id: 'fam-1',
          explorer_name: 'Builder Parent',
          role: 'BUILDER',
          level: 1,
          xp: 0,
          coins: 0,
          streak_days: 0,
        } as any,
        loading: false,
        error: null,
        login: vi.fn(),
        logout: logoutMock,
        refreshProfile: refreshProfileMock,
      })

      render(
        <MemoryRouter>
          <ProfilePage />
        </MemoryRouter>
      )

      expect(screen.getByText('Akun Administrator')).toBeInTheDocument()
      expect(screen.getByText('Administrator (BUILDER)')).toBeInTheDocument()
      expect(screen.queryByText(/Level/i)).toBeNull()
      expect(screen.queryByText('Pencairan Koin')).toBeNull()
    })
  })
})
