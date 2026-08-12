// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor, cleanup } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QuestView } from './QuestView'
import { useQuest } from '../../shared/hooks/useQuest'
import { useSession } from '../../shared/hooks/useSession'

vi.mock('../../shared/hooks/useQuest', () => ({
  useQuest: vi.fn(),
}))

vi.mock('../../shared/hooks/useSession', () => ({
  useSession: vi.fn(),
}))

describe('QuestView', () => {
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

  it('renders Indonesian error and Home CTA when quest is not found (404)', async () => {
    vi.mocked(useQuest).mockReturnValue({
      quest: null,
      challenges: [],
      loading: false,
      error: 'Not found',
      startQuest: vi.fn(),
      completeChallenge: vi.fn(),
      selectBranch: vi.fn(),
    })

    render(
      <MemoryRouter>
        <QuestView questId={999} />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Belum ada misi yang bisa dikerjakan.')).toBeInTheDocument()
      expect(screen.getByText('Kembali ke Beranda')).toBeInTheDocument()
    })
  })
})
