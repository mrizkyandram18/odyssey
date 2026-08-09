import { useState, useEffect } from 'react'

import { apiClient } from '../../shared/lib/api'
import type { InventoryItem } from '../../shared/types'
import { RelicGrid } from './components/RelicGrid'
import { Card } from '../../shared/components/atoms/Card'

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
    return (
      <div className="flex h-64 w-full items-center justify-center max-w-6xl mx-auto">
        <div className="flex flex-col items-center gap-4 animate-pulse">
          <div className="text-4xl">✨</div>
          <p className="text-sm text-text-secondary">Unlocking the vault...</p>
        </div>
      </div>
    )
  }

  const totalCollected = relics.reduce((acc, curr) => acc + curr.owned_count, 0)

  return (
    <div className="flex flex-col gap-8 max-w-6xl mx-auto py-4">
      <header className="flex flex-col gap-2 mb-4">
        <h1 className="font-heading text-4xl md:text-5xl text-text-primary">Collection</h1>
        <p className="text-text-secondary text-lg">Treasures and artifacts discovered across the realms.</p>
      </header>

      {/* Stats Summary */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <Card className="text-center p-4">
          <p className="text-3xl font-heading text-accent-reward mb-1">{relics.length}</p>
          <p className="text-xs text-text-secondary uppercase tracking-wider font-bold">Unique Relics</p>
        </Card>
        <Card className="text-center p-4">
          <p className="text-3xl font-heading text-accent-magic mb-1">{totalCollected}</p>
          <p className="text-xs text-text-secondary uppercase tracking-wider font-bold">Total Collected</p>
        </Card>
      </div>

      {error && (
        <div className="bg-accent-danger/10 border border-accent-danger/30 p-4 rounded-lg">
          <p className="text-sm font-medium text-accent-danger">{error}</p>
        </div>
      )}

      {relics.length === 0 && !error ? (
        <Card className="text-center py-16 opacity-60 border-dashed border-border-subtle bg-transparent">
          <span className="text-6xl mb-4 block">🗝️</span>
          <p className="text-text-secondary text-lg">The vault is empty.</p>
          <p className="text-sm text-text-secondary">Complete quests and open chests to discover relics!</p>
        </Card>
      ) : (
        <RelicGrid relics={relics} />
      )}
    </div>
  )
}
