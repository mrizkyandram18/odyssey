import { useState, useEffect, useCallback } from 'react'

import { achievementsApi, loreApi, storyFragmentsApi } from '../../shared/lib/api'
import type { AchievementView, LoreView, StoryFragmentView } from '../../shared/types'
import { Card } from '../../shared/components/atoms/Card'
import { ProgressBar } from '../../shared/components/atoms/ProgressBar'

export function JournalPage() {
  const [tab, setTab] = useState<'achievements' | 'lore' | 'fragments'>('achievements')
  const [achievements, setAchievements] = useState<AchievementView[]>([])
  const [loreEntries, setLoreEntries] = useState<LoreView[]>([])
  const [fragments, setFragments] = useState<StoryFragmentView[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const loadJournal = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const [achData, loreData, fragData] = await Promise.all([
        achievementsApi.list().catch(() => []),
        loreApi.list().catch(() => []),
        storyFragmentsApi.list().catch(() => []),
      ])
      setAchievements(achData)
      setLoreEntries(loreData)
      setFragments(fragData)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'failed to load journal data')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadJournal()
  }, [loadJournal])

  const unlockedAchievementsCount = achievements.filter((a) => a.unlocked).length
  const unlockedLoreCount = loreEntries.filter((l) => l.unlocked).length
  const discoveredFragmentsCount = fragments.filter((f) => f.discovered).length

  return (
    <div className="flex flex-col gap-6 max-w-5xl mx-auto py-4">
      <header className="mb-2">
        <h1 className="font-heading text-4xl text-text-primary mb-2">Milestones & Jurnal Topik</h1>
        <p className="text-text-secondary">Catatan petualangan keluarga, cerita, dan kepingan kenangan yang terkumpul.</p>
      </header>

      {/* Tabs */}
      <div className="flex bg-surface-elevated rounded-lg p-1 border border-border-subtle shadow-md">
        <button
          onClick={() => setTab('achievements')}
          className={`flex-1 rounded-md py-3 text-center text-xs sm:text-sm md:text-base font-semibold transition-all duration-300 ${
            tab === 'achievements'
              ? 'bg-accent-reward text-black shadow-[0_0_15px_rgba(245,158,11,0.4)]'
              : 'text-text-secondary hover:text-text-primary hover:bg-surface-glass'
          }`}
        >
          Achievements ({unlockedAchievementsCount}/{achievements.length})
        </button>
        <button
          onClick={() => setTab('lore')}
          className={`flex-1 rounded-md py-3 text-center text-xs sm:text-sm md:text-base font-semibold transition-all duration-300 ${
            tab === 'lore'
              ? 'bg-accent-magic text-black shadow-[0_0_15px_rgba(6,182,222,0.4)]'
              : 'text-text-secondary hover:text-text-primary hover:bg-surface-glass'
          }`}
        >
          Story Lore ({unlockedLoreCount}/{loreEntries.length})
        </button>
        <button
          onClick={() => setTab('fragments')}
          className={`flex-1 rounded-md py-3 text-center text-xs sm:text-sm md:text-base font-semibold transition-all duration-300 ${
            tab === 'fragments'
              ? 'bg-accent-nature text-black shadow-[0_0_15px_rgba(16,185,129,0.4)]'
              : 'text-text-secondary hover:text-text-primary hover:bg-surface-glass'
          }`}
        >
          Story Fragments ({discoveredFragmentsCount}/{fragments.length})
        </button>
      </div>

      {error && (
        <div className="bg-accent-danger/10 border border-accent-danger/30 p-4 rounded-lg">
          <p className="text-sm font-medium text-accent-danger">{error}</p>
        </div>
      )}

      {loading ? (
        <div className="flex h-64 w-full items-center justify-center">
          <div className="flex flex-col items-center gap-4 animate-pulse">
            <div className="text-4xl">📖</div>
            <p className="text-sm text-text-secondary">Opening the chronicle...</p>
          </div>
        </div>
      ) : tab === 'achievements' ? (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 animate-in fade-in duration-500">
          {achievements.length === 0 ? (
            <div className="col-span-full py-12 text-center text-text-secondary italic bg-surface-elevated/30 rounded-lg border border-dashed border-border-subtle">
              No achievements configured yet.
            </div>
          ) : (
            achievements.map((ach) => {
              const progressPct = Math.min(100, Math.round((ach.progress / ach.threshold) * 100))

              return (
                <Card
                  key={ach.code || ach.slug || ach.id}
                  className={`flex flex-col justify-between p-6 transition-all duration-300 relative overflow-hidden ${
                    ach.unlocked
                      ? 'border-accent-reward/40 bg-surface-elevated shadow-lg shadow-accent-reward/5'
                      : 'opacity-60 bg-surface/50 border-border-subtle'
                  }`}
                >
                  <div className="flex items-start gap-4">
                    <div className="text-4xl shrink-0 p-2 bg-surface-elevated rounded-lg border border-border-subtle">
                      {ach.icon || '🏆'}
                    </div>
                    <div className="flex-1">
                      <div className="flex items-center justify-between gap-2 mb-1">
                        <h3 className="font-heading text-xl text-text-primary">{ach.title}</h3>
                        {ach.unlocked ? (
                          <span className="text-xs font-bold bg-accent-reward/20 text-accent-reward px-2 py-0.5 rounded border border-accent-reward/30">
                            UNLOCKED
                          </span>
                        ) : (
                          <span className="text-xs font-bold bg-surface border border-border-subtle text-text-secondary px-2 py-0.5 rounded">
                            LOCKED
                          </span>
                        )}
                      </div>
                      <p className="text-sm text-text-secondary mb-3">{ach.description || ach.code}</p>
                    </div>
                  </div>

                  {ach.threshold > 1 && !ach.unlocked && (
                    <div className="mt-auto pt-4">
                      <div className="flex justify-between text-xs text-text-secondary mb-1">
                        <span>Progress</span>
                        <span>{ach.progress} / {ach.threshold}</span>
                      </div>
                      <ProgressBar progress={progressPct} colorClass="bg-text-secondary" />
                    </div>
                  )}

                  {ach.awarded_at && (
                    <div className="mt-auto pt-4">
                      <p className="text-xs font-medium text-accent-reward border-t border-accent-reward/20 pt-2 text-right">
                        Awarded {new Date(ach.awarded_at).toLocaleDateString()}
                      </p>
                    </div>
                  )}
                </Card>
              )
            })
          )}
        </div>
      ) : tab === 'lore' ? (
        <div className="flex flex-col gap-4 animate-in fade-in duration-500">
          {loreEntries.length === 0 ? (
            <div className="col-span-full py-12 text-center text-text-secondary italic bg-surface-elevated/30 rounded-lg border border-dashed border-border-subtle">
              No story lore unlocked yet. Complete quests to discover the history of the realms!
            </div>
          ) : (
            loreEntries.map((lore) => (
              <Card
                key={lore.slug}
                className={`flex flex-col md:flex-row gap-6 items-start transition-all duration-500 relative overflow-hidden ${
                  lore.unlocked
                    ? 'border-accent-magic/30 bg-surface hover:border-accent-magic/60 shadow-lg hover:shadow-accent-magic/10'
                    : 'opacity-60 bg-surface-elevated/30'
                }`}
              >
                {lore.unlocked && <div className="absolute top-0 left-0 bottom-0 w-1 bg-accent-magic"></div>}

                <div className="flex-1 pl-2">
                  <div className="flex flex-wrap items-center justify-between gap-4 mb-3">
                    <div>
                      <span className="text-xs font-bold text-accent-magic uppercase tracking-wider block mb-1">
                        {lore.realm.replace(/-/g, ' ')}
                      </span>
                      <h3 className="font-heading text-2xl text-text-primary">{lore.title}</h3>
                    </div>
                    {!lore.unlocked && (
                      <span className="text-xs font-bold bg-surface border border-border-subtle text-text-secondary px-3 py-1 rounded">
                        LOCKED
                      </span>
                    )}
                  </div>

                  {lore.unlocked ? (
                    <div className="prose prose-invert max-w-none text-text-secondary/90 bg-surface-elevated/30 p-4 rounded-lg border border-border-subtle/50 italic leading-relaxed">
                      <p>"{lore.content}"</p>
                    </div>
                  ) : (
                    <p className="text-sm text-text-secondary italic flex items-center gap-2">
                      <span className="text-lg">🗝️</span>
                      Complete chapter quests in {lore.realm.replace(/-/g, ' ')} to reveal this memory.
                    </p>
                  )}
                </div>
              </Card>
            ))
          )}
        </div>
      ) : (
        /* Story Fragments Gallery */
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 animate-in fade-in duration-500">
          {fragments.length === 0 ? (
            <div className="col-span-full py-12 text-center text-text-secondary italic bg-surface-elevated/30 rounded-lg border border-dashed border-border-subtle">
              No story fragments found in the world.
            </div>
          ) : (
            fragments.map((frag) => (
              <Card
                key={frag.slug}
                className={`flex flex-col justify-between p-6 transition-all duration-300 relative overflow-hidden ${
                  frag.discovered
                    ? 'border-accent-nature/40 bg-surface-elevated shadow-lg shadow-accent-nature/5'
                    : 'opacity-60 bg-surface/50 border-border-subtle'
                }`}
              >
                <div>
                  <div className="flex items-center justify-between gap-2 mb-2">
                    <span className="text-xs font-bold text-accent-nature uppercase tracking-wider">
                      🌿 {frag.realm.replace(/-/g, ' ')}
                    </span>
                    {frag.discovered ? (
                      <span className="text-xs font-bold bg-accent-nature/20 text-accent-nature px-2 py-0.5 rounded border border-accent-nature/30">
                        DISCOVERED (+20 Poin)
                      </span>
                    ) : (
                      <span className="text-xs font-bold bg-surface border border-border-subtle text-text-secondary px-2 py-0.5 rounded">
                        {frag.is_hidden ? '🔒 REPLAY SECRET' : '🔒 UNDISCOVERED'}
                      </span>
                    )}
                  </div>

                  <h3 className="font-heading text-xl text-text-primary mb-2">{frag.title}</h3>

                  {frag.discovered ? (
                    <div className="bg-surface-elevated/50 p-4 rounded-lg border border-border-subtle/50 text-sm text-text-secondary italic leading-relaxed mb-3">
                      "{frag.content}"
                    </div>
                  ) : (
                    <p className="text-xs text-text-secondary italic mb-3">
                      {frag.is_hidden
                        ? 'Jelajahi kembali (Replay) ranah yang telah tamat untuk menemukan fragmen tersembunyi ini.'
                        : 'Jelajahi ranah dan selesaikan misi untuk menemukan fragmen cerita ini.'}
                    </p>
                  )}
                </div>

                {frag.discovered && frag.discovered_at && (
                  <div className="mt-auto pt-2 border-t border-border-subtle/40">
                    <p className="text-xs text-text-secondary text-right">
                      Ditemukan {new Date(frag.discovered_at).toLocaleDateString()}
                    </p>
                  </div>
                )}
              </Card>
            ))
          )}
        </div>
      )}
    </div>
  )
}
