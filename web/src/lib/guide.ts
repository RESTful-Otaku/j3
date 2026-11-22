// Shared helper for the hands-on walkthrough engine.
// Pages call guideAction() only AFTER an action has actually succeeded, so the
// guide can honestly refuse to advance until the user has done the real thing.

export const GUIDE_ACTION_EVENT = 'j3-guide-action';
const ACTION_KEY = 'j3-guide-actions';

export type GuideActionDetail = { key: string; [field: string]: unknown };

export function guideAction(key: string, detail: Record<string, unknown> = {}) {
  if (typeof window === 'undefined') return;
  recordGuideAction(key);
  window.dispatchEvent(
    new CustomEvent<GuideActionDetail>(GUIDE_ACTION_EVENT, { detail: { key, ...detail } }),
  );
}

export function recordGuideAction(key: string) {
  if (typeof sessionStorage === 'undefined') return;
  try {
    const seen = new Set<string>(JSON.parse(sessionStorage.getItem(ACTION_KEY) || '[]'));
    seen.add(key);
    sessionStorage.setItem(ACTION_KEY, JSON.stringify([...seen]));
  } catch {
    sessionStorage.setItem(ACTION_KEY, JSON.stringify([key]));
  }
}

export function hasGuideAction(key: string) {
  if (typeof sessionStorage === 'undefined') return false;
  try {
    const seen: string[] = JSON.parse(sessionStorage.getItem(ACTION_KEY) || '[]');
    return seen.includes(key);
  } catch {
    return false;
  }
}

export function clearGuideActions() {
  if (typeof sessionStorage !== 'undefined') sessionStorage.removeItem(ACTION_KEY);
}
