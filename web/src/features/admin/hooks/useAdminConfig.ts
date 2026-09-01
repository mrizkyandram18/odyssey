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
    handleSaveConfig,
    fetchConfig,
  }
}
