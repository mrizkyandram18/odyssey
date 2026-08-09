import { useState, useEffect, useCallback } from 'react'

import { achievementsApi, loreApi } from '../../shared/lib/api'
import type { AchievementView, LoreView } from '../../shared/types'
import { Card } from '../../shared/components/atoms/Card'
import { Button } from '../../shared/components/atoms/Button'
import { ProgressBar } from '../../shared/components/atoms/ProgressBar'

export function JournalPage() {
  const [tab, setTab] = useState<'achievements' | 'lore'>('achievements')
  const [achievements, setAchievements] = useState<AchievementView[]>([])
  const [loreEntries, setLoreEntries] = useState<LoreView[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const loadJournal = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const [achData, loreData] = await Promise.all([
        achievementsApi.list().catch(() => []),
        loreApi.list().catch(() => []),
      ])
      setAchievements(achData)
      setLoreEntries(loreData)
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

  return (
    <div className="flex flex-col gap-6 max-w-5xl mx-auto py-4">
      <header className="mb-2">
        <h1 className="font-heading text-4xl text-text-primary mb-2">Milestones</h1>
        <p className="text-text-secondary">Chronicles of your crew's legendary feats and discovered lore.</p>
      </header>

      {/* Tabs */}
      <div className="flex bg-surface-elevated rounded-lg p-1 border border-border-subtle shadow-md">
        <button
          onClick={() => setTab('achievements')}
          className={`flex-1 rounded-md py-3 text-center text-sm md:text-base font-semibold transition-all duration-300 ${
            tab === 'achievements'
              ? 'bg-accent-reward text-black shadow-[0_0_15px_rgba(245,158,11,0.4)]'
              : 'text-text-secondary hover:text-text-primary hover:bg-surface-glass'
          }`}
        >
          Achievements ({unlockedAchievementsCount}/{achievements.length})
        </button>
        <button
          onClick={() => setTab('lore')}
          className={`flex-1 rounded-md py-3 text-center text-sm md:text-base font-semibold transition-all duration-300 ${
            tab === 'lore'
              ? 'bg-accent-magic text-black shadow-[0_0_15px_rgba(6,182,222,0.4)]'
              : 'text-text-secondary hover:text-text-primary hover:bg-surface-glass'
          }`}
        >
          Story Lore ({unlockedLoreCount}/{loreEntries.length})
        </button>
      </div>

      {error && (
        <div className="flex items-center justify-between gap-4 rounded-lg bg-accent-danger/10 border border-accent-danger/30 p-4 text-sm font-medium text-accent-danger">
          <span>{error}</span>
          <Button size="sm" variant="danger" onClick={loadJournal}>
            Retry
          </Button>
        </div>
      )}

      {loading ? (
        <div className="flex h-64 w-full items-center justify-center">
          <div className="flex flex-col items-center gap-4 animate-pulse">
            <div className="text-4xl">📖</div>
            <p className="text-sm text-text-secondary">Dusting off the tomes...</p>
          </div>
        </div>
      ) : tab === 'achievements' ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 animate-in fade-in duration-500">
          {achievements.length === 0 ? (
            <div className="col-span-full py-12 text-center text-text-secondary italic bg-surface-elevated/30 rounded-lg border border-dashed border-border-subtle">
              No legendary feats have been recorded yet.
            </div>
          ) : (
            achievements.map((ach) => {
              const progressPct = ach.threshold ? Math.min(100, (ach.progress / ach.threshold) * 100) : 0

              return (
                <Card
                  key={ach.code}
                  className={`flex flex-col h-full transition-all duration-500 ${
                    ach.unlocked 
                      ? 'border-accent-reward/50 bg-gradient-to-b from-accent-reward/10 to-surface shadow-[0_4px_20px_rgba(245,158,11,0.05)] hover:-translate-y-1' 
                      : 'opacity-60 grayscale hover:grayscale-0'
                  }`}
                >
                  <div className="flex items-center justify-between mb-4">
                    <div className={`w-12 h-12 rounded-full flex items-center justify-center text-xl shrink-0 ${
                      ach.unlocked ? 'bg-accent-reward text-black shadow-[0_0_15px_rgba(245,158,11,0.5)]' : 'bg-surface-elevated text-text-secondary border border-border-subtle'
                    }`}>
                      {ach.unlocked ? '🏆' : '🔒'}
                    </div>
                    <span className={`text-xs font-bold px-2 py-1 rounded uppercase tracking-wider ${
                      ach.unlocked ? 'bg-accent-reward/20 text-accent-reward' : 'bg-surface border border-border-subtle text-text-secondary'
                    }`}>
                      {ach.unlocked ? 'Unlocked' : ach.kind}
                    </span>
                  </div>

                  <h3 className={`font-heading text-xl mb-2 ${ach.unlocked ? 'text-text-primary' : 'text-text-secondary'}`}>
                    {ach.title || ach.code}
                  </h3>
                  
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
      ) : (
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
      )}
    </div>
  )
}
