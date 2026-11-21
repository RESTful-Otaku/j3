import { demoSession, postJson, resetAuthCache } from './api';
import { setRole, type Role } from './nav';

export type Session = {
  token: string; email: string; role: string; expiresAt: string; userId?: string;
};

export function uiRole(role: string): Role {
  switch (role) {
    case 'tournament_runner':
    case 'runner':
      return 'runner';
    case 'official':
    case 'moderator':
      return 'official';
    case 'club_leader':
      return 'club_leader';
    case 'fighter':
      return 'fighter';
    default:
      return 'spectator';
  }
}

export function landingForRole(role: string): string {
  switch (uiRole(role)) {
    case 'runner': return '/admin';
    case 'official': return '/official';
    case 'club_leader': return '/club';
    case 'fighter': return '/public';
    default: return '/public';
  }
}

const KEY = 'j3-session';

export function session(): Session | null {
  try {
    const raw = localStorage.getItem(KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as Session;
    if (!parsed.token) return null;
    // Older demo sessions had no expiry; the server still validates them.
    if (parsed.expiresAt && Date.parse(parsed.expiresAt) < Date.now()) {
      localStorage.removeItem(KEY);
      resetAuthCache();
      return null;
    }
    return parsed;
  } catch { return null; }
}

export function saveSession(value: Session) {
  localStorage.setItem(KEY, JSON.stringify(value));
  resetAuthCache();
}
export function clearSession() {
  localStorage.removeItem(KEY);
  resetAuthCache();
}

// Local QA only: replace the bearer session with one issued for another seeded
// user. Both the API identity and the UI role guard must change together.
export async function switchDemoUser(userId: string) {
  const next = await demoSession(userId);
  setRole(uiRole(next.role));
  window.dispatchEvent(new CustomEvent('j3-session-changed', { detail: next }));
  return next;
}

// Keep the client-side route guard aligned with the verified session on load.
export function syncRoleFromSession() {
  const current = session();
  if (current) setRole(uiRole(current.role));
}

export async function login(email: string, provider: string, password?: string) {
  return postJson<{ challengeId: string; mfaRequired: boolean; demoCode: string; provider: string }>(
    '/api/v1/auth/login', { email, password, provider },
  );
}

export async function verify(email: string, code: string) {
  const result = await postJson<Session>('/api/v1/auth/verify', { email, code });
  saveSession(result);
  setRole(uiRole(result.role));
  return result;
}

export async function passkey(email: string) {
  if ('PublicKeyCredential' in window) {
    // Real relying-party verification is configured server-side in production.
    try {
      await navigator.credentials.create({ publicKey: {
        challenge: new Uint8Array(32), rp: { name: 'J3' },
        user: { id: new Uint8Array(16), name: email, displayName: email },
        pubKeyCredParams: [{ type: 'public-key', alg: -7 }],
      } });
    } catch { /* demo falls back to the MFA code */ }
  }
  return login(email, 'passkey');
}
