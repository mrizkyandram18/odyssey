import React, { useState, useMemo } from 'react'
import { Sliders, AlertCircle, Check, Calendar, Coins, ArrowRight, Shield, Award, Plus, Trash2, ChevronDown, ChevronUp } from 'lucide-react'
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
    autoBlockInput,
    setAutoBlockInput,
    monthlyTargetInput,
    setMonthlyTargetInput,
    monthlyCapInput,
    setMonthlyCapInput,
    maxCapCeilingInput,
    setMaxCapCeilingInput,
    levelCapBonusInput,
    setLevelCapBonusInput,
    handleSaveConfig,
  } = useAdminConfig()

  // State for visual Level Cap Bonus builder
  const [newTierLevel, setNewTierLevel] = useState('')
  const [newTierBonus, setNewTierBonus] = useState('')
  const [tierError, setTierError] = useState<string | null>(null)
  const [showAdvanced, setShowAdvanced] = useState(false)

  const convRateNum = parseInt(conversionRateInput, 10) || 0
  const targetRpNum = parseInt(targetRupiahInput, 10) || 0
  const targetCoinsCalc = convRateNum > 0 ? Math.floor(targetRpNum / convRateNum) : 0
  const baseCapNum = parseInt(monthlyCapInput, 10) || 0
  const ceilingNum = parseInt(maxCapCeilingInput, 10) || 0

  // Parse structured tiers from levelCapBonusInput
  const parsedTiers = useMemo(() => {
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

  const handleAddTier = (e?: React.MouseEvent) => {
    if (e) e.preventDefault()
    setTierError(null)
    const lvl = parseInt(newTierLevel, 10)
    const bns = parseInt(newTierBonus, 10)

    if (isNaN(lvl) || lvl < 1) {
      setTierError('Tingkat harus berupa angka >= 1')
      return
    }
    if (isNaN(bns) || bns < 0) {
      setTierError('Tambahan bonus koin harus berupa angka >= 0')
      return
    }

    const currentMap: Record<string, number> = {}
    parsedTiers.forEach((t) => {
      currentMap[String(t.level)] = t.bonus
    })
    currentMap[String(lvl)] = bns

    // Update underlying JSON string
    setLevelCapBonusInput(JSON.stringify(currentMap))
    setNewTierLevel('')
    setNewTierBonus('')
  }

  const handleRemoveTier = (levelToRemove: number) => {
    const currentMap: Record<string, number> = {}
    parsedTiers.forEach((t) => {
      if (t.level !== levelToRemove) {
        currentMap[String(t.level)] = t.bonus
      }
    })
    setLevelCapBonusInput(JSON.stringify(currentMap))
  }

  return (
    <div className="space-y-4">
      <div className="p-4 sm:p-5 rounded-2xl bg-surface border border-border-subtle shadow-xs space-y-4">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2.5 pb-3 border-b border-border-subtle/60">
          <div>
            <h3 className="font-heading font-bold text-text-primary text-sm sm:text-base flex items-center gap-2">
              <Sliders className="w-4 h-4 text-accent-magic" />
              <span>Pengaturan Ekonomi & Batas Koin Bulanan</span>
            </h3>
            <p className="text-[11px] text-text-secondary mt-0.5">
              Kelola batas koin bulanan, bonus berdasarkan Tingkat, nilai konversi koin, dan jadwal pencairan.
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
              {config.is_open ? 'Pencairan Dibuka' : 'Pencairan Ditutup'}
            </span>
          )}
        </div>

        <form onSubmit={handleSaveConfig} className="space-y-6">
          {/* Dynamic Formula & Visual Cap Breakdown */}
          <div className="p-4 rounded-2xl bg-surface-elevated/60 border border-border-subtle shadow-xs space-y-3">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-1 border-b border-border-subtle/60 pb-2">
              <div className="flex items-center gap-2">
                <Shield className="w-4 h-4 text-accent-magic" />
                <h4 className="text-xs font-bold uppercase tracking-wider text-text-primary">
                  Rumus Batas Perolehan Koin Bulanan
                </h4>
              </div>
              <span className="text-[10px] text-text-secondary">
                Siklus Kalender: Tgl 1 – Akhir Bulan
              </span>
            </div>

            {/* Formula Cards */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-2.5 items-stretch">
              <div className="p-3 rounded-xl bg-surface border border-border-subtle flex flex-col justify-between">
                <span className="text-[11px] font-bold text-text-secondary uppercase">1. Batas Standar</span>
                <p className="text-base font-bold text-text-primary font-mono mt-1">
                  {baseCapNum > 0 ? `${baseCapNum.toLocaleString('id-ID')} koin` : 'Tanpa Batas'}
                </p>
                <span className="text-[10px] text-text-secondary mt-0.5">Batas dasar untuk semua anggota</span>
              </div>

              <div className="p-3 rounded-xl bg-surface border border-border-subtle flex flex-col justify-between">
                <span className="text-[11px] font-bold text-text-secondary uppercase">2. Bonus Tingkat</span>
                <p className="text-base font-bold text-accent-gold font-mono mt-1">
                  {parsedTiers.length > 0 ? `+${parsedTiers.map((t) => t.bonus.toLocaleString('id-ID')).join(' / +')} koin` : 'Belum diatur'}
                </p>
                <span className="text-[10px] text-text-secondary mt-0.5">Otomatis aktif saat Tingkat tercapai</span>
              </div>

              <div className="p-3 rounded-xl bg-surface border border-border-subtle flex flex-col justify-between">
                <span className="text-[11px] font-bold text-text-secondary uppercase">3. Plafon Maksimum</span>
                <p className="text-base font-bold text-accent-reward font-mono mt-1">
                  {ceilingNum > 0 ? `${ceilingNum.toLocaleString('id-ID')} koin` : 'Tanpa Plafon'}
                </p>
                <span className="text-[10px] text-text-secondary mt-0.5">Batas absolut tertinggi</span>
              </div>
            </div>

            {/* Dynamic Simulation Preview Table based strictly on actual config */}
            {baseCapNum > 0 && parsedTiers.length > 0 && (
              <div className="p-3 rounded-xl bg-surface border border-border-subtle/80 space-y-2">
                <p className="text-[11px] font-bold text-text-secondary">Simulasi Batas Koin Nyata Berdasarkan Tingkat:</p>
                <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-2">
                  <div className="p-2 rounded-lg bg-surface-elevated border border-border-subtle/60 text-xs">
                    <p className="text-[10px] text-text-secondary font-bold">
                      Tingkat Awal {parsedTiers[0].level > 1 ? `(1–${parsedTiers[0].level - 1})` : ''}:
                    </p>
                    <p className="font-bold text-text-primary font-mono">{baseCapNum.toLocaleString('id-ID')} koin</p>
                  </div>
                  {parsedTiers.map((t, idx) => {
                    const raw = baseCapNum + t.bonus
                    const capped = ceilingNum > 0 && raw > ceilingNum ? ceilingNum : raw
                    const isClamped = ceilingNum > 0 && raw > ceilingNum
                    const nextTier = parsedTiers[idx + 1]
                    const levelLabel = nextTier ? `Tingkat ${t.level}–${nextTier.level - 1}` : `Tingkat ${t.level}+`
                    return (
                      <div key={t.level} className="p-2 rounded-lg bg-surface-elevated border border-border-subtle/60 text-xs">
                        <p className="text-[10px] text-text-secondary font-bold">{levelLabel}:</p>
                        <p className="font-bold text-text-primary font-mono">
                          {capped.toLocaleString('id-ID')} koin {isClamped ? <span className="text-[9px] text-accent-reward font-normal">(mentok plafon)</span> : ''}
                        </p>
                      </div>
                    )
                  })}
                </div>
              </div>
            )}
          </div>

          {/* Group 1: Batas Koin Bulanan & Bonus Tingkat */}
          <div className="space-y-4 p-4 rounded-xl bg-surface-elevated/40 border border-border-subtle">
            <div className="flex items-center justify-between">
              <h4 className="text-xs font-bold uppercase tracking-wider text-text-primary flex items-center gap-2">
                <Shield className="w-4 h-4 text-accent-magic" />
                <span>Input Batas Koin Bulanan & Bonus Tingkat</span>
              </h4>
              <span className="text-[10px] text-text-secondary font-medium">Bulan Kalender (Tgl 1 – Akhir Bulan)</span>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div className="space-y-1">
                <label htmlFor="input-monthly-cap" className="text-xs font-bold text-text-secondary flex items-center justify-between">
                  <span>Batas Koin Bulanan Standar <span className="text-status-error">*</span></span>
                  <span className="text-[10px] font-normal text-text-secondary">0 = tanpa batas</span>
                </label>
                <div className="relative">
                  <input
                    id="input-monthly-cap"
                    type="number"
                    min={0}
                    required
                    value={monthlyCapInput}
                    onChange={(e) => setMonthlyCapInput(e.target.value)}
                    className="w-full p-2.5 pr-24 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
                  />
                  <span className="absolute right-3 top-2.5 text-xs text-text-secondary">koin / bulan</span>
                </div>
                <div className="flex items-center justify-between text-[10px] text-text-secondary">
                  <span>Jumlah maksimum koin per bulan kalender.</span>
                  {baseCapNum > 0 && (
                    <span className="font-bold text-accent-magic">
                      Format: {baseCapNum.toLocaleString('id-ID')} koin
                    </span>
                  )}
                </div>
              </div>

              <div className="space-y-1">
                <label htmlFor="input-max-ceiling" className="text-xs font-bold text-text-secondary flex items-center justify-between">
                  <span>Plafon Maksimum (Ceiling)</span>
                  <span className="text-[10px] font-normal text-text-secondary">0 = tanpa plafon</span>
                </label>
                <div className="relative">
                  <input
                    id="input-max-ceiling"
                    type="number"
                    min={0}
                    placeholder="10000"
                    value={maxCapCeilingInput}
                    onChange={(e) => setMaxCapCeilingInput(e.target.value)}
                    className="w-full p-2.5 pr-24 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
                  />
                  <span className="absolute right-3 top-2.5 text-xs text-text-secondary">koin / bulan</span>
                </div>
                <div className="flex items-center justify-between text-[10px] text-text-secondary">
                  <span>Batas absolut tertinggi koin bulanan anggota.</span>
                  {ceilingNum > 0 && (
                    <span className="font-bold text-accent-reward">
                      Format: {ceilingNum.toLocaleString('id-ID')} koin
                    </span>
                  )}
                </div>
              </div>
            </div>

            {/* Visual Level Cap Bonus Builder */}
            <div className="pt-2 border-t border-border-subtle/80 space-y-3">
              <div className="flex items-center justify-between">
                <div>
                  <h5 className="text-xs font-bold text-text-primary flex items-center gap-1.5">
                    <Award className="w-3.5 h-3.5 text-accent-gold" />
                    <span>Bonus Batas Koin Berdasarkan Tingkat</span>
                  </h5>
                  <p className="text-[11px] text-text-secondary mt-0.5">
                    Tambahan batas koin bulanan yang otomatis aktif saat anggota mencapai Tingkat tertentu.
                  </p>
                </div>
              </div>

              {/* Tiers List */}
              {parsedTiers.length === 0 ? (
                <div className="p-3 text-center text-text-secondary text-xs bg-surface rounded-xl border border-dashed border-border-subtle">
                  Belum ada bonus tingkat yang diatur. Semua tingkat akan menggunakan batas standar.
                </div>
              ) : (
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2">
                  {parsedTiers.map((tier) => {
                    const effectivePreview = baseCapNum > 0
                      ? (ceilingNum > 0 && baseCapNum + tier.bonus > ceilingNum ? ceilingNum : baseCapNum + tier.bonus)
                      : 'Bebas'
                    return (
                      <div
                        key={tier.level}
                        className="p-2.5 rounded-xl bg-surface border border-border-subtle flex items-center justify-between gap-2 shadow-2xs"
                      >
                        <div>
                          <div className="flex items-center gap-1.5">
                            <span className="text-xs font-bold text-text-primary">Tingkat {tier.level}</span>
                            <span className="text-[10px] font-bold px-1.5 py-0.5 rounded-md bg-accent-gold/15 text-accent-gold">
                              +{tier.bonus.toLocaleString('id-ID')} Koin
                            </span>
                          </div>
                          <p className="text-[10px] text-text-secondary mt-0.5">
                            Total Batas: {effectivePreview.toLocaleString('id-ID')} koin
                          </p>
                        </div>

                        <button
                          type="button"
                          onClick={() => handleRemoveTier(tier.level)}
                          className="p-1.5 rounded-lg text-text-secondary hover:text-status-error hover:bg-status-error/10 transition-colors cursor-pointer"
                          title={`Hapus bonus tingkat ${tier.level}`}
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </button>
                      </div>
                    )
                  })}
                </div>
              )}

              {/* Add Tier Inline Form */}
              <div className="p-3 rounded-xl bg-surface border border-border-subtle space-y-2">
                <span className="text-[11px] font-bold text-text-secondary block">Tambah Bonus Tingkat Baru:</span>
                <div className="flex flex-col sm:flex-row items-center gap-2">
                  <div className="w-full sm:w-1/3">
                    <input
                      type="number"
                      min={1}
                      placeholder="Tingkat (misal: 15)"
                      value={newTierLevel}
                      onChange={(e) => setNewTierLevel(e.target.value)}
                      className="w-full p-2 rounded-lg bg-surface-elevated border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic font-mono"
                    />
                  </div>
                  <div className="w-full sm:w-1/2">
                    <input
                      type="number"
                      min={0}
                      placeholder="Tambahan Koin (misal: 750)"
                      value={newTierBonus}
                      onChange={(e) => setNewTierBonus(e.target.value)}
                      className="w-full p-2 rounded-lg bg-surface-elevated border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic font-mono"
                    />
                  </div>
                  <button
                    type="button"
                    onClick={handleAddTier}
                    className="w-full sm:w-auto px-4 py-2 rounded-lg bg-accent-magic text-white text-xs font-bold hover:brightness-110 active:scale-[0.99] transition-all flex items-center justify-center gap-1 cursor-pointer shrink-0"
                  >
                    <Plus className="w-3.5 h-3.5" />
                    <span>Tambah</span>
                  </button>
                </div>
                {tierError && <p className="text-[10px] text-status-error">{tierError}</p>}
              </div>

              {/* Technical JSON storage (kept for backwards-compatibility and tests) */}
              <input
                id="input-level-bonus"
                type="hidden"
                value={levelCapBonusInput}
                onChange={(e) => setLevelCapBonusInput(e.target.value)}
              />
            </div>
          </div>

          {/* Group 2: Target & Nilai Konversi Koin */}
          <div className="space-y-3 pt-2">
            <h4 className="text-xs font-bold uppercase tracking-wider text-text-secondary flex items-center gap-1.5">
              <Coins className="w-3.5 h-3.5 text-accent-gold" />
              <span>Target & Nilai Konversi Koin</span>
            </h4>

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <div className="space-y-1">
                <label htmlFor="input-conversion-rate" className="text-xs font-bold text-text-secondary flex items-center justify-between">
                  <span>Nilai 1 Koin (Rupiah) <span className="text-status-error">*</span></span>
                  {convRateNum > 0 && (
                    <span className="text-[10px] font-bold text-accent-gold">
                      1 Koin = Rp {convRateNum.toLocaleString('id-ID')}
                    </span>
                  )}
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
                <p className="text-[10px] text-text-secondary">Nilai rupiah untuk setiap 1 koin yang ditukarkan.</p>
              </div>

              <div className="space-y-1">
                <label htmlFor="input-target-rupiah" className="text-xs font-bold text-text-secondary flex items-center justify-between">
                  <span>Target Normal (Rupiah) <span className="text-status-error">*</span></span>
                  {targetRpNum > 0 && (
                    <span className="text-[10px] font-bold text-accent-magic">
                      Rp {targetRpNum.toLocaleString('id-ID')}
                    </span>
                  )}
                </label>
                <input
                  id="input-target-rupiah"
                  type="number"
                  min={0}
                  required
                  value={targetRupiahInput}
                  onChange={(e) => setTargetRupiahInput(e.target.value)}
                  className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs sm:text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
                />
                <p className="text-[10px] text-accent-magic font-bold">
                  Setara: {targetCoinsCalc.toLocaleString('id-ID')} Koin
                </p>
              </div>

              <div className="space-y-1">
                <label htmlFor="input-monthly-target" className="text-xs font-bold text-text-secondary flex items-center justify-between">
                  <span>Target Koin Awal Anggota <span className="text-status-error">*</span></span>
                  <span className="text-[10px] text-text-secondary">0 = fleksibel</span>
                </label>
                <input
                  id="input-monthly-target"
                  type="number"
                  min={0}
                  max={10000}
                  required
                  value={monthlyTargetInput}
                  onChange={(e) => setMonthlyTargetInput(e.target.value)}
                  className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs sm:text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
                />
                <p className="text-[10px] text-text-secondary">Target default anggota baru (0–10.000 koin).</p>
              </div>
            </div>
          </div>

          {/* Group 3: Jadwal & Periode Penukaran Koin */}
          <div className="space-y-3 pt-3 border-t border-border-subtle">
            <div>
              <h4 className="text-xs font-bold uppercase tracking-wider text-text-secondary flex items-center gap-1.5">
                <Calendar className="w-3.5 h-3.5 text-accent-magic" />
                <span>Pengaturan Periode Penukaran Koin</span>
              </h4>
              <p className="text-[11px] text-text-secondary mt-0.5">
                Jadwal kapan anggota dapat mengajukan pencairan saldo koin menjadi uang tunai/saldo e-wallet.
              </p>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-4 gap-3">
              <div className="space-y-1">
                <label htmlFor="input-start-day" className="text-xs font-bold text-text-secondary">
                  Tanggal Buka Pengajuan <span className="text-status-error">*</span>
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
                <p className="text-[10px] text-text-secondary">Tgl 1–31 buka pengajuan</p>
              </div>

              <div className="space-y-1">
                <label htmlFor="input-end-day" className="text-xs font-bold text-text-secondary">
                  Tanggal Tutup Pengajuan <span className="text-status-error">*</span>
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
                <p className="text-[10px] text-text-secondary">Batas akhir ajukan klaim</p>
              </div>

              <div className="space-y-1">
                <label htmlFor="input-payout-day" className="text-xs font-bold text-text-secondary">
                  Tanggal Gajian / Transfer <span className="text-status-error">*</span>
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
                <p className="text-[10px] text-text-secondary">Hari transfer dana reward</p>
              </div>

              <div className="space-y-1">
                <label htmlFor="input-max-payout" className="text-xs font-bold text-text-secondary">
                  Maksimal Penarikan (Koin) <span className="text-status-error">*</span>
                </label>
                <input
                  id="input-max-payout"
                  type="number"
                  min={1}
                  required
                  value={maxPayoutInput}
                  onChange={(e) => setMaxPayoutInput(e.target.value)}
                  className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs sm:text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
                />
                <p className="text-[10px] text-text-secondary">Batas tarik per transaksi</p>
              </div>
            </div>
          </div>

          {/* Group 4: Keamanan & Pengaturan Lanjutan */}
          <div className="space-y-3 pt-3 border-t border-border-subtle">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 items-start">
              <div className="space-y-1">
                <label htmlFor="input-auto-block" className="text-xs font-bold text-text-secondary">
                  Blokir Otomatis Setelah Inaktif (Hari) <span className="text-status-error">*</span>
                </label>
                <input
                  id="input-auto-block"
                  type="number"
                  min={0}
                  max={365}
                  required
                  value={autoBlockInput}
                  onChange={(e) => setAutoBlockInput(e.target.value)}
                  className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs sm:text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
                />
                <p className="text-[10px] text-text-secondary">0 = nonaktif (otomatis blokir jika inaktif)</p>
              </div>

              {/* Advanced Technical Settings Dropdown */}
              <div className="space-y-1">
                <button
                  type="button"
                  onClick={() => setShowAdvanced(!showAdvanced)}
                  className="text-xs font-bold text-text-secondary hover:text-text-primary flex items-center gap-1.5 pt-2 cursor-pointer transition-colors"
                >
                  <span>Parameter Sistem Lanjutan</span>
                  {showAdvanced ? <ChevronUp className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}
                </button>
                {showAdvanced ? (
                  <div className="grid grid-cols-2 gap-2 pt-1">
                    <div className="space-y-1">
                      <label htmlFor="input-earning-period" className="text-[10px] font-bold text-text-secondary">
                        Siklus Earning (Hari)
                      </label>
                      <input
                        id="input-earning-period"
                        type="number"
                        min={1}
                        max={365}
                        required
                        value={earningPeriodInput}
                        onChange={(e) => setEarningPeriodInput(e.target.value)}
                        className="w-full p-2 rounded-lg bg-surface-elevated border border-border-subtle text-xs font-mono"
                      />
                    </div>
                    <div className="space-y-1">
                      <label htmlFor="input-timezone" className="text-[10px] font-bold text-text-secondary">
                        Zona Waktu Sistem
                      </label>
                      <input
                        id="input-timezone"
                        type="text"
                        required
                        value={timezoneInput}
                        onChange={(e) => setTimezoneInput(e.target.value)}
                        className="w-full p-2 rounded-lg bg-surface-elevated border border-border-subtle text-xs font-mono"
                      />
                    </div>
                  </div>
                ) : (
                  <div className="hidden">
                    <input
                      id="input-earning-period"
                      type="hidden"
                      value={earningPeriodInput}
                      onChange={(e) => setEarningPeriodInput(e.target.value)}
                    />
                    <input
                      id="input-timezone"
                      type="hidden"
                      value={timezoneInput}
                      onChange={(e) => setTimezoneInput(e.target.value)}
                    />
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Compact Live KPI Preview */}
          <div className="p-3 rounded-xl bg-surface-elevated border border-border-subtle text-xs flex flex-col sm:flex-row sm:items-center justify-between gap-2">
            <div className="flex items-center gap-1.5 text-text-secondary font-bold text-[11px] uppercase tracking-wider">
              <span>Ringkasan Parameter Aktif</span>
              <ArrowRight className="w-3 h-3 text-accent-magic" />
            </div>
            <div className="font-bold text-text-primary text-[11px] sm:text-xs">
              Target Rp {targetRpNum.toLocaleString('id-ID')} ({targetCoinsCalc} koin) • Maks{' '}
              {maxPayoutInput} koin • Cap {monthlyCapInput || 'unlimited'} (Ceiling {maxCapCeilingInput || 'none'}) • Gajian tgl {payoutDayInput}
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
                <span>Simpan Pengaturan Periode & Ekonomi</span>
              </>
            )}
          </button>
        </form>
      </div>

      {/* Information Box linking to Hadiah Tab */}
      <div className="p-4 rounded-2xl bg-surface border border-border-subtle shadow-2xs flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-xl bg-accent-gold/10 text-accent-gold flex items-center justify-center shrink-0">
            <Award className="w-5 h-5" />
          </div>
          <div>
            <h4 className="text-xs font-bold text-text-primary">Katalog Hadiah & Koleksi</h4>
            <p className="text-[11px] text-text-secondary">
              Kelola item bingkai avatar dan efek visual kosmetik di tab khusus <strong>Hadiah</strong> pada navigasi panel admin.
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
