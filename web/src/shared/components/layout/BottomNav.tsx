import { Link, useLocation } from 'react-router-dom'
import { Home, Banknote, User, ShieldCheck } from 'lucide-react'
import { useSession } from '../../hooks/useSession'

export function BottomNav() {
  const location = useLocation()
  const { profile } = useSession()
  const isGuide = profile?.role === 'GUIDE'

  const navItems = [
    { label: 'Beranda', to: '/', icon: Home },
    { label: 'Pencairan Koin', to: '/shop', icon: Banknote },
    { label: 'Profil', to: '/profile', icon: User },
    ...(isGuide ? [{ label: 'Admin', to: '/admin', icon: ShieldCheck }] : []),
  ]

  return (
    <nav aria-label="Navigasi Utama" className="flex items-center justify-around py-2 px-2 safe-pb">
      {navItems.map((item) => {
        const Icon = item.icon
        const isActive = location.pathname === item.to
        return (
          <Link
            key={item.to}
            to={item.to}
            aria-current={isActive ? 'page' : undefined}
            className={`flex flex-col items-center gap-1 py-2 px-4 rounded-xl transition-colors min-w-[64px] ${
              isActive
                ? 'text-accent-magic bg-accent-magic/10 font-bold'
                : 'text-text-secondary hover:text-text-primary hover:bg-surface'
            }`}
          >
            <Icon size={20} className={isActive ? 'stroke-[2.5]' : 'stroke-2'} />
            <span className="text-[11px] font-semibold tracking-tight leading-none">{item.label}</span>
          </Link>
        )
      })}
    </nav>
  )
}
