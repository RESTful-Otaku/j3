import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { defineConfig } from 'vitest/config';

// Docs cards are populated from manifest.json after hydration, so the prerender
// crawler cannot discover /docs/[slug] by following links. Seed every guide
// explicitly; '*' keeps all static routes prerendered as before.
const rootDir = dirname(fileURLToPath(import.meta.url));
const docsManifest = JSON.parse(readFileSync(join(rootDir, 'static/docs/manifest.json'), 'utf8')) as {
  docs: Array<{ slug?: string; path?: string }>;
};
const docEntries = docsManifest.docs
  .map((doc) => doc.path ?? (doc.slug ? `/docs/${doc.slug}` : ''))
  .filter((entry): entry is `/${string}` => entry.startsWith('/'));

export default defineConfig({
  plugins: [
    sveltekit({
      adapter: adapter({ fallback: '200.html' }),
      prerender: { entries: ['*', ...docEntries] },
    }),
  ],
  server: {
    port: 5173,
    strictPort: true,
    watch: {
      ignored: ['**/android/**', '**/ios/**', '**/build/**', '**/.svelte-kit/**'],
    },
    // `^` makes these regex matches. A plain '/api' prefix would also swallow
    // the docs routes (/apidocs, /docs/api) and proxy them to the venue server.
    proxy: {
      '^/api/': process.env.VITE_API_TARGET ?? 'http://127.0.0.1:8787',
      '^/healthz': process.env.VITE_API_TARGET ?? 'http://127.0.0.1:8787',
      '^/readyz': process.env.VITE_API_TARGET ?? 'http://127.0.0.1:8787',
    },
  },
  test: {
    include: ['src/**/*.test.ts'],
    environment: 'node',
    setupFiles: ['src/test-setup.ts'],
  },
});
