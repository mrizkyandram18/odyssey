import type { ReactNode } from 'react'
import { useLocation } from 'react-router-dom'
import { motion, AnimatePresence } from 'framer-motion'
import { MobileNav } from './MobileNav'

interface AppLayoutProps {
  children: ReactNode
}

export function AppLayout({ children }: AppLayoutProps) {
  const location = useLocation()
  return (
    <div className="flex h-screen overflow-hidden bg-bg-app relative">
      {/* Ambient Animated Background */}
      <div className="absolute inset-0 z-0 overflow-hidden pointer-events-none">
        <motion.div
          className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] rounded-full bg-accent-magic/10 blur-[100px]"
          animate={{
            x: [0, 50, 0],
            y: [0, 30, 0],
          }}
          transition={{
            duration: 15,
            repeat: Infinity,
            ease: "easeInOut"
          }}
        />
        <motion.div
          className="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] rounded-full bg-accent-nature/10 blur-[120px]"
          animate={{
            x: [0, -40, 0],
            y: [0, -50, 0],
          }}
          transition={{
            duration: 20,
            repeat: Infinity,
            ease: "easeInOut"
          }}
        />
      </div>

      {/* Mobile-First Main Content Area */}
      <div className="flex-1 flex flex-col h-full relative overflow-y-auto overflow-x-hidden">
        <main className="flex-1 w-full max-w-md mx-auto p-4 md:p-6 border-x border-border-subtle/30 bg-surface-elevated/30 shadow-2xl relative min-h-full">
          <AnimatePresence>
            <motion.div
              key={location.pathname}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              transition={{ duration: 0.15, ease: 'easeOut' }}
            >
              {children}
            </motion.div>
          </AnimatePresence>
          {/* Spacer to prevent content from hiding behind the fixed bottom nav */}
          <div className="h-32 w-full shrink-0 pointer-events-none" />
        </main>
      </div>

      {/* Universal Mobile Navigation (Always at bottom) */}
      <div className="fixed bottom-0 left-0 right-0 z-20 pointer-events-none">
        <div className="max-w-md mx-auto w-full pointer-events-auto border-t border-border-subtle bg-surface-elevated/95 backdrop-blur-md pb-safe border-x">
          <MobileNav />
        </div>
      </div>
    </div>
  )
}
