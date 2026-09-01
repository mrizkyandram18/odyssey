import React from 'react'
import { Sliders, AlertCircle, Check, Calendar, Coins, ArrowRight } from 'lucide-react'
import { useAdminConfig } from '../../hooks/useAdminConfig'

export const EconomySettingsForm: React.FC = () => {
  const {
    config,
    isSaving,
    successMsg,
    errorMsg,
    startDayInput,
    setStartDayInput,
    endDayInput,
    setEndDayInput,
    payoutDayInput,
    setPayoutDayInput,
    conversionRateInput,
    setConversionRateInput,
    timezoneInput,
    setTimezoneInput,
    handleSaveConfig,
  } = useAdminConfig()

  return (
    <div className="space-y-4">
      <div className="p-4 sm:p-5 rounded-2xl bg-surface border border-border-subtle shadow-xs space-y-4">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2.5 pb-3 border-b border-border-subtle/60">
          <div>
            <h3 className="font-heading font-bold text-text-primary text-sm sm:text-base flex items-center gap-2">
              <Sliders className="w-4 h-4 text-accent-magic" />
              <span>Pengaturan Pencairan Koin</span>
            </h3>
            <p className="text-[11px] text-text-secondary mt-0.5">
              Kelola periode pengajuan pencairan dan aturan pembayaran reward.
            </p>
          </div>

          {config && (
            <span
              className={`inline-flex items-center gap-1.5 text-xs font-bold px-3 py-1 rounded-full self-start sm:self-auto ${
                config.is_open
                  ? 'bg-status-success/15 text-status-success border border-status-success/20'
                  : 'bg-surface-elevated text-text-secondary border border-border-subtle'
              }`}
            >
              <span className={`w-2 h-2 rounded-full ${config.is_open ? 'bg-status-success animate-pulse' : 'bg-text-secondary'}`} />
              {config.is_open ? 'Sedang Dibuka' : 'Sedang Ditutup'}
            </span>
          )}
        </div>

        <form onSubmit={handleSaveConfig} className="space-y-5">
          {/* Section A: Periode Pencairan */}
          <div className="space-y-3">
            <h4 className="text-xs font-bold uppercase tracking-wider text-text-secondary flex items-center gap-1.5">
              <Calendar className="w-3.5 h-3.5 text-accent-magic" />
              <span>Periode Pencairan</span>
            </h4>

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <div className="space-y-1">
                <label htmlFor="input-start-day" className="text-xs font-bold text-text-secondary">
                  Tanggal Mulai <span className="text-status-error">*</span>
                </label>
                <input
                  id="input-start-day"
                  type="number"
                  min={1}
                  max={31}
                  required
                  value={startDayInput}
                  onChange={(e) => setStartDayInput(e.target.value)}
                  className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs sm:text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
                />
                <p className="text-[10px] text-text-secondary">Hari awal user dapat mengajukan pencairan.</p>
              </div>

              <div className="space-y-1">
                <label htmlFor="input-end-day" className="text-xs font-bold text-text-secondary">
                  Tanggal Akhir <span className="text-status-error">*</span>
                </label>
                <input
                  id="input-end-day"
                  type="number"
                  min={1}
                  max={31}
                  required
                  value={endDayInput}
                  onChange={(e) => setEndDayInput(e.target.value)}
                  className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs sm:text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
                />
                <p className="text-[10px] text-text-secondary">Hari terakhir user dapat mengajukan pencairan.</p>
              </div>

              <div className="space-y-1">
                <label htmlFor="input-payout-day" className="text-xs font-bold text-text-secondary">
                  Tanggal Payday <span className="text-status-error">*</span>
                </label>
                <input
                  id="input-payout-day"
                  type="number"
                  min={1}
                  max={31}
                  required
                  value={payoutDayInput}
                  onChange={(e) => setPayoutDayInput(e.target.value)}
                  className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs sm:text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
                />
                <p className="text-[10px] text-text-secondary">Hari pembayaran reward.</p>
              </div>

              <div className="space-y-1 sm:col-span-3">
                <label htmlFor="input-timezone" className="text-xs font-bold text-text-secondary">
                  Zona Waktu <span className="text-status-error">*</span>
                </label>
                <input
                  id="input-timezone"
                  type="text"
                  required
                  value={timezoneInput}
                  onChange={(e) => setTimezoneInput(e.target.value)}
                  className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs sm:text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
                />
                <p className="text-[10px] text-text-secondary">Contoh: Asia/Jakarta</p>
              </div>
            </div>
          </div>

          {/* Section B: Konversi Koin */}
          <div className="space-y-3 pt-3 border-t border-border-subtle">
            <h4 className="text-xs font-bold uppercase tracking-wider text-text-secondary flex items-center gap-1.5">
              <Coins className="w-3.5 h-3.5 text-accent-gold" />
              <span>Konversi Koin</span>
            </h4>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div className="space-y-1">
                <label htmlFor="input-conversion-rate" className="text-xs font-bold text-text-secondary">
                  Nilai 1 Koin (Rupiah) <span className="text-status-error">*</span>
                </label>
                <input
                  id="input-conversion-rate"
                  type="number"
                  min={1}
                  required
                  value={conversionRateInput}
                  onChange={(e) => setConversionRateInput(e.target.value)}
                  className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs sm:text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
                />
                <p className="text-[10px] text-text-secondary">Nilai rupiah yang digunakan saat menghitung nominal pencairan.</p>
              </div>
            </div>
          </div>

          {/* Compact Live Preview — redemption-only */}
          <div className="p-3 rounded-xl bg-surface-elevated border border-border-subtle text-xs flex flex-col sm:flex-row sm:items-center justify-between gap-2">
            <div className="flex items-center gap-1.5 text-text-secondary font-bold text-[11px] uppercase tracking-wider">
              <span>Ringkasan Parameter Aktif</span>
              <ArrowRight className="w-3 h-3 text-accent-magic" />
            </div>
            <div className="font-bold text-text-primary text-[11px] sm:text-xs">
              Penukaran: tanggal {startDayInput}–{endDayInput} • Payday: tanggal {payoutDayInput} • Nilai: Rp{conversionRateInput}/koin • Zona waktu: {timezoneInput}
            </div>
          </div>

          {/* Alerts */}
          {errorMsg && (
            <div className="p-3 rounded-xl bg-status-error/15 border border-status-error/30 text-status-error text-xs flex items-center gap-2">
              <AlertCircle className="w-4 h-4 shrink-0" />
              <span>{errorMsg}</span>
            </div>
          )}

          {successMsg && (
            <div className="p-3 rounded-xl bg-status-success/15 border border-status-success/30 text-status-success text-xs flex items-center gap-2">
              <Check className="w-4 h-4 shrink-0" />
              <span>{successMsg}</span>
            </div>
          )}

          <button
            type="submit"
            disabled={isSaving}
            className="w-full py-3 rounded-xl bg-accent-magic hover:brightness-110 active:scale-[0.99] text-white font-bold text-xs sm:text-sm shadow-xs transition-all flex items-center justify-center gap-2 disabled:opacity-50 cursor-pointer"
          >
            {isSaving ? (
              <span>Menyimpan Pengaturan...</span>
            ) : (
              <>
                <Check className="w-4 h-4" />
                <span>Simpan Pengaturan Periode</span>
              </>
            )}
          </button>
        </form>
      </div>
    </div>
  )
}
