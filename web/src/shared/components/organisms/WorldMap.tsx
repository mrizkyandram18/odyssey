import type { RealmProgress } from '../../types'
import { Card } from '../atoms/Card'
import { ProgressBar } from '../atoms/ProgressBar'
import { getMergedRealmProgress, isRealmUnlocked } from '../../utils/realm'

export interface WorldMapProps {
  realms?: RealmProgress[]
  selectedRealm?: string
  onRealmSelect?: (realmSlug: string) => void
  onRealmPress?: (realm: RealmProgress) => void
}

export function WorldMap({
  realms,
  selectedRealm,
  onRealmSelect,
  onRealmPress,
}: WorldMapProps) {
  const mergedRealms = getMergedRealmProgress(realms)

  const handleSelect = (r: (typeof mergedRealms)[0]) => {
    if (!isRealmUnlocked(r.status)) return
    if (onRealmSelect) {
      onRealmSelect(r.slug)
    }
    if (onRealmPress) {
      const rp: RealmProgress = r.raw || {
        crew_id: '',
        realm: r.slug,
        status: r.status,
        progress: r.progress,
        updated_at: new Date().toISOString(),
      }
      onRealmPress(rp)
    }
  }

  return (
    <div className="flex flex-col gap-4 w-full" data-testid="world-map">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {mergedRealms.map((r) => {
          const unlocked = isRealmUnlocked(r.status)
          const isSelected = selectedRealm === r.slug

          return (
            <Card
              key={r.slug}
              data-testid={`world-map-realm-${r.slug}`}
              className={`flex flex-col justify-between p-5 relative overflow-hidden transition-all duration-300 ${
                !unlocked
                  ? 'opacity-60 grayscale bg-surface-elevated/30 border-border-subtle cursor-not-allowed'
                  : isSelected
                  ? 'border-accent-magic shadow-[0_0_20px_rgba(6,182,222,0.2)] bg-surface-elevated cursor-pointer scale-[1.02]'
                  : 'bg-surface border-border-subtle hover:border-accent-magic/50 hover:shadow-md cursor-pointer'
              }`}
              onClick={() => handleSelect(r)}
            >
              {/* Background Glow */}
              {unlocked && (
                <div className="absolute top-0 right-0 w-24 h-24 bg-accent-magic/10 rounded-full blur-2xl pointer-events-none" />
              )}

              <div className="flex flex-col gap-3 relative z-10">
                {/* Header: Icon & Status Badge */}
                <div className="flex items-center justify-between">
                  <span className="text-3xl" role="img" aria-label={r.name}>
                    {r.icon}
                  </span>
                  <span
                    className={`text-[10px] font-bold px-2.5 py-1 rounded-full uppercase tracking-wider ${
                      r.status === 'COMPLETE'
                        ? 'bg-accent-nature/20 text-accent-nature border border-accent-nature/30'
                        : r.status === 'ACTIVE'
                        ? 'bg-accent-magic/20 text-accent-magic border border-accent-magic/30'
                        : 'bg-surface-elevated text-text-secondary border border-border-subtle'
                    }`}
                  >
                    {r.status === 'COMPLETE'
                      ? 'Selesai'
                      : r.status === 'ACTIVE'
                      ? 'Aktif'
                      : '🔒 Terkunci'}
                  </span>
                </div>

                {/* Realm Info */}
                <div>
                  <h3 className="font-heading text-lg text-text-primary mb-1">
                    {r.name}
                  </h3>
                  <p className="text-xs text-text-secondary leading-relaxed line-clamp-2">
                    {r.description}
                  </p>
                </div>
              </div>

              {/* Progress Footer */}
              <div className="mt-4 pt-3 border-t border-border-subtle/50 relative z-10">
                <div className="flex justify-between items-center text-xs mb-1.5">
                  <span className="text-text-secondary font-medium">Progres Ranah</span>
                  <span className="font-bold text-text-primary">{r.progress}%</span>
                </div>
                <ProgressBar
                  progress={r.progress}
                  colorClass={
                    r.status === 'COMPLETE'
                      ? 'bg-accent-nature'
                      : r.status === 'ACTIVE'
                      ? 'bg-accent-magic'
                      : 'bg-text-secondary/30'
                  }
                />
              </div>
            </Card>
          )
        })}
      </div>
    </div>
  )
}
