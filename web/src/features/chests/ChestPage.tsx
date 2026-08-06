import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { chestsApi } from '../../shared/lib/api'
import type { ChestView } from '../../shared/types'

export function ChestPage() {
  const [chests, setChests] = useState<ChestView[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const loadChests = async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await chestsApi.list()
      setChests(data)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'failed to load chests')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadChests()
  }, [])

  const openChest = async (id: number) => {
    try {
      await chestsApi.open(id)
      await loadChests()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'failed to open chest')
    }
  }

  if (loading) {
    return <p className="p-4 text-sm text-muted-foreground">Loading chests...</p>
  }

  return (
    <div className="flex flex-col gap-4 p-4 pb-safe">
      <header className="flex items-center gap-2">
        <Link to="/" className="text-sm text-muted-foreground">Home</Link>
        <h1 className="text-xl font-semibold">Chests</h1>
      </header>

      {error && <p className="text-xs text-red-500">{error}</p>}

      {chests.length === 0 && (
        <p className="text-sm text-muted-foreground">No chests yet. Complete quests to earn chests!</p>
      )}

      <div className="flex flex-col gap-3">
        {chests.map((chest) => (
          <div
            key={chest.id}
            className={`rounded-lg border border-border bg-surface p-4 ${chest.opened ? 'opacity-60' : ''}`}
          >
            <div className="flex items-center gap-3">
              <span className="text-3xl">{chest.icon}</span>
              <div className="flex-1">
                <p className="font-medium">{chest.name}</p>
                <p className="text-xs text-muted-foreground">{chest.description}</p>
                <p className="text-xs text-muted-foreground mt-1">Rarity: {chest.rarity}</p>
              </div>
            </div>
            {!chest.opened && (
              <button
                onClick={() => openChest(chest.id)}
                className="mt-3 w-full rounded-lg bg-accent py-2 text-sm font-medium text-background transition hover:opacity-90"
              >
                Open Chest
              </button>
            )}
            {chest.opened && (
              <p className="mt-3 text-xs text-muted-foreground">Opened</p>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
