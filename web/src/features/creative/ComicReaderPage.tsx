import { useEffect, useState, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import { creativeApi } from '../../shared/lib/api'
import type { CreativeSubmission } from '../../shared/types'
import { parseComicPayload } from '../../shared/utils/comic'
import { toSvgDataUri } from '../../shared/utils/svg'

export function ComicReaderPage() {
  const { id } = useParams<{ id: string }>()
  const [submission, setSubmission] = useState<CreativeSubmission | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [panelIndex, setPanelIndex] = useState(0)

  const loadSubmission = useCallback(async () => {
    if (!id) return
    setLoading(true)
    setError(null)
    setPanelIndex(0)
    try {
      const data = await creativeApi.get(Number(id))
      setSubmission(data)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'failed to load comic')
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
          <p className="text-sm text-muted-foreground animate-pulse">Memuat komik…</p>
        </div>
      </div>
    )
  }

  if (error || !submission) {
    return (
      <div className="flex flex-col gap-4 p-4 pb-safe max-w-2xl mx-auto w-full">
        <h1 className="text-2xl font-bold">Pembaca Komik</h1>
        <div className="flex flex-col items-center gap-2 p-6">
          <p className="text-sm text-red-500">{error || 'Komik tidak ditemukan'}</p>
          <Link to="/gallery" className="text-sm text-primary underline">Kembali ke Galeri</Link>
        </div>
      </div>
    )
  }

  if (submission.kind !== 'COMIC') {
    return (
      <div className="flex flex-col gap-4 p-4 pb-safe max-w-2xl mx-auto w-full">
        <h1 className="text-2xl font-bold">Pembaca Komik</h1>
        <div className="flex flex-col items-center gap-2 p-6">
          <p className="text-sm text-red-500">Karya ini bukan komik.</p>
          <Link to="/gallery" className="text-sm text-primary underline">Kembali ke Galeri</Link>
        </div>
      </div>
    )
  }

  const comic = parseComicPayload(submission.content)
  if (!comic) {
    return (
      <div className="flex flex-col gap-4 p-4 pb-safe max-w-2xl mx-auto w-full">
        <h1 className="text-2xl font-bold">Pembaca Komik</h1>
        <div className="flex flex-col items-center gap-2 p-6">
          <p className="text-sm text-red-500">Komik tidak dapat ditampilkan.</p>
          <Link to="/gallery" className="text-sm text-primary underline">Kembali ke Galeri</Link>
        </div>
      </div>
    )
  }

  const panels = comic.panels
  const currentPanel = panels[panelIndex]
  const isFirst = panelIndex === 0
  const isLast = panelIndex === panels.length - 1

  return (
    <div className="flex flex-col gap-4 p-4 pb-safe max-w-2xl mx-auto w-full">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Pembaca Komik</h1>
        <Link to="/gallery" className="text-sm text-primary underline">Kembali ke Galeri</Link>
      </div>

      <div className="flex items-center gap-2">
        <span className="text-xs font-medium text-muted-foreground">
          Misi #{submission.mission_id} · Komik
        </span>
        <span className="text-xs text-muted-foreground">oleh {submission.author_uid}</span>
      </div>

      <div className="flex flex-col gap-3">
        <div className="rounded-lg border border-border bg-surface p-4">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-medium text-muted-foreground">
              Panel {panelIndex + 1} dari {panels.length}
            </span>
          </div>

          <div className="flex flex-col gap-3">
            <div
              className="rounded-md border border-border/60 bg-background/40 overflow-hidden"
            >
              {currentPanel.svg ? (
                <img
                  src={toSvgDataUri(currentPanel.svg)}
                  alt={`Comic panel ${panelIndex + 1}`}
                  className="w-full h-auto bg-white/5"
                />
              ) : null}
              {currentPanel.caption ? (
                <p className="px-3 py-2 text-sm leading-snug">{currentPanel.caption}</p>
              ) : null}
            </div>
          </div>
        </div>

        <div className="flex items-center justify-between">
          <button
            onClick={() => setPanelIndex((i) => i - 1)}
            disabled={isFirst}
            className="rounded-md bg-primary px-4 py-2 text-sm font-semibold text-black disabled:opacity-40"
          >
            Sebel
          </button>

          <div className="flex gap-1">
            {panels.map((_, i) => (
              <div
                key={i}
                className={`h-2 w-2 rounded-full transition-all ${
                  i === panelIndex ? 'bg-primary w-4' : 'bg-border'
                }`}
              />
            ))}
          </div>

          <button
            onClick={() => setPanelIndex((i) => i + 1)}
            disabled={isLast}
            className="rounded-md bg-primary px-4 py-2 text-sm font-semibold text-black disabled:opacity-40"
          >
            Lanjut
          </button>
        </div>
      </div>
    </div>
  )
}
