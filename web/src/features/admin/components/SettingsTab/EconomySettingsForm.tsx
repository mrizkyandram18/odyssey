import React from 'react'
import { Sliders, AlertCircle, Check } from 'lucide-react'
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
    earningPeriodInput,
    setEarningPeriodInput,
    conversionRateInput,
    setConversionRateInput,
    targetRupiahInput,
    setTargetRupiahInput,
    maxPayoutInput,
    setMaxPayoutInput,
    timezoneInput,
    setTimezoneInput,
    handleSaveConfig,
  } = useAdminConfig()

  const convRateNum = parseInt(conversionRateInput, 10) || 0
  const targetRpNum = parseInt(targetRupiahInput, 10) || 0
  const targetCoinsCalc = convRateNum > 0 ? Math.floor(targetRpNum / convRateNum) : 0

  return (
    <div className="space-y-4">
      <div className="p-5 rounded-2xl bg-surface border border-border-subtle shadow-sm space-y-4">
        <div className="flex items-start justify-between gap-3">
          <div>
            <h3 className="font-heading font-bold text-text-primary text-base flex items-center gap-2">
              <Sliders className="w-5 h-5 text-accent-magic" />
              <span>Pengaturan Periode Penukaran Koin</span>
            </h3>
            <p className="text-xs text-text-secondary mt-1 leading-relaxed">
              Tentukan tanggal kalender setiap bulan di mana anggota dapat mengajukan pencairan koin menjadi uang tunai.
            </p>
          </div>

          {config && (
            <span
              className={`text-xs font-bold px-3 py-1 rounded-full shrink-0 ${
                config.is_open
                  ? 'bg-status-success/20 text-status-success'
                  : 'bg-surface text-text-secondary border border-border-subtle'
              }`}
            >
              {config.is_open ? '● Sedang Dibuka' : '○ Sedang Ditutup'}
            </span>
          )}
        </div>

        <form onSubmit={handleSaveConfig} className="space-y-4 pt-2 border-t border-border-subtle">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <label htmlFor="input-conversion-rate" className="text-xs font-bold text-text-secondary">
                Nilai 1 Koin (Rupiah):
              </label>
              <input
                id="input-conversion-rate"
                type="number"
                min={1}
                required
                value={conversionRateInput}
                onChange={(e) => setConversionRateInput(e.target.value)}
                className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
              />
              <p className="text-[11px] text-text-secondary">
                Contoh: <strong>100</strong> (1 Koin = Rp 100)
              </p>
            </div>

            <div className="space-y-1.5">
              <label htmlFor="input-target-rupiah" className="text-xs font-bold text-text-secondary">
                Target Penghasilan Normal (Rupiah):
              </label>
              <input
                id="input-target-rupiah"
                type="number"
                min={0}
                required
                value={targetRupiahInput}
                onChange={(e) => setTargetRupiahInput(e.target.value)}
                className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
              />
              <p className="text-[11px] text-text-secondary">
                Hasil kalkulasi koin: <strong>{targetCoinsCalc} Koin</strong>
              </p>
            </div>

            <div className="space-y-1.5">
              <label htmlFor="input-max-payout" className="text-xs font-bold text-text-secondary">
                Batas Maksimum Pencairan (Koin):
              </label>
              <input
                id="input-max-payout"
                type="number"
                min={1}
                required
                value={maxPayoutInput}
                onChange={(e) => setMaxPayoutInput(e.target.value)}
                className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
              />
              <p className="text-[11px] text-text-secondary">
                Batas keras (cap) pencairan per periode
              </p>
            </div>

            <div className="space-y-1.5">
              <label htmlFor="input-payout-day" className="text-xs font-bold text-text-secondary">
                Tanggal Gajian / Payday (1–31):
              </label>
              <input
                id="input-payout-day"
                type="number"
                min={1}
                max={31}
                required
                value={payoutDayInput}
                onChange={(e) => setPayoutDayInput(e.target.value)}
                className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
              />
              <p className="text-[11px] text-text-secondary">
                Hari kalender perayaan gajian
              </p>
            </div>

            <div className="space-y-1.5">
              <label htmlFor="input-earning-period" className="text-xs font-bold text-text-secondary">
                Durasi Periode Earning (Hari):
              </label>
              <input
                id="input-earning-period"
                type="number"
                min={1}
                max={365}
                required
                value={earningPeriodInput}
                onChange={(e) => setEarningPeriodInput(e.target.value)}
                className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
              />
              <p className="text-[11px] text-text-secondary">
                Contoh: <strong>30</strong> (30 hari kerja)
              </p>
            </div>

            <div className="space-y-1.5">
              <label htmlFor="input-timezone" className="text-xs font-bold text-text-secondary">
                Timezone:
              </label>
              <input
                id="input-timezone"
                type="text"
                required
                value={timezoneInput}
                onChange={(e) => setTimezoneInput(e.target.value)}
                className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic"
              />
              <p className="text-[11px] text-text-secondary">
                Contoh: <strong>Asia/Jakarta</strong>
              </p>
            </div>

            <div className="space-y-1.5">
              <label htmlFor="input-start-day" className="text-xs font-bold text-text-secondary">
                Tanggal Mulai Penukaran (1–31):
              </label>
              <input
                id="input-start-day"
                type="number"
                min={1}
                max={31}
                required
                value={startDayInput}
                onChange={(e) => setStartDayInput(e.target.value)}
                className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
              />
            </div>

            <div className="space-y-1.5">
              <label htmlFor="input-end-day" className="text-xs font-bold text-text-secondary">
                Tanggal Akhir Penukaran (1–31):
              </label>
              <input
                id="input-end-day"
                type="number"
                min={1}
                max={31}
                required
                value={endDayInput}
                onChange={(e) => setEndDayInput(e.target.value)}
                className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
              />
            </div>
          </div>

          <div className="p-3 rounded-xl bg-surface border border-border-subtle text-xs">
            <span className="text-text-secondary font-medium">Pratinjau:</span>
            <span className="font-bold text-text-primary ml-1">
              Target Rp {targetRpNum.toLocaleString('id-ID')} ({targetCoinsCalc} koin) • Maks{' '}
              {maxPayoutInput} koin • Gajian tgl {payoutDayInput} • Penukaran tgl {startDayInput}–
              {endDayInput} • {earningPeriodInput} hari ({timezoneInput})
            </span>
          </div>

          {errorMsg && (
            <div className="p-3.5 rounded-xl bg-status-error/15 border border-status-error/30 text-status-error text-xs flex items-center gap-2">
              <AlertCircle className="w-4 h-4 shrink-0" />
              <span>{errorMsg}</span>
            </div>
          )}

          {successMsg && (
            <div className="p-3.5 rounded-xl bg-status-success/15 border border-status-success/30 text-status-success text-xs flex items-center gap-2">
              <Check className="w-4 h-4 shrink-0" />
              <span>{successMsg}</span>
            </div>
          )}

          <button
            type="submit"
            disabled={isSaving}
            className="w-full py-3.5 rounded-2xl bg-accent-magic hover:brightness-110 active:scale-[0.98] text-white font-heading font-bold text-sm shadow-md shadow-accent-magic/30 transition-all flex items-center justify-center gap-2 disabled:opacity-50"
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
