import type { ReactionType, ReactionState } from '../../lib/api'

const REACTION_CONFIG: { type: ReactionType; emoji: string; label: string }[] = [
  { type: 'HEART', emoji: '❤️', label: 'Heart' },
  { type: 'CLAP',  emoji: '👏', label: 'Clap' },
  { type: 'STAR',  emoji: '⭐', label: 'Star' },
]

interface ReactionBarProps {
  state: ReactionState
  loading: boolean
  onReact: (type: ReactionType) => void
}

/**
 * ReactionBar — displays HEART / CLAP / STAR buttons with counts.
 * Shows current user's selection with an active ring.
 * Clicking an already-selected reaction is idempotent (upsert semantics).
 * Clicking a different reaction replaces the previous one.
 */
export function ReactionBar({ state, loading, onReact }: ReactionBarProps) {
  return (
    <div className="flex items-center gap-1.5" role="group" aria-label="Reactions">
      {REACTION_CONFIG.map(({ type, emoji, label }) => {
        const count = state.counts[type]
        const isSelected = state.myReaction === type
        return (
          <button
            key={type}
            onClick={() => onReact(type)}
            disabled={loading}
            aria-label={`${label} reaction${count > 0 ? `, ${count} total` : ''}`}
            aria-pressed={isSelected}
            className={[
              'inline-flex items-center gap-1 rounded-full border px-2.5 py-1 text-sm font-medium transition-all duration-150',
              'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-1',
              'disabled:opacity-50 disabled:cursor-not-allowed',
              isSelected
                ? 'border-accent-magic bg-accent-magic/10 text-accent-magic shadow-[0_0_8px_rgba(6,182,222,0.25)] ring-1 ring-accent-magic/40'
                : 'border-border-subtle bg-surface text-text-secondary hover:border-accent-magic/30 hover:bg-surface-elevated hover:text-text-primary',
            ].join(' ')}
          >
            <span>{emoji}</span>
            {count > 0 && (
              <span className="text-xs tabular-nums">{count}</span>
            )}
          </button>
        )
      })}
    </div>
  )
}
