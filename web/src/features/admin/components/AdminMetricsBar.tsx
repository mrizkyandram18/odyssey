import React from 'react'
import { CheckCircle2, Coins, Calendar, Users, Sliders } from 'lucide-react'
import type { RedemptionConfig } from '../../../shared/types'

interface AdminMetricsBarProps {
  activeTab: 'submissions' | 'claims' | 'tasks' | 'members' | 'settings'
  onSelectTab: (tab: 'submissions' | 'claims' | 'tasks' | 'members' | 'settings') => void
  pendingSubmissionsCount?: number
  pendingClaimsCount?: number
  config?: RedemptionConfig | null
}

export const AdminMetricsBar: React.FC<AdminMetricsBarProps> = ({
  activeTab,
  onSelectTab,
  pendingSubmissionsCount,
  pendingClaimsCount,
  config,
}) => {
  return (
    <div className="grid grid-cols-2 lg:grid-cols-5 gap-3">
      <button
        onClick={() => onSelectTab('submissions')}
        aria-current={activeTab === 'submissions' ? 'true' : undefined}
        className={`text-left p-3 rounded-2xl border transition-colors ${
          activeTab === 'submissions'
            ? 'bg-accent-magic/10 border-accent-magic/30'
            : 'bg-surface border-border-subtle hover:bg-surface-elevated'
        }`}
      >
        <span className="text-[11px] font-bold text-text-secondary flex items-center gap-1">
          <CheckCircle2 className="w-3.5 h-3.5 text-accent-magic" />
          Verifikasi
        </span>
        <p className="mt-1 flex items-baseline gap-1">
          <span className="text-xl font-bold text-text-primary">
            {pendingSubmissionsCount !== undefined ? pendingSubmissionsCount : '—'}
          </span>
          <span className="text-[11px] text-text-secondary">antrean</span>
        </p>
      </button>

      <button
        onClick={() => onSelectTab('claims')}
        aria-current={activeTab === 'claims' ? 'true' : undefined}
        className={`text-left p-3 rounded-2xl border transition-colors ${
          activeTab === 'claims'
            ? 'bg-status-success/10 border-status-success/30'
            : 'bg-surface border-border-subtle hover:bg-surface-elevated'
        }`}
      >
        <span className="text-[11px] font-bold text-text-secondary flex items-center gap-1">
          <Coins className="w-3.5 h-3.5 text-status-success" />
          Pencairan
        </span>
        <p className="mt-1 flex items-baseline gap-1">
          <span className="text-xl font-bold text-text-primary">
            {pendingClaimsCount !== undefined ? pendingClaimsCount : '—'}
          </span>
          <span className="text-[11px] text-text-secondary">klaim</span>
        </p>
      </button>

      <button
        onClick={() => onSelectTab('tasks')}
        aria-current={activeTab === 'tasks' ? 'true' : undefined}
        className={`text-left p-3 rounded-2xl border transition-colors ${
          activeTab === 'tasks'
            ? 'bg-accent-cyan/10 border-accent-cyan/30'
            : 'bg-surface border-border-subtle hover:bg-surface-elevated'
        }`}
      >
        <span className="text-[11px] font-bold text-text-secondary flex items-center gap-1">
          <Calendar className="w-3.5 h-3.5 text-accent-cyan" />
          Tugas
        </span>
        <p className="mt-1 flex items-baseline gap-1">
          <span className="text-xl font-bold text-text-primary">Jadwal</span>
          <span className="text-[11px] text-text-secondary">harian</span>
        </p>
      </button>

      <button
        onClick={() => onSelectTab('members')}
        aria-current={activeTab === 'members' ? 'true' : undefined}
        className={`text-left p-3 rounded-2xl border transition-colors ${
          activeTab === 'members'
            ? 'bg-accent-magic/10 border-accent-magic/30'
            : 'bg-surface border-border-subtle hover:bg-surface-elevated'
        }`}
      >
        <span className="text-[11px] font-bold text-text-secondary flex items-center gap-1">
          <Users className="w-3.5 h-3.5 text-accent-magic" />
          Anggota
        </span>
        <p className="mt-1 flex items-baseline gap-1">
          <span className="text-xl font-bold text-text-primary">Keluarga</span>
          <span className="text-[11px] text-text-secondary">eksplorasi</span>
        </p>
      </button>

      <button
        onClick={() => onSelectTab('settings')}
        aria-current={activeTab === 'settings' ? 'true' : undefined}
        className={`text-left p-3 rounded-2xl border transition-colors col-span-2 lg:col-span-1 ${
          activeTab === 'settings'
            ? 'bg-accent-gold/10 border-accent-gold/30'
            : 'bg-surface border-border-subtle hover:bg-surface-elevated'
        }`}
      >
        <span className="text-[11px] font-bold text-text-secondary flex items-center gap-1">
          <Sliders className="w-3.5 h-3.5 text-accent-gold" />
          Periode
        </span>
        <p className="mt-1 flex items-baseline gap-1">
          <span className="text-lg font-bold text-text-primary">
            {config ? `${config.redemption_start_day}–${config.redemption_end_day}` : '21–26'}
          </span>
          <span
            className={`text-[10px] font-bold px-1.5 py-0.5 rounded-full ${
              config?.is_open
                ? 'bg-status-success/15 text-status-success'
                : 'bg-surface-elevated text-text-secondary border border-border-subtle'
            }`}
          >
            {config?.is_open ? 'Buka' : 'Tutup'}
          </span>
        </p>
      </button>
    </div>
  )
}
