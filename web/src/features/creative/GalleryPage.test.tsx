// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor, cleanup } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { GalleryPage } from './GalleryPage'
import { apiClient } from '../../shared/lib/api'
import { useSession } from '../../shared/hooks/useSession'

vi.mock('../../shared/lib/api', () => ({
  apiClient: {
    get: vi.fn(),
  },
}))

vi.mock('../../shared/hooks/useSession', () => ({
  useSession: vi.fn(),
}))

vi.mock('../../shared/components/molecules/CreativeCard', () => ({
  CreativeCard: ({ submission }: { submission: { id: number } }) => (
    <div data-testid={`creative-card-${submission.id}`}>CreativeCard</div>
  ),
}))

describe('GalleryPage', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.mocked(useSession).mockReturnValue({
      session: { uid: 'u1', crew_id: 'crew-1' } as any,
      loading: false,
      error: null,
      login: vi.fn(),
      logout: vi.fn(),
      refreshProfile: vi.fn(),
    })
  })

  afterEach(() => {
    cleanup()
  })

  it('renders loading state initially', async () => {
    vi.mocked(apiClient.get).mockImplementation(() => new Promise(() => {}))

    render(
      <MemoryRouter>
        <GalleryPage />
      </MemoryRouter>
    )

    expect(screen.getByText('Memuat galeri…')).toBeInTheDocument()
  })

  it('loads and displays submissions with All filter', async () => {
    vi.mocked(apiClient.get).mockResolvedValueOnce([
      { id: 1, kind: 'STORY', content: 's1', created_at: '2024-01-02T00:00:00Z' },
      { id: 2, kind: 'COMIC', content: 'c1', created_at: '2024-01-01T00:00:00Z' },
    ])

    render(
      <MemoryRouter>
        <GalleryPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByTestId('creative-card-1')).toBeInTheDocument()
    })
    expect(screen.getByTestId('creative-card-2')).toBeInTheDocument()
    expect(screen.getByText('Galeri Keluarga')).toBeInTheDocument()
  })

  it('filters submissions by kind when filter is selected', async () => {
    vi.mocked(apiClient.get)
      .mockResolvedValueOnce([
        { id: 1, kind: 'STORY', content: 's1', created_at: '2024-01-02T00:00:00Z' },
        { id: 2, kind: 'COMIC', content: 'c1', created_at: '2024-01-01T00:00:00Z' },
      ])
      .mockResolvedValueOnce([
        { id: 2, kind: 'COMIC', content: 'c1', created_at: '2024-01-01T00:00:00Z' },
      ])

    render(
      <MemoryRouter>
        <GalleryPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByTestId('creative-card-1')).toBeInTheDocument()
    })

    const comicButton = screen.getByRole('button', { name: 'Komik' })
    comicButton.click()

    await waitFor(() => {
      expect(apiClient.get).toHaveBeenNthCalledWith(2, '/api/creative?kind=COMIC')
    })

    expect(screen.queryByTestId('creative-card-1')).not.toBeInTheDocument()
    expect(screen.getByTestId('creative-card-2')).toBeInTheDocument()
  })

  it('renders error state with retry button', async () => {
    vi.mocked(apiClient.get).mockRejectedValueOnce(new Error('network error'))

    render(
      <MemoryRouter>
        <GalleryPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('network error')).toBeInTheDocument()
    })
    expect(screen.getByRole('button', { name: 'Coba Lagi' })).toBeInTheDocument()
  })

  it('renders empty state when no submissions exist', async () => {
    vi.mocked(apiClient.get).mockResolvedValueOnce([])

    render(
      <MemoryRouter>
        <GalleryPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Belum ada karya.')).toBeInTheDocument()
    })
  })

  it('calls API with correct params on initial load', async () => {
    vi.mocked(apiClient.get).mockResolvedValueOnce([])

    render(
      <MemoryRouter>
        <GalleryPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(apiClient.get).toHaveBeenCalledWith('/api/creative')
    })
  })
})
