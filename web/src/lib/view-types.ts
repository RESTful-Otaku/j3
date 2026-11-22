export type WidgetType =
  | 'clock' | 'scoreboard' | 'upcoming' | 'standings' | 'brackets' | 'medals'
  | 'charts' | 'kata' | 'pools' | 'pa' | 'details' | 'audit';

export type WidgetSize = 1 | 2 | 3;

export type ViewWidgetConfig = {
  id: string;
  type: WidgetType;
  title?: string;
  span: WidgetSize;
  rows: WidgetSize;
};

export type SavedView = {
  name: string;
  widgets: ViewWidgetConfig[];
  updatedAt: string;
};

export const WIDGET_PALETTE: { type: WidgetType; label: string; icon: string }[] = [
  { type: 'clock', label: 'Clock', icon: '🕒' },
  { type: 'scoreboard', label: 'Live scoreboard', icon: '🥊' },
  { type: 'upcoming', label: 'Up next', icon: '📋' },
  { type: 'standings', label: 'Standings', icon: '🏆' },
  { type: 'brackets', label: 'Bracket', icon: '🌳' },
  { type: 'medals', label: 'Medals', icon: '🥇' },
  { type: 'charts', label: 'Charts', icon: '📊' },
  { type: 'kata', label: 'Kata panel', icon: '🥋' },
  { type: 'pools', label: 'Fight pools', icon: '🧩' },
  { type: 'pa', label: 'Announcer', icon: '📣' },
  { type: 'details', label: 'Event details', icon: '📋' },
  { type: 'audit', label: 'Audit trail', icon: '📜' },
];

const KEY = 'j3-custom-views';
const TYPES = new Set<string>(WIDGET_PALETTE.map((entry) => entry.type));

/** Clamp a numeric drag delta to the three supported grid sizes. */
export function clampSize(value: number): WidgetSize {
  if (!Number.isFinite(value)) return 1;
  return Math.min(3, Math.max(1, Math.round(value))) as WidgetSize;
}

export const clampSpan = clampSize;
export const clampRows = clampSize;

/** Snap a pointer delta to the next grid track, including the gap. */
export function sizeFromDelta(
  start: WidgetSize,
  deltaPx: number,
  trackPx: number,
  gapPx: number,
): WidgetSize {
  const step = Math.max(1, trackPx + gapPx);
  return clampSize(start + Math.round(deltaPx / step));
}

export type RectLike = { x: number; y: number; width: number; height: number };

/** Nearest-centre insertion index for grid and list reordering. */
export function insertionIndex(pointer: { x: number; y: number }, rects: RectLike[]): number {
  if (rects.length === 0) return 0;
  let best = 0;
  let bestDistance = Number.POSITIVE_INFINITY;
  let after = false;
  rects.forEach((rect, index) => {
    const dx = pointer.x - (rect.x + rect.width / 2);
    const dy = pointer.y - (rect.y + rect.height / 2);
    const distance = dx * dx + dy * dy;
    if (distance < bestDistance) {
      bestDistance = distance;
      best = index;
      after = Math.abs(dx) > Math.abs(dy) ? dx > 0 : dy > 0;
    }
  });
  return best + (after ? 1 : 0);
}

function memoryStorage(): Storage | null {
  const data = new Map<string, string>();
  return {
    get length() { return data.size; },
    clear: () => data.clear(),
    getItem: (key) => data.get(key) ?? null,
    key: (index) => [...data.keys()][index] ?? null,
    removeItem: (key) => { data.delete(key); },
    setItem: (key, value) => { data.set(key, String(value)); },
  };
}

function storage(): Storage | null {
  if (typeof localStorage !== 'undefined') return localStorage;
  return memoryStorage();
}

export function loadViews(): SavedView[] {
  try {
    const parsed = JSON.parse(storage()?.getItem(KEY) || '[]');
    if (!Array.isArray(parsed)) return [];
    return parsed
      .filter((view): view is SavedView => Boolean(view) && typeof view.name === 'string' && Array.isArray(view.widgets))
      .map((view) => ({
        name: view.name,
        updatedAt: typeof view.updatedAt === 'string' ? view.updatedAt : '',
        widgets: view.widgets
          .filter((widget): widget is ViewWidgetConfig => Boolean(widget) && typeof widget.id === 'string' && TYPES.has(widget.type))
          .map((widget) => ({
            ...widget,
            span: clampSize(Number(widget.span) || 1),
            rows: clampSize(Number(widget.rows) || 1),
          })),
      }));
  } catch {
    return [];
  }
}

export function saveViews(views: SavedView[]) {
  storage()?.setItem(KEY, JSON.stringify(views));
}
