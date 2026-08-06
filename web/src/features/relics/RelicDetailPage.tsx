import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { apiClient } from '../../shared/lib/api'
import type { InventoryItem } from '../../shared/types'

export function RelicDetailPage() {
  const { slug } = useParams<{ slug: string }>()
  const [relic, setRelic] = useState<InventoryItem | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const load = async () => {
      if (!slug) return
      setLoading(true)
      setError(null)
      try {
        const data = await apiClient.get<InventoryItem[]>('/api/relics/inventory')
        const found = data.find((r) => r.relic_slug === slug)
        setRelic(found || null)
      } catch (e) {
        setError(e instanceof Error ? e.message : 'failed to load relic')
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [slug])

  if (loading) {
    return <p className="p-4 text-sm text-muted-foreground">Loading relic...</p>
  }

  if (error || !relic) {
    return (
      <div className="flex flex-col gap-4 p-4 pb-safe">
        <Link to="/relics" className="text-sm text-muted-foreground">Back to Collection</Link>
        <p className="text-sm text-red-500">{error || 'Relic not found'}</p>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4 p-4 pb-safe">
      <Link to="/relics" className="text-sm text-muted-foreground">Back to Collection</Link>

      <div className="flex flex-col items-center gap-3 rounded-lg border border-border bg-surface p-6">
        <span className="text-6xl">{relic.image}</span>
        <h1 className="text-xl font-semibold">{relic.name}</h1>
        <span className="text-xs text-muted-foreground">{relic.realm.replace(/-/g, ' ')}</span>
      </div>

      <div className="rounded-lg border border-border bg-surface p-4">
        <p className="text-sm">{relic.description}</p>
        <p className="mt-3 text-xs text-muted-foreground italic">{relic.lore}</p>
      </div>

      <div className="flex items-center justify-between rounded-lg border border-border bg-surface p-3">
        <span className="text-sm text-muted-foreground">Owned</span>
        <span className="text-sm font-medium">{relic.owned_count}</span>
      </div>

      {relic.is_new && (
        <div className="rounded-lg border border-accent bg-accent/10 p-3 text-center">
          <span className="text-sm font-medium text-accent">New Discovery!</span>
        </div>
      )}
    </div>
  )
}
