import { useState } from 'react'
import confetti from 'canvas-confetti'
import { X, Sparkles, Check } from 'lucide-react'
import { Card } from '../../shared/components/atoms/Card'
import { Button } from '../../shared/components/atoms/Button'
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

export function CapsuleModal({ onClose, onOpened }: CapsuleModalProps) {
  const { refreshProfile } = useSession()
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
      void confetti({ particleCount: 90, spread: 75, origin: { y: 0.6 } })
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
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4" role="dialog" aria-modal="true" aria-label="Buka Hadiah">
      <Card className="w-full max-w-sm p-5">
        <div className="flex items-center justify-between">
          <h3 className="text-base font-extrabold text-text-primary">🎁 Buka Hadiah</h3>
          <button onClick={onClose} aria-label="Tutup" className="rounded-lg p-1.5 text-text-secondary hover:bg-bg-app">
            <X size={18} />
          </button>
        </div>

        {!result ? (
          <div className="mt-4 flex flex-col items-center gap-4 text-center">
            <span className="flex h-16 w-16 items-center justify-center rounded-2xl bg-accent-magic/10 text-3xl">🎁</span>
            <p className="text-xs text-text-secondary">Gunakan 1 Tiket Hadiah untuk membuka kotak kejutan berisi hiasan profil.</p>
            {error && <p className="text-xs font-semibold text-accent-danger">{error}</p>}
            <Button size="lg" className="w-full" isLoading={opening} onClick={handleOpen}>
              Buka Hadiah
            </Button>
          </div>
        ) : (
          <div className="mt-4 flex flex-col items-center gap-3 text-center">
            <p className="text-sm font-extrabold text-text-primary">🎉 Kamu mendapatkan hadiah!</p>
            <span className="flex h-16 w-16 items-center justify-center rounded-2xl bg-accent-gold/15 text-3xl">
              {result.slot === 'frame' ? '🖼️' : '✨'}
            </span>
            <p className="text-sm font-bold text-text-primary">
              {SLOT_LABEL[result.slot] ?? 'Hadiah'}
              {result.tier > 1 ? ` ★${result.tier}` : ''}
            </p>
            <p className="text-xs text-text-secondary">
              {result.is_new ? 'Gunakan hadiah ini untuk menghias profilmu.' : 'Kamu sudah punya hadiah ini — tingkatnya naik!'}
            </p>
            {error && <p className="text-xs font-semibold text-accent-danger">{error}</p>}
            {!equipped ? (
              <Button size="lg" className="w-full" isLoading={equipping} onClick={handleEquip}>
                <Check size={16} /> Pasang di Profil
              </Button>
            ) : (
              <p className="inline-flex items-center gap-1.5 text-xs font-bold text-emerald-600">
                <Sparkles size={14} /> Terpasang di profil!
              </p>
            )}
          </div>
        )}
      </Card>
    </div>
  )
}
