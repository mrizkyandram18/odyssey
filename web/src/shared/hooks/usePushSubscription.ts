import { useState, useEffect, useCallback } from 'react'
import { isPushSupported, urlBase64ToUint8Array, arrayBufferToBase64 } from '../lib/push'
import { pushApi } from '../lib/api'

export type PushSubscriptionStatus =
  | 'unsupported'
  | 'permission_required'
  | 'enabled'
  | 'blocked'
  | 'error'
  | 'loading'

export interface UsePushSubscriptionReturn {
  status: PushSubscriptionStatus
  error: string | null
  subscribe: () => Promise<void>
  unsubscribe: () => Promise<void>
  checkSubscription: () => Promise<void>
}

export function usePushSubscription(): UsePushSubscriptionReturn {
  const [status, setStatus] = useState<PushSubscriptionStatus>('loading')
  const [error, setError] = useState<string | null>(null)

  const checkSubscription = useCallback(async () => {
    if (!isPushSupported()) {
      setStatus('unsupported')
      return
    }

    if (Notification.permission === 'denied') {
      setStatus('blocked')
      return
    }

    if (Notification.permission === 'default') {
      setStatus('permission_required')
      return
    }

    try {
      const reg = await navigator.serviceWorker.ready
      const sub = await reg.pushManager.getSubscription()
      if (sub) {
        setStatus('enabled')
      } else {
        setStatus('permission_required')
      }
    } catch (e) {
      console.error('Failed to check push subscription', e)
      setStatus('error')
      setError(e instanceof Error ? e.message : 'Subscription check failed')
    }
  }, [])

  useEffect(() => {
    void checkSubscription()
  }, [checkSubscription])

  const subscribe = useCallback(async () => {
    setError(null)
    if (!isPushSupported()) {
      setStatus('unsupported')
      return
    }

    try {
      setStatus('loading')
      const permission = await Notification.requestPermission()
      if (permission === 'denied') {
        setStatus('blocked')
        return
      }
      if (permission !== 'granted') {
        setStatus('permission_required')
        return
      }

      const vapidKey = import.meta.env.VITE_VAPID_PUBLIC_KEY || ''
      if (!vapidKey) {
        throw new Error('VAPID public key not configured (VITE_VAPID_PUBLIC_KEY missing)')
      }

      const reg = await navigator.serviceWorker.ready
      let sub = await reg.pushManager.getSubscription()
      if (!sub) {
        sub = await reg.pushManager.subscribe({
          userVisibleOnly: true,
          applicationServerKey: urlBase64ToUint8Array(vapidKey).buffer as ArrayBuffer,
        })
      }

      const p256dh = arrayBufferToBase64(sub.getKey('p256dh'))
      const authSecret = arrayBufferToBase64(sub.getKey('auth'))

      if (!p256dh || !authSecret) {
        throw new Error('Failed to extract subscription encryption keys')
      }

      await pushApi.subscribe({
        endpoint: sub.endpoint,
        keys: {
          p256dh,
          auth: authSecret,
        },
      })

      setStatus('enabled')
    } catch (e) {
      console.error('Push subscription failed', e)
      setStatus('error')
      setError(e instanceof Error ? e.message : 'Push subscription failed')
    }
  }, [])

  const unsubscribe = useCallback(async () => {
    setError(null)
    if (!isPushSupported()) return

    try {
      setStatus('loading')
      const reg = await navigator.serviceWorker.ready
      const sub = await reg.pushManager.getSubscription()
      const endpoint = sub?.endpoint

      if (sub) {
        await sub.unsubscribe()
      }

      await pushApi.unsubscribe(endpoint)
      setStatus('permission_required')
    } catch (e) {
      console.error('Push unsubscription failed', e)
      setStatus('error')
      setError(e instanceof Error ? e.message : 'Unsubscription failed')
    }
  }, [])

  return {
    status,
    error,
    subscribe,
    unsubscribe,
    checkSubscription,
  }
}
