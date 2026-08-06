import type { Quest } from '../../types'
import { Badge } from '../atoms/Badge'

export interface QuestCardProps {
  quest: Quest
  onPress?: (quest: Quest) => void
}

export function QuestCard({ quest, onPress }: QuestCardProps) {
  const statusVariant = {
    PENDING: 'default' as const,
    ACTIVE: 'primary' as const,
    DONE: 'success' as const,
  }[quest.status]

  return (
    <button
      onClick={() => onPress?.(quest)}
      className="flex flex-col gap-2 rounded-lg border border-border bg-surface p-3 text-left transition-colors hover:bg-surface/80"
    >
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold">{quest.title}</h3>
        <Badge variant={statusVariant} />
      </div>
      <p className="text-sm text-muted-foreground">
        {quest.template_slug}
      </p>
    </button>
  )
}
