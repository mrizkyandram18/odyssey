import React from 'react'
import { motion } from 'framer-motion'
import { Check, Lock, Clock, AlertTriangle, Play, FileText, Camera, Sparkles, HelpCircle, PenLine, Gamepad2 } from 'lucide-react'
import type { TaskView } from '../../shared/types'

interface StepNodeProps {
  task: TaskView
  index: number
  onClick: (task: TaskView) => void
}

export const StepNode: React.FC<StepNodeProps> = ({ task, onClick }) => {
  const getTaskIcon = () => {
    switch (task.task_type) {
      case 'VIDEO':
      case 'VIDEO_QUIZ':
      case 'YOUTUBE_VIDEO':
        return <Play className="w-5 h-5 fill-current" />
      case 'QUIZ':
        return <HelpCircle className="w-5 h-5" />
      case 'DOCUMENT_UPLOAD':
        return <FileText className="w-5 h-5" />
      case 'PHOTO_UPLOAD':
      case 'PHOTO_PROOF':
        return <Camera className="w-5 h-5" />
      case 'TEXT_RESPONSE':
        return <PenLine className="w-5 h-5" />
      case 'MINI_GAME':
        return <Gamepad2 className="w-5 h-5" />
      default:
        return <Sparkles className="w-5 h-5" />
    }
  }

  // Determine appearance based on status
  const isApproved = task.status === 'APPROVED'
  const isPending = task.status === 'PENDING'
  const isUnlocked = task.status === 'UNLOCKED'
  const isLocked = task.status === 'LOCKED' || task.is_locked
  const isRejected = task.status === 'REJECTED'

  const isActive = isUnlocked || isPending

  return (
    <div className="relative flex flex-col items-center py-3 w-full max-w-[200px]">
      {/* Node Button */}
      <motion.button
        whileHover={!isLocked ? { scale: 1.04 } : {}}
        whileTap={!isLocked ? { scale: 0.97 } : {}}
        onClick={() => onClick(task)}
        disabled={isLocked}
        aria-label={`${task.title} - ${task.status}`}
        className={`relative w-[68px] h-[68px] rounded-full flex items-center justify-center transition-all select-none shrink-0 ${
          isApproved
            ? 'bg-status-success text-white shadow-sm border-2 border-status-success'
            : isPending
            ? 'bg-accent-gold text-white shadow-sm border-2 border-accent-gold'
            : isUnlocked
            ? 'bg-accent-magic text-white shadow-md border-2 border-white ring-2 ring-accent-magic/30'
            : isRejected
            ? 'bg-status-error text-white shadow-sm border-2 border-status-error'
            : 'bg-surface text-text-secondary/50 border-2 border-border-subtle shadow-none'
        } ${isActive && !isLocked ? 'cursor-pointer' : isLocked ? 'cursor-not-allowed opacity-60' : ''}`}
      >
        {/* Main Icon */}
        {isApproved ? (
          <Check className="w-6 h-6 stroke-[2.5]" />
        ) : isPending ? (
          <Clock className="w-6 h-6" />
        ) : isLocked ? (
          <Lock className="w-5 h-5" />
        ) : isRejected ? (
          <AlertTriangle className="w-5 h-5" />
        ) : (
          getTaskIcon()
        )}

        {/* Step Badge */}
        <span
          className={`absolute -top-1 -right-1 w-6 h-6 rounded-full text-[11px] font-bold flex items-center justify-center shadow-sm border-2 border-white ${
            isApproved
              ? 'bg-status-success text-white'
              : isPending
              ? 'bg-accent-gold text-white'
              : isUnlocked
              ? 'bg-white text-accent-magic border-accent-magic/20'
              : 'bg-surface-elevated text-text-secondary'
          }`}
        >
          {task.step_order}
        </span>
      </motion.button>

      {/* Title & Reward Label Below Node */}
      <div className="mt-2 text-center max-w-[160px]">
        <p className={`text-xs font-bold line-clamp-2 leading-tight ${
          isLocked ? 'text-text-secondary/60' : 'text-text-primary'
        }`}>
          {task.title}
        </p>
        <span className={`inline-flex items-center mt-1 text-[11px] font-semibold px-2 py-0.5 rounded-full border ${
          isApproved ? 'bg-status-success/10 text-status-success border-status-success/20' :
          isPending ? 'bg-accent-gold/10 text-accent-gold border-accent-gold/20' :
          isUnlocked ? 'bg-accent-magic/10 text-accent-magic border-accent-magic/20' :
          isRejected ? 'bg-status-error/10 text-status-error border-status-error/20' :
          'bg-surface border-border-subtle text-text-secondary/60'
        }`}>
          {isApproved ? `Selesai +${task.coins_earned || task.reward_coins}` :
           isPending ? 'Menunggu review' :
           isUnlocked ? `Mulai +${task.reward_coins}` :
           isRejected ? 'Perlu revisi' : 'Terkunci'}
        </span>
      </div>
    </div>
  )
}
