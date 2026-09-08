import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Gift } from 'lucide-react'
import { Button } from '../../shared/components/atoms/Button'
import { rewardsApi } from '../../shared/lib/api'
import { CapsuleModal } from '../collection/CapsuleModal'

// Kartu Tiket Hadiah di beranda: alasan menetap setelah tugas selesai.
export function TicketCard() {
  const [tickets, setTickets] = useState(0)
  const [showCapsule, setShowCapsule] = useState(false)
  const [hint, setHint] = useState<string | null>(null)
  const [claiming, setClaiming] = useState(false)

  const reload = async () => {
    try {
      const st = await rewardsApi.status()
      setTickets(st.tickets ?? 0)
    } catch {
      // endpoint belum tersedia / offline: kartu disembunyikan
      setTickets(-1)
    }
  }

  useEffect(() => {
    void reload()
  }, [])

  if (tickets < 0) return null

  const handleClaim = async () => {
    setClaiming(true)
    setHint(null)
    try {
      const res = await rewardsApi.claimTicket()
      setTickets(res.tickets ?? 0)
      if (res.granted) setShowCapsule(true)
    } catch (e) {
      setHint(e instanceof Error ? e.message : 'Belum bisa mengambil tiket')
    } finally {
      setClaiming(false)
    }
  }

  return (
    <>
      <div className="relative overflow-hidden rounded-2xl border-2 border-accent-gold/40 bg-gradient-to-r from-amber-500/10 via-amber-400/5 to-transparent p-4 shadow-sm transition-all hover:border-accent-gold/60">
        {/* Ticket notch decorations on sides */}
        <div className="absolute -left-2.5 top-1/2 -mt-2.5 h-5 w-5 rounded-full bg-bg-app border-r-2 border-accent-gold/40" />
        <div className="absolute -right-2.5 top-1/2 -mt-2.5 h-5 w-5 rounded-full bg-bg-app border-l-2 border-accent-gold/40" />

        <div className="flex items-center gap-3.5 pl-2 pr-2">
          <div className="relative flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-gradient-to-br from-amber-400 to-amber-500 text-white shadow-md shadow-amber-400/20">
            <span className="text-2xl animate-bounce" style={{ animationDuration: '3s' }}>🎁</span>
            {tickets > 0 && (
              <span className="absolute -top-1 -right-1 flex h-5 w-5 items-center justify-center rounded-full bg-accent-magic text-[10px] font-black text-white shadow-sm ring-2 ring-white">
                {tickets}
              </span>
            )}
          </div>

          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <p className="text-sm font-black text-text-primary tracking-tight">
                Tiket Hadiah: {tickets}
              </p>
              {tickets > 0 && (
                <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-bold bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-300">
                  Siap Dibuka ✨
                </span>
              )}
            </div>
            <p className="truncate text-xs text-text-secondary mt-0.5">
              {tickets > 0
                ? 'Buka kotak kejutan berisi bingkai & efek profil.'
                : 'Selesaikan tugas hari ini untuk mendapatkan tiket.'}
            </p>
            {hint && <p className="mt-1 text-[11px] font-semibold text-accent-danger">{hint}</p>}
          </div>

          <div className="shrink-0">
            {tickets > 0 ? (
              <Button
                size="sm"
                className="bg-gradient-to-r from-amber-500 to-orange-500 hover:from-amber-600 hover:to-orange-600 text-white font-black shadow-md shadow-amber-500/25 px-4 min-h-[40px]"
                onClick={() => setShowCapsule(true)}
              >
                <Gift size={15} className="mr-1" /> Buka
              </Button>
            ) : (
              <Button
                variant="secondary"
                size="sm"
                isLoading={claiming}
                onClick={handleClaim}
                className="min-h-[40px] px-3 font-bold"
              >
                Ambil
              </Button>
            )}
          </div>
        </div>
      </div>

      <div className="flex items-center justify-between px-1 text-xs">
        <span className="text-text-secondary">Koleksi avatar & hiasan profil</span>
        <Link to="/koleksi" className="font-extrabold text-accent-magic hover:text-accent-magic/80 transition-colors inline-flex items-center gap-1">
          <span>Lihat Koleksi Saya</span>
          <span>→</span>
        </Link>
      </div>

      {showCapsule && (
        <CapsuleModal
          onClose={() => {
            setShowCapsule(false)
            void reload()
          }}
          onOpened={() => void reload()}
        />
      )}
    </>
  )
}
