// @vitest-environment jsdom
import React from 'react'
import '@testing-library/jest-dom/vitest'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor, cleanup } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { HomePage } from './HomePage'
import { tasksApi } from '../../shared/lib/api'
import { useSession } from '../../shared/hooks/useSession'

// Mock dependencies
vi.mock('../../shared/lib/api', () => ({
  tasksApi: {
    getToday: vi.fn(),
    submit: vi.fn(),
  },
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    request: vi.fn(),
  },
}))

vi.mock('../../shared/hooks/useSession', () => ({
  useSession: vi.fn(),
}))

describe('HomePage', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.mocked(useSession).mockReturnValue({
      session: { uid: 'u1' } as any,
      profile: {
        uid: 'u1',
        explorer_name: 'Tester',
        level: 2,
        xp: 250,
        coins: 150,
        streak_days: 3,
      } as any,
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

  it('renders linear task stepper with today tasks', async () => {
    vi.mocked(tasksApi.getToday).mockResolvedValueOnce({
      tasks: [
        {
          id: 1,
          title: 'Belajar Menabung',
          description: 'Nonton video kuis',
          task_type: 'VIDEO_QUIZ',
          step_order: 1,
          reward_coins: 50,
          reward_xp: 100,
          config: {},
          is_locked: false,
          status: 'UNLOCKED',
          coins_earned: 0,
          xp_earned: 0,
        },
      ],
    })

    render(
      <MemoryRouter>
        <HomePage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Petualangan Harian')).toBeInTheDocument()
      expect(screen.getByText('Belajar Menabung')).toBeInTheDocument()
      expect(screen.getByText('3 Hari')).toBeInTheDocument()
      expect(screen.getByText('150')).toBeInTheDocument()
    })
  })
})
