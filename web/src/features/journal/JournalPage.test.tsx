// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor, cleanup } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { JournalPage } from './JournalPage'
import { achievementsApi, loreApi, storyFragmentsApi } from '../../shared/lib/api'

vi.mock('../../shared/lib/api', () => ({
  achievementsApi: { list: vi.fn() },
  loreApi: { list: vi.fn() },
  storyFragmentsApi: { list: vi.fn() },
}))

describe('JournalPage', () => {
  beforeEach(() => {
    vi.resetAllMocks()
  })
  afterEach(() => {
    cleanup()
  })

  it('handles null API responses gracefully without crashing and shows empty states', async () => {
    vi.mocked(achievementsApi.list).mockResolvedValueOnce(null as any)
    vi.mocked(loreApi.list).mockResolvedValueOnce(null as any)
    vi.mocked(storyFragmentsApi.list).mockResolvedValueOnce(null as any)

    render(
      <MemoryRouter>
        <JournalPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Belum ada catatan perjalanan. Mulailah misi untuk mengukir ceritamu!')).toBeInTheDocument()
    })
  })

  it('handles API errors gracefully', async () => {
    vi.mocked(achievementsApi.list).mockRejectedValueOnce(new Error('API failed'))
    vi.mocked(loreApi.list).mockResolvedValueOnce([])
    vi.mocked(storyFragmentsApi.list).mockResolvedValueOnce([])

    render(
      <MemoryRouter>
        <JournalPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Belum ada catatan perjalanan. Mulailah misi untuk mengukir ceritamu!')).toBeInTheDocument()
    })
  })
})
