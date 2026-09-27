// @vitest-environment jsdom
import React from 'react'
import '@testing-library/jest-dom/vitest'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { HomePage } from './HomePage'
import { tasksApi } from '../../shared/lib/api'
import { useSession } from '../../shared/hooks/useSession'

// Mock dependencies
vi.mock('../../shared/lib/api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../shared/lib/api')>()
  return {
    ...actual,
    tasksApi: {
      getToday: vi.fn(),
      submit: vi.fn(),
    },
    apiClient: {
      get: vi.fn(),
      post: vi.fn(),
      request: vi.fn(),
    },
  }
})

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

    expect(await screen.findByText('Tugas Harian', undefined, { timeout: 3000 })).toBeInTheDocument()
    // "Belajar Menabung" now appears in both primary CTA pill and task list — allow multiple
    expect((await screen.findAllByText('Belajar Menabung', undefined, { timeout: 3000 })).length).toBeGreaterThan(0)
    // streak may appear in progress card + anchor — allow multiple
    expect((await screen.findAllByText('3 Hari', undefined, { timeout: 3000 })).length).toBeGreaterThan(0)
    expect((await screen.findAllByText(/150/, undefined, { timeout: 3000 })).length).toBeGreaterThan(0)
  })

  it('renders all-done closure with milestone, streak explainer, and level meaning', async () => {
    vi.mocked(useSession).mockReturnValue({
      session: { uid: 'u1' } as any,
      profile: {
        uid: 'u1',
        explorer_name: 'Tester',
        level: 3,
        xp: 500,
        coins: 200,
        streak_days: 7,
      } as any,
      loading: false,
      error: null,
      login: vi.fn(),
      logout: vi.fn(),
      refreshProfile: vi.fn(),
    })
    vi.mocked(tasksApi.getToday).mockResolvedValueOnce({
      tasks: [1, 2, 3].map((n) => ({
        id: n,
        title: `Tugas ${n}`,
        description: `Deskripsi ${n}`,
        task_type: 'QUIZ',
        step_order: n,
        reward_coins: 40,
        reward_xp: 100,
        config: {},
        is_locked: false,
        status: 'APPROVED',
        coins_earned: 40,
        xp_earned: 100,
      })),
    })

    render(
      <MemoryRouter>
        <HomePage />
      </MemoryRouter>
    )

    expect(await screen.findByText(/Misi Sempurna/, undefined, { timeout: 3000 })).toBeInTheDocument()
    // earnedToday sums actual credited coins/xp: 3×40 / 3×100
    expect((await screen.findAllByText(/\+120 Koin/, undefined, { timeout: 3000 })).length).toBeGreaterThan(0)
    expect((await screen.findAllByText(/7 hari berturut-turut/, undefined, { timeout: 3000 })).length).toBeGreaterThan(0)
    expect(screen.getByText('Cara kerja streak')).toBeInTheDocument()
    expect(screen.getByText(/Streak bertambah setiap ada tugas yang disetujui/)).toBeInTheDocument()
    expect((await screen.findAllByText(/Bintang \(XP\) menentukan tingkatmu/, undefined, { timeout: 3000 })).length).toBeGreaterThan(0)
  })
})
