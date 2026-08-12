import { useState, useEffect, useCallback } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '../../shared/components/atoms/Button'
import { ProgressBar } from '../../shared/components/atoms/ProgressBar'
import { Card } from '../../shared/components/atoms/Card'
import { useSession } from '../../shared/hooks/useSession'
import { apiClient, crewsApi } from '../../shared/lib/api'
import type {
  CosmeticCatalogItem,
  CosmeticsResponse,
  InventoryItem,
  CrewMember,
  GiftRelicResult,
} from '../../shared/types'
import { Avatar } from '../../shared/components/atoms/Avatar'
import { Shuffle } from 'lucide-react'

import { getRoleMastery } from '../../shared/utils/roleMastery'
import { PushNotificationToggle } from '../../shared/components/molecules/PushNotificationToggle'

export function ProfilePage() {
  const { profile, loading, error, refreshProfile, logout } = useSession()
  const [activeView, setActiveView] = useState<'overview' | 'settings'>('overview')
  const [cosmetics, setCosmetics] = useState<CosmeticCatalogItem[]>([])
  const [frame, setFrame] = useState('none')
  const [effect, setEffect] = useState('none')
  const [buying, setBuying] = useState(false)
  const [shopError, setShopError] = useState<string | null>(null)
  const [shopMsg, setShopMsg] = useState<string | null>(null)

  // Relic inventory + gift state
  const [inventory, setInventory] = useState<InventoryItem[]>([])
  const [crewMembers, setCrewMembers] = useState<CrewMember[]>([])
  const [giftRelic, setGiftRelic] = useState<InventoryItem | null>(null)
  const [giftRecipient, setGiftRecipient] = useState<string>('')
  const [gifting, setGifting] = useState(false)
  const [giftError, setGiftError] = useState<string | null>(null)

  const loadCosmetics = useCallback(async () => {
    try {
      const data = await apiClient.get<CosmeticsResponse>('/api/cosmetics')
      setCosmetics(data.items ?? [])
      setFrame(data.avatar_frame || 'none')
      setEffect(data.explorer_effect || 'none')
    } catch (e) {
      console.error('failed to load cosmetics', e)
    }
  }, [])

  const loadInventory = useCallback(async () => {
    try {
      const data = await apiClient.get<InventoryItem[]>('/api/relics/inventory')
      setInventory(data ?? [])
    } catch (e) {
      console.error('failed to load inventory', e)
    }
  }, [])

  const loadCrewMembers = useCallback(async () => {
    try {
      const members = await crewsApi.members()
      setCrewMembers(members as unknown as CrewMember[])
    } catch (e) {
      console.error('failed to load family members', e)
    }
  }, [])

  useEffect(() => {
    if (profile) {
      void loadCosmetics()
      void loadInventory()
      void loadCrewMembers()
      if (profile.avatar_frame) {
        setFrame(profile.avatar_frame)
      }
      if (profile.equipped_explorer_effect) {
        setEffect(profile.equipped_explorer_effect)
      }
    }
  }, [profile, loadCosmetics, loadInventory, loadCrewMembers])

  const purchaseGoldFrame = async (item: CosmeticCatalogItem) => {
    setBuying(true)
    setShopError(null)
    setShopMsg(null)
    try {
      const res = await apiClient.post<{
        status: string
        already_owned: boolean
        coins: number
        avatar_frame: string
        explorer_effect: string
      }>('/api/cosmetics/purchase', { cosmetic_id: item.id })
      if (item.kind === 'avatar_frame') {
        setFrame(res.avatar_frame || item.value)
      } else if (item.kind === 'explorer_effect') {
        setEffect(res.explorer_effect || item.value)
      }
      setShopMsg(res.already_owned ? 'Sudah terbuka — tidak dikenakan biaya.' : `Terbuka! Menghabiskan ${item.price} koin.`)
      await Promise.all([refreshProfile(), loadCosmetics()])
    } catch (e) {
      setShopError(e instanceof Error ? e.message : 'Gagal membeli')
    } finally {
      setBuying(false)
    }
  }

  const equipCosmetic = async (cosmeticId: string) => {
    setShopError(null)
    try {
      await apiClient.post('/api/cosmetics/equip', { cosmetic_id: cosmeticId })
      await Promise.all([refreshProfile(), loadCosmetics()])
      const item = cosmetics.find((c) => c.id === cosmeticId)
      if (cosmeticId === 'none') {
        setFrame('none')
        setEffect('none')
        setShopMsg('Bingkai dan efek dilepas.')
      } else if (item?.kind === 'avatar_frame') {
        setFrame(item.value)
        setShopMsg('Bingkai dipasang.')
      } else if (item?.kind === 'explorer_effect') {
        setEffect(item.value)
        setShopMsg('Efek dipasang.')
      } else {
        setShopMsg('Dipasang.')
      }
    } catch (e) {
      setShopError(e instanceof Error ? e.message : 'Gagal memasang')
    }
  }

  const doGiftRelic = async () => {
    if (!giftRelic || !giftRecipient) return
    setGifting(true)
    setGiftError(null)
    try {
      await apiClient.post<GiftRelicResult>('/api/relics/gift', {
        recipient_uid: giftRecipient,
        relic_slug: giftRelic.relic_slug,
      })
      setGiftRelic(null)
      setGiftRecipient('')
      await loadInventory()
    } catch (e) {
      setGiftError(e instanceof Error ? e.message : 'Gagal memberi hadiah')
    } finally {
      setGifting(false)
    }
  }

  if (loading) {
    return (
      <div className="flex h-64 w-full items-center justify-center max-w-4xl mx-auto">
        <div className="flex flex-col items-center gap-4 animate-pulse">
          <div className="text-4xl">👤</div>
          <p className="text-sm text-text-secondary">Memuat profil...</p>
        </div>
      </div>
    )
  }

  if (error || !profile) {
    return (
      <div className="flex flex-col gap-6 max-w-4xl mx-auto py-4">
        <header className="flex items-center justify-between">
          <Link to="/" className="text-sm font-medium text-text-secondary hover:text-text-primary transition-colors inline-flex items-center gap-2">
            <span>←</span> Beranda
          </Link>
        </header>
        <Card className="flex flex-col items-center justify-center gap-4 py-16 text-center border-accent-danger/30 bg-accent-danger/5">
          <p className="text-lg font-medium text-text-primary">
            {error || 'Profil tidak ditemukan. Silakan masuk untuk melihat profilmu.'}
          </p>
          <Link to="/" className="text-sm font-bold text-accent-magic hover:underline uppercase tracking-wider">
            Kembali ke Masuk
          </Link>
        </Card>
      </div>
    )
  }


  const role = profile.role || 'SEEKER'
  const mastery = getRoleMastery(role, profile.level)
  const xpPercent = Math.min(100, profile.xp % 100)

  return (
    <div className="flex flex-col gap-6 max-w-2xl mx-auto py-4 animate-in fade-in duration-500">
      <header className="flex items-center justify-between mb-2">
        {activeView === 'overview' ? (
          <Link to="/" className="text-sm font-medium text-text-secondary hover:text-text-primary transition-colors inline-flex items-center gap-2">
            <span>←</span> Beranda
          </Link>
        ) : (
          <button onClick={() => setActiveView('overview')} className="text-sm font-medium text-text-secondary hover:text-text-primary transition-colors inline-flex items-center gap-2 cursor-pointer">
            <span>←</span> Kembali ke Profil
          </button>
        )}
        <h1 className="font-heading text-2xl md:text-3xl text-text-primary">
          {activeView === 'overview' ? 'Profil' : 'Pengaturan'}
        </h1>
      </header>

      {activeView === 'overview' && (
        <>
          {/* Identity Hero */}
          <Card className="relative overflow-hidden p-6 border-border-subtle bg-surface-elevated/80 shadow-md">
            <div className="absolute top-0 right-0 p-8 opacity-10 blur-[2px] pointer-events-none transform translate-x-1/4 -translate-y-1/4 scale-150">
              <Avatar
                seed={profile.avatar_seed || profile.uid}
                style={profile.avatar_style || 'adventurer'}
                frame={frame}
                size="2xl"
              />
            </div>
            
            <div className="relative z-10 flex flex-col items-center text-center gap-4">
              <Avatar
                seed={profile.avatar_seed || profile.uid}
                style={profile.avatar_style || 'adventurer'}
                frame={frame}
                size="xl"
              />
              
              <div>
                <h2 className="font-heading text-3xl text-text-primary">{profile.explorer_name}</h2>
                <span className="text-accent-reward font-bold text-xs tracking-widest uppercase pb-1">{mastery.title}</span>
              </div>
              
              <div className="w-full bg-surface p-4 rounded-xl border border-border-subtle mt-2">
                <div className="flex justify-between text-xs font-bold uppercase tracking-wider mb-2">
                  <span className="text-accent-reward">Level {profile.level}</span>
                  <span className="text-text-secondary">{profile.xp} Poin</span>
                </div>
                <ProgressBar progress={xpPercent} colorClass="bg-accent-reward" />
              </div>
            </div>
          </Card>

          {/* Hub Navigation Grid */}
          <div className="grid grid-cols-2 gap-4 mt-2 mb-6">
            <Link to="/creative" className="flex flex-col items-center justify-center p-6 gap-3 text-center bg-surface-elevated border border-border-subtle rounded-xl hover:border-accent-magic transition-colors">
              <span className="text-4xl">👨‍👩‍👧‍👦</span>
              <h3 className="font-bold text-text-primary">Keluarga</h3>
              <p className="text-xs text-text-secondary">Aktivitas & Info</p>
            </Link>
            
            <Link to="/gallery" className="flex flex-col items-center justify-center p-6 gap-3 text-center bg-surface-elevated border border-border-subtle rounded-xl hover:border-accent-magic transition-colors">
              <span className="text-4xl">🎨</span>
              <h3 className="font-bold text-text-primary">Galeri</h3>
              <p className="text-xs text-text-secondary">Karya Kreatif</p>
            </Link>
            
            <Link to="/chests" className="flex flex-col items-center justify-center p-6 gap-3 text-center bg-surface-elevated border border-border-subtle rounded-xl hover:border-accent-magic transition-colors">
              <span className="text-4xl">🎁</span>
              <h3 className="font-bold text-text-primary">Peti Hadiah</h3>
              <p className="text-xs text-text-secondary">Buka Hadiah</p>
            </Link>
            
            <button onClick={() => setActiveView('settings')} className="flex flex-col items-center justify-center p-6 gap-3 text-center bg-surface-elevated border border-border-subtle rounded-xl hover:border-accent-magic transition-colors">
              <span className="text-4xl">⚙️</span>
              <h3 className="font-bold text-text-primary">Pengaturan</h3>
              <p className="text-xs text-text-secondary">Akun & Sistem</p>
            </button>
          </div>

          <Card className="p-6 flex flex-col items-center justify-center text-center bg-accent-reward/5 border-accent-reward/20">
             <span className="text-4xl mb-2">🪙</span>
             <h3 className="font-heading text-3xl font-bold text-accent-reward">{profile.coins ?? 0}</h3>
             <p className="text-sm font-semibold text-text-primary uppercase tracking-widest mt-1">Saldo Koin</p>
          </Card>

          <Card className="p-6" data-testid="cosmetic-shop">
            <h3 className="font-heading text-xl text-text-primary mb-1 flex items-center gap-2">
              <span>✨</span> Toko Kosmetik
            </h3>
            <p className="text-xs text-text-secondary mb-4">
              Gunakan koin untuk bingkai potret dan efek.
            </p>
            {shopError && (
              <p className="mb-3 text-sm text-accent-danger" data-testid="cosmetic-shop-error">{shopError}</p>
            )}
            {shopMsg && (
              <p className="mb-3 text-sm text-accent-nature" data-testid="cosmetic-shop-msg">{shopMsg}</p>
            )}
            <div className="flex flex-col gap-3">
              {cosmetics.map((item) => {
                const isFrame = item.kind === 'avatar_frame'
                const isEffect = item.kind === 'explorer_effect'
                const equippedFrame = isFrame ? (item.unlocked || frame === item.value ? item.value : 'none') : frame
                const equippedEffect = isEffect ? (item.unlocked || effect === item.value ? item.value : 'none') : 'none'
                return (
                <div
                  key={item.id}
                  className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 rounded-lg border border-border-subtle bg-surface p-4"
                  data-testid={`cosmetic-item-${item.id}`}
                >
                  <div className="flex items-center gap-3">
                    <Avatar
                      seed={profile.avatar_seed || profile.uid}
                      style={profile.avatar_style || 'adventurer'}
                      frame={equippedFrame}
                      effect={equippedEffect}
                      size="md"
                    />
                    <div>
                      <p className="font-medium text-text-primary">{item.name}</p>
                      <p className="text-xs text-text-secondary">{item.description}</p>
                      <p className="text-xs font-semibold text-accent-reward mt-1">
                        {item.unlocked ? 'Terbuka' : item.price === 0 ? 'Hadiah gratis' : `🔒 ${item.price} koin`}
                      </p>
                    </div>
                  </div>
                  <div className="flex gap-2 self-start sm:self-center">
                    {item.unlocked ? (
                      (isFrame && frame === item.value) || (isEffect && effect === item.value) ? (
                        <Button
                          size="sm"
                          variant="secondary"
                          onClick={() => void equipCosmetic('none')}
                          data-testid={`unequip-${item.id}`}
                        >
                          Lepas
                        </Button>
                      ) : (
                        <Button
                          size="sm"
                          variant="primary"
                          onClick={() => void equipCosmetic(item.id)}
                          data-testid={`equip-${item.id}`}
                        >
                          Pakai
                        </Button>
                      )
                    ) : (
                      <Button
                        size="sm"
                        variant="secondary"
                        isLoading={buying}
                        disabled={buying || (item.price > 0 && (profile.coins ?? 0) < item.price)}
                        onClick={() => void purchaseGoldFrame(item)}
                        data-testid={`buy-${item.id}`}
                      >
                        {item.price === 0 ? 'Klaim' : ((profile.coins ?? 0) < item.price ? 'Koin kurang' : `Beli ${item.price} 🪙`)}
                      </Button>
                    )}
                  </div>
                </div>
                )
              })}
              {cosmetics.length === 0 && (
                <p className="text-sm text-text-secondary italic">Memuat kosmetik…</p>
              )}
            </div>
          </Card>

          {inventory.length > 0 && (
            <Card className="p-6" data-testid="relic-inventory">
              <h3 className="font-heading text-xl text-text-primary mb-1 flex items-center gap-2">
                <span>🗝️</span> Brankas Relik
              </h3>
              <p className="text-xs text-text-secondary mb-4">
                Koleksi relik yang telah kamu kumpulkan.
              </p>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                {inventory.map((item) => (
                  <div
                    key={item.relic_slug}
                    className="flex items-center justify-between gap-3 rounded-lg border border-border-subtle bg-surface p-3"
                    data-testid={`relic-item-${item.relic_slug}`}
                  >
                    <div className="flex items-center gap-3 min-w-0">
                      <span className="text-2xl shrink-0">{item.image}</span>
                      <div className="min-w-0">
                        <p className="font-medium text-text-primary text-sm truncate">{item.name}</p>
                        <p className="text-[10px] text-text-secondary uppercase tracking-wider">
                          {item.rarity} · ×{item.owned_count}
                        </p>
                      </div>
                    </div>
                    {crewMembers.filter((m) => m.uid !== profile.uid).length > 0 && (
                      <Button
                        size="sm"
                        variant="ghost"
                        className="shrink-0 text-accent-magic border border-accent-magic/30 hover:bg-accent-magic/10"
                        onClick={() => {
                          setGiftRelic(item)
                          setGiftRecipient('')
                          setGiftError(null)
                        }}
                        data-testid={`gift-btn-${item.relic_slug}`}
                      >
                        Beri
                      </Button>
                    )}
                  </div>
                ))}
              </div>
            </Card>
          )}

          {/* Gift modal overlay for inventory */}
          {giftRelic && (
            <div
              className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
              data-testid="gift-modal"
              onClick={(e) => { if (e.target === e.currentTarget) { setGiftRelic(null) } }}
            >
              <div className="bg-surface-elevated border border-border-subtle rounded-xl shadow-2xl p-6 w-full max-w-sm flex flex-col gap-4">
                <h4 className="font-heading text-xl text-text-primary flex items-center gap-2">
                  🎁 Beri Hadiah Relik
                </h4>
                <div className="flex items-center gap-3 rounded-lg bg-surface border border-border-subtle p-3">
                  <span className="text-3xl">{giftRelic.image}</span>
                  <div>
                    <p className="font-medium text-text-primary">{giftRelic.name}</p>
                    <p className="text-xs text-text-secondary uppercase tracking-wider">{giftRelic.rarity} · ×{giftRelic.owned_count}</p>
                  </div>
                </div>
                <div>
                  <label htmlFor="gift-recipient" className="block text-xs font-bold uppercase tracking-wider text-text-secondary mb-2">
                    Pilih penerima
                  </label>
                  <select
                    id="gift-recipient"
                    value={giftRecipient}
                    onChange={(e) => setGiftRecipient(e.target.value)}
                    className="w-full rounded-lg border border-border-subtle bg-surface text-text-primary px-3 py-2 text-sm focus:outline-none focus:border-accent-magic"
                    data-testid="gift-recipient-select"
                  >
                    <option value="">— pilih keluarga —</option>
                    {crewMembers
                      .filter((m) => m.uid !== profile.uid)
                      .map((m) => (
                        <option key={m.uid} value={m.uid}>{m.explorer_name}</option>
                      ))
                    }
                  </select>
                </div>
                {giftError && (
                  <p className="text-sm text-accent-danger" data-testid="gift-error-msg">{giftError}</p>
                )}
                <div className="flex gap-3">
                  <Button
                    variant="secondary"
                    className="flex-1"
                    onClick={() => setGiftRelic(null)}
                    disabled={gifting}
                  >
                    Batal
                  </Button>
                  <Button
                    variant="primary"
                    className="flex-1"
                    onClick={() => void doGiftRelic()}
                    disabled={gifting || !giftRecipient}
                    isLoading={gifting}
                    data-testid="gift-confirm-btn"
                  >
                    Konfirmasi
                  </Button>
                </div>
              </div>
            </div>
          )}
        </>
      )}

      {activeView === 'settings' && (
        <div className="flex flex-col gap-6 animate-in fade-in slide-in-from-right-4">
          <Card className="p-6 flex flex-col gap-4">
             <h3 className="font-heading text-xl text-text-primary mb-2 flex items-center gap-2">
              <span>👤</span> Akun
            </h3>
            <div className="flex items-center justify-between border-b border-border-subtle/50 pb-4">
              <div>
                <p className="text-sm font-bold text-text-primary">Avatar Profil</p>
                <p className="text-xs text-text-secondary">Ubah gaya karakter acak</p>
              </div>
              <Button 
                variant="secondary" 
                size="sm"
                className="flex items-center gap-2"
                onClick={async () => {
                  const newSeed = Math.random().toString(36).substring(2, 10);
                  try {
                    await apiClient.patch('/api/me/avatar', { avatar_style: 'adventurer', avatar_seed: newSeed });
                    await refreshProfile()
                  } catch (e) {
                    console.error('failed to update avatar', e);
                  }
                }}
              >
                <Shuffle size={14} /> Acak Baru
              </Button>
            </div>
            
            <div className="flex items-center justify-between border-b border-border-subtle/50 pb-4">
              <div>
                <p className="text-sm font-bold text-text-primary">ID Penjelajah</p>
                <p className="text-xs text-text-secondary font-mono">{profile.uid}</p>
              </div>
            </div>
            
            <div className="py-2">
               <PushNotificationToggle />
            </div>
          </Card>
          
          <Button variant="danger" onClick={logout} className="w-full shadow-md">
            Keluar (Sign Out)
          </Button>
        </div>
      )}
    </div>
  )
}
