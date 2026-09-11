import React, { useState } from 'react'
import { Megaphone, X } from 'lucide-react'
import type { AnnouncementConfig } from '../../types'

interface AnnouncementBannerProps {
  announcement?: AnnouncementConfig | null
  storageKeyPrefix?: string
}

/**
 * Member-facing system announcement banner.
 * All content (enabled/title/body/audience/schedule/priority) originates
 * from the backend announcement config (odyssey_system_config
 * announcement_* keys via /api/shop/config). This component renders
 * whatever the backend marks visible and contains no business text.
 */
export const AnnouncementBanner: React.FC<AnnouncementBannerProps> = ({
  announcement,
  storageKeyPrefix = 'odyssey-announcement-dismissed',
}) => {
  const [dismissed, setDismissed] = useState(false)

  if (!announcement || !announcement.visible) return null
  if (dismissed) return null
  const title = (announcement.title || '').trim()
  const body = (announcement.body || '').trim()
  if (!title && !body) return null

  const storageKey = `${storageKeyPrefix}:${title}::${body}`.slice(0, 200)
  try {
    if (typeof localStorage !== 'undefined' && localStorage.getItem(storageKey) === '1') {
      return null
    }
  } catch {
    // ignore storage errors; still render
  }

  const handleDismiss = () => {
    try {
      if (typeof localStorage !== 'undefined') localStorage.setItem(storageKey, '1')
    } catch {
      // ignore
    }
    setDismissed(true)
  }

  const isUrgent = announcement.priority === 'urgent' || announcement.priority === 'high'

  return (
    <div
      data-testid="announcement-banner"
      className={`rounded-2xl border p-4 flex items-start gap-3 ${
        isUrgent
          ? 'bg-amber-50 dark:bg-amber-950/40 border-amber-300'
          : 'bg-surface border-border-subtle'
      }`}
    >
      <div className="w-9 h-9 rounded-xl bg-accent-magic/10 text-accent-magic flex items-center justify-center shrink-0">
        <Megaphone className="w-4 h-4" />
      </div>
      <div className="flex-1 min-w-0">
        {title && (
          <h3 data-testid="announcement-title" className="text-sm font-extrabold text-text-primary leading-snug">
            {title}
          </h3>
        )}
        {body && (
          <p data-testid="announcement-body" className="text-xs text-text-secondary leading-relaxed mt-1 whitespace-pre-wrap">
            {body}
          </p>
        )}
      </div>
      <button
        type="button"
        onClick={handleDismiss}
        aria-label="Tutup pengumuman"
        data-testid="announcement-dismiss"
        className="w-7 h-7 rounded-full bg-surface-elevated border border-border-subtle text-text-secondary hover:text-text-primary flex items-center justify-center transition-colors shrink-0"
      >
        <X className="w-3.5 h-3.5" />
      </button>
    </div>
  )
}

export default AnnouncementBanner
