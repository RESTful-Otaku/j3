import { test, expect } from '@playwright/test';
import { disableTours } from './helpers';

test('navigation stays in the shell while only the content changes', async ({ page }) => {
  await disableTours(page);
  await page.goto('/public');
  await page.evaluate(() => { (window as Window & { navPersisted?: string }).navPersisted = 'kept'; });

  await page.locator('.nav-shell').getByRole('link', { name: 'Brackets' }).click();
  await expect(page).toHaveURL(/\/brackets$/);
  await expect(page.locator('.nav-shell')).toBeVisible();
  expect(await page.evaluate(() => (window as Window & { navPersisted?: string }).navPersisted)).toBe('kept');

  await page.locator('.nav-shell').getByRole('link', { name: 'Public dashboard' }).click();
  await expect(page).toHaveURL(/\/public$/);
  await expect(page.locator('.nav-shell')).toBeVisible();
});

test('mobile keeps a persistent nav and home icon', async ({ page }) => {
  await disableTours(page);
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto('/brackets');

  await expect(page.locator('.mobile-bar .brand-mark')).toBeVisible();
  await expect(page.locator('.mobile-drawer')).toHaveCount(0);

  // Open the hamburger drawer and navigate from it — the bottom tab bar is gone.
  await page.getByRole('button', { name: 'Open navigation' }).click();
  const drawer = page.locator('.mobile-drawer');
  await expect(drawer).toBeVisible();
  await drawer.getByRole('link', { name: /Public/ }).click();
  await expect(page).toHaveURL(/\/public$/);
  await expect(drawer).toHaveCount(0);

  // Brand mark is a home affordance.
  await page.locator('.mobile-bar .brand-mark').click();
  await expect(page).toHaveURL(/\/public$/);
});
