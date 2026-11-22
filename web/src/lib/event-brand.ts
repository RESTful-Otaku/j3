import { api, type EventDisplayInfo } from './api';
import { eventTitle } from './event-title';

const browser = typeof window !== 'undefined';

type JsonObject = Record<string, unknown>;

export type EventBrand = {
  title: string;
  bannerUrl: string | null;
  logoUrl: string | null;
  primary: string | null;
  secondary: string | null;
  accent: string | null;
};

let loaded = false;
let brand: EventBrand | null = null;

function object(value: unknown): JsonObject {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return {};
  return value as JsonObject;
}

function parseObject(value: string | undefined | null): JsonObject {
  if (!value) return {};
  try {
    return object(JSON.parse(value));
  } catch {
    return {};
  }
}

function text(value: unknown): string | null {
  return typeof value === 'string' && value.trim() ? value.trim() : null;
}

function cssColour(value: unknown): string | null {
  const colour = text(value);
  if (!colour || /[;{}<>]/.test(colour)) return null;
  const supported = typeof CSS !== 'undefined' && typeof CSS.supports === 'function';
  if (supported ? CSS.supports('color', colour) : /^#(?:[0-9a-f]{3,4}|[0-9a-f]{6}|[0-9a-f]{8})$/i.test(colour)) {
    return colour;
  }
  return null;
}

function assetUrl(value: unknown): string | null {
  const raw = text(value);
  if (!raw) return null;
  try {
    const url = new URL(raw, browser ? window.location.origin : 'http://localhost');
    if (!['http:', 'https:'].includes(url.protocol) || url.username || url.password) return null;
    return raw;
  } catch {
    return null;
  }
}

function rgb(colour: string): [number, number, number] | null {
  const hex = colour.match(/^#([0-9a-f]{3,4}|[0-9a-f]{6}|[0-9a-f]{8})$/i);
  if (hex) {
    let value = hex[1];
    if (value.length === 3 || value.length === 4) value = value.split('').map((c) => c + c).join('');
    return [parseInt(value.slice(0, 2), 16), parseInt(value.slice(2, 4), 16), parseInt(value.slice(4, 6), 16)];
  }
  const fn = colour.match(/^rgba?\(\s*([\d.]+)\s*,\s*([\d.]+)\s*,\s*([\d.]+)/i);
  if (fn) return [Number(fn[1]), Number(fn[2]), Number(fn[3])];
  return null;
}

/** Pick black or white text so organiser-chosen colours stay readable. */
export function readableInk(colour: string): string {
  const parts = rgb(colour);
  if (!parts) return '#ffffff';
  const [r, g, b] = parts.map((channel) => {
    const value = channel / 255;
    return value <= 0.03928 ? value / 12.92 : Math.pow((value + 0.055) / 1.055, 2.4);
  });
  return 0.2126 * r + 0.7152 * g + 0.0722 * b > 0.179 ? '#0f172a' : '#ffffff';
}

function readBrand(data: EventDisplayInfo): EventBrand {
  const tokens = parseObject(data.themeTokens);
  const details = parseObject(data.detailsJson);
  const colours = object(tokens.colours ?? tokens.colors);
  const primary = cssColour(tokens.primary ?? tokens.main ?? colours.primary ?? colours.main);
  const secondary = cssColour(tokens.secondary ?? colours.secondary);
  const accent = cssColour(tokens.accent ?? colours.accent ?? colours.highlight);
  const name = text(tokens.title) ?? text(details.title) ?? text(data.name) ?? 'J3 Tournament';
  return {
    title: name,
    bannerUrl: assetUrl(tokens.bannerPath ?? tokens.banner ?? data.bannerPath),
    logoUrl: assetUrl(tokens.logoPath ?? tokens.logo ?? data.logoPath),
    primary,
    secondary,
    accent,
  };
}

/** Apply organiser overrides on top of whichever base theme is active. */
export function applyEventBrand(): void {
  if (!browser || !brand) return;
  const root = document.documentElement;
  const pairs: Array<[string, string]> = [];
  if (brand.primary) {
    pairs.push(['--brand-primary', brand.primary], ['--brand-on-primary', readableInk(brand.primary)]);
  }
  if (brand.secondary) {
    pairs.push(['--brand-secondary', brand.secondary], ['--brand-on-secondary', readableInk(brand.secondary)]);
  }
  if (brand.accent) {
    pairs.push(['--brand-accent', brand.accent], ['--brand-on-accent', readableInk(brand.accent)]);
  }
  if (brand.bannerUrl) pairs.push(['--event-banner', `url(${JSON.stringify(brand.bannerUrl)})`]);
  if (brand.logoUrl) pairs.push(['--event-logo', `url(${JSON.stringify(brand.logoUrl)})`]);
  for (const [key, value] of pairs) root.style.setProperty(key, value);
  if (brand.primary) {
    document.querySelector('meta[name="theme-color"]')?.setAttribute('content', brand.primary);
  }
}

/** Fetch once from the venue API or the bundled offline snapshot. */
export async function loadEventBrand(): Promise<void> {
  if (loaded) return;
  loaded = true;
  try {
    brand = readBrand(await api.event());
    eventTitle.set(brand.title);
    applyEventBrand();
  } catch {
    /* keep the J3 Tournament placeholder and default palette */
  }
}

/** Used after the setup wizard saves a new event identity. */
export async function reloadEventBrand(): Promise<void> {
  loaded = false;
  brand = null;
  await loadEventBrand();
}
