import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { achievementsApi, loreApi } from '../../shared/lib/api'
import type { AchievementView, LoreView } from '../../shared/types'
import { Badge } from '../../shared/components/atoms/Badge'

export function JournalPage() {
  const [tab, setTab] = useState<'achievements' | 'lore'>('achievements')
  const [achievements, setAchievements] = useState<AchievementView[]>([])
  const [loreEntries, setLoreEntries] = useState<LoreView[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const loadJournal = async () => {
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
    }
    loadJournal()
  }, [])

  const unlockedAchievementsCount = achievements.filter((a) => a.unlocked).length
  const unlockedLoreCount = loreEntries.filter((l) => l.unlocked).length

  return (
    <div className="flex flex-col gap-4 p-4 pb-safe">
      <header className="flex items-center justify-between">
        <Link to="/" className="text-sm text-muted-foreground hover:text-primary transition-colors">
          ← Home
        </Link>
        <h1 className="text-xl font-semibold">Family Journal</h1>
      </header>

      <div className="flex rounded-lg border border-border bg-surface p-1">
        <button
          onClick={() => setTab('achievements')}
          className={`flex-1 rounded-md py-2 text-center text-sm font-medium transition-colors ${
            tab === 'achievements'
              ? 'bg-primary text-black font-semibold'
              : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          Milestones ({unlockedAchievementsCount}/{achievements.length})
        </button>
        <button
          onClick={() => setTab('lore')}
          className={`flex-1 rounded-md py-2 text-center text-sm font-medium transition-colors ${
            tab === 'lore'
              ? 'bg-primary text-black font-semibold'
              : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          Story Lore ({unlockedLoreCount}/{loreEntries.length})
        </button>
      </div>

      {error && <p className="text-xs text-error bg-error/10 p-2 rounded">{error}</p>}

      {loading ? (
        <p className="text-sm text-muted-foreground">Loading journal entries...</p>
      ) : tab === 'achievements' ? (
        <div className="flex flex-col gap-3">
          {achievements.length === 0 ? (
            <p className="text-sm text-muted-foreground">No achievement definitions available yet.</p>
          ) : (
            achievements.map((ach) => {
              const progressPct = ach.threshold ? Math.min(100, (ach.progress / ach.threshold) * 100) : 0

              return (
                <div
                  key={ach.code}
                  className={`flex flex-col gap-2 rounded-lg border border-border bg-surface p-4 transition-all ${
                    ach.unlocked ? 'border-accent/40' : 'opacity-60'
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <h3 className="font-semibold text-sm">{ach.title || ach.code}</h3>
                    <Badge variant={ach.unlocked ? 'success' : 'default'} size="sm">
                      {ach.unlocked ? 'Unlocked' : ach.kind}
                    </Badge>
                  </div>

                  {ach.threshold > 1 && (
                    <div className="space-y-1">
                      <div className="flex justify-between text-xs text-muted-foreground">
                        <span>Progress</span>
                        <span>{ach.progress} / {ach.threshold}</span>
                      </div>
                      <div className="h-1.5 w-full rounded-full bg-border overflow-hidden">
                        <div
                          className="h-full bg-accent transition-all"
                          style={{ width: `${progressPct}%` }}
                        />
                      </div>
                    </div>
                  )}

                  {ach.awarded_at && (
                    <p className="text-xs text-muted-foreground text-right">
                      Awarded {new Date(ach.awarded_at).toLocaleDateString()}
                    </p>
                  )}
                </div>
              )
            })
          )}
        </div>
      ) : (
        <div className="flex flex-col gap-3">
          {loreEntries.length === 0 ? (
            <p className="text-sm text-muted-foreground">No story lore unlocked yet. Complete quests to discover lore!</p>
          ) : (
            loreEntries.map((lore) => (
              <div
                key={lore.slug}
                className={`flex flex-col gap-2 rounded-lg border border-border bg-surface p-4 transition-all ${
                  lore.unlocked ? 'border-secondary/40' : 'opacity-50'
                }`}
              >
                <div className="flex items-center justify-between">
                  <h3 className="font-semibold text-sm">{lore.title}</h3>
                  <Badge variant={lore.unlocked ? 'secondary' : 'default'} size="sm">
                    {lore.unlocked ? 'Discovered' : 'Locked'}
                  </Badge>
                </div>
                {lore.unlocked ? (
                  <p className="text-xs text-muted-foreground leading-relaxed italic">
                    "{lore.content}"
                  </p>
                ) : (
                  <p className="text-xs text-muted-foreground italic">
                    Complete chapter quests in {lore.realm.replace(/-/g, ' ')} to reveal this memory.
                  </p>
                )}
              </div>
            ))
          )}
        </div>
      )}
    </div>
  )
}
