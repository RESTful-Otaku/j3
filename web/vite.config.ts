/// <reference types="vitest/config" />
import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { defineConfig } from 'vitest/config';

// Docs cards are populated from manifest.json after hydration, so the prerender
// crawler cannot discover /docs/[slug] by following links. Seed every guide
// explicitly; '*' keeps all static routes prerendered as before.
import path from 'node:path';
import { storybookTest } from '@storybook/addon-vitest/vitest-plugin';
import { playwright } from '@vitest/browser-playwright';
// More info at: https://storybook.js.org/docs/next/writing-tests/integrations/vitest-addon
const rootDir = process.cwd();
const docsManifest = JSON.parse(readFileSync(rootDir + '/static/docs/manifest.json', 'utf8')) as {
  docs: Array<{
    slug?: string;
    path?: string;
  }>;
};
const docEntries = docsManifest.docs.map(doc => doc.path ?? (doc.slug ? `/docs/${doc.slug}` : '')).filter((entry): entry is `/${string}` => entry.startsWith('/'));
export default defineConfig({
  plugins: [sveltekit({
    adapter: adapter({
      fallback: '200.html'
    }),
    prerender: {
      entries: ['*', ...docEntries]
    }
  })],
  server: {
    port: 5173,
    strictPort: true,
    watch: {
      ignored: ['**/android/**', '**/ios/**', '**/build/**', '**/.svelte-kit/**']
    },
    // `^` makes these regex matches. A plain '/api' prefix would also swallow
    // the docs routes (/apidocs, /docs/api) and proxy them to the venue server.
    proxy: {
      '^/api/': process.env.VITE_API_TARGET ?? 'http://127.0.0.1:8787',
      '^/healthz': process.env.VITE_API_TARGET ?? 'http://127.0.0.1:8787',
      '^/readyz': process.env.VITE_API_TARGET ?? 'http://127.0.0.1:8787'
    }
  },
  test: {
    projects: [{
      extends: true,
      test: {
        name: 'unit',
        include: ['src/**/*.test.ts'],
        environment: 'node',
        setupFiles: ['src/test-setup.ts']
      }
    }, {
      extends: true,
      plugins: [
      // The plugin will run tests for the stories defined in your Storybook config
      // See options at: https://storybook.js.org/docs/next/writing-tests/integrations/vitest-addon#storybooktest
      storybookTest({
        configDir: path.join(rootDir, '.storybook')
      })],
      test: {
        name: 'storybook',
        browser: {
          enabled: true,
          headless: true,
          provider: playwright({}),
          instances: [{
            browser: 'chromium'
          }]
        }
      }
    }]
  }
});