import type { InventoryItem } from '../../../shared/types'
import { RelicCard } from './RelicCard'

export interface RelicGridProps {
  relics: InventoryItem[]
  onRelicClick?: (relic: InventoryItem) => void
}

export function RelicGrid({ relics, onRelicClick }: RelicGridProps) {
  if (relics.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">No relics collected yet.</p>
    )
  }

  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
      {relics.map((relic) => (
        <RelicCard
          key={relic.relic_slug}
          relic={relic}
          onClick={() => onRelicClick?.(relic)}
        />
      ))}
    </div>
  )
}
