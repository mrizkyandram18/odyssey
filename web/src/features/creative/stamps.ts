export interface StampDef {
  id: string
  name: string
  svg: string
  width: number
  height: number
}

export interface Stamp extends StampDef {
  x: number
  y: number
  zIndex: number
}

export const STAMP_LIBRARY: StampDef[] = [
  {
    id: 'star',
    name: 'Star',
    svg: '<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" fill="currentColor"/>',
    width: 24,
    height: 24,
  },
  {
    id: 'heart',
    name: 'Heart',
    svg: '<path d="M12 21l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21z" fill="currentColor"/>',
    width: 24,
    height: 24,
  },
  {
    id: 'circle',
    name: 'Circle',
    svg: '<circle cx="12" cy="12" r="10" fill="currentColor"/>',
    width: 24,
    height: 24,
  },
  {
    id: 'square',
    name: 'Square',
    svg: '<rect x="2" y="2" width="20" height="20" fill="currentColor"/>',
    width: 24,
    height: 24,
  },
  {
    id: 'triangle',
    name: 'Triangle',
    svg: '<polygon points="12,2 22,20 2,20" fill="currentColor"/>',
    width: 24,
    height: 24,
  },
  {
    id: 'arrow',
    name: 'Arrow',
    svg: '<path d="M12 2l5 10h-10l5-10z" fill="currentColor"/><rect x="2" y="14" width="20" height="6" fill="currentColor"/>',
    width: 24,
    height: 24,
  },
  {
    id: 'cloud',
    name: 'Cloud',
    svg: '<path d="M19.35 10.04C18.67 6.59 15.64 4 12 4 9.11 4 6.6 5.64 5.35 8.04 2.34 8.36 0 10.91 0 14c0 3.31 2.69 6 6 6h13c2.76 0 5-2.24 5-5 0-2.64-2.05-4.78-4.65-4.96z" fill="currentColor"/>',
    width: 24,
    height: 24,
  },
  {
    id: 'sun',
    name: 'Sun',
    svg: '<circle cx="12" cy="12" r="5" fill="currentColor"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42" stroke="currentColor" stroke-width="2" fill="none"/>',
    width: 24,
    height: 24,
  },
]

export function buildStamp(def: StampDef, x: number, y: number, zIndex: number): Stamp {
  return {
    ...def,
    x,
    y,
    zIndex,
  }
}
