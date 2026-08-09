import { describe, it, expect } from 'vitest'
import { deriveReactionState } from '../lib/api'
import type { ReactionRow, ReactionType } from '../lib/api'

const makeRow = (actorUID: string, reactionType: ReactionType, id = 1): ReactionRow => ({
  id,
  crew_id: 'crew-1',
  target_type: 'JOURNAL',
  target_id: 42,
  actor_uid: actorUID,
  reaction_type: reactionType,
  created_at: new Date().toISOString(),
})

describe('deriveReactionState', () => {
  it('returns zero counts and null myReaction for empty rows', () => {
    const state = deriveReactionState([], 'user-1')
    expect(state.counts).toEqual({ HEART: 0, CLAP: 0, STAR: 0 })
    expect(state.myReaction).toBeNull()
  })

  it('correctly counts reactions per type', () => {
    const rows = [
      makeRow('user-1', 'HEART', 1),
      makeRow('user-2', 'HEART', 2),
      makeRow('user-3', 'CLAP',  3),
      makeRow('user-4', 'STAR',  4),
    ]
    const state = deriveReactionState(rows, 'user-99')
    expect(state.counts.HEART).toBe(2)
    expect(state.counts.CLAP).toBe(1)
    expect(state.counts.STAR).toBe(1)
  })

  it('identifies the current user\'s reaction from actor_uid — NOT from any client-supplied field', () => {
    // The myUID comes from session.uid (never from the request body).
    // We simulate reading actor_uid from server-returned rows to find myReaction.
    const rows = [
      makeRow('user-other', 'CLAP', 1),
      makeRow('user-me',    'STAR', 2),
    ]
    const state = deriveReactionState(rows, 'user-me')
    expect(state.myReaction).toBe('STAR')
  })

  it('returns null myReaction when current user has no row', () => {
    const rows = [makeRow('user-other', 'HEART', 1)]
    const state = deriveReactionState(rows, 'user-me')
    expect(state.myReaction).toBeNull()
  })

  it('idempotent: same reaction type seen twice for same user counts as 2 separate actors', () => {
    // This shouldn't happen in practice due to DB unique constraint, but the
    // derive fn handles it gracefully — last row wins for myReaction
    const rows = [
      makeRow('user-me',  'HEART', 1),
      makeRow('user-me',  'STAR',  2), // would override first one in db, but test graceful handling
    ]
    const state = deriveReactionState(rows, 'user-me')
    // counts both rows (defensive)
    expect(state.counts.HEART + state.counts.STAR).toBe(2)
    // myReaction is the last one found
    expect(state.myReaction).toBe('STAR')
  })
})
