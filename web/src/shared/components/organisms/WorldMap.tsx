import type { RealmProgress } from '../../types'

export interface WorldMapProps {
  realms?: RealmProgress[]
  onRealmPress?: (realm: RealmProgress) => void
}

export function WorldMap({ realms, onRealmPress }: WorldMapProps) {
  void realms
  void onRealmPress

  return (
    <div className="flex h-full w-full items-center justify-center rounded-lg border border-border bg-surface">
      <p className="text-sm text-muted-foreground">World map — The Whispering Woods</p>
    </div>
  )
}
