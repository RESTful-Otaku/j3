import { test, expect } from '@playwright/test';
import { disableTours } from './helpers';

test.beforeEach(async ({ page }) => {
  await disableTours(page);
});

test('documentation home loads the guide index and server status', async ({ page }) => {
  await page.goto('/docs');
  await expect(page.getByRole('heading', { name: /everything you need to run j3/i })).toBeVisible();
  await expect(page.locator('.doc-card')).toHaveCount(13);
  await expect(page.locator('.status')).toContainText('Venue server online');
  await expect(page.getByRole('link', { name: /organiser guide/i }).first()).toBeVisible();
});

test('search finds guides and endpoints across the docs site', async ({ page }) => {
  await page.goto('/docs');
  await page.getByLabel('Search documentation').fill('scoring');
  await expect(page.getByRole('navigation', { name: 'Search results' }).locator('a')).not.toHaveCount(0);
  await expect(page.getByRole('navigation', { name: 'Search results' })).toContainText('Live scoring rules');
});

test('sidebar lists every guide and jumps to in-page sections', async ({ page }) => {
  await page.goto('/docs/guide');
  await expect(page.locator('aside.sidebar')).toContainText('Runner guide');
  await expect(page.locator('aside.sidebar')).toContainText('Live scoring rules');
  const sections = page.getByRole('navigation', { name: 'Sections in this guide' });
  await expect(sections.locator('a').first()).toBeVisible();
  expect(await sections.locator('a').count()).toBeGreaterThan(5);
  const sectionLink = sections.locator('a').nth(2);
  const sectionHref = await sectionLink.getAttribute('href');
  const sectionId = (sectionHref ?? '').split('#')[1];
  expect(sectionId).toBeTruthy();
  await sectionLink.click();
  expect(page.url()).toContain(`#${sectionId}`);
  await expect(page.locator(`#${sectionId}`)).toBeInViewport();
});

test('every in-page link points at a heading that actually exists', async ({ page }) => {
  await page.goto('/docs/guide');
  await expect(page.getByRole('navigation', { name: 'Sections in this guide' })).toBeVisible();
  const anchors = await page.locator('aside.sidebar nav[aria-label="Sections in this guide"] a, .toc a[href^="#"]').evaluateAll(
    (links) => links.map((link) => (link as HTMLAnchorElement).getAttribute('href') ?? ''),
  );
  expect(anchors.length).toBeGreaterThan(5);
  for (const href of anchors) {
    const id = href.split('#')[1];
    if (!id) continue;
    await expect(page.locator(`#${id}`), `${href} has no matching heading`).toHaveCount(1);
  }
});

test('FAQ links to the rendered API guide', async ({ page }) => {
  await page.goto('/docs/faq');
  const link = page.getByRole('link', { name: 'API guide', exact: true });
  await expect(link).toHaveAttribute('href', '/docs/api-guide');
  await link.click();
  await expect(page.getByRole('heading', { level: 1, name: /api guide/i })).toBeVisible();
});

test('sidebar links swap the rendered guide', async ({ page }) => {
  await page.goto('/docs/faq');
  await page.locator('aside.sidebar a').filter({ hasText: 'Organiser guide' }).first().click();
  await expect(page.getByRole('heading', { level: 1, name: /organiser guide/i })).toBeVisible();
  await page.locator('aside.sidebar a').filter({ hasText: 'Live scoring rules' }).first().click();
  await expect(page.getByRole('heading', { level: 1, name: /live scoring rules/i })).toBeVisible();
});

test('markdown guides render sections and a table of contents', async ({ page }) => {
  await page.goto('/docs/guide');
  await expect(page.locator('.prose h2').first()).toBeVisible();
  expect(await page.locator('.prose h2').count()).toBeGreaterThan(5);
  await expect(page.getByRole('navigation', { name: 'On this page' }).locator('a[href^="#"]').first()).toBeVisible();
  expect(await page.locator('.prose table').count()).toBeGreaterThan(0);
});

test('API explorer lists every contract operation and can call a public endpoint', async ({ page }) => {
  await page.goto('/docs/api');
  await expect(page.getByText(/79 operations/)).toBeVisible();
  await page.locator('.list button').filter({ hasText: 'Seeded data counts' }).first().click();
  await page.getByRole('button', { name: /Try it/ }).click();
  await expect(page.locator('.status')).toHaveText('200');
  await expect(page.locator('.response-panel pre')).toContainText('"clubs"');
});

test('legacy /apidocs URL redirects to the documentation API explorer', async ({ page }) => {
  await page.goto('/apidocs');
  await page.waitForURL('**/docs/api');
  await expect(page.getByRole('heading', { name: 'API explorer' })).toBeVisible();
});
