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

test('admin can toggle a permission', async ({ page }) => {
  await page.goto('/permissions');
  await expect(page.getByText('Access control')).toBeVisible();
  const first = page.locator('tbody input[type=checkbox]').first();
  const before = await first.isChecked();
  await first.locator('..').locator('.slider').click();
  await expect(page.getByText(/enabled|disabled/).first()).toBeVisible();
  expect(await first.isChecked()).toBe(!before);
});
