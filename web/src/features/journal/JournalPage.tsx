import { useState, useEffect, useCallback } from 'react'

import { achievementsApi, loreApi, storyFragmentsApi } from '../../shared/lib/api'
import type { AchievementView, LoreView, StoryFragmentView } from '../../shared/types'
import { Card } from '../../shared/components/atoms/Card'

export function JournalPage() {
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
      setAchievements(achData || [])
      setLoreEntries(loreData || [])
      setFragments(fragData || [])
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Gagal memuat data jurnal')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadJournal()
  }, [loadJournal])

  const unlockedAchievements = achievements.filter(a => a.unlocked)
  const unlockedLore = loreEntries.filter(l => l.unlocked)
  const discoveredFragments = fragments.filter(f => f.discovered)

  // Merge into a single timeline array
  const timelineItems = [
    ...unlockedAchievements.map(a => ({
      type: 'ACHIEVEMENT',
      id: `ach-${a.id}`,
      title: a.title,
      description: a.description,
      date: a.awarded_at ? new Date(a.awarded_at).getTime() : 0,
      dateStr: a.awarded_at,
      icon: a.icon || '🏆',
      color: 'text-accent-reward',
      bgColor: 'bg-accent-reward/10',
      borderColor: 'border-accent-reward/30',
    })),
    ...unlockedLore.map(l => ({
      type: 'LORE',
      id: `lore-${l.slug}`,
      title: l.title,
      description: l.content,
      date: l.unlocked_at ? new Date(l.unlocked_at).getTime() : 0,
      dateStr: l.unlocked_at,
      icon: '📖',
      color: 'text-accent-magic',
      bgColor: 'bg-accent-magic/10',
      borderColor: 'border-accent-magic/30',
    })),
    ...discoveredFragments.map(f => ({
      type: 'FRAGMENT',
      id: `frag-${f.slug}`,
      title: f.title,
      description: f.content,
      date: f.discovered_at ? new Date(f.discovered_at).getTime() : 0,
      dateStr: f.discovered_at,
      icon: '🧩',
      color: 'text-accent-nature',
      bgColor: 'bg-accent-nature/10',
      borderColor: 'border-accent-nature/30',
    }))
  ].sort((a, b) => b.date - a.date)

  return (
    <div className="flex flex-col gap-6 max-w-3xl mx-auto py-4">
      <header className="mb-2 text-center">
        <h1 className="font-heading text-3xl font-bold text-text-primary mb-2">Perjalanan Keluarga</h1>
        <p className="text-sm text-text-secondary">Jejak petualangan, pencapaian, dan cerita yang telah kita lalui bersama.</p>
      </header>

      {/* Summary Stats */}
      <div className="grid grid-cols-3 gap-3">
         <Card className="p-3 text-center bg-surface-elevated">
            <span className="block text-2xl mb-1">🏆</span>
            <div className="text-xl font-bold text-text-primary">{unlockedAchievements.length}</div>
            <div className="text-xs text-text-secondary">Pencapaian</div>
         </Card>
         <Card className="p-3 text-center bg-surface-elevated">
            <span className="block text-2xl mb-1">📖</span>
            <div className="text-xl font-bold text-text-primary">{unlockedLore.length}</div>
            <div className="text-xs text-text-secondary">Cerita</div>
         </Card>
         <Card className="p-3 text-center bg-surface-elevated">
            <span className="block text-2xl mb-1">🧩</span>
            <div className="text-xl font-bold text-text-primary">{discoveredFragments.length}</div>
            <div className="text-xs text-text-secondary">Kenangan</div>
         </Card>
      </div>

      {error && (
        <div className="bg-accent-danger/10 border border-accent-danger/30 p-4 rounded-lg">
          <p className="text-sm font-medium text-accent-danger">{error}</p>
        </div>
      )}

      {loading ? (
        <div className="flex h-64 w-full items-center justify-center">
          <div className="flex flex-col items-center gap-4 animate-pulse">
            <div className="text-4xl">🕰️</div>
            <p className="text-sm text-text-secondary">Menyusun kronologi...</p>
          </div>
        </div>
      ) : timelineItems.length === 0 ? (
        <div className="py-12 text-center text-text-secondary bg-surface-elevated/30 rounded-xl border border-dashed border-border-subtle">
          Belum ada catatan perjalanan. Mulailah misi untuk mengukir ceritamu!
        </div>
      ) : (
        <div className="relative border-l-2 border-border-subtle ml-4 sm:ml-6 mt-4">
          {timelineItems.map((item, index) => (
            <div key={item.id} className="mb-8 relative pl-6 sm:pl-8 animate-in fade-in slide-in-from-bottom-4" style={{ animationDelay: `${index * 50}ms` }}>
              {/* Timeline dot */}
              <div className={`absolute -left-[17px] top-1 w-8 h-8 rounded-full border-4 border-surface ${item.bgColor} ${item.color} flex items-center justify-center text-sm shadow-sm`}>
                {item.icon}
              </div>
              
              <Card className="p-5 bg-surface-elevated border border-border-subtle hover:border-accent-reward/30 transition-all hover:shadow-md">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between mb-2 gap-2">
                  <h3 className="font-heading text-lg font-bold text-text-primary">{item.title}</h3>
                  <span className="text-xs font-medium text-text-secondary">
                    {item.dateStr ? new Date(item.dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' }) : 'Selesai'}
                  </span>
                </div>
                
                {item.type === 'LORE' || item.type === 'FRAGMENT' ? (
                   <p className="text-sm text-text-secondary italic leading-relaxed bg-surface p-3 rounded-lg border border-border-subtle/50">
                     "{item.description}"
                   </p>
                ) : (
                   <p className="text-sm text-text-secondary">
                     {item.description}
                   </p>
                )}
                
                <div className="mt-3 inline-block">
                  <span className={`text-[10px] font-bold uppercase tracking-wider px-2 py-1 rounded border ${item.bgColor} ${item.color} ${item.borderColor}`}>
                    {item.type === 'ACHIEVEMENT' ? 'Pencapaian' : item.type === 'LORE' ? 'Cerita' : 'Kepingan Kenangan'}
                  </span>
                </div>
              </Card>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
