import { useState, useEffect } from 'react'
import { useParams, Link, Navigate, useNavigate } from 'react-router-dom'
import { apiClient } from '../../shared/lib/api'
import type { InventoryItem } from '../../shared/types'
import { Card } from '../../shared/components/atoms/Card'
import { useSession } from '../../shared/hooks/useSession'
import { Button } from '../../shared/components/atoms/Button'

export function CollectionDetailPage() {
  const { slug } = useParams<{ slug: string }>()
  const [relic, setRelic] = useState<InventoryItem | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const { profile } = useSession()
  const navigate = useNavigate()

  useEffect(() => {
    const load = async () => {
      if (!slug) return
      setLoading(true)
      setError(null)
      try {
        const data = await apiClient.get<InventoryItem[]>('/api/collections/inventory')
        const found = data.find((r) => r.collection_slug === slug)
        setRelic(found || null)
      } catch (e) {
        setError(e instanceof Error ? e.message : 'failed to load relic')
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [slug])

  if (profile?.role !== 'GUIDE') {
    return <Navigate to="/profile" replace />
  }

  if (loading) {
    return (
      <div className="flex h-64 w-full items-center justify-center max-w-4xl mx-auto">
        <div className="flex flex-col items-center justify-center h-64 opacity-50 animate-pulse">
          <span className="text-4xl mb-4">🔍</span>
          <p className="text-sm text-text-secondary">Memeriksa hadiah...</p>
        </div>
      </div>
    )
  }

  if (error || !relic) {
    return (
      <div className="flex flex-col max-w-4xl mx-auto py-8">
        <Button variant="ghost" size="sm" onClick={() => navigate('/collections')} className="mb-4 w-fit">
          ← Kembali ke Koleksi
        </Button>
        <div className="bg-accent-danger/10 border border-accent-danger/20 p-8 rounded-xl text-center max-w-md mx-auto">
          <p className="text-lg font-medium text-accent-danger">{error || 'Hadiah tidak ditemukan.'}</p>
          <Button variant="secondary" className="mt-4" onClick={() => navigate('/collections')}>
            Kembali ke Koleksi
          </Button>
        </div>
      </div>
    )
  }

  const rarityColors: Record<string, string> = {
    COMMON: 'border-border-subtle bg-surface',
    UNCOMMON: 'border-accent-nature/50 bg-accent-nature/5 shadow-[0_0_15px_rgba(16,185,129,0.1)]',
    RARE: 'border-accent-magic/50 bg-accent-magic/5 shadow-[0_0_15px_rgba(6,182,222,0.15)]',
    EPIC: 'border-accent-rare/50 bg-accent-rare/10 shadow-[0_0_15px_rgba(139,92,246,0.2)]',
    LEGENDARY: 'border-accent-reward/60 bg-accent-reward/10 shadow-[0_0_20px_rgba(245,158,11,0.25)]',
  }

  const rarityTextColors: Record<string, string> = {
    COMMON: 'text-text-secondary',
    UNCOMMON: 'text-accent-nature',
    RARE: 'text-accent-magic',
    EPIC: 'text-accent-rare',
    LEGENDARY: 'text-accent-reward',
  }

  return (
    <div className="flex flex-col gap-6 max-w-4xl mx-auto py-4">
      <Link to="/collections" className="text-sm font-medium text-text-secondary hover:text-text-primary transition-colors inline-flex items-center gap-2 w-fit">
        <span>←</span> Kembali ke Koleksi
      </Link>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-8 animate-in fade-in duration-500">
        <div className="flex flex-col gap-6">
          <Card className={`relative overflow-hidden p-0 aspect-square flex items-center justify-center ${rarityColors[relic.rarity] || rarityColors.COMMON}`}>
            <div className="absolute inset-0 bg-[radial-gradient(circle_at_center,rgba(255,255,255,0.1)_0%,transparent_100%)] pointer-events-none"></div>
            <span className="text-[120px] drop-shadow-2xl z-10 transform hover:scale-110 transition-transform duration-700 cursor-pointer">
              {relic.image}
            </span>
          </Card>
        </div>

        <div className="flex flex-col justify-center">
          <div className="mb-6">
            <div className="flex items-center gap-3 mb-2">
              <span className={`text-xs font-bold uppercase tracking-widest px-2 py-1 rounded border bg-surface-elevated ${rarityTextColors[relic.rarity] || rarityTextColors.COMMON} border-current/30`}>
                {relic.rarity}
              </span>
              <span className="text-xs text-text-secondary uppercase tracking-widest">
                {relic.journey.replace(/-/g, ' ')}
              </span>
            </div>
            
            <h1 className="font-heading text-4xl md:text-5xl text-text-primary mb-4">{relic.name}</h1>
            
            {relic.is_new && (
              <span className="inline-block bg-accent-magic text-black text-xs font-bold px-3 py-1 rounded shadow-[0_0_10px_rgba(6,182,222,0.5)] mb-4">
                PENEMUAN BARU
              </span>
            )}
          </div>

          <Card className="mb-6 p-6 bg-surface-elevated/50">
            <p className="text-lg text-text-primary leading-relaxed mb-4">{relic.description}</p>
            <p className="text-sm text-text-secondary italic border-l-2 border-accent-magic/30 pl-4 py-1">
              "{relic.concept}"
            </p>
          </Card>

          <div className="flex items-center justify-between border-t border-border-subtle pt-6 mt-auto">
            <div>
              <p className="text-xs text-text-secondary uppercase tracking-wider mb-1">Ditemukan pada</p>
              <p className="text-sm text-text-primary font-medium">{new Date(relic.discovered_at).toLocaleDateString()}</p>
            </div>
            <div className="text-right">
              <p className="text-xs text-text-secondary uppercase tracking-wider mb-1">Jumlah Dimiliki</p>
              <p className="text-sm text-text-primary font-medium bg-surface-elevated px-3 py-1 rounded inline-block">x{relic.owned_count}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
