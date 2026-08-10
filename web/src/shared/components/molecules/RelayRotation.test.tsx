// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/react'
import { RelayRotation } from './RelayRotation'
import type { Challenge, CrewMember, Quest } from '../../types'

const doneLeg = (id: number, by: string): Challenge => ({
  id,
  quest_id: 1,
  slug: `leg-${id}`,
  description: `Leg ${id} description`,
  status: 'DONE',
  completed_by: by,
  created_at: '2026-01-01T00:00:00Z',
})

const pendingLeg = (id: number, assignedTo?: string): Challenge => ({
  id,
  quest_id: 1,
  slug: `leg-${id}`,
  description: `Leg ${id} description`,
  status: 'PENDING',
  assigned_to: assignedTo ?? null,
  created_at: '2026-01-01T00:00:00Z',
})

const relayQuest = (overrides: Partial<Quest> = {}): Quest => ({
  id: 104,
  crew_id: 'crew-1',
  template_slug: 'shadow-trail',
  title: 'Shadow Trail',
  status: 'ACTIVE',
  quest_type: 'RELAY',
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

  it('renders nothing for non-relay quests', () => {
    const { container } = render(
      <RelayRotation quest={relayQuest({ quest_type: 'SOLO' })} challenges={[]} members={members} />,
    )
    expect(container).toBeEmptyDOMElement()
  })

  it('shows done legs with resolved names and the active assignee as the current turn', () => {
    render(
      <RelayRotation
        quest={relayQuest({ active_challenge_assigned_to: 'u2' })}
        challenges={[doneLeg(1, 'u1'), pendingLeg(2, 'u2')]}
        members={members}
      />,
    )
    expect(screen.getByRole('heading', { name: 'Crew Relay' })).toBeInTheDocument()
    expect(screen.getByText('Done by Leo')).toBeInTheDocument()
    expect(screen.getByText("Maya's turn")).toBeInTheDocument()
    expect(screen.queryByText('Your turn')).not.toBeInTheDocument()
  })

  it('flags the active leg as "Your turn" for the assigned explorer', () => {
    render(
      <RelayRotation
        quest={relayQuest({ active_challenge_assigned_to: 'u2' })}
        challenges={[doneLeg(1, 'u1'), pendingLeg(2, 'u2')]}
        members={members}
        myUID="u2"
      />,
    )
    expect(screen.getByText('Your turn')).toBeInTheDocument()
  })

  it('shows an unassigned active leg as Open (no premature handoff)', () => {
    render(
      <RelayRotation
        quest={relayQuest({ active_challenge_assigned_to: undefined })}
        challenges={[pendingLeg(1), pendingLeg(2)]}
        members={members}
        myUID="u1"
      />,
    )
    expect(screen.getByText('Open')).toBeInTheDocument()
    expect(screen.queryByText('Your turn')).not.toBeInTheDocument()
  })

  it('marks the leg after the active one as "Up next"', () => {
    render(
      <RelayRotation
        quest={relayQuest({ active_challenge_assigned_to: 'u2' })}
        challenges={[doneLeg(1, 'u1'), pendingLeg(2, 'u2'), pendingLeg(3)]}
        members={members}
      />,
    )
    expect(screen.getByText('Up next')).toBeInTheDocument()
  })

  it('renders summary counts of done and current legs', () => {
    render(
      <RelayRotation
        quest={relayQuest({ active_challenge_assigned_to: 'u2' })}
        challenges={[doneLeg(1, 'u1'), pendingLeg(2, 'u2'), pendingLeg(3)]}
        members={members}
      />,
    )
    expect(screen.getByText('1')).toBeInTheDocument()
    expect(screen.getByText('done')).toBeInTheDocument()
    expect(screen.getByText('Maya')).toBeInTheDocument()
    expect(screen.getByText('now')).toBeInTheDocument()
  })
})
