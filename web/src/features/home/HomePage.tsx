import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { Card } from '../../shared/components/atoms/Card'
import { Button } from '../../shared/components/atoms/Button'
import { apiClient, chestsApi, crewsApi } from '../../shared/lib/api'
import type { HomeResponse, Crew } from '../../shared/types'
import { ConnectedReactionBar } from '../../shared/components/molecules/ConnectedReactionBar'
import { YourTurnBadge } from '../../shared/components/molecules/YourTurnBadge'
import { useSession } from '../../shared/hooks/useSession'
import { isMyRelayTurn } from '../../shared/utils/questTurn'
import { OnboardingModal } from './OnboardingModal'
import { DailyActivitySection } from './DailyActivitySection'

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
  const { session } = useSession()
  const [home, setHome] = useState<HomeResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [openingChestId, setOpeningChestId] = useState<number | null>(null)
  const [crew, setCrew] = useState<Crew | null>(null)

  const loadCrew = useCallback(async () => {
    try {
      const data = await crewsApi.get()
      setCrew(data)
    } catch {
      // best-effort
    }
  }, [])

  // Production's serverless /api/home is a heavy aggregate (~8s warm, slower on
  // cold start). Bound the wait so the UI never appears stuck "Memuat dunia".
  const HOME_TIMEOUT_MS = 15_000

  const loadHome = useCallback(async () => {
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
  }, [])

  useEffect(() => {
    loadHome()
    loadCrew()
  }, [loadHome, loadCrew])



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

  const renderContent = () => {
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
        data-theme={crew?.theme || 'default'}
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
                Mari belajar keterampilan baru bersama keluarga hari ini!
              </p>
            </div>
            <div
              className="inline-flex items-center gap-2 self-center sm:self-start rounded-full border border-accent-reward/30 bg-accent-reward/10 px-3 py-1.5 shadow-sm"
              data-testid="home-coin-balance"
              title="Koin adalah hadiah virtual dari aktivitas dan misi. Koin bisa digunakan untuk kosmetik profil."
            >
              <span className="text-sm" aria-hidden>🪙</span>
              <span className="text-sm font-bold text-accent-reward tabular-nums">
                {home.player?.coins ?? 0}
              </span>
              <span className="text-[10px] font-semibold uppercase tracking-wider text-text-secondary">
                Koin
              </span>
            </div>
          </div>
        </motion.section>

        {/* Crew Banner */}
        {crew?.banner_url && (
          <motion.section variants={itemVariants} className="mt-2">
            <div
              className="w-full h-40 md:h-56 rounded-xl overflow-hidden border border-border-subtle shadow-lg"
              data-testid="crew-banner"
              style={{
                backgroundImage: `url(${crew.banner_url})`,
                backgroundSize: 'cover',
                backgroundPosition: 'center',
              }}
            />
          </motion.section>
        )}

        {/* 2. Giliran Harian / Daily Activity */}
        <motion.section variants={itemVariants} className="flex flex-col mt-2">
          <DailyActivitySection />
          
          {/* Primary CTA for New / First-time Users */}
          {activeQuests.length === 0 && completedToday.length === 0 && (
             <div className="mt-4 p-5 bg-surface-elevated border-2 border-accent-primary rounded-xl shadow-md text-center">
               <h3 className="font-heading text-lg text-text-primary mb-2">Mulai Aktivitas Hari Ini</h3>
               <p className="text-sm text-text-secondary mb-4">
                 Belajar singkat 30–60 detik tentang hal yang berguna sehari-hari.
               </p>
               <Button onClick={() => navigate('/quests')} className="w-full">
                 Mulai Sekarang
               </Button>
             </div>
          )}
          {home.daily_turn?.available === false && activeQuests.length === 0 && completedToday.length === 0 && (
             <div className="mt-4 p-5 bg-surface-elevated border-2 border-accent-reward rounded-xl shadow-md text-center">
               <h3 className="font-heading text-lg text-text-primary mb-2">Bagus!</h3>
               <p className="text-sm text-text-secondary mb-4">
                 Sekarang lanjutkan satu Misi bersama keluarga.
               </p>
               <Button onClick={() => navigate('/quests')} className="w-full">
                 Pilih Misi
               </Button>
             </div>
          )}

          <div className="flex justify-between items-center px-2 mt-2">
            <p className="text-xs text-text-secondary">Runtutan: {home.daily_turn?.streak_days || 0} Hari 🔥</p>
            <p className="text-xs text-text-secondary" data-testid="home-crew-streak">
              Runtutan keluarga: {home.daily_turn?.crew_streak ?? 0} hari bersama 🤝
            </p>
          </div>
        </motion.section>

        {/* 3. Petualangan Aktif Keluarga */}
        <motion.section variants={itemVariants} className="flex flex-col gap-3 mt-2">
          <div className="flex items-center justify-between">
            <h2 className="font-heading text-xl text-text-primary flex items-center gap-2">
              <span className="text-accent-reward">📜</span> Misi Aktif
            </h2>
            <Button variant="ghost" size="sm" onClick={() => navigate('/quests')}>Lihat Semua</Button>
          </div>
          
          {activeQuests.length > 0 ? (
            <div className="flex flex-col gap-3">
              {activeQuests.slice(0, 2).map(quest => (
                <Card 
                  key={quest.id} 
                  className="flex justify-between items-center p-5 bg-surface-elevated border border-border-subtle cursor-pointer hover:border-accent-reward/50 transition-colors shadow-sm hover:shadow-md"
                  onClick={() => navigate(`/quests/${quest.id}`)}
                >
                  <div>
                    <h3 className="font-medium text-text-primary text-lg mb-1">{quest.title}</h3>
                    <div className="flex items-center gap-2">
                      <p className="text-xs text-text-secondary">
                        {quest.completed_count}/{quest.challenge_count} Latihan Selesai
                      </p>
                      {isMyRelayTurn(quest, session?.uid) && <YourTurnBadge />}
                    </div>
                  </div>
                  <div className="text-accent-reward text-xl">➡️</div>
                </Card>
              ))}
              {activeQuests.length > 2 && (
                <p className="text-xs text-text-secondary text-center">
                  + {activeQuests.length - 2} Misi lainnya...
                </p>
              )}
            </div>
          ) : (
            <Card className="p-6 text-center border-dashed border-border-subtle bg-surface/50">
              <p className="text-text-secondary text-sm">Tidak ada aktivitas yang sedang berjalan. Bicarakan materi esok hari dengan keluargamu!</p>
            </Card>
          )}
        </motion.section>

        {/* 4. Peti Harta */}
        {availableChests.length > 0 && (
          <motion.section variants={itemVariants} className="flex flex-col mt-2 gap-3">
            <h2 className="font-heading text-xl text-text-primary flex items-center gap-2">
              <span>🎁</span> Peti Hadiah Tersedia
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

        {/* 5. Aktivitas Keluarga Hari Ini */}
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
                      <p className="text-xs text-text-secondary">Telah diselesaikan oleh keluarga</p>
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

  return (
    <>
      <OnboardingModal />
      {renderContent()}
    </>
  )
}

