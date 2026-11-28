const J3_KEYS = [
  'j3-session','j3-role','j3-settings-v1','j3-accessibility','j3-theme','j3-offline-state',
  'j3-write-outbox','j3-read-cache','j3-nav-expanded','j3-nav-width','j3-nav-qa','j3-nav-*',
  'j3-replay-collapsed','j3-pa','j3-replay','j3-sim','j3-offline','j3-*',
];
export async function resetAppToDefaults(reload = true): Promise<void> {
  try { localStorage.clear(); } catch { /* ignore */ }
  try { sessionStorage.clear(); } catch { /* ignore */ }
  // Clear PWA caches + service worker so next load is fresh seeded app.
  try { if ('caches' in window) { for (const key of await caches.keys()) await caches.delete(key); } } catch { /* ignore */ }
  try {
    if ('serviceWorker' in navigator) { for (const reg of await navigator.serviceWorker.getRegistrations()) await reg.unregister(); }
  } catch { /* ignore */ }
  if (reload) window.location.href = '/';
}
export { J3_KEYS };
