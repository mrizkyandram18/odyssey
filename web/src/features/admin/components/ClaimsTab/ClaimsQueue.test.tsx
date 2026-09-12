// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, cleanup, waitFor, fireEvent } from '@testing-library/react'
import { ClaimsQueue } from './ClaimsQueue'
import { adminTasksApi } from '../../../../shared/lib/api'

vi.mock('../../../../shared/lib/api', () => ({
  adminTasksApi: {
    getClaims: vi.fn(),
    processClaim: vi.fn(),
  },
}))

const claim = (id: number, status = 'PENDING') => ({
  id,
  status,
  target_type: 'EWALLET',
  target_value: `0812000000${id}`,
  user_name: `User ${id}`,
  created_at: new Date().toISOString(),
  coins_redeemed: 500,
})

const paged = (items: any[], page: number, total: number, hasNext: boolean, limit = 50) => ({
  items,
  pagination: { page, limit, total, has_next: hasNext },
})

describe('ClaimsQueue total (P0: backend total, never page length)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    cleanup()
  })

  it('renders the exact backend total, not claims.length', async () => {
    vi.mocked(adminTasksApi.getClaims).mockResolvedValue(
      paged([claim(1), claim(2)], 1, 1234, true)
    )
    render(<ClaimsQueue />)
    await waitFor(() => {
      expect(screen.getByText('Permintaan Pencairan Koin (1234 Menunggu / 1234 Total)')).toBeInTheDocument()
    })
    // Only the current page rows are rendered — never the full 1234.
    expect(screen.getByText(/User 1/)).toBeInTheDocument()
    expect(screen.getByText(/User 2/)).toBeInTheDocument()
  })

  it('keeps Menunggu count and Total in sync with backend pagination', async () => {
    vi.mocked(adminTasksApi.getClaims).mockResolvedValue(
      paged([claim(1), claim(2)], 1, 3, true)
    )
    render(<ClaimsQueue />)
    await waitFor(() => {
      expect(screen.getByText('Permintaan Pencairan Koin (3 Menunggu / 3 Total)')).toBeInTheDocument()
    })
  })
})

describe('ClaimsQueue terminology (display labels never change backend enum)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    cleanup()
  })

  it('shows canonical Menunggu badge while fetching with PENDING enum', async () => {
    vi.mocked(adminTasksApi.getClaims).mockResolvedValue(paged([claim(1)], 1, 1, false))
    render(<ClaimsQueue />)
    await waitFor(() => {
      // Canonical label appears both on the filter pill and the card status badge.
      expect(screen.getAllByText('Menunggu')).toHaveLength(2)
    })
    expect(adminTasksApi.getClaims).toHaveBeenCalledWith(
      expect.objectContaining({ status: 'PENDING' })
    )
  })

  it('shows Ditolak filter while fetching with REJECTED enum', async () => {
    vi.mocked(adminTasksApi.getClaims).mockImplementation(async (params: any) => {
      if (params?.status === 'REJECTED') return paged([claim(9, 'REJECTED')], 1, 1, false)
      if (params?.status === 'PENDING' && params?.limit === 1) return paged([], 1, 0, false, 1)
      return paged([claim(1)], 1, 1, false)
    })
    render(<ClaimsQueue />)
    await waitFor(() => {
      expect(screen.getByText(/User 1/)).toBeInTheDocument()
    })
    fireEvent.click(screen.getByRole('button', { name: 'Ditolak' }))
    await waitFor(() => {
      expect(screen.getByText(/User 9/)).toBeInTheDocument()
      // Canonical label appears both on the filter pill and the card status badge.
      expect(screen.getAllByText('Ditolak')).toHaveLength(2)
    })
    const rejectedCalls = vi.mocked(adminTasksApi.getClaims).mock.calls.filter(
      (c) => (c[0] as any)?.status === 'REJECTED'
    )
    expect(rejectedCalls.length).toBeGreaterThanOrEqual(1)
  })

  it('sends APPROVED enum when admin confirms Sudah Ditransfer', async () => {
    vi.mocked(adminTasksApi.getClaims).mockResolvedValue(paged([claim(1)], 1, 1, false))
    vi.mocked(adminTasksApi.processClaim).mockResolvedValue({ success: true, status: 'APPROVED' } as any)
    render(<ClaimsQueue />)
    await waitFor(() => {
      expect(screen.getByText(/User 1/)).toBeInTheDocument()
    })
    fireEvent.click(screen.getByRole('button', { name: /Selesaikan pencairan/ }))
    await waitFor(() => {
      expect(adminTasksApi.processClaim).toHaveBeenCalledWith(1, 'APPROVED', undefined)
    })
  })
})
