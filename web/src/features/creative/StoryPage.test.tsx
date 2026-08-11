// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor, cleanup } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { StoryPage } from './StoryPage'
import { creativeApi } from '../../shared/lib/api'
import { useSession } from '../../shared/hooks/useSession'

vi.mock('../../shared/lib/api', () => ({
  apiClient: {
    get: vi.fn(),
  },
  creativeApi: {
    get: vi.fn(),
  },
}))

vi.mock('../../shared/hooks/useSession', () => ({
  useSession: vi.fn(),
}))

const mockedCreativeApiGet = vi.mocked(creativeApi.get)

describe('StoryPage', () => {
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
    mockedCreativeApiGet.mockImplementation(() => new Promise(() => {}))

    render(
      <MemoryRouter initialEntries={['/stories/1']}>
        <Routes>
          <Route path="/stories/:id" element={<StoryPage />} />
        </Routes>
      </MemoryRouter>
    )

    expect(screen.getByText('Loading story…')).toBeInTheDocument()
  })

  it('loads and displays a drawing submission', async () => {
    mockedCreativeApiGet.mockResolvedValueOnce({
      id: 1,
      quest_id: 1,
      kind: 'DRAWING',
      content: '<svg xmlns="http://www.w3.org/2000/svg"></svg>',
      created_at: '2024-01-01T00:00:00Z',
      author_uid: 'u1',
      crew_id: 'crew-1',
      challenge_id: 1,
      status: 'APPROVED',
    })

    render(
      <MemoryRouter initialEntries={['/stories/1']}>
        <Routes>
          <Route path="/stories/:id" element={<StoryPage />} />
        </Routes>
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Quest #1 · Drawing')).toBeInTheDocument()
    })
    expect(screen.getByText('by u1')).toBeInTheDocument()
  })

  it('renders error state when API fails', async () => {
    mockedCreativeApiGet.mockRejectedValueOnce(new Error('network error'))

    render(
      <MemoryRouter initialEntries={['/stories/1']}>
        <Routes>
          <Route path="/stories/:id" element={<StoryPage />} />
        </Routes>
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('network error')).toBeInTheDocument()
    })
    expect(screen.getByRole('link', { name: 'Back to Gallery' })).toBeInTheDocument()
  })

  it('renders not-drawing error when submission kind is not DRAWING', async () => {
    mockedCreativeApiGet.mockResolvedValueOnce({
      id: 1,
      quest_id: 1,
      kind: 'COMIC',
      content: '{}',
      created_at: '2024-01-01T00:00:00Z',
      author_uid: 'u1',
      crew_id: 'crew-1',
      challenge_id: 1,
      status: 'APPROVED',
    })

    render(
      <MemoryRouter initialEntries={['/stories/1']}>
        <Routes>
          <Route path="/stories/:id" element={<StoryPage />} />
        </Routes>
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('This submission is not a drawing.')).toBeInTheDocument()
    })
  })

  it('renders not-found error when submission is missing', async () => {
    mockedCreativeApiGet.mockResolvedValueOnce(null)

    render(
      <MemoryRouter initialEntries={['/stories/1']}>
        <Routes>
          <Route path="/stories/:id" element={<StoryPage />} />
        </Routes>
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Story not found')).toBeInTheDocument()
    })
  })
})
