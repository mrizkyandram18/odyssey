import { useEffect, useRef, useState } from 'react'

const CURRENT_BUILD_TIME = (import.meta as any).env?.VITE_BUILD_TIME || ''

export const PWAUpdatePrompt: React.FC = () => {
  const [needRefresh, setNeedRefresh] = useState(false)
  const [isUpdating, setIsUpdating] = useState(false)
  const [updateError, setUpdateError] = useState<string | null>(null)
  const [dismissed, setDismissed] = useState(false)
  const controllerCleanupRef = useRef<(() => void) | null>(null)

  useEffect(() => {
    let isMounted = true
    let regRef: ServiceWorkerRegistration | null = null
    let newWorkerRef: ServiceWorker | null = null
    let onStateChangeRef: (() => void) | null = null
    let onUpdateFoundRef: (() => void) | null = null

    // 1. Listen for new Service Worker
    if ('serviceWorker' in navigator) {
      navigator.serviceWorker.ready.then((reg) => {
        if (!isMounted) return
        regRef = reg
        onUpdateFoundRef = () => {
          const newWorker = reg.installing
          if (!newWorker) return
          newWorkerRef = newWorker
          onStateChangeRef = () => {
            if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
              if (isMounted) setNeedRefresh(true)
            }
          }
          newWorker.addEventListener('statechange', onStateChangeRef)
        }
        reg.addEventListener('updatefound', onUpdateFoundRef)
      }).catch(() => { /* ignore */ })
    }

    // 2. Poll /version.json every 30s - if frontend build timestamp changed, notify user!
    let lastBuildTime: string = CURRENT_BUILD_TIME
    const checkVersion = async () => {
      try {
        const r = await fetch(`/version.json?t=${Date.now()}`, { cache: 'no-store' })
        if (r.ok) {
          const data = await r.json().catch(() => null)
          const serverBuild = data?.build_time
          if (serverBuild && lastBuildTime && serverBuild !== lastBuildTime) {
            if (isMounted) setNeedRefresh(true)
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
            if (isMounted) setNeedRefresh(true)
          }
        } catch { /* ignore */ }
      }
    }

    checkVersion()
    const timer = setInterval(checkVersion, 30000)

    // 3. Also check on visibility change (when user re-opens browser or switches tabs)
    const onVisible = () => {
      if (document.visibilityState === 'visible') {
        if ('serviceWorker' in navigator) {
          navigator.serviceWorker.getRegistration().then((r) => r?.update()).catch(() => { /* ignore */ })
        }
        checkVersion()
      }
    }
    document.addEventListener('visibilitychange', onVisible)

    return () => {
      isMounted = false
      clearInterval(timer)
      document.removeEventListener('visibilitychange', onVisible)
      if (regRef && onUpdateFoundRef) {
        regRef.removeEventListener('updatefound', onUpdateFoundRef)
      }
      if (newWorkerRef && onStateChangeRef) {
        newWorkerRef.removeEventListener('statechange', onStateChangeRef)
      }
      if (controllerCleanupRef.current) {
        controllerCleanupRef.current()
        controllerCleanupRef.current = null
      }
    }
  }, [])

  const handleUpdate = async () => {
    if (isUpdating) return
    setIsUpdating(true)
    setUpdateError(null)

    if (controllerCleanupRef.current) {
      controllerCleanupRef.current()
      controllerCleanupRef.current = null
    }

    try {
      if ('serviceWorker' in navigator) {
        const reg = await navigator.serviceWorker.getRegistration()
        if (reg?.waiting) {
          let reloaded = false
          let timeoutId: any = null

          const onControllerChange = () => {
            if (reloaded) return
            reloaded = true
            if (timeoutId) clearTimeout(timeoutId)
            navigator.serviceWorker.removeEventListener('controllerchange', onControllerChange)
            controllerCleanupRef.current = null
            window.location.reload()
          }

          navigator.serviceWorker.addEventListener('controllerchange', onControllerChange)

          // If activation does not complete within 10s, do NOT force reload.
          // Instead, report error and let user retry safely.
          timeoutId = setTimeout(() => {
            if (reloaded) return
            navigator.serviceWorker.removeEventListener('controllerchange', onControllerChange)
            controllerCleanupRef.current = null
            setIsUpdating(false)
            setUpdateError('Pembaruan belum dapat diterapkan. Silakan coba lagi.')
          }, 10000)

          controllerCleanupRef.current = () => {
            if (timeoutId) clearTimeout(timeoutId)
            navigator.serviceWorker.removeEventListener('controllerchange', onControllerChange)
          }

          reg.waiting.postMessage({ type: 'SKIP_WAITING' })
          return
        }
      }
      // If no waiting worker was found, reload to fetch fresh HTML/assets
      window.location.reload()
    } catch {
      setIsUpdating(false)
      setUpdateError('Gagal memeriksa pembaruan. Silakan coba lagi.')
    }
  }

  if (!needRefresh || dismissed) return null

  return (
    <div
      data-testid="pwa-update-prompt"
      className="fixed bottom-4 left-4 right-4 sm:left-auto sm:right-6 sm:max-w-md z-[9999] flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 rounded-2xl bg-zinc-900 text-white p-4 shadow-2xl border border-zinc-700 animate-slideUp"
    >
      <div className="flex items-start gap-2.5">
        <div className="w-2.5 h-2.5 rounded-full bg-accent-magic animate-ping shrink-0 mt-1" />
        <div className="text-xs">
          <p className="font-bold text-white">Pembaruan Sistem Tersedia</p>
          <p className="text-zinc-400 text-[11px] leading-relaxed">
            Versi baru Odyssey tersedia. Simpan pekerjaan Anda lalu perbarui.
          </p>
          {updateError && (
            <p className="text-rose-400 text-[11px] font-medium mt-1" data-testid="pwa-error-message">
              {updateError}
            </p>
          )}
        </div>
      </div>
      <div className="flex items-center gap-2 self-end sm:self-center shrink-0">
        <button
          type="button"
          data-testid="pwa-later-button"
          onClick={() => setDismissed(true)}
          disabled={isUpdating}
          className="px-2.5 py-1.5 rounded-xl text-zinc-400 hover:text-white text-xs font-semibold cursor-pointer transition-colors disabled:opacity-50"
        >
          Nanti
        </button>
        <button
          type="button"
          data-testid="pwa-update-button"
          onClick={handleUpdate}
          disabled={isUpdating}
          className="px-3.5 py-1.5 rounded-xl bg-accent-magic text-white hover:brightness-110 text-xs font-bold cursor-pointer transition-all shadow-md active:scale-95 disabled:opacity-50"
        >
          {isUpdating ? 'Memuat...' : updateError ? 'Coba Lagi' : 'Perbarui'}
        </button>
      </div>
    </div>
  )
}
