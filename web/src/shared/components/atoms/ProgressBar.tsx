interface ProgressBarProps {
  progress: number // 0 to 100
  colorClass?: string
  className?: string
  showLabel?: boolean
}

export function ProgressBar({ 
  progress, 
  colorClass = 'bg-accent-nature', 
  className = '',
  showLabel = false
}: ProgressBarProps) {
  const safeProgress = Math.min(100, Math.max(0, progress))
  
  return (
    <div className={`w-full ${className}`}>
      <div className="h-2 w-full rounded-full bg-surface-elevated border border-border-subtle overflow-hidden">
        <div
          className={`h-full rounded-full ${colorClass} transition-all duration-700 ease-out`}
          style={{ width: `${safeProgress}%` }}
        />
      </div>
      {showLabel && (
        <div className="mt-1 flex justify-end">
          <span className="text-xs font-medium text-text-secondary">{Math.round(safeProgress)}%</span>
        </div>
      )}
    </div>
  )
}
