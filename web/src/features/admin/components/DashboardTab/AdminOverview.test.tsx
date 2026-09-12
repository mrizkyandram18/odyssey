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

  it('renders overview header with action queue, concise schedule strip, and shortcuts', () => {
    const onNavigateTab = vi.fn()
    render(<AdminOverview onNavigateTab={onNavigateTab} />)

    expect(screen.getByText('Ringkasan Operasional Harian')).toBeInTheDocument()
    expect(screen.getByText('Semua Antrean Bersih & Terkendali')).toBeInTheDocument()
    // Concise schedule status replaces the old duplicated metric cards.
    expect(screen.getByTestId('admin-schedule-strip')).toBeInTheDocument()
    expect(screen.getByText(/Jadwal Pencairan:/)).toBeInTheDocument()
    // Duplicated metric cards must be gone.
    expect(screen.queryByText('Metrik Operasional Utama')).toBeNull()
    expect(screen.queryByText('Antrean Verifikasi Bukti Tugas')).toBeNull()
    expect(screen.queryByText('Status Anggota')).toBeNull()
    // Shortcuts stay.
    expect(screen.getByRole('button', { name: /Jadwal Tugas/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Katalog Hadiah/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Pengaturan Ekonomi/i })).toBeInTheDocument()
  })

  it('shows each pending queue exactly once (no metric/action duplication)', () => {
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

    const onNavigateTab = vi.fn()
    render(<AdminOverview onNavigateTab={onNavigateTab} />)

    expect(screen.getAllByText(/1 Bukti Tugas Menunggu/i)).toHaveLength(1)
    expect(screen.getAllByText(/1 Permintaan Pencairan/i)).toHaveLength(1)
    expect(screen.queryByText('Metrik Operasional Utama')).toBeNull()
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

    it('renders admin announcement preview at the bottom when config.announcement.visible is true', () => {
     const onNavigateTab = vi.fn()
     const announcementConfig = {
       enabled: true,
       title: 'Pengumuman Pencairan',
       body: 'Pencairan sedang dalam proses verifikasi.',
       audience: 'ALL',
       priority: 'normal',
       visible: true,
       start_at: '2026-09-12T00:00:00+07:00',
       end_at: '2026-09-30T23:59:59+07:00',
     }
     vi.mocked(useAdminConfig).mockReturnValue({
       config: {
         redemption_start_day: 24,
         redemption_end_day: 26,
         payout_day: 24,
         is_open: true,
         conversion_rate: 100,
         announcement: announcementConfig,
       },
       isFetching: false,
     } as any)

      render(<AdminOverview onNavigateTab={onNavigateTab} />)
      // Compact admin preview (not the member-facing banner) with a CTA to settings.
      expect(screen.getByTestId('admin-announcement-preview')).toBeTruthy()
      expect(screen.getByTestId('admin-announcement-title').textContent).toBe('Pengumuman Pencairan')
      expect(screen.queryByTestId('announcement-banner')).toBeNull()
      expect(screen.queryByTestId('announcement-dismiss')).toBeNull()
      const editCta = screen.getByRole('button', { name: /Ubah di Pengaturan/i })
      fireEvent.click(editCta)
      expect(onNavigateTab).toHaveBeenCalledWith('settings')
    })

    it('does not render admin announcement preview when config.announcement is undefined or not visible', () => {
      const onNavigateTab = vi.fn()
      render(<AdminOverview onNavigateTab={onNavigateTab} />)
      expect(screen.queryByTestId('admin-announcement-preview')).toBeNull()
      expect(screen.queryByTestId('announcement-banner')).toBeNull()
    })
})
