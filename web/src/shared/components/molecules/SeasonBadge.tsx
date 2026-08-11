import type { SeasonSummary } from '../../types'
import { getRealmMetadata } from '../../utils/realm'

const STATE_LABEL: Record<string, string> = {
  ACTIVE: 'Aktif',
  UPCOMING: 'Mendatang',
  EXPIRED: 'Berakhir',
  INACTIVE: 'Tidak Aktif',
}

const STATE_STYLE: Record<string, string> = {
  ACTIVE: 'bg-accent-nature/20 text-accent-nature border-accent-nature/30',
  UPCOMING: 'bg-accent-magic/20 text-accent-magic border-accent-magic/30',
  EXPIRED: 'bg-surface-elevated text-text-secondary border-border-subtle',
  INACTIVE: 'bg-surface-elevated text-text-secondary border-border-subtle',
}

export interface SeasonBadgeProps {
  season: SeasonSummary
  progress?: {
    quests_completed: number
    realm_progress: number
    realm_status: string
  }
}

export function SeasonBadge({ season, progress }: SeasonBadgeProps) {
  const realmMeta = getRealmMetadata(season.definition.realm)
  const isActive = season.state === 'ACTIVE'

  return (
    <div
      className={`flex flex-col gap-2 p-4 rounded-xl border ${
        isActive
          ? 'border-accent-nature/30 bg-accent-nature/5'
          : 'border-border-subtle bg-surface-elevated/30'
      }`}
      data-testid="season-badge"
    >
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-lg" aria-hidden>
            {realmMeta?.icon || '🗓️'}
          </span>
          <div>
            <p className="text-sm font-bold text-text-primary leading-tight">
              {season.definition.name}
            </p>
            <p className="text-[10px] text-text-secondary uppercase tracking-wider">
              {realmMeta?.name || season.definition.realm}
            </p>
          </div>
        </div>
        <span
          className={`text-[10px] font-bold px-2.5 py-1 rounded-full uppercase tracking-wider border ${
            STATE_STYLE[season.state] || STATE_STYLE.INACTIVE
          }`}
        >
          {STATE_LABEL[season.state] || season.state}
        </span>
      </div>

      {isActive && progress && (
        <div className="flex flex-col gap-1 mt-1">
          <div className="flex items-center justify-between text-[11px] text-text-secondary">
            <span>Progres Ranah</span>
            <span className="tabular-nums">{progress.realm_progress}%</span>
          </div>
          <div className="h-1.5 w-full rounded-full bg-surface-elevated overflow-hidden">
            <div
              className="h-full rounded-full bg-accent-nature transition-all duration-500"
              style={{ width: `${Math.min(100, Math.max(0, progress.realm_progress))}%` }}
            />
          </div>
          <div className="flex items-center justify-between text-[11px] text-text-secondary">
            <span>Misi Selesai</span>
            <span className="tabular-nums">{progress.quests_completed}</span>
          </div>
        </div>
      )}
    </div>
  )
}
