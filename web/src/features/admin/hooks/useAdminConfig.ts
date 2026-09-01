import { useState, useCallback, useEffect } from 'react'
import { adminTasksApi } from '../../../shared/lib/api'
import type { RedemptionConfig } from '../../../shared/types'

export function useAdminConfig() {
  const [config, setConfig] = useState<RedemptionConfig | null>(null)
  const [isFetching, setIsFetching] = useState(false)
  const [isSaving, setIsSaving] = useState(false)
  const [successMsg, setSuccessMsg] = useState<string | null>(null)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  // Redemption-only inputs (Phase 1 refactor: earning configs removed)
  const [startDayInput, setStartDayInput] = useState('21')
  const [endDayInput, setEndDayInput] = useState('26')
  const [payoutDayInput, setPayoutDayInput] = useState('24')
  const [conversionRateInput, setConversionRateInput] = useState('100')
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
        setPayoutDayInput(String(res.payout_day ?? 24))
        setConversionRateInput(String(res.conversion_rate ?? 100))
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
    const rate = parseInt(conversionRateInput, 10)

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
    if (isNaN(rate) || rate <= 0) {
      setErrorMsg('Nilai konversi koin harus lebih dari 0')
      return
    }

    setIsSaving(true)
    try {
      const updated = await adminTasksApi.updateConfig({
        start_day: start,
        end_day: end,
        payout_day: payout,
        conversion_rate: rate,
        timezone: timezoneInput.trim() || 'Asia/Jakarta',
      })
      setConfig(updated)
      setSuccessMsg('Konfigurasi pencairan berhasil disimpan!')
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
    conversionRateInput,
    setConversionRateInput,
    timezoneInput,
    setTimezoneInput,
    handleSaveConfig,
    fetchConfig,
  }
}
