import type { Relic } from '../../types'
import { Badge } from '../atoms/Badge'

export interface RelicDisplayProps {
  relic: Relic
  size?: 'sm' | 'md'
}

export function RelicDisplay({ relic, size = 'md' }: RelicDisplayProps) {
  const sizes = {
    sm: 'h-8 w-8',
    md: 'h-12 w-12',
  }

  return (
    <div className="flex items-center gap-2">
      <div className={`flex items-center justify-center rounded-lg bg-accent ${sizes[size]}`}>
        <span>💎</span>
      </div>
      <div className="flex flex-col">
        <span className="font-medium">{relic.code}</span>
        <Badge variant="default" size="sm">
          Collected
        </Badge>
      </div>
    </div>
  )
}
