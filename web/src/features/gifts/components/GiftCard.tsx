import type { ChestView } from '../../../shared/types'

export interface GiftCardProps {
  chest: ChestView
  onOpen?: () => void
}

export function GiftCard({ chest, onOpen }: GiftCardProps) {
  return (
    <div className="rounded-lg border border-border bg-surface p-4">
      <div className="flex items-center gap-3">
        <span className="text-3xl">{chest.icon}</span>
        <div className="flex-1">
          <p className="font-medium">{chest.name}</p>
          <p className="text-xs text-muted-foreground">{chest.description}</p>
          <span className="text-xs text-muted-foreground">{chest.rarity}</span>
        </div>
      </div>
      {!chest.opened && onOpen && (
        <button
          onClick={onOpen}
          className="mt-3 w-full rounded-lg bg-accent py-2 text-sm font-medium text-background transition hover:opacity-90"
        >
          Open
        </button>
      )}
    </div>
  )
}
