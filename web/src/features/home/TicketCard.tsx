import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Gift } from 'lucide-react'
import { Card } from '../../shared/components/atoms/Card'
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
      <Card className="flex items-center gap-3 p-4">
        <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-accent-magic/10 text-2xl">🎁</span>
        <div className="min-w-0 flex-1">
          <p className="text-sm font-extrabold text-text-primary">Tiket Hadiah: {tickets}</p>
          <p className="truncate text-[11px] text-text-secondary">
            {tickets > 0 ? 'Buka kotak kejutan berisi hiasan profil.' : 'Selesaikan tugas hari ini untuk mendapatkan tiket.'}
          </p>
          {hint && <p className="mt-1 text-[11px] font-semibold text-accent-danger">{hint}</p>}
        </div>
        {tickets > 0 ? (
          <Button size="sm" onClick={() => setShowCapsule(true)}>
            <Gift size={14} /> Buka
          </Button>
        ) : (
          <Button variant="secondary" size="sm" isLoading={claiming} onClick={handleClaim}>
            Ambil
          </Button>
        )}
      </Card>
      <div className="text-right">
        <Link to="/koleksi" className="text-[11px] font-bold text-accent-magic hover:underline">
          Lihat Koleksi Saya →
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
