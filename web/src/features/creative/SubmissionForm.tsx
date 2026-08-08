import { useState } from 'react'
import { apiClient } from '../../shared/lib/api'

export interface SubmissionFormProps {
  questId: number
  challengeId: number
  onComplete: () => void
  onSkip: () => void
}

export function SubmissionForm({ questId, challengeId, onComplete, onSkip }: SubmissionFormProps) {
  const [content, setContent] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!content.trim()) return

    setSubmitting(true)
    setError(null)
    try {
      await apiClient.post('/api/creative', {
        quest_id: questId,
        challenge_id: challengeId,
        kind: 'STORY',
        content,
      })
      onComplete()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save memory')
      setSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="flex w-full flex-col gap-4 rounded-xl border border-border bg-surface p-6 shadow-md">
      <div className="text-center">
        <h2 className="text-xl font-bold">Create a Memory</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          You completed a quest! Leave a short journal entry about your adventure today.
        </p>
      </div>

      <div className="flex flex-col gap-2">
        <textarea
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder="What happened today? (e.g., Dad found the treasure!)"
          className="min-h-[120px] w-full resize-none rounded-lg border border-border bg-background p-4 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
          required
        />
      </div>

      {error && <p className="text-center text-xs text-red-500">{error}</p>}

      <div className="flex flex-col gap-2 pt-2">
        <button
          type="submit"
          disabled={submitting || !content.trim()}
          className="w-full rounded-lg bg-primary py-3 text-sm font-bold text-primary-foreground transition-all hover:brightness-110 disabled:opacity-50"
        >
          {submitting ? 'Saving...' : 'Save Memory'}
        </button>
        <button
          type="button"
          onClick={onSkip}
          disabled={submitting}
          className="w-full rounded-lg py-3 text-sm font-semibold text-muted-foreground transition-all hover:bg-muted"
        >
          Skip for now
        </button>
      </div>
    </form>
  )
}
