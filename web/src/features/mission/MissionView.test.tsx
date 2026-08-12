// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor, cleanup } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { MissionView } from './MissionView'
import { useMission } from '../../shared/hooks/useMission'
import { useSession } from '../../shared/hooks/useSession'

vi.mock('../../shared/hooks/useMission', () => ({
  useMission: vi.fn(),
}))

vi.mock('../../shared/hooks/useSession', () => ({
  useSession: vi.fn(),
}))

describe('MissionView', () => {
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

  it('renders Indonesian error and Home CTA when Mission is not found (404)', async () => {
    vi.mocked(useMission).mockReturnValue({
      Mission: null,
      exercises: [],
      loading: false,
      error: 'Not found',
      startMission: vi.fn(),
      completeChallenge: vi.fn(),
      selectBranch: vi.fn(),
    })

    render(
      <MemoryRouter>
        <MissionView missionId={999} />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Belum ada misi yang bisa dikerjakan.')).toBeInTheDocument()
      expect(screen.getByText('Kembali ke Beranda')).toBeInTheDocument()
    })
  })
})
