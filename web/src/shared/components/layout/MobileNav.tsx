import { Link, useLocation } from 'react-router-dom'
import { Home, Scroll, BookOpen, Palette, Sparkles, User, Image } from 'lucide-react'
import { useSession } from '../../hooks/useSession'

export function MobileNav() {
  const location = useLocation()
  const { profile } = useSession()

  // Simplified nav for mobile
  let navItems = [
    { label: 'Beranda', path: '/', icon: <Home size={20} /> },
    { label: 'Petualangan', path: '/quests', icon: <Scroll size={20} /> },
    { label: 'Jurnal', path: '/journal', icon: <BookOpen size={20} /> },
    { label: 'Kenangan', path: '/creative', icon: <Palette size={20} /> },
    { label: 'Galeri', path: '/gallery', icon: <Image size={20} /> },
    { label: 'Koleksi', path: '/relics', icon: <Sparkles size={20} /> },
    { label: 'Profil', path: '/profile', icon: <User size={20} /> },
  ]

  if (profile?.role !== 'GUIDE') {
    navItems = navItems.filter((item) => item.path !== '/relics')
  }

  return (
    <nav className="flex items-center justify-around px-2 py-2">
      {navItems.map((item) => {
        const isActive = location.pathname === item.path || (item.path !== '/' && location.pathname.startsWith(item.path))
        
        return (
          <Link
            key={item.path}
            to={item.path}
            className={`flex flex-col items-center justify-center p-2 min-w-[60px] rounded-lg transition-all ${
              isActive 
                ? 'text-accent-magic' 
                : 'text-text-secondary hover:text-text-primary'
            }`}
          >
            <span className={`text-xl mb-1 ${isActive ? 'scale-110 drop-shadow-[0_0_8px_rgba(6,182,222,0.5)]' : ''} transition-transform`}>
              {item.icon}
            </span>
            <span className="text-[10px] font-medium">{item.label}</span>
          </Link>
        )
      })}
    </nav>
  )
}
