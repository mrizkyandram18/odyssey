import { useRef, useState, useCallback } from 'react'
import { ReactSketchCanvas, type ReactSketchCanvasRef } from 'react-sketch-canvas'
import type { Stamp, StampDef } from '../creative/stamps'
import { STAMP_LIBRARY, buildStamp } from '../creative/stamps'

export interface CreativeCanvasProps {
  onSubmit: (svg: string) => void
  onCancel: () => void
  isSubmitting?: boolean
  enableStamps?: boolean
}

const BACKGROUND_COLORS = [
  { id: 'dark', color: '#1a1b26' },
  { id: 'white', color: '#ffffff' },
  { id: 'yellow', color: '#fff8e1' },
  { id: 'blue', color: '#e3f2fd' },
  { id: 'pink', color: '#fce4ec' },
]

const STROKE_COLORS = ['#ffffff', '#ff3366', '#33ccff', '#ffcc00', '#33ff99']

export function CreativeCanvas({ onSubmit, onCancel, isSubmitting, enableStamps }: CreativeCanvasProps) {
  const canvasRef = useRef<ReactSketchCanvasRef>(null)
  const [strokeColor, setStrokeColor] = useState('#ffffff')
  const [strokeWidth] = useState(4)
  const [eraserMode, setEraserMode] = useState(false)
  const [background, setBackground] = useState(BACKGROUND_COLORS[0].color)
  const [stamps, setStamps] = useState<Stamp[]>([])
  const [selectedStampId, setSelectedStampId] = useState<string | null>(null)
  const [activeStampId, setActiveStampId] = useState<string | null>(null)

  const selectedStamp = stamps.find((s) => s.id === selectedStampId) || null

  const buildCompositeSvg = useCallback(async (): Promise<string | null> => {
    const canvasSvg = await canvasRef.current?.exportSvg()
    if (!canvasSvg) return null

    const parser = new DOMParser()
    const doc = parser.parseFromString(canvasSvg, 'image/svg+xml')
    const root = doc.documentElement
    if (root.tagName.toLowerCase() !== 'svg') return canvasSvg

    const width = root.getAttribute('width') || '800'
    const height = root.getAttribute('height') || '600'

    const canvasChildren = Array.from(root.childNodes)
      .map((node) =>
        node.nodeType === Node.ELEMENT_NODE ? (node as Element).outerHTML : node.textContent || ''
      )
      .join('')

    const sortedStamps = [...stamps].sort((a, b) => a.zIndex - b.zIndex)
    const stampsContent = sortedStamps
      .map(
        (stamp) =>
          `<g transform="translate(${stamp.x}, ${stamp.y})">` +
          `<svg width="${stamp.width}" height="${stamp.height}" viewBox="0 0 ${stamp.width} ${stamp.height}" xmlns="http://www.w3.org/2000/svg">` +
          stamp.svg +
          `</svg>` +
          `</g>`
      )
      .join('')

    return `<svg xmlns="http://www.w3.org/2000/svg" width="${width}" height="${height}" viewBox="0 0 ${width} ${height}">` +
      `<rect width="${width}" height="${height}" fill="${background}"/>` +
      canvasChildren +
      stampsContent +
      `</svg>`
  }, [stamps, background])

  const handleSubmit = async () => {
    if (!canvasRef.current) return
    try {
      const svg = enableStamps ? await buildCompositeSvg() : await canvasRef.current.exportSvg()
      if (!svg) return
      onSubmit(svg)
    } catch (e) {
      console.error('Failed to export SVG', e)
    }
  }

  const handleClear = () => {
    canvasRef.current?.clearCanvas()
  }

  const handleUndo = () => {
    canvasRef.current?.undo()
  }

  const toggleEraser = () => {
    setEraserMode((prev) => {
      const next = !prev
      canvasRef.current?.eraseMode(next)
      return next
    })
  }

  const handleCanvasClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!activeStampId) return
    const rect = e.currentTarget.getBoundingClientRect()
    const x = e.clientX - rect.left
    const y = e.clientY - rect.top
    const def = STAMP_LIBRARY.find((s) => s.id === activeStampId)
    if (!def) return
    const newStamp = buildStamp(def, x, y, stamps.length + 1)
    setStamps((prev) => [...prev, newStamp])
    setActiveStampId(null)
  }

  const handleSelectStamp = (def: StampDef) => {
    setActiveStampId(def.id)
    setSelectedStampId(null)
  }

  const handleSelectStampOnCanvas = (id: string) => {
    setSelectedStampId(id)
    setActiveStampId(null)
  }

  const bringForward = () => {
    if (!selectedStamp) return
    setStamps((prev) =>
      prev.map((s) => (s.id === selectedStamp.id ? { ...s, zIndex: s.zIndex + 1 } : s))
    )
  }

  const sendBackward = () => {
    if (!selectedStamp) return
    setStamps((prev) =>
      prev.map((s) => (s.id === selectedStamp.id ? { ...s, zIndex: Math.max(1, s.zIndex - 1) } : s))
    )
  }

  const removeSelectedStamp = () => {
    if (!selectedStampId) return
    setStamps((prev) => prev.filter((s) => s.id !== selectedStampId))
    setSelectedStampId(null)
  }

  return (
    <div className="flex flex-col gap-3 w-full h-[400px]">
      <div className="flex items-center justify-between flex-wrap gap-2">
        <div className="flex items-center gap-2">
          {STROKE_COLORS.map((color) => (
            <button
              key={color}
              type="button"
              className={`w-6 h-6 rounded-full border-2 ${
                strokeColor === color && !eraserMode ? 'border-primary scale-110' : 'border-transparent'
              }`}
              style={{ backgroundColor: color }}
              onClick={() => {
                setStrokeColor(color)
                if (eraserMode) {
                  setEraserMode(false)
                  canvasRef.current?.eraseMode(false)
                }
              }}
              aria-label={`Color ${color}`}
            />
          ))}
          {enableStamps &&
            BACKGROUND_COLORS.map((bg) => (
              <button
                key={bg.id}
                type="button"
                className={`w-6 h-6 rounded-full border-2 ${
                  background === bg.color ? 'border-primary scale-110' : 'border-transparent'
                }`}
                style={{ backgroundColor: bg.color }}
                onClick={() => setBackground(bg.color)}
                aria-label={`Background ${bg.id}`}
              />
            ))}
        </div>
        <div className="flex items-center gap-2">
          {enableStamps && (
            <>
              {STAMP_LIBRARY.map((def) => (
                <button
                  key={def.id}
                  type="button"
                  className={`w-8 h-8 rounded border-2 flex items-center justify-center ${
                    activeStampId === def.id ? 'border-primary scale-110' : 'border-transparent'
                  }`}
                  onClick={() => {
                    handleSelectStamp(def)
                  }}
                  aria-label={`Stamp ${def.name}`}
                  title={def.name}
                >
                  <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    xmlns="http://www.w3.org/2000/svg"
                    className="text-current"
                    dangerouslySetInnerHTML={{ __html: def.svg }}
                  />
                </button>
              ))}
            </>
          )}
          <button
            type="button"
            onClick={toggleEraser}
            className={`px-3 py-1 text-sm rounded ${
              eraserMode ? 'bg-primary text-primary-foreground' : 'bg-surface hover:bg-surface-hover'
            }`}
          >
            Eraser
          </button>
          <button
            type="button"
            onClick={handleUndo}
            className="px-3 py-1 text-sm bg-surface hover:bg-surface-hover rounded"
          >
            Undo
          </button>
          <button
            type="button"
            onClick={handleClear}
            className="px-3 py-1 text-sm bg-surface hover:bg-surface-hover rounded text-destructive"
          >
            Clear
          </button>
          {enableStamps && selectedStamp && (
            <>
              <button
                type="button"
                onClick={bringForward}
                className="px-2 py-1 text-xs bg-surface hover:bg-surface-hover rounded"
              >
                Bring Fwd
              </button>
              <button
                type="button"
                onClick={sendBackward}
                className="px-2 py-1 text-xs bg-surface hover:bg-surface-hover rounded"
              >
                Send Bwd
              </button>
              <button
                type="button"
                onClick={removeSelectedStamp}
                className="px-2 py-1 text-xs bg-destructive text-white rounded hover:opacity-80"
              >
                Remove
              </button>
            </>
          )}
        </div>
      </div>

      <div
        className="flex-1 rounded-lg overflow-hidden border-2 border-border/50 relative"
        style={{ backgroundColor: background }}
      >
        <ReactSketchCanvas
          ref={canvasRef}
          strokeWidth={strokeWidth}
          strokeColor={strokeColor}
          canvasColor="transparent"
          className="w-full h-full"
          style={{ pointerEvents: activeStampId ? 'none' : 'auto' }}
        />
        {enableStamps &&
          stamps.map((stamp) => (
            <div
              key={stamp.id}
              className={`absolute ${
                selectedStampId === stamp.id ? 'ring-2 ring-primary ring-offset-2 ring-offset-transparent' : ''
              }`}
              style={{
                left: stamp.x,
                top: stamp.y,
                zIndex: stamp.zIndex,
                width: stamp.width,
                height: stamp.height,
                cursor: 'pointer',
              }}
              onClick={() => handleSelectStampOnCanvas(stamp.id)}
              dangerouslySetInnerHTML={{ __html: stamp.svg }}
            />
          ))}
        {enableStamps && activeStampId && (
          <div
            className="absolute inset-0 cursor-crosshair"
            onClick={handleCanvasClick}
          />
        )}
      </div>

      <div className="flex items-center justify-between">
        {enableStamps && activeStampId && (
          <p className="text-xs text-muted-foreground">Click on the canvas to place the stamp</p>
        )}
        {enableStamps && activeStampId && (
          <button
            type="button"
            onClick={() => setActiveStampId(null)}
            className="text-xs text-muted-foreground underline"
          >
            Cancel stamp placement
          </button>
        )}
        <div className="flex items-center justify-end gap-3 mt-auto">
          <button
            type="button"
            onClick={onCancel}
            disabled={isSubmitting}
            className="px-4 py-2 text-sm font-medium hover:text-muted-foreground disabled:opacity-50"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={handleSubmit}
            disabled={isSubmitting}
            className="rounded-full bg-primary px-6 py-2 text-sm font-bold text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
          >
            {isSubmitting ? 'Submitting...' : 'Submit Drawing'}
          </button>
        </div>
      </div>
    </div>
  )
}
