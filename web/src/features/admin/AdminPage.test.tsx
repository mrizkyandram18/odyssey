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
import { adminTasksApi } from '../../shared/lib/api'

vi.mock('../../shared/lib/api', () => ({
  adminTasksApi: {
    getTasks: vi.fn(),
    createTask: vi.fn(),
    updateTask: vi.fn(),
    deleteTask: vi.fn(),
    getPendingSubmissions: vi.fn(),
    verifySubmission: vi.fn(),
    getClaims: vi.fn(),
    processClaim: vi.fn(),
  },
}))

describe('AdminPage', () => {
  afterEach(() => {
    vi.clearAllMocks()
  })

  it('redirects unauthorized users', () => {
    vi.spyOn(useSessionModule, 'useSession').mockReturnValue({
      session: { uid: '1', family_id: '1', role: 'SEEKER', kind: 'user', expires: 9999999999, token: 'abc' },
      profile: { uid: '1', role: 'SEEKER' },
      loading: false,
    } as any)

    render(
      <BrowserRouter>
        <AdminPage />
      </BrowserRouter>
    )

    expect(screen.queryByText('Admin Panel Keluarga')).toBeNull()
  })

  it('renders admin dashboard for GUIDE role', async () => {
    vi.spyOn(useSessionModule, 'useSession').mockReturnValue({
      session: { uid: '1', family_id: '1', role: 'GUIDE', kind: 'user', expires: 9999999999, token: 'abc' },
      profile: { uid: '1', role: 'GUIDE' },
      loading: false,
    } as any)

    vi.mocked(adminTasksApi.getTasks).mockResolvedValueOnce([
      { id: 1, title: 'Tugas 1', step_order: 1, reward_coins: 50, reward_xp: 100, task_type: 'VIDEO_QUIZ', is_locked: false, status: 'UNLOCKED', config: {}, coins_earned: 0, xp_earned: 0 },
    ])
    vi.mocked(adminTasksApi.getPendingSubmissions).mockResolvedValueOnce([])
    vi.mocked(adminTasksApi.getClaims).mockResolvedValueOnce([])

    render(
      <BrowserRouter>
        <AdminPage />
      </BrowserRouter>
    )

    expect(screen.getByText('Admin Panel Keluarga')).toBeInTheDocument()

    await waitFor(() => {
      expect(screen.getByText('Antrean Verifikasi (0)')).toBeInTheDocument()
      expect(screen.getByText('Tidak Ada Antrean Verifikasi')).toBeInTheDocument()
    })
  })
})
