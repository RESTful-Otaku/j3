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
export type BeltOption = components['schemas']['BeltOption'];
export type EventDisplayInfo = components['schemas']['EventDisplayInfo'];
export type StandingRow = components['schemas']['StandingRow'];
export type AuditEntry = components['schemas']['AuditEntry'];
export type CreateTournamentRequest = components['schemas']['CreateTournamentRequest'];
export type CreateTournamentResponse = components['schemas']['CreateTournamentResponse'];
export type Template = components['schemas']['Template'];
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
  simulate: (seed = 42, fighters = 50) =>
    getJson<SimulationResult>(`/api/v1/demo/simulate?seed=${seed}&fighters=${fighters}`),
  stats: (seed = 42, fighters = 50) =>
    getJson<StatsResult>(`/api/v1/demo/stats?seed=${seed}&fighters=${fighters}`),
  belts: () => getJson<{ belts: BeltOption[] }>('/api/v1/demo/belts'),
  event: () => getJson<EventDisplayInfo>('/api/v1/demo/event'),
  officials: (fightId?: string) =>
    getJson<{ officials: DisplayOfficial[] }>(
      `/api/v1/demo/officials${fightId ? `?fightId=${encodeURIComponent(fightId)}` : ''}`,
    ),
  standings: () => getJson<{ standings: StandingRow[] }>('/api/v1/demo/standings'),
  audit: (params: { actor?: string; action?: string; limit?: number } = {}) => {
    const query = new URLSearchParams();
    if (params.actor) query.set('actor', params.actor);
    if (params.action) query.set('action', params.action);
    if (params.limit) query.set('limit', String(params.limit));
    return getJson<{ entries: AuditEntry[] }>(`/api/v1/demo/audit?${query}`);
  },
  query: (params: Record<string, string | number>) => {
    const query = new URLSearchParams();
    for (const [key, value] of Object.entries(params)) query.set(key, String(value));
    return getJson<{ table: string; columns: string[]; rows: Record<string, unknown>[]; count: number }>(
      `/api/v1/query?${query}`,
    );
  },
  templates: () => getJson<{ templates: Template[] }>('/api/v1/templates'),
  createTournament: (body: CreateTournamentRequest) =>
    postJson<CreateTournamentResponse>('/api/v1/tournaments', body),
  upload: async (file: File, kind: string, role: string): Promise<{ url: string; colour: string; highlight: string }> => {
    const body = new FormData();
    body.append('kind', kind);
    body.append('ownerRole', role);
    body.append('file', file);
    const response = await fetch('/api/v1/uploads', { method: 'POST', body });
    if (!response.ok) throw new Error(`Upload failed (${response.status})`);
    return response.json() as Promise<{ url: string; colour: string; highlight: string }>;
  },
};

export type SimulationBout = {
  id: string; stage: string; round: string; redId: string; blueId: string;
  outcome: string; method: string; pointsRed: number; pointsBlue: number; sequence: number;
};
export type SimulationResult = {
  bouts: SimulationBout[];
  medals: { fighterId: string; kind: string }[];
  standings: { fighterId: string; name: string; club: string; wins: number; losses: number; draws: number; points: number }[];
  topMoves: Record<string, number>;
  clubPoints: Record<string, number>;
  clubMedals: Record<string, number>;
  fighterMedals: Record<string, number>;
  seed: number;
};

export type StatsResult = {
  topMoves: Record<string, number>;
  clubPoints: Record<string, number>;
  clubMedals: Record<string, number>;
  fighterMedals: Record<string, number>;
  standings: SimulationResult['standings'];
  medals: SimulationResult['medals'];
  boutCount: number;
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
