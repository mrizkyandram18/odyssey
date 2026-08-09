import type { InventoryItem } from '../../../shared/types'
import { Card } from '../../../shared/components/atoms/Card'

export interface RelicCardProps {
  relic: InventoryItem
  onClick?: () => void
}

export function RelicCard({ relic }: RelicCardProps) {
  const rarityColors: Record<string, string> = {
    COMMON: 'border-border-subtle bg-surface',
    UNCOMMON: 'border-accent-nature/50 bg-accent-nature/5 shadow-[0_0_15px_rgba(16,185,129,0.1)]',
    RARE: 'border-accent-magic/50 bg-accent-magic/5 shadow-[0_0_15px_rgba(6,182,222,0.15)]',
    EPIC: 'border-accent-rare/50 bg-accent-rare/10 shadow-[0_0_15px_rgba(139,92,246,0.2)]',
    LEGENDARY: 'border-accent-reward/60 bg-accent-reward/10 shadow-[0_0_20px_rgba(245,158,11,0.25)]',
  }

  const rarityTextColors: Record<string, string> = {
    COMMON: 'text-text-secondary',
    UNCOMMON: 'text-accent-nature',
    RARE: 'text-accent-magic',
    EPIC: 'text-accent-rare',
    LEGENDARY: 'text-accent-reward',
  }

  return (
    <Card
      hoverable
      className={`h-full flex flex-col p-0 overflow-hidden group ${rarityColors[relic.rarity] || rarityColors.COMMON}`}
    >
      <div className="flex-1 flex items-center justify-center p-6 relative">
        <div className="absolute inset-0 bg-gradient-to-b from-transparent to-surface-elevated/50 pointer-events-none"></div>
        <span className="text-6xl drop-shadow-2xl transform transition-transform duration-500 group-hover:scale-110 group-hover:-translate-y-2 relative z-10">
          {relic.image}
        </span>
        {relic.is_new && (
          <span className="absolute top-2 right-2 z-20 bg-accent-magic text-black text-[10px] font-bold px-2 py-0.5 rounded shadow-[0_0_10px_rgba(6,182,222,0.5)]">
            NEW
          </span>
        )}
      </div>
      
      <div className="p-4 border-t border-border-subtle/50 bg-surface-elevated/80 backdrop-blur-sm">
        <p className="font-heading text-lg text-text-primary line-clamp-1 mb-1">{relic.name}</p>
        <div className="flex justify-between items-center mt-2">
          <span className={`text-[10px] font-bold uppercase tracking-wider ${rarityTextColors[relic.rarity] || rarityTextColors.COMMON}`}>
            {relic.rarity}
          </span>
          <span className="text-xs font-medium bg-surface border border-border-subtle px-2 py-0.5 rounded text-text-secondary">
            x{relic.owned_count}
          </span>
        </div>
      </div>
    </Card>
  )
}
