import { useState, useEffect, useCallback } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '../../shared/components/atoms/Button'
import { ProgressBar } from '../../shared/components/atoms/ProgressBar'
import { Card } from '../../shared/components/atoms/Card'
import { useSession } from '../../shared/hooks/useSession'
import { apiClient } from '../../shared/lib/api'
import type { CosmeticCatalogItem, CosmeticsResponse, RewardLedgerEntry } from '../../shared/types'
import { Avatar } from '../../shared/components/atoms/Avatar'
import { Shuffle } from 'lucide-react'

const ROLE_DESCRIPTIONS: Record<string, string> = {
  SEEKER: 'Curious explorer who seeks out hidden details, riddles, and lore.',
  BUILDER: 'Creative craftsman who solves puzzles and constructs family artifacts.',
  GUIDE: 'Seasoned mentor who helps coordinate crew quests and validates discoveries.',
}

export function ProfilePage() {
  const { profile, session, loading, error, logout, refreshProfile } = useSession()
  const [ledgers, setLedgers] = useState<RewardLedgerEntry[]>([])
  const [cosmetics, setCosmetics] = useState<CosmeticCatalogItem[]>([])
  const [frame, setFrame] = useState('none')
  const [buying, setBuying] = useState(false)
  const [shopError, setShopError] = useState<string | null>(null)
  const [shopMsg, setShopMsg] = useState<string | null>(null)

  const loadCosmetics = useCallback(async () => {
    try {
      const data = await apiClient.get<CosmeticsResponse>('/api/cosmetics')
      setCosmetics(data.items ?? [])
      setFrame(data.avatar_frame || 'none')
    } catch (e) {
      console.error('failed to load cosmetics', e)
    }
  }, [])

  useEffect(() => {
    if (profile) {
      apiClient.get<RewardLedgerEntry[]>('/api/rewards')
        .then(setLedgers)
        .catch(console.error)
      void loadCosmetics()
      if (profile.avatar_frame) {
        setFrame(profile.avatar_frame)
      }
    }
  }, [profile, loadCosmetics])

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
      }>('/api/cosmetics/purchase', { cosmetic_id: item.id })
      setFrame(res.avatar_frame || item.value)
      setShopMsg(res.already_owned ? 'Already unlocked — no charge.' : `Unlocked! Spent ${item.price} coins.`)
      await Promise.all([refreshProfile(), loadCosmetics()])
      const led = await apiClient.get<RewardLedgerEntry[]>('/api/rewards')
      setLedgers(led)
    } catch (e) {
      setShopError(e instanceof Error ? e.message : 'Purchase failed')
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
      setShopMsg(cosmeticId === 'none' ? 'Frame unequipped.' : 'Frame equipped.')
    } catch (e) {
      setShopError(e instanceof Error ? e.message : 'Equip failed')
    }
  }

  if (loading) {
    return (
      <div className="flex h-64 w-full items-center justify-center max-w-4xl mx-auto">
        <div className="flex flex-col items-center gap-4 animate-pulse">
          <div className="text-4xl">👤</div>
          <p className="text-sm text-text-secondary">Summoning profile...</p>
        </div>
      </div>
    )
  }

  if (error || !profile) {
    return (
      <div className="flex flex-col gap-6 max-w-4xl mx-auto py-4">
        <header className="flex items-center justify-between">
          <Link to="/" className="text-sm font-medium text-text-secondary hover:text-text-primary transition-colors inline-flex items-center gap-2">
            <span>←</span> Home
          </Link>
        </header>
        <Card className="flex flex-col items-center justify-center gap-4 py-16 text-center border-accent-danger/30 bg-accent-danger/5">
          <p className="text-lg font-medium text-text-primary">
            {error || 'No active profile found. Please sign in to view your explorer profile.'}
          </p>
          <Link to="/" className="text-sm font-bold text-accent-magic hover:underline uppercase tracking-wider">
            Return to Sign In
          </Link>
        </Card>
      </div>
    )
  }

  const role = profile.role || 'SEEKER'
  const xpPercent = Math.min(100, profile.xp % 100)

  return (
    <div className="flex flex-col gap-6 max-w-4xl mx-auto py-4 animate-in fade-in duration-500">
      <header className="flex items-center justify-between mb-2">
        <Link to="/" className="text-sm font-medium text-text-secondary hover:text-text-primary transition-colors inline-flex items-center gap-2">
          <span>←</span> Home
        </Link>
        <h1 className="font-heading text-2xl md:text-3xl text-text-primary">Dossier</h1>
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
              <Shuffle size={12} /> Randomize
            </Button>
          </div>
          
          <div className="flex-1 text-center md:text-left">
            <div className="flex flex-col md:flex-row md:items-end gap-3 mb-2">
              <h2 className="font-heading text-4xl text-text-primary">{profile.explorer_name}</h2>
              <span className="text-accent-magic font-bold text-sm tracking-widest uppercase pb-1">{role}</span>
            </div>
            
            <p className="text-text-secondary text-sm mb-6 max-w-md mx-auto md:mx-0">
              {ROLE_DESCRIPTIONS[role] || ROLE_DESCRIPTIONS.SEEKER}
            </p>
            
            <div className="w-full max-w-md bg-surface p-4 rounded-lg border border-border-subtle">
              <div className="flex justify-between text-xs font-bold uppercase tracking-wider mb-2">
                <span className="text-accent-reward">Level {profile.level}</span>
                <span className="text-text-secondary">{profile.xp} XP</span>
              </div>
              <ProgressBar progress={xpPercent} colorClass="bg-accent-reward shadow-[0_0_10px_rgba(245,158,11,0.5)]" />
              <div
                className="mt-3 flex items-center justify-between rounded-md border border-accent-reward/20 bg-accent-reward/5 px-3 py-2"
                data-testid="profile-coin-balance"
              >
                <span className="text-xs font-bold uppercase tracking-wider text-text-secondary">Coins</span>
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
          <span>✨</span> Cosmetic shop
        </h3>
        <p className="text-xs text-text-secondary mb-4">
          Spend coins on a portrait frame. Fixed price · fictional currency only.
        </p>
        {shopError && (
          <p className="mb-3 text-sm text-accent-danger" data-testid="cosmetic-shop-error">{shopError}</p>
        )}
        {shopMsg && (
          <p className="mb-3 text-sm text-accent-nature" data-testid="cosmetic-shop-msg">{shopMsg}</p>
        )}
        <div className="flex flex-col gap-3">
          {cosmetics.map((item) => (
            <div
              key={item.id}
              className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 rounded-lg border border-border-subtle bg-surface p-4"
              data-testid={`cosmetic-item-${item.id}`}
            >
              <div className="flex items-center gap-3">
                <Avatar
                  seed={profile.avatar_seed || profile.uid}
                  style={profile.avatar_style || 'adventurer'}
                  frame={item.unlocked || frame === item.value ? item.value : 'none'}
                  size="md"
                />
                <div>
                  <p className="font-medium text-text-primary">{item.name}</p>
                  <p className="text-xs text-text-secondary">{item.description}</p>
                  <p className="text-xs font-semibold text-accent-reward mt-1">
                    {item.unlocked ? 'Unlocked' : `🔒 ${item.price} coins`}
                  </p>
                </div>
              </div>
              <div className="flex gap-2 self-start sm:self-center">
                {item.unlocked ? (
                  frame === item.value ? (
                    <Button
                      size="sm"
                      variant="secondary"
                      onClick={() => void equipCosmetic('none')}
                      data-testid={`unequip-${item.id}`}
                    >
                      Unequip
                    </Button>
                  ) : (
                    <Button
                      size="sm"
                      variant="primary"
                      onClick={() => void equipCosmetic(item.id)}
                      data-testid={`equip-${item.id}`}
                    >
                      Equip
                    </Button>
                  )
                ) : (
                  <Button
                    size="sm"
                    variant="secondary"
                    isLoading={buying}
                    disabled={buying || (profile.coins ?? 0) < item.price}
                    onClick={() => void purchaseGoldFrame(item)}
                    data-testid={`buy-${item.id}`}
                  >
                    {(profile.coins ?? 0) < item.price ? 'Need more coins' : `Buy for ${item.price} 🪙`}
                  </Button>
                )}
              </div>
            </div>
          ))}
          {cosmetics.length === 0 && (
            <p className="text-sm text-text-secondary italic">Loading cosmetics…</p>
          )}
        </div>
      </Card>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Crew Info */}
        <Card className="p-6">
          <h3 className="font-heading text-xl text-text-primary mb-6 flex items-center gap-2">
            <span>🛡️</span> Alliance
          </h3>
          <div className="space-y-4">
            <div className="flex justify-between items-center py-2 border-b border-border-subtle/50">
              <span className="text-sm text-text-secondary font-medium">Crew Identification</span>
              <span className="font-mono text-sm bg-surface-elevated px-2 py-1 rounded text-text-primary">
                {profile.crew_id || session?.crew_id || 'Shared Crew'}
              </span>
            </div>
            <div className="flex justify-between items-center py-2 border-b border-border-subtle/50">
              <span className="text-sm text-text-secondary font-medium">Explorer ID</span>
              <span className="font-mono text-sm bg-surface-elevated px-2 py-1 rounded text-text-primary">
                {profile.uid}
              </span>
            </div>
            <div className="flex justify-between items-center py-2 border-b border-border-subtle/50">
              <span className="text-sm text-text-secondary font-medium">Coin balance</span>
              <span className="text-sm font-bold text-accent-reward tabular-nums">🪙 {profile.coins ?? 0}</span>
            </div>
          </div>
        </Card>

        {/* Ledger — earn + spend */}
        <Card className="p-6 flex flex-col max-h-[300px]">
          <h3 className="font-heading text-xl text-text-primary mb-1 flex items-center gap-2">
            <span>🪙</span> Coin ledger
          </h3>
          <p className="text-xs text-text-secondary mb-4">
            Quest +5 · Daily +1 · Gold frame −3. Fictional coins only.
          </p>
          
          <div className="flex-1 overflow-y-auto pr-2 custom-scrollbar space-y-3">
            {ledgers.length === 0 ? (
              <p className="text-sm text-text-secondary italic text-center py-8">No coin activity yet.</p>
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
          Leave Realm (Sign Out)
        </Button>
      </div>
    </div>
  )
}
