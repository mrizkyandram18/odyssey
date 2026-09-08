// Level curve helpers. Backend formula (SOT): level = floor(sqrt(xp / 100)) + 1.
// Level L spans [(L-1)^2 * 100, L^2 * 100). Display-only; server owns XP/level.
export interface LevelProgress {
  level: number
  xp: number
  rangeStart: number
  nextLevelAt: number
  have: number
  required: number
  percent: number
}

export function levelProgress(xpRaw: number, levelRaw: number): LevelProgress {
  const xp = Math.max(0, Math.floor(xpRaw ?? 0))
  const level = Math.max(1, Math.floor(levelRaw ?? 1))
  const rangeStart = (level - 1) * (level - 1) * 100
  const nextLevelAt = level * level * 100
  const have = Math.max(0, xp - rangeStart)
  const required = Math.max(1, nextLevelAt - rangeStart)
  return {
    level,
    xp,
    rangeStart,
    nextLevelAt,
    have,
    required,
    percent: Math.min(100, Math.round((have / required) * 100)),
  }
}
