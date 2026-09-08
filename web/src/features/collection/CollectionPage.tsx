import { useEffect, useState, useMemo } from 'react'
import { Link } from 'react-router-dom'
import { Gift, Check, Lock, ArrowLeft } from 'lucide-react'
import { Card } from '../../shared/components/atoms/Card'
import { Button } from '../../shared/components/atoms/Button'
import { Avatar } from '../../shared/components/atoms/Avatar'
import { rewardsApi } from '../../shared/lib/api'
import { useSession } from '../../shared/hooks/useSession'
import { CapsuleModal } from './CapsuleModal'
import type { CollectionItem } from '../../shared/types'

const COSMETIC_NAMES: Record<string, string> = {
  gold: 'Emas Legendaris',
  sparkle: 'Kilau Bintang',
  float: 'Melayang Tenang',
  trail: 'Jejak Petualang',
}

export function CollectionPage() {
  const { profile, refreshProfile } = useSession()
  const [items, setItems] = useState<CollectionItem[]>([])
  const [loading, setLoading] = useState(true)
  const [showCapsule, setShowCapsule] = useState(false)
  const [tickets, setTickets] = useState(0)
  const [filterSlot, setFilterSlot] = useState<'all' | 'frame' | 'effect'>('all')

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

  const filteredItems = useMemo(() => {
    if (filterSlot === 'all') return items
    return items.filter((it) => it.slot === filterSlot)
  }, [items, filterSlot])

  const equippedFrameItem = items.find((i) => i.slot === 'frame' && i.equipped)
  const equippedEffectItem = items.find((i) => i.slot === 'effect' && i.equipped)

  return (
    <div className="w-full flex flex-col gap-4 max-w-xl mx-auto pb-6">
      {/* Header */}
      <header className="flex items-center justify-between px-1">
        <div className="flex items-center gap-2">
          <Link
            to="/"
            className="p-2 -ml-2 rounded-xl hover:bg-surface text-text-secondary hover:text-text-primary transition-colors"
            aria-label="Kembali ke Beranda"
          >
            <ArrowLeft size={18} />
          </Link>
          <div>
            <h1 className="text-xl font-extrabold text-text-primary">🎨 Koleksi Saya</h1>
            <p className="text-xs text-text-secondary mt-0.5">
              Hiasan profil dari Kotak Kejutan. Tiket Hadiah tersisa: {tickets}
            </p>
          </div>
        </div>

        {/* Ticket indicator */}
        <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-2xl bg-amber-500/10 border border-amber-500/25 text-amber-700 dark:text-amber-300 font-extrabold text-xs shadow-xs">
          <span>🎁</span>
          <span>{tickets} Tiket</span>
        </div>
      </header>

      {/* 1. Hero Avatar Showcase (What I'm currently wearing) */}
      <div className="relative overflow-hidden rounded-3xl border border-border-subtle bg-gradient-to-b from-accent-magic/10 via-surface to-surface p-6 text-center shadow-sm">
        <div className="flex flex-col items-center">
          <div className="relative p-2">
            <Avatar
              seed={profile?.avatar_seed ?? 'koleksi'}
              frame={profile?.avatar_frame || 'none'}
              effect={profile?.equipped_explorer_effect || 'none'}
              size="2xl"
            />
          </div>

          <h2 className="text-lg font-extrabold text-text-primary mt-2">
            {profile?.explorer_name || 'Petualang'}
          </h2>
          <p className="text-xs font-bold text-accent-magic">
            ⭐ Tingkat {profile?.level || 1}
          </p>

          {/* Equipped Badges */}
          <div className="mt-3 flex items-center justify-center gap-2 flex-wrap">
            <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-surface border border-border-subtle text-[11px] font-bold text-text-secondary shadow-xs">
              <span>🖼️ Bingkai:</span>
              <span className="text-text-primary font-extrabold">
                {equippedFrameItem
                  ? (COSMETIC_NAMES[equippedFrameItem.asset] || equippedFrameItem.asset)
                  : 'Standar'}
              </span>
            </span>
            <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-surface border border-border-subtle text-[11px] font-bold text-text-secondary shadow-xs">
              <span>✨ Efek:</span>
              <span className="text-text-primary font-extrabold">
                {equippedEffectItem
                  ? (COSMETIC_NAMES[equippedEffectItem.asset] || equippedEffectItem.asset)
                  : 'Standar'}
              </span>
            </span>
          </div>
        </div>
      </div>

      {/* 2. Open Reward CTA Ticket Banner */}
      <div className="rounded-2xl border-2 border-amber-400/40 bg-gradient-to-r from-amber-500/10 via-amber-400/5 to-transparent p-4 flex items-center justify-between gap-3 shadow-xs">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-amber-400 to-amber-500 text-white flex items-center justify-center text-xl shadow-md shadow-amber-400/20 shrink-0">
            🎁
          </div>
          <div>
            <p className="text-xs font-black uppercase tracking-wider text-amber-700 dark:text-amber-300">
              {tickets > 0 ? `${tickets} Tiket Hadiah Tersedia` : 'Buka Kotak Kejutan'}
            </p>
            <p className="text-xs text-text-secondary mt-0.5">
              {tickets > 0
                ? 'Buka kotak kejutan untuk mendapatkan bingkai & efek baru.'
                : 'Selesaikan tugas harian untuk mengumpulkan tiket.'}
            </p>
          </div>
        </div>

        <Button
          size="md"
          className="bg-gradient-to-r from-amber-500 to-orange-500 hover:from-amber-600 hover:to-orange-600 text-white font-black shadow-md shadow-amber-500/25 px-4 shrink-0 min-h-[40px]"
          onClick={() => setShowCapsule(true)}
        >
          <Gift size={16} className="mr-1" /> Buka Hadiah
        </Button>
      </div>

      {/* 3. Category Filter Tabs */}
      <div className="flex p-1 bg-surface rounded-2xl border border-border-subtle gap-1">
        <button
          onClick={() => setFilterSlot('all')}
          className={`flex-1 py-2 rounded-xl text-xs font-bold transition-all ${
            filterSlot === 'all'
              ? 'bg-accent-magic text-white shadow-sm'
              : 'text-text-secondary hover:text-text-primary hover:bg-surface-elevated'
          }`}
        >
          Semua Koleksi ({items.length})
        </button>
        <button
          onClick={() => setFilterSlot('frame')}
          className={`flex-1 py-2 rounded-xl text-xs font-bold transition-all ${
            filterSlot === 'frame'
              ? 'bg-accent-magic text-white shadow-sm'
              : 'text-text-secondary hover:text-text-primary hover:bg-surface-elevated'
          }`}
        >
          🖼️ Bingkai
        </button>
        <button
          onClick={() => setFilterSlot('effect')}
          className={`flex-1 py-2 rounded-xl text-xs font-bold transition-all ${
            filterSlot === 'effect'
              ? 'bg-accent-magic text-white shadow-sm'
              : 'text-text-secondary hover:text-text-primary hover:bg-surface-elevated'
          }`}
        >
          ✨ Efek
        </button>
      </div>

      {/* 4. Collection Grid */}
      {loading ? (
        <Card className="p-8 text-center text-xs text-text-secondary">
          <div className="animate-pulse flex flex-col items-center gap-2">
            <span className="text-2xl">🎨</span>
            <span>Memuat koleksi hadiah...</span>
          </div>
        </Card>
      ) : filteredItems.length === 0 ? (
        <div className="rounded-3xl border-2 border-dashed border-border-subtle bg-surface/60 p-8 text-center">
          <span className="text-3xl">🎁</span>
          <p className="text-sm font-extrabold text-text-primary mt-2">Belum ada koleksi</p>
          <p className="mt-1 text-xs text-text-secondary max-w-[28ch] mx-auto leading-relaxed">
            Selesaikan tugas harian untuk mendapatkan Tiket Hadiah dan buka kotak pertamamu!
          </p>
          <Link
            to="/"
            className="mt-4 inline-flex items-center gap-1.5 px-4 py-2 rounded-xl bg-accent-magic text-white font-bold text-xs shadow-sm hover:brightness-110 transition-colors"
          >
            Lihat Tugas Hari Ini →
          </Link>
        </div>
      ) : (
        <div className="grid grid-cols-2 gap-3">
          {filteredItems.map((it) => {
            const isOwned = it.owned ?? (it.tier > 0)
            const friendlyName = COSMETIC_NAMES[it.asset] || (it.slot === 'frame' ? 'Bingkai' : 'Efek')
            const stars = it.tier > 0 ? '★'.repeat(Math.min(3, it.tier)) : ''

            return (
              <div
                key={it.cosmetic_id}
                className={`flex flex-col items-center justify-between rounded-2xl p-4 text-center border transition-all duration-200 ${
                  it.equipped
                    ? 'border-emerald-500/50 bg-emerald-500/[0.04] shadow-sm ring-2 ring-emerald-500/20'
                    : isOwned
                    ? 'border-border-subtle bg-surface hover:border-accent-magic/40 shadow-xs'
                    : 'border-border-subtle/60 bg-surface/50 opacity-60'
                }`}
              >
                {/* Avatar Preview */}
                <div className="relative my-1">
                  <Avatar
                    seed={profile?.avatar_seed ?? 'koleksi'}
                    frame={it.slot === 'frame' ? it.asset : 'none'}
                    effect={it.slot === 'effect' ? it.asset : 'none'}
                    size="lg"
                  />
                  {!isOwned && (
                    <span className="absolute -top-1 -right-1 flex h-6 w-6 items-center justify-center rounded-full bg-zinc-800 text-[10px] text-white shadow-sm ring-2 ring-white">
                      <Lock size={11} />
                    </span>
                  )}
                  {it.equipped && (
                    <span className="absolute -top-1 -right-1 flex h-6 w-6 items-center justify-center rounded-full bg-emerald-500 text-[10px] text-white shadow-sm ring-2 ring-white">
                      <Check size={12} strokeWidth={3} />
                    </span>
                  )}
                </div>

                {/* Cosmetic Details */}
                <div className="mt-2 space-y-0.5">
                  <p className="text-xs font-extrabold text-text-primary">
                    {it.slot === 'frame' ? 'Bingkai' : 'Efek'}
                    {stars ? ` ${stars}` : ''}
                  </p>
                  <p className="text-[11px] font-semibold text-accent-magic truncate max-w-[130px]">
                    {friendlyName}
                  </p>
                </div>

                {/* Equip / Status Action */}
                <div className="mt-3 w-full">
                  {!isOwned ? (
                    <span className="inline-block text-[11px] font-bold text-text-secondary py-1">
                      Belum dimiliki
                    </span>
                  ) : it.equipped ? (
                    <span className="inline-flex items-center justify-center gap-1 w-full py-1.5 rounded-xl bg-emerald-500/15 text-emerald-700 dark:text-emerald-300 text-[11px] font-black">
                      ✓ Terpasang
                    </span>
                  ) : (
                    <Button
                      variant="secondary"
                      size="sm"
                      className="w-full text-xs font-bold min-h-[34px]"
                      onClick={() => void handleEquip(it.cosmetic_id)}
                    >
                      Pasang di Profil
                    </Button>
                  )}
                </div>
              </div>
            )
          })}
        </div>
      )}

      {showCapsule && (
        <CapsuleModal
          onClose={() => {
            setShowCapsule(false)
            void reload()
          }}
          onOpened={() => void reload()}
        />
      )}
    </div>
  )
}
