import { useState } from 'react'
import { apiClient } from '../../shared/lib/api'
import { CreativeCanvas } from '../quest/CreativeCanvas'
import {
  buildComicPayload,
  isComicReady,
  MAX_COMIC_CAPTION,
  MAX_COMIC_PANELS,
  MIN_COMIC_PANELS,
  type ComicPanel,
} from '../../shared/utils/comic'

export interface SubmissionFormProps {
  questId: number
  challengeId: number
  onComplete: () => void
  onSkip: () => void
}

type Mode = 'STORY' | 'DRAWING' | 'COMIC'

const emptyPanels = (): ComicPanel[] => [
  { caption: '' },
  { caption: '' },
]

export function SubmissionForm({ questId, challengeId, onComplete, onSkip }: SubmissionFormProps) {
  const [mode, setMode] = useState<Mode>('STORY')
  const [content, setContent] = useState('')
  const [panels, setPanels] = useState<ComicPanel[]>(emptyPanels)
  const [sketchPanelIndex, setSketchPanelIndex] = useState<number | null>(null)
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

  const handleComicSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!isComicReady(panels)) return
    await submitCreative('COMIC', buildComicPayload(panels))
  }

  const submitCreative = async (kind: 'STORY' | 'DRAWING' | 'COMIC', payload: string) => {
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
        <h2 className="text-xl font-bold">Create a Memory</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          You completed a quest! Capture the moment.
        </p>
      </div>

      <div className="flex flex-wrap justify-center gap-2 mb-2">
        {modeBtn('STORY', 'Write Story')}
        {modeBtn('DRAWING', 'Draw Canvas')}
        {modeBtn('COMIC', 'Comic Strip')}
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
      ) : mode === 'DRAWING' ? (
        <CreativeCanvas
          onSubmit={handleCanvasSubmit}
          onCancel={onSkip}
          isSubmitting={submitting}
        />
      ) : sketchPanelIndex !== null ? (
        <div className="flex flex-col gap-3">
          <p className="text-sm text-center text-muted-foreground">
            Sketch for panel {sketchPanelIndex + 1}
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
            2–4 panels. Each panel needs a short caption and/or a sketch.
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
                    Remove
                  </button>
                )}
              </div>
              <textarea
                value={panel.caption}
                onChange={(e) => updatePanelCaption(index, e.target.value)}
                maxLength={MAX_COMIC_CAPTION}
                placeholder={`What happens in panel ${index + 1}?`}
                className="min-h-[72px] w-full resize-none rounded-lg border border-border bg-background p-3 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                disabled={submitting}
              />
              <div className="flex flex-wrap items-center gap-2">
                {panel.svg ? (
                  <>
                    <span className="text-xs font-medium text-primary">Sketch attached</span>
                    <button
                      type="button"
                      onClick={() => setSketchPanelIndex(index)}
                      className="text-xs text-primary underline"
                      disabled={submitting}
                    >
                      Redraw
                    </button>
                    <button
                      type="button"
                      onClick={() => clearPanelSketch(index)}
                      className="text-xs text-muted-foreground underline"
                      disabled={submitting}
                    >
                      Clear sketch
                    </button>
                  </>
                ) : (
                  <button
                    type="button"
                    onClick={() => setSketchPanelIndex(index)}
                    className="text-xs font-semibold text-primary underline"
                    disabled={submitting}
                  >
                    Add sketch (optional)
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
              + Add panel
            </button>
          )}

          <div className="flex flex-col gap-2 pt-2">
            <button
              type="submit"
              disabled={submitting || !isComicReady(panels)}
              className="w-full rounded-lg bg-primary py-3 text-sm font-bold text-primary-foreground transition-all hover:brightness-110 disabled:opacity-50"
            >
              {submitting ? 'Saving...' : 'Save Comic'}
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
      )}
    </div>
  )
}
