import { useState } from 'react'
import { apiClient } from '../../shared/lib/api'
import type { CreativeSubmission, SubmissionKind } from '../../shared/types'
import { Badge } from '../../shared/components/atoms/Badge'

const KIND_LABELS: Record<SubmissionKind, string> = {
  STORY: 'Cerita',
  COMIC: 'Komik',
  PHOTO: 'Foto',
  VIDEO: 'Video',
  DRAWING: 'Gambar',
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
      setError(e instanceof Error ? e.message : 'gagal memuat kiriman')
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
      setError(e instanceof Error ? e.message : 'gagal mengirim')
    } finally {
      setSubmitting(false)
    }
  }

  if (loading && submissions.length === 0) {
    return <p className="p-4 text-sm text-muted-foreground">Memuat...</p>
  }

  return (
    <div className="flex flex-col gap-4 p-4 pb-safe">
      <h1 className="text-xl font-semibold">Ruang Kreatif</h1>
      <p className="text-sm text-muted-foreground">
        Kirim karya kreatifmu untuk tantangan misi.
      </p>

      <form onSubmit={handleSubmit} className="flex flex-col gap-3 rounded-lg border border-border bg-surface p-4">
        <h2 className="text-sm font-medium">Kiriman Baru</h2>
        <div className="flex flex-col gap-2">
          <label className="text-xs text-muted-foreground">ID Misi</label>
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
          <label className="text-xs text-muted-foreground">ID Tantangan</label>
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
          <label className="text-xs text-muted-foreground">
            Tipe Kiriman (CERITA / GAMBAR / KOMIK / FOTO / VIDEO)
          </label>
          <select
            value={kind}
            onChange={(e) => setKind(e.target.value as SubmissionKind)}
            className="rounded-md border border-border bg-background p-2 text-sm"
          >
            <option value="STORY">Potongan Cerita</option>
            <option value="DRAWING">Gambar (SVG)</option>
            <option value="COMIC">Strip Komik (Panel JSON)</option>
            <option value="PHOTO">Foto (JSON data URI)</option>
            <option value="VIDEO">Video (JSON data URI)</option>
          </select>
        </div>
        <div className="flex flex-col gap-2">
          <label className="text-xs text-muted-foreground">Konten</label>
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            className="rounded-md border border-border bg-background p-2 text-sm"
            rows={4}
            placeholder="Deskripsikan karya kreatifmu..."
          />
        </div>
        <button
          type="submit"
          disabled={submitting}
          className="rounded-md bg-primary p-2 text-sm font-semibold text-black disabled:opacity-50"
        >
          {submitting ? 'Mengirim...' : 'Kirim'}
        </button>
      </form>

      {error && <p className="text-xs text-red-500">{error}</p>}

      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-medium text-muted-foreground">Kiriman</h2>
          <button
            onClick={loadSubmissions}
            className="text-xs text-primary"
          >
            Segarkan
          </button>
        </div>
        {loading ? (
          <p className="text-sm text-muted-foreground">Memuat...</p>
        ) : submissions.length === 0 ? (
          <p className="text-sm text-muted-foreground">Belum ada kiriman.</p>
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
                Misi #{sub.quest_id} · Tantangan #{sub.challenge_id}
              </p>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
