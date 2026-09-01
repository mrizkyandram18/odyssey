import React from 'react'
import { Video, HelpCircle, Camera, FileText, AlignLeft, Gamepad2 } from 'lucide-react'
import type { TaskType } from '../../../../shared/types'

interface TaskTypeBadgeProps {
  type: TaskType
}

export const TaskTypeBadge: React.FC<TaskTypeBadgeProps> = ({ type }) => {
  switch (type) {
    case 'VIDEO':
      return (
        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[11px] font-bold bg-rose-500/10 text-rose-700 dark:text-rose-300 border border-rose-500/20">
          <Video className="w-3 h-3" /> Video
        </span>
      )
    case 'QUIZ':
      return (
        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[11px] font-bold bg-amber-500/10 text-amber-700 dark:text-amber-300 border border-amber-500/20">
          <HelpCircle className="w-3 h-3" /> Kuis
        </span>
      )
    case 'PHOTO_UPLOAD':
      return (
        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[11px] font-bold bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 border border-emerald-500/20">
          <Camera className="w-3 h-3" /> Foto
        </span>
      )
    case 'DOCUMENT_UPLOAD':
      return (
        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[11px] font-bold bg-indigo-500/10 text-indigo-700 dark:text-indigo-300 border border-indigo-500/20">
          <FileText className="w-3 h-3" /> Dokumen
        </span>
      )
    case 'TEXT_RESPONSE':
      return (
        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[11px] font-bold bg-cyan-500/10 text-cyan-700 dark:text-cyan-300 border border-cyan-500/20">
          <AlignLeft className="w-3 h-3" /> Teks Esai
        </span>
      )
    case 'MINI_GAME':
      return (
        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[11px] font-bold bg-purple-500/10 text-purple-700 dark:text-purple-300 border border-purple-500/20">
          <Gamepad2 className="w-3 h-3" /> Mini Game
        </span>
      )
    default:
      return (
        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[11px] font-bold bg-surface-elevated text-text-secondary border border-border-subtle">
          {type}
        </span>
      )
  }
}
