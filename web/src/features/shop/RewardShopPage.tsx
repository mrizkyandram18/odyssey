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
    <div className="w-full max-w-lg mx-auto min-h-[calc(100vh-80px)] pb-24 flex flex-col space-y-4">
      {/* 1. Header Bar */}
      <header className="flex items-center justify-between pb-2 border-b border-border-subtle">
        <div>
          <h1 className="font-heading font-extrabold text-text-primary text-xl flex items-center gap-2">
            <Coins className="w-6 h-6 text-accent-gold" />
            <span>Toko Hadiah</span>
          </h1>
          <p className="text-xs text-text-secondary mt-0.5">
            Pencairan koin reward menjadi uang tunai atau saldo digital
          </p>
        </div>
        <button
          onClick={loadData}
          disabled={loading}
          className="p-2 rounded-xl bg-surface-elevated border border-border-subtle text-text-secondary hover:text-text-primary active:scale-95 transition-all shadow-sm"
          title="Perbarui Data"
        >
          <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
        </button>
      </header>

      {/* 2. Hero Card: Saldo Koin & Nilai Tunai */}
      <div className="relative overflow-hidden p-5 rounded-3xl bg-gradient-to-br from-surface-elevated via-surface-elevated/90 to-surface-base border border-border-subtle shadow-md">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <span className="text-[11px] font-bold uppercase tracking-wider text-text-secondary">
              Saldo Koin Kamu
            </span>
            <div className="flex items-center gap-2 mt-1">
              <span className="text-3xl font-heading font-extrabold text-accent-gold">
                {userCoins.toLocaleString('id-ID')}
              </span>
              <span className="text-sm font-bold text-text-secondary">Koin</span>
            </div>
          </div>

          <div className="sm:text-right border-t sm:border-t-0 pt-3 sm:pt-0 border-border-subtle/50">
            <span className="text-[11px] font-bold uppercase tracking-wider text-text-secondary">
              Nilai Penukaran Tunai
            </span>
            <div className="flex sm:justify-end items-center gap-1.5 mt-1">
              <Banknote className="w-5 h-5 text-status-success" />
              <span className="text-2xl font-heading font-extrabold text-status-success">
                Rp {estimatedCash.toLocaleString('id-ID')}
              </span>
            </div>
            <p className="text-[10px] text-text-secondary mt-0.5">
              1 Koin = Rp {conversionRate.toLocaleString('id-ID')}
            </p>
          </div>
        </div>
      </div>

      {/* 3. Navigation Tabs (Collision-Free) */}
      <div className="flex p-1 bg-surface-base rounded-2xl border border-border-subtle gap-1">
        <button
          data-testid="tab-redeem"
          onClick={() => setActiveTab('redeem')}
          className={`flex-1 py-2.5 px-3 rounded-xl font-heading font-bold text-xs flex items-center justify-center gap-1.5 transition-all text-center ${
            activeTab === 'redeem'
              ? 'bg-accent-magic text-white shadow-md shadow-accent-magic/30'
              : 'text-text-secondary hover:text-text-primary'
          }`}
        >
          <Banknote className="w-4 h-4 shrink-0" />
          <span className="truncate">Penukaran Koin</span>
        </button>
        <button
          data-testid="tab-history"
          onClick={() => setActiveTab('history')}
          className={`flex-1 py-2.5 px-3 rounded-xl font-heading font-bold text-xs flex items-center justify-center gap-1.5 transition-all text-center ${
            activeTab === 'history'
              ? 'bg-accent-magic text-white shadow-md shadow-accent-magic/30'
              : 'text-text-secondary hover:text-text-primary'
          }`}
        >
          <Clock className="w-4 h-4 shrink-0" />
          <span className="truncate">Riwayat Penukaran</span>
          {claims.length > 0 && (
            <span className="px-1.5 py-0.5 rounded-full bg-surface-base text-text-primary font-mono text-[10px] shrink-0">
              {claims.length}
            </span>
          )}
        </button>
      </div>

      {/* 4. Tab Contents */}
      {loading && !config ? (
        <div className="py-20 flex flex-col items-center gap-3">
          <div className="w-10 h-10 rounded-full border-4 border-accent-magic border-t-transparent animate-spin" />
          <p className="text-xs text-text-secondary">Memuat data penukaran koin...</p>
        </div>
      ) : error ? (
        <div className="p-6 text-center bg-status-error/10 border border-status-error/20 rounded-3xl space-y-3">
          <AlertCircle className="w-8 h-8 text-status-error mx-auto" />
          <p className="text-sm text-status-error font-bold">{error}</p>
          <button
            onClick={loadData}
            className="px-4 py-2 rounded-xl bg-surface-elevated text-xs font-bold text-text-primary shadow hover:bg-surface-base"
          >
            Coba Lagi
          </button>
        </div>
      ) : activeTab === 'redeem' ? (
        <div className="space-y-4">
          {/* Status Period Card */}
          <div
            className={`p-5 rounded-3xl border transition-all ${
              isOpen
                ? 'bg-status-success/10 border-status-success/30'
                : 'bg-surface-elevated border-border-subtle'
            }`}
          >
            <div className="flex items-start justify-between gap-3">
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <span
                    className={`text-[11px] font-extrabold px-3 py-1 rounded-full uppercase tracking-wider ${
                      isOpen
                        ? 'bg-status-success text-white shadow-sm shadow-status-success/30'
                        : 'bg-surface-base text-text-secondary border border-border-subtle'
                    }`}
                  >
                    {isOpen ? '● Penukaran Dibuka' : '○ Penukaran Ditutup'}
                  </span>
                </div>
                <h3 className="font-heading font-bold text-text-primary text-base pt-1">
                  {isOpen ? 'Periode Penukaran Sedang Aktif' : 'Penukaran Koin Belum Dibuka'}
                </h3>
                <p className="text-xs text-text-secondary leading-relaxed">
                  {isOpen
                    ? `Kamu dapat mengajukan penukaran koin menjadi cash pada periode tanggal ${startDay}–${endDay} bulan ini.`
                    : `Penukaran saat ini ditutup. Periode penukaran dibuka setiap tanggal ${startDay}–${endDay} setiap bulan.`}
                </p>
              </div>
            </div>

            {/* Action CTA */}
            <div className="mt-4 pt-3 border-t border-border-subtle/50">
              {isOpen ? (
                <button
                  onClick={() => setIsRedeemModalOpen(true)}
                  disabled={userCoins <= 0 || pendingClaims.length > 0}
                  className="w-full py-3.5 px-4 rounded-2xl bg-status-success hover:brightness-110 active:scale-[0.98] text-white font-heading font-bold text-sm shadow-lg shadow-status-success/30 transition-all flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <Banknote className="w-5 h-5" />
                  <span>
                    {pendingClaims.length > 0
                      ? 'Ada Pengajuan Pending yang Sedang Diproses'
                      : userCoins <= 0
                      ? 'Saldo Koin 0 (Kerjakan Tugas Terlebih Dahulu)'
                      : 'Tukarkan Koin Sekarang'}
                  </span>
                </button>
              ) : (
                <div className="flex items-center gap-2 text-xs text-text-secondary p-3 rounded-xl bg-surface-base border border-border-subtle">
                  <Info className="w-4 h-4 shrink-0 text-accent-magic" />
                  <span>
                    Tombol penukaran akan otomatis aktif ketika tanggal kalender memasuki periode{' '}
                    <strong>
                      {startDay}–{endDay}
                    </strong>
                    .
                  </span>
                </div>
              )}
            </div>
          </div>

          {/* Pending Claim Banner if exists */}
          {pendingClaims.length > 0 && (
            <div className="p-4 rounded-2xl bg-accent-gold/10 border border-accent-gold/30 flex items-start gap-3">
              <Clock className="w-5 h-5 text-accent-gold shrink-0 mt-0.5" />
              <div className="space-y-0.5">
                <h4 className="font-heading font-bold text-xs text-text-primary">
                  Pengajuan Pencairan Sedang Menunggu Verifikasi Admin
                </h4>
                <p className="text-[11px] text-text-secondary leading-relaxed">
                  Kamu telah mengajukan pencairan{' '}
                  <strong>{pendingClaims[0].coins_redeemed.toLocaleString('id-ID')} Koin</strong> (
                  {pendingClaims[0].target_value}). Mohon tunggu admin keluarga mentransfer dana.
                </p>
              </div>
            </div>
          )}

          {/* Information Guide Card */}
          <div className="p-5 rounded-3xl bg-surface-elevated border border-border-subtle space-y-3">
            <h4 className="font-heading font-bold text-text-primary text-xs uppercase tracking-wider flex items-center gap-1.5">
              <Info className="w-4 h-4 text-accent-magic" />
              <span>Cara Kerja Pencairan Reward</span>
            </h4>
            <div className="space-y-2.5 text-xs text-text-secondary">
              <div className="flex items-start gap-2.5">
                <div className="w-5 h-5 rounded-full bg-accent-magic/15 text-accent-magic font-bold text-[11px] flex items-center justify-center shrink-0 mt-0.5">
                  1
                </div>
                <p>
                  Selesaikan tugas harian (video, kuis, foto, dokumen) untuk mengumpulkan koin.
                </p>
              </div>
              <div className="flex items-start gap-2.5">
                <div className="w-5 h-5 rounded-full bg-accent-magic/15 text-accent-magic font-bold text-[11px] flex items-center justify-center shrink-0 mt-0.5">
                  2
                </div>
                <p>
                  Setiap tanggal <strong>{startDay}–{endDay}</strong>, periode penukaran dibuka.
                </p>
              </div>
              <div className="flex items-start gap-2.5">
                <div className="w-5 h-5 rounded-full bg-accent-magic/15 text-accent-magic font-bold text-[11px] flex items-center justify-center shrink-0 mt-0.5">
                  3
                </div>
                <p>
                  Tukarkan koin ke rekening Bank, E-Wallet (GoPay, DANA, OVO), atau uang tunai langsung.
                </p>
              </div>
            </div>
          </div>
        </div>
      ) : (
        /* History Tab */
        <div className="space-y-3">
          {claims.length === 0 ? (
            <div className="py-16 text-center bg-surface-elevated rounded-3xl border border-border-subtle space-y-2 p-6">
              <p className="text-3xl">📭</p>
              <p className="font-heading font-bold text-text-primary text-sm">
                Belum Ada Riwayat Penukaran
              </p>
              <p className="text-xs text-text-secondary max-w-xs mx-auto">
                Kumpulkan koin dengan menyelesaikan tugas harian, lalu tukarkan saat periode penukaran dibuka.
              </p>
            </div>
          ) : (
            claims.map((claim) => {
              const claimCash = claim.coins_redeemed * conversionRate
              return (
                <div
                  key={claim.id}
                  className="p-4 rounded-2xl bg-surface-elevated border border-border-subtle shadow-sm space-y-2.5"
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
                        <span className="font-heading font-bold text-text-primary text-sm">
                          Pencairan {claim.target_type}
                        </span>
                      </div>
                      <p className="text-xs text-text-secondary font-mono mt-0.5">
                        {claim.target_value}
                      </p>
                    </div>

                    <span
                      className={`text-[10px] font-bold px-2.5 py-1 rounded-full flex items-center gap-1 shrink-0 ${
                        claim.status === 'APPROVED'
                          ? 'bg-status-success/15 text-status-success'
                          : claim.status === 'PENDING'
                          ? 'bg-accent-gold/15 text-accent-gold animate-pulse'
                          : 'bg-status-error/15 text-status-error'
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
