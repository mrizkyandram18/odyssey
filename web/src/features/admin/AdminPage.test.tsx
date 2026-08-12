/**
 * @vitest-environment jsdom
 */
import React from 'react'
import '@testing-library/jest-dom/vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { vi, describe, it, expect, afterEach } from 'vitest'
import { BrowserRouter } from 'react-router-dom'
import { AdminPage } from './AdminPage'
import * as useSessionModule from '../../shared/hooks/useSession'
import { adminApi } from '../../shared/lib/api'

vi.mock('../../shared/lib/api', () => ({
  adminApi: {
    getStats: vi.fn(),
    getQuests: vi.fn(),
    getDailyActivities: vi.fn(),
    toggleQuest: vi.fn(),
    toggleActivity: vi.fn(),
  },
}))

describe('AdminPage', () => {
  afterEach(() => {
    vi.clearAllMocks()
  })

  it('redirects unauthorized users', () => {
    vi.spyOn(useSessionModule, 'useSession').mockReturnValue({
      session: { uid: '1', crew_id: '1', role: 'SEEKER', kind: 'user', expires: 9999999999, token: 'abc' },
      loading: false,
    } as any)

    render(
      <BrowserRouter>
        <AdminPage />
      </BrowserRouter>
    )

    // Navigate to / replace should happen. We can mock Navigate or just check there is no Admin text.
    expect(screen.queryByText('Admin Dashboard')).toBeNull()
  })

  it('renders admin dashboard for GUIDE role', async () => {
    vi.spyOn(useSessionModule, 'useSession').mockReturnValue({
      session: { uid: '1', crew_id: '1', role: 'GUIDE', kind: 'user', expires: 9999999999, token: 'abc' },
      loading: false,
    } as any)

    vi.mocked(adminApi.getStats).mockResolvedValueOnce({
      total_users: 10,
      active_users_7d: 5,
      active_users_30d: 8,
      quest_completions: 42,
      daily_activity_completions_today: 3,
    })
    vi.mocked(adminApi.getQuests).mockResolvedValueOnce([
      { slug: 'q1', title: 'Quest 1', published: true, completion_count: 10 },
    ])
    vi.mocked(adminApi.getDailyActivities).mockResolvedValueOnce([
      { id: 1, slug: 'a1', title: 'Act 1', active: true, completion_count: 5 },
    ])

    render(
      <BrowserRouter>
        <AdminPage />
      </BrowserRouter>
    )

    expect(screen.getByText('Admin Dashboard')).toBeInTheDocument()
    
    await waitFor(() => {
      expect(screen.getByText('Ringkasan')).toBeInTheDocument()
      expect(screen.getByText('42')).toBeInTheDocument() // quest_completions
      expect(screen.getByText('Quest 1')).toBeInTheDocument()
      expect(screen.getByText('Act 1')).toBeInTheDocument()
    })
  })
})
