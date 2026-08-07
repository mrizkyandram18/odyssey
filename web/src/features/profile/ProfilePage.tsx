import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { ExplorerIcon } from '../../shared/components/atoms/ExplorerIcon'
import { Badge } from '../../shared/components/atoms/Badge'
import { Button } from '../../shared/components/atoms/Button'
import { useSession } from '../../shared/hooks/useSession'
import { apiClient } from '../../shared/lib/api'
import type { RewardLedgerEntry } from '../../shared/types'

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
      <div className="flex flex-col gap-6 p-4 pb-safe">
        <header className="flex items-center justify-between">
          <Link to="/" className="text-sm text-muted-foreground hover:text-primary transition-colors">
            ← Home
          </Link>
          <h1 className="text-xl font-semibold">Explorer Profile</h1>
        </header>
        <div className="h-24 w-full animate-pulse rounded-lg border border-border bg-surface/50" />
        <div className="h-20 w-full animate-pulse rounded-lg border border-border bg-surface/50" />
        <div className="h-24 w-full animate-pulse rounded-lg border border-border bg-surface/50" />
      </div>
    )
  }

  if (error || !profile) {
    return (
      <div className="flex flex-col gap-4 p-4 pb-safe">
        <header className="flex items-center justify-between">
          <Link to="/" className="text-sm text-muted-foreground hover:text-primary transition-colors">
            ← Home
          </Link>
        </header>
        <div className="flex flex-col items-center justify-center gap-3 rounded-lg border border-border bg-surface p-6 text-center">
          <p className="text-sm text-muted-foreground">
            {error || 'No active profile found. Please sign in to view your explorer profile.'}
          </p>
          <Link to="/" className="text-xs font-semibold text-primary hover:underline">
            Return to Sign In
          </Link>
        </div>
      </div>
    )
  }

  const role = profile.role || 'SEEKER'
  const xpPercent = Math.min(100, profile.xp % 100)

  return (
    <div className="flex flex-col gap-6 p-4 pb-safe">
      <header className="flex items-center justify-between">
        <Link to="/" className="text-sm text-muted-foreground hover:text-primary transition-colors">
          ← Home
        </Link>
        <h1 className="text-xl font-semibold">Explorer Profile</h1>
      </header>

      <section className="flex items-center gap-4 rounded-lg border border-border bg-surface p-4">
        <ExplorerIcon role={role} size="lg" />
        <div className="flex-1 space-y-1">
          <h2 className="text-lg font-bold">{profile.explorer_name}</h2>
          <div className="flex items-center gap-2">
            <Badge variant="primary">Level {profile.level}</Badge>
            <Badge variant="secondary">{role}</Badge>
          </div>
        </div>
      </section>

      <section className="rounded-lg border border-border bg-surface p-4 space-y-3">
        <h3 className="text-sm font-medium text-muted-foreground">Experience & Progress</h3>
        <div>
          <div className="flex justify-between text-xs text-muted-foreground mb-1">
            <span>Progress to Level {profile.level + 1}</span>
            <span>{profile.xp} Total XP</span>
          </div>
          <div className="h-3 w-full rounded-full bg-border overflow-hidden">
            <div
              className="h-full bg-primary transition-all duration-300"
              style={{ width: `${xpPercent}%` }}
            />
          </div>
        </div>
      </section>

      <section className="rounded-lg border border-border bg-surface p-4 space-y-2">
        <h3 className="text-sm font-medium text-muted-foreground">Role Framing</h3>
        <p className="text-sm font-medium">{role}</p>
        <p className="text-xs text-muted-foreground">
          {ROLE_DESCRIPTIONS[role] || ROLE_DESCRIPTIONS.SEEKER}
        </p>
      </section>

      <section className="rounded-lg border border-border bg-surface p-4 space-y-2">
        <h3 className="text-sm font-medium text-muted-foreground">Identity & Crew</h3>
        <div className="flex justify-between text-sm">
          <span className="text-muted-foreground">User ID</span>
          <span className="font-mono text-xs">{profile.uid}</span>
        </div>
        <div className="flex justify-between text-sm">
          <span className="text-muted-foreground">Crew ID</span>
          <span className="font-mono text-xs">{profile.crew_id || session?.crew_id || 'Shared Crew'}</span>
        </div>
      </section>

      <section className="rounded-lg border border-border bg-surface p-4 space-y-3">
        <h3 className="text-sm font-medium text-muted-foreground">Reward History</h3>
        {ledgers.length === 0 ? (
          <p className="text-xs text-muted-foreground text-center py-2">No reward history yet.</p>
        ) : (
          <div className="flex flex-col gap-2">
            {ledgers.slice(0, 10).map((l) => (
              <div key={l.id} className="flex justify-between items-center border-b border-border pb-2 last:border-0 last:pb-0">
                <div className="flex flex-col">
                  <span className="text-sm">{l.source.replace(/_/g, ' ')}</span>
                  <span className="text-xs text-muted-foreground">{new Date(l.created_at).toLocaleDateString()}</span>
                </div>
                <div className="text-sm font-medium text-primary">
                  +{l.amount} {l.reward_type === 'COINS' ? '🪙' : l.reward_type}
                </div>
              </div>
            ))}
          </div>
        )}
      </section>

      <Button variant="ghost" onClick={logout} className="w-full text-error border border-error/50 hover:bg-error/10">
        Sign Out
      </Button>
    </div>
  )
}
