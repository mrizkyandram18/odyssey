import { Link, useLocation } from 'react-router-dom'
import { Home, Sparkles, Banknote, User, ShieldCheck } from 'lucide-react'
import { useSession } from '../../hooks/useSession'

export function BottomNav() {
  const location = useLocation()
  const { profile } = useSession()
  const isAdmin = profile?.role === 'ADMIN' || profile?.role === 'GUIDE' || profile?.role === 'BUILDER'

  const navItems = isAdmin
    ? [
        { label: 'Admin Panel', to: '/admin', icon: ShieldCheck },
        { label: 'Profil', to: '/profile', icon: User },
      ]
    : [
        { label: 'Beranda', to: '/', icon: Home },
        { label: 'Koleksi', to: '/koleksi', icon: Sparkles },
        { label: 'Pencairan Koin', to: '/shop', icon: Banknote },
        { label: 'Profil', to: '/profile', icon: User },
      ]

  return (
    <nav aria-label="Navigasi Utama" className="flex items-center justify-around py-1.5 px-2 safe-pb">
      {navItems.map((item) => {
        const Icon = item.icon
        const isActive = location.pathname === item.to
        return (
          <Link
            key={item.to}
            to={item.to}
            aria-current={isActive ? 'page' : undefined}
            className={`flex flex-col items-center justify-center gap-1 py-1.5 px-2.5 rounded-2xl transition-all duration-200 flex-1 max-w-[90px] min-h-[48px] ${
              isActive
                ? 'text-accent-magic bg-accent-magic/10 font-bold shadow-sm'
                : 'text-text-secondary hover:text-text-primary hover:bg-surface/60'
            }`}
          >
            <div className={`p-1 rounded-xl transition-transform ${isActive ? 'scale-110' : ''}`}>
              <Icon size={19} className={isActive ? 'stroke-[2.5]' : 'stroke-2'} />
            </div>
            <span className="text-[10.5px] tracking-tight leading-none text-center truncate max-w-full">
              {item.label}
            </span>
          </Link>
        )
      })}
    </nav>
  )
}
