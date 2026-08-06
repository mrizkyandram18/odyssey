// Minimal service worker for PWA shell caching.
// Full offline-first support is deferred to a future phase (see docs/future.md).
const CACHE_NAME = 'odyssey-shell-v1'

const urlsToCache = [
  '/',
  '/index.html',
  '/manifest.webmanifest',
]

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => cache.addAll(urlsToCache))
  )
})

self.addEventListener('fetch', (event) => {
  event.respondWith(
    caches.match(event.request).then((response) => {
      return response || fetch(event.request)
    })
  )
})
