import type { Exercise, CrewMember } from '../types'

export type RelayLegState = 'done' | 'active' | 'next' | 'open'

export interface RelayLeg {
  challenge: Exercise
  state: RelayLegState
  isMyTurn: boolean
}

/**
 * Derives the relay rotation state for each leg of a RELAY Mission.
 *
 * The active leg is anchored to `active_challenge_assigned_to` (the backend
 * source of truth, set by the round-robin rotation after each completion).
 * This util only classifies existing data — it never re-implements the
 * rotation algorithm.
 *
 * States:
 *  - done:   challenge is DONE
 *  - active: the current leg — PENDING and (when assigned) matching
 *            active_challenge_assigned_to; falls back to the first PENDING
 *            leg when no leg is assigned yet
 *  - next:   the first PENDING leg after the active one (becomes the turn
 *            once the active leg is completed)
 *  - open:   any remaining PENDING leg
 */
export function deriveRelayLegs(
  exercises: Exercise[],
  activeAssignee: string | undefined,
  myUID?: string | null,
): RelayLeg[] {
  if (exercises.length === 0) return []

  const assignedIndex = exercises.findIndex(
    (c) => c.status === 'PENDING' && activeAssignee != null && c.assigned_to === activeAssignee,
  )
  const activeIndex = assignedIndex >= 0
    ? assignedIndex
    : exercises.findIndex((c) => c.status === 'PENDING')

  let nextMarked = false
  return exercises.map((challenge, i) => {
    if (challenge.status === 'DONE') {
      return { challenge, state: 'done' as const, isMyTurn: false }
    }
    if (i === activeIndex) {
      return {
        challenge,
        state: 'active' as const,
        isMyTurn: challenge.assigned_to === myUID && challenge.assigned_to != null,
      }
    }
    if (!nextMarked) {
      nextMarked = true
      return { challenge, state: 'next' as const, isMyTurn: false }
    }
    return { challenge, state: 'open' as const, isMyTurn: false }
  })
}

/** Resolves a member's display name, falling back to the raw UID. */
export function memberName(members: CrewMember[] | undefined, uid?: string | null): string {
  if (!uid) return ''
  const member = members?.find((m) => m.uid === uid)
  return member?.explorer_name || uid
}
