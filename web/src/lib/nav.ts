export type Role = 'runner' | 'official' | 'club_leader' | 'fighter' | 'spectator';

export type NavItem = {
  href: string; label: string; icon: string; roles: Role[];
};

export const NAV_ITEMS: NavItem[] = [
  { href: '/', label: 'Dashboard', icon: '🏠', roles: ['runner', 'official', 'club_leader', 'fighter', 'spectator'] },
  { href: '/public', label: 'Public board', icon: '📺', roles: ['runner', 'official', 'club_leader', 'fighter', 'spectator'] },
  { href: '/display', label: 'Live scoreboard', icon: '🥊', roles: ['runner', 'official'] },
  { href: '/ceremony', label: 'Medal ceremony', icon: '🏅', roles: ['runner', 'official', 'club_leader', 'fighter', 'spectator'] },
  { href: '/official', label: 'Official console', icon: '🎯', roles: ['runner', 'official'] },
  { href: '/mat', label: 'Mat control', icon: '🎛️', roles: ['runner', 'official'] },
  { href: '/brackets', label: 'Brackets', icon: '🏆', roles: ['runner', 'official', 'club_leader', 'fighter', 'spectator'] },
  { href: '/tables', label: 'Tables & rankings', icon: '📊', roles: ['runner', 'official', 'club_leader', 'fighter', 'spectator'] },
  { href: '/pools', label: 'Fight pools', icon: '🧩', roles: ['runner', 'official', 'club_leader', 'fighter', 'spectator'] },
  { href: '/event', label: 'Event config', icon: '📋', roles: ['runner', 'official', 'club_leader', 'fighter', 'spectator'] },
  { href: '/simulate', label: 'Simulator', icon: '🎬', roles: ['runner', 'official'] },
  { href: '/setup', label: 'Setup wizard', icon: '✨', roles: ['runner'] },
  { href: '/builder', label: 'View builder', icon: '🧱', roles: ['runner', 'official'] },
  { href: '/operations', label: 'Check-in & import', icon: '📥', roles: ['runner', 'official', 'club_leader'] },
  { href: '/audit', label: 'Audit trail', icon: '📜', roles: ['runner', 'official'] },
  { href: '/data', label: 'Data explorer', icon: '🔎', roles: ['runner', 'official'] },
  { href: '/control', label: 'Display control', icon: '🖥️', roles: ['runner', 'official'] },
  { href: '/admin', label: 'Account admin', icon: '👥', roles: ['runner'] },
  { href: '/club', label: 'My club', icon: '🏠', roles: ['club_leader'] },
  { href: '/notifications', label: 'Notifications', icon: '🔔', roles: ['runner', 'official', 'club_leader'] },
  { href: '/permissions', label: 'Access control', icon: '🔐', roles: ['runner'] },
  { href: '/apidocs', label: 'API explorer', icon: '🔌', roles: ['runner', 'official', 'club_leader'] },
  { href: '/settings', label: 'Settings', icon: '⚙️', roles: ['runner', 'official', 'club_leader', 'fighter', 'spectator'] },
  { href: '/ui', label: 'Components', icon: '🧩', roles: ['runner'] },
];

const KEY = 'j3-role';

export function currentRole(): Role {
  const value = localStorage.getItem(KEY);
  return (['runner', 'official', 'club_leader', 'fighter', 'spectator'] as Role[]).includes(value as Role)
    ? (value as Role) : 'runner';
}

export function setRole(role: Role) { localStorage.setItem(KEY, role); }

export function itemsForRole(role: Role) {
  return NAV_ITEMS.filter((item) => item.roles.includes(role));
}
