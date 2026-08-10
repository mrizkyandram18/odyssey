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
    post: vi.fn(),
  },
  chestsApi: {
    open: vi.fn(),
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
      session: { user: { id: 'u1' } } as any,
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

  it('renders ConnectedReactionBar with correct props for completed quests', async () => {
    // Setup mock API response with a completed quest
    vi.mocked(apiClient.request).mockResolvedValueOnce({
      player: { explorer_name: 'Tester', coins: 10 },
      realm_progress: [{ realm: 'Whispering Woods', progress: 50, status: 'ACTIVE' }],
      daily_turn: { available: true, streak_days: 1 },
      active_quests: [],
      completed_quests_today: [
        { id: 101, title: 'Morning Light', challenge_count: 1, completed_count: 1, status: 'DONE' }
      ],
      available_chests: [],
    })

    render(
      <MemoryRouter>
        <HomePage />
      </MemoryRouter>
    )

    // Wait for the page to load
    await waitFor(() => {
      expect(screen.getByText('Jejak Hari Ini')).toBeInTheDocument()
    })

    // Assert the mocked ConnectedReactionBar is rendered with QUEST target and correct ID
    const reactionBar = screen.getByTestId('reaction-bar-QUEST-101')
    expect(reactionBar).toBeInTheDocument()
  })

  it('does not render reaction bar if there are no completed quests', async () => {
    // Setup mock API response with NO completed quests
    vi.mocked(apiClient.request).mockResolvedValueOnce({
      player: { explorer_name: 'Tester', coins: 10 },
      realm_progress: [{ realm: 'Whispering Woods', progress: 50, status: 'ACTIVE' }],
      daily_turn: { available: true, streak_days: 1 },
      active_quests: [],
      completed_quests_today: [], // Empty
      available_chests: [],
    })

    render(
      <MemoryRouter>
        <HomePage />
      </MemoryRouter>
    )

    // Wait for the title to be sure it has loaded
    await waitFor(() => {
      expect(screen.getByText('Halo, Tester!')).toBeInTheDocument()
    })

    // Assert that the 'Jejak Hari Ini' section is absent
    expect(screen.queryByText('Jejak Hari Ini')).not.toBeInTheDocument()

    // Assert no reaction bar is rendered
    expect(screen.queryByTestId(/reaction-bar/)).not.toBeInTheDocument()
  })
})
