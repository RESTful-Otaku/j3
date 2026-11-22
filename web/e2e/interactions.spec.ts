import { test, expect, type Page } from '@playwright/test';
import { disableTours } from './helpers';

test.beforeEach(async ({ page }) => { await disableTours(page); });

function watchErrors(page: Page) {
  const errors: string[] = [];
  page.on('pageerror', (error) => errors.push(error.message));
  return errors;
}

test.describe('Interactive controls', () => {
  test('mat console: start, pause, resume and confirm', async ({ page }) => {
    const errors = watchErrors(page);
    await page.goto('/mat');
    await page.getByRole('button', { name: /^▶ Start$/ }).click();
    await expect(page.locator('.clock').first()).toBeVisible();
    await page.getByRole('button', { name: /⏸ Pause/ }).click();
    await page.getByRole('button', { name: /▶ Resume/ }).click();
    await page.getByRole('button', { name: /Submit result/ }).first().click();
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();
    await dialog.getByRole('button', { name: /Submit result/ }).click();
    expect(errors).toEqual([]);
  });

  test('mat console: score, foul, hold and undo', async ({ page }) => {
    const errors = watchErrors(page);
    await page.goto('/mat');
    await page.locator('.bracket-match, .card').first().waitFor();
    await page.getByRole('button', { name: 'Red +2' }).click();
    await expect(page.locator('.big-score').first()).toHaveText('2');
    await page.getByRole('button', { name: /Red foul/ }).click();
    await page.getByRole('button', { name: /Red hold/ }).click();
    await expect(page.getByText(/HOLD/)).toBeVisible();
    await page.getByRole('button', { name: /Undo/ }).click();
    expect(errors).toEqual([]);
  });

  test('pools: regenerate from a confirmation dialog', async ({ page }) => {
    const errors = watchErrors(page);
    await page.goto('/pools');
    await page.getByRole('button', { name: /Regenerate pools/ }).click();
    await page.getByRole('dialog').getByRole('button', { name: 'Regenerate' }).click();
    await expect(page.locator('.card').first()).toBeVisible();
    expect(errors).toEqual([]);
  });

  test('data explorer: search and sort', async ({ page }) => {
    const errors = watchErrors(page);
    await page.goto('/data');
    await page.getByPlaceholder('Search…').fill('f01');
    await page.getByRole('button', { name: 'Apply' }).click();
    await expect(page.locator('tbody tr').first()).toBeVisible();
    await page.locator('th button').first().click();
    expect(errors).toEqual([]);
  });

  test('audit: filter by action and expand a change', async ({ page }) => {
    const errors = watchErrors(page);
    await page.goto('/audit');
    await page.getByPlaceholder('Filter by action…').fill('demo.seed');
    await page.getByRole('button', { name: 'Refresh' }).click();
    const entry = page.locator('.entry').first();
    await expect(entry).toBeVisible();
    await entry.getByRole('button', { name: /View change/ }).click();
    await expect(entry.locator('pre')).toBeVisible();
    expect(errors).toEqual([]);
  });

  test('simulator: step, play and pause', async ({ page }) => {
    const errors = watchErrors(page);
    await page.goto('/simulate');
    await page.getByRole('button', { name: /Step/ }).click();
    await expect(page.getByText(/Bout \d+/)).toBeVisible();
    await page.getByRole('button', { name: /Play full tournament/ }).click();
    await page.getByRole('button', { name: /Pause/ }).click();
    expect(errors).toEqual([]);
  });

  test('details: generate a sparring rotation', async ({ page }) => {
    const errors = watchErrors(page);
    await page.goto('/details');
    await page.getByRole('button', { name: /Generate fair rotation/ }).click();
    await expect(page.getByText(/^R1 ·/).first()).toBeVisible();
    expect(errors).toEqual([]);
  });

  test('ceremony: step through the podium', async ({ page }) => {
    const errors = watchErrors(page);
    await page.goto('/ceremony');
    await expect(page.locator('.step.active')).toBeVisible();
    await page.getByRole('button', { name: /Next:/ }).click();
    await expect(page.locator('.step.active')).toBeVisible();
    expect(errors).toEqual([]);
  });

  test('settings: persist profile and appearance', async ({ page }) => {
    const errors = watchErrors(page);
    await page.goto('/settings');
    await page.getByRole('button', { name: 'Imperial' }).click();
    await page.locator('input[type="text"], input:not([type])').first().fill('QA Fighter');
    await page.getByRole('button', { name: 'Save settings' }).click();
    await page.reload();
    await expect(page.locator('input').first()).toHaveValue('QA Fighter');
    expect(errors).toEqual([]);
  });

  test('control: theme, PA toast and template save', async ({ page }) => {
    const errors = watchErrors(page);
    await page.goto('/control');
    await page.getByRole('button', { name: /Street Fighter/ }).click();
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'street');
    await page.getByPlaceholder(/Message, e.g/).fill('QA call to Mat 1');
    await page.getByRole('button', { name: /Announce/ }).click();
    await expect(page.getByText('QA call to Mat 1').first()).toBeVisible();
    await page.getByPlaceholder('Template name…').fill('QA Template');
    await page.getByRole('button', { name: /Save template/ }).click();
    await expect(page.getByText('QA Template')).toBeVisible();
    expect(errors).toEqual([]);
  });

  test('setup: presets, ranges, invites and templates', async ({ page }) => {
    const errors = watchErrors(page);
    await page.goto('/setup');
    await page.locator('.step').nth(1).click();
    await page.getByRole('button', { name: /BJJ points/ }).click();
    await page.getByRole('button', { name: /\+ Add discipline/ }).click();
    await page.locator('.step').nth(2).click();
    await page.getByRole('button', { name: /\+ Add range/ }).first().click();
    await page.locator('.step').nth(3).click();
    await page.locator('textarea').fill('qa@example.com, +447700000000, official, Riverside');
    await page.getByRole('button', { name: /Add invitations/ }).click();
    await expect(page.getByText('qa@example.com')).toBeVisible();
    await page.locator('.step').nth(4).click();
    await page.getByPlaceholder('Template name…').fill('QA Setup');
    await page.getByRole('button', { name: /Save template/ }).click();
    await expect(page.getByText('QA Setup')).toBeVisible();
    expect(errors).toEqual([]);
  });

  test('component gallery exercises theme, dialog and toasts', async ({ page }) => {
    const errors = watchErrors(page);
    await page.goto('/ui');
    await page.getByRole('button', { name: 'street' }).click();
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'street');
    await page.getByRole('button', { name: 'Open dialog' }).click();
    await expect(page.getByRole('dialog')).toBeVisible();
    await page.getByRole('dialog').getByRole('button', { name: 'Confirm' }).click();
    await page.getByRole('button', { name: 'Success' }).click();
    await expect(page.getByText('Saved successfully').first()).toBeVisible();
    expect(errors).toEqual([]);
  });

  test('setup review shows readiness and auto-generates', async ({ page }) => {
    const errors = watchErrors(page);
    await page.goto('/setup');
    await page.locator('.step').nth(4).click();
    await expect(page.getByText('Readiness check')).toBeVisible();
    await page.getByRole('button', { name: /Auto-generate pools/ }).click();
    await expect(page.getByText(/Auto-generated/).first()).toBeVisible();
    expect(errors).toEqual([]);
  });

  test('home shows the tournament journey', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByText('Tournament journey')).toBeVisible();
    await expect(page.locator('.journey a').first()).toBeVisible();
    await expect(page.locator('.journey').getByRole('link', { name: /Live scoreboard/ })).toBeVisible();
  });

  test('operations: check-in and add a fighter', async ({ page }) => {
    const errors = watchErrors(page);
    await page.goto('/operations');
    await expect(page.getByText('Check-in & weigh-in')).toBeVisible();
    await page.locator('input[aria-label^="Weight for"]').first().fill('74');
    await page.locator('input[aria-label^="Weight for"]').first().dispatchEvent('change');
    await expect(page.getByText(/checked in|out of class/).first()).toBeVisible();
    await page.getByRole('textbox', { name: 'Name', exact: true }).fill('E2E Added Fighter');
    await page.getByRole('button', { name: 'Add fighter' }).click();
    await expect(page.getByText(/E2E Added Fighter added/).first()).toBeVisible();
    expect(errors).toEqual([]);
  });

  test('view builder: drag-resize, reorder, save, clone and delete a view', async ({ page }) => {
    const errors = watchErrors(page);
    await page.goto('/builder');
    await page.getByRole('button', { name: 'Add Live scoreboard' }).click();
    await page.getByRole('button', { name: 'Add Clock' }).click();
    await expect(page.locator('.canvas-item')).toHaveCount(2);

    const first = page.locator('.canvas-item').first();
    const widthHandle = first.locator('[data-resize-axis="x"]');
    const widthBox = await widthHandle.boundingBox();
    expect(widthBox).toBeTruthy();
    await page.mouse.move(widthBox!.x + widthBox!.width / 2, widthBox!.y + widthBox!.height / 2);
    await page.mouse.down();
    await page.mouse.move(widthBox!.x + 300, widthBox!.y + widthBox!.height / 2, { steps: 8 });
    await expect(page.locator('.drag-ghost')).toBeVisible();
    await page.mouse.up();
    await expect(first).toHaveAttribute('data-span', '2');

    const heightHandle = first.locator('[data-resize-axis="y"]');
    const heightBox = await heightHandle.boundingBox();
    expect(heightBox).toBeTruthy();
    await page.mouse.move(heightBox!.x + heightBox!.width / 2, heightBox!.y + heightBox!.height / 2);
    await page.mouse.down();
    await page.mouse.move(heightBox!.x + heightBox!.width / 2, heightBox!.y + 180, { steps: 8 });
    await expect(page.locator('.drag-ghost')).toContainText('row');
    await page.mouse.up();
    await expect(first).toHaveAttribute('data-rows', '2');
    await expect(first).toHaveClass(/tall/);

    // Button sizing remains available as an accessible fallback.
    await page.locator('.canvas-item').nth(1).getByRole('button', { name: 'M' }).click();
    await expect(page.locator('.canvas-item').nth(1)).toHaveAttribute('data-span', '2');

    const reorderHandle = first.locator('.drag-handle');
    const target = page.locator('.canvas-item').nth(1);
    const fromBox = await reorderHandle.boundingBox();
    const toBox = await target.boundingBox();
    expect(fromBox).toBeTruthy();
    expect(toBox).toBeTruthy();
    await page.mouse.move(fromBox!.x + fromBox!.width / 2, fromBox!.y + fromBox!.height / 2);
    await page.mouse.down();
    await page.mouse.move(toBox!.x + toBox!.width / 2, toBox!.y + toBox!.height / 2, { steps: 10 });
    await expect(page.locator('.drag-ghost')).toContainText('Drop at position');
    await page.mouse.up();
    await expect(page.locator('.canvas-item').first()).toHaveAttribute('data-widget-type', 'clock');

    await page.getByPlaceholder('View name…').fill(`QA View ${Date.now()}`);
    await page.getByRole('button', { name: /Save view/ }).click();
    await expect(page.locator('.saved').first()).toBeVisible();
    await page.locator('.saved').first().getByRole('button', { name: 'Clone' }).click();
    await expect(page.locator('.saved')).toHaveCount(2);
    await page.locator('.saved').last().getByRole('button', { name: 'Delete' }).click();
    await expect(page.locator('.saved')).toHaveCount(1);
    expect(errors).toEqual([]);
  });

  test('sign up: preview compatible categories and submit', async ({ page }) => {
    const errors = watchErrors(page);
    await page.goto('/');
    await page.getByRole('button', { name: 'Sign up' }).click();
    await page.getByLabel('Full name').fill('QA Signup');
    await page.getByLabel('Email').fill(`qa-signup-${Date.now()}@example.com`);
    await page.locator('input[type=password]').fill('password123');
    await page.getByRole('button', { name: /Check compatible categories/ }).click();
    await expect(page.getByText(/Adult -76/).first()).toBeVisible();
    await page.getByRole('button', { name: 'Create account' }).click();
    await expect(page.getByText(/Account created/)).toBeVisible();
    expect(errors).toEqual([]);
  });
});
