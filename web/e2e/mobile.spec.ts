import { test, expect } from '@playwright/test';
import { disableTours } from './helpers';

test.beforeEach(async ({ page }) => { await disableTours(page); });

test.use({ viewport: { width: 390, height: 844 }, isMobile: true, hasTouch: true });

test.describe('Mobile navigation & layouts', () => {
  test('drawer opens, navigates and closes', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('.mobile-bar')).toBeVisible();
    await page.getByRole('button', { name: 'Open navigation' }).click();
    const drawer = page.locator('.mobile-drawer');
    await expect(drawer).toBeVisible();
    await drawer.getByRole('link', { name: /Brackets/ }).click();
    await expect(page).toHaveURL(/\/brackets/);
    await expect(drawer).toHaveCount(0);
  });

  test('phone views fit the viewport without sideways scrolling', async ({ page }) => {
    for (const route of ['/', '/login', '/signup', '/public', '/brackets', '/tables', '/pools', '/ceremony', '/event', '/details', '/notifications', '/settings', '/guides', '/docs', '/view']) {
      await page.goto(route);
      await expect(page.locator('body')).not.toBeEmpty();
      const metrics = await page.evaluate(() => ({
        width: document.documentElement.clientWidth,
        scrollWidth: document.documentElement.scrollWidth,
        scrollHeight: document.documentElement.scrollHeight,
      }));
      expect(metrics.scrollWidth, `${route} should not scroll sideways`).toBeLessThanOrEqual(metrics.width + 1);
      expect(metrics.scrollHeight, `${route} should fit reasonably on a phone`).toBeLessThan(6000);
    }
  });

  test('operator screens stay on tablet and desktop', async ({ page }) => {
    for (const route of ['/setup', '/display', '/mat', '/control', '/data', '/audit', '/builder', '/ui', '/operations', '/overlay', '/official', '/permissions', '/simulate']) {
      await page.goto(route);
      await expect(page).toHaveURL(/\/public$/);
    }
    await page.getByRole('button', { name: 'Open navigation' }).click();
    const drawer = page.locator('.mobile-drawer');
    await expect(drawer.getByRole('link', { name: /Official console/ })).toHaveCount(0);
    await expect(drawer.getByRole('link', { name: /Create tournament/ })).toHaveCount(0);
    await expect(drawer.getByText(/Scoring, setup and audit tools are available on tablet or desktop/)).toBeVisible();
  });

  test('bundled seeded data works with the venue server offline', async ({ page }) => {
    await page.route('**/api/**', (route) => route.abort());
    await page.route('**/healthz', (route) => route.abort());
    await page.goto('/public?offline=1');
    await expect(page.locator('body')).toContainText('Okami Jiu Jitsu Challengers Tournament 2027');
    await expect(page.locator('body')).toContainText('On the mats now');
  });
});

test.describe('Tablet layouts', () => {
  test('portrait and landscape render without overflow', async ({ browser }) => {
    for (const viewport of [{ width: 768, height: 1024 }, { width: 1024, height: 768 }]) {
      const context = await browser.newContext({ viewport });
      const page = await context.newPage();
      for (const route of ['/', '/brackets', '/official', '/tables', '/event']) {
        await page.goto(route);
        await expect(page.locator('body')).not.toBeEmpty();
        const overflow = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
        expect(overflow, `${route} at ${viewport.width}x${viewport.height}`).toBeLessThan(40);
      }
      await context.close();
    }
  });
});
