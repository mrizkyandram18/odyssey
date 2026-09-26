import React, { useMemo } from 'react'
import { AlertCircle, Check, Award } from 'lucide-react'
import { useAdminConfig } from '../../hooks/useAdminConfig'
import {
  EconomyHeader,
  CapFormulaHero,
  MonthlyCapSection,
  ConversionTargetSection,
  RedemptionScheduleSection,
  SecurityAdvancedSection,
  KpiPreviewStrip,
  AnnouncementSection,
  type ParsedTier,
} from './EconomySettingsSections'

export const EconomySettingsForm: React.FC = () => {
  const settings = useAdminConfig()
  const {
    config,
    isSaving,
    successMsg,
    errorMsg,
    conversionRateInput,
    targetRupiahInput,
    maxPayoutInput,
    monthlyCapInput,
    maxCapCeilingInput,
    payoutDayInput,
    levelCapBonusInput,
    handleSaveSchedule,
    handleSaveEconomy,
    handleSaveTechnical,
    handleSaveAnnouncement,
  } = settings

  const convRateNum = parseInt(conversionRateInput, 10) || 0
  const targetRpNum = parseInt(targetRupiahInput, 10) || 0
  const targetCoinsCalc = convRateNum > 0 ? Math.floor(targetRpNum / convRateNum) : 0
  const baseCapNum = parseInt(monthlyCapInput, 10) || 0
  const ceilingNum = parseInt(maxCapCeilingInput, 10) || 0

  // Parse structured tiers from levelCapBonusInput (display only; writes go
  // through setLevelCapBonusInput in MonthlyCapSection).
  const parsedTiers: ParsedTier[] = useMemo(() => {
    try {
      if (!levelCapBonusInput.trim()) return []
      const obj = JSON.parse(levelCapBonusInput)
      if (typeof obj !== 'object' || obj === null || Array.isArray(obj)) return []
      return Object.entries(obj)
        .map(([k, v]) => ({ level: parseInt(k, 10), bonus: Number(v) }))
        .filter((t) => !isNaN(t.level) && t.level > 0 && !isNaN(t.bonus) && t.bonus >= 0)
        .sort((a, b) => a.level - b.level)
    } catch {
      return []
    }
  }, [levelCapBonusInput])

  const sectionSaveButton = (label: string) => (
    <button
      type="submit"
      disabled={isSaving}
      className="w-full py-2.5 rounded-xl bg-accent-magic hover:brightness-110 active:scale-[0.99] text-white font-bold text-xs sm:text-sm shadow-xs transition-all flex items-center justify-center gap-2 disabled:opacity-50 cursor-pointer"
    >
      {isSaving ? (
        <span>Menyimpan...</span>
      ) : (
        <>
          <Check className="w-4 h-4" />
          <span>{label}</span>
        </>
      )}
    </button>
  )

  return (
    <div className="space-y-4">
      <div className="p-4 sm:p-5 rounded-2xl bg-surface border border-border-subtle shadow-xs space-y-4">
        <EconomyHeader isOpen={config ? config.is_open : null} />

        {/* Shared alerts: each independent save reports here. */}
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

        {/* Form 1 — Aturan pencairan (independent save). */}
        <form
          onSubmit={(e) => {
            e.preventDefault()
            void handleSaveSchedule()
          }}
          className="space-y-4 p-4 rounded-2xl bg-surface-elevated/40 border border-border-subtle"
        >
          <RedemptionScheduleSection settings={settings} />
          {sectionSaveButton('Simpan Aturan Pencairan')}
        </form>

        {/* Form 2 — Target & batas koin (independent save). */}
        <form
          onSubmit={(e) => {
            e.preventDefault()
            void handleSaveEconomy()
          }}
          className="space-y-4 p-4 rounded-2xl bg-surface-elevated/40 border border-border-subtle"
        >
          <CapFormulaHero baseCapNum={baseCapNum} ceilingNum={ceilingNum} parsedTiers={parsedTiers} />

          <MonthlyCapSection
            settings={settings}
            baseCapNum={baseCapNum}
            ceilingNum={ceilingNum}
            parsedTiers={parsedTiers}
          />

          <ConversionTargetSection
            settings={settings}
            convRateNum={convRateNum}
            targetRpNum={targetRpNum}
            targetCoinsCalc={targetCoinsCalc}
          />

          <KpiPreviewStrip
            targetRpNum={targetRpNum}
            targetCoinsCalc={targetCoinsCalc}
            maxPayoutInput={maxPayoutInput}
            monthlyCapInput={monthlyCapInput}
            maxCapCeilingInput={maxCapCeilingInput}
            payoutDayInput={payoutDayInput}
          />
          {sectionSaveButton('Simpan Target & Batas Koin')}
        </form>

        {/* Form 3 — Teknis (independent save, rarely touched). */}
        <form
          onSubmit={(e) => {
            e.preventDefault()
            void handleSaveTechnical()
          }}
          className="space-y-4 p-4 rounded-2xl bg-surface-elevated/40 border border-border-subtle"
        >
          <SecurityAdvancedSection settings={settings} />
          {sectionSaveButton('Simpan Pengaturan Teknis')}
        </form>

        {/* Form 4 — Pengumuman (independent save). */}
        <form
          onSubmit={(e) => {
            e.preventDefault()
            void handleSaveAnnouncement()
          }}
          className="space-y-4 p-4 rounded-2xl bg-surface-elevated/40 border border-border-subtle"
        >
          <AnnouncementSection settings={settings} />
          {sectionSaveButton('Simpan Pengumuman')}
        </form>
      </div>

      {/* Information Box linking to Kustomisasi Tab */}
      <div className="p-4 rounded-2xl bg-surface border border-border-subtle shadow-2xs flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-xl bg-accent-gold/10 text-accent-gold flex items-center justify-center shrink-0">
            <Award className="w-5 h-5" />
          </div>
          <div>
            <h4 className="text-xs font-bold text-text-primary">Katalog Kustomisasi</h4>
            <p className="text-[11px] text-text-secondary">
              Kelola item bingkai avatar dan efek visual kosmetik di tab khusus <strong>Kustomisasi</strong> pada navigasi panel admin.
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
