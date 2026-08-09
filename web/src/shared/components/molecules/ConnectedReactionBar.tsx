import { ReactionBar } from './ReactionBar'
import { useReactions } from '../../hooks/useReactions'
import type { TargetType } from '../../lib/api'
import type { ReactionType } from '../../lib/api'

interface ConnectedReactionBarProps {
  targetType: TargetType
  targetId: number
}

/**
 * ConnectedReactionBar — self-contained component that fetches, displays and
 * writes reactions for any target (JOURNAL | QUEST). Uses the canonical
 * useReactions hook for optimistic UI with rollback + race prevention.
 */
export function ConnectedReactionBar({ targetType, targetId }: ConnectedReactionBarProps) {
  const { state, loading, react } = useReactions({ targetType, targetId })

  const handleReact = (type: ReactionType) => {
    void react(type)
  }

  return <ReactionBar state={state} loading={loading} onReact={handleReact} />
}
