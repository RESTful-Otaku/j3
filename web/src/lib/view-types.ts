export type WidgetType =
  | 'clock' | 'scoreboard' | 'upcoming' | 'standings' | 'brackets' | 'medals'
  | 'charts' | 'kata' | 'pools' | 'pa' | 'details' | 'audit';

export type ViewWidgetConfig = {
  id: string;
  type: WidgetType;
  title?: string;
  span: 1 | 2 | 3;
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

export function loadViews(): SavedView[] {
  try { return JSON.parse(localStorage.getItem(KEY) || '[]'); } catch { return []; }
}

export function saveViews(views: SavedView[]) {
  localStorage.setItem(KEY, JSON.stringify(views));
}
