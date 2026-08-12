// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor, cleanup } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { MissionsPage } from './MissionsPage'
import { MissionsApi, JourneyProgressApi } from '../../shared/lib/api'

vi.mock('../../shared/lib/api', () => ({
  MissionsApi: { list: vi.fn(), available: vi.fn() },
  JourneyProgressApi: { list: vi.fn() },
}))

vi.mock('../../shared/hooks/useSession', () => ({
  useSession: () => ({ session: { uid: 'u1' } }),
}))

describe('MissionsPage', () => {
  beforeEach(() => {
    vi.resetAllMocks()
  })
  afterEach(() => {
    cleanup()
  })

  it('filters pending missions by available definitions to show in Misi Tersedia', async () => {
    // Database instances
    vi.mocked(MissionsApi.list).mockResolvedValueOnce([
      { id: 10, title: 'Active Mission', template_slug: 'q1', status: 'ACTIVE' } as any,
      { id: 20, title: 'Pending Instantiated Mission', template_slug: 'q2', status: 'PENDING' } as any,
      { id: 30, title: 'Pending Future Mission', template_slug: 'q3', status: 'PENDING' } as any,
    ])
    // Definitions available
    vi.mocked(MissionsApi.available).mockResolvedValueOnce([
      { template_slug: 'q1' } as any,
      { template_slug: 'q2' } as any,
    ])
    vi.mocked(JourneyProgressApi.list).mockResolvedValueOnce([])

    render(
      <MemoryRouter>
        <MissionsPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      // Should show Active Mission in Misi Aktif
      expect(screen.getByText('Active Mission')).toBeInTheDocument()
      
      // Should show Pending Instantiated Mission because it matches available definition q2
      expect(screen.getByText('Pending Instantiated Mission')).toBeInTheDocument()

      // Should NOT show Pending Future Mission because q3 is not available yet
      expect(screen.queryByText('Pending Future Mission')).not.toBeInTheDocument()
    })
  })
})
