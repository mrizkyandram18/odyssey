// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, cleanup } from '@testing-library/react'
import { AdminOverview } from './AdminOverview'
import { useAdminSubmissions } from '../../hooks/useAdminSubmissions'
import { useAdminClaims } from '../../hooks/useAdminClaims'
import { useAdminMembers } from '../../hooks/useAdminMembers'
import { useAdminConfig } from '../../hooks/useAdminConfig'

vi.mock('../../hooks/useAdminSubmissions', () => ({
  useAdminSubmissions: vi.fn(),
}))

vi.mock('../../hooks/useAdminClaims', () => ({
  useAdminClaims: vi.fn(),
}))

vi.mock('../../hooks/useAdminMembers', () => ({
  useAdminMembers: vi.fn(),
}))

vi.mock('../../hooks/useAdminConfig', () => ({
  useAdminConfig: vi.fn(),
}))

describe('AdminOverview Component', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useAdminSubmissions).mockReturnValue({
      submissions: [],
      pendingTotal: 0,
      isFetching: false,
    } as any)

    vi.mocked(useAdminClaims).mockReturnValue({
      claims: [],
      isFetching: false,
    } as any)

    vi.mocked(useAdminMembers).mockReturnValue({
      members: [],
      isFetching: false,
    } as any)

    vi.mocked(useAdminConfig).mockReturnValue({
      config: {
        redemption_start_day: 24,
        redemption_end_day: 26,
        payout_day: 24,
        is_open: true,
        conversion_rate: 100,
      },
      isFetching: false,
    } as any)
  })

  afterEach(() => {
    cleanup()
  })

  it('renders overview header and essential metrics cards in clean state', () => {
    const onNavigateTab = vi.fn()
    render(<AdminOverview onNavigateTab={onNavigateTab} />)

    expect(screen.getByText('Ringkasan Operasional Harian')).toBeInTheDocument()
    expect(screen.getByText('Semua Antrean Bersih & Terkendali')).toBeInTheDocument()
    expect(screen.getByText('Antrean Verifikasi Bukti Tugas')).toBeInTheDocument()
    expect(screen.getByText('Tidak Ada Antrean Verifikasi')).toBeInTheDocument()
    expect(screen.getByText('Permintaan Pencairan')).toBeInTheDocument()
    expect(screen.getByText('Status Anggota')).toBeInTheDocument()
    expect(screen.getByText('Jadwal Pencairan')).toBeInTheDocument()
  })

  it('renders action queue items with CTAs when pending items exist and triggers tab navigation', () => {
    vi.mocked(useAdminSubmissions).mockReturnValue({
      submissions: [
        {
          id: 101,
          task_title: 'Membaca Buku 15 Menit',
          user_name: 'Adit',
          status: 'PENDING',
        },
      ],
      pendingTotal: 1,
      isFetching: false,
    } as any)

    vi.mocked(useAdminClaims).mockReturnValue({
      claims: [
        {
          id: 201,
          coins_requested: 500,
          status: 'PENDING',
        },
      ],
      isFetching: false,
    } as any)

    vi.mocked(useAdminMembers).mockReturnValue({
      members: [
        {
          uid: 'user-1',
          explorer_name: 'Budi',
          is_active: false,
          inactive_days: 10,
          earned_coins_this_month: 3320,
          effective_earning_cap: 3320,
        },
      ],
      isFetching: false,
    } as any)

    const onNavigateTab = vi.fn()
    render(<AdminOverview onNavigateTab={onNavigateTab} />)

    // Action Queue items
    expect(screen.getByText(/1 Bukti Tugas Menunggu/i)).toBeInTheDocument()
    expect(screen.getByText(/1 Permintaan Pencairan/i)).toBeInTheDocument()
    expect(screen.getByText(/Perhatian Anggota/i)).toBeInTheDocument()

    // Click "Periksa Tugas" CTA
    const verifyCta = screen.getByRole('button', { name: /Periksa Tugas/i })
    fireEvent.click(verifyCta)
    expect(onNavigateTab).toHaveBeenCalledWith('submissions')

    // Click "Proses Pencairan" CTA
    const claimsCta = screen.getByRole('button', { name: /Proses Pencairan/i })
    fireEvent.click(claimsCta)
    expect(onNavigateTab).toHaveBeenCalledWith('claims')

    // Click "Lihat Anggota" CTA
    const membersCta = screen.getByRole('button', { name: /Lihat Anggota/i })
    fireEvent.click(membersCta)
    expect(onNavigateTab).toHaveBeenCalledWith('members')
  })

  it('clicks quick action shortcuts to navigate to Tasks, Rewards, and Settings', () => {
    const onNavigateTab = vi.fn()
    render(<AdminOverview onNavigateTab={onNavigateTab} />)

    const tasksShortcut = screen.getByRole('button', { name: /Jadwal Tugas/i })
    fireEvent.click(tasksShortcut)
    expect(onNavigateTab).toHaveBeenCalledWith('tasks')

    const rewardsShortcut = screen.getByRole('button', { name: /Katalog Hadiah/i })
    fireEvent.click(rewardsShortcut)
    expect(onNavigateTab).toHaveBeenCalledWith('rewards')

    const settingsShortcut = screen.getByRole('button', { name: /Pengaturan Ekonomi/i })
    fireEvent.click(settingsShortcut)
    expect(onNavigateTab).toHaveBeenCalledWith('settings')
  })
})
