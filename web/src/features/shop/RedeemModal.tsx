import React, { useState } from 'react'
import { motion } from 'framer-motion'
import { Coins, Wallet, Building2, AlertCircle, CheckCircle2, Banknote, X } from 'lucide-react'
import { shopApi } from '../../shared/lib/api'

interface RedeemModalProps {
  userCoins: number
  conversionRate: number
  onClose: () => void
  onSuccess: () => void
}

export const RedeemModal: React.FC<RedeemModalProps> = ({
  userCoins,
  conversionRate,
  onClose,
  onSuccess,
}) => {
  const [targetType, setTargetType] = useState<'EWALLET' | 'BANK' | 'PHONE'>('EWALLET')
  const [provider, setProvider] = useState('GoPay')
  const [accountNumber, setAccountNumber] = useState('')
  const [accountName, setAccountName] = useState('')
  const [coinsToRedeem, setCoinsToRedeem] = useState<number>(userCoins)
  const [submitting, setSubmitting] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)

  const calculatedCash = (coinsToRedeem || 0) * conversionRate
  const isValidAmount = coinsToRedeem > 0 && coinsToRedeem <= userCoins

  const handleTypeChange = (type: 'EWALLET' | 'BANK' | 'PHONE') => {
    setTargetType(type)
    if (type === 'EWALLET') setProvider('GoPay')
    else if (type === 'BANK') setProvider('BCA')
    else setProvider('Telkomsel')
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!accountNumber.trim()) {
      setErrorMessage('Nomor rekening / nomor HP tujuan wajib diisi')
      return
    }
    if (!isValidAmount) {
      setErrorMessage('Jumlah koin yang ditukarkan tidak valid')
      return
    }

    setSubmitting(true)
    setErrorMessage(null)

    try {
      const fullTarget = accountName.trim()
        ? `${provider} - ${accountNumber.trim()} (a.n ${accountName.trim()})`
        : `${provider} - ${accountNumber.trim()}`

      const res = await shopApi.redeem({
        coins: Number(coinsToRedeem),
        target_type: targetType,
        target_value: fullTarget,
      })

      if (res.success) {
        setSuccess(true)
      } else {
        setErrorMessage('Gagal mengajukan penukaran koin')
      }
    } catch (err: any) {
      setErrorMessage(err.message || 'Gagal mengajukan klaim penukaran.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fadeIn">
      <motion.div
        initial={{ scale: 0.95, opacity: 0, y: 15 }}
        animate={{ scale: 1, opacity: 1, y: 0 }}
        exit={{ scale: 0.95, opacity: 0, y: 15 }}
        className="w-full max-w-md bg-surface-elevated border border-border-subtle rounded-2xl shadow-xl overflow-hidden flex flex-col max-h-[90vh]"
      >
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-border-subtle bg-surface">
          <div className="flex items-center gap-2">
            <Banknote className="w-5 h-5 text-status-success" />
            <h3 className="font-heading font-bold text-text-primary text-base">
              Pencairan Koin ke Cash
            </h3>
          </div>
          <button
            onClick={onClose}
            className="w-8 h-8 rounded-full bg-surface-elevated text-text-secondary hover:text-text-primary flex items-center justify-center transition-colors"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Content */}
        <div className="p-6 overflow-y-auto space-y-4">
          {!success ? (
            <form onSubmit={handleSubmit} className="space-y-4">
              {/* Method Selector */}
              <div className="space-y-1.5">
                <label className="text-xs font-bold text-text-secondary">
                  Pilih Metode Pencairan:
                </label>
                <div className="grid grid-cols-2 gap-2">
                  <button
                    type="button"
                    onClick={() => handleTypeChange('EWALLET')}
                    className={`p-3 rounded-2xl border text-xs font-heading font-bold flex flex-col items-center gap-1.5 transition-all ${
                      targetType === 'EWALLET'
                        ? 'bg-accent-magic/15 border-accent-magic text-accent-magic shadow-sm'
                        : 'bg-surface border-border-subtle text-text-secondary hover:text-text-primary'
                    }`}
                  >
                    <Wallet className="w-4 h-4" />
                    <span>Dompet Digital</span>
                  </button>

                  <button
                    type="button"
                    onClick={() => handleTypeChange('BANK')}
                    className={`p-3 rounded-2xl border text-xs font-heading font-bold flex flex-col items-center gap-1.5 transition-all ${
                      targetType === 'BANK'
                        ? 'bg-status-success/15 border-status-success text-status-success shadow-sm'
                        : 'bg-surface border-border-subtle text-text-secondary hover:text-text-primary'
                    }`}
                  >
                    <Building2 className="w-4 h-4" />
                    <span>Transfer Bank</span>
                  </button>
                </div>
              </div>

              {/* Provider Selection */}
              <div className="space-y-1.5">
                <label className="text-xs font-bold text-text-secondary">
                  {targetType === 'EWALLET' ? 'Pilihan Dompet Digital:' : 'Pilihan Bank:'}
                </label>
                <select
                  value={provider}
                  onChange={(e) => setProvider(e.target.value)}
                  className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic font-medium"
                >
                  {targetType === 'EWALLET' && (
                    <>
                      <option value="GoPay">GoPay</option>
                      <option value="DANA">DANA</option>
                      <option value="OVO">OVO Cash</option>
                      <option value="ShopeePay">ShopeePay</option>
                    </>
                  )}
                  {targetType === 'BANK' && (
                    <>
                      <option value="BCA">Bank BCA</option>
                      <option value="Mandiri">Bank Mandiri</option>
                      <option value="BRI">Bank BRI</option>
                      <option value="BNI">Bank BNI</option>
                      <option value="BSI">Bank Syariah Indonesia (BSI)</option>
                    </>
                  )}
                </select>
              </div>

              {/* Target Number */}
              <div className="space-y-1.5">
                <label className="text-xs font-bold text-text-secondary">
                  {targetType === 'BANK' ? 'Nomor Rekening:' : 'Nomor Akun Tujuan:'}
                </label>
                <input
                  type="text"
                  required
                  value={accountNumber}
                  onChange={(e) => setAccountNumber(e.target.value)}
                  placeholder={
                    targetType === 'BANK'
                      ? 'Contoh: 1234567890'
                      : 'Contoh: 081234567890'
                  }
                  className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary placeholder:text-text-secondary/50 focus:outline-none focus:border-accent-magic font-mono"
                />
              </div>

              {/* Account Name */}
              <div className="space-y-1.5">
                <label className="text-xs font-bold text-text-secondary">
                  Nama Pemilik Rekening / Akun (Opsional):
                </label>
                <input
                  type="text"
                  value={accountName}
                  onChange={(e) => setAccountName(e.target.value)}
                  placeholder="Contoh: Budi Santoso"
                  className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary placeholder:text-text-secondary/50 focus:outline-none focus:border-accent-magic"
                />
              </div>

              {/* Amount to Redeem */}
              <div className="space-y-1.5">
                <div className="flex items-center justify-between">
                  <label className="text-xs font-bold text-text-secondary">
                    Jumlah Koin yang Ditukarkan:
                  </label>
                  <span className="text-[11px] text-text-secondary">
                    Saldo: <strong>{userCoins.toLocaleString('id-ID')} Koin</strong>
                  </span>
                </div>
                <div className="relative">
                  <input
                    type="number"
                    min={1}
                    max={userCoins}
                    value={coinsToRedeem}
                    onChange={(e) => setCoinsToRedeem(Number(e.target.value))}
                    className="w-full p-3.5 rounded-xl bg-surface border border-border-subtle text-base font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
                  />
                  <div className="absolute right-3 top-3.5 flex items-center gap-1 text-accent-gold text-xs font-bold">
                    <Coins className="w-4 h-4" />
                    <span>Koin</span>
                  </div>
                </div>

                {/* Quick select buttons */}
                <div className="flex gap-2 pt-1">
                  <button
                    type="button"
                    onClick={() => setCoinsToRedeem(userCoins)}
                    className="px-3 py-1 rounded-lg bg-surface border border-border-subtle text-[11px] font-bold text-text-secondary hover:text-text-primary"
                  >
                    Semua Koin ({userCoins})
                  </button>
                  {userCoins >= 2 && (
                    <button
                      type="button"
                      onClick={() => setCoinsToRedeem(Math.floor(userCoins / 2))}
                      className="px-3 py-1 rounded-lg bg-surface border border-border-subtle text-[11px] font-bold text-text-secondary hover:text-text-primary"
                    >
                      50% ({Math.floor(userCoins / 2)})
                    </button>
                  )}
                </div>
              </div>

              {/* Conversion Result Preview */}
              <div className="p-4 rounded-2xl bg-gradient-to-r from-status-success/15 to-accent-magic/10 border border-status-success/30 flex items-center justify-between">
                <div>
                  <span className="text-[11px] text-text-secondary font-medium">
                    Total Dana Diterima:
                  </span>
                  <p className="font-heading font-extrabold text-lg text-status-success">
                    Rp {calculatedCash.toLocaleString('id-ID')}
                  </p>
                </div>
                <div className="text-right text-[10px] text-text-secondary font-mono">
                  {coinsToRedeem.toLocaleString('id-ID')} × Rp {conversionRate}
                </div>
              </div>

              {errorMessage && (
                <div className="p-3 rounded-xl bg-status-error/15 border border-status-error/30 text-status-error text-xs flex items-center gap-2">
                  <AlertCircle className="w-4 h-4 shrink-0" />
                  <span>{errorMessage}</span>
                </div>
              )}

              <button
                type="submit"
                disabled={!isValidAmount || submitting}
                className="w-full py-4 rounded-2xl bg-status-success hover:brightness-110 active:scale-[0.98] text-white font-heading font-bold text-sm shadow-lg shadow-status-success/30 transition-all flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {submitting ? (
                  <span>Mengirim Pengajuan...</span>
                ) : (
                  <>
                    <CheckCircle2 className="w-4 h-4" />
                    <span>Konfirmasi Cairkan Rp {calculatedCash.toLocaleString('id-ID')}</span>
                  </>
                )}
              </button>
            </form>
          ) : (
            <div className="py-6 text-center space-y-4">
              <div className="w-16 h-16 mx-auto rounded-full bg-status-success/20 text-status-success flex items-center justify-center shadow-inner">
                <CheckCircle2 className="w-10 h-10" />
              </div>
              <div className="space-y-1">
                <h4 className="font-heading font-bold text-xl text-text-primary">
                  Pengajuan Berhasil Dikirim! 🎉
                </h4>
                <p className="text-xs text-text-secondary max-w-xs mx-auto leading-relaxed">
                  Koinmu telah berhasil dipotong dan dicatat dalam riwayat penukaran. Admin akan segera memproses transfer dana pencairan.
                </p>
              </div>

              <button
                onClick={() => {
                  onSuccess()
                  onClose()
                }}
                className="w-full py-3.5 rounded-2xl bg-accent-magic text-white font-heading font-bold text-sm shadow-md shadow-accent-magic/30 hover:brightness-110 transition-all"
              >
                Kembali ke Halaman Penukaran
              </button>
            </div>
          )}
        </div>
      </motion.div>
    </div>
  )
}
