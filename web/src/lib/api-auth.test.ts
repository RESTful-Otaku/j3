// @vitest-environment happy-dom
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { getJson, resetAuthCache } from './api';

type Call = { url: string; auth: string | null };

function mockApi(handler: (url: string, callCount: number) => { status: number; body: unknown }) {
  const calls: Call[] = [];
  vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
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
    vi.unstubAllGlobals();
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
});
