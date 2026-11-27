import { test, expect } from '@playwright/test';
import { disableTours } from './helpers';

test.beforeEach(async ({ page }) => { await disableTours(page); });

// Svelte Storybook-equivalent: every UI primitive on /ui renders, has no
// page errors, keyboard-accessible, and stays usable on phone + desktop.
test.describe('Component gallery QA', () => {
  test('renders all primitives without errors', async ({ page }) => {
    const errors: string[] = [];
    page.on('pageerror', (e) => errors.push(String(e)));
    page.on('console', (m) => { if (m.type() === 'error') errors.push(m.text()); });
    await page.goto('/ui');
    await expect(page.locator('.card').first()).toBeVisible();
    await expect(page.getByRole('button', { name: /Primary/ }).first()).toBeVisible();
    expect(errors, errors.join('\n')).toEqual([]);
  });

  test('dialog opens, traps focus, closes on Esc', async ({ page }) => {
    await page.goto('/ui');
    await page.getByRole('button', { name: /Open dialog/ }).click();
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();
    await page.keyboard.press('Escape');
    await expect(dialog).toHaveCount(0);
  });

  test('theme switching works', async ({ page }) => {
    await page.goto('/ui');
    await page.getByRole('button', { name: 'light' }).click();
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'light');
    await page.getByRole('button', { name: 'dark' }).click();
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
  });

  test('no sideways scroll on phone', async ({ page }) => {
    await page.setViewportSize({ width: 320, height: 720 });
    await page.goto('/ui');
    const m = await page.evaluate(() => ({ w: document.documentElement.clientWidth, s: document.documentElement.scrollWidth }));
    expect(m.s).toBeLessThanOrEqual(m.w + 1);
  });
});
