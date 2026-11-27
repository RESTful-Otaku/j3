export function sanitizeText(value: unknown, max = 1000): string {
  if (value === null || value === undefined) return '';
  return String(value).replace(/[<>]/g, '').replace(/\u0000/g, '').trim().slice(0, max);
}
export function sanitizeUrl(value: unknown): string | null {
  if (typeof value !== 'string') return null;
  const raw = value.trim();
  if (!raw) return null;
  try {
    const u = new URL(raw, typeof location !== 'undefined' ? location.origin : 'http://localhost');
    if (!['http:', 'https:', 'mailto:'].includes(u.protocol)) return null;
    if (u.username || u.password) return null;
    if (u.protocol === 'file:') return null;
    return raw;
  } catch { return null; }
}
export function humanizeKey(key: string): string {
  return String(key).replace(/[_-]+/g, ' ').replace(/([a-z0-9])([A-Z])/g, '$1 $2').replace(/\b\w/g, (c) => c.toUpperCase());
}
