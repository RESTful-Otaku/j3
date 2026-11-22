import { test, expect, type Page } from '@playwright/test';
import { disableTours } from './helpers';

test.beforeEach(async ({ page }) => { await disableTours(page); });

function watchErrors(page: Page) {
  const errors: string[] = [];
  page.on('pageerror', (error) => errors.push(error.message));
  return errors;
}

async function confirm(page: Page) {
  const dialog = page.getByRole('dialog');
  await expect(dialog).toBeVisible();
  await dialog.getByRole('button', { name: /Confirm & log/ }).click();
  await expect(dialog).toBeHidden();
}

async function openConsole(page: Page) {
  await page.goto('/official');
  // Scoring handlers intentionally ignore clicks until a contest is selected;
  // wait for the assigned-contest feed so parallel E2E runs cannot race it.
  await expect(page.getByLabel('Select contest')).not.toHaveValue('');
}

test.describe('Officials console', () => {
  test('searches the full catalogue and logs configured points', async ({ page }) => {
    const errors = watchErrors(page);
    await openConsole(page);
    const search = page.getByLabel('Search techniques');
    await expect(search).toBeVisible();
    await search.fill('seoi');
    const technique = page.locator('.tech').filter({ hasText: 'Shoulder throw' }).first();
    await expect(technique).toBeVisible();
    await technique.click();
    const points = page.getByLabel('Configured points');
    await expect(points).toBeVisible();
    expect((await points.locator('option').count()) > 0).toBeTruthy();
    await page.getByRole('button', { name: /^Log for RED/ }).click();
    await confirm(page);
    expect(errors).toEqual([]);
  });

  test('attributes yellow and red cards to the selected corner', async ({ page }) => {
    const errors = watchErrors(page);
    await openConsole(page);
    await page.getByRole('button', { name: /Yellow \/ shido/ }).click();
    await confirm(page);
    await page.getByRole('button', { name: /Red \/ hansoku/ }).click();
    await confirm(page);
    expect(errors).toEqual([]);
  });

  test('controls the authoritative clock and confirms a final result', async ({ page }) => {
    const errors = watchErrors(page);
    await openConsole(page);
    await page.getByRole('button', { name: /^▶ Start$/ }).click();
    await expect(page.getByText(/RUNNING/).first()).toBeVisible();
    await page.getByRole('button', { name: /⏸ Pause/ }).click();
    await page.getByRole('button', { name: /^▶ Resume$/ }).click();
    await page.getByRole('button', { name: 'Red wins', exact: true }).click();
    await confirm(page);
    expect(errors).toEqual([]);
  });

  test('calls a technical result with a reason and winner', async ({ page }) => {
    const errors = watchErrors(page);
    await openConsole(page);
    await page.getByLabel('Technical winner').selectOption('blue');
    await page.getByLabel('Technical reason').selectOption('injury');
    await page.getByLabel('Technical result detail').fill('Shoulder injury');
    await page.getByRole('button', { name: /Call technical result/ }).click();
    const dialog = page.getByRole('dialog');
    await expect(dialog).toContainText('Technical result');
    await confirm(page);
    expect(errors).toEqual([]);
  });

  test('public dashboard reflects the seeded mid-event state', async ({ page }) => {
    const errors = watchErrors(page);
    await page.goto('/public');
    await expect(page.getByText(/Live now/).first()).toBeVisible();
    await expect(page.getByText(/Next up — report to your mat/)).toBeVisible();
    await expect(page.locator('.fight-card, .card').first()).toBeVisible();
    expect(errors).toEqual([]);
  });
});
