import { useEffect, useState } from 'react'

export const PWAUpdatePrompt: React.FC = () => {
  const [needRefresh, setNeedRefresh] = useState(false)

  useEffect(() => {
    if (!('serviceWorker' in navigator)) return
    let timeout: any

    // Listen for new SW
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

    // Poll /version every 60s - if version changes, show prompt (works even without SW)
    let lastVersion: string | null = null
    const check = async () => {
      try {
        const r = await fetch('/version', { cache: 'no-store' })
        if (!r.ok) return
        const j = await r.json().catch(() => null) as any
        const v = j?.version ?? j?.schema_version ?? await r.text()
        if (lastVersion && v && v !== lastVersion) setNeedRefresh(true)
        else if (!lastVersion) lastVersion = v
      } catch {}
    }
    check()
    timeout = setInterval(check, 60000)

    // Also check on visibility change (when user re-opens app)
    const onVisible = () => {
      if (document.visibilityState === 'visible') {
        navigator.serviceWorker.getRegistration().then((r) => r?.update())
        check()
      }
    }
    document.addEventListener('visibilitychange', onVisible)

    return () => {
      clearInterval(timeout)
      document.removeEventListener('visibilitychange', onVisible)
    }
  }, [])

  // Auto-reload for HP awam (SMA) - no tap needed, reload after 1.5s
  useEffect(() => {
    if (needRefresh) {
      const t = setTimeout(() => window.location.reload(), 1500)
      return () => clearTimeout(t)
    }
  }, [needRefresh])

  if (!needRefresh) return null

  return (
    <div className="fixed bottom-4 left-4 right-4 z-[9999] flex items-center justify-between gap-3 rounded-2xl bg-zinc-900 text-white px-4 py-3 shadow-xl border border-zinc-700">
      <span className="text-xs font-bold">Update tersedia — memuat ulang...</span>
      <span className="shrink-0 rounded-xl bg-white text-zinc-900 px-3 py-1.5 text-xs font-extrabold animate-pulse">Reload</span>
    </div>
  )
}
