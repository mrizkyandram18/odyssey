// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, cleanup, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { RewardShopPage } from './RewardShopPage'
import { useSession } from '../../shared/hooks/useSession'
import { shopApi } from '../../shared/lib/api'

vi.mock('../../shared/hooks/useSession', () => ({
  useSession: vi.fn(),
}))

vi.mock('../../shared/lib/api', () => ({
  shopApi: {
    getConfig: vi.fn(),
    getCatalog: vi.fn(),
    redeem: vi.fn(),
    getMyClaims: vi.fn(),
  },
  adminTasksApi: {
    getConfig: vi.fn(),
    updateConfig: vi.fn(),
    getTasks: vi.fn(),
    getPendingSubmissions: vi.fn(),
    getClaims: vi.fn(),
    verifySubmission: vi.fn(),
    processClaim: vi.fn(),
    createTask: vi.fn(),
    updateTask: vi.fn(),
    deleteTask: vi.fn(),
  },
  tasksApi: {
    getToday: vi.fn(),
    submit: vi.fn(),
  },
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
  },
}))

describe('RewardShopPage Component', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useSession).mockReturnValue({
      session: { uid: 'user-1', role: 'SEEKER' } as any,
      profile: { uid: 'user-1', coins: 1250, role: 'SEEKER' } as any,
      loading: false,
      error: null,
      login: vi.fn(),
      logout: vi.fn(),
      refreshProfile: vi.fn(),
    })
    vi.mocked(shopApi.getConfig).mockResolvedValue({
      redemption_start_day: 21,
      redemption_end_day: 26,
      is_open: true,
      current_day: 22,
      conversion_rate: 10,
      timezone: 'Asia/Jakarta',
    })
    vi.mocked(shopApi.getMyClaims).mockResolvedValue([])
  })

  afterEach(() => {
    cleanup()
  })

  it('renders coin balance and calculated cash conversion correctly', async () => {
    render(
      <MemoryRouter>
        <RewardShopPage />
      </MemoryRouter>
    )

    expect(await screen.findByText('● Penukaran Dibuka')).toBeInTheDocument()
    expect(screen.getByText(/1[.,]?250/)).toBeInTheDocument()
    expect(screen.getByText(/12[.,]?500/)).toBeInTheDocument()
    expect(screen.getByText('Tukarkan Koin Sekarang')).toBeInTheDocument()
  })

  it('displays closed state with dynamic configured dates when window is closed', async () => {
    vi.mocked(useSession).mockReturnValue({
      session: { uid: 'user-1', role: 'SEEKER' } as any,
      profile: { uid: 'user-1', coins: 500, role: 'SEEKER' } as any,
      loading: false,
      error: null,
      login: vi.fn(),
      logout: vi.fn(),
      refreshProfile: vi.fn(),
    })

    vi.mocked(shopApi.getConfig).mockResolvedValue({
      redemption_start_day: 10,
      redemption_end_day: 15,
      is_open: false,
      current_day: 2,
      conversion_rate: 10,
      timezone: 'Asia/Jakarta',
    })
    vi.mocked(shopApi.getMyClaims).mockResolvedValue([])

    render(
      <MemoryRouter>
        <RewardShopPage />
      </MemoryRouter>
    )

    expect(await screen.findByText('○ Penukaran Ditutup')).toBeInTheDocument()
    expect(screen.getAllByText(/10[–-]15/).length).toBeGreaterThan(0)
    expect(screen.queryByText('Tukarkan Koin Sekarang')).toBeNull()
  })

  it('switches to history tab without layout collision and displays claims', async () => {
    vi.mocked(useSession).mockReturnValue({
      session: { uid: 'user-1', role: 'SEEKER' } as any,
      profile: { uid: 'user-1', coins: 300, role: 'SEEKER' } as any,
      loading: false,
      error: null,
      login: vi.fn(),
      logout: vi.fn(),
      refreshProfile: vi.fn(),
    })

    vi.mocked(shopApi.getConfig).mockResolvedValue({
      redemption_start_day: 21,
      redemption_end_day: 26,
      is_open: false,
      current_day: 30,
      conversion_rate: 10,
      timezone: 'Asia/Jakarta',
    })
    vi.mocked(shopApi.getMyClaims).mockResolvedValue([
      {
        id: 1,
        user_uid: 'user-1',
        coins_redeemed: 1000,
        target_type: 'EWALLET',
        target_value: 'GoPay - 0812345678',
        status: 'APPROVED',
        created_at: '2026-08-25T10:00:00Z',
      },
    ])

    render(
      <MemoryRouter>
        <RewardShopPage />
      </MemoryRouter>
    )

    const historyTabBtn = await screen.findByTestId('tab-history')
    fireEvent.click(historyTabBtn)

    expect(await screen.findByText('Pencairan EWALLET')).toBeInTheDocument()
    expect(screen.getByText('GoPay - 0812345678')).toBeInTheDocument()
    expect(screen.getByText('Berhasil Ditransfer')).toBeInTheDocument()
  })
})
