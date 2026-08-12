import { Link, useLocation } from 'react-router-dom'
import { Home, Scroll, Book, User, Compass } from 'lucide-react'
import { useSession } from '../../hooks/useSession'
import { ProgressBar } from '../atoms/ProgressBar'
import { Avatar } from '../atoms/Avatar'
import { getRoleMastery } from '../../utils/roleMastery'

export function Sidebar() {
  const { profile } = useSession()
  const location = useLocation()

  const navItems = [
    { label: 'Beranda', path: '/', icon: <Home size={20} /> },
    { label: 'Misi', path: '/quests', icon: <Scroll size={20} /> },
    { label: 'Jurnal', path: '/journal', icon: <Book size={20} /> },
    { label: 'Profil', path: '/profile', icon: <User size={20} /> },
  ]

  const xpPercent = profile ? Math.min(100, (profile.xp % 100)) : 0

  return (
    <div className="flex flex-col h-full py-6 px-4">
      <div className="flex items-center gap-2 mb-8 px-2">
        <div className="flex h-8 w-8 items-center justify-center rounded-full bg-surface border border-accent-magic/30 text-accent-magic">
          <Compass size={18} />
        </div>
        <h1 className="font-heading text-xl text-text-primary tracking-wide">Odyssey</h1>
      </div>

      {profile && (
        <div className="mb-8 px-2 flex flex-col items-center text-center">
          <div className="w-20 h-20 rounded-full border-2 border-accent-nature p-1 bg-surface-glass mb-3 relative overflow-hidden">
            <Avatar
              seed={profile.avatar_seed || profile.uid}
              style={profile.avatar_style || 'adventurer'}
              frame={profile.avatar_frame || 'none'}
              size="xl"
            />
          </div>
          <h2 className="font-semibold text-lg text-text-primary">{profile.explorer_name}</h2>
          <p className="text-xs text-text-secondary mb-1">Level {profile.level} &bull; {getRoleMastery(profile.role, profile.level).title}</p>
          <p className="text-xs font-semibold text-accent-reward mb-3 tabular-nums" data-testid="sidebar-coin-balance">
            🪙 {profile.coins ?? 0} koin
          </p>
          
          <div className="w-full relative group">
            <ProgressBar progress={xpPercent} colorClass="bg-accent-nature" />
            <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
              <span className="text-[10px] font-bold text-white shadow-black drop-shadow-md">{xpPercent}% Poin</span>
            </div>
          </div>
        </div>
      )}

      <nav className="flex-1 space-y-1">
        {navItems.map((item) => {
          // Simple active check. In real app, might want to match sub-routes too.
          const isActive = location.pathname === item.path || (item.path !== '/' && location.pathname.startsWith(item.path))
          
          return (
            <Link
              key={item.path}
              to={item.path}
              className={`flex items-center gap-3 px-3 py-3 rounded-lg transition-all duration-200 ${
                isActive 
                  ? 'bg-accent-magic/10 text-accent-magic border border-accent-magic/20 shadow-[inset_0_0_12px_rgba(6,182,222,0.1)]' 
                  : 'text-text-secondary hover:text-text-primary hover:bg-surface-glass border border-transparent'
              }`}
            >
              <span className="text-lg">{item.icon}</span>
              <span className="font-medium text-sm">{item.label}</span>
            </Link>
          )
        })}
      </nav>
      
      <div className="mt-auto pt-4 px-2">
        <div className="rounded-lg bg-surface-glass border border-border-subtle p-3">
          <p className="text-xs text-text-secondary text-center">Keluargamu sedang menunggu.</p>
        </div>
      </div>
    </div>
  )
}
