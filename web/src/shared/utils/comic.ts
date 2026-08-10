/** Multi-panel comic payload stored in CreativeSubmission.content (JSON string). */

export interface ComicPanel {
  caption: string
  svg?: string
}

export interface ComicPayload {
  v?: number
  panels: ComicPanel[]
}

export const MIN_COMIC_PANELS = 2
export const MAX_COMIC_PANELS = 4
export const MAX_COMIC_CAPTION = 280

export function parseComicPayload(content: string): ComicPayload | null {
  if (!content || !content.trim().startsWith('{')) return null
  try {
    const parsed = JSON.parse(content) as ComicPayload
    if (!parsed || !Array.isArray(parsed.panels) || parsed.panels.length === 0) {
      return null
    }
    return {
      v: parsed.v,
      panels: parsed.panels.map((p) => ({
        caption: typeof p?.caption === 'string' ? p.caption : '',
        svg: typeof p?.svg === 'string' && p.svg.trim() ? p.svg : undefined,
      })),
    }
  } catch {
    return null
  }
}

export function buildComicPayload(panels: ComicPanel[]): string {
  const normalized = panels.map((p) => {
    const caption = p.caption.trim()
    const svg = p.svg?.trim()
    const out: ComicPanel = { caption }
    if (svg) out.svg = svg
    return out
  })
  return JSON.stringify({ v: 1, panels: normalized })
}

export function isComicReady(panels: ComicPanel[]): boolean {
  if (panels.length < MIN_COMIC_PANELS || panels.length > MAX_COMIC_PANELS) {
    return false
  }
  return panels.every((p) => {
    const caption = p.caption.trim()
    const svg = p.svg?.trim() ?? ''
    if (!caption && !svg) return false
    if (caption.length > MAX_COMIC_CAPTION) return false
    return true
  })
}
