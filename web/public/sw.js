// Service worker removed — offline-first deferred to a future phase.
// This script unregisters itself to clear stale caches from prior deploys.
self.addEventListener('install', () => self.skipWaiting())
self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then(keys =>
      Promise.all(keys.map(k => caches.delete(k)))
    ).then(() => self.registration.unregister())
  )
})
