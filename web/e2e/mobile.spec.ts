import { test, expect } from '@playwright/test';

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

  test('key views stay usable without endless scrolling', async ({ page }) => {
    for (const route of ['/', '/setup', '/display', '/mat', '/public', '/brackets', '/tables', '/pools', '/control', '/data', '/audit', '/settings', '/details', '/simulate', '/ceremony', '/builder', '/ui', '/view', '/operations', '/overlay', '/event', '/official', '/login', '/permissions']) {
      await page.goto(route);
      await expect(page.locator('body')).not.toBeEmpty();
      const scrollHeight = await page.evaluate(() => document.documentElement.scrollHeight);
      expect(scrollHeight, `${route} should fit reasonably on a phone`).toBeLessThan(8000);
    }
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
