import React from 'react'
import { Shield, Info, Target } from 'lucide-react'
import type { MemberView } from '../../../../../shared/types'

interface MemberCapSectionProps {
  member: MemberView
  form: {
    income_status?: 'NOT_CONFIGURED' | 'CONFIGURED' | 'DISABLED'
    reason?: string
    monthly_coin_target: number | null
    monthly_earning_cap: number
  }
  setForm: React.Dispatch<React.SetStateAction<any>>
}

export const MemberCapSection: React.FC<MemberCapSectionProps> = ({
  member,
  form,
  setForm,
}) => {
  const currentStatus = form.income_status || 'NOT_CONFIGURED'
  const earned = member.earned_this_period ?? 0

  const handleStatusChange = (newStatus: 'NOT_CONFIGURED' | 'CONFIGURED' | 'DISABLED') => {
    setForm((prev: any) => {
      let nextCap = 0
      if (newStatus === 'CONFIGURED') {
        if (prev.monthly_earning_cap > 0) {
          nextCap = prev.monthly_earning_cap
        } else if (member.monthly_earning_cap && member.monthly_earning_cap > 0) {
          nextCap = member.monthly_earning_cap
        } else {
          nextCap = 0
        }
      }
      return {
        ...prev,
        income_status: newStatus,
        monthly_earning_cap: nextCap,
      }
    })
  }
  // NULL (or missing) = inherit system default · 0 = no coin payout · N = explicit pool.
  const targetState =
    form.monthly_coin_target == null
      ? 'Ikut default'
      : form.monthly_coin_target === 0
        ? 'Tanpa payout'
        : `${form.monthly_coin_target.toLocaleString('id-ID')} koin`

  const effectiveTarget = form.monthly_coin_target ?? 3320

  return (
    <div className="space-y-4">
      {/* Panduan Edukatif Target vs Batas Koin */}
      <div className="p-3.5 rounded-xl bg-accent-magic/5 border border-accent-magic/20 space-y-2.5">
        <div className="flex items-center gap-1.5 font-bold text-accent-magic text-[11px] uppercase tracking-wider">
          <Info className="w-3.5 h-3.5" />
          <span>Panduan: Perbedaan Target vs Batas Koin</span>
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-[11px] text-text-secondary leading-relaxed">
          <div className="p-2.5 rounded-lg bg-surface border border-border-subtle space-y-1">
            <strong className="text-text-primary flex items-center gap-1">
              <Target className="w-3.5 h-3.5 text-accent-magic" />
              <span>Target Koin Bulanan</span>
            </strong>
            <p>
              Estimasi reward yang dirancang untuk dicapai anggota per bulan. Sistem membagi koin pada tiap tugas harian agar total sebulan mencapai angka ini.
            </p>
          </div>
          <div className="p-2.5 rounded-lg bg-surface border border-border-subtle space-y-1">
            <strong className="text-text-primary flex items-center gap-1">
              <Shield className="w-3.5 h-3.5 text-accent-gold" />
              <span>Batas Koin (Earning Cap)</span>
            </strong>
            <p>
              Plafon keamanan maksimal per bulan. Jika anggota mengerjakan banyak bonus quest, koin berhenti bertambah setelah batas ini tercapai.
            </p>
          </div>
        </div>
        <p className="text-[11px] text-text-secondary leading-relaxed">
          💡 <strong>Rekomendasi Admin:</strong> Set Batas Koin <strong>sama dengan Target Koin</strong>, atau beri toleransi bonus 10%–20% (misal Target 3.320 &rarr; Batas 3.500–4.000 koin).
        </p>
      </div>

      {/* Target Koin Bulanan */}
      <div className="space-y-1.5 p-3.5 rounded-xl bg-surface-elevated border border-border-subtle">
        <div className="flex items-center justify-between">
          <label htmlFor="input-monthly-target" className="text-xs font-bold text-text-primary flex items-center gap-1.5">
            <Target className="w-3.5 h-3.5 text-accent-magic" />
            <span>Target Koin Bulanan Anggota</span>
          </label>
          <span className="text-[10px] text-text-secondary font-mono">{targetState}</span>
        </div>
        <div className="relative">
          <input
            id="input-monthly-target"
            type="number"
            min={0}
            max={10000}
            value={form.monthly_coin_target ?? ''}
            placeholder="Ikut default sistem"
            onChange={(e) =>
              setForm((prev: any) => ({
                ...prev,
                monthly_coin_target: e.target.value === '' ? null : parseInt(e.target.value, 10),
              }))
            }
            className="w-full p-2.5 pr-16 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
          />
          <span className="absolute right-3 top-2.5 text-xs text-text-secondary">koin</span>
        </div>
        <p className="text-[11px] text-text-secondary">
          Target perolehan koin bulanan untuk anggota ini. Kosongkan untuk ikut default sistem.
          Nilai 0 berarti tanpa payout koin (XP dan streak tetap berjalan).
        </p>
      </div>

      {/* Batas Koin Bulanan (Cap Earning) & Status Konfigurasi */}
      <div className="space-y-3 p-3.5 rounded-xl bg-surface-elevated border border-border-subtle">
        <div className="flex items-center justify-between">
          <span className="text-xs font-bold text-text-primary flex items-center gap-1.5">
            <Shield className="w-3.5 h-3.5 text-accent-gold" />
            <span>Batas Koin Bulanan (Earning Cap)</span>
          </span>
          <span className="text-[10px] font-bold px-2 py-0.5 rounded-md bg-surface border border-border-subtle text-text-secondary">
            {currentStatus === 'CONFIGURED'
              ? (form.monthly_earning_cap > 0
                ? `Batas Khusus: ${form.monthly_earning_cap.toLocaleString('id-ID')} koin`
                : 'Batas Belum Diisi')
              : currentStatus === 'DISABLED'
                ? 'Dinonaktifkan (Rp0)'
                : 'Belum Diatur (Rp0)'}
          </span>
        </div>

        {/* Status Konfigurasi Selector */}
        <div className="space-y-2">
          <label className="text-[11px] font-bold text-text-secondary block">
            Status Konfigurasi Earning:
          </label>
          <div className="grid grid-cols-1 gap-2">
            <label className={`flex items-start gap-2.5 p-2.5 rounded-xl border cursor-pointer transition-all ${
              currentStatus === 'CONFIGURED'
                ? 'bg-accent-magic/10 border-accent-magic text-text-primary'
                : 'bg-surface border-border-subtle text-text-secondary hover:border-border-strong'
            }`}>
              <input
                type="radio"
                name="income_status"
                value="CONFIGURED"
                checked={currentStatus === 'CONFIGURED'}
                onChange={() => handleStatusChange('CONFIGURED')}
                className="mt-0.5 text-accent-magic cursor-pointer"
              />
              <div className="text-xs">
                <span className="font-bold text-text-primary block">Batas Koin Aktif (Dikonfigurasi)</span>
                <span className="text-[11px] text-text-secondary">Anggota dapat memperoleh koin tugas hingga batas bulanan yang ditentukan.</span>
              </div>
            </label>

            <label className={`flex items-start gap-2.5 p-2.5 rounded-xl border cursor-pointer transition-all ${
              currentStatus === 'NOT_CONFIGURED'
                ? 'bg-status-warning/10 border-status-warning text-text-primary'
                : 'bg-surface border-border-subtle text-text-secondary hover:border-border-strong'
            }`}>
              <input
                type="radio"
                name="income_status"
                value="NOT_CONFIGURED"
                checked={currentStatus === 'NOT_CONFIGURED'}
                onChange={() => handleStatusChange('NOT_CONFIGURED')}
                className="mt-0.5 text-status-warning cursor-pointer"
              />
              <div className="text-xs">
                <span className="font-bold text-text-primary block">Belum Diatur (Batas Rp0 — Antrean Triage)</span>
                <span className="text-[11px] text-text-secondary">Batas koin efektif 0 koin (Rp0). Anggota masuk antrean triage admin sampai dikonfigurasi.</span>
              </div>
            </label>

            <label className={`flex items-start gap-2.5 p-2.5 rounded-xl border cursor-pointer transition-all ${
              currentStatus === 'DISABLED'
                ? 'bg-zinc-500/10 border-zinc-500 text-text-primary'
                : 'bg-surface border-border-subtle text-text-secondary hover:border-border-strong'
            }`}>
              <input
                type="radio"
                name="income_status"
                value="DISABLED"
                checked={currentStatus === 'DISABLED'}
                onChange={() => handleStatusChange('DISABLED')}
                className="mt-0.5 text-zinc-500 cursor-pointer"
              />
              <div className="text-xs">
                <span className="font-bold text-text-primary block">Dinonaktifkan (Batas Rp0 — Ditutup)</span>
                <span className="text-[11px] text-text-secondary">Earning koin dinonaktifkan (Rp0). Anggota tidak memperoleh koin dan tidak masuk triage.</span>
              </div>
            </label>
          </div>
        </div>

        {/* Input Batas Koin Khusus (Hanya bila CONFIGURED) */}
        {currentStatus === 'CONFIGURED' ? (
          <div className="pt-2 border-t border-border-subtle/60 space-y-2">
            <div className="flex items-center justify-between">
              <label htmlFor="input-custom-cap" className="text-[11px] font-bold text-text-secondary">
                Batas Koin Bulanan (Wajib &gt; 0, Maks 10.000):
              </label>
              <div className="flex items-center gap-1 text-[10px] text-text-secondary">
                <span>Pilihan Cepat:</span>
                {effectiveTarget > 0 && (
                  <>
                    <button
                      type="button"
                      onClick={() => setForm((prev: any) => ({ ...prev, monthly_earning_cap: effectiveTarget }))}
                      className="px-2 py-0.5 rounded bg-surface border border-border-subtle hover:text-text-primary hover:border-accent-magic"
                    >
                      = Target ({effectiveTarget})
                    </button>
                    <button
                      type="button"
                      onClick={() => setForm((prev: any) => ({ ...prev, monthly_earning_cap: Math.round(effectiveTarget * 1.1) }))}
                      className="px-2 py-0.5 rounded bg-surface border border-border-subtle hover:text-text-primary hover:border-accent-magic"
                    >
                      +10% ({Math.round(effectiveTarget * 1.1)})
                    </button>
                    <button
                      type="button"
                      onClick={() => setForm((prev: any) => ({ ...prev, monthly_earning_cap: Math.round(effectiveTarget * 1.2) }))}
                      className="px-2 py-0.5 rounded bg-surface border border-border-subtle hover:text-text-primary hover:border-accent-magic"
                    >
                      +20% ({Math.round(effectiveTarget * 1.2)})
                    </button>
                  </>
                )}
              </div>
            </div>
            <div className="relative">
              <input
                id="input-custom-cap"
                type="number"
                min={1}
                max={10000}
                value={form.monthly_earning_cap > 0 ? form.monthly_earning_cap : ''}
                onChange={(e) =>
                  setForm((prev: any) => ({
                    ...prev,
                    monthly_earning_cap: Math.max(0, parseInt(e.target.value || '0', 10)),
                  }))
                }
                placeholder="Contoh: 3500"
                className={`w-full p-2.5 pr-24 rounded-xl bg-surface border text-xs sm:text-sm font-bold text-text-primary focus:outline-none font-mono ${
                  form.monthly_earning_cap <= 0
                    ? 'border-status-error focus:border-status-error'
                    : 'border-border-subtle focus:border-accent-magic'
                }`}
              />
              <span className="absolute right-3 top-2.5 text-xs text-text-secondary">koin / bulan</span>
            </div>
            {form.monthly_earning_cap <= 0 && (
              <p className="text-[11px] text-status-error font-medium">
                Batas koin harus diisi angka positif (minimal 1 koin, maksimal 10.000 koin).
              </p>
            )}
            <p className="text-[10px] text-text-secondary">
              Batas khusus ini hanya berlaku untuk anggota ini. 1 Koin bernilai Rp100.
            </p>
          </div>
        ) : (
          <div className="p-2.5 rounded-lg bg-surface border border-border-subtle text-[11px] text-text-secondary space-y-1">
            <span className="font-bold text-text-primary block">
              Batas Koin Efektif: 0 koin (Rp0)
            </span>
            <p>
              {currentStatus === 'NOT_CONFIGURED'
                ? 'Anggota belum memiliki batas koin yang dikonfigurasi. Anggota tidak akan menerima koin dari reward tugas sampai status diubah ke CONFIGURED.'
                : 'Earning koin dinonaktifkan secara eksplisit untuk anggota ini.'}
            </p>
          </div>
        )}

        {/* Current status summary */}
        <div className="pt-2 border-t border-border-subtle/60 space-y-2">
          {member.monthly_adjustment != null && member.monthly_adjustment > 0 && (
            <div className="p-2.5 rounded-lg bg-surface border border-accent-gold/30 text-[11px] space-y-1">
              <div className="font-bold text-accent-gold flex items-center justify-between">
                <span>Rincian Budget & Penyesuaian Historis:</span>
                <span className="font-mono">-{member.monthly_adjustment.toLocaleString('id-ID')} koin</span>
              </div>
              <p className="text-text-secondary text-[10px]">
                Base Budget: <strong>{(member.base_earning_cap ?? form.monthly_earning_cap).toLocaleString('id-ID')} koin</strong> - Alokasi Kelebihan: <strong>{member.monthly_adjustment.toLocaleString('id-ID')} koin</strong> = Budget Efektif: <strong className="text-text-primary">{(member.effective_earning_cap ?? (form.monthly_earning_cap - member.monthly_adjustment)).toLocaleString('id-ID')} koin (Rp {(((member.effective_earning_cap ?? (form.monthly_earning_cap - member.monthly_adjustment))) * 100).toLocaleString('id-ID')})</strong>
              </p>
            </div>
          )}

          <div className="flex items-center justify-between text-[11px]">
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
        </div>

        {/* Audit Reason Input */}
        <div className="pt-2 border-t border-border-subtle/60 space-y-1.5">
          <label htmlFor="input-cap-reason" className="text-[11px] font-bold text-text-primary">
            Alasan Perubahan Konfigurasi (Catatan Audit):
          </label>
          <input
            id="input-cap-reason"
            type="text"
            value={form.reason ?? ''}
            onChange={(e) => setForm((prev: any) => ({ ...prev, reason: e.target.value }))}
            placeholder="Contoh: Penyesuaian budget reward bulanan keluarga"
            className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic"
          />
        </div>

        <p className="text-[11px] text-text-secondary leading-relaxed">
          Jika perolehan koin bulan ini mencapai batas, anggota tidak dapat memperoleh koin tambahan dari tugas sampai awal bulan berikutnya. Saldo koin yang telah diperoleh tetap aman dan dapat ditarik kapan saja (minimal 500 koin).
        </p>
      </div>
    </div>
  )
}
