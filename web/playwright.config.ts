import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  timeout: 45_000,
  expect: { timeout: 10_000 },
  fullyParallel: true,
  reporter: [['list']],
  use: {
    baseURL: 'http://127.0.0.1:5180',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  webServer: [
    {
      command: 'cd .. && go run ./cmd/j3d -addr 127.0.0.1:8791 -db /tmp/j3-e2e.db -seed',
      url: 'http://127.0.0.1:8791/healthz',
      reuseExistingServer: true,
      timeout: 120_000,
    },
    {
      command: 'bun run dev -- --port 5180',
      url: 'http://127.0.0.1:5180/',
      env: { VITE_API_TARGET: 'http://127.0.0.1:8791' },
      reuseExistingServer: true,
      timeout: 120_000,
    },
  ],
});
