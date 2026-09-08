// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor, cleanup } from '@testing-library/react'
import { CapsuleModal } from './CapsuleModal'
import { useSession } from '../../shared/hooks/useSession'
import { rewardsApi } from '../../shared/lib/api'

vi.mock('../../shared/hooks/useSession', () => ({
  useSession: vi.fn(),
}))

vi.mock('../../shared/lib/api', () => ({
  rewardsApi: {
    openReward: vi.fn(),
    equip: vi.fn(),
  },
}))

vi.mock('canvas-confetti', () => ({
  default: vi.fn(),
}))

describe('CapsuleModal Component', () => {
  const refreshProfileMock = vi.fn()
  const onCloseMock = vi.fn()
  const onOpenedMock = vi.fn()

  afterEach(() => {
    cleanup()
  })

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useSession).mockReturnValue({
      profile: { uid: 'u1' } as any,
      refreshProfile: refreshProfileMock,
      loading: false,
      error: null,
      login: vi.fn(),
      logout: vi.fn(),
      session: { uid: 'u1' } as any,
    })
  })

  it('renders Buka Hadiah button and opens reward on click', async () => {
    vi.mocked(rewardsApi.openReward).mockResolvedValue({
      cosmetic_id: 'frame-gold',
      slot: 'frame',
      asset: 'gold',
      tier: 1,
      is_new: true,
    })

    render(<CapsuleModal onClose={onCloseMock} onOpened={onOpenedMock} />)

    expect(screen.getByText('Gunakan 1 Tiket Hadiah untuk membuka kotak kejutan berisi hiasan profil.')).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /Buka Hadiah/i }))

    await waitFor(() => {
      expect(screen.getByText(/Kamu mendapatkan hadiah!/i)).toBeInTheDocument()
      expect(screen.getByText(/Bingkai Profil/i)).toBeInTheDocument()
    })

    expect(onOpenedMock).toHaveBeenCalled()
  })

  it('allows equipping reward directly and calls refreshProfile', async () => {
    vi.mocked(rewardsApi.openReward).mockResolvedValue({
      cosmetic_id: 'effect-sparkle',
      slot: 'effect',
      asset: 'sparkle',
      tier: 2,
      is_new: false,
    })
    vi.mocked(rewardsApi.equip).mockResolvedValue({ status: 'success' })

    render(<CapsuleModal onClose={onCloseMock} onOpened={onOpenedMock} />)

    fireEvent.click(screen.getByRole('button', { name: /Buka Hadiah/i }))

    await waitFor(() => {
      expect(screen.getByText(/Efek Profil ★2/i)).toBeInTheDocument()
    })

    const equipBtn = screen.getByRole('button', { name: /Pasang di Profil/i })
    fireEvent.click(equipBtn)

    await waitFor(() => {
      expect(rewardsApi.equip).toHaveBeenCalledWith('effect-sparkle')
      expect(refreshProfileMock).toHaveBeenCalled()
      expect(screen.getByText(/Terpasang di profil!/i)).toBeInTheDocument()
    })
  })
})
