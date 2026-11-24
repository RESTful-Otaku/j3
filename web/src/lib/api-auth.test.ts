// @vitest-environment happy-dom
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { bootstrapQaSession, getJson, postJson, resetAuthCache, type FightState } from './api';
import { switchDemoUser } from './auth';

type Call = { url: string; auth: string | null };

function stubGlobal(name: string, value: unknown) {
  Object.defineProperty(globalThis, name, { value, configurable: true, writable: true });
}


function mockApi(handler: (url: string, callCount: number) => { status: number; body: unknown }) {
  const calls: Call[] = [];
  stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);
    const auth = new Headers(init?.headers).get('Authorization');
    calls.push({ url, auth });
    const count = calls.filter((call) => call.url === url).length;
    const { status, body } = handler(url, count);
    return new Response(JSON.stringify(body), {
      status,
      headers: { 'Content-Type': 'application/json' },
    });
  }));
  return calls;
}

const freshSession = {
  token: 'fresh', email: 'runner@demo.j3', role: 'runner',
  expiresAt: '2099-01-01T00:00:00Z',
};

describe('venue session handling', () => {
  beforeEach(() => {
    localStorage.clear();
    resetAuthCache();
    vi.restoreAllMocks();
  });

  it('re-authenticates once and retries after a stale session (DB reset)', async () => {
    localStorage.setItem('j3-session', JSON.stringify({ token: 'stale', role: 'runner' }));
    const calls = mockApi((url, count) => {
      if (url.includes('/api/v1/auth/demo')) return { status: 200, body: freshSession };
      if (url.includes('/api/v1/demo/accounts')) {
        return count === 1
          ? { status: 401, body: { error: 'authentication required' } }
          : { status: 200, body: { accounts: [] } };
      }
      return { status: 200, body: {} };
    });

    await expect(getJson('/api/v1/demo/accounts')).resolves.toEqual({ accounts: [] });
    expect(calls.map((call) => call.url)).toEqual([
      '/api/v1/demo/accounts', '/api/v1/auth/demo', '/api/v1/demo/accounts',
    ]);
    expect(calls[0].auth).toBe('Bearer stale');
    expect(calls[2].auth).toBe('Bearer fresh');
    expect(JSON.parse(localStorage.getItem('j3-session')!).token).toBe('fresh');
  });

  it('retries only once and reports a clear error if the session stays invalid', async () => {
    localStorage.setItem('j3-session', JSON.stringify({ token: 'stale', role: 'runner' }));
    const calls = mockApi((url) => {
      if (url.includes('/api/v1/auth/demo')) return { status: 200, body: freshSession };
      return { status: 401, body: { error: 'authentication required' } };
    });

    await expect(getJson('/api/v1/demo/accounts')).rejects.toThrow(/401.*demo/i);
    expect(calls.filter((call) => call.url === '/api/v1/demo/accounts')).toHaveLength(2);
  });

  it('explains a failed demo login instead of failing silently', async () => {
    mockApi((url) => {
      if (url.includes('/api/v1/auth/demo')) return { status: 500, body: { error: 'internal server error' } };
      return { status: 401, body: { error: 'authentication required' } };
    });
    await expect(getJson('/api/v1/demo/accounts')).rejects.toThrow(/demo sign-in failed \(500\)/);
  });

  it('clears an expired local session before attaching it', async () => {
    localStorage.setItem('j3-session', JSON.stringify({
      token: 'old', role: 'runner', expiresAt: '2000-01-01T00:00:00Z',
    }));
    const calls = mockApi((url) => {
      if (url.includes('/api/v1/auth/demo')) return { status: 200, body: freshSession };
      return { status: 200, body: { accounts: [] } };
    });
    await getJson('/api/v1/demo/accounts');
    expect(calls[0].url).toBe('/api/v1/auth/demo');
    expect(calls[1].auth).toBe('Bearer fresh');
  });

  it('switching demo identity replaces the bearer token and the UI role guard', async () => {
    localStorage.setItem('j3-session', JSON.stringify(freshSession));
    const calls = mockApi((url) => {
      if (url.includes('/api/v1/auth/demo')) {
        return { status: 200, body: {
          token: 'fighter-token', email: 'fighter@demo.j3', role: 'fighter',
          userId: 'u-fighter', expiresAt: '2099-01-01T00:00:00Z',
        } };
      }
      return { status: 200, body: {} };
    });
    const next = await switchDemoUser('u-fighter');
    expect(next.token).toBe('fighter-token');
    expect(calls[0].url).toBe('/api/v1/auth/demo');
    expect(JSON.parse(localStorage.getItem('j3-session')!).token).toBe('fighter-token');
    expect(localStorage.getItem('j3-role')).toBe('fighter');

    await getJson('/api/v1/demo/standings');
    expect(calls.at(-1)?.auth).toBe('Bearer fighter-token');
  });

  it('serves and mutates bundled demo data when the WebView has no venue server', async () => {
    window.history.replaceState({}, '', '/?offline=1');
    stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes('/j3-demo-data.json')) {
        return new Response(JSON.stringify({
          session: {
            token: 'offline-demo', email: 'owner@demo.j3', role: 'admin',
            userId: 'u-owner', expiresAt: '2099-01-01T00:00:00Z',
          },
          '/api/v1/demo/standings': { standings: [{ fighterId: 'f01', wins: 3 }] },
          '/api/v1/query?table=fighters&limit=100': {
            table: 'fighters', columns: ['id'], rows: [{ id: 'f01' }], count: 1,
          },
          '/api/v1/fights/fight-01/state': {
            fightId: 'fight-01', state: 'scheduled', outcome: '', running: false,
            seconds: 180, round: 1, redScore: 0, blueScore: 0, redPenalty: 0,
            bluePenalty: 0, redPenaltyPoints: 0, bluePenaltyPoints: 0,
            holdSeconds: 0, holdCorner: '', events: [],
          },
        }), { status: 200, headers: { 'Content-Type': 'application/json' } });
      }
      return new Response('<!doctype html><html>offline</html>', {
        status: 200,
        headers: { 'Content-Type': 'text/html' },
      });
    }));
    localStorage.setItem('j3-session', JSON.stringify(freshSession));

    await expect(getJson('/api/v1/demo/standings')).resolves.toEqual({
      standings: [{ fighterId: 'f01', wins: 3 }],
    });
    await expect(getJson('/api/v1/query?table=fighters&limit=50&offset=0&order=asc')).resolves.toEqual({
      table: 'fighters', columns: ['id'], rows: [{ id: 'f01' }], count: 1,
    });
    const started = await postJson<FightState>('/api/v1/fights/fight-01/start', { seconds: 180, round: 1 });
    expect(started.running).toBe(true);
    const scored = await postJson<FightState>('/api/v1/fights/fight-01/event', {
      kind: 'score', corner: 'red', points: 10, techniqueId: 'ippon',
    });
    expect(scored.redScore).toBe(10);
    await expect(getJson('/api/v1/fights/fight-01/state')).resolves.toMatchObject({
      redScore: 10, state: 'live',
    });
  });

  it('QA bootstrap signs in as the bundled admin when the venue server is offline', async () => {
    window.history.replaceState({}, '', '/?offline=1');
    stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes('/j3-demo-data.json')) {
        return new Response(JSON.stringify({
          session: {
            token: 'offline-demo', email: 'owner@demo.j3', role: 'admin',
            userId: 'u-owner', expiresAt: '2099-01-01T00:00:00Z',
          },
        }), { status: 200, headers: { 'Content-Type': 'application/json' } });
      }
      throw new Error('offline');
    }));

    const next = await bootstrapQaSession();
    expect(next?.role).toBe('admin');
    expect(next?.userId).toBe('u-owner');
    expect(localStorage.getItem('j3-session')).toContain('offline-demo');
  });
});
