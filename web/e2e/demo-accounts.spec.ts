import { test, expect } from '@playwright/test';
import { disableTours } from './helpers';

// These checks mutate the shared demo database (sessions and invitations).
test.describe.configure({ mode: 'serial' });

type CredentialAccount = {
  id: string;
  name: string;
  email: string;
  role: string;
  roleLabel: string;
  password: string;
};
type Credentials = {
  password: string;
  accounts: CredentialAccount[];
  quickAccounts: CredentialAccount[];
  total: number;
};

test.beforeEach(async ({ page }) => { await disableTours(page); });

async function credentials(page: import('@playwright/test').Page): Promise<Credentials> {
  const response = await page.request.get('/api/v1/demo/credentials');
  expect(response.ok()).toBeTruthy();
  return response.json() as Promise<Credentials>;
}

test('sign-in shows an endpoint-backed demo account sheet', async ({ page }) => {
  const sheet = await credentials(page);
  await page.goto('/login');

  await expect(page.getByRole('heading', { name: 'Demo accounts — one click to test any role' })).toBeVisible();
  await expect(page.getByText(`Every demo account shares the password ${sheet.password}`)).toBeVisible();
  const table = page.locator('.demo-sheet table');
  await expect(table).toBeVisible();
  await expect(table.locator('thead')).toContainText('Role');
  await expect(table.locator('thead')).toContainText('Club');
  await expect(table.locator('tbody tr')).toHaveCount(sheet.quickAccounts.length);
  await expect(page.getByRole('link', { name: /View all/ })).toHaveAttribute('href', '/demo-accounts');
});

test('one-click demo identity switch changes the active role', async ({ page }) => {
  await page.goto('/login');
  const row = page.locator('.demo-sheet tbody tr').filter({ hasText: /Official \/ moderator/ }).first();
  await expect(row).toBeVisible();
  await row.getByRole('button', { name: 'Use account' }).click();

  await expect(page).toHaveURL(/\/official$/);
  await expect.poll(async () => page.evaluate(() => JSON.parse(localStorage.getItem('j3-session') || '{}').role)).toBe('official');
  expect(await page.evaluate(() => localStorage.getItem('j3-role'))).toBe('official');
});

test('demo accounts page lists the full credential sheet', async ({ page }) => {
  const sheet = await credentials(page);
  await page.goto('/demo-accounts');

  await expect(page.getByRole('heading', { name: 'Demo accounts', exact: true })).toBeVisible();
  await expect(page.getByText(`Shared demo password: ${sheet.password}`)).toBeVisible();
  const rows = page.locator('.table-wrap tbody tr');
  await expect(rows).toHaveCount(sheet.total);
  await expect(page.getByLabel('Search demo accounts')).toBeVisible();
  await expect(page.getByLabel('Filter by role')).toBeVisible();
  await page.getByLabel('Search demo accounts').fill('no-such-account');
  await expect(rows).toHaveCount(1);
  await expect(page.getByText('No accounts match those filters.')).toBeVisible();
});

test('an invite can be accepted from its shared link', async ({ page }) => {
  await page.goto('/admin');
  await page.getByLabel('Invite name').fill('E2E Invitee');
  await page.getByLabel('Invite email').fill(`e2e-invitee-${Date.now()}@demo.j3`);
  await page.getByLabel('Invite role').selectOption('fighter');
  await page.getByRole('button', { name: 'Send invite' }).click();
  await expect(page.locator('.invite-result')).toBeVisible();
  const token = (await page.locator('.invite-result code').nth(1).textContent())?.trim() ?? '';
  expect(token).toBeTruthy();

  await page.goto(`/invite/${encodeURIComponent(token)}`);
  await expect(page.getByRole('heading', { name: 'Accept your invitation' })).toBeVisible();
  await page.getByLabel('Your name').fill('E2E Invitee');
  await page.getByLabel('Password', { exact: true }).fill('invited-password');
  await page.getByLabel('Confirm password').fill('invited-password');
  await page.getByRole('button', { name: 'Accept invitation' }).click();
  await expect(page.getByRole('heading', { name: "You're in" })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Go to sign in' })).toBeVisible();
});

test('admin invite form renders every invitable role', async ({ page }) => {
  await page.goto('/admin');
  await expect(page.getByRole('heading', { name: 'Invite an account' })).toBeVisible();
  await expect(page.getByLabel('Invite name')).toBeVisible();
  await expect(page.getByLabel('Invite email')).toBeVisible();
  await expect(page.getByLabel('Club name')).toBeVisible();
  const role = page.getByLabel('Invite role');
  for (const value of ['admin', 'tournament_runner', 'official', 'club_leader', 'fighter', 'spectator']) {
    await expect(role.locator(`option[value="${value}"]`)).toHaveCount(1);
  }
  await expect(page.getByRole('button', { name: 'Send invite' })).toBeVisible();
});
