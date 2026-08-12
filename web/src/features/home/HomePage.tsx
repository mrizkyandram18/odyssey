import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { Card } from '../../shared/components/atoms/Card'
import { Button } from '../../shared/components/atoms/Button'
import { apiClient, crewsApi } from '../../shared/lib/api'
import type { HomeResponse, Crew } from '../../shared/types'
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
  const [home, setHome] = useState<HomeResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
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

    return (
      <motion.div
        className="flex flex-col gap-6 max-w-md mx-auto pb-8"
        data-theme={crew?.theme || 'default'}
        variants={containerVariants}
        initial="hidden"
        animate="show"
      >
        {/* Header - No Coins */}
        <motion.section variants={itemVariants} className="mt-4 flex items-center justify-between">
          <div>
            <h1 className="font-heading text-2xl font-bold text-text-primary mb-1">
              Odyssey
            </h1>
          </div>
          <div className="flex items-center gap-3">
            <span className="font-semibold text-text-primary">Hi, {home.player?.explorer_name || 'Explorer'}!</span>
            <div className="w-10 h-10 bg-accent-reward/20 rounded-full flex items-center justify-center text-accent-reward">
              👤
            </div>
          </div>
        </motion.section>

        {/* 1. Primary: Aktivitas Hari Ini */}
        <motion.section variants={itemVariants} className="flex flex-col mt-2">
          <DailyActivitySection />
        </motion.section>

        {/* 2. Secondary: Misi Berikutnya */}
        <motion.section variants={itemVariants} className="flex flex-col gap-3 mt-4">
          <div className="flex items-center justify-between">
            <h2 className="font-heading text-lg font-bold text-text-primary uppercase tracking-wide">
              Misi Berikutnya
            </h2>
            <Button variant="ghost" size="sm" onClick={() => navigate('/quests')}>Lihat Semua</Button>
          </div>
          
          {activeQuests.length > 0 ? (
            <div className="flex flex-col gap-3">
              {activeQuests.slice(0, 1).map(quest => (
                <Card 
                  key={quest.id} 
                  hoverable
                  className="p-0 overflow-hidden flex flex-col"
                  onClick={() => navigate(`/quests/${quest.id}`)}
                >
                  <div className="p-5 flex justify-between items-center bg-surface-elevated">
                    <div>
                      <h3 className="font-bold text-text-primary text-xl mb-1">{quest.title}</h3>
                      <p className="text-sm text-text-secondary">
                        Misi {quest.id}: Eksplorasi Pengetahuan
                      </p>
                    </div>
                  </div>
                  <div className="px-5 pb-5 pt-2">
                     <Button variant="primary" className="w-auto">Mulai Misi</Button>
                  </div>
                </Card>
              ))}
            </div>
          ) : (
            <Card className="p-6 text-center border-dashed bg-surface-elevated">
              <p className="text-text-secondary text-sm">Tidak ada Misi aktif. Pilih materi baru untuk dipelajari!</p>
              <Button onClick={() => navigate('/quests')} className="mt-4">Pilih Misi</Button>
            </Card>
          )}
        </motion.section>

        {/* 3. Tertiary: Perjalanan Belajar */}
        <motion.section variants={itemVariants} className="flex flex-col gap-3 mt-4">
           <h2 className="font-heading text-lg font-bold text-text-primary uppercase tracking-wide">
              Perjalanan Belajarmu
            </h2>
            <Card className="p-5 bg-surface-elevated relative overflow-hidden">
               <div className="absolute right-0 top-0 opacity-10 pointer-events-none">
                 <svg width="200" height="200" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2L15.09 8.26L22 9.27L17 14.14L18.18 21.02L12 17.77L5.82 21.02L7 14.14L2 9.27L8.91 8.26L12 2Z"/></svg>
               </div>
               <div className="flex justify-between items-end mb-4 relative z-10">
                 <div>
                   <div className="text-accent-reward font-bold text-3xl">{home.player?.level ?? 1}</div>
                   <div className="text-sm font-semibold text-text-primary mt-1">Level {home.player?.level ?? 1} | Petualang Pengetahuan</div>
                 </div>
                 <div className="text-right">
                   <div className="text-xs font-semibold text-text-secondary mb-1">Target Bulan Ini: Selesaikan 5 Misi</div>
                   <div className="text-sm font-bold">{completedToday.length}/5</div>
                 </div>
               </div>
               <div className="w-full bg-border-subtle rounded-full h-2.5 relative z-10">
                  <div className="bg-accent-reward h-2.5 rounded-full" style={{ width: '40%' }}></div>
               </div>
            </Card>
        </motion.section>

        {/* 4. Keluarga */}
        <motion.section variants={itemVariants} className="flex flex-col gap-3 mt-4">
            <h2 className="font-heading text-lg font-bold text-text-primary uppercase tracking-wide">
              Keluarga
            </h2>
            <Card className="p-5 bg-surface-elevated flex flex-col gap-4">
              <div>
                <h3 className="font-bold text-text-primary">Kelompok Jelajah</h3>
                <p className="text-xs text-text-secondary">Runtutan: {home.daily_turn?.crew_streak ?? 0} hari belajar bersama</p>
              </div>
              <div className="flex gap-3 overflow-x-auto pb-2">
                 {[1,2,3,4].map((i) => (
                   <div key={i} className="flex flex-col items-center gap-1 min-w-[60px]">
                     <div className="w-12 h-12 rounded-full bg-border-subtle border-2 border-surface flex items-center justify-center text-xl">
                       {i === 1 ? '👩' : i === 2 ? '👨' : '👦'}
                     </div>
                     <span className="text-xs font-medium text-text-secondary text-center">Anggota {i}</span>
                   </div>
                 ))}
              </div>
              {completedToday.length > 0 && (
                <div className="mt-2 pt-4 border-t border-border-subtle">
                   <p className="text-sm font-medium">Hari ini: 1 Aktivitas diselesaikan</p>
                </div>
              )}
            </Card>
        </motion.section>
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

