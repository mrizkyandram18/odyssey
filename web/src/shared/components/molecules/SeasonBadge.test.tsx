// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/react'
import { SeasonBadge } from './SeasonBadge'
import type { SeasonSummary } from '../../types'

const activeSeason = (): SeasonSummary => ({
  definition: {
    id: 1,
    slug: 'season-autumn-2026',
    name: 'Autumn 2026',
    description: 'Autumn in Clockwork City',
    start_at: '2026-09-01T00:00:00Z',
    end_at: '2026-11-30T23:59:59Z',
    journey: 'clockwork-city',
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
  },
  state: 'ACTIVE',
})

const upcomingSeason = (): SeasonSummary => ({
  definition: {
    id: 2,
    slug: 'season-winter-2026',
    name: 'Winter 2026',
    description: 'Winter in Starlit Library',
    start_at: '2026-12-01T00:00:00Z',
    end_at: '2026-12-31T23:59:59Z',
    journey: 'starlit-library',
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
  },
  state: 'UPCOMING',
})

describe('SeasonBadge', () => {
  afterEach(() => cleanup())

  it('renders active season with progress bar', () => {
    render(
      <SeasonBadge
        season={activeSeason()}
        progress={{ missions_completed: 3, journey_progress: 65, journey_status: 'ACTIVE' }}
      />
    )

    expect(screen.getByTestId('season-badge')).toBeInTheDocument()
    expect(screen.getByText('Autumn 2026')).toBeInTheDocument()
    expect(screen.getByText('Aktif')).toBeInTheDocument()
    expect(screen.getByText('65%')).toBeInTheDocument()
    expect(screen.getByText('3')).toBeInTheDocument()
  })

  it('renders upcoming season without progress bar', () => {
    render(<SeasonBadge season={upcomingSeason()} />)

    expect(screen.getByTestId('season-badge')).toBeInTheDocument()
    expect(screen.getByText('Winter 2026')).toBeInTheDocument()
    expect(screen.getByText('Mendatang')).toBeInTheDocument()
    expect(screen.queryByText('65%')).not.toBeInTheDocument()
  })

  it('renders journey name from metadata', () => {
    render(<SeasonBadge season={activeSeason()} />)

    expect(screen.getByText('Clockwork City')).toBeInTheDocument()
  })
})
