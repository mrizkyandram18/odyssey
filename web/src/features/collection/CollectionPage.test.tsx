// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { CollectionPage } from './CollectionPage'
import { useSession } from '../../shared/hooks/useSession'
import { rewardsApi } from '../../shared/lib/api'

vi.mock('../../shared/hooks/useSession', () => ({
  useSession: vi.fn(),
}))

vi.mock('../../shared/lib/api', () => ({
  rewardsApi: {
    status: vi.fn(),
    claimTicket: vi.fn(),
    openReward: vi.fn(),
    collection: vi.fn(),
    equip: vi.fn(),
  },
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
  },
}))

describe('CollectionPage Component', () => {
  const refreshProfileMock = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useSession).mockReturnValue({
      profile: {
        uid: 'user-1',
        explorer_name: 'Explorer One',
        avatar_seed: 'seed1',
        avatar_frame: 'gold',
        equipped_explorer_effect: 'sparkle',
      } as any,
      refreshProfile: refreshProfileMock,
      loading: false,
      error: null,
      login: vi.fn(),
      logout: vi.fn(),
      session: { uid: 'user-1', family_id: 'fam-1', role: 'MEMBER', kind: 'user' } as any,
    })

    vi.mocked(rewardsApi.status).mockResolvedValue({ tickets: 3 })
    vi.mocked(rewardsApi.collection).mockResolvedValue({
      items: [
        {
          cosmetic_id: 'frame-gold',
          slot: 'frame',
          asset: 'gold',
          tier: 2,
          owned: true,
          equipped: true,
        },
        {
          cosmetic_id: 'effect-sparkle',
          slot: 'effect',
          asset: 'sparkle',
          tier: 1,
          owned: true,
          equipped: false,
        },
        {
          cosmetic_id: 'effect-trail',
          slot: 'effect',
          asset: 'trail',
          tier: 0,
          owned: false,
          equipped: false,
        },
      ],
    })
    vi.mocked(rewardsApi.equip).mockResolvedValue({ status: 'success' })
  })

  it('renders title, tickets left, and cosmetic items with ownership states', async () => {
    render(
      <MemoryRouter>
        <CollectionPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('🎨 Koleksi Saya')).toBeInTheDocument()
    })

    expect(screen.getByText(/Tiket Hadiah tersisa: 3/i)).toBeInTheDocument()
    expect(screen.getByText(/Bingkai ★★/i)).toBeInTheDocument()
    expect(screen.getByText(/Efek ★/i)).toBeInTheDocument()
    expect(screen.getByText('✓ Terpasang')).toBeInTheDocument()
    expect(screen.getByText('Pasang di Profil')).toBeInTheDocument()
    expect(screen.getByText('Belum dimiliki')).toBeInTheDocument()
  })

  it('calls equip API and refreshProfile when clicking Pasang di Profil', async () => {
    render(
      <MemoryRouter>
        <CollectionPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getAllByRole('button', { name: /Pasang di Profil/i })[0]).toBeInTheDocument()
    })

    fireEvent.click(screen.getAllByRole('button', { name: /Pasang di Profil/i })[0])

    await waitFor(() => {
      expect(rewardsApi.equip).toHaveBeenCalledWith('effect-sparkle')
      expect(refreshProfileMock).toHaveBeenCalled()
    })
  })
})
