import type { components } from './api-types';

export type DemoAccount = components['schemas']['DemoAccount'];
export type DemoCounts = components['schemas']['DemoCounts'];
export type DemoFight = components['schemas']['DemoFight'];
export type Role = components['schemas']['Role'];
export type Action = components['schemas']['Action'];
export type RoleInfo = components['schemas']['RoleInfo'];

async function getJson<T>(url: string): Promise<T> {
  const response = await fetch(url, { headers: { Accept: 'application/json' } });
  if (!response.ok) {
    throw new Error(`${response.status} ${response.statusText} from ${url}`);
  }
  return (await response.json()) as T;
}

export const api = {
  health: () => getJson<{ status: string }>('/healthz'),
  version: () => getJson<{ name: string; version: string }>('/api/v1/version'),
  roles: () => getJson<{ roles: RoleInfo[] }>('/api/v1/roles'),
  demo: () => getJson<{ counts: DemoCounts }>('/api/v1/demo'),
  accounts: () => getJson<{ accounts: DemoAccount[] }>('/api/v1/demo/accounts'),
  fights: () => getJson<{ fights: DemoFight[] }>('/api/v1/demo/fights'),
  view: (actor?: string) =>
    getJson<{ actor: DemoAccount; actions: Action[]; counts: DemoCounts }>(
      `/api/v1/demo/view${actor ? `?actor=${encodeURIComponent(actor)}` : ''}`,
    ),
};

export const allActions: Action[] = [
  'manage_tenant',
  'manage_event',
  'manage_users',
  'manage_club',
  'invite_fighters',
  'score_fight',
  'register_self',
  'view_private',
  'view_public',
];
