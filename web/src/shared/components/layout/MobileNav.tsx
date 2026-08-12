import { Link, useLocation } from 'react-router-dom'
import { Home, Scroll, BookOpen, User, MoreHorizontal, Image, Box, Users } from 'lucide-react'
import { useState, useRef, useEffect } from 'react'

export function MobileNav() {
  const location = useLocation()
  const [showMore, setShowMore] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setShowMore(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  // Simplified nav for mobile
  const navItems = [
    { label: 'Beranda', path: '/', icon: <Home size={20} /> },
    { label: 'Misi', path: '/quests', icon: <Scroll size={20} /> },
    { label: 'Jurnal', path: '/journal', icon: <BookOpen size={20} /> },
    { label: 'Profil', path: '/profile', icon: <User size={20} /> },
  ]

  return (
    <div className="relative">
      {/* Secondary Menu Popup */}
      {showMore && (
        <div 
          ref={menuRef}
          className="absolute bottom-[100%] right-2 mb-2 w-48 bg-surface-elevated border border-border-subtle rounded-xl shadow-xl overflow-hidden animate-in fade-in slide-in-from-bottom-2 z-50"
        >
          <div className="flex flex-col py-2">
            <Link to="/creative" onClick={() => setShowMore(false)} className="flex items-center gap-3 px-4 py-3 text-sm text-text-secondary hover:bg-surface hover:text-text-primary">
              <Users size={18} /> Keluarga
            </Link>
            <Link to="/gallery" onClick={() => setShowMore(false)} className="flex items-center gap-3 px-4 py-3 text-sm text-text-secondary hover:bg-surface hover:text-text-primary">
              <Image size={18} /> Galeri
            </Link>
            <Link to="/chests" onClick={() => setShowMore(false)} className="flex items-center gap-3 px-4 py-3 text-sm text-text-secondary hover:bg-surface hover:text-text-primary">
              <Box size={18} /> Peti Hadiah
            </Link>
          </div>
        </div>
      )}
      
      <nav className="flex items-center justify-around px-2 py-2 relative z-40 bg-surface-elevated">
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
        
        {/* Lainnya Button */}
        <button
          onClick={() => setShowMore(!showMore)}
          className={`flex flex-col items-center justify-center p-2 min-w-[60px] rounded-lg transition-all ${
            showMore 
              ? 'text-accent-magic' 
              : 'text-text-secondary hover:text-text-primary'
          }`}
        >
          <span className="text-xl mb-1 transition-transform">
            <MoreHorizontal size={20} />
          </span>
          <span className="text-[10px] font-medium">Lainnya</span>
        </button>
      </nav>
    </div>
  )
}
