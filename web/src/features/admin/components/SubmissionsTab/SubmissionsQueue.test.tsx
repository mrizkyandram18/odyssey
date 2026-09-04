// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, cleanup, waitFor, fireEvent } from '@testing-library/react'
import { SubmissionsQueue } from './SubmissionsQueue'
import { adminTasksApi } from '../../../../shared/lib/api'

vi.mock('../../../../shared/lib/api', () => ({
  adminTasksApi: {
    getSubmissions: vi.fn(),
    verifySubmission: vi.fn(),
    editSubmission: vi.fn(),
  },
}))

const sub = (id: number, status = 'PENDING') => ({
  id,
  task_id: 100 + id,
  task_title: `Tugas ${id}`,
  task_type: 'TEXT_RESPONSE',
  user_uid: `user-${id}`,
  user_name: `User ${id}`,
  submission_type: 'MANUAL_VERIFY',
  status,
  payload: { text: `jawaban ${id}` },
  created_at: new Date().toISOString(),
  reward_coins: 50,
  reward_xp: 100,
})

const paged = (items: any[], page: number, total: number, hasNext: boolean, limit = 50) => ({
  items,
  pagination: { page, limit, total, has_next: hasNext },
})

describe('SubmissionsQueue pagination', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    cleanup()
  })

  it('renders the exact backend total, not submissions.length', async () => {
    vi.mocked(adminTasksApi.getSubmissions).mockResolvedValue(
      paged([sub(1), sub(2)], 1, 1234, true)
    )
    render(<SubmissionsQueue />)
    await waitFor(() => {
      expect(screen.getByText('Antrean Verifikasi Bukti Tugas (1234 Menunggu / 1234 Total)')).toBeInTheDocument()
    })
    // Only the current page rows are rendered — never the full 1234.
    expect(screen.getByText('Tugas 1')).toBeInTheDocument()
    expect(screen.getByText('Tugas 2')).toBeInTheDocument()
    expect(screen.queryByText('Tugas 3')).toBeNull()
  })

  it('renders numbered pages and fetches only the clicked page', async () => {
    const calls: any[] = []
    vi.mocked(adminTasksApi.getSubmissions).mockImplementation(async (params: any) => {
      calls.push(params)
      if (params?.page === 2) {
        return paged([sub(51), sub(52)], 2, 1234, true)
      }
      return paged([sub(1), sub(2)], 1, 1234, true)
    })
    render(<SubmissionsQueue />)
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Halaman 2' })).toBeInTheDocument()
    })
    // Current page is clearly indicated.
    expect(screen.getByRole('button', { name: 'Halaman 1' })).toHaveAttribute('aria-current', 'page')

    fireEvent.click(screen.getByRole('button', { name: 'Halaman 2' }))
    await waitFor(() => {
      expect(screen.getByText('Tugas 51')).toBeInTheDocument()
    })
    const page2Calls = calls.filter((c) => c?.page === 2)
    expect(page2Calls.length).toBeGreaterThanOrEqual(1)
    expect(page2Calls[0]).toMatchObject({ status: 'PENDING', page: 2, limit: 50 })
  })

  it('changing filter resets to page 1', async () => {
    const calls: any[] = []
    vi.mocked(adminTasksApi.getSubmissions).mockImplementation(async (params: any) => {
      calls.push(params)
      if (params?.status === 'APPROVED') {
        return paged([sub(9, 'APPROVED')], 1, 5, false)
      }
      if (params?.status === 'PENDING' && params?.limit === 1) {
        return paged([], 1, 12, false, 1)
      }
      return paged([sub(1)], 1, 12, false)
    })
    render(<SubmissionsQueue />)
    await waitFor(() => {
      expect(screen.getByText('Tugas 1')).toBeInTheDocument()
    })
    fireEvent.click(screen.getByRole('button', { name: 'Disetujui' }))
    await waitFor(() => {
      expect(screen.getByText('Tugas 9')).toBeInTheDocument()
    })
    const approvedCalls = calls.filter((c) => c?.status === 'APPROVED')
    expect(approvedCalls.length).toBeGreaterThanOrEqual(1)
    // Every APPROVED fetch must be page 1 — filter change never stays on an old page.
    for (const c of approvedCalls) {
      expect(c.page).toBe(1)
    }
  })

  it('shows the exact global Menunggu count when filter is Semua', async () => {
    vi.mocked(adminTasksApi.getSubmissions).mockImplementation(async (params: any) => {
      if (params?.status === 'PENDING' && params?.limit === 1) {
        return paged([], 1, 7, false, 1)
      }
      if (!params?.status) {
        return paged([sub(1), sub(2, 'APPROVED')], 1, 100, true)
      }
      return paged([sub(1)], 1, 7, false)
    })
    render(<SubmissionsQueue />)
    await waitFor(() => {
      expect(screen.getByText('Tugas 1')).toBeInTheDocument()
    })
    fireEvent.click(screen.getByRole('button', { name: 'Semua' }))
    await waitFor(() => {
      // Menunggu = exact global pending total (7), Total = exact filter total (100).
      expect(screen.getByText('Antrean Verifikasi Bukti Tugas (7 Menunggu / 100 Total)')).toBeInTheDocument()
    })
  })
})
