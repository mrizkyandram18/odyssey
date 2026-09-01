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
        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-semibold bg-rose-500/20 text-rose-400 border border-rose-500/30">
          <Video className="w-3 h-3" /> Video
        </span>
      )
    case 'QUIZ':
      return (
        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-semibold bg-amber-500/20 text-amber-400 border border-amber-500/30">
          <HelpCircle className="w-3 h-3" /> Kuis
        </span>
      )
    case 'PHOTO_UPLOAD':
      return (
        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-semibold bg-emerald-500/20 text-emerald-400 border border-emerald-500/30">
          <Camera className="w-3 h-3" /> Foto
        </span>
      )
    case 'DOCUMENT_UPLOAD':
      return (
        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-semibold bg-indigo-500/20 text-indigo-400 border border-indigo-500/30">
          <FileText className="w-3 h-3" /> Dokumen
        </span>
      )
    case 'TEXT_RESPONSE':
      return (
        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-semibold bg-cyan-500/20 text-cyan-400 border border-cyan-500/30">
          <AlignLeft className="w-3 h-3" /> Teks Esai
        </span>
      )
    case 'MINI_GAME':
      return (
        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-semibold bg-purple-500/20 text-purple-400 border border-purple-500/30">
          <Gamepad2 className="w-3 h-3" /> Mini Game
        </span>
      )
    default:
      return (
        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-semibold bg-slate-500/20 text-slate-400 border border-slate-500/30">
          {type}
        </span>
      )
  }
}
