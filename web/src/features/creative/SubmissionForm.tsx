import { useState } from 'react'
import { apiClient } from '../../shared/lib/api'
import { CreativeCanvas } from '../quest/CreativeCanvas'

export interface SubmissionFormProps {
  questId: number
  challengeId: number
  onComplete: () => void
  onSkip: () => void
}

export function SubmissionForm({ questId, challengeId, onComplete, onSkip }: SubmissionFormProps) {
  const [mode, setMode] = useState<'STORY' | 'DRAWING'>('STORY')
  const [content, setContent] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmitText = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!content.trim()) return
    await submitCreative('STORY', content)
  }

  const handleCanvasSubmit = async (svg: string) => {
    await submitCreative('DRAWING', svg)
  }

  const submitCreative = async (kind: 'STORY' | 'DRAWING', payload: string) => {
    setSubmitting(true)
    setError(null)
    try {
      await apiClient.post('/api/creative', {
        quest_id: questId,
        challenge_id: challengeId,
        kind,
        content: payload,
      })
      onComplete()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save memory')
      setSubmitting(false)
    }
  }

  return (
    <div className="flex w-full flex-col gap-4 rounded-xl border border-border bg-surface p-6 shadow-md">
      <div className="text-center mb-2">
        <h2 className="text-xl font-bold">Create a Memory</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          You completed a quest! Capture the moment.
        </p>
      </div>

      <div className="flex justify-center gap-4 mb-2">
        <button
          type="button"
          onClick={() => setMode('STORY')}
          className={`px-4 py-2 rounded-full text-sm font-bold ${
            mode === 'STORY' ? 'bg-primary text-primary-foreground' : 'bg-background hover:bg-surface-hover'
          }`}
        >
          Write Story
        </button>
        <button
          type="button"
          onClick={() => setMode('DRAWING')}
          className={`px-4 py-2 rounded-full text-sm font-bold ${
            mode === 'DRAWING' ? 'bg-primary text-primary-foreground' : 'bg-background hover:bg-surface-hover'
          }`}
        >
          Draw Canvas
        </button>
      </div>

      {error && <p className="text-center text-xs text-red-500">{error}</p>}

      {mode === 'STORY' ? (
        <form onSubmit={handleSubmitText} className="flex flex-col gap-4">
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder="What happened today? (e.g., Dad found the treasure!)"
            className="min-h-[120px] w-full resize-none rounded-lg border border-border bg-background p-4 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
            required
            disabled={submitting}
          />
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
      ) : (
        <CreativeCanvas
          onSubmit={handleCanvasSubmit}
          onCancel={onSkip}
          isSubmitting={submitting}
        />
      )}
    </div>
  )
}
