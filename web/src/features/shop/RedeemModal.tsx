import React, { useState } from 'react'
import { motion } from 'framer-motion'
import { Coins, Smartphone, Wallet, Building2, AlertCircle, CheckCircle2, Sparkles } from 'lucide-react'
import type { RewardCatalogItem } from '../../shared/types'
import { shopApi } from '../../shared/lib/api'

interface RedeemModalProps {
  item: RewardCatalogItem
  userCoins: number
  onClose: () => void
  onSuccess: () => void
}

export const RedeemModal: React.FC<RedeemModalProps> = ({ item, userCoins, onClose, onSuccess }) => {
  const targetType = item.category === 'PULSA' ? 'PULSA' : item.category === 'EWALLET' ? 'EWALLET' : 'BANK'
  const [targetValue, setTargetValue] = useState('')
  const [provider, setProvider] = useState(
    item.category === 'PULSA' ? 'Telkomsel' : item.category === 'EWALLET' ? 'GoPay' : 'BCA'
  )
  const [submitting, setSubmitting] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)

  const canAfford = userCoins >= item.cost_coins

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!targetValue.trim()) {
      setErrorMessage('Nomor HP / Nomor Rekening tujuan wajib diisi')
      return
    }

    setSubmitting(true)
    setErrorMessage(null)

    try {
      const fullTarget = `${provider} - ${targetValue.trim()}`
      const res = await shopApi.redeem({
        reward_id: item.id,
        coins: item.cost_coins,
        target_type: targetType,
        target_value: fullTarget,
      })

      if (res.success) {
        setSuccess(true)
      } else {
        setErrorMessage('Gagal mengajukan penukaran koin')
      }
    } catch (err: any) {
      setErrorMessage(err.message || 'Gagal mengajukan klaim. Pastikan saldo koin mencukupi.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fadeIn">
      <motion.div
        initial={{ scale: 0.95, opacity: 0, y: 20 }}
        animate={{ scale: 1, opacity: 1, y: 0 }}
        exit={{ scale: 0.95, opacity: 0, y: 20 }}
        className="w-full max-w-md bg-surface-elevated border border-border-subtle rounded-3xl shadow-2xl overflow-hidden flex flex-col"
      >
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-border-subtle bg-surface-base">
          <div className="flex items-center gap-2">
            <span className="text-xl">🎁</span>
            <h3 className="font-heading font-bold text-text-primary text-base">
              Tukar Koin Hadiah
            </h3>
          </div>
          <button
            onClick={onClose}
            className="w-8 h-8 rounded-full bg-surface-elevated text-text-secondary hover:text-text-primary flex items-center justify-center transition-colors"
          >
            ✕
          </button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-5">
          {!success ? (
            <form onSubmit={handleSubmit} className="space-y-4">
              {/* Item Card Preview */}
              <div className="p-4 rounded-2xl bg-surface-base border border-border-subtle flex items-center justify-between">
                <div>
                  <h4 className="font-heading font-bold text-text-primary text-sm">
                    {item.title}
                  </h4>
                  <p className="text-xs text-text-secondary mt-0.5">{item.description}</p>
                </div>
                <div className="px-3 py-1.5 rounded-full bg-accent-gold/15 text-accent-gold font-bold text-xs flex items-center gap-1 shrink-0">
                  <Coins className="w-3.5 h-3.5" />
                  <span>{item.cost_coins.toLocaleString('id-ID')}</span>
                </div>
              </div>

              {/* Provider Selection */}
              {item.category === 'PULSA' && (
                <div className="space-y-1.5">
                  <label className="text-xs font-bold text-text-secondary flex items-center gap-1.5">
                    <Smartphone className="w-3.5 h-3.5" />
                    <span>Operator Seluler:</span>
                  </label>
                  <select
                    value={provider}
                    onChange={(e) => setProvider(e.target.value)}
                    className="w-full p-3 rounded-xl bg-surface-base border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic"
                  >
                    <option value="Telkomsel">Telkomsel</option>
                    <option value="Indosat (IM3)">Indosat (IM3)</option>
                    <option value="XL Axiata">XL Axiata</option>
                    <option value="Tri (3)">Tri (3)</option>
                    <option value="Smartfren">Smartfren</option>
                  </select>
                </div>
              )}

              {item.category === 'EWALLET' && (
                <div className="space-y-1.5">
                  <label className="text-xs font-bold text-text-secondary flex items-center gap-1.5">
                    <Wallet className="w-3.5 h-3.5" />
                    <span>Dompet Digital (E-Wallet):</span>
                  </label>
                  <select
                    value={provider}
                    onChange={(e) => setProvider(e.target.value)}
                    className="w-full p-3 rounded-xl bg-surface-base border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic"
                  >
                    <option value="GoPay">GoPay</option>
                    <option value="DANA">DANA</option>
                    <option value="OVO">OVO</option>
                    <option value="ShopeePay">ShopeePay</option>
                  </select>
                </div>
              )}

              {item.category === 'CASH' && (
                <div className="space-y-1.5">
                  <label className="text-xs font-bold text-text-secondary flex items-center gap-1.5">
                    <Building2 className="w-3.5 h-3.5" />
                    <span>Metode Penerimaan:</span>
                  </label>
                  <select
                    value={provider}
                    onChange={(e) => setProvider(e.target.value)}
                    className="w-full p-3 rounded-xl bg-surface-base border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic"
                  >
                    <option value="BCA">Transfer Bank BCA</option>
                    <option value="Mandiri">Transfer Bank Mandiri</option>
                    <option value="BRI">Transfer Bank BRI</option>
                    <option value="Uang Tunai">Uang Tunai Langsung</option>
                  </select>
                </div>
              )}

              {/* Destination Input */}
              <div className="space-y-1.5">
                <label className="text-xs font-bold text-text-secondary">
                  {item.category === 'CASH' && provider.includes('Tunai')
                    ? 'Nama Penerima:'
                    : 'Nomor HP / Nomor Rekening Tujuan:'}
                </label>
                <input
                  type="text"
                  required
                  value={targetValue}
                  onChange={(e) => setTargetValue(e.target.value)}
                  placeholder={
                    item.category === 'PULSA' || item.category === 'EWALLET'
                      ? 'Contoh: 081234567890'
                      : 'Contoh: 1234567890 (a.n Nama Pemilik)'
                  }
                  className="w-full p-3.5 rounded-xl bg-surface-base border border-border-subtle text-sm text-text-primary placeholder:text-text-secondary/50 focus:outline-none focus:border-accent-magic"
                />
              </div>

              {/* Balance comparison */}
              <div className="flex items-center justify-between text-xs p-3 rounded-xl bg-surface-base">
                <span className="text-text-secondary">Saldo Koin Kamu:</span>
                <span className={`font-bold ${canAfford ? 'text-text-primary' : 'text-status-error'}`}>
                  {userCoins.toLocaleString('id-ID')} 🪙
                </span>
              </div>

              {errorMessage && (
                <div className="p-3.5 rounded-xl bg-status-error/15 border border-status-error/30 text-status-error text-sm flex items-center gap-2">
                  <AlertCircle className="w-5 h-5 shrink-0" />
                  <span>{errorMessage}</span>
                </div>
              )}

              <button
                type="submit"
                disabled={!canAfford || submitting}
                className="w-full py-4 rounded-2xl bg-accent-gold text-text-primary font-heading font-bold shadow-lg shadow-accent-gold/30 hover:brightness-105 disabled:opacity-50 disabled:cursor-not-allowed active:scale-[0.98] transition-all flex items-center justify-center gap-2"
              >
                {submitting ? (
                  <span>Mengajukan Penukaran...</span>
                ) : !canAfford ? (
                  <span>Koin Belum Mencukupi</span>
                ) : (
                  <>
                    <Sparkles className="w-4 h-4" />
                    <span>Konfirmasi Tukar {item.cost_coins} Koin</span>
                  </>
                )}
              </button>
            </form>
          ) : (
            <div className="py-6 text-center space-y-4">
              <div className="w-16 h-16 mx-auto rounded-full bg-status-success/20 text-status-success flex items-center justify-center">
                <CheckCircle2 className="w-10 h-10" />
              </div>
              <div>
                <h4 className="font-heading font-bold text-xl text-text-primary">
                  Pengajuan Berhasil Dikirim! 🎉
                </h4>
                <p className="text-xs text-text-secondary mt-1">
                  Koinmu telah dipotong. Admin keluarga akan segera memproses transfer/top-up pulsa.
                </p>
              </div>

              <button
                onClick={() => {
                  onSuccess()
                  onClose()
                }}
                className="w-full py-3.5 rounded-2xl bg-accent-magic text-white font-heading font-bold"
              >
                Kembali ke Toko
              </button>
            </div>
          )}
        </div>
      </motion.div>
    </div>
  )
}
