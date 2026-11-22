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

// Inline until the generated contract includes the demo credential sheet and
// invite responses. Shape mirrors GET /api/v1/demo/credentials and POST
// /api/v1/invites; keeping it here avoids hand-editing api-types.d.ts.
export type DemoCredentialAccount = {
  id: string;
  name: string;
  email: string;
  role: string;
  roleLabel: string;
  club: string;
  password: string;
  status: string;
};
export type DemoCredentialsSheet = {
  password: string;
  accounts: DemoCredentialAccount[];
  quickAccounts: DemoCredentialAccount[];
  total: number;
};
export type InviteRequest = {
  email: string;
  name?: string;
  role: string;
  clubName?: string;
  clubId?: string;
  phone?: string;
  eventId?: string;
};
export type InviteLink = {
  id: string;
  email: string;
  name: string;
  role: string;
  club: string;
  state: string;
  expiresAt: string;
  createdAt: string;
  inviteUrl?: string;
  token?: string;
};
export type Technique = components['schemas']['Technique'];
export type TechniqueGroup = components['schemas']['TechniqueGroup'];
export type TechniquePointOption = components['schemas']['TechniquePointOption'];
export type TechniqueCatalogue = {
  eventId: string;
  count: number;
  techniques: Technique[];
  groups: TechniqueGroup[];
  pointOptions: TechniquePointOption[];
};
export type TechnicalResultRequest = {
  winner: 'red' | 'blue';
  reason: 'injury' | 'forfeit' | 'cheating' | 'dq' | 'no_show' | 'withdrawal' | 'referee' | 'other';
  detail?: string;
};

const SESSION_KEY = 'j3-session';

// The server refuses every protected route until a generated password is
// replaced. Callers can catch this to guide the user back to the change form.
export class PasswordChangeRequiredError extends Error {
  constructor() {
    super('A new password is required before you can continue.');
    this.name = 'PasswordChangeRequiredError';
  }
}

// On web this is empty, so relative URLs keep working. Native builds bake the
// venue server URL into the bundle and load their UI from local Capacitor
// assets, so every API path must be rewritten to that origin.
const API_BASE = (import.meta.env.VITE_API_BASE_URL as string | undefined || '').replace(/\/+$/, '');
const OFFLINE_DEMO = import.meta.env.VITE_OFFLINE_DEMO === '1';

// Native builds ship a seeded mid-event snapshot plus the SQLite file it came
// from. If the venue LAN is unreachable, read-only phone screens fall back to
// that snapshot so the app still opens with real data.
type LocalSnapshot = Record<string, unknown> & { session?: DemoSession };
let localSnapshot: Promise<LocalSnapshot | null> | null = null;

function offlineDemoEnabled() {
  if (OFFLINE_DEMO) return true;
  if (typeof window === 'undefined') return false;
  return new URLSearchParams(window.location.search).get('offline') === '1';
}

function bundledSnapshot(): Promise<LocalSnapshot | null> {
  if (!offlineDemoEnabled()) return Promise.resolve(null);
  if (!localSnapshot) {
    localSnapshot = fetch('/j3-demo-data.json', { cache: 'no-store' })
      .then((response) => response.ok ? response.json() as Promise<LocalSnapshot> : null)
      .catch(() => null);
  }
  return localSnapshot;
}

async function localResponse(url: string): Promise<Response | null> {
  const snapshot = await bundledSnapshot();
  if (!snapshot) return null;
  const value = snapshot[url] ?? snapshot[url.split('?')[0]];
  if (value === undefined) return null;
  return new Response(JSON.stringify(value), {
    status: 200,
    headers: { 'Content-Type': 'application/json' },
  });
}

export function apiUrl(path: string): string {
  if (!API_BASE || !path.startsWith('/')) return path;
  return `${API_BASE}${path}`;
}

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

// Map a canonical RBAC role (or a shortened session role) to the UI role used
// by the route guard and nav. Mirrors demoUIRole on the venue server.
function offlineUiRole(role: string): string {
  switch (role) {
    case 'admin': return 'admin';
    case 'tournament_runner':
    case 'runner': return 'runner';
    case 'official':
    case 'moderator': return 'official';
    case 'club_leader': return 'club_leader';
    case 'fighter': return 'fighter';
    default: return 'spectator';
  }
}

// A bundled snapshot may include the credential sheet as an endpoint key or
// under `credentials`; either shape lets an offline APK sign in as any seeded
// account without a live venue server.
function snapshotSession(snapshot: LocalSnapshot, userId?: string): DemoSession | null {
  // The default snapshot session is fine for auto-login, but an explicit
  // identity switch must come from the credential sheet so every seeded role
  // remains selectable while the LAN server is unavailable.
  if (snapshot.session && !userId) return snapshot.session;
  const raw = snapshot.credentials ?? snapshot['/api/v1/demo/credentials'];
  if (!raw || typeof raw !== 'object') return null;
  const sheet = raw as Partial<DemoCredentialsSheet>;
  const accounts = Array.isArray(sheet.accounts) ? sheet.accounts : [];
  if (!accounts.length) return null;
  const candidates = Array.isArray(sheet.quickAccounts) && sheet.quickAccounts.length
    ? [...sheet.quickAccounts, ...accounts]
    : accounts;
  const account = userId
    ? candidates.find((item) => item.id === userId)
    : candidates.find((item) => item.role === 'admin')
      ?? candidates.find((item) => item.role === 'tournament_runner' || item.role === 'runner')
      ?? accounts[0];
  if (!account?.id) {
    if (userId && snapshot.session?.userId === userId) return snapshot.session;
    return null;
  }
  return {
    token: `offline-${account.id}`,
    email: account.email ?? '',
    role: offlineUiRole(account.role),
    userId: account.id,
    expiresAt: '2099-01-01T00:00:00Z',
  };
}

async function fetchDemoSession(userId?: string): Promise<DemoSession> {
  const init: RequestInit = { method: 'POST' };
  if (userId) {
    init.headers = { 'Content-Type': 'application/json' };
    init.body = JSON.stringify({ userId });
  }
  try {
    const response = await fetch(apiUrl('/api/v1/auth/demo'), init);
    if (!response.ok) throw new Error(`demo sign-in failed (${response.status})`);
    return await response.json() as DemoSession;
  } catch (error) {
    const snapshot = await bundledSnapshot();
    const offline = snapshot ? snapshotSession(snapshot, userId) : null;
    if (offline) return offline;
    throw error;
  }
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
    return await fetch(apiUrl(url), { ...init, headers });
  } catch {
    const local = await localResponse(url);
    if (local) return local;
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
    let detail: { error?: string } = {};
    try { detail = await response.clone().json() as { error?: string }; } catch { /* plain error */ }
    if (response.status === 403 && detail.error === 'password_change_required') {
      if (typeof window !== 'undefined') {
        localStorage.removeItem(SESSION_KEY);
        resetAuthCache();
        window.dispatchEvent(new Event('j3-password-change-required'));
      }
      throw new PasswordChangeRequiredError();
    }
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
  credentials: () => getJson<DemoCredentialsSheet>('/api/v1/demo/credentials'),
  changePassword: (currentPassword: string, newPassword: string) =>
    postJson<{ changed: boolean; mustChangePassword: boolean }>('/api/v1/auth/password', {
      currentPassword, newPassword,
    }),
  accounts: () => getJson<{ accounts: DemoAccount[] }>('/api/v1/demo/accounts'),
  fights: () => getJson<{ fights: DemoFight[] }>('/api/v1/demo/fights'),
  myContests: () => getJson<{ contests: DemoFight[]; role: string }>('/api/v1/demo/my-contests'),
  techniques: (params: { q?: string; group?: string; limit?: number; eventId?: string } = {}) => {
    const query = new URLSearchParams();
    if (params.q) query.set('q', params.q);
    if (params.group) query.set('group', params.group);
    if (params.limit) query.set('limit', String(params.limit));
    if (params.eventId) query.set('eventId', params.eventId);
    return getJson<TechniqueCatalogue>(`/api/v1/demo/techniques?${query}`);
  },
  view: () => getJson<{ actor: DemoAccount; actions: Action[]; counts: DemoCounts }>('/api/v1/demo/view'),
  signupPreview: (body: SignupRequest) =>
    postJson<SignupPreview>('/api/v1/signup/preview', body),
  signup: (body: SignupRequest) => postJson<SignupResult>('/api/v1/signup', body),
  roleRequests: () =>
    getJson<{ requests: RoleRequest[] }>('/api/v1/demo/role-requests'),
  approveRoleRequest: (id: string) =>
    postJson<{ status: string }>(`/api/v1/demo/role-requests/${encodeURIComponent(id)}/approve`, {}),
  pools: () => getJson<PoolsResponse>('/api/v1/demo/pools'),
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
  schedule: () => getJson<{ items: ScheduleItem[] }>('/api/v1/demo/schedule'),
  setDrift: (minutes: number) =>
    postJson<{ items: ScheduleItem[] }>('/api/v1/demo/schedule/drift', { minutes }),
  completeScheduleItem: (id: string) =>
    postJson<{ items: ScheduleItem[] }>(`/api/v1/demo/schedule/${encodeURIComponent(id)}/complete`, {}),
  bracketStages: (eventId?: string) =>
    getJson<{ stages: BracketStage[] }>(
      eventId ? `/api/v1/demo/brackets?eventId=${encodeURIComponent(eventId)}` : '/api/v1/demo/brackets',
    ),
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
  technicalResult: (id: string, body: TechnicalResultRequest) =>
    postJson<FightState>(`/api/v1/fights/${encodeURIComponent(id)}/technical`, body),
  eventStream: (onEvent: (event: { type: string; fightId?: string; payload?: FightState }) => void) => {
    const source = new EventSource(apiUrl('/api/v1/events/stream'));
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
  exportUrl: (type: string) => apiUrl(`/api/v1/demo/export?type=${encodeURIComponent(type)}`),
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
  exportCsv: async (type: string): Promise<{ ok: boolean; error?: string }> => {
    try {
      const headers = await authorisedHeaders();
      const response = await fetch(apiUrl(`/api/v1/demo/export?type=${encodeURIComponent(type)}`), { headers });
      if (!response.ok) return { ok: false, error: `HTTP ${response.status}` };
      const blob = await response.blob();
      const link = document.createElement('a');
      link.href = URL.createObjectURL(blob);
      link.download = `j3-export-${type}-${Date.now()}.csv`;
      document.body.appendChild(link);
      link.click();
      link.remove();
      return { ok: true };
    } catch (e) {
      return { ok: false, error: e instanceof Error ? e.message : String(e) };
    }
  },
  createTournament: (body: CreateTournamentRequest) =>
    postJson<CreateTournamentResponse>('/api/v1/tournaments', body),
  invites: () => getJson<{ invites: InviteLink[] }>('/api/v1/invites'),
  createInvite: (body: InviteRequest) => postJson<InviteLink>('/api/v1/invites', body),
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
  redPenaltyPoints: number; bluePenaltyPoints: number;
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
