import { get, writable } from 'svelte/store';
import { applyEventBrand } from './event-brand';

export type ThemeChoice = 'system' | 'light' | 'dark' | 'street';
export type ResolvedTheme = Extract<ThemeChoice, 'light' | 'dark' | 'street'>;

const KEY = 'j3-theme';
const LEGACY_KEY = 'j3-settings-v1';
const choices: ThemeChoice[] = ['system', 'light', 'dark', 'street'];

/** Shared appearance choice. All screens use this instead of local copies. */
export const themeChoice = writable<ThemeChoice>('street');

const browser = typeof window !== 'undefined';

let initialised = false;

function isChoice(value: unknown): value is ThemeChoice {
  return typeof value === 'string' && choices.includes(value as ThemeChoice);
}

export function resolveTheme(choice: ThemeChoice): ResolvedTheme {
  if (choice !== 'system') return choice;
  if (!browser) return 'street';
  return window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark';
}

function apply(choice: ThemeChoice): void {
  if (!browser) return;
  const resolved = resolveTheme(choice);
  document.documentElement.dataset.theme = resolved;
  document.documentElement.style.colorScheme = resolved === 'light' ? 'light' : 'dark';
  applyEventBrand();
}

function readStoredChoice(): ThemeChoice {
  if (!browser) return 'street';
  const direct = localStorage.getItem(KEY);
  if (isChoice(direct)) return direct;
  try {
    const legacy = JSON.parse(localStorage.getItem(LEGACY_KEY) || '{}') as { appearance?: unknown };
    if (isChoice(legacy.appearance)) return legacy.appearance;
  } catch {
    /* ignore malformed legacy settings */
  }
  return 'street';
}

/** Call once from the root layout. Keeps storage, DOM and event overrides in sync. */
export function initTheme(): void {
  if (!browser || initialised) return;
  initialised = true;
  themeChoice.set(readStoredChoice());
  themeChoice.subscribe((value) => {
    localStorage.setItem(KEY, value);
    apply(value);
  });
  window.matchMedia('(prefers-color-scheme: light)').addEventListener('change', () => {
    apply(get(themeChoice));
  });
}

export function setTheme(choice: ThemeChoice): void {
  themeChoice.set(choice);
}

/** Cycle used by the dashboard quick switch. */
export function cycleTheme(): void {
  const order: ThemeChoice[] = ['street', 'light', 'dark'];
  const current = get(themeChoice);
  setTheme(order[(order.indexOf(current === 'system' ? 'street' : current) + 1) % order.length]);
}

export { applyEventBrand };
