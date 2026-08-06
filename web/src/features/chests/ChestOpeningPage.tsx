import { useState, useEffect } from 'react'
import { Link, useParams } from 'react-router-dom'
import { chestsApi } from '../../shared/lib/api'
import type { OpenResult } from '../../shared/types'
import { RelicCard } from '../relics/components/RelicCard'

export function ChestOpeningPage() {
  const { chestId } = useParams<{ chestId: string }>()
  const [result, setResult] = useState<OpenResult | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const open = async () => {
      if (!chestId) return
      setLoading(true)
      setError(null)
      try {
        const data = await chestsApi.open(Number(chestId))
        setResult(data)
      } catch (e) {
        setError(e instanceof Error ? e.message : 'failed to open chest')
      } finally {
        setLoading(false)
      }
    }
    open()
  }, [chestId])

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center gap-4 p-4 pb-safe">
        <div className="text-4xl animate-pulse">📦</div>
        <p className="text-sm text-muted-foreground">Opening chest...</p>
      </div>
    )
  }

  if (error || !result) {
    return (
      <div className="flex flex-col gap-4 p-4 pb-safe">
        <Link to="/chests" className="text-sm text-muted-foreground">Back to Chests</Link>
        <p className="text-sm text-red-500">{error || 'Failed to open chest'}</p>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4 p-4 pb-safe">
      <Link to="/chests" className="text-sm text-muted-foreground">Back to Chests</Link>

      <div className="flex flex-col items-center gap-2 rounded-lg border border-border bg-surface p-6">
        <span className="text-5xl animate-bounce">{result.chest.icon}</span>
        <h1 className="text-xl font-semibold">{result.chest.name}</h1>
        <p className="text-sm text-muted-foreground">{result.chest.description}</p>
      </div>

      <div className="flex flex-col gap-2">
        <h2 className="text-sm font-medium text-muted-foreground">Rewards</h2>
        {result.rewards.map((reward, idx) => (
          <RelicCard
            key={idx}
            relic={{
              relic_id: 0,
              relic_slug: reward.relic_slug,
              name: reward.name,
              description: '',
              realm: '',
              rarity: reward.rarity,
              image: '',
              lore: '',
              owned_count: 1,
              is_new: reward.is_new,
              discovered_at: new Date().toISOString(),
              created_at: new Date().toISOString(),
            }}
          />
        ))}
      </div>

      <div className="flex items-center justify-between rounded-lg border border-border bg-surface p-3">
        <span className="text-sm text-muted-foreground">New discoveries</span>
        <span className="text-sm font-medium text-accent">{result.new_count}</span>
      </div>
      <div className="flex items-center justify-between rounded-lg border border-border bg-surface p-3">
        <span className="text-sm text-muted-foreground">Duplicates</span>
        <span className="text-sm font-medium">{result.duplicate_count}</span>
      </div>

      <Link
        to="/relics"
        className="w-full rounded-lg bg-accent py-3 text-center text-sm font-medium text-background transition hover:opacity-90"
      >
        View Collection
      </Link>
    </div>
  )
}
