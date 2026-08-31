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
  beforeEach(() => {
    vi.resetAllMocks()
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
      logout: vi.fn(),
      refreshProfile: vi.fn(),
    })
  })

  afterEach(() => {
    cleanup()
  })

  it('renders user profile details and coins correctly', () => {
    render(
      <MemoryRouter>
        <ProfilePage />
      </MemoryRouter>
    )

    expect(screen.getByText('Budi Santoso')).toBeInTheDocument()
    expect(screen.getByText('Anggota')).toBeInTheDocument()
    expect(screen.getByText('Level 3')).toBeInTheDocument()
    expect(screen.getByText('1500')).toBeInTheDocument()
    expect(screen.getByText('5 Hari Streak')).toBeInTheDocument()
    expect(screen.getByText('Pencairan Koin')).toBeInTheDocument()
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
  })

  it('calls randomize avatar API when clicking acak avatar', async () => {
    vi.mocked(apiClient.patch).mockResolvedValueOnce({})

    render(
      <MemoryRouter>
        <ProfilePage />
      </MemoryRouter>
    )

    // Go to settings
    fireEvent.click(screen.getByText('Pengaturan'))
    const randomizeBtn = screen.getByText('Acak Avatar')
    fireEvent.click(randomizeBtn)

    await waitFor(() => {
      expect(apiClient.patch).toHaveBeenCalledWith('/api/me/avatar', expect.objectContaining({
        avatar_style: 'adventurer',
      }))
    })
  })
})
