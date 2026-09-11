import { useState, useCallback, useEffect } from 'react'
import { adminTasksApi } from '../../../shared/lib/api'
import type { RedemptionConfig } from '../../../shared/types'

export function useAdminConfig() {
  const [config, setConfig] = useState<RedemptionConfig | null>(null)
  const [isFetching, setIsFetching] = useState(false)
  const [isSaving, setIsSaving] = useState(false)
  const [successMsg, setSuccessMsg] = useState<string | null>(null)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  // Input states
  const [startDayInput, setStartDayInput] = useState('21')
  const [endDayInput, setEndDayInput] = useState('26')
  const [payoutDayInput, setPayoutDayInput] = useState('28')
  const [earningPeriodInput, setEarningPeriodInput] = useState('30')
  const [conversionRateInput, setConversionRateInput] = useState('100')
  const [targetRupiahInput, setTargetRupiahInput] = useState('320000')
  const [maxPayoutInput, setMaxPayoutInput] = useState('3200')
  const [timezoneInput, setTimezoneInput] = useState('Asia/Jakarta')
  const [autoBlockInput, setAutoBlockInput] = useState('5')
  const [monthlyTargetInput, setMonthlyTargetInput] = useState('0')
  const [monthlyCapInput, setMonthlyCapInput] = useState('0')
  const [maxCapCeilingInput, setMaxCapCeilingInput] = useState('')
  const [levelCapBonusInput, setLevelCapBonusInput] = useState('')
  // Announcement (existing odyssey_system_config announcement_* keys).
  const [announcementEnabledInput, setAnnouncementEnabledInput] = useState(false)
  const [announcementTitleInput, setAnnouncementTitleInput] = useState('')
  const [announcementBodyInput, setAnnouncementBodyInput] = useState('')
  const [announcementAudienceInput, setAnnouncementAudienceInput] = useState('ALL')
  const [announcementStartAtInput, setAnnouncementStartAtInput] = useState('')
  const [announcementEndAtInput, setAnnouncementEndAtInput] = useState('')
  const [announcementPriorityInput, setAnnouncementPriorityInput] = useState('normal')

  const fetchConfig = useCallback(async () => {
    setIsFetching(true)
    setErrorMsg(null)
    try {
      const res = await adminTasksApi.getConfig()
      setConfig(res)
      if (res) {
        setStartDayInput(String(res.redemption_start_day ?? 21))
        setEndDayInput(String(res.redemption_end_day ?? 26))
        setPayoutDayInput(String(res.payout_day ?? 28))
        setEarningPeriodInput(String(res.earning_period_days ?? 30))
        setConversionRateInput(String(res.conversion_rate ?? 100))
        setTargetRupiahInput(String(res.payout_target_rupiah ?? 320000))
        setMaxPayoutInput(String(res.max_payout_coins ?? 3200))
        setTimezoneInput(res.timezone || 'Asia/Jakarta')
        setAutoBlockInput(String(res.auto_block_inactivity_days ?? 5))
        setMonthlyTargetInput(String(res.default_monthly_coin_target ?? 0))
        setMonthlyCapInput(res.default_monthly_earning_cap != null && res.default_monthly_earning_cap >= 0 ? String(res.default_monthly_earning_cap) : '0')
        setMaxCapCeilingInput(res.max_monthly_earning_cap_ceiling != null ? String(res.max_monthly_earning_cap_ceiling) : '')
        setLevelCapBonusInput(res.level_cap_bonus ? JSON.stringify(res.level_cap_bonus) : '')
        const ann = (res as any).announcement
        if (ann && typeof ann === 'object') {
          setAnnouncementEnabledInput(Boolean(ann.enabled))
          setAnnouncementTitleInput(String(ann.title ?? ''))
          setAnnouncementBodyInput(String(ann.body ?? ''))
          setAnnouncementAudienceInput(String(ann.audience ?? 'ALL'))
          setAnnouncementStartAtInput(String(ann.start_at ?? ''))
          setAnnouncementEndAtInput(String(ann.end_at ?? ''))
          setAnnouncementPriorityInput(String(ann.priority ?? 'normal'))
        }
      }
    } catch (err: any) {
      setErrorMsg(err?.message || 'Gagal memuat konfigurasi ekonomi')
    } finally {
      setIsFetching(false)
    }
  }, [])

  useEffect(() => {
    fetchConfig()
  }, [fetchConfig])

  const handleSaveConfig = async (e?: React.FormEvent) => {
    if (e) e.preventDefault()
    setSuccessMsg(null)
    setErrorMsg(null)

    const start = parseInt(startDayInput, 10)
    const end = parseInt(endDayInput, 10)
    const payout = parseInt(payoutDayInput, 10)
    const earningPeriod = parseInt(earningPeriodInput, 10)
    const rate = parseInt(conversionRateInput, 10)
    const targetRp = parseInt(targetRupiahInput, 10)
    const maxPayout = parseInt(maxPayoutInput, 10)
    const autoBlock = parseInt(autoBlockInput, 10)
    const monthlyTarget = parseInt(monthlyTargetInput, 10)
    const monthlyCap = parseInt(monthlyCapInput, 10)
    const ceiling = maxCapCeilingInput.trim() !== '' ? parseInt(maxCapCeilingInput, 10) : undefined

    if (isNaN(start) || start < 1 || start > 31) {
      setErrorMsg('Tanggal mulai harus antara 1 sampai 31')
      return
    }
    if (isNaN(end) || end < 1 || end > 31) {
      setErrorMsg('Tanggal selesai harus antara 1 sampai 31')
      return
    }
    if (start > end) {
      setErrorMsg('Tanggal mulai tidak boleh lebih besar dari tanggal selesai')
      return
    }
    if (isNaN(payout) || payout < 1 || payout > 31) {
      setErrorMsg('Tanggal gajian (payout day) harus antara 1 sampai 31')
      return
    }
    if (isNaN(earningPeriod) || earningPeriod < 1 || earningPeriod > 365) {
      setErrorMsg('Durasi periode earning harus antara 1 sampai 365 hari')
      return
    }
    if (isNaN(rate) || rate <= 0) {
      setErrorMsg('Nilai konversi koin harus lebih dari 0')
      return
    }
    if (isNaN(targetRp) || targetRp <= 0) {
      setErrorMsg('Target rupiah harus lebih dari 0')
      return
    }
    if (isNaN(maxPayout) || maxPayout <= 0) {
      setErrorMsg('Batas penarikan koin maksimum harus lebih dari 0')
      return
    }
    if (isNaN(autoBlock) || autoBlock < 0 || autoBlock > 365) {
      setErrorMsg('Batas inaktivitas auto-block harus antara 0 sampai 365 (0 = nonaktif)')
      return
    }
    if (isNaN(monthlyTarget) || monthlyTarget < 0 || monthlyTarget > 10000) {
      setErrorMsg('Target koin bulanan default harus antara 0 sampai 10000')
      return
    }
    if (isNaN(monthlyCap) || monthlyCap < 0) {
      setErrorMsg('Batas earning bulanan default harus >= 0 (0 = unlimited)')
      return
    }
    if (ceiling !== undefined && (isNaN(ceiling) || ceiling < 0)) {
      setErrorMsg('Plafon earning cap ceiling harus >= 0')
      return
    }
    if (ceiling !== undefined && ceiling > 0 && monthlyCap > ceiling) {
      setErrorMsg(`Batas earning bulanan default (${monthlyCap}) tidak boleh melebihi ceiling (${ceiling})`)
      return
    }

    let parsedBonusMap: Record<string, number> | undefined
    if (levelCapBonusInput.trim() !== '') {
      try {
        const parsed = JSON.parse(levelCapBonusInput)
        if (typeof parsed !== 'object' || parsed === null || Array.isArray(parsed)) {
          throw new Error('Format harus objek JSON')
        }
        for (const [k, v] of Object.entries(parsed)) {
          const numKey = parseInt(k, 10)
          if (isNaN(numKey) || numKey <= 0) {
            throw new Error(`Level threshold "${k}" harus bilangan bulat positif`)
          }
          if (typeof v !== 'number' || v < 0) {
            throw new Error(`Bonus koin untuk level ${k} harus angka >= 0`)
          }
        }
        parsedBonusMap = parsed as Record<string, number>
      } catch (err: any) {
        setErrorMsg(`Format level bonus tidak valid: ${err?.message || 'JSON error'}`)
        return
      }
    }

    // Announcement validation (mirrors backend ValidateAnnouncementInput).
    const annTitle = announcementTitleInput.trim()
    const annBody = announcementBodyInput.trim()
    const annAudience = announcementAudienceInput.trim().toUpperCase() || 'ALL'
    const annPriority = announcementPriorityInput.trim().toLowerCase() || 'normal'
    const annStart = announcementStartAtInput.trim()
    const annEnd = announcementEndAtInput.trim()
    if (!['ALL', 'MEMBER', 'ADMIN'].includes(annAudience)) {
      setErrorMsg('Audiens pengumuman harus ALL, MEMBER, atau ADMIN')
      return
    }
    if (!['low', 'normal', 'high', 'urgent'].includes(annPriority)) {
      setErrorMsg('Prioritas pengumuman harus low, normal, high, atau urgent')
      return
    }
    if (annTitle.length > 255) {
      setErrorMsg('Judul pengumuman maksimal 255 karakter')
      return
    }
    if (annBody.length > 5000) {
      setErrorMsg('Isi pengumuman maksimal 5000 karakter')
      return
    }
    const parseRfc3339 = (s: string): number | null => {
      if (!s) return null
      const t = Date.parse(s)
      return isNaN(t) ? null : t
    }
    const startTs = parseRfc3339(annStart)
    const endTs = parseRfc3339(annEnd)
    if (annStart && startTs === null) {
      setErrorMsg('Waktu mulai pengumuman harus format RFC3339 (mis. 2026-09-12T00:00:00+07:00)')
      return
    }
    if (annEnd && endTs === null) {
      setErrorMsg('Waktu selesai pengumuman harus format RFC3339 (mis. 2026-09-30T23:59:59+07:00)')
      return
    }
    if (startTs !== null && endTs !== null && startTs > endTs) {
      setErrorMsg('Waktu mulai pengumuman tidak boleh setelah waktu selesai')
      return
    }
    if (announcementEnabledInput && !annTitle && !annBody) {
      setErrorMsg('Judul atau isi pengumuman wajib diisi saat pengumuman diaktifkan')
      return
    }

    setIsSaving(true)
    try {
      const updated = await adminTasksApi.updateConfig({
        start_day: start,
        end_day: end,
        payout_day: payout,
        earning_period_days: earningPeriod,
        conversion_rate: rate,
        payout_target_rupiah: targetRp,
        payout_target_coins: Math.round(targetRp / rate),
        max_payout_coins: maxPayout,
        timezone: timezoneInput.trim() || 'Asia/Jakarta',
        auto_block_inactivity_days: autoBlock,
        default_monthly_coin_target: monthlyTarget,
        default_monthly_earning_cap: monthlyCap,
        max_monthly_earning_cap_ceiling: ceiling,
        level_cap_bonus: parsedBonusMap,
        announcement_enabled: announcementEnabledInput,
        announcement_title: annTitle,
        announcement_body: annBody,
        announcement_audience: annAudience,
        announcement_start_at: annStart,
        announcement_end_at: annEnd,
        announcement_priority: annPriority,
      })
      setConfig(updated)
      setSuccessMsg('Konfigurasi ekonomi berhasil disimpan!')
      setTimeout(() => setSuccessMsg(null), 4000)
    } catch (err: any) {
      setErrorMsg(err?.message || 'Gagal menyimpan konfigurasi')
    } finally {
      setIsSaving(false)
    }
  }

  return {
    config,
    isFetching,
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
    announcementEnabledInput,
    setAnnouncementEnabledInput,
    announcementTitleInput,
    setAnnouncementTitleInput,
    announcementBodyInput,
    setAnnouncementBodyInput,
    announcementAudienceInput,
    setAnnouncementAudienceInput,
    announcementStartAtInput,
    setAnnouncementStartAtInput,
    announcementEndAtInput,
    setAnnouncementEndAtInput,
    announcementPriorityInput,
    setAnnouncementPriorityInput,
    handleSaveConfig,
    fetchConfig,
  }
}
