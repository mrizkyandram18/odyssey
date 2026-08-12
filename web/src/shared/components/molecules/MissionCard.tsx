import { Link } from 'react-router-dom'
import type { Mission, MissionView } from '../../types'
import { Badge } from '../atoms/Badge'
import { YourTurnBadge } from './YourTurnBadge'

export interface MissionCardProps {
  Mission: Mission | MissionView
  onPress?: (Mission: Mission | MissionView) => void
  isMyTurn?: boolean
}

export function MissionCard({ Mission, onPress, isMyTurn }: MissionCardProps) {
  const statusVariant = {
    PENDING: 'default' as const,
    ACTIVE: 'primary' as const,
    DONE: 'success' as const,
  }[Mission.status]

  const isMissionView = 'challenge_count' in Mission

  const cardContent = (
    <div className={`flex flex-col gap-2 rounded-lg border bg-surface p-4 text-left transition-colors hover:bg-surface/80 ${isMyTurn ? 'border-yellow-500/50 shadow-[0_0_10px_rgba(234,179,8,0.1)]' : 'border-border'}`}>
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold">{Mission.title}</h3>
        <div className="flex items-center gap-2">
          {isMyTurn && <YourTurnBadge />}
          <Badge variant={statusVariant}>{Mission.status}</Badge>
        </div>
      </div>
      <div className="flex items-center justify-between text-sm text-muted-foreground">
        <span className="capitalize">{Mission.template_slug.replace(/-/g, ' ')}</span>
        {isMissionView && (
          <span className="text-xs">
            {Mission.completed_count} / {Mission.challenge_count} Exercises
          </span>
        )}
      </div>
    </div>
  )

  if (onPress) {
    return (
      <button onClick={() => onPress(Mission)} className="w-full text-left">
        {cardContent}
      </button>
    )
  }

  return (
    <Link to={`/missions/${Mission.id}`} className="block w-full">
      {cardContent}
    </Link>
  )
}

