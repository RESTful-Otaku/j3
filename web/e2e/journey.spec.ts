import { test, expect, type Page } from '@playwright/test';

const routes = [
  '/', '/setup', '/display', '/mat', '/public', '/brackets', '/tables',
  '/pools', '/control', '/data', '/audit', '/settings', '/details', '/simulate', '/ceremony', '/builder', '/ui', '/view', '/operations', '/overlay',
];

async function expectNoPageErrors(page: Page) {
  const errors: string[] = [];
  page.on('pageerror', (error) => errors.push(error.message));
  return errors;
}

test.describe('J3 end-to-end journeys', () => {
  for (const route of routes) {
    test(`route ${route} renders without runtime errors`, async ({ page }) => {
      const errors = await expectNoPageErrors(page);
      await page.goto(route);
      await expect(page.locator('body')).not.toBeEmpty();
      await expect(page.locator('h1, h2, h3, .brand-name, .display-header, .card, .placeholder').first()).toBeVisible();
      expect(errors, `page errors on ${route}`).toEqual([]);
    });
  }

  test('setup wizard persists a draft and validates before launch', async ({ page }) => {
    const errors = await expectNoPageErrors(page);
    await page.goto('/setup');
    for (let step = 0; step < 4; step++) {
      await page.getByRole('button', { name: 'Continue →' }).click();
    }
    await expect(page.getByText(/need attention before launch/)).toBeVisible();
    await page.locator('.step').first().click();

    await page.locator('input').first().fill('E2E Autumn Cup');
    await page.reload();
    await expect(page.locator('input').first()).toHaveValue('E2E Autumn Cup');
    expect(errors).toEqual([]);
  });

  test('account switcher changes visible permissions', async ({ page }) => {
    const errors = await expectNoPageErrors(page);
    await page.goto('/');
    await page.getByRole('button', { name: 'Accounts & roles' }).click();
    await page.getByRole('button', { name: /Aiko Tanaka \(Runner\)/ }).click();
    await expect(page.locator('.cap.on')).toHaveCount(9);
    await page.getByRole('button', { name: /Sam Viewer/ }).click();
    await expect(page.locator('.cap.on')).toHaveCount(1);
    expect(errors).toEqual([]);
  });

  test('official bracket renders connector lines and rounds', async ({ page }) => {
    const errors = await expectNoPageErrors(page);
    await page.goto('/brackets');
    await page.getByRole('button', { name: /Build official bracket/ }).click();
    await expect(page.locator('.bracket-column').first()).toBeVisible();
    await expect(page.locator('.bracket-lines path').first()).toBeAttached();
    expect(await page.locator('.bracket-column').count()).toBeGreaterThanOrEqual(4);
    expect(errors).toEqual([]);
  });

  test('scorekeeper can start a fight from the mat console', async ({ page }) => {
    const errors = await expectNoPageErrors(page);
    await page.goto('/mat');
    await page.getByRole('button', { name: /Start/ }).first().click();
    await expect(page.getByText(/running|paused/)).toBeVisible();
    expect(errors).toEqual([]);
  });
});
