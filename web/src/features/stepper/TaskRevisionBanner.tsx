import React from 'react'
import { RotateCcw } from 'lucide-react'

interface TaskRevisionBannerProps {
  adminNotes?: string | null
  className?: string
}

export const TaskRevisionBanner: React.FC<TaskRevisionBannerProps> = ({
  adminNotes,
  className = '',
}) => {
  const trimmedNotes = (adminNotes || '').trim()

  return (
    <div
      data-testid="task-revision-banner"
      className={`rounded-2xl bg-amber-500/10 border border-amber-500/30 p-3.5 space-y-2 text-left animate-fadeIn ${className}`}
    >
      <div className="flex items-center gap-2 text-amber-700 dark:text-amber-400 font-extrabold text-xs">
        <RotateCcw className="w-4 h-4 shrink-0" />
        <span>Perlu Diperbaiki</span>
      </div>

      {Boolean(trimmedNotes) && (
        <div className="space-y-1">
          <p className="text-[11px] font-bold text-text-secondary">Catatan dari admin:</p>
          <div className="p-3 rounded-xl bg-surface/80 dark:bg-surface-elevated border border-border-subtle text-xs text-text-primary italic leading-relaxed whitespace-pre-wrap break-words font-medium">
            &ldquo;{trimmedNotes}&rdquo;
          </div>
        </div>
      )}

      <p className="text-[11px] text-text-secondary leading-relaxed">
        Silakan perbaiki bukti lalu kirim ulang agar dapat disetujui admin.
      </p>
    </div>
  )
}
