// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, afterEach, beforeEach } from 'vitest'
import { render, screen, cleanup, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { LinearPath } from './LinearPath'
import { tasksApi, shopApi } from '../../shared/lib/api'
import { useSession } from '../../shared/hooks/useSession'

vi.mock('../../shared/lib/api', () => ({
  tasksApi: {
    getToday: vi.fn(),
  },
  shopApi: {
    getConfig: vi.fn(),
  },
}))

vi.mock('../../shared/hooks/useSession', () => ({
  useSession: vi.fn(),
}))

vi.mock('../home/TicketCard', () => ({
  TicketCard: () => <div data-testid="mock-ticket-card">TicketCard</div>,
}))

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

describe('LinearPath Component', () => {
  beforeEach(() => {
    vi.mocked(useSession).mockReturnValue({
      profile: {
        uid: 'user-1',
        explorer_name: 'Budi',
        coins: 850,
        xp: 1200,
        streak_days: 3,
        level: 4,
        monthly_coin_target: 3320,
      },
      session: { uid: 'user-1' },
      refreshProfile: vi.fn(),
    } as any)

    vi.mocked(tasksApi.getToday).mockResolvedValue({
      tasks: [],
      earning_locked: false,
      earning_cap: 4000,
      earned: 1500,
    } as any)

    vi.mocked(shopApi.getConfig).mockResolvedValue({
      redemption_start_day: 1,
      redemption_end_day: 31,
      payout_day: 1,
      conversion_rate: 100,
      effective_monthly_target: 3320,
      earned_this_period: 1500,
      effective_minimum_withdrawal: 500,
    } as any)
  })

  it('renders Monthly Target Checkpoint card with pace progress', async () => {
    render(
      <MemoryRouter>
        <LinearPath />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByTestId('monthly-target-checkpoint')).toBeInTheDocument()
    })

    expect(screen.getByText('Target Koin Bulanan')).toBeInTheDocument()
    expect(screen.getByText(/1\.500 \/ 3\.320/)).toBeInTheDocument()
    expect(screen.getByText(/Bisa dicairkan sekarang \(min\. 500 koin\)/i)).toBeInTheDocument()
  })

  it('does not render checkpoint widget if target is 0 or unset', async () => {
    vi.mocked(shopApi.getConfig).mockResolvedValue({
      redemption_start_day: 1,
      redemption_end_day: 31,
      payout_day: 1,
      conversion_rate: 100,
      effective_monthly_target: 0,
      earned_this_period: 0,
    } as any)

    vi.mocked(useSession).mockReturnValue({
      profile: {
        uid: 'user-1',
        explorer_name: 'Budi',
        coins: 100,
        xp: 100,
        level: 1,
        monthly_coin_target: 0,
      },
      session: { uid: 'user-1' },
      refreshProfile: vi.fn(),
    } as any)

    render(
      <MemoryRouter>
        <LinearPath />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.queryByTestId('monthly-target-checkpoint')).toBeNull()
    })
  })
})
