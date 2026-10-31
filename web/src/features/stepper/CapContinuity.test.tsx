// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { LinearPath } from './LinearPath'
import * as api from '../../shared/lib/api'
import type { TodayTasksResponse, TaskView } from '../../shared/types'

vi.mock('../../shared/hooks/useSession', () => ({
  useSession: vi.fn().mockReturnValue({
    session: { uid: 'test-user-uid', family_id: 'demo-crew-1', kind: 'user' },
    profile: {
      uid: 'test-user-uid',
      explorer_name: 'Explorer Test',
      coins: 3320,
      xp: 500,
      level: 3,
      monthly_coin_target: 3320,
      streak_days: 5,
    },
    refreshProfile: vi.fn(),
  }),
}))

vi.mock('../home/TicketCard', () => ({
  TicketCard: () => <div data-testid="mock-ticket-card">TicketCard</div>,
}))

vi.mock('../../shared/lib/api', () => ({
  tasksApi: {
    getToday: vi.fn(),
    submitTask: vi.fn(),
  },
  shopApi: {
    getConfig: vi.fn().mockResolvedValue({
      effective_monthly_target: 3320,
      earned_this_period: 3320,
      conversion_rate: 100,
      effective_minimum_withdrawal: 500,
    }),
  },
  rewardsApi: {
    getTicketStatus: vi.fn().mockResolvedValue({ tickets: 0 }),
    claimDailyTicket: vi.fn().mockResolvedValue({ granted: false, tickets: 0 }),
  },
  profileApi: {
    getMe: vi.fn().mockResolvedValue({
      uid: 'test-user-uid',
      explorer_name: 'Explorer Test',
      coins: 3320,
      xp: 500,
      level: 3,
      monthly_coin_target: 3320,
    }),
  },
}))

describe('LinearPath Cap Continuity & Notification', () => {
  const mockTasks: TaskView[] = [
    {
      id: 101,
      title: 'Misi Kuis Harian',
      description: 'Jawab kuis berikut',
      task_type: 'QUIZ',
      evaluation_type: 'AUTO',
      step_order: 1,
      reward_coins: 20,
      reward_xp: 100,
      config: {},
      is_locked: false,
      earning_locked: true,
      status: 'UNLOCKED',
      coins_earned: 0,
      xp_earned: 0,
    },
    {
      id: 102,
      title: 'Misi Foto Bukti',
      description: 'Ambil foto',
      task_type: 'PHOTO_UPLOAD',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 2,
      reward_coins: 20,
      reward_xp: 100,
      config: {},
      is_locked: true,
      earning_locked: true,
      status: 'LOCKED',
      coins_earned: 0,
      xp_earned: 0,
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders persistent cap-reached banner when earningLocked is true', async () => {
    vi.mocked(api.tasksApi.getToday).mockResolvedValueOnce({
      tasks: mockTasks,
      earning_locked: true,
      earning_cap: 1883,
      earned: 1883,
    })

    render(
      <MemoryRouter>
        <LinearPath />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByTestId('cap-reached-banner')).toBeInTheDocument()
    })

    expect(screen.getByText('Target Coin Bulan Ini Tercapai!')).toBeInTheDocument()
    expect(
      screen.getByText(/Kamu tidak wajib mengerjakan misi lagi untuk mendapatkan Coin bulan ini/i)
    ).toBeInTheDocument()
  })

  it('allows clicking and opening tasks even when earningLocked is true', async () => {
    vi.mocked(api.tasksApi.getToday).mockResolvedValueOnce({
      tasks: mockTasks,
      earning_locked: true,
      earning_cap: 1883,
      earned: 1883,
    })

    render(
      <MemoryRouter>
        <LinearPath />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getAllByText('Misi Kuis Harian').length).toBeGreaterThan(0)
    })

    const taskButtons = screen.getAllByRole('button', { name: /Misi Kuis Harian/i })
    expect(taskButtons[0]).not.toBeDisabled()

    fireEvent.click(taskButtons[0])
    expect(taskButtons[0]).toBeInTheDocument()
  })
})
