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
  RewardLedgerEntry,
  InventoryItem,
  CrewMember,
  GiftRelicResult,
} from '../../shared/types'
import { Avatar } from '../../shared/components/atoms/Avatar'
import { Shuffle } from 'lucide-react'

import { getRoleMastery } from '../../shared/utils/roleMastery'
import { PushNotificationToggle } from '../../shared/components/molecules/PushNotificationToggle'

export function ProfilePage() {
  const { profile, session, loading, error, logout, refreshProfile } = useSession()
  const [ledgers, setLedgers] = useState<RewardLedgerEntry[]>([])
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
  const [giftMsg, setGiftMsg] = useState<string | null>(null)
  const [giftError, setGiftError] = useState<string | null>(null)

  // Crew customization state
  const [bannerUrl, setBannerUrl] = useState('')
  const [theme, setTheme] = useState('default')
  const [savingCrew, setSavingCrew] = useState(false)
  const [crewSaveMsg, setCrewSaveMsg] = useState<string | null>(null)
  const [crewSaveError, setCrewSaveError] = useState<string | null>(null)

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

  const loadCrew = useCallback(async () => {
    try {
      const data = await crewsApi.get()
      setBannerUrl(data.banner_url || '')
      setTheme(data.theme || 'default')
    } catch (e) {
      console.error('failed to load crew', e)
    }
  }, [])

  const saveCrewSettings = useCallback(async () => {
    setSavingCrew(true)
    setCrewSaveError(null)
    setCrewSaveMsg(null)
    try {
      const data = await crewsApi.patch({
        banner_url: bannerUrl,
        theme: theme,
      })
      setBannerUrl(data.banner_url || '')
      setTheme(data.theme || 'default')
      setCrewSaveMsg('Kustomisasi keluarga berhasil disimpan!')
    } catch (e) {
      setCrewSaveError(e instanceof Error ? e.message : 'Gagal menyimpan')
    } finally {
      setSavingCrew(false)
    }
  }, [bannerUrl, theme])

  useEffect(() => {
    if (profile) {
      apiClient.get<RewardLedgerEntry[]>('/api/rewards')
        .then(data => setLedgers(data ?? []))
        .catch(console.error)
      void loadCosmetics()
      void loadInventory()
      void loadCrewMembers()
      void loadCrew()
      if (profile.avatar_frame) {
        setFrame(profile.avatar_frame)
      }
      if (profile.equipped_explorer_effect) {
        setEffect(profile.equipped_explorer_effect)
      }
    }
  }, [profile, loadCosmetics, loadInventory, loadCrewMembers, loadCrew])

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
      const led = await apiClient.get<RewardLedgerEntry[]>('/api/rewards')
      setLedgers(led)
    } catch (e) {
      setShopError(e instanceof Error ? e.message : 'Gagal membeli')
    } finally {
      setBuying(false)
    }
  }

  const equipCosmetic = async (cosmeticId: string) => {
    setShopError(null)
    setShopMsg(null)
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
    setGiftMsg(null)
    try {
      const res = await apiClient.post<GiftRelicResult>('/api/relics/gift', {
        recipient_uid: giftRecipient,
        relic_slug: giftRelic.relic_slug,
      })
      const recipientName = crewMembers.find((m) => m.uid === giftRecipient)?.explorer_name ?? 'mereka'
      setGiftMsg(`🎁 ${res.relic_name} diberikan kepada ${recipientName}! Kamu masih memiliki ${res.sender_remaining_count}.`)
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
    <div className="flex flex-col gap-6 max-w-4xl mx-auto py-4 animate-in fade-in duration-500">
      <header className="flex items-center justify-between mb-2">
        <Link to="/" className="text-sm font-medium text-text-secondary hover:text-text-primary transition-colors inline-flex items-center gap-2">
          <span>←</span> Beranda
        </Link>
        <h1 className="font-heading text-2xl md:text-3xl text-text-primary">Profil</h1>
      </header>

      {/* Identity Hero */}
      <Card className="relative overflow-hidden p-8 border-accent-magic/30 bg-surface-elevated/80 shadow-[0_0_30px_rgba(6,182,222,0.1)]">
        <div className="absolute top-0 right-0 p-8 opacity-10 blur-[2px] pointer-events-none transform translate-x-1/4 -translate-y-1/4 scale-150">
          <Avatar
            seed={profile.avatar_seed || profile.uid}
            style={profile.avatar_style || 'adventurer'}
            frame={frame}
            size="2xl"
          />
        </div>
        
        <div className="relative z-10 flex flex-col md:flex-row items-center md:items-start gap-8">
          <div className="flex flex-col items-center gap-3 shrink-0">
            <Avatar
              seed={profile.avatar_seed || profile.uid}
              style={profile.avatar_style || 'adventurer'}
              frame={frame}
              size="xl"
            />
            <Button 
              variant="ghost" 
              className="text-xs py-1 px-3 flex items-center gap-1 opacity-70 hover:opacity-100"
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
              <Shuffle size={12} /> Acak
            </Button>
          </div>
          
          <div className="flex-1 text-center md:text-left">
            <div className="flex flex-col md:flex-row md:items-end gap-3 mb-2">
              <h2 className="font-heading text-4xl text-text-primary">{profile.explorer_name}</h2>
              <span className="text-accent-magic font-bold text-sm tracking-widest uppercase pb-1">{mastery.title}</span>
            </div>
            
            <p className="text-text-secondary text-sm mb-6 max-w-md mx-auto md:mx-0">
              {mastery.flavor}
            </p>
            
            <div className="w-full max-w-md bg-surface p-4 rounded-lg border border-border-subtle">
              <div className="flex justify-between text-xs font-bold uppercase tracking-wider mb-2">
                <span className="text-accent-reward">Level {profile.level}</span>
                <span className="text-text-secondary">{profile.xp} Poin</span>
              </div>
              <ProgressBar progress={xpPercent} colorClass="bg-accent-reward shadow-[0_0_10px_rgba(245,158,11,0.5)]" />
              <div
                className="mt-3 flex items-center justify-between rounded-md border border-accent-reward/20 bg-accent-reward/5 px-3 py-2"
                data-testid="profile-coin-balance"
              >
                <span className="text-xs font-bold uppercase tracking-wider text-text-secondary">Koin</span>
                <span className="text-sm font-bold text-accent-reward tabular-nums">
                  🪙 {profile.coins ?? 0}
                </span>
              </div>
            </div>
          </div>
        </div>
      </Card>

      {/* Slice 2.2 — first spend path */}
      <Card className="p-6" data-testid="cosmetic-shop">
        <h3 className="font-heading text-xl text-text-primary mb-1 flex items-center gap-2">
          <span>✨</span> Toko Kosmetik
        </h3>
        <p className="text-xs text-text-secondary mb-4">
          Gunakan koin untuk bingkai potret. Harga tetap · hanya koin fiktif.
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
                    {item.price === 0 ? 'Klaim' : ((profile.coins ?? 0) < item.price ? 'Koin tidak cukup' : `Beli seharga ${item.price} 🪙`)}
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

      {/* Relic Inventory + Gifting — Slice 2.11 */}
      {inventory.length > 0 && (
        <Card className="p-6" data-testid="relic-inventory">
          <h3 className="font-heading text-xl text-text-primary mb-1 flex items-center gap-2">
            <span>🗝️</span> Brankas Relik
          </h3>
          <p className="text-xs text-text-secondary mb-4">
            Berikan relik ke keluarga · gratis · tanpa potong koin
          </p>
          {giftMsg && (
            <p className="mb-3 text-sm text-accent-nature" data-testid="gift-success-msg">{giftMsg}</p>
          )}
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
                      setGiftMsg(null)
                    }}
                    data-testid={`gift-btn-${item.relic_slug}`}
                  >
                    🎁 Beri Hadiah
                  </Button>
                )}
              </div>
            ))}
          </div>
        </Card>
      )}

      {/* Gift modal overlay */}
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
                Konfirmasi Hadiah
              </Button>
            </div>
          </div>
        </div>
      )}

      <PushNotificationToggle />

      {/* Crew Customization — Slice 4.4 */}
      <Card className="p-6" data-testid="crew-customization">
        <h3 className="font-heading text-xl text-text-primary mb-1 flex items-center gap-2">
          <span>⚓</span> Kustomisasi Keluarga
        </h3>
        <p className="text-xs text-text-secondary mb-4">
          Atur URL banner dan tema bersama untuk keluarga kalian.
        </p>
        {crewSaveError && (
          <p className="mb-3 text-sm text-accent-danger" data-testid="crew-save-error">{crewSaveError}</p>
        )}
        {crewSaveMsg && (
          <p className="mb-3 text-sm text-accent-nature" data-testid="crew-save-msg">{crewSaveMsg}</p>
        )}
        <div className="flex flex-col gap-4">
          <div>
            <label htmlFor="banner-url" className="block text-xs font-bold uppercase tracking-wider text-text-secondary mb-2">
              URL Banner
            </label>
            <input
              id="banner-url"
              type="url"
              value={bannerUrl}
              onChange={(e) => setBannerUrl(e.target.value)}
              placeholder="https://example.com/banner.png"
              className="w-full rounded-lg border border-border-subtle bg-surface text-text-primary px-3 py-2 text-sm focus:outline-none focus:border-accent-magic"
              data-testid="banner-url-input"
            />
            {bannerUrl && (
              <div className="mt-3 rounded-lg border border-border-subtle overflow-hidden h-32 bg-surface">
                <img
                  src={bannerUrl}
                  alt="Crew banner preview"
                  className="w-full h-full object-cover"
                  data-testid="banner-preview"
                />
              </div>
            )}
          </div>
          <div>
            <label htmlFor="theme-select" className="block text-xs font-bold uppercase tracking-wider text-text-secondary mb-2">
              Tema Bersama
            </label>
            <select
              id="theme-select"
              value={theme}
              onChange={(e) => setTheme(e.target.value)}
              className="w-full rounded-lg border border-border-subtle bg-surface text-text-primary px-3 py-2 text-sm focus:outline-none focus:border-accent-magic"
              data-testid="theme-select"
            >
              <option value="default">Default (Midnight)</option>
              <option value="forest">Forest (Emerald)</option>
              <option value="city">City (Cyan)</option>
              <option value="library">Library (Violet)</option>
            </select>
          </div>
          <Button
            variant="primary"
            onClick={() => void saveCrewSettings()}
            isLoading={savingCrew}
            data-testid="save-crew-settings"
          >
            Simpan Pengaturan Keluarga
          </Button>
        </div>
      </Card>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Crew Info */}
        <Card className="p-6">
          <h3 className="font-heading text-xl text-text-primary mb-6 flex items-center gap-2">
            <span>🛡️</span> Keluarga
          </h3>
          <div className="space-y-4">
            <div className="flex justify-between items-center py-2 border-b border-border-subtle/50">
              <span className="text-sm text-text-secondary font-medium">Identifikasi Keluarga</span>
              <span className="font-mono text-sm bg-surface-elevated px-2 py-1 rounded text-text-primary">
                {profile.crew_id || session?.crew_id || 'Keluarga Bersama'}
              </span>
            </div>
            <div className="flex justify-between items-center py-2 border-b border-border-subtle/50">
              <span className="text-sm text-text-secondary font-medium">ID Penjelajah</span>
              <span className="font-mono text-sm bg-surface-elevated px-2 py-1 rounded text-text-primary">
                {profile.uid}
              </span>
            </div>
            <div className="flex justify-between items-center py-2 border-b border-border-subtle/50">
              <span className="text-sm text-text-secondary font-medium">Saldo Koin</span>
              <span className="text-sm font-bold text-accent-reward tabular-nums">🪙 {profile.coins ?? 0}</span>
            </div>
          </div>
        </Card>

        {/* Ledger — earn + spend */}
        <Card className="p-6 flex flex-col max-h-[300px]">
          <h3 className="font-heading text-xl text-text-primary mb-1 flex items-center gap-2">
            <span>🪙</span> Catatan Koin
          </h3>
          <p className="text-xs text-text-secondary mb-4">
            Misi +5 · Harian +1 · Bingkai emas −3. Hanya koin fiktif.
          </p>
          
          <div className="flex-1 overflow-y-auto pr-2 custom-scrollbar space-y-3">
            {ledgers.length === 0 ? (
              <p className="text-sm text-text-secondary italic text-center py-8">Belum ada aktivitas koin.</p>
            ) : (
              ledgers.slice(0, 15).map((l) => (
                <div key={l.id} className="flex justify-between items-center bg-surface p-3 rounded border border-border-subtle hover:border-accent-reward/30 transition-colors">
                  <div className="flex flex-col">
                    <span className="text-sm font-medium text-text-primary capitalize">
                      {l.source.replace(/_/g, ' ').toLowerCase()}
                    </span>
                    <span className="text-[10px] text-text-secondary uppercase tracking-wider">
                      {new Date(l.created_at).toLocaleDateString()}
                    </span>
                  </div>
                  <div className={`text-sm font-bold px-2 py-1 rounded ${
                    l.amount < 0
                      ? 'text-accent-danger bg-accent-danger/10'
                      : 'text-accent-reward bg-accent-reward/10'
                  }`}>
                    {l.amount > 0 ? '+' : ''}{l.amount} {l.reward_type === 'COINS' ? '🪙' : l.reward_type}
                  </div>
                </div>
              ))
            )}
          </div>
        </Card>
      </div>

      <div className="mt-8 flex justify-center">
        <Button variant="danger" onClick={logout} className="w-full max-w-xs shadow-lg shadow-accent-danger/20">
          Keluar (Sign Out)
        </Button>
      </div>
    </div>
  )
}
