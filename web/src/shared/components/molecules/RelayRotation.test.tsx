// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/react'
import { RelayRotation } from './RelayRotation'
import type { Exercise, CrewMember, Mission } from '../../types'

const doneLeg = (id: number, by: string): Exercise => ({
  id,
  mission_id: 1,
  slug: `leg-${id}`,
  description: `Leg ${id} description`,
  status: 'DONE',
  completed_by: by,
  created_at: '2026-01-01T00:00:00Z',
})

const pendingLeg = (id: number, assignedTo?: string): Exercise => ({
  id,
  mission_id: 1,
  slug: `leg-${id}`,
  description: `Leg ${id} description`,
  status: 'PENDING',
  assigned_to: assignedTo ?? null,
  created_at: '2026-01-01T00:00:00Z',
})

const relayMission = (overrides: Partial<Mission> = {}): Mission => ({
  id: 104,
  family_id: 'crew-1',
  template_slug: 'shadow-trail',
  title: 'Shadow Trail',
  status: 'ACTIVE',
  Mission_type: 'RELAY',
  active_challenge_assigned_to: 'u2',
  created_at: '2026-01-01T00:00:00Z',
  ...overrides,
})

const members: CrewMember[] = [
  { uid: 'u1', explorer_name: 'Leo', role: 'SEEKER' },
  { uid: 'u2', explorer_name: 'Maya', role: 'GUIDE' },
  { uid: 'u3', explorer_name: 'Sam', role: 'BUILDER' },
]

describe('RelayRotation', () => {
  afterEach(() => cleanup())

  it('renders nothing for non-relay missions', () => {
    const { container } = render(
      <RelayRotation Mission={relayMission({ Mission_type: 'SOLO' })} exercises={[]} members={members} />,
    )
    expect(container).toBeEmptyDOMElement()
  })

  it('shows done legs with resolved names and the active assignee as the current turn', () => {
    render(
      <RelayRotation
        Mission={relayMission({ active_challenge_assigned_to: 'u2' })}
        exercises={[doneLeg(1, 'u1'), pendingLeg(2, 'u2')]}
        members={members}
      />,
    )
    expect(screen.getByRole('heading', { name: 'Giliran Keluarga' })).toBeInTheDocument()
    expect(screen.getByText('Done by Leo')).toBeInTheDocument()
    expect(screen.getByText("Maya's turn")).toBeInTheDocument()
    expect(screen.queryByText('Your turn')).not.toBeInTheDocument()
  })

  it('flags the active leg as "Your turn" for the assigned explorer', () => {
    render(
      <RelayRotation
        Mission={relayMission({ active_challenge_assigned_to: 'u2' })}
        exercises={[doneLeg(1, 'u1'), pendingLeg(2, 'u2')]}
        members={members}
        myUID="u2"
      />,
    )
    expect(screen.getByText('Your turn')).toBeInTheDocument()
  })

  it('shows an unassigned active leg as Open (no premature handoff)', () => {
    render(
      <RelayRotation
        Mission={relayMission({ active_challenge_assigned_to: undefined })}
        exercises={[pendingLeg(1), pendingLeg(2)]}
        members={members}
        myUID="u1"
      />,
    )
    expect(screen.getByText('Siapa saja')).toBeInTheDocument()
    expect(screen.queryByText('Your turn')).not.toBeInTheDocument()
  })

  it('marks the leg after the active one as "Up next"', () => {
    render(
      <RelayRotation
        Mission={relayMission({ active_challenge_assigned_to: 'u2' })}
        exercises={[doneLeg(1, 'u1'), pendingLeg(2, 'u2'), pendingLeg(3)]}
        members={members}
      />,
    )
    expect(screen.getByText('Up next')).toBeInTheDocument()
  })

  it('renders summary counts of done and current legs', () => {
    render(
      <RelayRotation
        Mission={relayMission({ active_challenge_assigned_to: 'u2' })}
        exercises={[doneLeg(1, 'u1'), pendingLeg(2, 'u2'), pendingLeg(3)]}
        members={members}
      />,
    )
    expect(screen.getByText('1')).toBeInTheDocument()
    expect(screen.getByText('done')).toBeInTheDocument()
    expect(screen.getByText('Maya')).toBeInTheDocument()
    expect(screen.getByText('sekarang')).toBeInTheDocument()
  })
})
