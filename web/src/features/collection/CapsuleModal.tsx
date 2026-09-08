import { useState } from 'react'
import confetti from 'canvas-confetti'
import { X, Sparkles, Check, Gift } from 'lucide-react'
import { Card } from '../../shared/components/atoms/Card'
import { Button } from '../../shared/components/atoms/Button'
import { Avatar } from '../../shared/components/atoms/Avatar'
import { rewardsApi } from '../../shared/lib/api'
import { useSession } from '../../shared/hooks/useSession'
import type { OpenRewardResponse } from '../../shared/types'

interface CapsuleModalProps {
  onClose: () => void
  onOpened?: (res: OpenRewardResponse) => void
}

const SLOT_LABEL: Record<string, string> = {
  frame: 'Bingkai Profil',
  effect: 'Efek Profil',
}

const COSMETIC_NAMES: Record<string, string> = {
  gold: 'Emas Legendaris',
  sparkle: 'Kilau Bintang',
  float: 'Melayang Tenang',
  trail: 'Jejak Petualang',
}

export function CapsuleModal({ onClose, onOpened }: CapsuleModalProps) {
  const { profile, refreshProfile } = useSession()
  const [opening, setOpening] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [result, setResult] = useState<OpenRewardResponse | null>(null)
  const [equipping, setEquipping] = useState(false)
  const [equipped, setEquipped] = useState(false)

  const handleOpen = async () => {
    setOpening(true)
    setError(null)
    try {
      const res = await rewardsApi.openReward()
      setResult(res)
      onOpened?.(res)
      void confetti({
        particleCount: 100,
        spread: 80,
        origin: { y: 0.55 },
        colors: ['#F59E0B', '#38BDF8', '#10B981', '#EC4899', '#8B5CF6'],
      })
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Gagal membuka hadiah')
    } finally {
      setOpening(false)
    }
  }

  const handleEquip = async () => {
    if (!result) return
    setEquipping(true)
    try {
      await rewardsApi.equip(result.cosmetic_id)
      setEquipped(true)
      if (refreshProfile) void refreshProfile()
    } catch {
      setError('Gagal memasang hadiah di profil')
    } finally {
      setEquipping(false)
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm px-4 animate-fadeIn"
      role="dialog"
      aria-modal="true"
      aria-label="Buka Hadiah"
    >
      <Card className="w-full max-w-sm p-6 overflow-hidden relative border-2 border-amber-400/30 shadow-2xl bg-surface-elevated">
        {/* Top Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-xl">🎁</span>
            <h3 className="text-base font-extrabold text-text-primary tracking-tight">
              Buka Hadiah Petualangan
            </h3>
          </div>
          <button
            onClick={onClose}
            aria-label="Tutup"
            className="w-8 h-8 rounded-full bg-surface text-text-secondary hover:text-text-primary flex items-center justify-center transition-colors"
          >
            <X size={18} />
          </button>
        </div>

        {!result ? (
          /* 1. Anticipation Screen */
          <div className="mt-5 flex flex-col items-center gap-4 text-center">
            <div className="relative">
              {/* Glowing Aura */}
              <div className="absolute inset-0 rounded-3xl bg-amber-400/20 blur-xl animate-pulse" />
              <div className="relative flex h-24 w-24 items-center justify-center rounded-3xl bg-gradient-to-br from-amber-400 via-amber-500 to-orange-500 text-white shadow-xl shadow-amber-500/30 border-2 border-amber-300/60">
                <span className="text-4xl animate-bounce" style={{ animationDuration: '2s' }}>
                  🎁
                </span>
              </div>
            </div>

            <div className="space-y-1">
              <h4 className="font-extrabold text-text-primary text-sm">
                Kotak Kejutan Petualang
              </h4>
              <p className="text-xs text-text-secondary leading-relaxed px-2">
                Gunakan 1 Tiket Hadiah untuk membuka kotak kejutan berisi hiasan profil.
              </p>
            </div>

            {error && (
              <div className="w-full p-2.5 rounded-xl bg-status-error/10 border border-status-error/20 text-status-error text-xs font-semibold">
                {error}
              </div>
            )}

            <Button
              size="lg"
              className="w-full bg-gradient-to-r from-amber-500 to-orange-500 hover:from-amber-600 hover:to-orange-600 text-white font-black shadow-lg shadow-amber-500/30 min-h-[46px] active:scale-[0.98]"
              isLoading={opening}
              onClick={handleOpen}
            >
              <Gift size={18} className="mr-2" /> Buka Hadiah
            </Button>
          </div>
        ) : (
          /* 2. Reveal Screen */
          <div className="mt-5 flex flex-col items-center gap-3 text-center">
            <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-amber-500/10 text-amber-700 dark:text-amber-300 border border-amber-500/25 font-black text-xs">
              <Sparkles size={13} className="text-amber-500" />
              <span>🎉 Kamu mendapatkan hadiah!</span>
            </div>

            {/* Live Avatar Preview Wearing Reward */}
            <div className="relative my-2 p-3 rounded-2xl bg-gradient-to-b from-amber-500/10 via-surface to-surface border border-border-subtle shadow-inner">
              <Avatar
                seed={profile?.avatar_seed || 'petualang'}
                frame={result.slot === 'frame' ? result.asset : profile?.avatar_frame || 'none'}
                effect={result.slot === 'effect' ? result.asset : profile?.equipped_explorer_effect || 'none'}
                size="xl"
              />
              <span className="absolute -bottom-1 -right-1 flex h-7 w-7 items-center justify-center rounded-full bg-accent-magic text-white text-xs shadow-md border-2 border-white">
                {result.slot === 'frame' ? '🖼️' : '✨'}
              </span>
            </div>

            {/* Cosmetic Name & Stars */}
            <div>
              <p className="text-base font-extrabold text-text-primary tracking-tight">
                {SLOT_LABEL[result.slot] ?? 'Hadiah'}
                {result.tier > 1 ? ` ★${result.tier}` : ''}
              </p>
              {COSMETIC_NAMES[result.asset] && (
                <p className="text-xs font-bold text-accent-magic mt-0.5">
                  Tema: {COSMETIC_NAMES[result.asset]}
                </p>
              )}
            </div>

            {/* Star Rating Visualization */}
            <div className="flex items-center justify-center gap-1 text-amber-500 text-base">
              {Array.from({ length: 3 }).map((_, idx) => (
                <span key={idx} className={idx < result.tier ? 'opacity-100' : 'opacity-25'}>
                  ★
                </span>
              ))}
            </div>

            {/* Ownership & Upgrade Feedback */}
            <p className="text-xs text-text-secondary leading-relaxed px-2">
              {result.is_new
                ? 'Gunakan hadiah ini untuk menghias profilmu.'
                : `Kamu sudah punya hadiah ini — tingkatnya naik! (Tingkat ★${result.tier})`}
            </p>

            {error && (
              <div className="w-full p-2.5 rounded-xl bg-status-error/10 border border-status-error/20 text-status-error text-xs font-semibold">
                {error}
              </div>
            )}

            <div className="w-full space-y-2 mt-2">
              {!equipped ? (
                <Button
                  size="lg"
                  className="w-full bg-accent-magic hover:brightness-110 text-white font-black shadow-md min-h-[44px]"
                  isLoading={equipping}
                  onClick={handleEquip}
                >
                  <Check size={16} className="mr-1.5" /> Pasang di Profil
                </Button>
              ) : (
                <div className="p-2.5 rounded-xl bg-emerald-500/10 border border-emerald-500/25 flex items-center justify-center gap-2 text-xs font-extrabold text-emerald-600 dark:text-emerald-400">
                  <Sparkles size={15} />
                  <span>Terpasang di profil!</span>
                </div>
              )}

              <button
                type="button"
                onClick={onClose}
                className="w-full py-2.5 text-xs font-extrabold text-text-secondary hover:text-text-primary transition-colors"
              >
                Selesai
              </button>
            </div>
          </div>
        )}
      </Card>
    </div>
  )
}
