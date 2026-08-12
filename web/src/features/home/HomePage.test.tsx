// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor, cleanup } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { HomePage } from './HomePage'
import { apiClient, crewsApi } from '../../shared/lib/api'
import { useSession } from '../../shared/hooks/useSession'

// Mock dependencies
vi.mock('../../shared/lib/api', () => ({
  apiClient: {
    request: vi.fn(),
    get: vi.fn().mockResolvedValue({ data: {
      id: 1, title: 'Test', question: 'Q?', type: 'MCQ', options: ['A'], completed: false, xp_reward: 10
    } }),
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
    vi.mocked(apiClient.get).mockResolvedValue({
      data: {
        id: 1, title: 'Test', question: 'Q?', type: 'MCQ', options: ['A'], completed: false, xp_reward: 10
      }
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

  it('renders Your Turn badge on an active relay quest assigned to me', async () => {
    vi.mocked(apiClient.request).mockResolvedValueOnce({
      player: { explorer_name: 'Tester', coins: 10 },
      realm_progress: [{ realm: 'Whispering Woods', progress: 50, status: 'ACTIVE' }],
      daily_turn: { available: true, streak_days: 1 },
      active_quests: [
        { id: 201, title: 'Shadow Trail', challenge_count: 2, completed_count: 1, status: 'ACTIVE', quest_type: 'RELAY', active_challenge_assigned_to: 'u1' },
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
      expect(screen.getByText('Your Turn')).toBeInTheDocument()
    })
  })

  it('does not render Your Turn badge when the relay turn belongs to someone else', async () => {
    vi.mocked(apiClient.request).mockResolvedValueOnce({
      player: { explorer_name: 'Tester', coins: 10 },
      realm_progress: [{ realm: 'Whispering Woods', progress: 50, status: 'ACTIVE' }],
      daily_turn: { available: true, streak_days: 1 },
      active_quests: [
        { id: 201, title: 'Shadow Trail', challenge_count: 2, completed_count: 1, status: 'ACTIVE', quest_type: 'RELAY', active_challenge_assigned_to: 'u2' },
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
      expect(screen.getByText('Shadow Trail')).toBeInTheDocument()
    })
    expect(screen.queryByText('Your Turn')).not.toBeInTheDocument()
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

  it('renders crew streak as shared progress', async () => {
    vi.mocked(apiClient.request).mockResolvedValueOnce({
      player: { explorer_name: 'Tester', coins: 10 },
      realm_progress: [{ realm: 'Whispering Woods', progress: 50, status: 'ACTIVE' }],
      daily_turn: { available: true, streak_days: 1, crew_streak: 4 },
      active_quests: [],
      completed_quests_today: [],
      available_chests: [],
    })

    render(
      <MemoryRouter>
        <HomePage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByTestId('home-crew-streak')).toHaveTextContent('Runtutan keluarga: 4 hari bersama')
    })
  })

  it('falls back to 0 crew streak when field is missing', async () => {
    vi.mocked(apiClient.request).mockResolvedValueOnce({
      player: { explorer_name: 'Tester', coins: 10 },
      realm_progress: [{ realm: 'Whispering Woods', progress: 50, status: 'ACTIVE' }],
      daily_turn: { available: true, streak_days: 1 },
      active_quests: [],
      completed_quests_today: [],
      available_chests: [],
    })

    render(
      <MemoryRouter>
        <HomePage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByTestId('home-crew-streak')).toHaveTextContent('Runtutan keluarga: 0 hari bersama')
    })
  })

  it('renders crew banner when crew has a banner_url', async () => {
    vi.mocked(apiClient.request).mockResolvedValueOnce({
      player: { explorer_name: 'Tester', coins: 10 },
      realm_progress: [{ realm: 'Whispering Woods', progress: 50, status: 'ACTIVE' }],
      daily_turn: { available: true, streak_days: 1 },
      active_quests: [],
      completed_quests_today: [],
      available_chests: [],
    })
    vi.mocked(crewsApi.get).mockResolvedValueOnce({
      id: 'crew-1',
      name: 'Test Crew',
      banner_url: 'https://example.com/banner.png',
      theme: 'default',
      created_at: '',
      updated_at: '',
    })

    render(
      <MemoryRouter>
        <HomePage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByTestId('crew-banner')).toBeInTheDocument()
    })
  })
})
