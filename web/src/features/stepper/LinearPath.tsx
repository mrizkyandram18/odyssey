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
    <div className="w-full flex flex-col gap-4">
      {/* 1. Compact Status Header */}
      <header className="bg-surface rounded-2xl border border-border-subtle p-4 shadow-sm">
        <div className="flex items-center justify-between gap-2">
          <div className="flex items-center gap-2">
            <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-accent-danger/10 border border-accent-danger/20 text-accent-danger font-bold text-xs">
              <Flame className="w-3.5 h-3.5" />
              <span>{userStreak} Hari</span>
            </span>
            <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-accent-gold/15 border border-accent-gold/20 text-accent-gold font-bold text-xs">
              <Coins className="w-3.5 h-3.5" />
              <span>{userCoins.toLocaleString('id-ID')}</span>
            </span>
          </div>
          <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-accent-magic/10 border border-accent-magic/20 text-accent-magic font-bold text-xs">
            <Trophy className="w-3.5 h-3.5" />
            <span>Lv. {userLevel}</span>
          </span>
        </div>
        <div className="mt-3 flex items-center gap-2">
          <div className="flex-1 h-2 bg-surface border border-border-subtle rounded-full overflow-hidden">
            <motion.div
              initial={{ width: 0 }}
              animate={{ width: `${xpProgressPercent}%` }}
              transition={{ duration: 0.6, ease: 'easeOut' }}
              className="h-full bg-accent-magic rounded-full"
            />
          </div>
          <span className="text-[11px] text-text-secondary font-semibold shrink-0">
            {xpInCurrentLevel} / {xpRequiredForLevel} XP
          </span>
        </div>
      </header>

      {/* 2. Today's Banner */}
      <div className="flex items-start justify-between gap-3">
        <div>
          <h2 className="text-lg font-bold text-text-primary flex items-center gap-2">
            <span>Tugas Harian</span>
            <Sparkles className="w-4 h-4 text-accent-gold" />
          </h2>
          <p className="text-xs text-text-secondary mt-0.5 flex items-center gap-1">
            <Calendar className="w-3.5 h-3.5" />
            <span>{new Date().toLocaleDateString('id-ID', { dateStyle: 'long' })}</span>
          </p>
        </div>
        <button
          onClick={loadTasks}
          disabled={loading}
          aria-label="Muat ulang tugas"
          className="p-2 rounded-xl bg-surface border border-border-subtle text-text-secondary hover:text-text-primary hover:bg-surface-elevated transition-colors shrink-0"
        >
          <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
        </button>
      </div>

      {/* 3. Main Linear Task Stepper */}
      <div className="flex-1 flex flex-col items-center relative">
        {loading && tasks.length === 0 ? (
          <div className="py-16 flex flex-col items-center gap-3">
            <div className="w-8 h-8 rounded-full border-2 border-accent-magic border-t-transparent animate-spin" />
            <p className="text-xs text-text-secondary font-semibold tracking-wide uppercase">
              Memuat tugas...
            </p>
          </div>
        ) : error ? (
          <div className="w-full p-5 text-center bg-surface border border-status-error/20 rounded-2xl">
            <p className="text-sm text-status-error font-semibold">{error}</p>
            <button
              onClick={loadTasks}
              className="mt-3 px-4 py-2 rounded-xl bg-surface-elevated border border-border-subtle text-xs font-bold text-text-primary hover:bg-surface transition-colors"
            >
              Coba Lagi
            </button>
          </div>
        ) : tasks.length === 0 ? (
          <div className="w-full p-8 text-center bg-surface border border-border-subtle rounded-2xl space-y-3">
            <div className="w-12 h-12 mx-auto rounded-full bg-accent-magic/10 flex items-center justify-center text-xl">
              🌟
            </div>
            <h4 className="font-bold text-text-primary text-sm">
              Belum Ada Tugas Hari Ini
            </h4>
            <p className="text-xs text-text-secondary leading-relaxed max-w-[28ch] mx-auto">
              Admin belum menambahkan jadwal tugas untuk hari ini. Silakan hubungi pengelola keluarga.
            </p>
          </div>
        ) : (
          <div className="w-full flex flex-col items-center relative py-2">
            {/* Subtle vertical line */}
            <div className="absolute top-6 bottom-16 left-1/2 -translate-x-1/2 w-px bg-border-subtle" aria-hidden />

            {/* Step Nodes Sequence */}
            {tasks.map((task, idx) => (
              <StepNode
                key={task.id}
                task={task}
                index={idx}
                onClick={handleTaskClick}
              />
            ))}

            {/* End of Path */}
            <div className="mt-6 flex flex-col items-center text-center gap-2">
              <div className="w-12 h-12 rounded-2xl bg-accent-gold/15 border border-accent-gold/20 text-accent-gold flex items-center justify-center">
                <Trophy className="w-6 h-6" />
              </div>
              <p className="text-xs font-bold text-text-secondary">
                Selesaikan semua langkah
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
