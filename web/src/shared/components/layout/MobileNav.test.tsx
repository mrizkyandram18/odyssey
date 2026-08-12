// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { MobileNav } from './MobileNav'
import { useSession } from '../../hooks/useSession'

vi.mock('../../hooks/useSession', () => ({
  useSession: vi.fn(),
}))

describe('MobileNav', () => {
  it('renders exactly Beranda, Misi, Jurnal, and Profil', () => {
    vi.mocked(useSession).mockReturnValue({ profile: { role: 'EXPLORER' } } as any)
    
    render(
      <MemoryRouter>
        <MobileNav />
      </MemoryRouter>
    )

    expect(screen.getByText('Beranda')).toBeInTheDocument()
    expect(screen.getByText('Misi')).toBeInTheDocument()
    expect(screen.getByText('Jurnal')).toBeInTheDocument()
    expect(screen.getByText('Profil')).toBeInTheDocument()

    // Ensure hidden items are not there
    expect(screen.queryByText('Kenangan')).not.toBeInTheDocument()
    expect(screen.queryByText('Koleksi')).not.toBeInTheDocument()
    expect(screen.queryByText('Galeri')).not.toBeInTheDocument()
  })
})
