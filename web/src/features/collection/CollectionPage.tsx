import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Gift } from 'lucide-react'
import { Card } from '../../shared/components/atoms/Card'
import { Button } from '../../shared/components/atoms/Button'
import { Avatar } from '../../shared/components/atoms/Avatar'
import { rewardsApi } from '../../shared/lib/api'
import { useSession } from '../../shared/hooks/useSession'
import { CapsuleModal } from './CapsuleModal'
import type { CollectionItem } from '../../shared/types'

export function CollectionPage() {
  const { profile, refreshProfile } = useSession()
  const [items, setItems] = useState<CollectionItem[]>([])
  const [loading, setLoading] = useState(true)
  const [showCapsule, setShowCapsule] = useState(false)
  const [tickets, setTickets] = useState(0)

  const reload = async () => {
    setLoading(true)
    try {
      const [col, st] = await Promise.all([rewardsApi.collection(), rewardsApi.status()])
      setItems(col.items ?? [])
      setTickets(st.tickets ?? 0)
    } catch {
      setItems([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void reload()
  }, [])

  const handleEquip = async (id: string) => {
    try {
      await rewardsApi.equip(id)
      if (refreshProfile) void refreshProfile()
      await reload()
    } catch {
      // biarkan list apa adanya; error kecil tidak memblokir halaman
    }
  }

  return (
    <div className="mx-auto flex w-full max-w-md flex-col gap-4 px-4 py-6">
      <div className="text-center">
        <h1 className="text-xl font-extrabold text-text-primary">🎨 Koleksi Saya</h1>
        <p className="mt-1 text-xs text-text-secondary">Hiasan profil dari Kotak Kejutan. Tiket Hadiah tersisa: {tickets}</p>
      </div>

      <Button size="lg" className="w-full" onClick={() => setShowCapsule(true)}>
        <Gift size={16} /> Buka Hadiah
      </Button>

      {loading ? (
        <Card className="p-5 text-center text-xs text-text-secondary">Memuat koleksi…</Card>
      ) : items.length === 0 ? (
        <Card className="p-5 text-center">
          <p className="text-sm font-bold text-text-primary">Belum ada koleksi</p>
          <p className="mt-1 text-xs text-text-secondary">Selesaikan tugas harian untuk mendapatkan Tiket Hadiah.</p>
          <Link to="/" className="mt-3 inline-block text-xs font-bold text-accent-magic hover:underline">
            Lihat Tugas Hari Ini
          </Link>
        </Card>
      ) : (
        <div className="grid grid-cols-2 gap-3">
          {items.map((it) => {
            const isOwned = it.owned ?? (it.tier > 0)
            const stars = it.tier > 0 ? '★'.repeat(Math.min(3, it.tier)) : ''
            return (
              <Card
                key={it.cosmetic_id}
                className={`flex flex-col items-center gap-2 p-4 text-center ${!isOwned ? 'opacity-60 border-dashed bg-surface/50' : ''}`}
              >
                <div className="relative">
                  <Avatar
                    seed={profile?.avatar_seed ?? 'koleksi'}
                    frame={it.slot === 'frame' ? it.asset : 'none'}
                    effect={it.slot === 'effect' ? it.asset : 'none'}
                    size="lg"
                  />
                  {!isOwned && (
                    <span className="absolute -top-1 -right-1 flex h-5 w-5 items-center justify-center rounded-full bg-zinc-700 text-[10px] text-white">
                      🔒
                    </span>
                  )}
                </div>
                <p className="text-xs font-bold text-text-primary">
                  {it.slot === 'frame' ? 'Bingkai' : 'Efek'}
                  {stars ? ` ${stars}` : ''}
                </p>
                {!isOwned ? (
                  <span className="text-[11px] font-medium text-text-secondary">Belum dimiliki</span>
                ) : it.equipped ? (
                  <span className="text-[11px] font-bold text-emerald-600">✓ Terpasang</span>
                ) : (
                  <Button variant="secondary" size="sm" onClick={() => void handleEquip(it.cosmetic_id)}>
                    Pasang di Profil
                  </Button>
                )}
              </Card>
            )
          })}
        </div>
      )}

      {showCapsule && <CapsuleModal onClose={() => { setShowCapsule(false); void reload() }} onOpened={() => void reload()} />}
    </div>
  )
}
