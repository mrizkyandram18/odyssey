// Odyssey Service Worker for PWA Web Push Notifications

self.addEventListener('install', () => {
  self.skipWaiting()
})

self.addEventListener('activate', (event) => {
  event.waitUntil(self.clients.claim())
})

self.addEventListener('push', (event) => {
  let data = {
    title: 'Odyssey',
    body: "It's your turn!",
    url: '/#/home',
    tag: 'odyssey-turn',
  }

  if (event.data) {
    try {
      const parsed = event.data.json()
      data = { ...data, ...parsed }
    } catch {
      data.body = event.data.text() || data.body
    }
  }

  const options = {
    body: data.body,
    icon: '/favicon-32x32.png',
    badge: '/favicon-32x32.png',
    tag: data.tag || 'odyssey-notification',
    data: {
      url: data.url || '/#/home',
    },
  }

  event.waitUntil(self.registration.showNotification(data.title, options))
})

self.addEventListener('notificationclick', (event) => {
  event.notification.close()

  const targetUrl = new URL(event.notification.data?.url || '/#/home', self.location.origin).href

  event.waitUntil(
    self.clients.matchAll({ type: 'window', includeUncontrolled: true }).then((windowClients) => {
      for (const client of windowClients) {
        if (client.url === targetUrl && 'focus' in client) {
          return client.focus()
        }
      }
      if (self.clients.openWindow) {
        return self.clients.openWindow(targetUrl)
      }
    })
  )
})
