import type { ReactNode } from 'react'
import { Button } from '../atoms/Button'

export interface dailyMissionBannerProps {
  remaining: number
  onTakeTurn: () => void
  loading?: boolean
}

export function dailyMissionBanner({ remaining, onTakeTurn, loading }: dailyMissionBannerProps) {
  const canTake = remaining > 0 && !loading

  let label: ReactNode
  if (remaining <= 0) {
    label = 'No turns remaining today'
  } else {
    label = `Take your daily turn`
  }

  return (
    <div className="flex items-center justify-between rounded-lg border border-border bg-surface p-3">
      <span className="text-sm font-medium">{label}</span>
      {canTake && (
        <Button size="sm" onClick={onTakeTurn} isLoading={loading}>
          Go
        </Button>
      )}
    </div>
  )
}
