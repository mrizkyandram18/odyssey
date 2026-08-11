import { useEffect, useState, useCallback } from 'react'
import { apiClient } from '../../shared/lib/api'
import type { CreativeSubmission, SubmissionKind } from '../../shared/types'
import { CreativeCard } from '../../shared/components/molecules/CreativeCard'

type FilterKind = 'ALL' | SubmissionKind

const FILTERS: { value: FilterKind; label: string }[] = [
  { value: 'ALL', label: 'All' },
  { value: 'STORY', label: 'Stories' },
  { value: 'COMIC', label: 'Comics' },
  { value: 'PHOTO', label: 'Photos' },
  { value: 'VIDEO', label: 'Videos' },
]

export function GalleryPage() {
  const [submissions, setSubmissions] = useState<CreativeSubmission[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [filter, setFilter] = useState<FilterKind>('ALL')

  const loadGallery = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const path = filter === 'ALL' ? '/api/creative' : `/api/creative?kind=${filter}`
      const data = await apiClient.get<CreativeSubmission[]>(path)
      const sorted = [...data].sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
      setSubmissions(sorted)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'failed to load gallery')
    } finally {
      setLoading(false)
    }
  }, [filter])

  useEffect(() => {
    loadGallery()
  }, [loadGallery])

  return (
    <div className="flex flex-col gap-6 p-4 pb-safe max-w-2xl mx-auto w-full">
      <div>
        <h1 className="text-2xl font-bold">Family Gallery</h1>
        <p className="text-sm text-muted-foreground">
          Creative contributions from your crew.
        </p>
      </div>

      <div className="flex gap-2 overflow-x-auto pb-1 -mx-1 px-1">
        {FILTERS.map((f) => (
          <button
            key={f.value}
            onClick={() => setFilter(f.value)}
            className={`whitespace-nowrap rounded-full px-4 py-1.5 text-sm font-medium transition-all ${
              filter === f.value
                ? 'bg-primary text-black shadow-md'
                : 'bg-surface border border-border text-text-secondary hover:text-text-primary'
            }`}
          >
            {f.label}
          </button>
        ))}
      </div>

      {loading && submissions.length === 0 ? (
        <div className="flex h-32 items-center justify-center">
          <p className="text-sm text-muted-foreground animate-pulse">Loading gallery…</p>
        </div>
      ) : error ? (
        <div className="flex flex-col items-center gap-2 p-6">
          <p className="text-sm text-red-500">{error}</p>
          <button onClick={loadGallery} className="text-sm text-primary underline">Retry</button>
        </div>
      ) : submissions.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border bg-surface/50 p-12 text-center">
          <p className="text-muted-foreground">No contributions yet.</p>
          <p className="mt-1 text-sm text-muted-foreground">Complete quests and submit creative work to fill the gallery.</p>
        </div>
      ) : (
        <div className="flex flex-col gap-4">
          {submissions.map((sub) => (
            <CreativeCard key={sub.id} submission={sub} />
          ))}
        </div>
      )}
    </div>
  )
}
