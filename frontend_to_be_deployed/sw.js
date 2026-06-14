// Network-only service worker.
// Present so the app is installable (PWA/TWA), but it intentionally does NOT
// cache anything — the app requires a live connection to the Go server.
self.addEventListener('install', () => self.skipWaiting());
self.addEventListener('activate', (e) => e.waitUntil(self.clients.claim()));
self.addEventListener('fetch', (event) => {
  // Always go to the network. If offline, the request fails (no offline mode).
  event.respondWith(fetch(event.request));
});
