import { test, expect } from '@playwright/test';
import { disableTours } from './helpers';

test('hands-on guide gates progress until the real action is done', async ({ page }) => {
  await page.goto('/');
  await page.evaluate(() => {
    localStorage.removeItem('j3-tour-progress');
    localStorage.removeItem('j3-tour-seen');
    localStorage.removeItem('j3-guide-progress-v1');
    sessionStorage.removeItem('j3-guide-actions');
  });
  await page.reload();
  await expect(page.getByRole('button', { name: 'Start guided tour' }).first()).toBeVisible();
  await page.getByRole('button', { name: 'Start guided tour' }).first().click();

  const dialog = page.getByRole('dialog', { name: 'Guided walkthrough' });
  await expect(dialog).toBeVisible();
  await expect(dialog.locator('.spotlight')).toBeVisible();
  await expect(dialog).toContainText('Hands-on guide');

  // The first step only asks the user to read the dashboard, so it unlocks at once.
  await dialog.locator('[data-guide="tour-next"]').click();
  await expect(dialog.locator('h2')).toHaveText('Name your tournament');

  // The second step needs a real field value before Next is available.
  const next = dialog.locator('[data-guide="tour-next"]');
  await expect(next).toBeDisabled();
  await page.locator('[data-guide="setup-name"] input').fill('QA Demo Challenge');
  await expect(next).toBeEnabled();
  await next.click();
  await expect(dialog.locator('h2')).not.toHaveText('Name your tournament');

  await dialog.locator('[data-guide="tour-skip"]').click();
  await expect(dialog).toHaveCount(0);
});

test('guide progress page shows a checklist, jump and reset', async ({ page }) => {
  await disableTours(page);
  await page.goto('/guides');
  await expect(page.getByRole('heading', { name: /learn the whole tournament/i })).toBeVisible();
  await expect(page.locator('[data-guide="guides-checklist"] li').first()).toBeVisible();
  expect(await page.locator('[data-guide="guides-checklist"] li').count()).toBeGreaterThan(5);

  await page.locator('[data-guide="guides-checklist"] li').nth(1).getByRole('button').click();
  const dialog = page.getByRole('dialog', { name: 'Guided walkthrough' });
  await expect(dialog.locator('h2')).toHaveText('Name your tournament');
  await dialog.locator('[data-guide="tour-skip"]').click();
  await page.goto('/guides');

  await page.getByRole('button', { name: 'Reset all guides' }).click();
  await expect(page.locator('[data-guide="guides-hero"]')).toContainText('0%');
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
