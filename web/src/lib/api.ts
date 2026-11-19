import type { components } from './api-types';

export type SignupRequest = components['schemas']['SignupRequest'];
export type SignupPreview = components['schemas']['SignupPreview'];
export type SignupResult = components['schemas']['SignupResult'];
export type Eligibility = components['schemas']['Eligibility'];
export type PoolsResponse = components['schemas']['PoolsResponse'];
export type PoolView = components['schemas']['PoolView'];
export type GeneratePoolsRequest = components['schemas']['GeneratePoolsRequest'];
export type RoleRequest = components['schemas']['RoleRequest'];
export type DisplayOfficial = components['schemas']['DisplayOfficial'];
export type EventDisplayInfo = components['schemas']['EventDisplayInfo'];
export type StandingRow = components['schemas']['StandingRow'];
export type DemoAccount = components['schemas']['DemoAccount'];
export type DemoCounts = components['schemas']['DemoCounts'];
export type DemoFight = components['schemas']['DemoFight'];
export type DemoFightEvent = components['schemas']['DemoFightEvent'];
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

async function postJson<T>(url: string, body: unknown): Promise<T> {
  const response = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
    body: JSON.stringify(body),
  });
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
  signupPreview: (body: SignupRequest) =>
    postJson<SignupPreview>('/api/v1/signup/preview', body),
  signup: (body: SignupRequest) => postJson<SignupResult>('/api/v1/signup', body),
  roleRequests: () =>
    getJson<{ requests: RoleRequest[] }>('/api/v1/demo/role-requests'),
  approveRoleRequest: (id: string) =>
    postJson<{ status: string }>(`/api/v1/demo/role-requests/${encodeURIComponent(id)}/approve`, {}),
  pools: () => getJson<PoolsResponse>('/api/v1/demo/pools?eventId=ev-demo'),
  generatePools: (body: GeneratePoolsRequest) =>
    postJson<PoolsResponse>('/api/v1/demo/pools/generate', body),
  event: () => getJson<EventDisplayInfo>('/api/v1/demo/event'),
  officials: (fightId?: string) =>
    getJson<{ officials: DisplayOfficial[] }>(
      `/api/v1/demo/officials${fightId ? `?fightId=${encodeURIComponent(fightId)}` : ''}`,
    ),
  standings: () => getJson<{ standings: StandingRow[] }>('/api/v1/demo/standings'),
};

// Polls a data source and keeps all open display tabs live-synced.
export function livePoll<T>(load: () => Promise<T>, intervalMs: number, onData: (data: T) => void) {
  let stopped = false;
  const tick = async () => {
    if (stopped) return;
    try {
      onData(await load());
    } catch {
      // Keep the last good frame when the venue server is briefly unavailable.
    }
  };
  void tick();
  const handle = setInterval(() => void tick(), intervalMs);
  return () => {
    stopped = true;
    clearInterval(handle);
  };
}

// Broadcasts announcer/PA messages to every open tab via localStorage events.
export function broadcastPA(message: string, fighterId = '', seconds = 6) {
  const payload = { message, fighterId, seconds, at: Date.now() };
  localStorage.setItem('j3-pa', JSON.stringify(payload));
  window.dispatchEvent(new CustomEvent('j3-pa', { detail: payload }));
}

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
