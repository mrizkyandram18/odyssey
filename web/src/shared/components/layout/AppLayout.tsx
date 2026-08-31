import type { ReactNode } from 'react'
import { useLocation } from 'react-router-dom'
import { motion, AnimatePresence } from 'framer-motion'
import { BottomNav } from './BottomNav'

interface AppLayoutProps {
  children: ReactNode
}

export function AppLayout({ children }: AppLayoutProps) {
  const location = useLocation()
  const isAdmin = location.pathname.startsWith('/admin')

  return (
    <div className="flex min-h-screen bg-bg-app">
      {/* Main Content Area */}
      <div className="flex-1 flex flex-col min-h-screen">
        <main
          className={`flex-1 w-full mx-auto px-4 pt-4 pb-28 md:px-6 md:pt-6 relative transition-all duration-200 ${
            isAdmin
              ? 'max-w-5xl'
              : 'max-w-lg'
          }`}
        >
          <AnimatePresence mode="wait">
            <motion.div
              key={location.pathname}
              initial={{ opacity: 0, y: 6 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -6 }}
              transition={{ duration: 0.15, ease: 'easeOut' }}
            >
              {children}
            </motion.div>
          </AnimatePresence>
        </main>
      </div>

      {/* Bottom Navigation */}
      <div className="fixed bottom-0 left-0 right-0 z-20">
        <div
          className={`mx-auto w-full border-t border-border-subtle bg-surface-elevated/95 backdrop-blur-md shadow-[0_-4px_24px_rgba(0,0,0,0.06)] transition-all duration-200 ${
            isAdmin ? 'max-w-5xl md:border-x' : 'max-w-lg md:border-x'
          }`}
          style={{ paddingBottom: 'env(safe-area-inset-bottom)' }}
        >
          <BottomNav />
        </div>
      </div>
    </div>
  )
}
