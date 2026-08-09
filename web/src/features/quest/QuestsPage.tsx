import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { Card } from '../../shared/components/atoms/Card'
import { Button } from '../../shared/components/atoms/Button'
import { questsApi } from '../../shared/lib/api'
import type { QuestView } from '../../shared/types'

export function QuestsPage() {
  const [quests, setQuests] = useState<QuestView[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const loadQuests = async () => {
      try {
        const data = await questsApi.list()
        setQuests(data)
      } catch (e) {
        setError(e instanceof Error ? e.message : 'failed to load quests')
      } finally {
        setLoading(false)
      }
    }
    loadQuests()
  }, [])

  if (loading) {
    return (
      <div className="flex h-64 w-full items-center justify-center">
        <div className="flex flex-col items-center gap-4 animate-pulse">
          <div className="text-4xl">📜</div>
          <p className="text-sm text-text-secondary">Unrolling the scrolls...</p>
        </div>
      </div>
    )
  }

  const activeQuests = quests.filter(q => q.status === 'ACTIVE')
  const pendingQuests = quests.filter(q => q.status === 'PENDING')
  const completedQuests = quests.filter(q => q.status === 'DONE')

  const QuestList = ({ title, list, emptyMsg }: { title: string, list: QuestView[], emptyMsg: string }) => (
    <section className="mb-10">
      <h2 className="font-heading text-2xl text-text-primary mb-4 border-b border-border-subtle pb-2">{title}</h2>
      {list.length === 0 ? (
        <Card className="text-center py-8 opacity-60 bg-transparent border-dashed">
          <p className="text-text-secondary">{emptyMsg}</p>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {list.map(quest => (
            <Card key={quest.id} hoverable className="flex flex-col">
              <div className="flex justify-between items-start mb-3">
                <div>
                  <h3 className="font-heading text-xl text-text-primary">{quest.title}</h3>
                  <p className="text-xs text-accent-magic uppercase tracking-wider">{quest.template_slug.replace(/-/g, ' ')}</p>
                </div>
                {quest.status === 'ACTIVE' && <span className="text-xs font-bold bg-accent-magic/20 text-accent-magic px-2 py-1 rounded">ACTIVE</span>}
                {quest.status === 'DONE' && <span className="text-xs font-bold bg-accent-nature/20 text-accent-nature px-2 py-1 rounded">DONE</span>}
              </div>
              
              <div className="flex items-center gap-2 mb-6">
                <div className="h-1.5 flex-1 bg-surface-elevated rounded-full overflow-hidden">
                  <div 
                    className={`h-full rounded-full transition-all ${quest.status === 'DONE' ? 'bg-accent-nature' : 'bg-accent-magic'}`}
                    style={{ width: `${quest.challenge_count > 0 ? (quest.completed_count / quest.challenge_count) * 100 : 0}%` }}
                  />
                </div>
                <span className="text-xs text-text-secondary whitespace-nowrap">
                  {quest.completed_count} / {quest.challenge_count}
                </span>
              </div>
              
              <div className="mt-auto">
                <Link to={`/quests/${quest.id}`} className="block">
                  <Button variant={quest.status === 'ACTIVE' ? 'primary' : 'secondary'} className="w-full">
                    {quest.status === 'DONE' ? 'Review Quest' : 'View Details'}
                  </Button>
                </Link>
              </div>
            </Card>
          ))}
        </div>
      )}
    </section>
  )

  return (
    <div className="max-w-5xl mx-auto flex flex-col gap-6">
      <header className="mb-6">
        <h1 className="font-heading text-4xl text-text-primary mb-2">Quests</h1>
        <p className="text-text-secondary">Your active missions and past triumphs.</p>
      </header>

      {error && (
        <div className="bg-accent-danger/10 border border-accent-danger/20 p-4 rounded-lg mb-6">
          <p className="text-sm text-accent-danger">{error}</p>
        </div>
      )}

      <QuestList title="Active Adventures" list={activeQuests} emptyMsg="No active quests right now. Time for a rest!" />
      <QuestList title="Available Quests" list={pendingQuests} emptyMsg="No new quests available." />
      <QuestList title="Completed" list={completedQuests} emptyMsg="You haven't completed any quests yet." />
    </div>
  )
}
