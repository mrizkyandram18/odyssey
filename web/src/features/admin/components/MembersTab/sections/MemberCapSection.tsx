import React from 'react'
import { Shield } from 'lucide-react'
import type { MemberView } from '../../../../../shared/types'

interface MemberCapSectionProps {
  member: MemberView
  form: {
    monthly_coin_target: number
    monthly_earning_cap: number
  }
  setForm: React.Dispatch<React.SetStateAction<any>>
}

export const MemberCapSection: React.FC<MemberCapSectionProps> = ({
  member,
  form,
  setForm,
}) => {
  const isCustomCap = form.monthly_earning_cap > 0
  const earned = member.earned_this_period ?? 0

  return (
    <div className="space-y-4">
      {/* Target Koin Bulanan */}
      <div className="space-y-1.5 p-3.5 rounded-xl bg-surface-elevated border border-border-subtle">
        <div className="flex items-center justify-between">
          <label htmlFor="input-monthly-target" className="text-xs font-bold text-text-primary">
            Target Koin Bulanan Anggota
          </label>
          <span className="text-[10px] text-text-secondary">0 = Ikuti target sistem</span>
        </div>
        <div className="relative">
          <input
            id="input-monthly-target"
            type="number"
            min={0}
            max={10000}
            value={form.monthly_coin_target}
            onChange={(e) =>
              setForm((prev: any) => ({
                ...prev,
                monthly_coin_target: parseInt(e.target.value || '0', 10),
              }))
            }
            className="w-full p-2.5 pr-16 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
          />
          <span className="absolute right-3 top-2.5 text-xs text-text-secondary">koin</span>
        </div>
        <p className="text-[11px] text-text-secondary">
          Target perolehan koin bulanan untuk anggota ini. Sistem menggunakan angka ini untuk memandu bobot tugas.
        </p>
      </div>

      {/* Batas Koin Bulanan (Cap Earning) */}
      <div className="space-y-3 p-3.5 rounded-xl bg-surface-elevated border border-border-subtle">
        <div className="flex items-center justify-between">
          <span className="text-xs font-bold text-text-primary flex items-center gap-1.5">
            <Shield className="w-3.5 h-3.5 text-accent-magic" />
            <span>Batas Koin Bulanan (Earning Cap)</span>
          </span>
          <span className="text-[10px] font-bold px-2 py-0.5 rounded-md bg-surface border border-border-subtle text-text-secondary">
            {isCustomCap ? 'Batas Khusus' : 'Batas Standar Sistem'}
          </span>
        </div>

        <div className="space-y-2">
          <label className="flex items-center gap-2 cursor-pointer text-xs text-text-primary">
            <input
              type="radio"
              name="edit_cap_mode"
              checked={!isCustomCap}
              onChange={() => setForm((prev: any) => ({ ...prev, monthly_earning_cap: 0 }))}
              className="text-accent-magic cursor-pointer"
            />
            <span>Gunakan batas koin bulanan standar sistem (global)</span>
          </label>

          <label className="flex items-center gap-2 cursor-pointer text-xs text-text-primary">
            <input
              type="radio"
              name="edit_cap_mode"
              checked={isCustomCap}
              onChange={() =>
                setForm((prev: any) => ({
                  ...prev,
                  monthly_earning_cap: prev.monthly_earning_cap > 0 ? prev.monthly_earning_cap : 3000,
                }))
              }
              className="text-accent-magic cursor-pointer"
            />
            <span>Atur batas koin khusus untuk anggota ini</span>
          </label>
        </div>

        {isCustomCap && (
          <div className="pt-2 border-t border-border-subtle/60 space-y-1">
            <label htmlFor="input-custom-cap" className="text-[11px] font-bold text-text-secondary">
              Batas Koin Khusus (Per Bulan):
            </label>
            <div className="relative">
              <input
                id="input-custom-cap"
                type="number"
                min={1}
                max={10000}
                value={form.monthly_earning_cap}
                onChange={(e) =>
                  setForm((prev: any) => ({
                    ...prev,
                    monthly_earning_cap: Math.max(0, parseInt(e.target.value || '0', 10)),
                  }))
                }
                placeholder="Contoh: 3500"
                className="w-full p-2.5 pr-24 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
              />
              <span className="absolute right-3 top-2.5 text-xs text-text-secondary">koin / bulan</span>
            </div>
            <p className="text-[10px] text-text-secondary">
              Batas khusus ini hanya berlaku untuk anggota ini dan menggantikan batas standar.
            </p>
          </div>
        )}

        {/* Current status summary */}
        <div className="pt-2 border-t border-border-subtle/60 flex items-center justify-between text-[11px]">
          <span className="text-text-secondary">
            Perolehan bulan ini: <strong className="text-text-primary font-bold">{earned.toLocaleString('id-ID')} koin</strong>
          </span>
          <span
            className={`font-bold px-2 py-0.5 rounded-full text-[10px] ${
              member.earning_locked
                ? 'bg-accent-reward/15 text-accent-reward'
                : 'bg-status-success/15 text-status-success'
            }`}
          >
            {member.earning_locked ? '🔒 Batas Tercapai' : '✓ Masih Bisa Memperoleh Koin'}
          </span>
        </div>

        <p className="text-[11px] text-text-secondary leading-relaxed">
          Jika perolehan koin bulan ini mencapai batas, anggota tidak dapat memperoleh koin tambahan sampai awal bulan berikutnya. Saldo koin yang telah diperoleh tetap aman dan tidak berkurang.
        </p>
      </div>
    </div>
  )
}
