import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { Card } from '../../shared/components/atoms/Card'
import { Button } from '../../shared/components/atoms/Button'
import { apiClient, chestsApi } from '../../shared/lib/api'
import type { HomeResponse } from '../../shared/types'
import { ConnectedReactionBar } from '../../shared/components/molecules/ConnectedReactionBar'
import { YourTurnBadge } from '../../shared/components/molecules/YourTurnBadge'
import { useSession } from '../../shared/hooks/useSession'
import { WorldMap } from '../../shared/components/organisms/WorldMap'
import { isMyRelayTurn } from '../../shared/utils/questTurn'

const containerVariants = {
  hidden: { opacity: 0 },
  show: {
    opacity: 1,
    transition: { staggerChildren: 0.05 }
  }
}

const itemVariants = {
  hidden: { opacity: 0, y: 20 },
  show: { opacity: 1, y: 0 }
}

export function HomePage() {
  const navigate = useNavigate()
  const { session, refreshProfile } = useSession()
  const [home, setHome] = useState<HomeResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [takingTurn, setTakingTurn] = useState(false)
  const [openingChestId, setOpeningChestId] = useState<number | null>(null)

  // Production's serverless /api/home is a heavy aggregate (~8s warm, slower on
  // cold start). Bound the wait so the UI never appears stuck "Memuat dunia".
  const HOME_TIMEOUT_MS = 15_000

  const loadHome = async () => {
    setLoading(true)
    setError(null)
    const controller = new AbortController()
    const timer = setTimeout(() => controller.abort(), HOME_TIMEOUT_MS)
    try {
      const data = await apiClient.request<HomeResponse>('/api/home', {
        method: 'GET',
        signal: controller.signal,
      })
      setHome(data)
    } catch (e: any) {
      if (e?.name === 'AbortError') {
        setError('Beranda membutuhkan waktu terlalu lama. Cek koneksi dan coba lagi.')
      } else {
        setError(e?.message || 'Gagal memuat data beranda.')
      }
      console.error('Gagal memuat data beranda', e)
    } finally {
      clearTimeout(timer)
      setLoading(false)
    }
  }

  useEffect(() => {
    loadHome()
  }, [])

  const takeTurn = async () => {
    setTakingTurn(true)
    try {
      await apiClient.post('/api/daily_turns/consume', { quest_slug: 'daily-turn' })
      // Refresh home + session profile so coin balance is not stale after +1 daily earn.
      await Promise.all([loadHome(), refreshProfile()])
    } catch (e) {
      console.error('Gagal mengambil giliran', e)
    } finally {
      setTakingTurn(false)
    }
  }

  const openChest = async (chestId: number) => {
    setOpeningChestId(chestId)
    try {
      await chestsApi.open(chestId)
      await loadHome()
    } catch (e) {
      console.error('Gagal membuka peti', e)
    } finally {
      setOpeningChestId(null)
    }
  }

  if (loading && !home) {
    return (
      <div className="flex h-64 w-full items-center justify-center">
        <div className="flex flex-col items-center gap-4 animate-pulse">
          <div className="text-4xl">🧭</div>
          <p className="text-sm text-text-secondary">Memuat dunia...</p>
        </div>
      </div>
    )
  }

  if (!home) {
    return (
      <div className="flex h-64 w-full items-center justify-center max-w-md mx-auto">
        <div className="text-center">
          <p className="text-accent-danger mb-4">{error}</p>
          <Button size="sm" variant="secondary" onClick={loadHome}>
            Coba Lagi
          </Button>
        </div>
      </div>
    )
  }

  const activeQuests = home.active_quests || []
  const completedToday = home.completed_quests_today || []
  const availableChests = home.available_chests || []

  return (
    <motion.div 
      className="flex flex-col gap-6 max-w-md mx-auto pb-8"
      variants={containerVariants}
      initial="hidden"
      animate="show"
    >
      {/* 1. Sapaan - Odyssey Vibe */}
      <motion.section variants={itemVariants} className="mt-2 text-center md:text-left">
        <div className="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <h1 className="font-heading text-3xl text-text-primary mb-1">
              Halo, {home.player?.explorer_name || 'Explorer'}!
            </h1>
            <p className="text-text-secondary text-sm italic">
              Setiap langkah yang kita ambil bersama akan menuliskan bab baru dalam legenda kita.
            </p>
          </div>
          <div
            className="inline-flex items-center gap-2 self-center sm:self-start rounded-full border border-accent-reward/30 bg-accent-reward/10 px-3 py-1.5 shadow-sm"
            data-testid="home-coin-balance"
            title="Coin balance (fictional — no real money)"
          >
            <span className="text-sm" aria-hidden>🪙</span>
            <span className="text-sm font-bold text-accent-reward tabular-nums">
              {home.player?.coins ?? 0}
            </span>
            <span className="text-[10px] font-semibold uppercase tracking-wider text-text-secondary">
              Coins
            </span>
          </div>
        </div>
      </motion.section>

      {/* 2. Progres & Peta Dunia */}
      <motion.section variants={itemVariants} className="flex flex-col gap-3">
        <div className="flex items-center justify-between">
          <h2 className="font-heading text-xl text-text-primary flex items-center gap-2">
            <span>🌍</span> Peta Dunia & Ranah
          </h2>
          <Button variant="ghost" size="sm" onClick={() => navigate('/quests')}>
            Jelajah Misi
          </Button>
        </div>
        <WorldMap
          realms={home.realm_progress}
          onRealmSelect={(realmSlug) => navigate(`/quests?realm=${realmSlug}`)}
        />
      </motion.section>

      {/* 3. Giliran Harian */}
      <motion.section variants={itemVariants} className="flex flex-col mt-2">
        <h2 className="font-heading text-xl text-text-primary mb-3 flex items-center gap-2">
          <span>✍️</span> Tugas Harian
        </h2>
        <Card className="flex flex-row items-center justify-between p-4 bg-surface border-l-4 border-l-accent-nature">
          <div className="flex flex-col gap-1">
            <h3 className="font-medium text-text-primary text-sm">Giliran Hari Ini</h3>
            <p className="text-xs text-text-secondary">Runtutan: {home.daily_turn.streak_days} Hari 🔥</p>
            <p className="text-xs text-text-secondary" data-testid="home-crew-streak">
              Runtutan kru: {home.daily_turn.crew_streak ?? 0} hari bersama 🤝
            </p>
          </div>
          <Button 
            variant="secondary" 
            size="sm"
            onClick={takeTurn}
            isLoading={takingTurn}
            disabled={!home.daily_turn.available}
          >
            {home.daily_turn.available ? 'Kerjakan' : 'Selesai'}
          </Button>
        </Card>
      </motion.section>

      {/* 4. Petualangan Aktif Keluarga */}
      <motion.section variants={itemVariants} className="flex flex-col gap-3 mt-2">
        <div className="flex items-center justify-between">
          <h2 className="font-heading text-xl text-text-primary flex items-center gap-2">
            <span className="text-accent-reward">📜</span> Buku Misi Aktif
          </h2>
          <Button variant="ghost" size="sm" onClick={() => navigate('/quests')}>Lihat Semua</Button>
        </div>
        
        {activeQuests.length > 0 ? (
          <div className="flex flex-col gap-3">
            {activeQuests.map(quest => (
              <Card 
                key={quest.id} 
                className="flex justify-between items-center p-5 bg-surface-elevated border border-border-subtle cursor-pointer hover:border-accent-reward/50 transition-colors shadow-sm hover:shadow-md"
                onClick={() => navigate(`/quests/${quest.id}`)}
              >
                <div>
                  <h3 className="font-medium text-text-primary text-lg mb-1">{quest.title}</h3>
                  <div className="flex items-center gap-2">
                    <p className="text-xs text-text-secondary">
                      {quest.completed_count}/{quest.challenge_count} Tantangan Selesai
                    </p>
                    {isMyRelayTurn(quest, session?.uid) && <YourTurnBadge />}
                  </div>
                </div>
                <div className="text-accent-reward text-xl">➡️</div>
              </Card>
            ))}
          </div>
        ) : (
          <Card className="p-6 text-center border-dashed border-border-subtle bg-surface/50">
            <p className="text-text-secondary text-sm">Tidak ada cerita yang aktif saat ini. Beristirahatlah sejenak dan bicarakan petualangan esok hari!</p>
          </Card>
        )}
      </motion.section>

      {/* 5. Peti Harta */}
      {availableChests.length > 0 && (
        <motion.section variants={itemVariants} className="flex flex-col mt-2 gap-3">
          <h2 className="font-heading text-xl text-text-primary flex items-center gap-2">
            <span>🎁</span> Harta Karun Tersedia
          </h2>
          <div className="grid grid-cols-2 gap-3">
            {availableChests.map(chest => (
              <Card key={chest.id} className="flex flex-col items-center justify-center p-4 bg-surface gap-2 text-center border-accent-reward/30">
                <span className="text-4xl drop-shadow-md">📦</span>
                <span className="text-xs text-text-secondary font-medium">Dari {chest.source.replace('_', ' ')}</span>
                <Button 
                  size="sm" 
                  className="w-full mt-2"
                  onClick={() => openChest(chest.id)}
                  isLoading={openingChestId === chest.id}
                >
                  Buka
                </Button>
              </Card>
            ))}
          </div>
        </motion.section>
      )}

      {/* 6. Aktivitas Keluarga Hari Ini */}
      {completedToday.length > 0 && (
        <motion.section variants={itemVariants} className="flex flex-col mt-2 gap-3">
          <h2 className="font-heading text-xl text-text-primary flex items-center gap-2">
            <span>🏆</span> Jejak Hari Ini
          </h2>
          <div className="flex flex-col gap-2">
            {completedToday.map(quest => (
              <Card key={quest.id} className="flex flex-col gap-3 p-3 bg-surface/50 border-border-subtle">
                <div className="flex items-center gap-3">
                  <div className="bg-accent-reward/20 text-accent-reward p-2 rounded-full">
                    ✓
                  </div>
                  <div className="flex-1">
                    <h4 className="text-sm font-medium text-text-primary">{quest.title}</h4>
                    <p className="text-xs text-text-secondary">Telah diselesaikan oleh kru</p>
                  </div>
                </div>
                <div className="flex justify-end border-t border-border-subtle/50 pt-2">
                  <ConnectedReactionBar targetType="QUEST" targetId={quest.id} />
                </div>
              </Card>
            ))}
          </div>
        </motion.section>
      )}

    </motion.div>
  )
}
