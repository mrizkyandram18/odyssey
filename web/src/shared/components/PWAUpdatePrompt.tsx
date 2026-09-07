import { useEffect, useState } from 'react'

const CURRENT_BUILD_TIME = (import.meta as any).env?.VITE_BUILD_TIME || ''

export const PWAUpdatePrompt: React.FC = () => {
  const [needRefresh, setNeedRefresh] = useState(false)

  useEffect(() => {
    // 1. Listen for new Service Worker
    if ('serviceWorker' in navigator) {
      navigator.serviceWorker.ready.then((reg) => {
        reg.addEventListener('updatefound', () => {
          const newWorker = reg.installing
          if (!newWorker) return
          newWorker.addEventListener('statechange', () => {
            if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
              setNeedRefresh(true)
            }
          })
        })
      })
    }

    // 2. Poll /version.json every 30s - if frontend build timestamp changed, auto reload!
    let lastBuildTime: string = CURRENT_BUILD_TIME
    const checkVersion = async () => {
      try {
        const r = await fetch(`/version.json?t=${Date.now()}`, { cache: 'no-store' })
        if (r.ok) {
          const data = await r.json().catch(() => null)
          const serverBuild = data?.build_time
          if (serverBuild && lastBuildTime && serverBuild !== lastBuildTime) {
            setNeedRefresh(true)
            return
          } else if (serverBuild && !lastBuildTime) {
            lastBuildTime = serverBuild
          }
        }
      } catch {
        // Fallback: poll /version API
        try {
          const r = await fetch('/version', { cache: 'no-store' })
          if (!r.ok) return
          const j = await r.json().catch(() => null) as any
          const v = j?.schema_version
          if (v && lastBuildTime && v !== lastBuildTime) {
            setNeedRefresh(true)
          }
        } catch { /* ignore */ }
      }
    }

    checkVersion()
    const timer = setInterval(checkVersion, 30000)

    // 3. Also check immediately on visibility change (when user re-opens browser or switches tabs)
    const onVisible = () => {
      if (document.visibilityState === 'visible') {
        if ('serviceWorker' in navigator) {
          navigator.serviceWorker.getRegistration().then((r) => r?.update())
        }
        checkVersion()
      }
    }
    document.addEventListener('visibilitychange', onVisible)

    return () => {
      clearInterval(timer)
      document.removeEventListener('visibilitychange', onVisible)
    }
  }, [])

  // Auto-reload for HP awam (SMA) - no tap needed, reload after 800ms
  useEffect(() => {
    if (needRefresh) {
      if ('caches' in window) {
        caches.keys().then((names) => {
          names.forEach((name) => caches.delete(name))
        })
      }
      const t = setTimeout(() => window.location.reload(), 800)
      return () => clearTimeout(t)
    }
  }, [needRefresh])

  if (!needRefresh) return null

  return (
    <div className="fixed bottom-4 left-4 right-4 z-[9999] flex items-center justify-between gap-3 rounded-2xl bg-zinc-900 text-white px-4 py-3 shadow-xl border border-zinc-700">
      <span className="text-xs font-bold">Pembaruan sistem — memuat versi terbaru...</span>
      <span className="shrink-0 rounded-xl bg-accent-magic text-white px-3 py-1.5 text-xs font-extrabold animate-pulse">Memuat</span>
    </div>
  )
}
