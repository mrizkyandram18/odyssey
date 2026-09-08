import React from 'react'
import { Coins, Calendar } from 'lucide-react'

interface MemberPayoutSectionProps {
  form: {
    payout_frequency: 'THRESHOLD' | 'WEEKLY' | 'MONTHLY'
    minimum_withdrawal_coins: number
    payout_weekday: number
    payout_month_start_day: number
    payout_month_end_day: number
  }
  setForm: React.Dispatch<React.SetStateAction<any>>
}

const WEEKDAY_OPTIONS = [
  { value: 1, label: 'Senin' },
  { value: 2, label: 'Selasa' },
  { value: 3, label: 'Rabu' },
  { value: 4, label: 'Kamis' },
  { value: 5, label: 'Jumat' },
  { value: 6, label: 'Sabtu' },
  { value: 0, label: 'Minggu' },
]

export const MemberPayoutSection: React.FC<MemberPayoutSectionProps> = ({
  form,
  setForm,
}) => {
  return (
    <div className="space-y-4">
      <div className="p-3.5 rounded-xl bg-surface-elevated border border-border-subtle space-y-3">
        <div className="flex items-center gap-2">
          <Coins className="w-4 h-4 text-accent-gold" />
          <h4 className="text-xs font-bold uppercase tracking-wider text-text-primary">
            Kebijakan Pencairan Khusus Anggota
          </h4>
        </div>
        <p className="text-[11px] text-text-secondary">
          Atur jadwal dan ketentuan pencairan koin khusus untuk anggota ini jika berbeda dari pengaturan umum.
        </p>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-1">
          <div className="space-y-1">
            <label htmlFor="select-payout-freq" className="text-xs font-bold text-text-secondary">
              Kapan Pencairan Dilakukan?
            </label>
            <select
              id="select-payout-freq"
              value={form.payout_frequency}
              onChange={(e) =>
                setForm((prev: any) => ({
                  ...prev,
                  payout_frequency: e.target.value as 'THRESHOLD' | 'WEEKLY' | 'MONTHLY',
                }))
              }
              className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs font-bold text-text-primary focus:outline-none focus:border-accent-magic"
            >
              <option value="THRESHOLD">Saat Batas Koin Tercapai (Fleksibel)</option>
              <option value="WEEKLY">Setiap Minggu</option>
              <option value="MONTHLY">Setiap Bulan</option>
            </select>
          </div>

          <div className="space-y-1">
            <label htmlFor="input-min-withdrawal" className="text-xs font-bold text-text-secondary">
              Minimal Penarikan (Koin)
            </label>
            <input
              id="input-min-withdrawal"
              type="number"
              min={1}
              max={100000}
              value={form.minimum_withdrawal_coins}
              onChange={(e) =>
                setForm((prev: any) => ({
                  ...prev,
                  minimum_withdrawal_coins: parseInt(e.target.value || '500', 10),
                }))
              }
              className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
            />
          </div>
        </div>

        {form.payout_frequency === 'WEEKLY' && (
          <div className="pt-2 border-t border-border-subtle space-y-1">
            <label htmlFor="select-payout-weekday" className="text-xs font-bold text-text-secondary flex items-center gap-1.5">
              <Calendar className="w-3.5 h-3.5 text-accent-magic" />
              <span>Hari Pencairan Mingguan:</span>
            </label>
            <select
              id="select-payout-weekday"
              value={form.payout_weekday}
              onChange={(e) =>
                setForm((prev: any) => ({
                  ...prev,
                  payout_weekday: parseInt(e.target.value || '1', 10),
                }))
              }
              className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs font-bold text-text-primary focus:outline-none focus:border-accent-magic"
            >
              {WEEKDAY_OPTIONS.map((w) => (
                <option key={w.value} value={w.value}>
                  Hari {w.label}
                </option>
              ))}
            </select>
          </div>
        )}

        {form.payout_frequency === 'MONTHLY' && (
          <div className="pt-2 border-t border-border-subtle grid grid-cols-2 gap-2">
            <div className="space-y-1">
              <label htmlFor="input-month-start" className="text-[11px] font-bold text-text-secondary">
                Tanggal Mulai Pengajuan (1–31)
              </label>
              <input
                id="input-month-start"
                type="number"
                min={1}
                max={31}
                value={form.payout_month_start_day}
                onChange={(e) =>
                  setForm((prev: any) => ({
                    ...prev,
                    payout_month_start_day: parseInt(e.target.value || '24', 10),
                  }))
                }
                className="w-full p-2 rounded-lg bg-surface border border-border-subtle text-xs font-mono"
              />
            </div>
            <div className="space-y-1">
              <label htmlFor="input-month-end" className="text-[11px] font-bold text-text-secondary">
                Tanggal Selesai Pengajuan (1–31)
              </label>
              <input
                id="input-month-end"
                type="number"
                min={1}
                max={31}
                value={form.payout_month_end_day}
                onChange={(e) =>
                  setForm((prev: any) => ({
                    ...prev,
                    payout_month_end_day: parseInt(e.target.value || '26', 10),
                  }))
                }
                className="w-full p-2 rounded-lg bg-surface border border-border-subtle text-xs font-mono"
              />
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
