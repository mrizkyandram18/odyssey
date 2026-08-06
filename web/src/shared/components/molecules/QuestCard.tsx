import { Link } from 'react-router-dom'
import type { Quest, QuestView } from '../../types'
import { Badge } from '../atoms/Badge'

export interface QuestCardProps {
  quest: Quest | QuestView
  onPress?: (quest: Quest | QuestView) => void
}

export function QuestCard({ quest, onPress }: QuestCardProps) {
  const statusVariant = {
    PENDING: 'default' as const,
    ACTIVE: 'primary' as const,
    DONE: 'success' as const,
  }[quest.status]

  const isQuestView = 'challenge_count' in quest

  const cardContent = (
    <div className="flex flex-col gap-2 rounded-lg border border-border bg-surface p-4 text-left transition-colors hover:bg-surface/80">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold">{quest.title}</h3>
        <Badge variant={statusVariant}>{quest.status}</Badge>
      </div>
      <div className="flex items-center justify-between text-sm text-muted-foreground">
        <span className="capitalize">{quest.template_slug.replace(/-/g, ' ')}</span>
        {isQuestView && (
          <span className="text-xs">
            {quest.completed_count} / {quest.challenge_count} Challenges
          </span>
        )}
      </div>
    </div>
  )

  if (onPress) {
    return (
      <button onClick={() => onPress(quest)} className="w-full text-left">
        {cardContent}
      </button>
    )
  }

  return (
    <Link to={`/quests/${quest.id}`} className="block w-full">
      {cardContent}
    </Link>
  )
}

