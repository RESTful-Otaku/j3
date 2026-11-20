import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

const routes = ['/', '/setup', '/public', '/brackets', '/tables', '/settings', '/details', '/mat'];

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
