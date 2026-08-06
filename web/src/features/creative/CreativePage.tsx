import { useState } from 'react'
import { apiClient } from '../../shared/lib/api'
import type { CreativeSubmission, SubmissionKind } from '../../shared/types'
import { Badge } from '../../shared/components/atoms/Badge'

const KIND_LABELS: Record<SubmissionKind, string> = {
  STORY: 'Story',
  COMIC: 'Comic',
  PHOTO: 'Photo',
  VIDEO: 'Video',
}

const STATUS_VARIANT: Record<string, 'default' | 'primary' | 'secondary' | 'success' | 'warning' | 'error'> = {
  PENDING: 'warning',
  APPROVED: 'success',
  REJECTED: 'error',
}

export function CreativePage() {
  const [submissions, setSubmissions] = useState<CreativeSubmission[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [questId, setQuestId] = useState('')
  const [challengeId, setChallengeId] = useState('')
  const [kind, setKind] = useState<SubmissionKind>('STORY')
  const [content, setContent] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const loadSubmissions = async () => {
    setLoading(true)
    setError(null)
    try {
      const qid = questId || '1'
      const data = await apiClient.get<CreativeSubmission[]>(`/api/creative?quest_id=${qid}`)
      setSubmissions(data)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'failed to load submissions')
    } finally {
      setLoading(false)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      await apiClient.post('/api/creative', {
        quest_id: Number(questId) || 1,
        challenge_id: Number(challengeId) || 1,
        kind,
        content,
      })
      setContent('')
      await loadSubmissions()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'failed to submit')
    } finally {
      setSubmitting(false)
    }
  }

  if (loading && submissions.length === 0) {
    return <p className="p-4 text-sm text-muted-foreground">Loading...</p>
  }

  return (
    <div className="flex flex-col gap-4 p-4 pb-safe">
      <h1 className="text-xl font-semibold">Creative Space</h1>
      <p className="text-sm text-muted-foreground">
        Submit your creative work for quest challenges.
      </p>

      <form onSubmit={handleSubmit} className="flex flex-col gap-3 rounded-lg border border-border bg-surface p-4">
        <h2 className="text-sm font-medium">New Submission</h2>
        <div className="flex flex-col gap-2">
          <label className="text-xs text-muted-foreground">Quest ID</label>
          <input
            type="number"
            value={questId}
            onChange={(e) => setQuestId(e.target.value)}
            className="rounded-md border border-border bg-background p-2 text-sm"
            placeholder="1"
            min={1}
          />
        </div>
        <div className="flex flex-col gap-2">
          <label className="text-xs text-muted-foreground">Challenge ID</label>
          <input
            type="number"
            value={challengeId}
            onChange={(e) => setChallengeId(e.target.value)}
            className="rounded-md border border-border bg-background p-2 text-sm"
            placeholder="1"
            min={1}
          />
        </div>
        <div className="flex flex-col gap-2">
          <label className="text-xs text-muted-foreground">Kind</label>
          <select
            value={kind}
            onChange={(e) => setKind(e.target.value as SubmissionKind)}
            className="rounded-md border border-border bg-background p-2 text-sm"
          >
            <option value="STORY">Story</option>
            <option value="COMIC">Comic</option>
            <option value="PHOTO">Photo</option>
            <option value="VIDEO">Video</option>
          </select>
        </div>
        <div className="flex flex-col gap-2">
          <label className="text-xs text-muted-foreground">Content</label>
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            className="rounded-md border border-border bg-background p-2 text-sm"
            rows={4}
            placeholder="Describe your creative work..."
          />
        </div>
        <button
          type="submit"
          disabled={submitting}
          className="rounded-md bg-primary p-2 text-sm font-semibold text-black disabled:opacity-50"
        >
          {submitting ? 'Submitting...' : 'Submit'}
        </button>
      </form>

      {error && <p className="text-xs text-red-500">{error}</p>}

      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-medium text-muted-foreground">Submissions</h2>
          <button
            onClick={loadSubmissions}
            className="text-xs text-primary"
          >
            Refresh
          </button>
        </div>
        {loading ? (
          <p className="text-sm text-muted-foreground">Loading...</p>
        ) : submissions.length === 0 ? (
          <p className="text-sm text-muted-foreground">No submissions yet.</p>
        ) : (
          submissions.map((sub) => (
            <div key={sub.id} className="rounded-lg border border-border bg-surface p-3">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium">{KIND_LABELS[sub.kind]}</span>
                <Badge variant={STATUS_VARIANT[sub.status] || 'default'} size="sm">
                  {sub.status}
                </Badge>
              </div>
              <p className="mt-1 text-sm text-muted-foreground line-clamp-2">
                {sub.content}
              </p>
              <p className="mt-1 text-xs text-muted-foreground">
                Quest #{sub.quest_id} · Challenge #{sub.challenge_id}
              </p>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
