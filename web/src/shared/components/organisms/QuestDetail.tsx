import { useState } from 'react'
import type { Challenge, Quest, CompleteChallengeResult } from '../../types'
import { Badge } from '../atoms/Badge'
import { Button } from '../atoms/Button'

export interface QuestDetailProps {
  quest: Quest
  challenges: Challenge[]
  onStartQuest?: () => Promise<void>
  onCompleteChallenge?: (challengeId: number) => Promise<CompleteChallengeResult | null>
}

export function QuestDetail({
  quest,
  challenges,
  onStartQuest,
  onCompleteChallenge,
}: QuestDetailProps) {
  const [starting, setStarting] = useState(false)
  const [completingId, setCompletingId] = useState<number | null>(null)
  const [lastResult, setLastResult] = useState<CompleteChallengeResult | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)

  const statusVariant = {
    PENDING: 'default' as const,
    ACTIVE: 'primary' as const,
    DONE: 'success' as const,
  }[quest.status]

  const handleStart = async () => {
    if (!onStartQuest) return
    setStarting(true)
    setActionError(null)
    try {
      await onStartQuest()
    } catch (e) {
      setActionError(e instanceof Error ? e.message : 'Failed to start quest')
    } finally {
      setStarting(false)
    }
  }

  const handleComplete = async (challengeId: number) => {
    if (!onCompleteChallenge) return
    setCompletingId(challengeId)
    setActionError(null)
    try {
      const res = await onCompleteChallenge(challengeId)
      if (res) {
        setLastResult(res)
      }
    } catch (e) {
      setActionError(e instanceof Error ? e.message : 'Failed to complete challenge')
    } finally {
      setCompletingId(null)
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold">{quest.title}</h1>
          <p className="text-xs text-muted-foreground capitalize">
            {quest.template_slug.replace(/-/g, ' ')}
          </p>
        </div>
        <Badge variant={statusVariant}>{quest.status}</Badge>
      </div>

      {actionError && (
        <p className="text-xs text-error bg-error/10 p-2 rounded">{actionError}</p>
      )}

      {lastResult && (
        <div className="rounded-lg border border-accent bg-accent/10 p-4 text-center space-y-1">
          <p className="text-sm font-bold text-accent">
            + {lastResult.xp} XP Earned!
          </p>
          {lastResult.level_up && (
            <p className="text-xs font-semibold text-primary">
              🎉 Level Up! You are now Explorer Level {lastResult.new_level}!
            </p>
          )}
          {lastResult.quest_completed && (
            <p className="text-xs font-semibold text-success">
              🏆 Quest Completed! Check your home screen for chest rewards!
            </p>
          )}
        </div>
      )}

      {quest.status === 'PENDING' && onStartQuest && (
        <Button isLoading={starting} onClick={handleStart} className="w-full">
          Start Quest
        </Button>
      )}

      <div className="flex flex-col gap-3">
        <h2 className="text-sm font-medium text-muted-foreground">Challenges</h2>
        {challenges.length === 0 ? (
          <p className="text-xs text-muted-foreground">No challenges in this quest.</p>
        ) : (
          challenges.map((c) => {
            const isDone = c.status === 'DONE'
            const isCompleting = completingId === c.id

            return (
              <div
                key={c.id}
                className={`flex flex-col gap-2 rounded-lg border border-border bg-surface p-4 transition-all ${
                  isDone ? 'opacity-70' : ''
                }`}
              >
                <div className="flex items-start justify-between gap-2">
                  <p className="text-sm font-medium">{c.description}</p>
                  <Badge variant={isDone ? 'success' : 'default'} size="sm">
                    {c.status}
                  </Badge>
                </div>

                {isDone && (
                  <p className="text-xs text-muted-foreground">
                    Completed {c.completed_by ? `by ${c.completed_by}` : ''}
                  </p>
                )}

                {!isDone && quest.status === 'ACTIVE' && onCompleteChallenge && (
                  <Button
                    size="sm"
                    variant="ghost"
                    isLoading={isCompleting}
                    onClick={() => handleComplete(c.id)}
                    className="self-end mt-1 border border-border"
                  >
                    Complete Challenge
                  </Button>
                )}
              </div>
            )
          })
        )}
      </div>
    </div>
  )
}
