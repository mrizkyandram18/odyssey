import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { apiClient } from '../../shared/lib/api'
import type { InventoryItem } from '../../shared/types'
import { RelicGrid } from './components/RelicGrid'

export function RelicInventoryPage() {
  const [relics, setRelics] = useState<InventoryItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const load = async () => {
      setLoading(true)
      setError(null)
      try {
        const data = await apiClient.get<InventoryItem[]>('/api/relics/inventory')
        setRelics(data)
      } catch (e) {
        setError(e instanceof Error ? e.message : 'failed to load relics')
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [])

  if (loading) {
    return <p className="p-4 text-sm text-muted-foreground">Loading relics...</p>
  }

  return (
    <div className="flex flex-col gap-4 p-4 pb-safe">
      <header className="flex items-center gap-2">
        <Link to="/" className="text-sm text-muted-foreground">Home</Link>
        <h1 className="text-xl font-semibold">Relic Collection</h1>
      </header>

      {error && <p className="text-xs text-red-500">{error}</p>}

      {relics.length === 0 && (
        <p className="text-sm text-muted-foreground">No relics yet. Open chests to discover relics!</p>
      )}

      <RelicGrid relics={relics} />
    </div>
  )
}
