// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { TicketCard } from './TicketCard'
import { rewardsApi } from '../../shared/lib/api'
import { useSession } from '../../shared/hooks/useSession'

vi.mock('../../shared/hooks/useSession', () => ({
  useSession: vi.fn(),
}))

vi.mock('../../shared/lib/api', () => ({
  rewardsApi: {
    status: vi.fn(),
    claimTicket: vi.fn(),
    openReward: vi.fn(),
  },
}))

describe('TicketCard Component', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useSession).mockReturnValue({
      profile: { uid: 'u1' } as any,
      refreshProfile: vi.fn(),
      loading: false,
      error: null,
      login: vi.fn(),
      logout: vi.fn(),
      session: { uid: 'u1' } as any,
    })
  })

  it('renders ticket balance when tickets > 0 and shows Buka button', async () => {
    vi.mocked(rewardsApi.status).mockResolvedValue({ tickets: 2 })

    render(
      <MemoryRouter>
        <TicketCard />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Tiket Hadiah: 2')).toBeInTheDocument()
      expect(screen.getByText('Buka')).toBeInTheDocument()
    })
  })

  it('renders Ambil button when 0 tickets and claims ticket on click', async () => {
    vi.mocked(rewardsApi.status).mockResolvedValue({ tickets: 0 })
    vi.mocked(rewardsApi.claimTicket).mockResolvedValue({ granted: true, tickets: 1 })

    render(
      <MemoryRouter>
        <TicketCard />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Tiket Hadiah: 0')).toBeInTheDocument()
      expect(screen.getByText('Ambil')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByText('Ambil'))

    await waitFor(() => {
      expect(rewardsApi.claimTicket).toHaveBeenCalled()
    })
  })
})
