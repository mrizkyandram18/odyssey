import { useState, useEffect } from 'react'
import { DailyTurnBanner } from '../../shared/components/molecules/DailyTurnBanner'
import { QuestCard } from '../../shared/components/molecules/QuestCard'
import { StreakBadge } from '../../shared/components/molecules/StreakBadge'
import { apiClient } from '../../shared/lib/api'
import { chestsApi } from '../../shared/lib/api'
import type { HomeResponse, QuestView, RealmProgress, ChestView } from '../../shared/types'
import { Link } from 'react-router-dom'

export function HomePage() {
  const [home, setHome] = useState<HomeResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [takingTurn, setTakingTurn] = useState(false)

  const xpPercent = home ? Math.min(100, (home.player.xp % 100)) : 0

  const takeTurn = async () => {
    setTakingTurn(true)
    setError(null)
    try {
      await apiClient.post('/api/daily_turns/consume', { quest_slug: 'daily-turn' })
      await loadHome()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'failed to take turn')
    } finally {
      setTakingTurn(false)
    }
  }

  const openChest = async (chestId: number) => {
    setError(null)
    try {
      await chestsApi.open(chestId)
      await loadHome()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'failed to open chest')
    }
  }

  const loadHome = async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await apiClient.get<HomeResponse>('/api/home')
      setHome(data)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'failed to load home')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadHome()
  }, [])

  if (loading && !home) {
    return <p className="p-4 text-sm text-muted-foreground">Loading...</p>
  }

  const greeting = `Hello, ${home?.player.explorer_name || 'Explorer'}`

  return (
    <div className="flex flex-col gap-4 p-4 pb-safe">
      <header className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">{greeting}</h1>
        <StreakBadge days={home?.daily_turn.streak_days ?? 0} />
      </header>

      <section className="flex items-center gap-3 rounded-lg border border-border bg-surface p-3">
        <div className="flex-1">
          <p className="text-xs text-muted-foreground">Level {home?.player.level ?? 1} Explorer</p>
          <div className="mt-1 h-2 w-full rounded-full bg-border">
            <div
              className="h-2 rounded-full bg-primary transition-all"
              style={{ width: `${xpPercent}%` }}
            />
          </div>
          <div className="mt-1 flex items-center justify-between text-xs text-muted-foreground">
            <span>{home?.player.xp ?? 0} XP</span>
            <span className="flex items-center gap-1">
              <span>🪙</span>
              <span>{home?.player.coins ?? 0} Coins</span>
            </span>
          </div>
        </div>
      </section>

      <DailyTurnBanner
        remaining={home?.daily_turn.remaining_turns ?? 0}
        onTakeTurn={takeTurn}
        loading={takingTurn}
      />
      {error && <p className="text-xs text-red-500">{error}</p>}

      {home?.realm_progress && home.realm_progress.length > 0 && (
        <section className="flex flex-col gap-2">
          <h2 className="text-sm font-medium text-muted-foreground">Realm Progress</h2>
          {home.realm_progress.map((realm: RealmProgress) => (
            <div key={realm.realm} className="rounded-lg border border-border bg-surface p-3">
              <div className="flex items-center justify-between">
                <h3 className="font-semibold capitalize">{realm.realm.replace(/-/g, ' ')}</h3>
                <span className="text-xs text-muted-foreground">{realm.status}</span>
              </div>
              <div className="mt-1 h-2 w-full rounded-full bg-border">
                <div
                  className="h-2 rounded-full bg-primary transition-all"
                  style={{ width: `${realm.progress}%` }}
                />
              </div>
              <p className="mt-1 text-xs text-muted-foreground">{realm.progress}% complete</p>
            </div>
          ))}
        </section>
      )}

      {home?.available_chests && home.available_chests.length > 0 && (
        <section className="flex flex-col gap-2">
          <h2 className="text-sm font-medium text-muted-foreground">Available Chests</h2>
          {home.available_chests.map((chest: ChestView) => (
            <button
              key={chest.id}
              onClick={() => openChest(chest.id)}
              className="rounded-lg border border-border bg-surface p-3 text-left transition hover:border-accent"
            >
              <div className="flex items-center gap-3">
                <span className="text-2xl">{chest.icon}</span>
                <div className="flex-1">
                  <p className="font-medium">{chest.name}</p>
                  <p className="text-xs text-muted-foreground">{chest.description}</p>
                </div>
                <span className="text-xs text-accent">Open</span>
              </div>
            </button>
          ))}
        </section>
      )}

      {home?.latest_relic && (
        <section className="flex flex-col gap-2">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-medium text-muted-foreground">Latest Relic</h2>
            <Link to="/relics" className="text-xs text-accent">View all</Link>
          </div>
          <div className="rounded-lg border border-border bg-surface p-3">
            <div className="flex items-center gap-3">
              <span className="text-3xl">{home.latest_relic.image}</span>
              <div className="flex-1">
                <p className="font-medium">{home.latest_relic.name}</p>
                <p className="text-xs text-muted-foreground">{home.latest_relic.description}</p>
              </div>
            </div>
          </div>
        </section>
      )}

      <section className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-medium text-muted-foreground">Collection Progress</h2>
          <Link to="/relics" className="text-xs text-accent">View all</Link>
        </div>
        <div className="rounded-lg border border-border bg-surface p-3">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium">
              {home?.collection_progress.collected ?? 0} / {home?.collection_progress.total ?? 0}
            </span>
            <span className="text-xs text-muted-foreground">Relics</span>
          </div>
          <div className="mt-2 h-2 w-full rounded-full bg-border">
            <div
              className="h-2 rounded-full bg-accent transition-all"
              style={{
                width: home?.collection_progress.total
                  ? `${(home.collection_progress.collected / home.collection_progress.total) * 100}%`
                  : '0%',
              }}
            />
          </div>
        </div>
      </section>

      {home?.completed_quests_today && home.completed_quests_today.length > 0 && (
        <section className="flex flex-col gap-2">
          <h2 className="text-sm font-medium text-muted-foreground">Completed Today</h2>
          {home.completed_quests_today.map((q: QuestView) => (
            <QuestCard key={q.id} quest={q} />
          ))}
        </section>
      )}

      {(home?.pending_creative_review ?? 0) > 0 && (
        <section className="flex flex-col gap-2">
          <h2 className="text-sm font-medium text-muted-foreground">Pending Creative Review</h2>
          <p className="text-sm text-muted-foreground">
            {home!.pending_creative_review} submission{home!.pending_creative_review === 1 ? '' : 's'} awaiting review
          </p>
        </section>
      )}

      {home?.last_submission && (
        <section className="flex flex-col gap-2">
          <h2 className="text-sm font-medium text-muted-foreground">Last Submission</h2>
          <div className="rounded-lg border border-border bg-surface p-3">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">{home.last_submission.kind}</span>
              <span className="text-xs text-muted-foreground">{home.last_submission.status}</span>
            </div>
            <p className="mt-1 text-sm text-muted-foreground line-clamp-2">
              {home.last_submission.content}
            </p>
          </div>
        </section>
      )}

      <section className="flex flex-col gap-2">
        <h2 className="text-sm font-medium text-muted-foreground">Active Quests</h2>
        {loading ? (
          <p className="text-sm text-muted-foreground">Loading quests...</p>
        ) : (
          home?.active_quests && home.active_quests.length > 0 ? (
            home.active_quests.map((q: QuestView) => (
              <QuestCard key={q.id} quest={q} isMyTurn={q.active_challenge_assigned_to === home?.player.uid} />
            ))
          ) : (
            <p className="text-sm text-muted-foreground">No active quests. Check back soon!</p>
          )
        )}
      </section>

      <section className="mt-4 flex gap-2">
        <Link to="/creative" className="flex-1 rounded-lg bg-surface border border-border p-3 text-center transition-colors hover:border-primary">
          <p className="text-sm font-semibold text-primary">Family Journal</p>
          <p className="text-xs text-muted-foreground mt-1">Read your stories</p>
        </Link>
        <Link to="/journal" className="flex-1 rounded-lg bg-surface border border-border p-3 text-center transition-colors hover:border-accent">
          <p className="text-sm font-semibold text-accent">Milestones</p>
          <p className="text-xs text-muted-foreground mt-1">Lore & Achievements</p>
        </Link>
      </section>
    </div>
  )
}
