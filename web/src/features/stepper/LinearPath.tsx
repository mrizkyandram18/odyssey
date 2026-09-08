import React, { useState, useEffect, useCallback, useMemo } from 'react'
import { Link } from 'react-router-dom'
import { Flame, Coins, Trophy, Calendar, RefreshCw, CheckCircle2, Clock, Lock, ArrowRight, AlertTriangle, Play, Video, FileText, Camera, HelpCircle, PenLine, Gamepad2, Sparkles, Banknote } from 'lucide-react'
import type { TaskView, RedemptionConfig } from '../../shared/types'
import { tasksApi, shopApi } from '../../shared/lib/api'
import { useSession } from '../../shared/hooks/useSession'
import { Card } from '../../shared/components/atoms/Card'
import { Button } from '../../shared/components/atoms/Button'
import { VideoQuizModal } from './VideoQuizModal'
import { VideoRecordModal } from './VideoRecordModal'
import { DecisionGameModal } from './DecisionGameModal'
import { DocUploadModal } from './DocUploadModal'
import { CameraCaptureModal } from './CameraCaptureModal'
import { LiveCameraCaptureModal } from './LiveCameraCaptureModal'
import { TextResponseModal } from './TextResponseModal'
import { MiniGameModal } from './MiniGameModal'
import { TicketCard } from '../home/TicketCard'
import { levelProgress } from '../../shared/lib/level'

// Helper: greeting by time
function getGreeting() {
  const h = new Date().getHours()
  if (h < 11) return 'Selamat pagi'
  if (h < 15) return 'Selamat siang'
  if (h < 18) return 'Selamat sore'
  return 'Selamat malam'
}

function getTaskIcon(task: TaskView) {
  if (task.task_type === 'VIDEO' && (task.config as any)?.recording?.enabled) {
    return <Video className="w-4 h-4" />
  }
  switch (task.task_type) {
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

// Helper for legacy camera-only tasks that might not have camera_only in config yet
export function isLegacyCameraOnlyTask(task: TaskView): boolean {
  const title = (task.title || '').trim()
  return (
    title === 'Foto Langsung Kesiapan Profil CV (Rapi & Profesional)' ||
    title === 'Bukti Diskusi'
  )
}

export const LinearPath: React.FC = () => {
  const { profile, refreshProfile, session } = useSession()
  const [tasks, setTasks] = useState<TaskView[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeModalTask, setActiveModalTask] = useState<TaskView | null>(null)
  const [shopConfig, setShopConfig] = useState<RedemptionConfig | null>(null)
  const [earningLocked, setEarningLocked] = useState(false)
  const [earningCap, setEarningCap] = useState<number | null>(null)
  const [earned, setEarned] = useState<number | null>(null)
  const requestIdRef = React.useRef(0)

  const loadTasks = useCallback(async () => {
    const reqId = ++requestIdRef.current
    try {
      setLoading(true)
      setError(null)
      const data = await tasksApi.getToday()
      // Race guard: ignore stale response
      if (reqId !== requestIdRef.current) return
      setTasks(data.tasks || [])
      setEarningLocked(Boolean(data.earning_locked))
      setEarningCap(data.earning_cap ?? null)
      setEarned(data.earned ?? null)
    } catch (err: any) {
      if (reqId !== requestIdRef.current) return
      const msg = err.message || 'Gagal memuat alur tugas harian.'
      setError(msg)
      // If cap reached, also lock UI from error path
      if (msg.includes('EARNING_CAP_REACHED') || msg.includes('batas earning')) {
        setEarningLocked(true)
      }
    } finally {
      if (reqId === requestIdRef.current) setLoading(false)
    }
  }, [])

  // Load on mount and when session (user) changes — prevents User A data leaking to User B
  useEffect(() => { loadTasks() }, [loadTasks, session?.uid])
  // Clear stale tasks immediately when user switches
  useEffect(() => {
    setTasks([])
    setEarningLocked(false)
    setEarningCap(null)
    setEarned(null)
  }, [session?.uid])

  useEffect(() => {
    shopApi.getConfig().then((c) => setShopConfig(c)).catch(() => {})
  }, [])

  // Refetch on window focus / reconnect / visibility — ensures period rollover and cap updates propagate without manual reload
  useEffect(() => {
    const onFocus = () => loadTasks()
    const onOnline = () => loadTasks()
    const onVisibility = () => { if (document.visibilityState === 'visible') loadTasks() }
    window.addEventListener('focus', onFocus)
    window.addEventListener('online', onOnline)
    document.addEventListener('visibilitychange', onVisibility)
    return () => {
      window.removeEventListener('focus', onFocus)
      window.removeEventListener('online', onOnline)
      document.removeEventListener('visibilitychange', onVisibility)
    }
  }, [loadTasks])

  const handleTaskClick = (task: TaskView) => {
    if (earningLocked) return
    if (task.earning_locked || task.is_locked || task.status === 'LOCKED') return
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
      const currentIndex = tasks.findIndex((t) => t.id === currentTask.id)
      const next = currentIndex >= 0 ? tasks[currentIndex + 1] : undefined
      if (next) {
        // Do not open next if earning is locked
        if (earningLocked || next.earning_locked) {
          setActiveModalTask(null)
          return
        }
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

  // Level progress curve based on server-side formula: level = floor(sqrt(xp/100)) + 1
  const userProgress = levelProgress(userXP, userLevel)

  return (
    <div className="w-full flex flex-col gap-4">
      {/* 1. Greeting & Adventurer Status Bar */}
      <div className="px-1 pt-1 flex items-start justify-between gap-3">
        <div>
          <p className="text-[11px] font-black tracking-wider uppercase text-accent-magic flex items-center gap-1.5">
            <span>🧭</span> {getGreeting()}
          </p>
          <h1 className="text-[22px] font-extrabold text-text-primary leading-tight mt-0.5 tracking-tight">
            {explorerName ? `Halo, ${explorerName}!` : 'Petualangan Hari Ini'}
          </h1>
          <p className="text-xs text-text-secondary mt-1 flex items-center gap-1.5">
            <Calendar className="w-3.5 h-3.5 shrink-0 text-text-secondary/70" />
            <span>{new Date().toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })}</span>
          </p>
        </div>

        {/* Status Badges */}
        <div className="flex flex-col items-end gap-1.5 shrink-0">
          {userStreak > 0 ? (
            <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-amber-500/10 text-amber-600 dark:text-amber-400 font-extrabold text-xs border border-amber-500/20 shadow-sm">
              <Flame className="w-3.5 h-3.5 fill-amber-500 text-amber-500" />
              <span>{userStreak} Hari</span>
            </span>
          ) : (
            <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-accent-magic/10 text-accent-magic font-bold text-xs border border-accent-magic/20">
              <Sparkles className="w-3.5 h-3.5" /> Petualang
            </span>
          )}
          <span className="text-[11px] font-extrabold px-2 py-0.5 rounded-full bg-surface-elevated border border-border-subtle text-text-secondary">
            ⭐ Tingkat {userProgress.level}
          </span>
        </div>
      </div>

      {/* 2. Hero Adventure Card ("Petualangan Hari Ini") */}
      {loading && tasks.length === 0 ? (
        <Card className="p-5">
          <div className="animate-pulse space-y-3">
            <div className="h-3 w-28 bg-surface-elevated rounded-full" />
            <div className="h-6 w-48 bg-surface-elevated rounded-lg" />
            <div className="h-3 w-full bg-surface-elevated rounded-full" />
            <div className="h-12 w-full bg-surface-elevated rounded-xl" />
          </div>
        </Card>
      ) : error ? (
        <Card className="p-5 text-center">
          <div className="w-11 h-11 mx-auto rounded-2xl bg-status-error/10 text-status-error flex items-center justify-center">
            <AlertTriangle className="w-6 h-6" />
          </div>
          <h2 className="text-sm font-bold text-text-primary mt-3">Petualangan belum bisa dimuat</h2>
          <p className="text-xs text-text-secondary mt-1 leading-relaxed">Periksa koneksi internetmu dan coba lagi.</p>
          <Button onClick={loadTasks} variant="secondary" size="md" className="w-full mt-4 font-bold">
            <RefreshCw className="w-4 h-4 mr-2" /> Coba Lagi
          </Button>
          {error && <p className="text-[11px] text-text-secondary mt-2 break-words">{error}</p>}
        </Card>
      ) : tasks.length === 0 ? (
        <div className="rounded-3xl border-2 border-dashed border-border-subtle bg-surface/80 p-8 text-center backdrop-blur-sm">
          <div className="w-14 h-14 mx-auto rounded-2xl bg-accent-magic/10 flex items-center justify-center text-2xl shadow-inner">
            🌟
          </div>
          <h2 className="text-base font-extrabold text-text-primary mt-3.5">Belum ada petualangan hari ini</h2>
          <p className="text-xs text-text-secondary mt-1.5 leading-relaxed max-w-[32ch] mx-auto">
            Misi harian sedang disiapkan. Kamu akan mendapatkan notifikasi begitu misi baru tersedia!
          </p>
        </div>
      ) : stats.isAllDone ? (
        <div className="relative overflow-hidden rounded-3xl border border-emerald-500/30 bg-gradient-to-br from-emerald-500/10 via-emerald-500/5 to-transparent p-5 shadow-sm">
          <div className="flex items-start gap-3.5">
            <div className="w-12 h-12 rounded-2xl bg-status-success text-white flex items-center justify-center shrink-0 shadow-md shadow-emerald-500/25">
              <CheckCircle2 className="w-7 h-7 stroke-[2.5]" />
            </div>
            <div className="flex-1 min-w-0">
              <span className="inline-flex items-center gap-1 text-[11px] font-black uppercase tracking-wider text-status-success">
                Misi Sempurna ✨
              </span>
              <h2 className="text-base font-extrabold text-text-primary leading-tight mt-0.5">
                Semua tugas hari ini selesai 🎉
              </h2>
              <p className="text-xs text-text-secondary mt-1 leading-relaxed">
                Hebat sekali! Kamu telah menuntaskan seluruh {stats.total} petualangan hari ini. Koin & Bintang sudah masuk ke kantongmu!
              </p>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-2 mt-4 pt-3 border-t border-emerald-500/20">
            <Link
              to="/shop"
              className="inline-flex items-center justify-center gap-1.5 py-3 rounded-xl bg-accent-magic text-white font-bold text-xs shadow-sm hover:brightness-110 active:scale-95 transition-all min-h-[44px]"
            >
              <Coins className="w-4 h-4" /> Tukar Koin
            </Link>
            <Link
              to="/koleksi"
              className="inline-flex items-center justify-center gap-1.5 py-3 rounded-xl bg-gradient-to-r from-amber-500 to-orange-500 text-white font-bold text-xs shadow-sm hover:brightness-110 active:scale-95 transition-all min-h-[44px]"
            >
              🎁 Buka Hadiah
            </Link>
            <Link
              to="/profile"
              className="col-span-2 inline-flex items-center justify-center gap-1.5 py-2.5 rounded-xl bg-surface border border-border-subtle text-text-primary font-bold text-xs hover:bg-surface-elevated transition-colors"
            >
              <Trophy className="w-3.5 h-3.5 text-accent-gold" /> Lihat Perkembangan Profil
            </Link>
          </div>
        </div>
      ) : stats.nextTask ? (
        <div className="relative overflow-hidden rounded-3xl border border-accent-magic/25 bg-gradient-to-br from-accent-magic/10 via-surface to-surface p-5 shadow-sm">
          {/* Header pill & Progress count */}
          <div className="flex items-center justify-between gap-2 flex-wrap">
            <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-accent-magic/15 text-accent-magic border border-accent-magic/20 font-black text-[11px] tracking-wide uppercase">
              <Sparkles className="w-3.5 h-3.5" />
              <span>Tugas Harian</span>
              <span>•</span>
              <span>Langkah {stats.nextTask.step_order} dari {stats.total}</span>
            </span>

            <span className="text-[11px] font-extrabold text-text-secondary">
              {stats.completed}/{stats.total} Misi Selesai
            </span>
          </div>

          {/* Quest Title & Description */}
          <div className="mt-3">
            <h2 className="text-lg font-extrabold text-text-primary leading-tight tracking-tight">
              {stats.nextTask.title}
            </h2>
            <p className="text-xs text-text-secondary mt-1.5 leading-relaxed">
              {stats.nextTask.description || (stats.completed === 0
                ? 'Selesaikan tugas untuk mengumpulkan koin & membuka langkah berikutnya!'
                : `Masih ada ${stats.total - stats.completed} tugas lagi. Ayo lanjutkan perjalananmu!`)}
            </p>
          </div>

          {/* Reward Preview Chips */}
          <div className="mt-3.5 flex items-center gap-2 flex-wrap">
            <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg bg-amber-500/10 text-amber-700 dark:text-amber-300 border border-amber-500/20 font-black text-xs">
              <Coins className="w-3.5 h-3.5 text-amber-500" />
              <span>+{stats.nextTask.reward_coins} Koin</span>
            </span>
            <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg bg-sky-500/10 text-sky-700 dark:text-sky-300 border border-sky-500/20 font-black text-xs">
              <Sparkles className="w-3.5 h-3.5 text-sky-500" />
              <span>+{stats.nextTask.reward_xp} Bintang</span>
            </span>
          </div>

          {/* Primary CTA Button */}
          <button
            onClick={() => handleTaskClick(stats.nextTask!)}
            disabled={earningLocked}
            className="w-full mt-4 py-3.5 px-4 rounded-2xl bg-gradient-to-r from-accent-magic to-sky-600 hover:from-accent-magic/90 hover:to-sky-600/90 text-white font-black text-sm shadow-md shadow-accent-magic/25 hover:shadow-lg hover:shadow-accent-magic/35 active:scale-[0.98] transition-all flex items-center justify-center gap-2 min-h-[46px] disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <span>{earningLocked ? 'Batas Koin Tercapai' : stats.completed === 0 ? 'Mulai Tugas' : 'Lanjutkan Tugas'}</span>
            <ArrowRight className="w-4 h-4" />
          </button>

          {stats.rejected > 0 && (
            <p className="text-xs text-status-error text-center mt-2.5 font-bold">
              ⚠️ Ada {stats.rejected} tugas perlu revisi — cek daftar di bawah.
            </p>
          )}
        </div>
      ) : (
        <Card className="p-5 text-center">
          <div className="w-11 h-11 mx-auto rounded-2xl bg-accent-gold/15 text-accent-gold flex items-center justify-center">
            <Clock className="w-6 h-6" />
          </div>
          <h2 className="text-sm font-bold text-text-primary mt-3">Menunggu verifikasi admin</h2>
          <p className="text-xs text-text-secondary mt-1 leading-relaxed">
            Semua tugas sudah kamu kumpulkan. Koin akan otomatis masuk setelah admin memeriksa dokumen bukti.
          </p>
        </Card>
      )}

      {/* 3. Ticket Card — prominent voucher */}
      <TicketCard />

      {/* Earning cap HALTED banner — authoritative from backend */}
      {earningLocked && !loading && (
        <div className="rounded-2xl border-2 border-amber-300 bg-amber-50 dark:bg-amber-950/40 p-4">
          <div className="flex items-start gap-3">
            <div className="w-9 h-9 rounded-xl bg-amber-200/80 dark:bg-amber-900 flex items-center justify-center shrink-0">
              <Lock className="w-4 h-4 text-amber-800 dark:text-amber-200" />
            </div>
            <div className="flex-1 min-w-0">
              <h3 className="text-sm font-extrabold text-amber-950 dark:text-amber-200">Batas Koin Bulanan tercapai</h3>
              <p className="text-xs text-amber-800/90 dark:text-amber-300 mt-1 leading-relaxed">
                Kamu sudah mencapai {earned ?? '—'} / {earningCap ?? '—'} koin periode ini (1–24). Tugas tetap terlihat, tapi tidak menghasilkan koin sampai periode berikutnya. Saldo {userCoins.toLocaleString('id-ID')} koin tetap aman.
              </p>
            </div>
          </div>
        </div>
      )}

      {/* 4. Gamified Progression & Currency Strip */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        {/* Tingkat & Bintang Journey Card */}
        <div className="rounded-2xl bg-surface border border-border-subtle p-4 flex flex-col justify-between gap-2.5">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-xl bg-accent-magic/10 text-accent-magic flex items-center justify-center font-black text-sm">
                ⭐
              </div>
              <div>
                <p className="text-[10px] font-black uppercase tracking-wider text-text-secondary">Tingkat Penjelajah</p>
                <p className="text-sm font-black text-text-primary leading-none mt-0.5">
                  Tingkat {userProgress.level}
                </p>
              </div>
            </div>
            <span className="text-xs font-extrabold text-accent-magic">
              {userProgress.have}/{userProgress.required} Bintang
            </span>
          </div>

          <div className="space-y-1">
            <div className="h-2 w-full bg-surface-elevated rounded-full overflow-hidden border border-border-subtle/50">
              <div
                className="h-full bg-gradient-to-r from-accent-magic to-sky-400 rounded-full transition-all duration-500"
                style={{ width: `${userProgress.percent}%` }}
              />
            </div>
            <p className="text-[10.5px] text-text-secondary text-right">
              {userProgress.required - userProgress.have} Bintang lagi menuju Tingkat {userProgress.level + 1}
            </p>
          </div>
        </div>

        {/* Coin Balance Card */}
        <div className="rounded-2xl bg-surface border border-border-subtle p-4 flex flex-col justify-between gap-2.5">
          {(() => {
            const conv = shopConfig ? shopConfig.conversion_rate : 0
            const estCash = userCoins * conv
            const isPayoutDay = shopConfig?.is_payout_day
            return (
              <>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2.5">
                    <div className="w-8 h-8 rounded-xl bg-amber-500/10 text-amber-500 flex items-center justify-center text-lg">
                      🪙
                    </div>
                    <div>
                      <p className="text-[10px] font-black uppercase tracking-wider text-text-secondary">Saldo Koin</p>
                      <p className="text-base font-black text-text-primary leading-none mt-0.5">
                        {userCoins.toLocaleString('id-ID')} <span className="text-xs font-semibold text-text-secondary">Koin</span>
                      </p>
                    </div>
                  </div>

                  <Link
                    to="/shop"
                    className="inline-flex items-center gap-1 px-3 py-1.5 rounded-xl bg-accent-magic/10 hover:bg-accent-magic/20 text-accent-magic font-extrabold text-xs transition-colors"
                  >
                    <span>Tukar</span>
                    <ArrowRight className="w-3.5 h-3.5" />
                  </Link>
                </div>

                <div className="text-[11px] text-text-secondary flex items-center justify-between pt-1 border-t border-border-subtle/60">
                  <span>Estimasi Uang Tunai</span>
                  <span className="font-bold text-text-primary">≈ Rp {estCash.toLocaleString('id-ID')}</span>
                </div>

                {isPayoutDay && userCoins > 0 && (
                  <div className="p-2 rounded-xl bg-status-success/10 border border-status-success/20 flex items-center gap-1.5 text-xs text-status-success font-bold">
                    <Banknote className="w-3.5 h-3.5 shrink-0" />
                    <span>Periode pencairan aktif!</span>
                  </div>
                )}
              </>
            )
          })()}
        </div>
      </div>

      {/* 5. Quest Stepper List ("Misi Petualangan Hari Ini") */}
      {tasks.length > 0 && !error && (
        <div className="flex flex-col gap-2.5 mt-1">
          <div className="flex items-center justify-between px-1">
            <div className="flex items-center gap-2">
              <span className="text-sm">🗺️</span>
              <h3 className="text-xs font-extrabold tracking-wider uppercase text-text-primary">
                Jalur Petualangan Hari Ini
              </h3>
              <span className="text-[11px] font-bold px-2 py-0.5 rounded-full bg-accent-magic/10 text-accent-magic border border-accent-magic/15">
                {stats.completed}/{stats.total} Selesai
              </span>
            </div>
            <button
              onClick={loadTasks}
              aria-label="Muat ulang tugas"
              disabled={loading}
              className="w-7 h-7 rounded-full bg-surface border border-border-subtle text-text-secondary hover:text-text-primary disabled:opacity-50 flex items-center justify-center transition-colors shadow-sm"
            >
              <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
            </button>
          </div>

          <div className="flex flex-col gap-2">
            {tasks.map((task) => {
              const isApproved = task.status === 'APPROVED'
              const isPending = task.status === 'PENDING'
              const isUnlocked = task.status === 'UNLOCKED' && !earningLocked && !task.earning_locked
              const isLocked = task.status === 'LOCKED' || task.is_locked || earningLocked || Boolean(task.earning_locked)
              const isRejected = task.status === 'REJECTED'

              return (
                <button
                  key={task.id}
                  onClick={() => handleTaskClick(task)}
                  disabled={isLocked}
                  className={`w-full text-left p-3.5 rounded-2xl border transition-all duration-200 flex items-center gap-3 min-h-[64px] ${
                    isApproved
                      ? 'bg-emerald-500/[0.04] border-emerald-500/20 hover:border-emerald-500/35'
                      : isPending
                      ? 'bg-amber-500/[0.04] border-amber-500/20 hover:border-amber-500/35'
                      : isUnlocked
                      ? 'bg-surface border-accent-magic/30 shadow-sm hover:shadow-md hover:border-accent-magic hover:bg-surface-elevated'
                      : isRejected
                      ? 'bg-red-500/[0.04] border-red-500/20 hover:border-red-500/35'
                      : 'bg-surface/60 border-border-subtle opacity-65 cursor-not-allowed'
                  }`}
                >
                  {/* Step Order / Status Badge Node */}
                  <span
                    className={`w-10 h-10 rounded-2xl flex items-center justify-center shrink-0 font-extrabold text-sm shadow-sm transition-transform ${
                      isApproved
                        ? 'bg-status-success text-white shadow-emerald-500/20'
                        : isPending
                        ? 'bg-accent-gold text-white shadow-amber-500/20'
                        : isUnlocked
                        ? 'bg-accent-magic text-white shadow-accent-magic/25 ring-2 ring-accent-magic/20 scale-105'
                        : isRejected
                        ? 'bg-status-error text-white shadow-red-500/20'
                        : 'bg-surface-elevated text-text-secondary/70 border border-border-subtle'
                    }`}
                  >
                    {isApproved ? (
                      <CheckCircle2 className="w-5 h-5 stroke-[2.5]" />
                    ) : isPending ? (
                      <Clock className="w-5 h-5" />
                    ) : isLocked ? (
                      <Lock className="w-4 h-4" />
                    ) : isRejected ? (
                      <AlertTriangle className="w-4 h-4" />
                    ) : (
                      getTaskIcon(task)
                    )}
                  </span>

                  {/* Task Content */}
                  <span className="flex-1 min-w-0">
                    <span className="flex items-center gap-1.5 flex-wrap">
                      <span className="text-[11px] font-bold text-text-secondary">
                        Langkah #{task.step_order}
                      </span>
                    </span>
                    <span
                      className={`text-sm font-extrabold leading-snug block truncate mt-0.5 ${
                        isLocked ? 'text-text-secondary' : 'text-text-primary'
                      }`}
                    >
                      {task.title}
                    </span>
                    <span className="text-xs text-text-secondary line-clamp-1 mt-0.5">
                      {task.description || (isLocked ? 'Selesaikan tugas sebelumnya untuk membuka' : `+${task.reward_coins} koin • +${task.reward_xp} Bintang`)}
                    </span>
                  </span>

                  {/* Right Status / Action Pill */}
                  <span
                    className={`text-[11px] font-black px-3 py-1.5 rounded-xl shrink-0 whitespace-nowrap shadow-xs ${
                      isApproved
                        ? 'bg-status-success/15 text-status-success border border-status-success/20'
                        : isPending
                        ? 'bg-accent-gold/15 text-amber-700 dark:text-amber-300 border border-accent-gold/25'
                        : isUnlocked
                        ? 'bg-accent-magic text-white shadow-sm shadow-accent-magic/25'
                        : isRejected
                        ? 'bg-status-error/15 text-status-error border border-status-error/25'
                        : 'bg-surface-elevated text-text-secondary border border-border-subtle'
                    }`}
                  >
                    {isApproved
                      ? '✓ Selesai'
                      : isPending
                      ? 'Menunggu'
                      : isUnlocked
                      ? 'Kerjakan'
                      : isRejected
                      ? 'Revisi'
                      : 'Terkunci'}
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
        const isLiveCamera =
          cfg.camera_only !== undefined
            ? Boolean(cfg.camera_only)
            : isLegacyCameraOnlyTask(activeModalTask)
        const isDoc = activeModalTask.task_type === 'DOCUMENT_UPLOAD' || Boolean(cfg.attachment_url)
        const isPhoto = activeModalTask.task_type === 'PHOTO_UPLOAD' || activeModalTask.task_type === 'PHOTO_PROOF'
        const isVideoRecording = activeModalTask.task_type === 'VIDEO' && Boolean(cfg.recording?.enabled)
        const hasDecisionScenario = activeModalTask.task_type === 'MINI_GAME' && Array.isArray(cfg.scenario?.events) && cfg.scenario.events.length > 0
        const isGame = activeModalTask.task_type === 'MINI_GAME' || Boolean(cfg.game)
        const isText = activeModalTask.task_type === 'TEXT_RESPONSE' || (!isDoc && !isPhoto && Boolean(cfg.minimum_characters || cfg.prompt))
        // eslint-disable-next-line react-hooks/refs -- onNextTask is event handler, not ref read during render
        if (isVideoRecording) return <VideoRecordModal task={activeModalTask} onClose={() => setActiveModalTask(null)} onSuccess={handleModalSuccess} onNextTask={() => handleNextTask(activeModalTask)} />
        if (isLiveCamera && isPhoto) return <LiveCameraCaptureModal task={activeModalTask} onClose={() => setActiveModalTask(null)} onSuccess={handleModalSuccess} onNextTask={() => handleNextTask(activeModalTask)} />
        if (isDoc) return <DocUploadModal task={activeModalTask} onClose={() => setActiveModalTask(null)} onSuccess={handleModalSuccess} onNextTask={() => handleNextTask(activeModalTask)} />
        if (isPhoto) return <CameraCaptureModal task={activeModalTask} onClose={() => setActiveModalTask(null)} onSuccess={handleModalSuccess} onNextTask={() => handleNextTask(activeModalTask)} />
        if (hasDecisionScenario) return <DecisionGameModal task={activeModalTask} onClose={() => setActiveModalTask(null)} onSuccess={handleModalSuccess} onNextTask={() => handleNextTask(activeModalTask)} />
        if (isText) return <TextResponseModal task={activeModalTask} onClose={() => setActiveModalTask(null)} onSuccess={handleModalSuccess} onNextTask={() => handleNextTask(activeModalTask)} />
        if (isGame) return <MiniGameModal task={activeModalTask} onClose={() => setActiveModalTask(null)} onSuccess={handleModalSuccess} onNextTask={() => handleNextTask(activeModalTask)} />
        return <VideoQuizModal task={activeModalTask} onClose={() => setActiveModalTask(null)} onSuccess={handleModalSuccess} onNextTask={() => handleNextTask(activeModalTask)} />
      })()}
    </div>
  )
}
