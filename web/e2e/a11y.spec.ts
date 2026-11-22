import { test, expect } from '@playwright/test';
import { disableTours } from './helpers';

test.beforeEach(async ({ page }) => { await disableTours(page); });
import AxeBuilder from '@axe-core/playwright';

const routes = [
  '/', '/welcome', '/login', '/signup', '/setup', '/public', '/display', '/mat',
  '/official', '/brackets', '/tables', '/pools', '/control', '/data', '/audit',
  '/settings', '/guides', '/details', '/simulate', '/ceremony', '/builder', '/ui', '/view',
  '/operations', '/overlay', '/event', '/notifications', '/admin', '/club', '/permissions',
  '/docs', '/docs/guide', '/docs/faq', '/docs/development', '/docs/ci', '/docs/api', '/docs/api-guide',
];

for (const route of routes) {
  test(`accessibility scan: ${route}`, async ({ page }) => {
    await page.goto(route);
    const results = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa'])
      .analyze();
    const serious = results.violations.filter(
      (violation) => violation.impact === 'serious' || violation.impact === 'critical',
    );
    expect(
      serious.map((violation) => `${violation.id}: ${violation.nodes.map((n) => n.target).join(', ')}`),
      `serious a11y violations on ${route}`,
    ).toEqual([]);
  });
}
