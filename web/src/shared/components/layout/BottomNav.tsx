import { Link, useLocation } from 'react-router-dom'
import { Home, ShoppingBag, User, ShieldCheck } from 'lucide-react'
import { useSession } from '../../hooks/useSession'

export function BottomNav() {
  const location = useLocation()
  const { profile } = useSession()
  const isGuide = profile?.role === 'GUIDE'

  const navItems = [
    { label: 'Beranda', to: '/', icon: Home },
    { label: 'Toko Hadiah', to: '/shop', icon: ShoppingBag },
    { label: 'Profil', to: '/profile', icon: User },
    ...(isGuide ? [{ label: 'Admin', to: '/admin', icon: ShieldCheck }] : []),
  ]

  return (
    <nav aria-label="Navigasi Utama" className="flex items-center justify-around py-2 px-4">
      {navItems.map((item) => {
        const Icon = item.icon
        const isActive = location.pathname === item.to
        return (
          <Link
            key={item.to}
            to={item.to}
            className={`flex flex-col items-center gap-1 py-1 px-3 rounded-xl transition-all duration-200 ${
              isActive
                ? 'text-accent-reward font-bold scale-105'
                : 'text-text-secondary hover:text-text-primary'
            }`}
          >
            <Icon size={20} className={isActive ? 'stroke-[2.5]' : 'stroke-2'} />
            <span className="text-[11px] font-medium tracking-tight">{item.label}</span>
          </Link>
        )
      })}
    </nav>
  )
}
