import React, { useState, useEffect, useCallback } from 'react'
import {
  Coins,
  Banknote,
  Clock,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Info,
  Wallet,
  Building2,
  Smartphone,
  RefreshCw,
} from 'lucide-react'
import type { ClaimView, RedemptionConfig } from '../../shared/types'
import { shopApi } from '../../shared/lib/api'
import { useSession } from '../../shared/hooks/useSession'
import { RedeemModal } from './RedeemModal'

export const RewardShopPage: React.FC = () => {
  const { profile, refreshProfile } = useSession()
  const [activeTab, setActiveTab] = useState<'redeem' | 'history'>('redeem')
  const [claims, setClaims] = useState<ClaimView[]>([])
  const [config, setConfig] = useState<RedemptionConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isRedeemModalOpen, setIsRedeemModalOpen] = useState(false)

  const userCoins = profile?.coins || 0

  const loadData = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const [cfg, myClaims] = await Promise.all([
        shopApi.getConfig(),
        shopApi.getMyClaims(),
      ])
      setConfig(cfg)
      setClaims(myClaims || [])
    } catch (err: any) {
      setError(err.message || 'Gagal memuat data penukaran hadiah')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadData()
  }, [loadData])

  const handleRedeemSuccess = () => {
    loadData()
    if (refreshProfile) {
      refreshProfile()
    }
  }

  const conversionRate = config?.conversion_rate || 10
  const estimatedCash = userCoins * conversionRate
  const isOpen = config?.is_open ?? false
  const startDay = config?.redemption_start_day ?? 21
  const endDay = config?.redemption_end_day ?? 26

  const pendingClaims = claims.filter((c) => c.status === 'PENDING')

  return (
    <div className="w-full flex flex-col gap-4">
      {/* 1. Header */}
      <header className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-bold text-text-primary flex items-center gap-2">
            <Coins className="w-5 h-5 text-accent-gold" />
            <span>Toko Hadiah</span>
          </h1>
          <p className="text-xs text-text-secondary mt-0.5">
            Tukarkan koin menjadi saldo digital
          </p>
        </div>
        <button
          onClick={loadData}
          disabled={loading}
          aria-label="Muat ulang"
          className="p-2 rounded-xl bg-surface border border-border-subtle text-text-secondary hover:text-text-primary transition-colors"
        >
          <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
        </button>
      </header>

      {/* 2. Hero Balance Card */}
      <div className="p-4 rounded-2xl bg-surface border border-border-subtle shadow-sm">
        <div className="flex items-center justify-between gap-4">
          <div>
            <p className="text-[11px] font-bold uppercase tracking-wide text-text-secondary">Saldo Koin</p>
            <p className="mt-1 flex items-baseline gap-1.5">
              <span className="text-2xl font-bold text-accent-gold">{userCoins.toLocaleString('id-ID')}</span>
              <span className="text-xs font-semibold text-text-secondary">Koin</span>
            </p>
          </div>
          <div className="text-right">
            <p className="text-[11px] font-bold uppercase tracking-wide text-text-secondary">Estimasi Tunai</p>
            <p className="mt-1 flex items-center justify-end gap-1 text-status-success font-bold">
              <Banknote className="w-4 h-4" />
              <span className="text-lg">Rp {estimatedCash.toLocaleString('id-ID')}</span>
            </p>
            <p className="text-[11px] text-text-secondary">1 Koin = Rp {conversionRate.toLocaleString('id-ID')}</p>
          </div>
        </div>
      </div>

      {/* 3. Tabs */}
      <div className="flex p-1 bg-surface rounded-xl border border-border-subtle gap-1">
        <button
          data-testid="tab-redeem"
          onClick={() => setActiveTab('redeem')}
          className={`flex-1 py-2.5 px-3 rounded-lg font-bold text-xs flex items-center justify-center gap-1.5 transition-colors ${
            activeTab === 'redeem'
              ? 'bg-accent-magic text-white shadow-sm'
              : 'text-text-secondary hover:text-text-primary'
          }`}
        >
          <Banknote className="w-4 h-4 shrink-0" />
          <span>Penukaran</span>
        </button>
        <button
          data-testid="tab-history"
          onClick={() => setActiveTab('history')}
          className={`flex-1 py-2.5 px-3 rounded-lg font-bold text-xs flex items-center justify-center gap-1.5 transition-colors ${
            activeTab === 'history'
              ? 'bg-accent-magic text-white shadow-sm'
              : 'text-text-secondary hover:text-text-primary'
          }`}
        >
          <Clock className="w-4 h-4 shrink-0" />
          <span>Riwayat</span>
          {claims.length > 0 && (
            <span className="px-1.5 py-0.5 rounded-full bg-white/20 text-white font-mono text-[10px] shrink-0">
              {claims.length}
            </span>
          )}
        </button>
      </div>

      {/* 4. Tab Contents */}
      {loading && !config ? (
        <div className="py-16 flex flex-col items-center gap-3">
          <div className="w-8 h-8 rounded-full border-2 border-accent-magic border-t-transparent animate-spin" />
          <p className="text-xs text-text-secondary font-medium">Memuat data...</p>
        </div>
      ) : error ? (
        <div className="p-5 text-center bg-surface border border-status-error/20 rounded-2xl space-y-3">
          <AlertCircle className="w-7 h-7 text-status-error mx-auto" />
          <p className="text-sm text-status-error font-semibold">{error}</p>
          <button
            onClick={loadData}
            className="px-4 py-2 rounded-xl bg-surface border border-border-subtle text-xs font-bold text-text-primary hover:bg-surface-elevated transition-colors"
          >
            Coba Lagi
          </button>
        </div>
      ) : activeTab === 'redeem' ? (
        <div className="space-y-4">
          {/* Status Period Card */}
          <div
            className={`p-4 rounded-2xl border ${
              isOpen
                ? 'bg-status-success/10 border-status-success/20'
                : 'bg-surface border-border-subtle'
            }`}
          >
            <span
              className={`inline-flex text-[11px] font-bold px-2.5 py-1 rounded-full ${
                isOpen
                  ? 'bg-status-success text-white'
                  : 'bg-surface-elevated text-text-secondary border border-border-subtle'
              }`}
            >
              {isOpen ? '● Penukaran Dibuka' : '○ Penukaran Ditutup'}
            </span>
            <h3 className="font-bold text-text-primary text-sm mt-2">
              {isOpen ? 'Periode Penukaran Sedang Aktif' : 'Penukaran Koin Belum Dibuka'}
            </h3>
            <p className="text-xs text-text-secondary leading-relaxed mt-1">
              {isOpen
                ? `Ajukan penukaran koin menjadi cash pada tanggal ${startDay}–${endDay} bulan ini.`
                : `Penukaran dibuka setiap tanggal ${startDay}–${endDay} setiap bulan.`}
            </p>

            <div className="mt-4">
              {isOpen ? (
                <button
                  onClick={() => setIsRedeemModalOpen(true)}
                  disabled={userCoins <= 0 || pendingClaims.length > 0}
                  className="w-full py-3 px-4 rounded-xl bg-accent-magic hover:brightness-110 text-white font-bold text-sm shadow-sm transition-all flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <Banknote className="w-4 h-4" />
                  <span>
                    {pendingClaims.length > 0
                      ? 'Menunggu verifikasi admin'
                      : userCoins <= 0
                      ? 'Kerjakan tugas untuk kumpulkan koin'
                      : 'Tukarkan Koin Sekarang'}
                  </span>
                </button>
              ) : (
                <div className="flex items-start gap-2 text-xs text-text-secondary p-3 rounded-xl bg-surface-elevated border border-border-subtle">
                  <Info className="w-4 h-4 shrink-0 text-accent-magic mt-0.5" />
                  <span>
                    Penukaran aktif otomatis pada tanggal <strong>{startDay}–{endDay}</strong>.
                  </span>
                </div>
              )}
            </div>
          </div>

          {pendingClaims.length > 0 && (
            <div className="p-3 rounded-xl bg-accent-gold/10 border border-accent-gold/20 flex items-start gap-2.5">
              <Clock className="w-4 h-4 text-accent-gold shrink-0 mt-0.5" />
              <div>
                <p className="font-bold text-xs text-text-primary">Pengajuan menunggu verifikasi</p>
                <p className="text-xs text-text-secondary leading-relaxed mt-0.5">
                  {pendingClaims[0].coins_redeemed.toLocaleString('id-ID')} Koin — {pendingClaims[0].target_value}
                </p>
              </div>
            </div>
          )}

          <div className="p-4 rounded-2xl bg-surface border border-border-subtle space-y-3">
            <h4 className="font-bold text-text-primary text-xs flex items-center gap-1.5">
              <Info className="w-4 h-4 text-accent-magic" />
              <span>Cara kerja</span>
            </h4>
            <ol className="space-y-2 text-xs text-text-secondary list-decimal list-inside leading-relaxed">
              <li>Selesaikan tugas harian untuk kumpulkan koin</li>
              <li>Tunggu periode penukaran tanggal <strong>{startDay}–{endDay}</strong></li>
              <li>Tukarkan ke Bank, E-Wallet, atau tunai</li>
            </ol>
          </div>
        </div>
      ) : (
        <div className="space-y-3">
          {claims.length === 0 ? (
            <div className="py-12 text-center bg-surface rounded-2xl border border-border-subtle p-6 space-y-2">
              <p className="text-2xl">📭</p>
              <p className="font-bold text-text-primary text-sm">Belum ada riwayat</p>
              <p className="text-xs text-text-secondary max-w-xs mx-auto">
                Selesaikan tugas untuk kumpulkan koin, lalu tukarkan saat periode dibuka.
              </p>
            </div>
          ) : (
            claims.map((claim) => {
              const claimCash = claim.coins_redeemed * conversionRate
              return (
                <div
                  key={claim.id}
                  className="p-4 rounded-2xl bg-surface border border-border-subtle shadow-sm space-y-3"
                >
                  <div className="flex items-start justify-between gap-2">
                    <div>
                      <div className="flex items-center gap-1.5">
                        {claim.target_type === 'EWALLET' ? (
                          <Wallet className="w-4 h-4 text-accent-magic" />
                        ) : claim.target_type === 'BANK' ? (
                          <Building2 className="w-4 h-4 text-status-success" />
                        ) : (
                          <Smartphone className="w-4 h-4 text-accent-cyan" />
                        )}
                        <span className="font-bold text-text-primary text-sm">
                          Pencairan {claim.target_type}
                        </span>
                      </div>
                      <p className="text-xs text-text-secondary font-mono mt-0.5 break-all">
                        {claim.target_value}
                      </p>
                    </div>

                    <span
                      className={`text-[10px] font-bold px-2.5 py-1 rounded-full flex items-center gap-1 shrink-0 border ${
                        claim.status === 'APPROVED'
                          ? 'bg-status-success/10 text-status-success border-status-success/20'
                          : claim.status === 'PENDING'
                          ? 'bg-accent-gold/10 text-accent-gold border-accent-gold/20'
                          : 'bg-status-error/10 text-status-error border-status-error/20'
                      }`}
                    >
                      {claim.status === 'APPROVED' ? (
                        <>
                          <CheckCircle2 className="w-3 h-3" />
                          <span>Berhasil Ditransfer</span>
                        </>
                      ) : claim.status === 'PENDING' ? (
                        <>
                          <Clock className="w-3 h-3" />
                          <span>Menunggu Admin</span>
                        </>
                      ) : (
                        <>
                          <XCircle className="w-3 h-3" />
                          <span>Ditolak & Koin Dikembalikan</span>
                        </>
                      )}
                    </span>
                  </div>

                  <div className="flex items-center justify-between text-xs pt-2 border-t border-border-subtle">
                    <span className="text-text-secondary">
                      {new Date(claim.created_at).toLocaleDateString('id-ID', {
                        day: 'numeric',
                        month: 'short',
                        year: 'numeric',
                        hour: '2-digit',
                        minute: '2-digit',
                      })}
                    </span>
                    <div className="text-right">
                      <span className="font-bold text-text-primary">
                        -{claim.coins_redeemed.toLocaleString('id-ID')} Koin
                      </span>
                      <span className="text-[11px] text-status-success font-bold ml-1.5">
                        (Rp {claimCash.toLocaleString('id-ID')})
                      </span>
                    </div>
                  </div>

                  {claim.admin_notes && (
                    <div className="text-[11px] p-2.5 rounded-xl bg-surface-base border border-border-subtle text-text-secondary">
                      <strong className="text-text-primary">Catatan Admin:</strong>{' '}
                      {claim.admin_notes}
                    </div>
                  )}
                </div>
              )
            })
          )}
        </div>
      )}

      {/* Redeem Modal Dialog */}
      {isRedeemModalOpen && (
        <RedeemModal
          userCoins={userCoins}
          conversionRate={conversionRate}
          onClose={() => setIsRedeemModalOpen(false)}
          onSuccess={() => {
            setIsRedeemModalOpen(false)
            handleRedeemSuccess()
          }}
        />
      )}
    </div>
  )
}
