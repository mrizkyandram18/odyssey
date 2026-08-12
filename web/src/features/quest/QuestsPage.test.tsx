// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor, cleanup } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QuestsPage } from './QuestsPage'
import { questsApi, realmProgressApi } from '../../shared/lib/api'

vi.mock('../../shared/lib/api', () => ({
  questsApi: { list: vi.fn(), available: vi.fn() },
  realmProgressApi: { list: vi.fn() },
}))

vi.mock('../../shared/hooks/useSession', () => ({
  useSession: () => ({ session: { uid: 'u1' } }),
}))

describe('QuestsPage', () => {
  beforeEach(() => {
    vi.resetAllMocks()
  })
  afterEach(() => {
    cleanup()
  })

  it('filters pending quests by available definitions to show in Misi Tersedia', async () => {
    // Database instances
    vi.mocked(questsApi.list).mockResolvedValueOnce([
      { id: 10, title: 'Active Quest', template_slug: 'q1', status: 'ACTIVE' } as any,
      { id: 20, title: 'Pending Instantiated Quest', template_slug: 'q2', status: 'PENDING' } as any,
      { id: 30, title: 'Pending Future Quest', template_slug: 'q3', status: 'PENDING' } as any,
    ])
    // Definitions available
    vi.mocked(questsApi.available).mockResolvedValueOnce([
      { template_slug: 'q1' } as any,
      { template_slug: 'q2' } as any,
    ])
    vi.mocked(realmProgressApi.list).mockResolvedValueOnce([])

    render(
      <MemoryRouter>
        <QuestsPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      // Should show Active Quest in Misi Aktif
      expect(screen.getByText('Active Quest')).toBeInTheDocument()
      
      // Should show Pending Instantiated Quest because it matches available definition q2
      expect(screen.getByText('Pending Instantiated Quest')).toBeInTheDocument()

      // Should NOT show Pending Future Quest because q3 is not available yet
      expect(screen.queryByText('Pending Future Quest')).not.toBeInTheDocument()
    })
  })
})
