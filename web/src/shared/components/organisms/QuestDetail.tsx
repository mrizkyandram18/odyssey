import type { ReactNode } from 'react'
import type { Challenge, Quest } from '../../types'

export interface QuestDetailProps {
  quest: Quest
  challenges: Challenge[]
  children?: ReactNode
}

export function QuestDetail({ quest, challenges, children }: QuestDetailProps) {
  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-xl font-semibold">{quest.title}</h1>
      {children}
      <div className="flex flex-col gap-2">
        {challenges.map((c) => (
          <div key={c.id} className="rounded-md border border-border bg-surface p-2">
            <span className="text-sm">{c.description}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
