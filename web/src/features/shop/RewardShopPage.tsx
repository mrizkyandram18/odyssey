import React, { useState, useEffect, useCallback } from 'react'
import { motion } from 'framer-motion'
import { Coins, Gift, Smartphone, Wallet, Banknote, Clock, CheckCircle2, XCircle } from 'lucide-react'
import type { RewardCatalogItem, ClaimView } from '../../shared/types'
import { shopApi } from '../../shared/lib/api'
import { useSession } from '../../shared/hooks/useSession'
import { RedeemModal } from './RedeemModal'

export const RewardShopPage: React.FC = () => {
  const { profile, refreshProfile } = useSession()
  const [activeTab, setActiveTab] = useState<'catalog' | 'history'>('catalog')
  const [catalog, setCatalog] = useState<RewardCatalogItem[]>([])
  const [claims, setClaims] = useState<ClaimView[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedItem, setSelectedItem] = useState<RewardCatalogItem | null>(null)

  const userCoins = profile?.coins || 0

  const loadData = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const [items, myClaims] = await Promise.all([
        shopApi.getCatalog(),
        shopApi.getMyClaims(),
      ])
      setCatalog(items || [])
      setClaims(myClaims || [])
    } catch (err: any) {
      setError(err.message || 'Gagal memuat data hadiah')
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

  const getItemIcon = (category: string) => {
    switch (category) {
      case 'PULSA':
        return <Smartphone className="w-6 h-6 text-accent-cyan" />
      case 'EWALLET':
        return <Wallet className="w-6 h-6 text-accent-magic" />
      case 'CASH':
        return <Banknote className="w-6 h-6 text-status-success" />
      default:
        return <Gift className="w-6 h-6 text-accent-gold" />
    }
  }

  return (
    <div className="w-full max-w-md mx-auto min-h-[calc(100vh-80px)] pb-24 flex flex-col">
      {/* 1. Header Bar with Coins Balance */}
      <header className="sticky top-0 z-20 bg-surface-elevated/95 backdrop-blur-md border-b border-border-subtle px-4 py-3 shadow-sm flex items-center justify-between">
        <div>
          <h1 className="font-heading font-extrabold text-text-primary text-lg">
            Toko Hadiah 🎁
          </h1>
          <p className="text-xs text-text-secondary">
            Tukarkan koinmu dengan uang atau pulsa
          </p>
        </div>
        <div className="flex items-center gap-1.5 px-3.5 py-1.5 rounded-full bg-accent-gold/20 border border-accent-gold/40 text-accent-gold font-bold text-sm shadow-sm">
          <Coins className="w-4 h-4" />
          <span>{userCoins.toLocaleString('id-ID')} Koin</span>
        </div>
      </header>

      {/* 2. Navigation Tabs */}
      <div className="p-4 flex gap-2">
        <button
          onClick={() => setActiveTab('catalog')}
          className={`flex-1 py-2.5 rounded-2xl font-heading font-bold text-xs transition-all ${
            activeTab === 'catalog'
              ? 'bg-accent-magic text-white shadow-md shadow-accent-magic/30'
              : 'bg-surface-elevated text-text-secondary border border-border-subtle hover:text-text-primary'
          }`}
        >
          🎁 Katalog Hadiah
        </button>
        <button
          onClick={() => setActiveTab('history')}
          className={`flex-1 py-2.5 rounded-2xl font-heading font-bold text-xs transition-all ${
            activeTab === 'history'
              ? 'bg-accent-magic text-white shadow-md shadow-accent-magic/30'
              : 'bg-surface-elevated text-text-secondary border border-border-subtle hover:text-text-primary'
          }`}
        >
          📜 Riwayat Penukaran ({claims.length})
        </button>
      </div>

      {/* 3. Tab Contents */}
      <div className="flex-1 px-4 space-y-4">
        {loading && catalog.length === 0 ? (
          <div className="py-20 flex flex-col items-center gap-3">
            <div className="w-10 h-10 rounded-full border-4 border-accent-magic border-t-transparent animate-spin" />
            <p className="text-xs text-text-secondary">Memuat katalog hadiah...</p>
          </div>
        ) : error ? (
          <div className="p-6 text-center bg-status-error/10 border border-status-error/20 rounded-3xl">
            <p className="text-sm text-status-error font-bold mb-3">{error}</p>
            <button
              onClick={loadData}
              className="px-4 py-2 rounded-xl bg-surface-elevated text-xs font-bold text-text-primary shadow"
            >
              Coba Lagi
            </button>
          </div>
        ) : activeTab === 'catalog' ? (
          <div className="space-y-3">
            {catalog.map((item) => {
              const canAfford = userCoins >= item.cost_coins
              return (
                <motion.div
                  key={item.id}
                  whileHover={{ scale: 1.01 }}
                  className="p-4 rounded-2xl bg-surface-elevated border border-border-subtle shadow-sm flex items-center justify-between gap-3"
                >
                  <div className="flex items-center gap-3">
                    <div className="w-12 h-12 rounded-2xl bg-surface-base border border-border-subtle flex items-center justify-center shrink-0">
                      {getItemIcon(item.category)}
                    </div>
                    <div>
                      <h4 className="font-heading font-bold text-text-primary text-sm">
                        {item.title}
                      </h4>
                      <p className="text-xs text-text-secondary line-clamp-1">
                        {item.description}
                      </p>
                      <div className="flex items-center gap-1 text-xs font-bold text-accent-gold mt-1">
                        <Coins className="w-3.5 h-3.5" />
                        <span>{item.cost_coins.toLocaleString('id-ID')} Koin</span>
                      </div>
                    </div>
                  </div>

                  <button
                    onClick={() => setSelectedItem(item)}
                    disabled={!canAfford}
                    className={`px-4 py-2.5 rounded-xl font-heading font-bold text-xs shrink-0 transition-all ${
                      canAfford
                        ? 'bg-accent-magic text-white hover:brightness-110 active:scale-95 shadow-sm shadow-accent-magic/30'
                        : 'bg-surface-base text-text-secondary/50 border border-border-subtle cursor-not-allowed'
                    }`}
                  >
                    {canAfford ? 'Tukar 🪙' : 'Kurang'}
                  </button>
                </motion.div>
              )
            })}
          </div>
        ) : (
          <div className="space-y-3">
            {claims.length === 0 ? (
              <div className="py-16 text-center space-y-2">
                <p className="text-3xl">📭</p>
                <p className="font-heading font-bold text-text-primary text-sm">
                  Belum Ada Riwayat Penukaran
                </p>
                <p className="text-xs text-text-secondary max-w-xs mx-auto">
                  Kerjakan tugas harian untuk mengumpulkan koin dan tukar dengan pulsa atau uang nyata.
                </p>
              </div>
            ) : (
              claims.map((claim) => (
                <div
                  key={claim.id}
                  className="p-4 rounded-2xl bg-surface-elevated border border-border-subtle shadow-sm space-y-2"
                >
                  <div className="flex items-center justify-between">
                    <span className="font-heading font-bold text-text-primary text-sm">
                      {claim.reward_title || claim.target_type}
                    </span>
                    <span
                      className={`text-[10px] font-bold px-2.5 py-1 rounded-full flex items-center gap-1 ${
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
                          <span>Ditolak (Koin Refund)</span>
                        </>
                      )}
                    </span>
                  </div>

                  <p className="text-xs text-text-secondary font-mono">
                    Tujuan: {claim.target_value}
                  </p>

                  <div className="flex items-center justify-between text-[11px] text-text-secondary pt-1 border-t border-border-subtle">
                    <span>
                      {new Date(claim.created_at).toLocaleDateString('id-ID', {
                        day: 'numeric',
                        month: 'short',
                        hour: '2-digit',
                        minute: '2-digit',
                      })}
                    </span>
                    <span className="font-bold text-text-primary">
                      -{claim.coins_redeemed.toLocaleString('id-ID')} 🪙
                    </span>
                  </div>

                  {claim.admin_notes && (
                    <div className="text-[11px] p-2 rounded-lg bg-surface-base text-text-secondary">
                      <strong>Catatan Admin:</strong> {claim.admin_notes}
                    </div>
                  )}
                </div>
              ))
            )}
          </div>
        )}
      </div>

      {/* Redeem Modal */}
      {selectedItem && (
        <RedeemModal
          item={selectedItem}
          userCoins={userCoins}
          onClose={() => setSelectedItem(null)}
          onSuccess={handleRedeemSuccess}
        />
      )}
    </div>
  )
}
