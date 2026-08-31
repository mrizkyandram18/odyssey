// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { BottomNav } from './BottomNav'
import { useSession } from '../../hooks/useSession'

vi.mock('../../hooks/useSession', () => ({
  useSession: vi.fn(),
}))

describe('BottomNav Component', () => {
  beforeEach(() => {
    vi.resetAllMocks()
  })

  afterEach(() => {
    cleanup()
  })

  it('renders standard navigation links for seeker members', () => {
    vi.mocked(useSession).mockReturnValue({
      session: { uid: 'u1', role: 'SEEKER' } as any,
      profile: { uid: 'u1', role: 'SEEKER' } as any,
      loading: false,
      error: null,
      login: vi.fn(),
      logout: vi.fn(),
      refreshProfile: vi.fn(),
    })

    render(
      <MemoryRouter initialEntries={['/']}>
        <BottomNav />
      </MemoryRouter>
    )

    expect(screen.getByText('Beranda')).toBeInTheDocument()
    expect(screen.getByText('Toko Hadiah')).toBeInTheDocument()
    expect(screen.getByText('Profil')).toBeInTheDocument()
    expect(screen.queryByText('Admin')).not.toBeInTheDocument()
  })

  it('renders Admin navigation link for GUIDE role', () => {
    vi.mocked(useSession).mockReturnValue({
      session: { uid: 'u1', role: 'GUIDE' } as any,
      profile: { uid: 'u1', role: 'GUIDE' } as any,
      loading: false,
      error: null,
      login: vi.fn(),
      logout: vi.fn(),
      refreshProfile: vi.fn(),
    })

    render(
      <MemoryRouter initialEntries={['/']}>
        <BottomNav />
      </MemoryRouter>
    )

    expect(screen.getByText('Beranda')).toBeInTheDocument()
    expect(screen.getByText('Toko Hadiah')).toBeInTheDocument()
    expect(screen.getByText('Profil')).toBeInTheDocument()
    expect(screen.getByText('Admin')).toBeInTheDocument()
  })
})
