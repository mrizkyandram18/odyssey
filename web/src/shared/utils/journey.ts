import type { JourneyProgress } from '../types'

export interface RealmMetadata {
  slug: string
  name: string
  description: string
  icon: string
  order: number
}

export const KNOWN_REALMS: RealmMetadata[] = [
  {
    slug: 'whispering-woods',
    name: 'Whispering Woods',
    description: 'Hutan kuno yang dipenuhi bisikan angin dan rahasia alam.',
    icon: '🌲',
    order: 1,
  },
  {
    slug: 'clockwork-city',
    name: 'Clockwork City',
    description: 'Kota mekanis serba roda gigi, uap, dan keajaiban tembaga.',
    icon: '⚙️',
    order: 2,
  },
  {
    slug: 'starlit-library',
    name: 'Starlit Library',
    description: 'Perpustakaan megah di bawah naungan bintang berkerlap-kerlip.',
    icon: '📚',
    order: 3,
  },
]

export function getRealmMetadata(slug: string): RealmMetadata | undefined {
  const normalized = slug.toLowerCase().trim()
  return KNOWN_REALMS.find((r) => r.slug.toLowerCase() === normalized)
}

export function formatRealmName(slug: string): string {
  const meta = getRealmMetadata(slug)
  if (meta) return meta.name
  return slug
    .replace(/-/g, ' ')
    .replace(/\b\w/g, (char) => char.toUpperCase())
}

const CLOCKWORK_SLUGS = new Set([
  'clockwork-intro',
  'gear-hunt',
  'the-copper-key',
  'clockwork-story',
  'gear-drawing',
])

const STARLIT_SLUGS = new Set(['star-observation', 'library-concept'])

export function getRealmForMission(Mission: { template_slug: string; journey?: string }): string {
  if (Mission.journey && Mission.journey.trim() !== '') {
    return Mission.journey.toLowerCase().trim()
  }
  const slug = (Mission.template_slug || '').toLowerCase().trim()
  if (CLOCKWORK_SLUGS.has(slug)) {
    return 'clockwork-city'
  }
  if (STARLIT_SLUGS.has(slug)) {
    return 'starlit-library'
  }
  return 'whispering-woods'
}

export function isRealmUnlocked(status: string): boolean {
  return status === 'ACTIVE' || status === 'COMPLETE'
}

export function getMergedJourneyProgress(serverRealms?: JourneyProgress[]): (RealmMetadata & {
  status: 'LOCKED' | 'ACTIVE' | 'COMPLETE'
  progress: number
  raw?: JourneyProgress
})[] {
  const progressMap = new Map<string, JourneyProgress>()
  if (serverRealms) {
    for (const r of serverRealms) {
      if (r && r.journey) {
        progressMap.set(r.journey.toLowerCase().trim(), r)
      }
    }
  }

  return KNOWN_REALMS.map((meta) => {
    const raw = progressMap.get(meta.slug)
    let status: 'LOCKED' | 'ACTIVE' | 'COMPLETE' = meta.order === 1 ? 'ACTIVE' : 'LOCKED'
    let progress = 0

    if (raw) {
      status = (raw.status as 'LOCKED' | 'ACTIVE' | 'COMPLETE') || status
      progress = typeof raw.progress === 'number' ? raw.progress : 0
    }

    return {
      ...meta,
      status,
      progress,
      raw,
    }
  })
}
