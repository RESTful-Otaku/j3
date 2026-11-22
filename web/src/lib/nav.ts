export type Role = 'runner' | 'official' | 'club_leader' | 'fighter' | 'spectator';

export type NavGroup = 'Run today' | 'Plan & setup' | 'Manage' | 'Learn' | 'QA tools';

export type NavItem = {
  href: string;
  label: string;
  icon: string;
  roles: Role[];
  group: NavGroup;
};

const all: Role[] = ['runner', 'official', 'club_leader', 'fighter', 'spectator'];
const operators: Role[] = ['runner', 'official'];

// One canonical label and icon per route. Reuse these in dashboards and control
// screens instead of inventing a second vocabulary for the same destination.
export const NAV_ITEMS: NavItem[] = [
  { href: '/public', label: 'Public board', icon: '📺', roles: all, group: 'Run today' },
  { href: '/display', label: 'Live scoreboard', icon: '🥊', roles: operators, group: 'Run today' },
  { href: '/mat', label: 'Mat control', icon: '🎛️', roles: operators, group: 'Run today' },
  { href: '/official', label: 'Official console', icon: '🎯', roles: operators, group: 'Run today' },
  { href: '/operations', label: 'Check-in & weigh-in', icon: '📥', roles: ['runner', 'official', 'club_leader'], group: 'Run today' },
  { href: '/brackets', label: 'Brackets', icon: '🏆', roles: all, group: 'Run today' },
  { href: '/tables', label: 'Tables & rankings', icon: '📊', roles: all, group: 'Run today' },
  { href: '/pools', label: 'Fight pools', icon: '🧩', roles: all, group: 'Run today' },
  { href: '/ceremony', label: 'Medal ceremony', icon: '🏅', roles: all, group: 'Run today' },
  { href: '/control', label: 'Display control', icon: '🖥️', roles: operators, group: 'Run today' },
  { href: '/overlay', label: 'Broadcast overlay', icon: '📡', roles: operators, group: 'Run today' },
  { href: '/event', label: 'Event config', icon: '📋', roles: all, group: 'Run today' },

  { href: '/setup', label: 'Create tournament', icon: '✨', roles: ['runner'], group: 'Plan & setup' },
  { href: '/builder', label: 'View builder', icon: '🧱', roles: operators, group: 'Plan & setup' },

  { href: '/club', label: 'My club', icon: '👥', roles: ['club_leader'], group: 'Manage' },
  { href: '/audit', label: 'Audit trail', icon: '📜', roles: operators, group: 'Manage' },
  { href: '/data', label: 'Data explorer', icon: '🔎', roles: operators, group: 'Manage' },
  { href: '/admin', label: 'Account admin', icon: '👥', roles: ['runner'], group: 'Manage' },
  { href: '/notifications', label: 'Notifications', icon: '🔔', roles: ['runner', 'official', 'club_leader'], group: 'Manage' },
  { href: '/permissions', label: 'Access control', icon: '🔐', roles: ['runner'], group: 'Manage' },

  { href: '/guides', label: 'Guided walkthroughs', icon: '🎓', roles: all, group: 'Learn' },
  { href: '/docs', label: 'Documentation', icon: '📚', roles: all, group: 'Learn' },
  { href: '/settings', label: 'Settings', icon: '⚙️', roles: all, group: 'Learn' },

  { href: '/simulate', label: 'Event simulator', icon: '🎬', roles: operators, group: 'QA tools' },
  { href: '/ui', label: 'Component gallery', icon: '🧩', roles: ['runner'], group: 'QA tools' },
];

export const NAV_GROUPS: NavGroup[] = ['Run today', 'Plan & setup', 'Manage', 'Learn', 'QA tools'];

// The public board is the always-available home for every role: the J3 icon in
// the top-left takes users here so they can never get stranded in a deep view.
export const HOME_ROUTE = '/public';

// A small, persistent mobile nav keeps the main destinations reachable without
// forcing everyone through the full drawer.
export const MOBILE_PRIMARY_HREFS = ['/public', '/mat', '/brackets', '/tables', '/guides'];

const KEY = 'j3-role';

export function currentRole(): Role {
  const value = localStorage.getItem(KEY);
  return all.includes(value as Role) ? (value as Role) : 'runner';
}

export function setRole(role: Role) {
  localStorage.setItem(KEY, role);
}

export function itemsForRole(role: Role) {
  return NAV_ITEMS.filter((item) => item.roles.includes(role));
}

// Public, read-only or demo destinations that do not require a nav entry.
export const OPEN_ROUTES = new Set([
  '/', '/login', '/signup', '/welcome', '/public', '/brackets', '/tables', '/pools',
  '/event', '/details', '/simulate', '/view', '/ui', '/apidocs', '/overlay',
]);

export function canAccessPath(role: Role, path: string) {
  if (OPEN_ROUTES.has(path) || path.startsWith('/docs')) return true;
  return itemsForRole(role).some((item) => item.href === path);
}
