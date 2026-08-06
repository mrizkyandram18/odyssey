import { useState, useEffect } from 'react'
import type { OpenResult } from '../../../shared/types'

export interface ChestOpeningAnimationProps {
  result: OpenResult
  onComplete?: () => void
}

export function ChestOpeningAnimation({ result, onComplete }: ChestOpeningAnimationProps) {
  const [phase, setPhase] = useState(0)

  useEffect(() => {
    const timers = [
      setTimeout(() => setPhase(1), 300),
      setTimeout(() => setPhase(2), 800),
      setTimeout(() => setPhase(3), 1500),
      setTimeout(() => onComplete?.(), 2500),
    ]
    return () => timers.forEach(clearTimeout)
  }, [onComplete])

  return (
    <div className="flex flex-col items-center justify-center gap-4 p-4">
      {phase === 0 && (
        <div className="flex flex-col items-center gap-2">
          <span className="text-6xl animate-bounce">{result.chest.icon}</span>
          <p className="text-sm text-muted-foreground">Opening {result.chest.name}...</p>
        </div>
      )}

      {phase === 1 && (
        <div className="flex flex-col items-center gap-2">
          <div className="text-4xl animate-pulse">✨</div>
          <p className="text-sm text-muted-foreground">Revealing rewards...</p>
        </div>
      )}

      {phase === 2 && (
        <div className="flex flex-col items-center gap-2">
          <div className="text-4xl animate-spin">🔄</div>
          <p className="text-sm text-muted-foreground">Almost there...</p>
        </div>
      )}

      {phase === 3 && (
        <div className="flex flex-col items-center gap-4 w-full">
          <div className="text-4xl animate-bounce">🎉</div>
          <p className="text-lg font-semibold">Rewards revealed!</p>
          <div className="flex flex-col gap-2 w-full">
            {result.rewards.map((reward, idx) => (
              <div
                key={idx}
                className="flex items-center gap-3 rounded-lg border border-border bg-surface p-3 animate-[fadeIn_0.5s_ease-in]"
                style={{ animationDelay: `${idx * 200}ms` }}
              >
                <span className="text-2xl">💎</span>
                <div className="flex-1">
                  <p className="font-medium">{reward.name}</p>
                  <p className="text-xs text-muted-foreground">{reward.rarity}</p>
                </div>
                {reward.is_new && <span className="text-xs text-accent">NEW</span>}
              </div>
            ))}
          </div>
          <div className="flex gap-4 text-sm">
            <span className="text-muted-foreground">New: <span className="text-accent font-medium">{result.new_count}</span></span>
            <span className="text-muted-foreground">Duplicates: <span className="font-medium">{result.duplicate_count}</span></span>
          </div>
        </div>
      )}
    </div>
  )
}
