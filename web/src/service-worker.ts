/// <reference types="@sveltejs/kit" />
/// <reference no-default-lib="true"/>
/// <reference lib="esnext" />
/// <reference lib="webworker" />
import { immutable, assets, prerendered } from '$app/manifest';
import { version } from '$app/env';
import { resolve } from '$app/paths';

declare const self: ServiceWorkerGlobalScope;

const CACHE = `j3-${version}`;
const ASSETS = [...immutable, ...assets, ...prerendered, resolve('/'), resolve('/manifest.webmanifest')];

self.addEventListener('install', (event) => {
  event.waitUntil(caches.open(CACHE).then((cache) => cache.addAll(ASSETS)).then(() => self.skipWaiting()));
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((keys) => Promise.all(keys.filter((k) => k !== CACHE).map((k) => caches.delete(k)))).then(() => self.clients.claim())
  );
});

self.addEventListener('fetch', (event) => {
  const { request } = event;
  if (request.method !== 'GET') return; // never cache mutations/secrets
  const url = new URL(request.url);
  if (url.origin !== self.location.origin) return;

  // API reads: network-first, fallback to cache. Never cache auth endpoints.
  if (url.pathname.startsWith('/api/') || url.pathname === '/healthz') {
    if (url.pathname.startsWith('/api/v1/auth')) return;
    event.respondWith((async () => {
      try {
        const fresh = await fetch(request);
        if (fresh.ok) {
          const cache = await caches.open(CACHE);
          cache.put(request, fresh.clone()).catch(() => {});
        }
        return fresh;
      } catch {
        const cached = await caches.match(request);
        if (cached) return cached;
        return new Response(JSON.stringify({ error: 'offline' }), { status: 503, headers: { 'Content-Type': 'application/json' } });
      }
    })());
    return;
  }

  // App shell + assets: cache-first, fallback to network, then shell.
  event.respondWith((async () => {
    const cached = await caches.match(request);
    if (cached) return cached;
    try { return await fetch(request); } catch { return (await caches.match('/')) ?? new Response('offline', { status: 503 }); }
  })());
});
