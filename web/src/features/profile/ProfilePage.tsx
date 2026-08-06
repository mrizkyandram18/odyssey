import { Link } from 'react-router-dom'
import { ExplorerIcon } from '../../shared/components/atoms/ExplorerIcon'
import { Badge } from '../../shared/components/atoms/Badge'
import { Button } from '../../shared/components/atoms/Button'
import { useSession } from '../../shared/hooks/useSession'

const ROLE_DESCRIPTIONS: Record<string, string> = {
  SEEKER: 'Curious explorer who seeks out hidden details, riddles, and lore.',
  BUILDER: 'Creative craftsman who solves puzzles and constructs family artifacts.',
  GUIDE: 'Seasoned mentor who helps coordinate crew quests and validates discoveries.',
}

export function ProfilePage() {
  const { profile, session, loading, logout } = useSession()

  if (loading) {
    return (
      <div className="flex flex-col gap-4 p-4 pb-safe">
        <Link to="/" className="text-sm text-muted-foreground">← Home</Link>
        <p className="text-sm text-muted-foreground">Loading profile...</p>
      </div>
    )
  }

  if (!profile) {
    return (
      <div className="flex flex-col gap-4 p-4 pb-safe">
        <Link to="/" className="text-sm text-muted-foreground">← Home</Link>
        <p className="text-sm text-muted-foreground">No profile data available.</p>
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

      <Button variant="ghost" onClick={logout} className="w-full text-error border border-error/50 hover:bg-error/10">
        Sign Out
      </Button>
    </div>
  )
}
