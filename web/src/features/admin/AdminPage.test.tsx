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
    getSubmissions: vi.fn(),
    getPendingSubmissions: vi.fn(),
    getClaims: vi.fn(),
    verifySubmission: vi.fn(),
    editSubmission: vi.fn(),
    processClaim: vi.fn(),
    createTask: vi.fn(),
    updateTask: vi.fn(),
    deleteTask: vi.fn(),
  },
  adminMembersApi: {
    getMembers: vi.fn().mockResolvedValue([]),
    createMember: vi.fn(),
    updateMember: vi.fn(),
  },
  adminCosmeticsApi: {
    getCosmetics: vi.fn().mockResolvedValue({ items: [] }),
    createCosmetic: vi.fn(),
    updateCosmetic: vi.fn(),
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
    vi.mocked(adminTasksApi.getSubmissions).mockResolvedValue([])
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

  it('keeps monthly coin target 0 from API through display to save (no 3200 fallback)', async () => {
    vi.mocked(useSession).mockReturnValue({
      session: { uid: '1', family_id: '1', role: 'ADMIN', kind: 'user', expires: 9999999999, token: 'abc' },
      profile: { uid: '1', role: 'ADMIN' },
      loading: false,
    } as any)

    vi.mocked(adminTasksApi.getConfig).mockResolvedValue({
      ...mockConfig,
      default_monthly_coin_target: 0,
      default_monthly_earning_cap: 3320,
      auto_block_inactivity_days: 5,
    })

    render(
      <MemoryRouter>
        <AdminPage />
      </MemoryRouter>
    )

    const settingsTabBtn = screen.getByTestId('admin-tab-settings')
    fireEvent.click(settingsTabBtn)

    await waitFor(() => {
      expect(screen.getByText('Pengaturan Periode Penukaran Koin')).toBeInTheDocument()
    })

    // API value 0 must render as "0", not empty and not replaced by a legacy fallback
    const targetInput = document.getElementById('input-monthly-target') as HTMLInputElement
    const capInput = document.getElementById('input-monthly-cap') as HTMLInputElement
    const autoBlockInput = document.getElementById('input-auto-block') as HTMLInputElement
    await waitFor(() => {
      expect(targetInput.value).toBe('0')
      expect(capInput.value).toBe('3320')
      expect(autoBlockInput.value).toBe('5')
    })

    const saveBtn = screen.getByRole('button', { name: /Simpan Pengaturan Periode/i })
    fireEvent.submit(saveBtn.closest('form')!)

    await waitFor(() => {
      expect(adminTasksApi.updateConfig).toHaveBeenCalledWith(expect.objectContaining({
        default_monthly_coin_target: 0,
        default_monthly_earning_cap: 3320,
        auto_block_inactivity_days: 5,
      }))
    })
    const sent = vi.mocked(adminTasksApi.updateConfig).mock.calls[0][0] as Record<string, unknown>
    expect(sent.default_monthly_coin_target).toBe(0)
    expect(sent.default_monthly_earning_cap).toBe(3320)
  })

  it('renders pending submission with edit and reject with penalty buttons', async () => {
    vi.mocked(useSession).mockReturnValue({
      session: { uid: '1', family_id: '1', role: 'ADMIN', kind: 'user', expires: 9999999999, token: 'abc' },
      profile: { uid: '1', role: 'ADMIN' },
      loading: false,
    } as any)

    const mockPendingSub = [
      {
        id: 10,
        task_id: 100,
        task_title: 'Tugas Menulis',
        task_type: 'TEXT_RESPONSE',
        user_uid: 'user-1',
        user_name: 'Budi',
        submission_type: 'MANUAL_VERIFY',
        status: 'PENDING',
        payload: { text: 'Jawaban salah' },
        created_at: new Date().toISOString(),
        reward_coins: 50,
        reward_xp: 100,
      } as any,
    ]
    vi.mocked(adminTasksApi.getSubmissions).mockResolvedValue(mockPendingSub)
    vi.mocked(adminTasksApi.getPendingSubmissions).mockResolvedValue(mockPendingSub)

    render(
      <MemoryRouter>
        <AdminPage initialTab="submissions" />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Tugas Menulis')).toBeInTheDocument()
      expect(screen.getByText('Edit Jawaban')).toBeInTheDocument()
      expect(screen.getByText(/Penalti Koin jika Ditolak/i)).toBeInTheDocument()
    })

    // Click Edit Jawaban button
    const editBtn = screen.getByText('Edit Jawaban')
    fireEvent.click(editBtn)

    await waitFor(() => {
      expect(screen.getByText('Edit Jawaban Submission')).toBeInTheDocument()
      expect(screen.getByDisplayValue('Jawaban salah')).toBeInTheDocument()
    })
  })

  it('hides XP inputs and badges from Admin UI while preserving default reward_xp', async () => {
    vi.mocked(useSession).mockReturnValue({
      session: { uid: '1', family_id: '1', role: 'ADMIN', kind: 'user', expires: 9999999999, token: 'abc' },
      profile: { uid: '1', role: 'ADMIN' },
      loading: false,
    } as any)

    vi.mocked(adminTasksApi.getTasks).mockResolvedValue([
      { id: 1, title: 'Tugas Harian 1', step_order: 1, reward_coins: 50, reward_xp: 250, task_type: 'VIDEO', is_locked: false, status: 'UNLOCKED', config: {}, coins_earned: 0, xp_earned: 0 },
    ])

    render(
      <MemoryRouter>
        <AdminPage initialTab="submissions" />
      </MemoryRouter>
    )

    // Check verification subtitle does not contain "& EXP"
    await waitFor(() => {
      expect(screen.getByText('Antrean Verifikasi Bukti Tugas (0 Menunggu / 0 Total)')).toBeInTheDocument()
      expect(screen.getByText(/Approve memberi koin otomatis/i)).toBeInTheDocument()
      expect(screen.queryByText(/Approve memberi koin & EXP otomatis/i)).toBeNull()
    })

    // Navigate to tasks tab
    const tasksTabBtn = screen.getByTestId('admin-tab-tasks')
    fireEvent.click(tasksTabBtn)

    await waitFor(() => {
      expect(screen.getByText('Tugas Harian 1')).toBeInTheDocument()
      expect(screen.getByText('+50 🪙')).toBeInTheDocument()
      // XP badge should not be in the task card
      expect(screen.queryByText('+250 XP')).toBeNull()
    })

    // Open Edit Task modal
    const editTaskBtn = screen.getByRole('button', { name: /Edit tugas Tugas Harian 1/i })
    fireEvent.click(editTaskBtn)

    await waitFor(() => {
      expect(screen.getByText('Edit Tugas')).toBeInTheDocument()
      // XP input should not exist in the modal
      expect(screen.queryByText(/XP ✨/i)).toBeNull()
    })

    // Save Edit Task and verify reward_xp is preserved
    const saveTaskBtn = screen.getByRole('button', { name: /Simpan Perubahan/i })
    fireEvent.click(saveTaskBtn)

    await waitFor(() => {
      expect(adminTasksApi.updateTask).toHaveBeenCalledWith(1, expect.objectContaining({
        reward_coins: 50,
        reward_xp: 250,
      }))
    })
  })

  it('navigates to Hadiah tab and renders cosmetic catalog separately from economy settings', async () => {
    vi.mocked(useSession).mockReturnValue({
      session: { uid: '1', family_id: '1', role: 'ADMIN', kind: 'user', expires: 9999999999, token: 'abc' },
      profile: { uid: '1', role: 'ADMIN' },
      loading: false,
    } as any)

    render(
      <MemoryRouter>
        <AdminPage />
      </MemoryRouter>
    )

    const rewardsTabBtn = screen.getByTestId('admin-tab-rewards')
    fireEvent.click(rewardsTabBtn)

    await waitFor(() => {
      expect(screen.getByText('Katalog Hadiah & Koleksi')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /Tambah Hadiah/i })).toBeInTheDocument()
    })
  })

  it('renders visual level cap bonus tiers and allows adding a tier without touching raw JSON', async () => {
    vi.mocked(useSession).mockReturnValue({
      session: { uid: '1', family_id: '1', role: 'ADMIN', kind: 'user', expires: 9999999999, token: 'abc' },
      profile: { uid: '1', role: 'ADMIN' },
      loading: false,
    } as any)

    vi.mocked(adminTasksApi.getConfig).mockResolvedValue({
      ...mockConfig,
      default_monthly_earning_cap: 3320,
      max_monthly_earning_cap_ceiling: 10000,
      level_cap_bonus: { '5': 300 },
    })

    render(
      <MemoryRouter>
        <AdminPage />
      </MemoryRouter>
    )

    const settingsTabBtn = screen.getByTestId('admin-tab-settings')
    fireEvent.click(settingsTabBtn)

    await waitFor(() => {
      expect(screen.getByText('Bonus Batas Koin Berdasarkan Tingkat')).toBeInTheDocument()
      expect(screen.getByText('Tingkat 5')).toBeInTheDocument()
      expect(screen.getByText('+300 Koin')).toBeInTheDocument()
    })

    // Add new tier (Tingkat 10 -> +500)
    const levelInput = screen.getByPlaceholderText(/Tingkat \(misal: 15\)/i)
    const bonusInput = screen.getByPlaceholderText(/Tambahan Koin \(misal: 750\)/i)
    fireEvent.change(levelInput, { target: { value: '10' } })
    fireEvent.change(bonusInput, { target: { value: '500' } })

    const addTierBtn = screen.getByRole('button', { name: /Tambah/i })
    fireEvent.click(addTierBtn)

    await waitFor(() => {
      expect(screen.getByText('Tingkat 10')).toBeInTheDocument()
      expect(screen.getByText('+500 Koin')).toBeInTheDocument()
    })

    // Submit form and verify updated level_cap_bonus
    const saveBtn = screen.getByRole('button', { name: /Simpan Pengaturan Periode & Ekonomi/i })
    fireEvent.submit(saveBtn.closest('form')!)

    await waitFor(() => {
      expect(adminTasksApi.updateConfig).toHaveBeenCalledWith(expect.objectContaining({
        level_cap_bonus: { '5': 300, '10': 500 },
      }))
    })
  })
})
