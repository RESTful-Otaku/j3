import { test, expect } from '@playwright/test';

test('guided tour can be started from help, navigated and skipped', async ({ page }) => {
  await page.goto('/');
  await page.evaluate(() => {
    localStorage.removeItem('j3-tour-progress');
    localStorage.removeItem('j3-tour-seen');
  });
  await page.reload();
  await expect(page.getByRole('button', { name: 'Start guided tour' }).first()).toBeVisible();
  await page.getByRole('button', { name: 'Start guided tour' }).first().click();
  const dialog = page.getByRole('dialog', { name: 'Guided walkthrough' });
  await expect(dialog).toBeVisible();
  await expect(dialog.locator('.spotlight')).toBeVisible();
  const firstTitle = await dialog.locator('h2').innerText();
  await dialog.getByRole('button', { name: /Next/ }).click();
  await expect(dialog.locator('h2')).not.toHaveText(firstTitle);
  await dialog.getByRole('button', { name: 'Skip tour' }).click();
  await expect(dialog).toHaveCount(0);
});

test('minor registration requires guardian consent', async ({ page }) => {
  await page.goto('/signup');
  await page.locator('.fields input').first().fill('Minor Tester');
  await page.locator('.fields input').nth(1).fill('minor-tester@example.com');
  await page.locator('input[type=date]').fill('2014-05-01');
  await expect(page.getByText('Safeguarding · under 16')).toBeVisible();
  await expect(page.getByLabel('Guardian name')).toBeVisible();
  await expect(page.getByLabel('Guardian email or phone')).toBeVisible();
});
