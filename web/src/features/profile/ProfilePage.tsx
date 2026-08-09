import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '../../shared/components/atoms/Button'
import { ProgressBar } from '../../shared/components/atoms/ProgressBar'
import { Card } from '../../shared/components/atoms/Card'
import { useSession } from '../../shared/hooks/useSession'
import { apiClient } from '../../shared/lib/api'
import type { RewardLedgerEntry } from '../../shared/types'
import { Avatar } from '../../shared/components/atoms/Avatar'
import { Shuffle } from 'lucide-react'

const ROLE_DESCRIPTIONS: Record<string, string> = {
  SEEKER: 'Curious explorer who seeks out hidden details, riddles, and lore.',
  BUILDER: 'Creative craftsman who solves puzzles and constructs family artifacts.',
  GUIDE: 'Seasoned mentor who helps coordinate crew quests and validates discoveries.',
}

export function ProfilePage() {
  const { profile, session, loading, error, logout } = useSession()
  const [ledgers, setLedgers] = useState<RewardLedgerEntry[]>([])

  useEffect(() => {
    if (profile) {
      apiClient.get<RewardLedgerEntry[]>('/api/rewards')
        .then(setLedgers)
        .catch(console.error)
    }
  }, [profile])

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
          <Avatar seed={profile.avatar_seed || profile.uid} style={profile.avatar_style || 'adventurer'} size="2xl" />
        </div>
        
        <div className="relative z-10 flex flex-col md:flex-row items-center md:items-start gap-8">
          <div className="flex flex-col items-center gap-3 shrink-0">
            <Avatar seed={profile.avatar_seed || profile.uid} style={profile.avatar_style || 'adventurer'} size="xl" />
            <Button 
              variant="ghost" 
              className="text-xs py-1 px-3 flex items-center gap-1 opacity-70 hover:opacity-100"
              onClick={async () => {
                const newSeed = Math.random().toString(36).substring(2, 10);
                try {
                  await apiClient.patch('/api/me/avatar', { avatar_style: 'adventurer', avatar_seed: newSeed });
                  // Option 1: useSession refreshProfile (if available)
                  // Let's just reload the page for now to get fresh data everywhere easily, or we can use refreshProfile
                  window.location.reload(); 
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
            </div>
          </div>
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
          </div>
        </Card>

        {/* Ledger */}
        <Card className="p-6 flex flex-col max-h-[300px]">
          <h3 className="font-heading text-xl text-text-primary mb-4 flex items-center gap-2">
            <span>🪙</span> Spoils
          </h3>
          
          <div className="flex-1 overflow-y-auto pr-2 custom-scrollbar space-y-3">
            {ledgers.length === 0 ? (
              <p className="text-sm text-text-secondary italic text-center py-8">No treasures claimed yet.</p>
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
                  <div className="text-sm font-bold text-accent-reward bg-accent-reward/10 px-2 py-1 rounded">
                    +{l.amount} {l.reward_type === 'COINS' ? '🪙' : l.reward_type}
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
