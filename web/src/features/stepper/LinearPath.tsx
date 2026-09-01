import React, { useState, useEffect, useCallback, useMemo } from 'react'
import { Link } from 'react-router-dom'
import { Flame, Coins, Trophy, Calendar, RefreshCw, CheckCircle2, Clock, Lock, ArrowRight, AlertTriangle, Play, FileText, Camera, HelpCircle, PenLine, Gamepad2, Sparkles, Banknote } from 'lucide-react'
import type { TaskView, RedemptionConfig } from '../../shared/types'
import { tasksApi, shopApi } from '../../shared/lib/api'
import { useSession } from '../../shared/hooks/useSession'
import { Card } from '../../shared/components/atoms/Card'
import { Button } from '../../shared/components/atoms/Button'
import { ProgressBar } from '../../shared/components/atoms/ProgressBar'
import { VideoQuizModal } from './VideoQuizModal'
import { DocUploadModal } from './DocUploadModal'
import { CameraCaptureModal } from './CameraCaptureModal'
import { LiveCameraCaptureModal } from './LiveCameraCaptureModal'
import { TextResponseModal } from './TextResponseModal'
import { MiniGameModal } from './MiniGameModal'

// Helper: greeting by time
function getGreeting() {
  const h = new Date().getHours()
  if (h < 11) return 'Selamat pagi'
  if (h < 15) return 'Selamat siang'
  if (h < 18) return 'Selamat sore'
  return 'Selamat malam'
}

function getTaskIcon(taskType: string) {
  switch (taskType) {
    case 'VIDEO':
    case 'VIDEO_QUIZ':
    case 'YOUTUBE_VIDEO': return <Play className="w-4 h-4" />
    case 'QUIZ': return <HelpCircle className="w-4 h-4" />
    case 'DOCUMENT_UPLOAD': return <FileText className="w-4 h-4" />
    case 'PHOTO_UPLOAD':
    case 'PHOTO_PROOF': return <Camera className="w-4 h-4" />
    case 'TEXT_RESPONSE': return <PenLine className="w-4 h-4" />
    case 'MINI_GAME': return <Gamepad2 className="w-4 h-4" />
    default: return <Sparkles className="w-4 h-4" />
  }
}

export const LinearPath: React.FC = () => {
  const { profile, refreshProfile } = useSession()
  const [tasks, setTasks] = useState<TaskView[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeModalTask, setActiveModalTask] = useState<TaskView | null>(null)
  const [shopConfig, setShopConfig] = useState<RedemptionConfig | null>(null)

  const loadTasks = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await tasksApi.getToday()
      setTasks(data.tasks || [])
    } catch (err: any) {
      setError(err.message || 'Gagal memuat alur tugas harian.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { loadTasks() }, [loadTasks])
  useEffect(() => {
    shopApi.getConfig().then((c) => setShopConfig(c)).catch(() => {})
  }, [])

  const handleTaskClick = (task: TaskView) => {
    if (task.is_locked || task.status === 'LOCKED') return
    setActiveModalTask(task)
  }
  const handleModalSuccess = () => {
    loadTasks()
    if (refreshProfile) refreshProfile()
  }

  const handleNextTask = (currentTask?: TaskView) => {
    loadTasks()
    if (refreshProfile) refreshProfile()
    if (currentTask) {
      const next = tasks.find((t) => t.step_order === currentTask.step_order + 1)
      if (next) {
        setActiveModalTask({ ...next, is_locked: false, status: next.status === 'LOCKED' ? 'UNLOCKED' : next.status })
        return
      }
    }
    setActiveModalTask(null)
  }

  const userCoins = profile?.coins || 0
  const userLevel = profile?.level || 1
  const userXP = profile?.xp || 0
  const userStreak = (profile as any)?.streak_days || 0
  const explorerName = profile?.explorer_name || ''

  const stats = useMemo(() => {
    const total = tasks.length
    const completed = tasks.filter(t => t.status === 'APPROVED').length
    const pending = tasks.filter(t => t.status === 'PENDING').length
    const rejected = tasks.filter(t => t.status === 'REJECTED').length
    const nextTask = tasks.find(t => t.status === 'UNLOCKED' || t.status === 'REJECTED')
    // fallback: first locked? but unlocked is actionable
    const isAllDone = total > 0 && completed === total
    const progressPercent = total === 0 ? 0 : Math.round((completed / total) * 100)
    return { total, completed, pending, rejected, nextTask, isAllDone, progressPercent }
  }, [tasks])

  // XP bar still as secondary info
  const currentLevelBaseXP = Math.pow(userLevel - 1, 2) * 100
  const nextLevelXP = Math.pow(userLevel, 2) * 100
  const xpInCurrentLevel = Math.max(0, userXP - currentLevelBaseXP)
  const xpRequiredForLevel = Math.max(1, nextLevelXP - currentLevelBaseXP)
  const xpProgressPercent = Math.min(100, Math.round((xpInCurrentLevel / xpRequiredForLevel) * 100))

  return (
    <div className="w-full flex flex-col gap-4">
      {/* 1. Greeting — compact, hierarchy: greeting > name > date */}
      <div className="px-1 pt-1">
        <p className="text-[11px] font-bold tracking-widest uppercase text-text-secondary">{getGreeting()} 👋</p>
        <h1 className="text-[22px] font-extrabold text-text-primary leading-tight mt-1 tracking-tight">
          {explorerName ? `Halo, ${explorerName}` : 'Yuk lanjutkan aktivitasmu'}
        </h1>
        <p className="text-xs text-text-secondary/80 mt-1.5 flex items-center gap-1.5">
          <Calendar className="w-3.5 h-3.5 shrink-0" />
          <span>{new Date().toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })}</span>
        </p>
      </div>

      {/* 2. Primary Action — unified hierarchy */}
      {loading && tasks.length === 0 ? (
        <Card className="p-5">
          <div className="animate-pulse space-y-3">
            <div className="h-3 w-24 bg-surface-elevated rounded-full" />
            <div className="h-5 w-48 bg-surface-elevated rounded-lg" />
            <div className="h-2 w-full bg-surface-elevated rounded-full" />
            <div className="h-11 w-full bg-surface-elevated rounded-xl" />
          </div>
        </Card>
      ) : error ? (
        <Card className="p-5 text-center">
          <div className="w-10 h-10 mx-auto rounded-full bg-status-error/10 text-status-error flex items-center justify-center">
            <AlertTriangle className="w-5 h-5" />
          </div>
          <h2 className="text-sm font-bold text-text-primary mt-3">Data belum bisa dimuat</h2>
          <p className="text-xs text-text-secondary mt-1 leading-relaxed">Periksa koneksi internetmu dan coba lagi.</p>
          <Button onClick={loadTasks} variant="secondary" size="md" className="w-full mt-4">
            <RefreshCw className="w-4 h-4 mr-2" /> Coba Lagi
          </Button>
          {error && <p className="text-[11px] text-text-secondary mt-2 break-words">{error}</p>}
        </Card>
      ) : tasks.length === 0 ? (
        <div className="rounded-2xl border border-dashed border-border-subtle bg-surface/60 p-6 text-center">
          <div className="w-11 h-11 mx-auto rounded-full bg-accent-magic/10 flex items-center justify-center text-lg">🌟</div>
          <h2 className="text-sm font-bold text-text-primary mt-3">Belum ada tugas hari ini</h2>
          <p className="text-xs text-text-secondary mt-1 leading-relaxed max-w-[30ch] mx-auto">
            Admin belum menambahkan tugas. Kamu akan mendapat notifikasi saat tugas baru tersedia.
          </p>
        </div>
      ) : stats.isAllDone ? (
        <Card className="p-5">
          <div className="flex items-start gap-3">
            <div className="w-10 h-10 rounded-full bg-status-success/15 text-status-success flex items-center justify-center shrink-0">
              <CheckCircle2 className="w-5 h-5" />
            </div>
            <div className="flex-1 min-w-0">
              <h2 className="text-[15px] font-extrabold text-text-primary leading-tight">Semua tugas hari ini selesai 🎉</h2>
              <p className="text-xs text-text-secondary mt-1 leading-relaxed">
                Kerja bagus! {stats.total} dari {stats.total} tugas selesai. Koin sudah masuk ke saldomu.
              </p>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-2 mt-4">
            <Link to="/shop" className="inline-flex items-center justify-center gap-1.5 py-3 rounded-xl bg-accent-magic text-white font-bold text-sm shadow-sm hover:brightness-110 transition-colors min-h-[44px]">
              <Coins className="w-4 h-4" /> Tukar Koin
            </Link>
            <Link to="/profile" className="inline-flex items-center justify-center gap-1.5 py-3 rounded-xl bg-surface-elevated border border-border-subtle text-text-primary font-bold text-sm hover:bg-surface transition-colors min-h-[44px]">
              <Trophy className="w-4 h-4" /> Perkembangan
            </Link>
          </div>
        </Card>
      ) : stats.nextTask ? (
        <Card className="p-5">
          <p className="text-[11px] font-bold tracking-widest uppercase text-text-secondary">Tugas Harian • Langkah {stats.nextTask.step_order} dari {stats.total}</p>
          <span className="sr-only">Tugas Harian</span>
          <h2 className="text-[15px] font-extrabold text-text-primary mt-1.5 leading-tight">
            {stats.completed === 0 ? 'Mulai tugas pertamamu' : 'Lanjutkan tugas berikutnya'}
          </h2>
          <p className="text-xs text-text-secondary mt-1.5 leading-relaxed">
            {stats.completed === 0
              ? 'Selesaikan tugas untuk mendapatkan koin dan membuka langkah berikutnya.'
              : `Masih ada ${stats.total - stats.completed} tugas lagi. Ayo selesaikan!`}
          </p>
          <div className="mt-3 flex items-center gap-2 flex-wrap">
            <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-accent-magic/10 text-accent-magic border border-accent-magic/15 font-bold text-xs">
              {getTaskIcon(stats.nextTask.task_type)} {stats.nextTask.title}
            </span>
            <span className="text-[11px] font-semibold text-text-secondary whitespace-nowrap">
              <span className="text-accent-gold font-bold">+{stats.nextTask.reward_coins} koin</span> • +{stats.nextTask.reward_xp} XP
            </span>
          </div>
          <Button onClick={() => handleTaskClick(stats.nextTask!)} size="lg" className="w-full mt-4">
            {stats.completed === 0 ? 'Mulai Tugas' : 'Lanjutkan Tugas'} <ArrowRight className="w-4 h-4 ml-2" />
          </Button>
          {stats.rejected > 0 && (
            <p className="text-xs text-status-error text-center mt-2.5 font-medium">Ada {stats.rejected} tugas perlu revisi — cek daftar di bawah.</p>
          )}
        </Card>
      ) : (
        <Card className="p-5 text-center">
          <div className="w-10 h-10 mx-auto rounded-full bg-accent-gold/15 text-accent-gold flex items-center justify-center">
            <Clock className="w-5 h-5" />
          </div>
          <h2 className="text-sm font-bold text-text-primary mt-3">Menunggu verifikasi</h2>
          <p className="text-xs text-text-secondary mt-1 leading-relaxed">Semua tugas sudah dikumpulkan. Koin akan masuk setelah admin memeriksa.</p>
        </Card>
      )}

      {/* 3. Progress — compact, single hierarchy */}
      {tasks.length > 0 && !loading && !error && (
        <div className="rounded-2xl bg-surface border border-border-subtle px-4 py-3.5">
          <div className="flex items-center justify-between gap-2">
            <h3 className="text-xs font-bold text-text-primary">Perkembangan hari ini</h3>
            <span className="text-[11px] font-bold px-2 py-0.5 rounded-full bg-accent-magic/10 text-accent-magic border border-accent-magic/15">{stats.completed}/{stats.total} selesai • {stats.progressPercent}%</span>
          </div>
          <div className="mt-2.5">
            <ProgressBar progress={stats.progressPercent} colorClass="bg-accent-magic" />
          </div>
          <div className="mt-2.5 flex items-center justify-between text-[11px] gap-2">
            <span className="inline-flex items-center gap-1.5 text-text-secondary">
              <Trophy className="w-3.5 h-3.5 text-accent-magic shrink-0" />
              <span className="font-bold text-text-primary">Lv. {userLevel}</span>
              <span>{xpInCurrentLevel}/{xpRequiredForLevel} XP</span>
            </span>
            <span className="inline-flex items-center gap-1 text-text-secondary shrink-0">
              <Flame className="w-3.5 h-3.5 text-accent-danger" /> <span className="font-bold">{userStreak} Hari</span>
            </span>
          </div>
          <div className="mt-2 h-1 bg-surface-elevated rounded-full overflow-hidden">
            <div className="h-full bg-accent-magic/35 rounded-full transition-all" style={{ width: `${xpProgressPercent}%` }} />
          </div>
        </div>
      )}

      {/* 4. Coins — supporting, grouped, no card heaviness */}
      <div className="rounded-2xl bg-surface border border-border-subtle px-4 py-3.5">
        {(() => {
          if (!shopConfig) {
            return (
              <div className="flex items-center justify-between gap-3 animate-pulse">
                <div className="h-10 w-36 bg-surface-elevated rounded-xl" />
                <div className="h-8 w-24 bg-surface-elevated rounded-xl" />
              </div>
            )
          }
          const conv = shopConfig.conversion_rate
          const maxCoins = shopConfig.max_payout_coins
          const estCash = userCoins * conv
          const maxCash = maxCoins * conv
          const targetCash = shopConfig.payout_target_rupiah
          const isPayoutDay = shopConfig.is_payout_day
          return (
            <>
              <div className="flex items-center justify-between gap-3">
                <div className="flex items-center gap-2.5 min-w-0">
                  <span className="w-9 h-9 rounded-xl bg-accent-gold/12 border border-accent-gold/15 flex items-center justify-center text-base shrink-0">🪙</span>
                  <div className="min-w-0">
                    <p className="text-[11px] font-bold tracking-widest uppercase text-text-secondary">Saldo Koin</p>
                    <p className="text-[17px] font-extrabold text-text-primary leading-none mt-1">
                      {userCoins.toLocaleString('id-ID')} <span className="text-xs font-semibold text-text-secondary">Koin</span>
                    </p>
                    <p className="text-[11px] text-text-secondary mt-1">
                      ≈ Rp {estCash.toLocaleString('id-ID')} • Maks Rp {maxCash.toLocaleString('id-ID')}
                    </p>
                  </div>
                </div>
                <Link to="/shop" className="shrink-0 inline-flex items-center gap-1 px-3 py-2.5 rounded-xl bg-accent-magic text-white font-bold text-xs shadow-sm hover:brightness-110 transition-colors min-h-[38px]">
                  Tukar <ArrowRight className="w-3.5 h-3.5" />
                </Link>
              </div>
              {isPayoutDay && userCoins > 0 && (
                <div className="mt-3 p-2.5 rounded-xl bg-status-success/10 border border-status-success/15 flex items-start gap-2">
                  <Banknote className="w-4 h-4 text-status-success shrink-0 mt-0.5" />
                  <p className="text-xs leading-relaxed text-text-secondary">
                    <span className="font-bold text-status-success">Saatnya tukarkan!</span> {userCoins.toLocaleString('id-ID')} Koin = Rp {estCash.toLocaleString('id-ID')} (target Rp {targetCash.toLocaleString('id-ID')}). Periode tgl {shopConfig.redemption_start_day}–{shopConfig.redemption_end_day}.
                  </p>
                </div>
              )}
            </>
          )
        })()}
      </div>

      {/* 5. Task list — compact, clean rows */}
      {tasks.length > 0 && !error && (
        <div className="flex flex-col gap-2">
          <div className="flex items-center justify-between px-1">
            <h3 className="text-xs font-bold tracking-widest uppercase text-text-secondary">Daftar tugas • {stats.completed}/{stats.total}</h3>
            <button onClick={loadTasks} aria-label="Muat ulang tugas" disabled={loading} className="w-7 h-7 rounded-full bg-surface border border-border-subtle text-text-secondary hover:text-text-primary disabled:opacity-50 flex items-center justify-center">
              <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
            </button>
          </div>
          <div className="flex flex-col gap-1.5">
            {tasks.map((task) => {
              const isApproved = task.status === 'APPROVED'
              const isPending = task.status === 'PENDING'
              const isUnlocked = task.status === 'UNLOCKED'
              const isLocked = task.status === 'LOCKED' || task.is_locked
              const isRejected = task.status === 'REJECTED'
              return (
                <button
                  key={task.id}
                  onClick={() => handleTaskClick(task)}
                  disabled={isLocked}
                  className={`w-full text-left px-3 py-2.5 rounded-xl border flex items-center gap-2.5 transition-colors min-h-[56px] ${isApproved ? 'bg-status-success/[0.04] border-status-success/15' : isPending ? 'bg-accent-gold/[0.04] border-accent-gold/15' : isUnlocked ? 'bg-surface border-border-subtle hover:border-accent-magic/25 hover:bg-surface-elevated' : isRejected ? 'bg-status-error/[0.04] border-status-error/15' : 'bg-surface border-border-subtle opacity-60'}`}
                >
                  <span className={`w-8 h-8 rounded-lg flex items-center justify-center shrink-0 ${isApproved ? 'bg-status-success text-white' : isPending ? 'bg-accent-gold text-white' : isUnlocked ? 'bg-accent-magic text-white' : isRejected ? 'bg-status-error text-white' : 'bg-surface-elevated text-text-secondary border border-border-subtle'}`}>
                    {isApproved ? <CheckCircle2 className="w-4 h-4" /> : isPending ? <Clock className="w-4 h-4" /> : isLocked ? <Lock className="w-3.5 h-3.5" /> : isRejected ? <AlertTriangle className="w-3.5 h-3.5" /> : getTaskIcon(task.task_type)}
                  </span>
                  <span className="flex-1 min-w-0">
                    <span className={`text-[13px] font-bold leading-tight block truncate ${isLocked ? 'text-text-secondary' : 'text-text-primary'}`}>{task.title}</span>
                    <span className="text-[11px] text-text-secondary line-clamp-1 leading-snug">{task.description || (isLocked ? 'Selesaikan tugas sebelumnya untuk membuka' : `Langkah ${task.step_order} • +${task.reward_coins} koin`)}</span>
                  </span>
                  <span className={`text-[10px] font-bold px-2 py-1 rounded-full shrink-0 whitespace-nowrap ${isApproved ? 'bg-status-success text-white' : isPending ? 'bg-accent-gold/12 text-accent-gold border border-accent-gold/15' : isUnlocked ? 'bg-accent-magic text-white' : isRejected ? 'bg-status-error/10 text-status-error border border-status-error/15' : 'bg-surface-elevated text-text-secondary border border-border-subtle'}`}>
                    {isApproved ? 'Selesai' : isPending ? 'Menunggu' : isUnlocked ? 'Kerjakan' : isRejected ? 'Revisi' : 'Terkunci'}
                  </span>
                </button>
              )
            })}
          </div>
        </div>
      )}

      {/* Modals — preserve existing, camera-only for new task */}
      {activeModalTask && (() => {
        let cfg: any = activeModalTask.config || {}
        if (typeof cfg === 'string') {
          try {
            cfg = JSON.parse(cfg)
          } catch {
            cfg = {}
          }
        }
        const isLiveCamera = Boolean(cfg.camera_only) || activeModalTask.title === 'Foto Langsung Kesiapan Profil CV (Rapi & Profesional)'
        const isDoc = activeModalTask.task_type === 'DOCUMENT_UPLOAD' || Boolean(cfg.attachment_url)
        const isPhoto = activeModalTask.task_type === 'PHOTO_UPLOAD' || activeModalTask.task_type === 'PHOTO_PROOF'
        const isGame = activeModalTask.task_type === 'MINI_GAME' || Boolean(cfg.game)
        const isText = activeModalTask.task_type === 'TEXT_RESPONSE' || (!isDoc && !isPhoto && Boolean(cfg.minimum_characters || cfg.prompt))
        if (isLiveCamera && isPhoto) return <LiveCameraCaptureModal task={activeModalTask} onClose={() => setActiveModalTask(null)} onSuccess={handleModalSuccess} onNextTask={() => handleNextTask(activeModalTask)} />
        if (isDoc) return <DocUploadModal task={activeModalTask} onClose={() => setActiveModalTask(null)} onSuccess={handleModalSuccess} onNextTask={() => handleNextTask(activeModalTask)} />
        if (isPhoto) return <CameraCaptureModal task={activeModalTask} onClose={() => setActiveModalTask(null)} onSuccess={handleModalSuccess} onNextTask={() => handleNextTask(activeModalTask)} />
        if (isText) return <TextResponseModal task={activeModalTask} onClose={() => setActiveModalTask(null)} onSuccess={handleModalSuccess} onNextTask={() => handleNextTask(activeModalTask)} />
        if (isGame) return <MiniGameModal task={activeModalTask} onClose={() => setActiveModalTask(null)} onSuccess={handleModalSuccess} onNextTask={() => handleNextTask(activeModalTask)} />
        return <VideoQuizModal task={activeModalTask} onClose={() => setActiveModalTask(null)} onSuccess={handleModalSuccess} onNextTask={() => handleNextTask(activeModalTask)} />
      })()}
    </div>
  )
}
