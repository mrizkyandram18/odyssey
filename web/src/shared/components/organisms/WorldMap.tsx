import { useState } from 'react'
import type { RealmProgress, ReplayResult } from '../../types'
import { Card } from '../atoms/Card'
import { Button } from '../atoms/Button'
import { ProgressBar } from '../atoms/ProgressBar'
import { getMergedRealmProgress, isRealmUnlocked } from '../../utils/realm'
import { storyFragmentsApi } from '../../lib/api'

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
  const [replayingSlug, setReplayingSlug] = useState<string | null>(null)
  const [replayModal, setReplayModal] = useState<ReplayResult | null>(null)

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

  const handleReplayClick = async (e: React.MouseEvent, slug: string) => {
    e.stopPropagation()
    setReplayingSlug(slug)
    try {
      const res = await storyFragmentsApi.replay(slug)
      setReplayModal(res)
    } catch {
      // Best-effort error handle
    } finally {
      setReplayingSlug(null)
    }
  }

  return (
    <div className="flex flex-col gap-4 w-full" data-testid="world-map">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {mergedRealms.map((r) => {
          const unlocked = isRealmUnlocked(r.status)
          const isSelected = selectedRealm === r.slug
          const isComplete = r.status === 'COMPLETE'

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
                      isComplete
                        ? 'bg-accent-nature/20 text-accent-nature border border-accent-nature/30'
                        : r.status === 'ACTIVE'
                        ? 'bg-accent-magic/20 text-accent-magic border border-accent-magic/30'
                        : 'bg-surface-elevated text-text-secondary border border-border-subtle'
                    }`}
                  >
                    {isComplete
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

              {/* Progress & Replay Footer */}
              <div className="mt-4 pt-3 border-t border-border-subtle/50 relative z-10 flex flex-col gap-2">
                <div className="flex justify-between items-center text-xs">
                  <span className="text-text-secondary font-medium">Progres Ranah</span>
                  <span className="font-bold text-text-primary">{r.progress}%</span>
                </div>
                <ProgressBar
                  progress={r.progress}
                  colorClass={
                    isComplete
                      ? 'bg-accent-nature'
                      : r.status === 'ACTIVE'
                      ? 'bg-accent-magic'
                      : 'bg-text-secondary/30'
                  }
                />

                {isComplete && (
                  <Button
                    size="sm"
                    variant="secondary"
                    isLoading={replayingSlug === r.slug}
                    onClick={(e) => handleReplayClick(e, r.slug)}
                    className="mt-2 text-xs border-accent-nature/50 hover:border-accent-nature text-accent-nature bg-accent-nature/10"
                  >
                    🔁 Jelajahi Rahasia Replay
                  </Button>
                )}
              </div>
            </Card>
          )
        })}
      </div>

      {/* Replay Dialogue Modal Banner */}
      {replayModal && (
        <Card className="p-6 border-accent-nature/40 bg-surface-elevated/90 relative mt-2 animate-in fade-in duration-300">
          <div className="flex items-start justify-between gap-4">
            <div>
              <div className="flex items-center gap-2 mb-1">
                <span className="text-2xl">✨</span>
                <h4 className="font-heading text-xl text-text-primary">
                  Dialog Bonus Replay: {replayModal.realm.replace(/-/g, ' ')}
                </h4>
              </div>
              <p className="text-sm text-text-secondary italic mb-3">
                "{replayModal.bonus_dialogue}"
              </p>

              {replayModal.unlocked_fragments.length > 0 && (
                <div className="bg-accent-nature/10 border border-accent-nature/30 p-3 rounded-lg text-xs font-bold text-accent-nature">
                  🎉 Rahasia Terbuka: Fragmen cerita tersembunyi "
                  {replayModal.unlocked_fragments[0].title}" telah ditambahkan ke Jurnal!
                </div>
              )}
            </div>
            <button
              onClick={() => setReplayModal(null)}
              className="text-text-secondary hover:text-text-primary font-bold text-lg p-1"
            >
              ✕
            </button>
          </div>
        </Card>
      )}
    </div>
  )
}
