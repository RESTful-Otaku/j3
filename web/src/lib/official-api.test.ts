import { beforeEach, describe, expect, it, vi } from 'vitest';
import { api, resetAuthCache } from './api';

const store = new Map<string, string>();

beforeEach(() => {
  store.clear();
  store.set('j3-session', JSON.stringify({
    token: 'tok-qa',
    expiresAt: new Date(Date.now() + 3_600_000).toISOString(),
  }));
  vi.stubGlobal('localStorage', {
    getItem: (key: string) => store.get(key) ?? null,
    setItem: (key: string, value: string) => { store.set(key, String(value)); },
    removeItem: (key: string) => { store.delete(key); },
  });
  resetAuthCache();
});

describe('official console API client', () => {
  it('passes search, group, limit and event filters to the catalogue', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ techniques: [], groups: [], pointOptions: [], count: 0 }),
    });
    vi.stubGlobal('fetch', fetchMock);
    await api.techniques({ q: 'seoi', group: 'throws', limit: 25, eventId: 'ev-okami' });
    const url = String(fetchMock.mock.calls[0][0]);
    expect(url).toContain('/api/v1/demo/techniques?');
    expect(url).toContain('q=seoi');
    expect(url).toContain('group=throws');
    expect(url).toContain('limit=25');
    expect(url).toContain('eventId=ev-okami');
  });

  it('posts a technical result with winner and reason', async () => {
    const fetchMock = vi.fn().mockResolvedValue({ ok: true, json: async () => ({ state: 'forfeit' }) });
    vi.stubGlobal('fetch', fetchMock);
    await api.technicalResult('ok-fight-24', { winner: 'blue', reason: 'injury', detail: 'shoulder' });
    const url = String(fetchMock.mock.calls[0][0]);
    const init = fetchMock.mock.calls[0][1] as RequestInit;
    expect(url).toContain('/api/v1/fights/ok-fight-24/technical');
    expect(init.method).toBe('POST');
    expect(JSON.parse(String(init.body))).toEqual({
      winner: 'blue', reason: 'injury', detail: 'shoulder',
    });
  });
});
