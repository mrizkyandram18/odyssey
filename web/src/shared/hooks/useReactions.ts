import { useState, useEffect, useRef, useCallback } from 'react'
import { reactionsApi, deriveReactionState } from '../lib/api'
import type { ReactionType, TargetType, ReactionRow, ReactionState } from '../lib/api'
import { useSession } from './useSession'

interface UseReactionsOptions {
  targetType: TargetType
  targetId: number
  enabled?: boolean
}

interface UseReactionsResult {
  state: ReactionState
  loading: boolean
  error: string | null
  react: (reactionType: ReactionType) => Promise<void>
}

/**
 * useReactions — canonical hook for reading and writing peer reactions.
 *
 * Design decisions:
 * - Optimistic UI: updates state immediately, rolls back on error.
 * - Race prevention: an inflight flag blocks concurrent requests for same target.
 * - Identity: actor_uid is derived from session.uid (never from response or input).
 * - Semantics: upsert-only (HEART→STAR changes reaction, same reaction is idempotent).
 */
export function useReactions({ targetType, targetId, enabled = true }: UseReactionsOptions): UseReactionsResult {
  const { session } = useSession()
  const myUID = session?.uid ?? ''

  const [rows, setRows] = useState<ReactionRow[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const inflightRef = useRef(false)

  const fetch = useCallback(async () => {
    if (!enabled || !targetId) return
    setLoading(true)
    setError(null)
    try {
      const res = await reactionsApi.list(targetType, targetId)
      setRows(res.reactions ?? [])
    } catch (e) {
      setError(e instanceof Error ? e.message : 'failed to load reactions')
    } finally {
      setLoading(false)
    }
  }, [targetType, targetId, enabled])

  useEffect(() => {
    void fetch()
  }, [fetch])

  const react = useCallback(async (reactionType: ReactionType) => {
    if (!myUID) return
    // Prevent rapid-click duplicate / race
    if (inflightRef.current) return
    inflightRef.current = true

    const previousRows = rows

    // Optimistic update: derive what the new row set looks like
    // Upsert semantics: replace actor's existing reaction or insert new
    const optimisticRows: ReactionRow[] = [
      ...rows.filter(r => r.actor_uid !== myUID),
      {
        id: -1, // temp sentinel
        crew_id: '',
        target_type: targetType,
        target_id: targetId,
        actor_uid: myUID,
        reaction_type: reactionType,
        created_at: new Date().toISOString(),
      },
    ]
    setRows(optimisticRows)

    try {
      const serverRow = await reactionsApi.upsert(targetType, targetId, reactionType)
      // Replace optimistic row with real server row
      setRows(prev => [
        ...prev.filter(r => r.actor_uid !== myUID),
        serverRow,
      ])
    } catch (e) {
      // Rollback on failure
      setRows(previousRows)
      setError(e instanceof Error ? e.message : 'failed to save reaction')
    } finally {
      inflightRef.current = false
    }
  }, [rows, targetType, targetId, myUID])

  const state = deriveReactionState(rows, myUID)

  return { state, loading, error, react }
}
