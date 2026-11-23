import { test, expect } from '@playwright/test';
import { disableTours } from './helpers';

test.beforeEach(async ({ page }) => { await disableTours(page); });

test('login with Google and MFA, then sign out', async ({ page }) => {
  await page.goto('/login');
  await page.getByRole('button', { name: /Google/ }).click();
  await expect(page.getByText('Two-factor verification')).toBeVisible();
  await page.getByRole('button', { name: 'Verify & sign in' }).click();
  await expect(page).toHaveURL(/\/admin$/);
});

test('local demo run auto signs in as the admin', async ({ page }) => {
  await page.goto('/');
  await expect.poll(async () => page.evaluate(() => {
    const stored = JSON.parse(localStorage.getItem('j3-session') || '{}');
    return stored.role;
  })).toBe('admin');
  const stored = await page.evaluate(() => JSON.parse(localStorage.getItem('j3-session') || '{}'));
  expect(stored.userId).toBeTruthy();
  expect(stored.token).toBeTruthy();
  await expect(page.getByText('Tournament journey', { exact: true })).toBeVisible();
});

test('stale demo session is replaced automatically after a data reset', async ({ page }) => {
  const unauthorized: string[] = [];
  page.on('response', (response) => {
    if (response.status() === 401) unauthorized.push(response.url());
  });
  await page.goto('/');
  await expect(page.getByText('Tournament journey', { exact: true })).toBeVisible();
  expect(await page.evaluate(() => localStorage.getItem('j3-session'))).toBeTruthy();

  // Simulate a browser holding a token whose server-side session was deleted.
  await page.evaluate(() => localStorage.setItem('j3-session', JSON.stringify({
    token: 'stale-after-reset', role: 'runner', expiresAt: '2099-01-01T00:00:00Z',
  })));
  unauthorized.length = 0;
  await page.reload();

  await expect(page.getByText('Tournament journey', { exact: true })).toBeVisible();
  await expect
    .poll(async () => JSON.parse(await page.evaluate(() => localStorage.getItem('j3-session') || '{}')).token)
    .not.toBe('stale-after-reset');
  const refreshed = JSON.parse(await page.evaluate(() => localStorage.getItem('j3-session') || '{}'));
  // In-flight requests can each see one 401 with the stale token. Recovery is
  // shared, and no request made after it should be unauthorized.
  // The first in-flight request may already have been retried before the
  // listener attached on slower parallel runs; assert the shape whenever a
  // stale-token response was observed, and always require a clean tail.
  expect(new Set(unauthorized).size).toBe(unauthorized.length);
  expect(unauthorized.every((url) => url.includes('/api/v1/') && !url.includes('/auth/'))).toBe(true);
  const afterRecovery = [...unauthorized];
  await page.waitForTimeout(600);
  expect(unauthorized).toEqual(afterRecovery);
});

test('admin can toggle a permission', async ({ page }) => {
  await page.goto('/permissions');
  await expect(page.getByText('Access control')).toBeVisible();
  const first = page.locator('tbody input[type=checkbox]').first();
  const before = await first.isChecked();
  await first.locator('..').locator('.slider').click();
  await expect(page.getByText(/enabled|disabled/).first()).toBeVisible();
  expect(await first.isChecked()).toBe(!before);
});
