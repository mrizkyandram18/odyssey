import React from 'react'
import { CheckCircle2, Coins, Calendar, Users, Sliders } from 'lucide-react'
import type { RedemptionConfig } from '../../../shared/types'

interface AdminMetricsBarProps {
  pendingSubmissionsCount?: number
  pendingClaimsCount?: number
  config?: RedemptionConfig | null
}

export const AdminMetricsBar: React.FC<AdminMetricsBarProps> = ({
  pendingSubmissionsCount,
  pendingClaimsCount,
  config,
}) => {
  return (
    <div className="grid grid-cols-2 lg:grid-cols-5 gap-2.5">
      <div className="p-3 rounded-2xl bg-surface border border-border-subtle">
        <span className="text-[11px] font-bold text-text-secondary flex items-center gap-1">
          <CheckCircle2 className="w-3.5 h-3.5 text-accent-magic" />
          Verifikasi
        </span>
        <p className="mt-1 flex items-baseline gap-1">
          <span className="text-lg font-bold text-text-primary">
            {pendingSubmissionsCount !== undefined ? pendingSubmissionsCount : 0}
          </span>
          <span className="text-[11px] text-text-secondary">antrean</span>
        </p>
      </div>

      <div className="p-3 rounded-2xl bg-surface border border-border-subtle">
        <span className="text-[11px] font-bold text-text-secondary flex items-center gap-1">
          <Coins className="w-3.5 h-3.5 text-status-success" />
          Pencairan
        </span>
        <p className="mt-1 flex items-baseline gap-1">
          <span className="text-lg font-bold text-text-primary">
            {pendingClaimsCount !== undefined ? pendingClaimsCount : 0}
          </span>
          <span className="text-[11px] text-text-secondary">klaim</span>
        </p>
      </div>

      <div className="p-3 rounded-2xl bg-surface border border-border-subtle">
        <span className="text-[11px] font-bold text-text-secondary flex items-center gap-1">
          <Calendar className="w-3.5 h-3.5 text-accent-cyan" />
          Tugas
        </span>
        <p className="mt-1 flex items-baseline gap-1">
          <span className="text-lg font-bold text-text-primary">Jadwal</span>
          <span className="text-[11px] text-text-secondary">harian</span>
        </p>
      </div>

      <div className="p-3 rounded-2xl bg-surface border border-border-subtle">
        <span className="text-[11px] font-bold text-text-secondary flex items-center gap-1">
          <Users className="w-3.5 h-3.5 text-accent-magic" />
          Anggota
        </span>
        <p className="mt-1 flex items-baseline gap-1">
          <span className="text-lg font-bold text-text-primary">Keluarga</span>
          <span className="text-[11px] text-text-secondary">eksplorasi</span>
        </p>
      </div>

      <div className="p-3 rounded-2xl bg-surface border border-border-subtle col-span-2 lg:col-span-1">
        <span className="text-[11px] font-bold text-text-secondary flex items-center gap-1">
          <Sliders className="w-3.5 h-3.5 text-accent-gold" />
          Periode
        </span>
        <p className="mt-1 flex items-baseline gap-1">
          <span className="text-base font-bold text-text-primary">
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
      </div>
    </div>
  )
}
