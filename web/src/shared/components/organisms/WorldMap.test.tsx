// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, fireEvent, cleanup } from '@testing-library/react'
import { WorldMap } from './WorldMap'

describe('WorldMap', () => {
  afterEach(() => {
    cleanup()
  })

  it('renders all default 3 realms with status badges', () => {
    render(
      <WorldMap
        realms={[
          { crew_id: 'c1', realm: 'whispering-woods', status: 'ACTIVE', progress: 50, updated_at: '' },
          { crew_id: 'c1', realm: 'clockwork-city', status: 'LOCKED', progress: 0, updated_at: '' },
          { crew_id: 'c1', realm: 'starlit-library', status: 'LOCKED', progress: 0, updated_at: '' },
        ]}
      />
    )

    expect(screen.getByText('Whispering Woods')).toBeInTheDocument()
    expect(screen.getByText('Clockwork City')).toBeInTheDocument()
    expect(screen.getByText('Starlit Library')).toBeInTheDocument()

    expect(screen.getByText('Aktif')).toBeInTheDocument()
    expect(screen.getAllByText('🔒 Terkunci')).toHaveLength(2)
  })

  it('triggers onRealmSelect when clicking an unlocked realm', () => {
    const handleSelect = vi.fn()
    render(
      <WorldMap
        realms={[
          { crew_id: 'c1', realm: 'whispering-woods', status: 'COMPLETE', progress: 100, updated_at: '' },
          { crew_id: 'c1', realm: 'clockwork-city', status: 'ACTIVE', progress: 20, updated_at: '' },
        ]}
        onRealmSelect={handleSelect}
      />
    )

    fireEvent.click(screen.getByText('Clockwork City'))
    expect(handleSelect).toHaveBeenCalledWith('clockwork-city')

    fireEvent.click(screen.getByText('Whispering Woods'))
    expect(handleSelect).toHaveBeenCalledWith('whispering-woods')
  })

  it('does not trigger onRealmSelect when clicking a locked realm', () => {
    const handleSelect = vi.fn()
    render(
      <WorldMap
        realms={[
          { crew_id: 'c1', realm: 'whispering-woods', status: 'ACTIVE', progress: 50, updated_at: '' },
          { crew_id: 'c1', realm: 'clockwork-city', status: 'LOCKED', progress: 0, updated_at: '' },
        ]}
        onRealmSelect={handleSelect}
      />
    )

    fireEvent.click(screen.getByText('Clockwork City'))
    expect(handleSelect).not.toHaveBeenCalled()
  })
})
