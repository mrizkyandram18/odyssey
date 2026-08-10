import { useState } from 'react'
import { apiClient } from '../../lib/api'

/**
 * STALE (Phase 2a superseded): posts legacy fields (target_user_id / emoji_code)
 * that no longer match /api/reactions (target_type / target_id / reaction_type).
 * Prefer ConnectedReactionBar + useReactions for JOURNAL reactions.
 * Kept for reference only — do not wire into new UI.
 */
export interface ReactionPickerProps {
  targetUserId: string
  questId?: string
  onReactionAdded?: () => void
}

const EMOJIS = ['❤️', '🔥', '👍', '👏', '🎉', '😂']

export function ReactionPicker({ targetUserId, questId, onReactionAdded }: ReactionPickerProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [loading, setLoading] = useState(false)

  const handleSelect = async (emojiCode: string) => {
    setLoading(true)
    try {
      await apiClient.post('/api/reactions', {
        target_user_id: targetUserId,
        quest_id: questId,
        emoji_code: emojiCode,
      })
      if (onReactionAdded) {
        onReactionAdded()
      }
      setIsOpen(false)
    } catch (e) {
      console.error('failed to add reaction', e)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="relative inline-block">
      <button 
        onClick={() => setIsOpen(!isOpen)}
        className="flex h-8 items-center gap-1 rounded-full border border-border bg-surface px-3 text-sm text-muted-foreground transition hover:bg-accent/10 hover:text-accent disabled:opacity-50"
        disabled={loading}
      >
        <span>+</span>
        <span className="text-xs">React</span>
      </button>

      {isOpen && (
        <div className="absolute bottom-full left-0 mb-2 flex gap-1 rounded-full border border-border bg-background p-2 shadow-md">
          {EMOJIS.map(emoji => (
            <button
              key={emoji}
              onClick={() => handleSelect(emoji)}
              className="flex h-8 w-8 items-center justify-center rounded-full text-lg transition hover:bg-accent/20"
            >
              {emoji}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
