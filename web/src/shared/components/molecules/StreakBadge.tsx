export interface StreakBadgeProps {
  days: number
  maxDays?: number
}

export function StreakBadge({ days, maxDays = 7 }: StreakBadgeProps) {
  const active = days >= maxDays
  return (
    <div className="flex items-center gap-1 rounded-full border border-accent/20 bg-accent/10 px-2 py-1">
      <span className="text-xs font-medium">🔥 {days}d</span>
      {active && (
        <span className="text-xs text-accent">★</span>
      )}
    </div>
  )
}
