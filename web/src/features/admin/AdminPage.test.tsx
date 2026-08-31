// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, cleanup, waitFor, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { AdminPage } from './AdminPage'
import { useSession } from '../../shared/hooks/useSession'
import { adminTasksApi } from '../../shared/lib/api'

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

describe('AdminPage Component', () => {
  const mockConfig = {
    redemption_start_day: 24,
    redemption_end_day: 26,
    payout_day: 24,
    earning_period_days: 30,
    is_open: true,
    is_payout_day: true,
    current_day: 24,
    conversion_rate: 100,
    payout_target_rupiah: 320000,
    payout_target_coins: 3200,
    max_payout_coins: 3200,
    timezone: 'Asia/Jakarta',
  }

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(adminTasksApi.getTasks).mockResolvedValue([])
    vi.mocked(adminTasksApi.getPendingSubmissions).mockResolvedValue([])
    vi.mocked(adminTasksApi.getClaims).mockResolvedValue([])
    vi.mocked(adminTasksApi.getConfig).mockResolvedValue(mockConfig)
    vi.mocked(adminTasksApi.updateConfig).mockResolvedValue({
      ...mockConfig,
      redemption_start_day: 10,
      redemption_end_day: 15,
      current_day: 12,
    })
  })

  afterEach(() => {
    cleanup()
  })

  it('redirects unauthorized users', () => {
    vi.mocked(useSession).mockReturnValue({
      session: { uid: '1', family_id: '1', role: 'SEEKER', kind: 'user', expires: 9999999999, token: 'abc' },
      profile: { uid: '1', role: 'SEEKER' },
      loading: false,
    } as any)

    render(
      <MemoryRouter>
        <AdminPage />
      </MemoryRouter>
    )

    expect(screen.queryByText('Panel Operasional Admin')).toBeNull()
  })

  it('renders operations dashboard for ADMIN role with metric tiles', async () => {
    vi.mocked(useSession).mockReturnValue({
      session: { uid: '1', family_id: '1', role: 'ADMIN', kind: 'user', expires: 9999999999, token: 'abc' },
      profile: { uid: '1', role: 'ADMIN' },
      loading: false,
    } as any)

    vi.mocked(adminTasksApi.getTasks).mockResolvedValue([
      { id: 1, title: 'Tugas 1', step_order: 1, reward_coins: 50, reward_xp: 100, task_type: 'VIDEO', is_locked: false, status: 'UNLOCKED', config: {}, coins_earned: 0, xp_earned: 0 },
    ])

    render(
      <MemoryRouter>
        <AdminPage />
      </MemoryRouter>
    )

    expect(screen.getByText('Panel Operasional Admin')).toBeInTheDocument()

    await waitFor(() => {
      expect(screen.getByText(/Antrean Verifikasi Bukti Tugas/i)).toBeInTheDocument()
      expect(screen.getByText('Tidak Ada Antrean Verifikasi')).toBeInTheDocument()
      expect(screen.getByText(/24[–-]26/)).toBeInTheDocument()
    })
  })

  it('allows configuring redemption period start and end days with validation', async () => {
    vi.mocked(useSession).mockReturnValue({
      session: { uid: '1', family_id: '1', role: 'GUIDE', kind: 'user', expires: 9999999999, token: 'abc' },
      profile: { uid: '1', role: 'GUIDE' },
      loading: false,
    } as any)

    render(
      <MemoryRouter>
        <AdminPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Panel Operasional Admin')).toBeInTheDocument()
    })

    const settingsTabBtn = screen.getByTestId('admin-tab-settings')
    fireEvent.click(settingsTabBtn)

    await waitFor(() => {
      expect(screen.getByText('Pengaturan Periode Penukaran Koin')).toBeInTheDocument()
    })

    const saveBtn = screen.getByRole('button', { name: /Simpan Pengaturan Periode/i })
    const form = saveBtn.closest('form')!
    fireEvent.submit(form)

    await waitFor(() => {
      expect(adminTasksApi.updateConfig).toHaveBeenCalled()
    })
  })
})
