import type { CreativeItem } from '../../types'

export interface CreativeCanvasProps {
  items?: CreativeItem[]
  onSubmit?: (payload: unknown) => void
}

export function CreativeCanvas({ items, onSubmit }: CreativeCanvasProps) {
  void items
  void onSubmit

  return (
    <div className="flex h-64 w-full flex-col items-center justify-center rounded-lg border border-dashed border-border bg-surface">
      <p className="text-sm text-muted-foreground">Creative canvas — MVP supports text submissions</p>
    </div>
  )
}
