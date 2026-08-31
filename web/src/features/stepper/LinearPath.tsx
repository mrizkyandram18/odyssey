import React, { useState, useEffect, useCallback, useMemo } from 'react'
import { Link } from 'react-router-dom'
import { Flame, Coins, Trophy, Calendar, RefreshCw, CheckCircle2, Clock, Lock, ArrowRight, AlertTriangle, Play, FileText, Camera, HelpCircle, PenLine, Gamepad2, Sparkles } from 'lucide-react'
import type { TaskView } from '../../shared/types'
import { tasksApi } from '../../shared/lib/api'
import { useSession } from '../../shared/hooks/useSession'
import { Card } from '../../shared/components/atoms/Card'
import { Button } from '../../shared/components/atoms/Button'
import { ProgressBar } from '../../shared/components/atoms/ProgressBar'
import { VideoQuizModal } from './VideoQuizModal'
import { DocUploadModal } from './DocUploadModal'
import { CameraCaptureModal } from './CameraCaptureModal'
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

  const handleTaskClick = (task: TaskView) => {
    if (task.is_locked || task.status === 'LOCKED') return
    setActiveModalTask(task)
  }
  const handleModalSuccess = () => {
    loadTasks()
    if (refreshProfile) refreshProfile()
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
      {/* 1. Greeting — compact, human */}
      <div className="px-1 pt-1">
        <p className="text-sm text-text-secondary">{getGreeting()} 👋</p>
        <h1 className="text-xl font-bold text-text-primary leading-tight mt-0.5">
          {explorerName ? `Halo, ${explorerName}` : 'Yuk lanjutkan aktivitasmu'}
        </h1>
        <p className="text-xs text-text-secondary mt-1 flex items-center gap-1.5">
          <Calendar className="w-3.5 h-3.5" />
          <span>{new Date().toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })}</span>
        </p>
      </div>

      {/* 2. Primary Action — state aware */}
      {loading && tasks.length === 0 ? (
        <Card className="p-5">
          <div className="animate-pulse space-y-3">
            <div className="h-4 w-32 bg-surface border border-border-subtle rounded-full" />
            <div className="h-6 w-48 bg-surface border border-border-subtle rounded-lg" />
            <div className="h-3 w-full bg-surface border border-border-subtle rounded-full" />
            <div className="h-11 w-full bg-surface border border-border-subtle rounded-xl" />
          </div>
        </Card>
      ) : error ? (
        <Card className="p-5 text-center">
          <div className="w-10 h-10 mx-auto rounded-full bg-status-error/10 text-status-error flex items-center justify-center">
            <AlertTriangle className="w-5 h-5" />
          </div>
          <h2 className="text-sm font-bold text-text-primary mt-3">Data belum bisa dimuat</h2>
          <p className="text-xs text-text-secondary mt-1">Periksa koneksi internetmu dan coba lagi.</p>
          <Button onClick={loadTasks} variant="secondary" size="md" className="w-full mt-4">
            <RefreshCw className="w-4 h-4 mr-2" /> Coba Lagi
          </Button>
          {error && <p className="text-[11px] text-text-secondary mt-2 break-words">{error}</p>}
        </Card>
      ) : tasks.length === 0 ? (
        <Card className="p-6 text-center">
          <div className="w-12 h-12 mx-auto rounded-full bg-accent-magic/10 flex items-center justify-center text-xl">
            🌟
          </div>
          <h2 className="text-sm font-bold text-text-primary mt-3">Belum ada tugas hari ini</h2>
          <p className="text-xs text-text-secondary mt-1 leading-relaxed max-w-[30ch] mx-auto">
            Admin keluarga belum menambahkan tugas. Kamu akan mendapat notifikasi saat tugas baru tersedia.
          </p>
        </Card>
      ) : stats.isAllDone ? (
        <Card className="p-5">
          <div className="flex items-start gap-3">
            <div className="w-10 h-10 rounded-full bg-status-success/15 text-status-success flex items-center justify-center shrink-0">
              <CheckCircle2 className="w-5 h-5" />
            </div>
            <div className="flex-1 min-w-0">
              <h2 className="text-base font-bold text-text-primary">Semua tugas hari ini selesai 🎉</h2>
              <p className="text-xs text-text-secondary mt-1 leading-relaxed">
                Kerja bagus! Kamu telah menyelesaikan {stats.total} dari {stats.total} tugas. Koin sudah masuk ke saldomu.
              </p>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-2 mt-4">
            <Link to="/shop" className="inline-flex items-center justify-center gap-1.5 py-3 rounded-xl bg-accent-magic text-white font-bold text-sm shadow-sm hover:brightness-110 transition-colors">
              <Coins className="w-4 h-4" /> Lihat Hadiah
            </Link>
            <Link to="/profile" className="inline-flex items-center justify-center gap-1.5 py-3 rounded-xl bg-surface-elevated border border-border-subtle text-text-primary font-bold text-sm hover:bg-surface transition-colors">
              <Trophy className="w-4 h-4" /> Perkembangan
            </Link>
          </div>
        </Card>
      ) : stats.nextTask ? (
        <Card className="p-5">
          <div className="flex items-start justify-between gap-3">
            <div className="flex-1 min-w-0">
              <p className="text-[11px] font-bold tracking-wide uppercase text-text-secondary">Tugas Harian</p>
              <h2 className="text-base font-bold text-text-primary mt-1 leading-tight">
                {stats.completed === 0 ? 'Mulai tugas pertamamu' : 'Lanjutkan tugas berikutnya'}
              </h2>
              <p className="text-xs text-text-secondary mt-1 leading-relaxed">
                {stats.completed === 0
                  ? 'Selesaikan tugas untuk mendapatkan koin dan melanjutkan ke langkah berikutnya.'
                  : `Masih ada ${stats.total - stats.completed} tugas lagi. Ayo selesaikan!`}
              </p>
              <div className="mt-3 flex items-center gap-2 text-xs">
                <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-accent-magic/10 text-accent-magic border border-accent-magic/20 font-bold">
                  {getTaskIcon(stats.nextTask.task_type)} {stats.nextTask.title}
                </span>
              </div>
              <p className="text-[11px] text-text-secondary mt-2">
                Hadiah: <span className="font-bold text-accent-gold">+{stats.nextTask.reward_coins} koin</span> • +{stats.nextTask.reward_xp} XP • Langkah {stats.nextTask.step_order} dari {stats.total}
              </p>
            </div>
          </div>
          <Button onClick={() => handleTaskClick(stats.nextTask!)} size="lg" className="w-full mt-4">
            {stats.completed === 0 ? 'Mulai Tugas' : 'Lanjutkan Tugas'} <ArrowRight className="w-4 h-4 ml-2" />
          </Button>
          {stats.rejected > 0 && (
            <p className="text-xs text-status-error text-center mt-2 font-medium">Ada {stats.rejected} tugas perlu revisi — cek daftar di bawah.</p>
          )}
        </Card>
      ) : (
        <Card className="p-5 text-center">
          <div className="w-10 h-10 mx-auto rounded-full bg-accent-gold/15 text-accent-gold flex items-center justify-center">
            <Clock className="w-5 h-5" />
          </div>
          <h2 className="text-sm font-bold text-text-primary mt-3">Menunggu verifikasi</h2>
          <p className="text-xs text-text-secondary mt-1">Semua tugas sudah dikumpulkan. Koin akan masuk setelah admin memeriksa.</p>
        </Card>
      )}

      {/* 3. Progress — simple, non-technical */}
      {tasks.length > 0 && !loading && !error && (
        <Card className="p-4">
          <div className="flex items-center justify-between">
            <h3 className="text-xs font-bold text-text-primary">Perkembangan hari ini</h3>
            <span className="text-xs font-bold text-text-secondary">{stats.completed} dari {stats.total} selesai</span>
          </div>
          <div className="mt-3">
            <ProgressBar progress={stats.progressPercent} colorClass="bg-accent-magic" />
          </div>
          <div className="mt-3 flex items-center justify-between text-xs">
            <span className="inline-flex items-center gap-1.5 text-text-secondary">
              <Trophy className="w-3.5 h-3.5 text-accent-magic" />
              <span className="font-semibold">Lv. {userLevel}</span>
              <span className="text-text-secondary">• {xpInCurrentLevel}/{xpRequiredForLevel} XP</span>
            </span>
            <span className="inline-flex items-center gap-1 text-text-secondary">
              <Flame className="w-3.5 h-3.5 text-accent-danger" /> <span className="font-bold">{userStreak} Hari</span>
            </span>
          </div>
          {/* secondary XP bar kept faint for low-tech secondary */}
          <div className="mt-2 flex items-center gap-2 opacity-60">
            <div className="flex-1 h-1 bg-surface border border-border-subtle rounded-full overflow-hidden">
              <div className="h-full bg-accent-magic/50 rounded-full" style={{ width: `${xpProgressPercent}%` }} />
            </div>
          </div>
        </Card>
      )}

      {/* 4. Coins / Reward — supporting, not dominant */}
      <Card className="p-4">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-3 min-w-0">
            <span className="text-xl p-2 rounded-xl bg-accent-gold/15 border border-accent-gold/20 shrink-0">🪙</span>
            <div className="min-w-0">
              <p className="text-xs font-bold text-text-secondary uppercase tracking-wide">Saldo Koin</p>
              <p className="text-lg font-bold text-text-primary leading-none mt-0.5">
                {userCoins.toLocaleString('id-ID')} <span className="text-xs font-semibold text-text-secondary">Koin</span>
              </p>
              <p className="text-[11px] text-text-secondary mt-1">Kumpulkan dan tukarkan dengan hadiah.</p>
            </div>
          </div>
          <Link to="/shop" className="shrink-0 inline-flex items-center gap-1.5 px-3.5 py-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-text-primary font-bold text-xs hover:bg-surface transition-colors">
            Lihat Hadiah <ArrowRight className="w-3.5 h-3.5" />
          </Link>
        </div>
      </Card>

      {/* 5. Task list — readable, no bubble navigation */}
      {tasks.length > 0 && !error && (
        <div className="flex flex-col gap-2">
          <div className="flex items-center justify-between px-1">
            <h3 className="text-sm font-bold text-text-primary">Daftar tugas hari ini</h3>
            <button onClick={loadTasks} aria-label="Muat ulang tugas" disabled={loading} className="p-1.5 rounded-lg bg-surface border border-border-subtle text-text-secondary hover:text-text-primary disabled:opacity-50">
              <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
            </button>
          </div>
          <div className="flex flex-col gap-2">
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
                  className={`w-full text-left p-3 rounded-2xl border flex items-center gap-3 transition-colors ${isApproved ? 'bg-status-success/5 border-status-success/20' : isPending ? 'bg-accent-gold/5 border-accent-gold/20' : isUnlocked ? 'bg-surface border-border-subtle hover:border-accent-magic/30 hover:bg-surface-elevated shadow-sm' : isRejected ? 'bg-status-error/5 border-status-error/20' : 'bg-surface border-border-subtle opacity-60'}`}
                >
                  <span className={`w-9 h-9 rounded-xl flex items-center justify-center shrink-0 border ${isApproved ? 'bg-status-success text-white border-status-success' : isPending ? 'bg-accent-gold text-white border-accent-gold' : isUnlocked ? 'bg-accent-magic text-white border-accent-magic' : isRejected ? 'bg-status-error text-white border-status-error' : 'bg-surface-elevated text-text-secondary border-border-subtle'}`}>
                    {isApproved ? <CheckCircle2 className="w-4 h-4" /> : isPending ? <Clock className="w-4 h-4" /> : isLocked ? <Lock className="w-4 h-4" /> : isRejected ? <AlertTriangle className="w-4 h-4" /> : getTaskIcon(task.task_type)}
                  </span>
                  <span className="flex-1 min-w-0">
                    <span className={`text-sm font-bold leading-tight block truncate ${isLocked ? 'text-text-secondary' : 'text-text-primary'}`}>{task.title}</span>
                    <span className="text-xs text-text-secondary line-clamp-1">{task.description || (isLocked ? 'Selesaikan tugas sebelumnya untuk membuka' : `Langkah ${task.step_order} • +${task.reward_coins} koin`)}</span>
                  </span>
                  <span className={`text-[11px] font-bold px-2 py-1 rounded-full border shrink-0 ${isApproved ? 'bg-status-success text-white border-status-success' : isPending ? 'bg-accent-gold/15 text-accent-gold border-accent-gold/20' : isUnlocked ? 'bg-accent-magic text-white border-accent-magic' : isRejected ? 'bg-status-error/15 text-status-error border-status-error/20' : 'bg-surface-elevated text-text-secondary border-border-subtle'}`}>
                    {isApproved ? 'Selesai' : isPending ? 'Menunggu' : isUnlocked ? 'Kerjakan' : isRejected ? 'Revisi' : 'Terkunci'}
                  </span>
                </button>
              )
            })}
          </div>
        </div>
      )}

      {/* Modals — preserved */}
      {activeModalTask && (() => {
        const cfg = activeModalTask.config || {}
        const isDoc = activeModalTask.task_type === 'DOCUMENT_UPLOAD' || Boolean(cfg.attachment_url)
        const isPhoto = activeModalTask.task_type === 'PHOTO_UPLOAD' || activeModalTask.task_type === 'PHOTO_PROOF'
        const isGame = activeModalTask.task_type === 'MINI_GAME' || Boolean(cfg.game)
        const isText = activeModalTask.task_type === 'TEXT_RESPONSE' || (!isDoc && !isPhoto && Boolean(cfg.minimum_characters || cfg.prompt))
        if (isDoc) return <DocUploadModal task={activeModalTask} onClose={() => setActiveModalTask(null)} onSuccess={handleModalSuccess} />
        if (isPhoto) return <CameraCaptureModal task={activeModalTask} onClose={() => setActiveModalTask(null)} onSuccess={handleModalSuccess} />
        if (isText) return <TextResponseModal task={activeModalTask} onClose={() => setActiveModalTask(null)} onSuccess={handleModalSuccess} />
        if (isGame) return <MiniGameModal task={activeModalTask} onClose={() => setActiveModalTask(null)} onSuccess={handleModalSuccess} />
        return <VideoQuizModal task={activeModalTask} onClose={() => setActiveModalTask(null)} onSuccess={handleModalSuccess} />
      })()}
    </div>
  )
}
