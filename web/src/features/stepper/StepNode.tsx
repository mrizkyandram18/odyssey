import React from 'react'
import { motion } from 'framer-motion'
import { Check, Lock, Clock, AlertTriangle, Play, FileText, Camera, Sparkles, HelpCircle, PenLine, Gamepad2 } from 'lucide-react'
import type { TaskView } from '../../shared/types'

interface StepNodeProps {
  task: TaskView
  index: number
  onClick: (task: TaskView) => void
}

export const StepNode: React.FC<StepNodeProps> = ({ task, index, onClick }) => {
  const getTaskIcon = () => {
    switch (task.task_type) {
      case 'VIDEO':
      case 'VIDEO_QUIZ':
      case 'YOUTUBE_VIDEO':
        return <Play className="w-6 h-6 fill-current" />
      case 'QUIZ':
        return <HelpCircle className="w-6 h-6" />
      case 'DOCUMENT_UPLOAD':
        return <FileText className="w-6 h-6" />
      case 'PHOTO_UPLOAD':
      case 'PHOTO_PROOF':
        return <Camera className="w-6 h-6" />
      case 'TEXT_RESPONSE':
        return <PenLine className="w-6 h-6" />
      case 'MINI_GAME':
        return <Gamepad2 className="w-6 h-6" />
      default:
        return <Sparkles className="w-6 h-6" />
    }
  }

  // Determine appearance based on status
  const isApproved = task.status === 'APPROVED'
  const isPending = task.status === 'PENDING'
  const isUnlocked = task.status === 'UNLOCKED'
  const isLocked = task.status === 'LOCKED' || task.is_locked
  const isRejected = task.status === 'REJECTED'

  // Dynamic alternating horizontal offset for winding snake path
  const xOffset = index % 2 === 0 ? 'translate-x-0' : index % 4 === 1 ? 'translate-x-8' : '-translate-x-8'

  return (
    <div className={`relative flex flex-col items-center my-3 transition-transform ${xOffset}`}>
      {/* Node Button */}
      <motion.button
        whileHover={!isLocked ? { scale: 1.06 } : {}}
        whileTap={!isLocked ? { scale: 0.95 } : {}}
        onClick={() => onClick(task)}
        disabled={isLocked}
        className={`relative w-20 h-20 rounded-full flex items-center justify-center transition-all shadow-lg select-none ${
          isApproved
            ? 'bg-status-success text-white shadow-status-success/30 border-4 border-status-success/40'
            : isPending
            ? 'bg-accent-gold text-white shadow-accent-gold/40 border-4 border-accent-gold/40 animate-pulse'
            : isUnlocked
            ? 'bg-accent-magic text-white shadow-accent-magic/50 border-4 border-white dark:border-surface-elevated ring-4 ring-accent-magic/40 animate-bounce'
            : isRejected
            ? 'bg-status-error text-white shadow-status-error/30 border-4 border-status-error/40'
            : 'bg-surface-base text-text-secondary/40 border-4 border-border-subtle opacity-70 cursor-not-allowed shadow-none'
        }`}
      >
        {/* Main Icon */}
        {isApproved ? (
          <Check className="w-8 h-8 stroke-[3]" />
        ) : isPending ? (
          <Clock className="w-8 h-8 animate-spin" />
        ) : isLocked ? (
          <Lock className="w-7 h-7" />
        ) : isRejected ? (
          <AlertTriangle className="w-7 h-7" />
        ) : (
          getTaskIcon()
        )}

        {/* Step Badge */}
        <span
          className={`absolute -top-1 -right-1 w-6 h-6 rounded-full text-xs font-bold flex items-center justify-center shadow-md ${
            isApproved
              ? 'bg-white text-status-success'
              : isPending
              ? 'bg-white text-accent-gold'
              : isUnlocked
              ? 'bg-accent-gold text-text-primary'
              : 'bg-surface-elevated text-text-secondary'
          }`}
        >
          {task.step_order}
        </span>
      </motion.button>

      {/* Title & Reward Label Below Node */}
      <div className="mt-2.5 text-center max-w-[140px]">
        <p className={`text-xs font-heading font-bold line-clamp-1 ${
          isLocked ? 'text-text-secondary/60' : 'text-text-primary'
        }`}>
          {task.title}
        </p>

        {isApproved ? (
          <span className="inline-block mt-0.5 text-[10px] text-status-success font-bold bg-status-success/15 px-2 py-0.5 rounded-full">
            ✅ Selesai (+{task.coins_earned || task.reward_coins}🪙)
          </span>
        ) : isPending ? (
          <span className="inline-block mt-0.5 text-[10px] text-accent-gold font-bold bg-accent-gold/15 px-2 py-0.5 rounded-full">
            ⏳ Menunggu Review
          </span>
        ) : isUnlocked ? (
          <span className="inline-block mt-0.5 text-[10px] text-accent-magic font-bold bg-accent-magic/15 px-2 py-0.5 rounded-full animate-pulse">
            ⚡ Mulai (+{task.reward_coins}🪙)
          </span>
        ) : isRejected ? (
          <span className="inline-block mt-0.5 text-[10px] text-status-error font-bold bg-status-error/15 px-2 py-0.5 rounded-full">
            ⚠️ Perlu Revisi
          </span>
        ) : (
          <span className="inline-block mt-0.5 text-[10px] text-text-secondary/60 font-medium">
            🔒 Terkunci
          </span>
        )}
      </div>
    </div>
  )
}
