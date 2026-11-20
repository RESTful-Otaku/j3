import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  plugins: [sveltekit({ adapter: adapter({ fallback: '200.html' }) })],
  server: {
    port: 5173,
    strictPort: true,
    watch: {
      ignored: ['**/android/**', '**/ios/**', '**/build/**', '**/.svelte-kit/**'],
    },
    proxy: {
      '/api': process.env.VITE_API_TARGET ?? 'http://127.0.0.1:8787',
      '/healthz': process.env.VITE_API_TARGET ?? 'http://127.0.0.1:8787',
      '/readyz': process.env.VITE_API_TARGET ?? 'http://127.0.0.1:8787',
    },
  },
  test: {
    include: ['src/**/*.test.ts'],
    environment: 'node',
  },
});
