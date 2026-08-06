import type { InventoryItem } from '../../../shared/types'
import { NewRelicBadge } from './NewRelicBadge'

export interface RelicCardProps {
  relic: InventoryItem
  onClick?: () => void
}

export function RelicCard({ relic, onClick }: RelicCardProps) {
  const rarityColors: Record<string, string> = {
    COMMON: 'border-gray-400',
    UNCOMMON: 'border-green-400',
    RARE: 'border-blue-400',
    EPIC: 'border-purple-400',
    LEGENDARY: 'border-amber-400',
  }

  return (
    <button
      onClick={onClick}
      className={`w-full rounded-lg border-2 ${rarityColors[relic.rarity] || 'border-border'} bg-surface p-3 text-left transition hover:scale-[1.01]`}
    >
      <div className="flex items-center gap-3">
        <span className="text-3xl">{relic.image}</span>
        <div className="flex-1">
          <div className="flex items-center gap-2">
            <p className="font-medium">{relic.name}</p>
            {relic.is_new && <NewRelicBadge />}
          </div>
          <p className="text-xs text-muted-foreground">{relic.description}</p>
          <div className="mt-1 flex items-center gap-2">
            <span className="text-xs text-muted-foreground">{relic.realm.replace(/-/g, ' ')}</span>
            <span className="text-xs text-muted-foreground">x{relic.owned_count}</span>
          </div>
        </div>
      </div>
    </button>
  )
}
