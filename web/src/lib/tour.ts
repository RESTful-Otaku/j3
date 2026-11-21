export type Role = 'runner' | 'official' | 'club_leader' | 'fighter' | 'spectator';

export type TourStep = {
  selector: string; title: string; body: string; href?: string;
};

export type Tour = { id: string; label: string; steps: TourStep[] };

export const TOURS: Record<Role, Tour> = {
  runner: {
    id: 'runner', label: 'Tournament runner',
    steps: [
      { selector: '.journey', href: '/', title: 'Your tournament journey', body: 'Start here: create your event, configure it, then run it on the day.' },
      { selector: '.topbar', href: '/setup', title: 'Create in five steps', body: 'The setup wizard guides you through identity, rules, categories, people and review. Everything saves as you go.' },
      { selector: '.card', href: '/event', title: 'Check the event configuration', body: 'Review kit, schedule, rules and staff so competitors know exactly what to expect.' },
      { selector: '.card', href: '/operations', title: 'Check in your athletes', body: 'Record weigh-ins and flag anyone out of class before the first bout.' },
      { selector: '.card', href: '/mat', title: 'Run a mat', body: 'Start fights, award points, pause for the referee and submit results with confirmation.' },
      { selector: '.card', href: '/control', title: 'Put views on screens', body: 'Open the public board, scoreboards and ceremony in separate tabs on different displays.' },
    ],
  },
  official: {
    id: 'official', label: 'Official / moderator',
    steps: [
      { selector: '.topbar', href: '/official', title: 'Your official console', body: 'Select the contest and the active corner, then tap a technique to log it.' },
      { selector: '.card', href: '/official', title: 'Log with confidence', body: 'Every entry shows the English and romaji technique name and its points, and needs confirmation.' },
      { selector: '.card', href: '/mat', title: 'Control the clock', body: 'Start, pause during referee stoppages and resume when the bout restarts.' },
      { selector: '.card', href: '/tables', title: 'Kata judging', body: 'Open the Kata tab to score etiquette, demonstration and application for each performer.' },
    ],
  },
  club_leader: {
    id: 'club_leader', label: 'Club leader',
    steps: [
      { selector: '.topbar', href: '/club', title: 'Your club area', body: 'Invite your members and see their statuses here.' },
      { selector: '.card', href: '/operations', title: 'Check in your team', body: 'Record weights and see out-of-class flags early.' },
      { selector: '.card', href: '/public', title: 'Follow your fighters', body: 'The public board shows live contests, next-up callouts and rankings.' },
    ],
  },
  fighter: {
    id: 'fighter', label: 'Participant',
    steps: [
      { selector: '.topbar', href: '/public', title: 'Your tournament board', body: 'See upcoming bouts, live scores and medal tables here.' },
      { selector: '.card', href: '/event', title: 'Check requirements', body: 'Read the kit list, schedule and course attendance rule before the day.' },
      { selector: '.card', href: '/settings', title: 'Your profile', body: 'Keep your stats, belt grade, club and preferences up to date.' },
    ],
  },
  spectator: {
    id: 'spectator', label: 'Spectator',
    steps: [
      { selector: '.topbar', href: '/public', title: 'Follow the event', body: 'Live scores, upcoming contests and results are all on the public dashboard.' },
      { selector: '.card', href: '/brackets', title: 'Watch progression', body: 'Brackets show who advances through every round.' },
      { selector: '.card', href: '/event', title: 'Event information', body: 'Times, amenities and venue details are available here.' },
    ],
  },
};

const KEY = 'j3-tour-progress';

type Progress = Record<string, number>;
export function progress(): Progress {
  try { return JSON.parse(localStorage.getItem(KEY) || '{}'); } catch { return {}; }
}
export function stepDone(role: Role, index: number) {
  const all = progress();
  all[role] = Math.max(all[role] ?? 0, index + 1);
  localStorage.setItem(KEY, JSON.stringify(all));
}
export function resetTour(role?: Role) {
  if (role) { const all = progress(); delete all[role]; localStorage.setItem(KEY, JSON.stringify(all)); }
  else localStorage.removeItem(KEY);
}
export function isTourDone(role: Role) {
  return (progress()[role] ?? 0) >= TOURS[role].steps.length;
}
