import React, { useState, useEffect, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Flame, Coins, Trophy, Sparkles, RefreshCw, Calendar } from 'lucide-react'
import type { TaskView } from '../../shared/types'
import { tasksApi } from '../../shared/lib/api'
import { useSession } from '../../shared/hooks/useSession'
import { StepNode } from './StepNode'
import { VideoQuizModal } from './VideoQuizModal'
import { DocUploadModal } from './DocUploadModal'
import { CameraCaptureModal } from './CameraCaptureModal'
import { TextResponseModal } from './TextResponseModal'
import { MiniGameModal } from './MiniGameModal'

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

  useEffect(() => {
    loadTasks()
  }, [loadTasks])

  const handleTaskClick = (task: TaskView) => {
    if (task.is_locked || task.status === 'LOCKED') return
    setActiveModalTask(task)
  }

  const handleModalSuccess = () => {
    loadTasks()
    if (refreshProfile) {
      refreshProfile()
    }
  }

  // Calculate XP percentage to next level
  const userXP = profile?.xp || 0
  const userLevel = profile?.level || 1
  const userCoins = profile?.coins || 0
  const userStreak = (profile as any)?.streak_days || 1

  const currentLevelBaseXP = Math.pow(userLevel - 1, 2) * 100
  const nextLevelXP = Math.pow(userLevel, 2) * 100
  const xpInCurrentLevel = Math.max(0, userXP - currentLevelBaseXP)
  const xpRequiredForLevel = Math.max(1, nextLevelXP - currentLevelBaseXP)
  const xpProgressPercent = Math.min(100, Math.round((xpInCurrentLevel / xpRequiredForLevel) * 100))

  return (
    <div className="w-full max-w-md mx-auto min-h-[calc(100vh-80px)] pb-24 flex flex-col">
      {/* 1. Gamified Header Bar */}
      <header className="sticky top-0 z-20 bg-surface-elevated/95 backdrop-blur-md border-b border-border-subtle px-4 py-3 shadow-sm">
        <div className="flex items-center justify-between gap-2">
          {/* Streak Flame */}
          <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-accent-danger/10 border border-accent-danger/20 text-accent-danger font-bold text-xs shadow-inner">
            <Flame className="w-4 h-4 fill-current animate-pulse" />
            <span>{userStreak} Hari</span>
          </div>

          {/* Coins Balance */}
          <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-accent-gold/15 border border-accent-gold/30 text-accent-gold font-bold text-xs shadow-inner">
            <Coins className="w-4 h-4" />
            <span>{userCoins.toLocaleString('id-ID')}</span>
          </div>

          {/* Level Badge */}
          <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-accent-magic/10 border border-accent-magic/20 text-accent-magic font-bold text-xs">
            <Trophy className="w-3.5 h-3.5" />
            <span>Lv. {userLevel}</span>
          </div>
        </div>

        {/* EXP Progress Bar */}
        <div className="mt-2 flex items-center gap-2">
          <div className="flex-1 h-2 bg-surface-base rounded-full overflow-hidden border border-border-subtle">
            <motion.div
              initial={{ width: 0 }}
              animate={{ width: `${xpProgressPercent}%` }}
              transition={{ duration: 0.8, ease: 'easeOut' }}
              className="h-full bg-gradient-to-r from-accent-magic to-accent-cyan rounded-full"
            />
          </div>
          <span className="text-[10px] text-text-secondary font-medium shrink-0">
            {xpInCurrentLevel} / {xpRequiredForLevel} XP
          </span>
        </div>
      </header>

      {/* 2. Today's Banner */}
      <div className="p-4 flex items-center justify-between">
        <div>
          <h2 className="text-xl font-heading font-extrabold text-text-primary flex items-center gap-2">
            <span>Tugas Harian</span>
            <Sparkles className="w-4 h-4 text-accent-gold" />
          </h2>
          <p className="text-xs text-text-secondary mt-0.5 flex items-center gap-1">
            <Calendar className="w-3.5 h-3.5" />
            <span>{new Date().toLocaleDateString('id-ID', { dateStyle: 'full' })}</span>
          </p>
        </div>
        <button
          onClick={loadTasks}
          disabled={loading}
          className="p-2 rounded-xl bg-surface-elevated border border-border-subtle text-text-secondary hover:text-text-primary active:scale-95 transition-all"
        >
          <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
        </button>
      </div>

      {/* 3. Main Linear Task Stepper */}
      <div className="flex-1 px-4 flex flex-col items-center justify-start relative my-4">
        {loading && tasks.length === 0 ? (
          <div className="py-20 flex flex-col items-center gap-3">
            <div className="w-12 h-12 rounded-full border-4 border-accent-magic border-t-transparent animate-spin" />
            <p className="text-xs text-text-secondary font-heading tracking-wider uppercase">
              Memuat Alur Tugas...
            </p>
          </div>
        ) : error ? (
          <div className="p-6 text-center bg-status-error/10 border border-status-error/20 rounded-3xl max-w-xs my-10">
            <p className="text-sm text-status-error font-bold mb-3">{error}</p>
            <button
              onClick={loadTasks}
              className="px-4 py-2 rounded-xl bg-surface-elevated text-xs font-bold text-text-primary shadow"
            >
              Coba Lagi
            </button>
          </div>
        ) : tasks.length === 0 ? (
          <div className="p-8 text-center bg-surface-elevated border border-border-subtle rounded-3xl my-12 max-w-xs space-y-3">
            <div className="w-16 h-16 mx-auto rounded-full bg-accent-magic/10 flex items-center justify-center text-2xl">
              🌟
            </div>
            <h4 className="font-heading font-bold text-text-primary text-base">
              Belum Ada Tugas Hari Ini
            </h4>
            <p className="text-xs text-text-secondary leading-relaxed">
              Admin belum menambahkan jadwal tugas untuk hari ini. Silakan hubungi pengelola keluarga.
            </p>
          </div>
        ) : (
          <div className="w-full flex flex-col items-center relative py-4">
            {/* SVG Connecting Curve Line */}
            <svg
              className="absolute top-8 left-0 right-0 w-full h-[calc(100%-60px)] pointer-events-none stroke-border-subtle"
              style={{ zIndex: 0 }}
            >
              <defs>
                <linearGradient id="pathGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                  <stop offset="0%" stopColor="var(--color-accent-magic)" stopOpacity="0.4" />
                  <stop offset="100%" stopColor="var(--color-accent-gold)" stopOpacity="0.4" />
                </linearGradient>
              </defs>
            </svg>

            {/* Step Nodes Sequence */}
            {tasks.map((task, idx) => (
              <StepNode
                key={task.id}
                task={task}
                index={idx}
                onClick={handleTaskClick}
              />
            ))}

            {/* End of Path Trophy Chest */}
            <div className="mt-8 flex flex-col items-center text-center">
              <div className="w-14 h-14 rounded-2xl bg-gradient-to-br from-accent-gold to-accent-magic text-white flex items-center justify-center shadow-lg shadow-accent-gold/30">
                <Trophy className="w-7 h-7" />
              </div>
              <p className="text-xs font-heading font-bold text-text-secondary mt-2">
                Target Hari Ini Selesai
              </p>
            </div>
          </div>
        )}
      </div>

      {/* 4. Active Modals based on Task Type */}
      <AnimatePresence>
        {activeModalTask && (() => {
          const cfg = activeModalTask.config || {}
          const isDoc = activeModalTask.task_type === 'DOCUMENT_UPLOAD' || Boolean(cfg.attachment_url)
          const isPhoto = activeModalTask.task_type === 'PHOTO_UPLOAD' || activeModalTask.task_type === 'PHOTO_PROOF'
          const isGame = activeModalTask.task_type === 'MINI_GAME' || Boolean(cfg.game)
          const isText = activeModalTask.task_type === 'TEXT_RESPONSE' || (!isDoc && !isPhoto && Boolean(cfg.minimum_characters || cfg.prompt))

          if (isDoc) {
            return (
              <DocUploadModal
                task={activeModalTask}
                onClose={() => setActiveModalTask(null)}
                onSuccess={handleModalSuccess}
              />
            )
          }

          if (isPhoto) {
            return (
              <CameraCaptureModal
                task={activeModalTask}
                onClose={() => setActiveModalTask(null)}
                onSuccess={handleModalSuccess}
              />
            )
          }

          if (isText) {
            return (
              <TextResponseModal
                task={activeModalTask}
                onClose={() => setActiveModalTask(null)}
                onSuccess={handleModalSuccess}
              />
            )
          }

          if (isGame) {
            return (
              <MiniGameModal
                task={activeModalTask}
                onClose={() => setActiveModalTask(null)}
                onSuccess={handleModalSuccess}
              />
            )
          }

          // Default / Video / Quiz / Composite Video+Quiz
          return (
            <VideoQuizModal
              task={activeModalTask}
              onClose={() => setActiveModalTask(null)}
              onSuccess={handleModalSuccess}
            />
          )
        })()}
      </AnimatePresence>
    </div>
  )
}
