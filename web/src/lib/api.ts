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
export type DemoSession = components['schemas']['Session'];
export type DemoIdentity = components['schemas']['DemoIdentity'];
export type DemoIdentitiesResponse = components['schemas']['DemoIdentitiesResponse'];

const SESSION_KEY = 'j3-session';

let authReady: Promise<void> | null = null;
let authRecovery: Promise<boolean> | null = null;
let lastAuthError = '';

// A session can outlive a database reset, so validate its shape and expiry
// before sending it. The server remains the authority on token validity.
function storedToken(): string {
  try {
    const raw = localStorage.getItem(SESSION_KEY);
    if (!raw) return '';
    const parsed = JSON.parse(raw) as { token?: string; expiresAt?: string };
    if (parsed.expiresAt && Date.parse(parsed.expiresAt) < Date.now()) {
      localStorage.removeItem(SESSION_KEY);
      return '';
    }
    return typeof parsed.token === 'string' ? parsed.token : '';
  } catch {
    localStorage.removeItem(SESSION_KEY);
    return '';
  }
}

async function fetchDemoSession(userId?: string): Promise<DemoSession> {
  const init: RequestInit = { method: 'POST' };
  if (userId) {
    init.headers = { 'Content-Type': 'application/json' };
    init.body = JSON.stringify({ userId });
  }
  const response = await fetch('/api/v1/auth/demo', init);
  if (!response.ok) throw new Error(`demo sign-in failed (${response.status})`);
  return response.json() as Promise<DemoSession>;
}

async function demoLogin(): Promise<boolean> {
  lastAuthError = '';
  try {
    const data = await fetchDemoSession();
    localStorage.setItem(SESSION_KEY, JSON.stringify(data));
    return true;
  } catch (e) {
    lastAuthError = e instanceof Error ? e.message : 'demo sign-in failed';
    return false;
  }
}

// Issues a real server-side session for a chosen QA account. Unlike the old
// display-only actor switch, every subsequent API call carries this identity.
export async function demoSession(userId: string): Promise<DemoSession> {
  const data = await fetchDemoSession(userId);
  localStorage.setItem(SESSION_KEY, JSON.stringify(data));
  resetAuthCache();
  return data;
}

async function ensureDemoAuth() {
  if (storedToken()) return;
  await demoLogin();
}

function ensureAuth() {
  if (!authReady) authReady = ensureDemoAuth();
  return authReady;
}

// Called by the login/sign-out store and by 401 recovery.
export function resetAuthCache() { authReady = null; }

async function authorisedHeaders(init?: RequestInit): Promise<Headers> {
  await ensureAuth();
  const headers = new Headers(init?.headers);
  headers.set('Accept', 'application/json');
  const token = storedToken();
  if (token) headers.set('Authorization', `Bearer ${token}`);
  return headers;
}

async function recoverAuth(): Promise<boolean> {
  localStorage.removeItem(SESSION_KEY);
  resetAuthCache();
  const ok = await demoLogin();
  resetAuthCache();
  return ok && Boolean(storedToken());
}

// Concurrent requests that all receive a 401 must share one re-authentication.
function recoverAuthOnce(): Promise<boolean> {
  if (!authRecovery) {
    authRecovery = recoverAuth().finally(() => { authRecovery = null; });
  }
  return authRecovery;
}

function isAuthEndpoint(url: string) {
  return url.startsWith('/api/v1/auth/') || url.startsWith('/api/v1/signup');
}

async function send(url: string, init: RequestInit | undefined, headers: Headers): Promise<Response> {
  try {
    return await fetch(url, { ...init, headers });
  } catch {
    throw new Error('Cannot reach the venue server — check that J3 is running');
  }
}

async function requestJson<T>(url: string, init?: RequestInit): Promise<T> {
  let headers = await authorisedHeaders(init);
  let response = await send(url, init, headers);
  if (response.status === 401 && !isAuthEndpoint(url) && await recoverAuthOnce()) {
    headers = await authorisedHeaders(init);
    response = await send(url, init, headers);
  }
  if (!response.ok) {
    const hint = response.status === 401 && lastAuthError ? `: ${lastAuthError}` : '';
    throw new Error(`${response.status} ${response.statusText} from ${url}${hint}`);
  }
  try {
    return (await response.json()) as T;
  } catch {
    throw new Error('Venue server returned an invalid response');
  }
}

export const getJson = <T>(url: string) => requestJson<T>(url);
export const postJson = <T>(url: string, body: unknown) => requestJson<T>(url, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
  body: JSON.stringify(body),
});

export const api = {
  health: () => getJson<{ status: string }>('/healthz'),
  version: () => getJson<{ name: string; version: string }>('/api/v1/version'),
  roles: () => getJson<{ roles: RoleInfo[] }>('/api/v1/roles'),
  demo: () => getJson<{ counts: DemoCounts }>('/api/v1/demo'),
  identities: () => getJson<DemoIdentitiesResponse>('/api/v1/demo/identities'),
  accounts: () => getJson<{ accounts: DemoAccount[] }>('/api/v1/demo/accounts'),
  fights: () => getJson<{ fights: DemoFight[] }>('/api/v1/demo/fights'),
  view: () => getJson<{ actor: DemoAccount; actions: Action[]; counts: DemoCounts }>('/api/v1/demo/view'),
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
  medals: () => getJson<{ medals: { id: string; fighterId: string; name: string; club: string; clubColour: string; kind: string }[] }>('/api/v1/demo/medals'),
  kataCatalogue: () => getJson<{ kata: { id: string; name: string; style: string; level: string; discipline: string; isWeapon: boolean }[] }>('/api/v1/demo/kata/catalogue'),
  training: () => getJson<{ sessions: { id: string; title: string; kind: string; leader: string; startsAt: string; duration: number }[] }>('/api/v1/demo/training'),
  rotation: (body: { fighters: string[]; mats: string[]; rounds: number }) =>
    postJson<{ pairs: { round: number; mat: string; red: string; blue: string }[]; count: number }>('/api/v1/demo/rotation', body),
  kata: () => getJson<{ performances: KataPerformance[] }>('/api/v1/demo/kata'),
  createKata: (body: Record<string, unknown>) =>
    postJson<{ performances: KataPerformance[] }>('/api/v1/demo/kata/performances', body),
  kataScore: (id: string, body: Record<string, unknown>) =>
    postJson<{ performances: KataPerformance[] }>(`/api/v1/demo/kata/${encodeURIComponent(id)}/scores`, body),
  advanceKata: (advanceCount = 8) =>
    postJson<{ advanced: number; round: number; performances: KataPerformance[] }>('/api/v1/demo/kata/advance', { advanceCount }),
  schedule: () => getJson<{ items: ScheduleItem[] }>('/api/v1/demo/schedule?eventId=ev-demo'),
  setDrift: (minutes: number) =>
    postJson<{ items: ScheduleItem[] }>('/api/v1/demo/schedule/drift', { minutes }),
  completeScheduleItem: (id: string) =>
    postJson<{ items: ScheduleItem[] }>(`/api/v1/demo/schedule/${encodeURIComponent(id)}/complete`, {}),
  bracketStages: (eventId = 'ev-demo') =>
    getJson<{ stages: BracketStage[] }>(`/api/v1/demo/brackets?eventId=${encodeURIComponent(eventId)}`),
  buildBracket: (body: Record<string, unknown>) =>
    postJson<{ format: string; stages: BracketStage[] }>('/api/v1/demo/brackets/build', body),
  promoteSlot: (id: string, body: Record<string, unknown>) =>
    postJson<{ stages: BracketStage[] }>(`/api/v1/demo/brackets/slots/${encodeURIComponent(id)}/promote`, body),
  fightState: (id: string) => getJson<FightState>(`/api/v1/fights/${encodeURIComponent(id)}/state`),
  startFight: (id: string, seconds = 180, round = 1) =>
    postJson<FightState>(`/api/v1/fights/${encodeURIComponent(id)}/start`, { seconds, round }),
  pauseFight: (id: string) => postJson<FightState>(`/api/v1/fights/${encodeURIComponent(id)}/pause`, {}),
  resumeFight: (id: string) => postJson<FightState>(`/api/v1/fights/${encodeURIComponent(id)}/resume`, {}),
  fightEvent: (id: string, body: Record<string, unknown>) =>
    postJson<FightState>(`/api/v1/fights/${encodeURIComponent(id)}/event`, body),
  startHold: (id: string, corner: string, seconds = 20) =>
    postJson<FightState>(`/api/v1/fights/${encodeURIComponent(id)}/hold`, { corner, seconds }),
  undoFightEvent: (id: string) => postJson<FightState>(`/api/v1/fights/${encodeURIComponent(id)}/undo`, {}),
  submitFight: (id: string, outcome: string) =>
    postJson<FightState>(`/api/v1/fights/${encodeURIComponent(id)}/submit`, { outcome }),
  eventStream: (onEvent: (event: { type: string; fightId?: string; payload?: FightState }) => void) => {
    const source = new EventSource('/api/v1/events/stream');
    source.onmessage = (message) => {
      try { onEvent(JSON.parse(message.data)); } catch { /* ignore malformed frames */ }
    };
    return () => source.close();
  },
  checkIn: (id: string, weightKg: number) =>
    postJson<{ entrantId: string; weightKg: number; outOfClass: boolean }>(`/api/v1/demo/entrants/${encodeURIComponent(id)}/checkin`, { weightKg }),
  teamRankings: () => getJson<{ teams: { club: string; gold: number; silver: number; bronze: number; points: number }[] }>('/api/v1/demo/team-rankings'),
  importEntries: (rows: Record<string, unknown>[]) =>
    postJson<{ imported: number }>('/api/v1/demo/import', { rows }),
  exportUrl: (type: string) => `/api/v1/demo/export?type=${encodeURIComponent(type)}`,
  readiness: () => getJson<{ categories: { id: string; name: string; entrants: number; ready: boolean; poolCount: number }[] }>('/api/v1/demo/readiness'),
  autoRun: () => postJson<{ pools: number; stages: number; officials: number }>('/api/v1/demo/auto-run', {}),
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
  upload: async (file: File, kind: string, _role?: string): Promise<{ url: string; colour: string; highlight: string }> => {
    const body = new FormData();
    body.append('kind', kind);
    body.append('file', file);
    let headers = await authorisedHeaders();
    let response = await send('/api/v1/uploads', { method: 'POST', body }, headers);
    if (response.status === 401 && await recoverAuthOnce()) {
      headers = await authorisedHeaders();
      response = await send('/api/v1/uploads', { method: 'POST', body }, headers);
    }
    if (!response.ok) throw new Error(`Upload failed (${response.status})`);
    return response.json() as Promise<{ url: string; colour: string; highlight: string }>;
  },
};

export type KataPerformance = {
  id: string; fighterId: string; name: string; club: string; belt: string;
  kataName: string; roundNumber: number; kataType: string; bunkaiScore: number;
  panelSize: number; status: string; total: number;
  scores: { judge: number; score: number }[];
};

export type ScheduleItem = {
  id: string; name: string; area: string; plannedStart: string;
  projectedStart: string; projectedEnd: string; duration: number;
  status: string; offsetMinutes: number;
};

export type BracketSlot = {
  id: string; position: number; fighterId: string; fighterName: string;
  club: string; clubColour: string; clubLogo: string; belt: string; state: string;
  resultIcon: string; source: string;
};
export type BracketStage = { id: string; name: string; kind: string; slots: BracketSlot[] };

export type FightState = {
  fightId: string; state: string; outcome: string; running: boolean;
  seconds: number; round: number; redScore: number; blueScore: number;
  redPenalty: number; bluePenalty: number;
  holdSeconds: number; holdCorner: string;
  events: { seq: number; kind: string; corner: string; points: number; technique: string; card: string; actor: string; createdAt: string }[];
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
