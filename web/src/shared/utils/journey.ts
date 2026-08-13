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
    slug: 'literasi-keluarga',
    name: 'Literasi Keluarga',
    description: 'Membangun kebiasaan baik, komunikasi, dan ikatan dalam keluarga.',
    icon: '🏡',
    order: 1,
  },
  {
    slug: 'literasi-finansial',
    name: 'Literasi Finansial',
    description: 'Mengenal konsep dasar keuangan, menabung, dan mengelola pengeluaran.',
    icon: '💰',
    order: 2,
  },
  {
    slug: 'persiapan-karier',
    name: 'Persiapan Karier',
    description: 'Menggali minat, bakat, dan keterampilan dasar untuk masa depan.',
    icon: '🚀',
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

export function getRealmForMission(Mission: { template_slug: string; journey?: string }): string {
  if (Mission.journey && Mission.journey.trim() !== '') {
    return Mission.journey.toLowerCase().trim()
  }
  
  const slug = (Mission.template_slug || '').toLowerCase().trim()
  
  if (slug.includes('finansial') || slug.includes('uang') || slug.includes('nabung') || slug.includes('diskon')) {
    return 'literasi-finansial'
  }
  
  if (slug.includes('karier') || slug.includes('pekerjaan') || slug.includes('bisnis') || slug.includes('presentasi')) {
    return 'persiapan-karier'
  }
  
  return 'literasi-keluarga'
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
