import React, { useState } from 'react'
import { motion } from 'framer-motion'
import { Coins, Wallet, Building2, AlertCircle, CheckCircle2, Banknote, X, ArrowLeft, ArrowRight, ShieldAlert } from 'lucide-react'
import { shopApi } from '../../shared/lib/api'

interface RedeemModalProps {
  userCoins: number
  conversionRate: number
  onClose: () => void
  onSuccess: () => void
}

const MANDATORY_WARNING_TEXT =
  'Pastikan data tujuan pencairan sudah benar. Kesalahan nomor HP/rekening atau data tujuan menjadi tanggung jawab pengguna dan dana yang sudah dikirim tidak dapat dikembalikan.'

function maskDestinationNumber(val: string): string {
  const clean = val.trim()
  if (clean.length <= 6) return clean
  const start = clean.slice(0, 4)
  const end = clean.slice(-4)
  const stars = '*'.repeat(Math.min(6, Math.max(3, clean.length - 8)))
  return `${start}${stars}${end}`
}

export const RedeemModal: React.FC<RedeemModalProps> = ({
  userCoins,
  conversionRate,
  onClose,
  onSuccess,
}) => {
  const [step, setStep] = useState<'form' | 'confirm' | 'success'>('form')
  const [targetType, setTargetType] = useState<'EWALLET' | 'BANK'>('EWALLET')
  const [provider, setProvider] = useState('GoPay')
  const [accountNumber, setAccountNumber] = useState('')
  const [accountName, setAccountName] = useState('')
  const [coinsToRedeem, setCoinsToRedeem] = useState<number>(userCoins)
  const [submitting, setSubmitting] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)

  const calculatedCash = (coinsToRedeem || 0) * conversionRate
  const isValidAmount = coinsToRedeem > 0 && coinsToRedeem <= userCoins

  const handleTypeChange = (type: 'EWALLET' | 'BANK') => {
    setTargetType(type)
    if (type === 'EWALLET') setProvider('GoPay')
    else setProvider('BCA')
    setErrorMessage(null)
  }

  const validateInput = (): boolean => {
    const rawNumber = accountNumber.trim()
    if (!rawNumber) {
      setErrorMessage(
        targetType === 'BANK'
          ? 'Nomor rekening bank tujuan wajib diisi'
          : 'Nomor HP / akun e-wallet tujuan wajib diisi'
      )
      return false
    }

    const digitsOnly = rawNumber.replace(/\D/g, '')
    if (digitsOnly.length < 5) {
      setErrorMessage(
        targetType === 'BANK'
          ? 'Nomor rekening minimal 5 digit angka'
          : 'Nomor HP / akun e-wallet minimal 9 digit angka'
      )
      return false
    }

    if (targetType === 'EWALLET' && (digitsOnly.length < 9 || digitsOnly.length > 15)) {
      setErrorMessage('Format nomor HP / e-wallet harus antara 9 hingga 15 digit')
      return false
    }

    if (!isValidAmount) {
      setErrorMessage('Jumlah koin yang ditukarkan tidak valid')
      return false
    }

    setErrorMessage(null)
    return true
  }

  const handleProceedToConfirm = (e: React.FormEvent) => {
    e.preventDefault()
    if (!validateInput()) return
    setStep('confirm')
  }

  const handleFinalSubmit = async () => {
    if (submitting) return
    if (!validateInput()) {
      setStep('form')
      return
    }

    setSubmitting(true)
    setErrorMessage(null)

    try {
      const trimmedNumber = accountNumber.trim()
      const trimmedName = accountName.trim()
      const fullTarget = trimmedName
        ? `${provider} - ${trimmedNumber} (a.n ${trimmedName})`
        : `${provider} - ${trimmedNumber}`

      const res = await shopApi.redeem({
        coins: Number(coinsToRedeem),
        target_type: targetType,
        target_value: fullTarget,
      })

      if (res.success) {
        setStep('success')
      } else {
        setErrorMessage('Gagal mengajukan penukaran koin')
        setStep('form')
      }
    } catch (err: any) {
      setErrorMessage(err.message || 'Gagal mengajukan klaim penukaran.')
      setStep('form')
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
            {step === 'confirm' ? (
              <button
                type="button"
                onClick={() => setStep('form')}
                className="w-7 h-7 -ml-1.5 rounded-full hover:bg-surface-elevated text-text-secondary hover:text-text-primary flex items-center justify-center transition-colors"
                aria-label="Kembali ke formulir"
              >
                <ArrowLeft className="w-4 h-4" />
              </button>
            ) : (
              <Banknote className="w-5 h-5 text-status-success" />
            )}
            <h3 className="font-heading font-bold text-text-primary text-base">
              {step === 'confirm'
                ? 'Konfirmasi Pencairan Dana'
                : step === 'success'
                ? 'Pengajuan Berhasil'
                : 'Pencairan Koin ke Cash'}
            </h3>
          </div>
          <button
            onClick={onClose}
            className="w-8 h-8 rounded-full bg-surface-elevated text-text-secondary hover:text-text-primary flex items-center justify-center transition-colors"
            aria-label="Tutup modal"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Content */}
        <div className="p-6 overflow-y-auto space-y-4">
          {step === 'form' && (
            <form
              noValidate
              onSubmit={handleProceedToConfirm}
              className="space-y-4"
            >
              {/* Method Selector */}
              <div className="space-y-1.5">
                <label className="text-xs font-bold text-text-secondary">
                  Pilih Metode Pencairan:
                </label>
                <div className="grid grid-cols-2 gap-2">
                  <button
                    type="button"
                    data-testid="method-ewallet"
                    onClick={() => handleTypeChange('EWALLET')}
                    className={`p-3 rounded-2xl border text-xs font-heading font-bold flex flex-col items-center gap-1.5 transition-all ${
                      targetType === 'EWALLET'
                        ? 'bg-accent-magic/15 border-accent-magic text-accent-magic shadow-sm'
                        : 'bg-surface border-border-subtle text-text-secondary hover:text-text-primary'
                    }`}
                  >
                    <Wallet className="w-4 h-4" />
                    <span>E-Wallet</span>
                  </button>

                  <button
                    type="button"
                    data-testid="method-bank"
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
                  {targetType === 'EWALLET' ? 'Pilihan E-Wallet:' : 'Pilihan Bank:'}
                </label>
                <select
                  value={provider}
                  data-testid="provider-select"
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
                  {targetType === 'BANK' ? 'Nomor Rekening:' : 'Nomor HP / Nomor E-Wallet:'}
                </label>
                <input
                  type="text"
                  data-testid="account-number-input"
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
                  data-testid="account-name-input"
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
                    data-testid="coins-input"
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

              {/* Mandatory Warning Banner */}
              <div className="p-3.5 rounded-xl bg-accent-gold/10 border border-accent-gold/30 text-text-primary text-xs flex items-start gap-2.5">
                <ShieldAlert className="w-4 h-4 text-accent-gold shrink-0 mt-0.5" />
                <p className="leading-relaxed text-[11px] text-text-secondary">
                  <strong className="text-text-primary font-semibold">Peringatan: </strong>
                  {MANDATORY_WARNING_TEXT}
                </p>
              </div>

              {errorMessage && (
                <div className="p-3 rounded-xl bg-status-error/15 border border-status-error/30 text-status-error text-xs flex items-center gap-2">
                  <AlertCircle className="w-4 h-4 shrink-0" />
                  <span>{errorMessage}</span>
                </div>
              )}

              <button
                type="submit"
                data-testid="proceed-confirm-btn"
                disabled={!isValidAmount}
                className="w-full py-4 rounded-2xl bg-accent-magic hover:brightness-110 active:scale-[0.98] text-white font-heading font-bold text-sm shadow-lg shadow-accent-magic/30 transition-all flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <span>Lanjut ke Konfirmasi</span>
                <ArrowRight className="w-4 h-4" />
              </button>
            </form>
          )}

          {step === 'confirm' && (
            <div className="space-y-4">
              {/* Review Card */}
              <div className="p-4 rounded-2xl bg-surface border border-border-subtle space-y-3">
                <h4 className="text-xs font-bold text-text-secondary uppercase tracking-wider">
                  Ringkasan Pencairan
                </h4>

                <div className="space-y-2.5 text-xs">
                  <div className="flex items-center justify-between pb-2 border-b border-border-subtle">
                    <span className="text-text-secondary">Metode Pencairan</span>
                    <span className="font-bold text-text-primary">
                      {targetType === 'EWALLET' ? 'E-Wallet' : 'Transfer Bank'}
                    </span>
                  </div>

                  <div className="flex items-center justify-between pb-2 border-b border-border-subtle">
                    <span className="text-text-secondary">
                      {targetType === 'EWALLET' ? 'Provider E-Wallet' : 'Bank Tujuan'}
                    </span>
                    <span className="font-bold text-text-primary">{provider}</span>
                  </div>

                  <div className="flex items-center justify-between pb-2 border-b border-border-subtle">
                    <span className="text-text-secondary">
                      {targetType === 'BANK' ? 'Nomor Rekening' : 'Nomor HP / Akun'}
                    </span>
                    <div className="text-right">
                      <span className="font-bold font-mono text-text-primary block">
                        {maskDestinationNumber(accountNumber)}
                      </span>
                      <span className="text-[10px] text-text-secondary font-mono">
                        ({accountNumber.trim()})
                      </span>
                    </div>
                  </div>

                  {accountName.trim() && (
                    <div className="flex items-center justify-between pb-2 border-b border-border-subtle">
                      <span className="text-text-secondary">Atas Nama</span>
                      <span className="font-bold text-text-primary">{accountName.trim()}</span>
                    </div>
                  )}

                  <div className="flex items-center justify-between pb-2 border-b border-border-subtle">
                    <span className="text-text-secondary">Jumlah Koin Ditukar</span>
                    <span className="font-bold text-accent-gold">
                      {coinsToRedeem.toLocaleString('id-ID')} Koin
                    </span>
                  </div>

                  <div className="flex items-center justify-between pt-1">
                    <span className="text-text-secondary font-semibold">Total Dana Diterima</span>
                    <span className="text-base font-extrabold text-status-success">
                      Rp {calculatedCash.toLocaleString('id-ID')}
                    </span>
                  </div>
                </div>
              </div>

              {/* Mandatory Warning Alert (Explicit & Clear) */}
              <div className="p-4 rounded-2xl bg-accent-gold/15 border-2 border-accent-gold/40 text-text-primary space-y-1.5">
                <div className="flex items-center gap-2 text-accent-gold font-bold text-xs">
                  <ShieldAlert className="w-4 h-4 shrink-0" />
                  <span>PENTING — BACA SEBELUM KONFIRMASI</span>
                </div>
                <p className="text-xs leading-relaxed text-text-secondary">
                  {MANDATORY_WARNING_TEXT}
                </p>
              </div>

              {errorMessage && (
                <div className="p-3 rounded-xl bg-status-error/15 border border-status-error/30 text-status-error text-xs flex items-center gap-2">
                  <AlertCircle className="w-4 h-4 shrink-0" />
                  <span>{errorMessage}</span>
                </div>
              )}

              {/* Action Buttons */}
              <div className="flex gap-2.5 pt-2">
                <button
                  type="button"
                  onClick={() => setStep('form')}
                  disabled={submitting}
                  className="flex-1 py-3.5 rounded-2xl bg-surface border border-border-subtle hover:bg-surface-elevated text-text-secondary hover:text-text-primary font-heading font-bold text-xs transition-all disabled:opacity-50"
                >
                  Ubah Data
                </button>

                <button
                  type="button"
                  data-testid="final-confirm-btn"
                  onClick={handleFinalSubmit}
                  disabled={submitting}
                  className="flex-[2] py-3.5 rounded-2xl bg-status-success hover:brightness-110 active:scale-[0.98] text-white font-heading font-bold text-xs shadow-lg shadow-status-success/30 transition-all flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {submitting ? (
                    <span>Memproses...</span>
                  ) : (
                    <>
                      <CheckCircle2 className="w-4 h-4" />
                      <span>Konfirmasi & Cairkan</span>
                    </>
                  )}
                </button>
              </div>
            </div>
          )}

          {step === 'success' && (
            <div className="py-6 text-center space-y-4">
              <div className="w-16 h-16 mx-auto rounded-full bg-status-success/20 text-status-success flex items-center justify-center shadow-inner">
                <CheckCircle2 className="w-10 h-10" />
              </div>
              <div className="space-y-1">
                <h4 className="font-heading font-bold text-xl text-text-primary">
                  Pengajuan Berhasil Dikirim! 🎉
                </h4>
                <p className="text-xs text-text-secondary max-w-xs mx-auto leading-relaxed">
                  Koinmu telah dipotong dan dicatat. Notifikasi telah diteruskan ke tim untuk verifikasi dan transfer dana.
                </p>
              </div>

              <button
                type="button"
                data-testid="success-close-btn"
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
