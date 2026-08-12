import { useState } from 'react'
import { apiClient } from '../../shared/lib/api'
import { CreativeCanvas } from '../mission/CreativeCanvas'
import {
  buildComicPayload,
  isComicReady,
  MAX_COMIC_CAPTION,
  MAX_COMIC_PANELS,
  MIN_COMIC_PANELS,
  type ComicPanel,
} from '../../shared/utils/comic'
import {
  buildPhotoPayload,
  fileToDataURL,
  isImageFile,
  isImageDataURL,
  MAX_PHOTO_CAPTION,
  buildVideoPayload,
  fileToVideoDataURL,
  isVideoFile,
  isVideoDataURL,
  MAX_VIDEO_CAPTION,
} from '../../shared/utils/media'

export interface SubmissionFormProps {
  missionId: number
  exerciseId: number
  onComplete: () => void
  onSkip: () => void
}

type Mode = 'STORY' | 'DRAWING' | 'COMIC' | 'PHOTO' | 'VIDEO'

const emptyPanels = (): ComicPanel[] => [
  { caption: '' },
  { caption: '' },
]

export function SubmissionForm({ missionId, exerciseId, onComplete, onSkip }: SubmissionFormProps) {
  const [mode, setMode] = useState<Mode>('STORY')
  const [content, setContent] = useState('')
  const [panels, setPanels] = useState<ComicPanel[]>(emptyPanels)
  const [sketchPanelIndex, setSketchPanelIndex] = useState<number | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const [photoURL, setPhotoURL] = useState('')
  const [photoCaption, setPhotoCaption] = useState('')

  const [videoURL, setVideoURL] = useState('')
  const [videoCaption, setVideoCaption] = useState('')

  const handleSubmitText = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!content.trim()) return
    await submitCreative('STORY', content)
  }

  const handleCanvasSubmit = async (svg: string) => {
    await submitCreative('DRAWING', svg)
  }

  const handleComicSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!isComicReady(panels)) return
    await submitCreative('COMIC', buildComicPayload(panels))
  }

  const handlePhotoSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!photoURL || !isImageDataURL(photoURL)) {
      setError('Pilih gambar')
      return
    }
    await submitCreative('PHOTO', buildPhotoPayload(photoURL, photoCaption))
  }

  const handleVideoSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!videoURL || !isVideoDataURL(videoURL)) {
      setError('Pilih video')
      return
    }
    await submitCreative('VIDEO', buildVideoPayload(videoURL, videoCaption))
  }

  const submitCreative = async (kind: 'STORY' | 'DRAWING' | 'COMIC' | 'PHOTO' | 'VIDEO', payload: string) => {
    setSubmitting(true)
    setError(null)
    try {
      await apiClient.post('/api/creative', {
        mission_id: missionId,
        exercise_id: exerciseId,
        kind,
        content: payload,
      })
      onComplete()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Gagal menyimpan kenangan')
      setSubmitting(false)
    }
  }

  const updatePanelCaption = (index: number, caption: string) => {
    setPanels((prev) => prev.map((p, i) => (i === index ? { ...p, caption } : p)))
  }

  const addPanel = () => {
    setPanels((prev) => (prev.length >= MAX_COMIC_PANELS ? prev : [...prev, { caption: '' }]))
  }

  const removePanel = (index: number) => {
    setPanels((prev) => {
      if (prev.length <= MIN_COMIC_PANELS) return prev
      return prev.filter((_, i) => i !== index)
    })
    if (sketchPanelIndex === index) setSketchPanelIndex(null)
  }

  const clearPanelSketch = (index: number) => {
    setPanels((prev) => prev.map((p, i) => (i === index ? { ...p, svg: undefined } : p)))
  }

  const modeBtn = (value: Mode, label: string) => (
    <button
      type="button"
      onClick={() => {
        setMode(value)
        setSketchPanelIndex(null)
        setError(null)
      }}
      className={`px-3 py-2 rounded-full text-sm font-bold ${
        mode === value ? 'bg-primary text-primary-foreground' : 'bg-background hover:bg-surface-hover'
      }`}
    >
      {label}
    </button>
  )

  return (
    <div className="flex w-full flex-col gap-4 rounded-xl border border-border bg-surface p-6 shadow-md">
      <div className="text-center mb-2">
        <h2 className="text-xl font-bold">Buat Kenangan</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Kamu menyelesaikan misi! Abadikan momen ini.
        </p>
      </div>

      <div className="flex flex-wrap justify-center gap-2 mb-2">
        {modeBtn('STORY', 'Tulis Cerita')}
        {modeBtn('DRAWING', 'Gambar Kanvas')}
        {modeBtn('COMIC', 'Strip Komik')}
        {modeBtn('PHOTO', 'Ambil Foto')}
        {modeBtn('VIDEO', 'Rekam Video')}
      </div>

      {error && <p className="text-center text-xs text-red-500">{error}</p>}

      {mode === 'STORY' ? (
        <form onSubmit={handleSubmitText} className="flex flex-col gap-4">
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder="Apa yang terjadi hari ini? (misal, Ayah menemukan harta karun!)"
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
              {submitting ? 'Menyimpan...' : 'Simpan Kenangan'}
            </button>
            <button
              type="button"
              onClick={onSkip}
              disabled={submitting}
              className="w-full rounded-lg py-3 text-sm font-semibold text-muted-foreground transition-all hover:bg-muted"
            >
              Lewati dulu
            </button>
          </div>
        </form>
      ) : mode === 'DRAWING' ? (
        <CreativeCanvas
          onSubmit={handleCanvasSubmit}
          onCancel={onSkip}
          isSubmitting={submitting}
        />
      ) : mode === 'PHOTO' ? (
        <form onSubmit={handlePhotoSubmit} className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <label className="text-xs text-muted-foreground">Foto</label>
            {photoURL ? (
              <div className="relative aspect-video w-full overflow-hidden rounded-lg border border-border bg-background/60">
                <img src={photoURL} alt="captured" className="h-full w-full object-cover" />
                <button
                  type="button"
                  onClick={() => setPhotoURL('')}
                  className="absolute right-1 top-1 rounded bg-black/60 px-2 py-1 text-xs text-white opacity-80 hover:opacity-100"
                >
                  Hapus
                </button>
              </div>
            ) : (
              <label className="flex cursor-pointer items-center justify-center gap-2 rounded-lg border border-dashed border-border py-8 text-center hover:bg-muted">
                <input
                  type="file"
                  accept="image/*"
                  capture="environment"
                  className="hidden"
                  onChange={async (e) => {
                    const file = e.target.files?.[0]
                    if (!file) return
                    if (!isImageFile(file)) {
                      setError('Hanya file gambar yang diizinkan')
                      return
                    }
                    setError(null)
                    try {
                      const url = await fileToDataURL(file)
                      setPhotoURL(url)
                    } catch (err) {
                      setError(err instanceof Error ? err.message : 'Gagal membaca foto')
                    }
                  }}
                />
                <span className="text-sm font-medium text-muted-foreground">
                  Ketuk untuk mengambil atau memilih foto
                </span>
              </label>
            )}
          </div>
          <div className="flex flex-col gap-2">
            <label className="text-xs text-muted-foreground">Keterangan (opsional)</label>
            <textarea
              value={photoCaption}
              onChange={(e) => setPhotoCaption(e.target.value.slice(0, MAX_PHOTO_CAPTION))}
              maxLength={MAX_PHOTO_CAPTION}
              placeholder="Apa yang terjadi di sini?"
              className="w-full resize-none rounded-lg border border-border bg-background p-3 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
              rows={3}
              disabled={submitting}
            />
          </div>
          <div className="flex flex-col gap-2 pt-2">
            <button
              type="submit"
              disabled={submitting || !photoURL || !isImageDataURL(photoURL)}
              className="w-full rounded-lg bg-primary py-3 text-sm font-bold text-primary-foreground transition-all hover:brightness-110 disabled:opacity-50"
            >
              {submitting ? 'Menyimpan...' : 'Simpan Foto'}
            </button>
            <button
              type="button"
              onClick={onSkip}
              disabled={submitting}
              className="w-full rounded-lg py-3 text-sm font-semibold text-muted-foreground transition-all hover:bg-muted"
            >
              Lewati dulu
            </button>
          </div>
        </form>
      ) : mode === 'VIDEO' ? (
        <form onSubmit={handleVideoSubmit} className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <label className="text-xs text-muted-foreground">Video</label>
            {videoURL ? (
              <div className="relative aspect-video w-full overflow-hidden rounded-lg border border-border bg-background/60">
                <video src={videoURL} controls className="h-full w-full object-cover" />
                <button
                  type="button"
                  onClick={() => setVideoURL('')}
                  className="absolute right-1 top-1 rounded bg-black/60 px-2 py-1 text-xs text-white opacity-80 hover:opacity-100"
                >
                  Hapus
                </button>
              </div>
            ) : (
              <label className="flex cursor-pointer items-center justify-center gap-2 rounded-lg border border-dashed border-border py-8 text-center hover:bg-muted">
                <input
                  type="file"
                  accept="video/*"
                  capture="environment"
                  className="hidden"
                  onChange={async (e) => {
                    const file = e.target.files?.[0]
                    if (!file) return
                    if (!isVideoFile(file)) {
                      setError('Hanya file video yang diizinkan')
                      return
                    }
                    setError(null)
                    try {
                      const url = await fileToVideoDataURL(file)
                      setVideoURL(url)
                    } catch (err) {
                      setError(err instanceof Error ? err.message : 'Gagal membaca video')
                    }
                  }}
                />
                <span className="text-sm font-medium text-muted-foreground">
                  Ketuk untuk merekam atau memilih video
                </span>
              </label>
            )}
          </div>
          <div className="flex flex-col gap-2">
            <label className="text-xs text-muted-foreground">Keterangan (opsional)</label>
            <textarea
              value={videoCaption}
              onChange={(e) => setVideoCaption(e.target.value.slice(0, MAX_VIDEO_CAPTION))}
              maxLength={MAX_VIDEO_CAPTION}
              placeholder="Apa yang terjadi di sini?"
              className="w-full resize-none rounded-lg border border-border bg-background p-3 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
              rows={3}
              disabled={submitting}
            />
          </div>
          <div className="flex flex-col gap-2 pt-2">
            <button
              type="submit"
              disabled={submitting || !videoURL || !isVideoDataURL(videoURL)}
              className="w-full rounded-lg bg-primary py-3 text-sm font-bold text-primary-foreground transition-all hover:brightness-110 disabled:opacity-50"
            >
              {submitting ? 'Menyimpan...' : 'Simpan Video'}
            </button>
            <button
              type="button"
              onClick={onSkip}
              disabled={submitting}
              className="w-full rounded-lg py-3 text-sm font-semibold text-muted-foreground transition-all hover:bg-muted"
            >
              Lewati dulu
            </button>
          </div>
        </form>
      ) : sketchPanelIndex !== null ? (
        <div className="flex flex-col gap-3">
          <p className="text-sm text-center text-muted-foreground">
            Sketsa untuk panel {sketchPanelIndex + 1}
          </p>
          <CreativeCanvas
            onSubmit={(svg) => {
              const idx = sketchPanelIndex
              setPanels((prev) => prev.map((p, i) => (i === idx ? { ...p, svg } : p)))
              setSketchPanelIndex(null)
            }}
            onCancel={() => setSketchPanelIndex(null)}
            isSubmitting={false}
          />
        </div>
      ) : (
        <form onSubmit={handleComicSubmit} className="flex flex-col gap-4">
          <p className="text-xs text-center text-muted-foreground">
            2–4 panel. Setiap panel butuh keterangan singkat dan/atau sketsa.
          </p>
          {panels.map((panel, index) => (
            <div
              key={index}
              className="flex flex-col gap-2 rounded-lg border border-border bg-background/60 p-3"
            >
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold uppercase tracking-wide text-muted-foreground">
                  Panel {index + 1}
                </span>
                {panels.length > MIN_COMIC_PANELS && (
                  <button
                    type="button"
                    onClick={() => removePanel(index)}
                    className="text-xs text-muted-foreground hover:text-destructive"
                    disabled={submitting}
                  >
                    Hapus
                  </button>
                )}
              </div>
              <textarea
                value={panel.caption}
                onChange={(e) => updatePanelCaption(index, e.target.value)}
                maxLength={MAX_COMIC_CAPTION}
                placeholder={`Apa yang terjadi di panel ${index + 1}?`}
                className="min-h-[72px] w-full resize-none rounded-lg border border-border bg-background p-3 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                disabled={submitting}
              />
              <div className="flex flex-wrap items-center gap-2">
                {panel.svg ? (
                  <>
                    <span className="text-xs font-medium text-primary">Sketsa terlampir</span>
                    <button
                      type="button"
                      onClick={() => setSketchPanelIndex(index)}
                      className="text-xs text-primary underline"
                      disabled={submitting}
                    >
                      Gambar ulang
                    </button>
                    <button
                      type="button"
                      onClick={() => clearPanelSketch(index)}
                      className="text-xs text-muted-foreground underline"
                      disabled={submitting}
                    >
                      Hapus sketsa
                    </button>
                  </>
                ) : (
                  <button
                    type="button"
                    onClick={() => setSketchPanelIndex(index)}
                    className="text-xs font-semibold text-primary underline"
                    disabled={submitting}
                  >
                    Tambah sketsa (opsional)
                  </button>
                )}
              </div>
            </div>
          ))}

          {panels.length < MAX_COMIC_PANELS && (
            <button
              type="button"
              onClick={addPanel}
              disabled={submitting}
              className="rounded-lg border border-dashed border-border py-2 text-sm font-semibold text-muted-foreground hover:bg-muted disabled:opacity-50"
            >
              + Tambah panel
            </button>
          )}

          <div className="flex flex-col gap-2 pt-2">
            <button
              type="submit"
              disabled={submitting || !isComicReady(panels)}
              className="w-full rounded-lg bg-primary py-3 text-sm font-bold text-primary-foreground transition-all hover:brightness-110 disabled:opacity-50"
            >
              {submitting ? 'Menyimpan...' : 'Simpan Komik'}
            </button>
            <button
              type="button"
              onClick={onSkip}
              disabled={submitting}
              className="w-full rounded-lg py-3 text-sm font-semibold text-muted-foreground transition-all hover:bg-muted"
            >
              Lewati dulu
            </button>
          </div>
        </form>
      )}
    </div>
  )
}
