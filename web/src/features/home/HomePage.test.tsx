// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor, cleanup } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { HomePage } from './HomePage'
import { apiClient } from '../../shared/lib/api'
import { useSession } from '../../shared/hooks/useSession'

// Mock dependencies
vi.mock('../../shared/lib/api', () => ({
  apiClient: {
    request: vi.fn(),
    get: vi.fn().mockResolvedValue({
      id: 1, title: 'Test', question: 'Q?', type: 'MCQ', options: ['A'], completed: false, xp_reward: 10
    }),
    post: vi.fn(),
  },
  chestsApi: {
    open: vi.fn(),
  },
  crewsApi: {
    get: vi.fn(),
  },
}))

vi.mock('../../shared/hooks/useSession', () => ({
  useSession: vi.fn(),
}))

// Mock ConnectedReactionBar to easily inspect its props
vi.mock('../../shared/components/molecules/ConnectedReactionBar', () => ({
  ConnectedReactionBar: ({ targetType, targetId }: { targetType: string, targetId: number }) => (
    <div data-testid={`reaction-bar-${targetType}-${targetId}`}>
      Mock Reaction Bar
    </div>
  )
}))

describe('HomePage', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.mocked(useSession).mockReturnValue({
      session: { uid: 'u1' } as any,
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

  it('renders correctly with Aktivitas Hari Ini and Misi Berikutnya sections', async () => {
    vi.mocked(apiClient.request).mockResolvedValueOnce({
      player: { explorer_name: 'Tester', coins: 10 },
      realm_progress: [{ realm: 'Whispering Woods', progress: 50, status: 'ACTIVE' }],
      daily_turn: { available: true, streak_days: 1 },
      active_quests: [
        { id: 201, title: 'Shadow Trail', challenge_count: 2, completed_count: 1, status: 'ACTIVE', quest_type: 'SOLO' },
      ],
      completed_quests_today: [],
      available_chests: [],
    })

    render(
      <MemoryRouter>
        <HomePage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Misi Berikutnya')).toBeInTheDocument()
      expect(screen.getByText('Shadow Trail')).toBeInTheDocument()
    })
  })

  it('renders onboarding immediately on first visit even before home loads', async () => {
    // delay mock to simulate slow load
    vi.mocked(apiClient.request).mockImplementationOnce(() => new Promise((resolve) => setTimeout(() => resolve({} as any), 1000)))
    localStorage.removeItem('odyssey_onboarded')
    
    render(
      <MemoryRouter>
        <HomePage />
      </MemoryRouter>
    )

    // Onboarding modal shows up immediately
    expect(screen.getByText(/Odyssey membantu kamu belajar/i)).toBeInTheDocument()
    // Home loader shows up under it
    expect(screen.getByText('Memuat dunia...')).toBeInTheDocument()
  })
})
