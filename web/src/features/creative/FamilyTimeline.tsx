import { useEffect, useState } from 'react'
import { apiClient } from '../../shared/lib/api'
import type { CreativeSubmission } from '../../shared/types'
import { CreativeCard } from '../../shared/components/molecules/CreativeCard'
import { SharedTextBoard } from './SharedTextBoard'

export function FamilyTimeline() {
  const [submissions, setSubmissions] = useState<CreativeSubmission[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const loadTimeline = async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await apiClient.get<CreativeSubmission[]>('/api/creative')
      // Sort newest first
      const sorted = [...data].sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
      setSubmissions(sorted)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'failed to load timeline')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadTimeline()
  }, [])

  return (
    <div className="flex flex-col gap-8 p-4 pb-safe max-w-2xl mx-auto w-full">
      <div>
        <h1 className="text-2xl font-bold">Family Journal</h1>
        <p className="text-sm text-muted-foreground">
          Shared board notes and quest memories from your adventures.
        </p>
      </div>

      {/* Slice 2.3: shared crew text board (append-only posts, not live collab) */}
      <SharedTextBoard />

      <section className="flex flex-col gap-4">
        <div>
          <h2 className="text-xl font-bold">Quest memories</h2>
          <p className="text-sm text-muted-foreground">
            STORY, DRAWING, and COMIC submissions from completed quests.
          </p>
        </div>

        {loading && submissions.length === 0 ? (
          <div className="flex h-32 items-center justify-center">
            <p className="text-sm text-muted-foreground animate-pulse">Loading memories…</p>
          </div>
        ) : error ? (
          <div className="flex flex-col items-center gap-2 p-6">
            <p className="text-sm text-red-500">{error}</p>
            <button onClick={loadTimeline} className="text-sm text-primary underline">Retry</button>
          </div>
        ) : submissions.length === 0 ? (
          <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border bg-surface/50 p-12 text-center">
            <p className="text-muted-foreground">No quest memories yet.</p>
            <p className="mt-1 text-sm text-muted-foreground">Complete quests to start writing your family story!</p>
          </div>
        ) : (
          <div className="flex flex-col gap-4">
            {submissions.map((sub) => (
              <CreativeCard key={sub.id} submission={sub} />
            ))}
          </div>
        )}
      </section>
    </div>
  )
}
