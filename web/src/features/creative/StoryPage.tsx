import { useEffect, useState, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import { creativeApi } from '../../shared/lib/api'
import type { CreativeSubmission } from '../../shared/types'
import { toSvgDataUri } from '../../shared/utils/svg'

export function StoryPage() {
  const { id } = useParams<{ id: string }>()
  const [submission, setSubmission] = useState<CreativeSubmission | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const loadSubmission = useCallback(async () => {
    if (!id) return
    setLoading(true)
    setError(null)
    try {
      const data = await creativeApi.get(Number(id))
      setSubmission(data)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'failed to load story')
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => {
    loadSubmission()
  }, [loadSubmission])

  if (loading) {
    return (
      <div className="flex flex-col gap-4 p-4 pb-safe max-w-2xl mx-auto w-full">
        <div className="flex h-32 items-center justify-center">
          <p className="text-sm text-muted-foreground animate-pulse">Memuat cerita…</p>
        </div>
      </div>
    )
  }

  if (error || !submission) {
    return (
      <div className="flex flex-col gap-4 p-4 pb-safe max-w-2xl mx-auto w-full">
        <h1 className="text-2xl font-bold">Halaman Cerita</h1>
        <div className="flex flex-col items-center gap-2 p-6">
          <p className="text-sm text-red-500">{error || 'Cerita tidak ditemukan'}</p>
          <Link to="/gallery" className="text-sm text-primary underline">Kembali ke Galeri</Link>
        </div>
      </div>
    )
  }

  if (submission.kind !== 'DRAWING') {
    return (
      <div className="flex flex-col gap-4 p-4 pb-safe max-w-2xl mx-auto w-full">
        <h1 className="text-2xl font-bold">Halaman Cerita</h1>
        <div className="flex flex-col items-center gap-2 p-6">
          <p className="text-sm text-red-500">Karya ini bukan gambar.</p>
          <Link to="/gallery" className="text-sm text-primary underline">Kembali ke Galeri</Link>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4 p-4 pb-safe max-w-2xl mx-auto w-full">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Halaman Cerita</h1>
        <Link to="/gallery" className="text-sm text-primary underline">Kembali ke Galeri</Link>
      </div>

      <div className="flex items-center gap-2">
        <span className="text-xs font-medium text-muted-foreground">
          Misi #{submission.quest_id} · Gambar
        </span>
        <span className="text-xs text-muted-foreground">oleh {submission.author_uid}</span>
      </div>

      <div className="rounded-lg border border-border bg-surface p-4">
        <img
          src={toSvgDataUri(submission.content)}
          alt={`Story by ${submission.author_uid}`}
          className="w-full h-auto bg-white/5 rounded"
        />
      </div>
    </div>
  )
}
