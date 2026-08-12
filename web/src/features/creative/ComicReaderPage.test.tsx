// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor, cleanup } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { ComicReaderPage } from './ComicReaderPage'
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

describe('ComicReaderPage', () => {
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
      <MemoryRouter initialEntries={['/comics/1']}>
        <Routes>
          <Route path="/comics/:id" element={<ComicReaderPage />} />
        </Routes>
      </MemoryRouter>
    )

    expect(screen.getByText('Memuat komik…')).toBeInTheDocument()
  })

  it('loads and displays comic panels with navigation', async () => {
    const comicContent = JSON.stringify({
      v: 1,
      panels: [
        { caption: 'Panel 1', svg: '<svg></svg>' },
        { caption: 'Panel 2' },
        { caption: 'Panel 3', svg: '<svg/>' },
      ],
    })

    mockedCreativeApiGet.mockResolvedValueOnce({
      id: 1,
      quest_id: 1,
      kind: 'COMIC',
      content: comicContent,
      created_at: '2024-01-01T00:00:00Z',
      author_uid: 'u1',
      crew_id: 'crew-1',
      challenge_id: 1,
      status: 'APPROVED',
    })

    render(
      <MemoryRouter initialEntries={['/comics/1']}>
        <Routes>
          <Route path="/comics/:id" element={<ComicReaderPage />} />
        </Routes>
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Panel 1')).toBeInTheDocument()
    })

    expect(screen.getByText('Panel 1 dari 3')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Lanjut' })).toBeInTheDocument()

    const nextButton = screen.getByRole('button', { name: 'Lanjut' })
    nextButton.click()

    await waitFor(() => {
      expect(screen.getByText('Panel 2')).toBeInTheDocument()
    })
    expect(screen.getByText('Panel 2 dari 3')).toBeInTheDocument()
  })

  it('disables prev button on first panel and next on last panel', async () => {
    const comicContent = JSON.stringify({
      v: 1,
      panels: [{ caption: 'Only panel' }],
    })

    mockedCreativeApiGet.mockResolvedValueOnce({
      id: 1,
      quest_id: 1,
      kind: 'COMIC',
      content: comicContent,
      created_at: '2024-01-01T00:00:00Z',
      author_uid: 'u1',
      crew_id: 'crew-1',
      challenge_id: 1,
      status: 'APPROVED',
    })

    render(
      <MemoryRouter initialEntries={['/comics/1']}>
        <Routes>
          <Route path="/comics/:id" element={<ComicReaderPage />} />
        </Routes>
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Only panel')).toBeInTheDocument()
    })

    expect(screen.getByRole('button', { name: 'Sebel' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Lanjut' })).toBeDisabled()
  })

  it('renders error state when API fails', async () => {
    mockedCreativeApiGet.mockRejectedValueOnce(new Error('network error'))

    render(
      <MemoryRouter initialEntries={['/comics/1']}>
        <Routes>
          <Route path="/comics/:id" element={<ComicReaderPage />} />
        </Routes>
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('network error')).toBeInTheDocument()
    })
    expect(screen.getByRole('link', { name: 'Kembali ke Galeri' })).toBeInTheDocument()
  })

  it('renders not-comic error when submission kind is not COMIC', async () => {
    mockedCreativeApiGet.mockResolvedValueOnce({
      id: 1,
      quest_id: 1,
      kind: 'STORY',
      content: 'A story',
      created_at: '2024-01-01T00:00:00Z',
      author_uid: 'u1',
      crew_id: 'crew-1',
      challenge_id: 1,
      status: 'APPROVED',
    })

    render(
      <MemoryRouter initialEntries={['/comics/1']}>
        <Routes>
          <Route path="/comics/:id" element={<ComicReaderPage />} />
        </Routes>
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Karya ini bukan komik.')).toBeInTheDocument()
    })
  })

  it('renders invalid comic payload error', async () => {
    mockedCreativeApiGet.mockResolvedValueOnce({
      id: 1,
      quest_id: 1,
      kind: 'COMIC',
      content: 'not a json',
      created_at: '2024-01-01T00:00:00Z',
      author_uid: 'u1',
      crew_id: 'crew-1',
      challenge_id: 1,
      status: 'APPROVED',
    })

    render(
      <MemoryRouter initialEntries={['/comics/1']}>
        <Routes>
          <Route path="/comics/:id" element={<ComicReaderPage />} />
        </Routes>
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Komik tidak dapat ditampilkan.')).toBeInTheDocument()
    })
  })

  it('shows progress dots', async () => {
    const comicContent = JSON.stringify({
      v: 1,
      panels: [
        { caption: 'A' },
        { caption: 'B' },
        { caption: 'C' },
        { caption: 'D' },
      ],
    })

    mockedCreativeApiGet.mockResolvedValueOnce({
      id: 1,
      quest_id: 1,
      kind: 'COMIC',
      content: comicContent,
      created_at: '2024-01-01T00:00:00Z',
      author_uid: 'u1',
      crew_id: 'crew-1',
      challenge_id: 1,
      status: 'APPROVED',
    })

    render(
      <MemoryRouter initialEntries={['/comics/1']}>
        <Routes>
          <Route path="/comics/:id" element={<ComicReaderPage />} />
        </Routes>
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Panel 1 dari 4')).toBeInTheDocument()
    })

    const dots = screen.getAllByRole('generic').filter(el =>
      el.className.includes('rounded-full') && el.className.includes('h-2')
    )
    expect(dots).toHaveLength(4)
  })
})
