import { useRef, useState } from 'react'
import { ReactSketchCanvas, type ReactSketchCanvasRef } from 'react-sketch-canvas'

export interface CreativeCanvasProps {
  onSubmit: (svg: string) => void
  onCancel: () => void
  isSubmitting?: boolean
}

export function CreativeCanvas({ onSubmit, onCancel, isSubmitting }: CreativeCanvasProps) {
  const canvasRef = useRef<ReactSketchCanvasRef>(null)
  const [strokeColor, setStrokeColor] = useState('#ffffff')
  const [strokeWidth] = useState(4)
  const [eraserMode, setEraserMode] = useState(false)

  const handleSubmit = async () => {
    if (!canvasRef.current) return
    try {
      const svg = await canvasRef.current.exportSvg()
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

  const colors = ['#ffffff', '#ff3366', '#33ccff', '#ffcc00', '#33ff99']

  return (
    <div className="flex flex-col gap-4 w-full h-[400px]">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {colors.map((color) => (
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
        </div>
        <div className="flex items-center gap-2">
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
        </div>
      </div>

      <div className="flex-1 rounded-lg overflow-hidden border-2 border-border/50 relative bg-[#1a1b26]">
        <ReactSketchCanvas
          ref={canvasRef}
          strokeWidth={strokeWidth}
          strokeColor={strokeColor}
          canvasColor="transparent"
          className="w-full h-full"
        />
      </div>

      <div className="flex items-center justify-end gap-3 mt-2">
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
  )
}
