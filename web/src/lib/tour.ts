import { hasGuideAction } from './guide';

export type Role = 'runner' | 'official' | 'club_leader' | 'fighter' | 'spectator';

export type GuideCheck =
  | { kind: 'route'; path: string }
  | { kind: 'value'; selector: string; minLength?: number }
  | { kind: 'action'; key: string }
  | { kind: 'visible'; selector: string };

export type GuideStep = {
  id: string;
  title: string;
  body: string;
  hint: string;
  href: string;
  target: string;
  mobileTarget?: string;
  /** For the setup wizard: which wizard tab must be open before this step. */
  wizardStep?: number;
  check: GuideCheck;
};

export type Guide = {
  id: string;
  label: string;
  shortLabel: string;
  icon: string;
  description: string;
  steps: GuideStep[];
};

const runner: Guide = {
  id: 'runner', label: 'Tournament runner', shortLabel: 'Runner', icon: '👑',
  description: 'Create a small event, generate its draw, check in a fighter, score and publish a result.',
  steps: [
    { id: 'runner-welcome', title: 'Start at your dashboard', body: 'This is the control room. Every guide step below uses the real app — nothing is simulated.', hint: 'You are already in the right place. Read the cards, then press Next.', href: '/', target: '[data-guide="nav-help-compact"]', mobileTarget: '.mobile-bar', check: { kind: 'route', path: '/' } },
    { id: 'runner-name', title: 'Name your tournament', body: 'Open the setup wizard and give your event a name of at least three characters. Use something simple like “Demo Challenge”.', hint: 'Type in the Tournament name field. Next unlocks when the name is saved.', href: '/setup', target: '[data-guide="setup-name"]', wizardStep: 0, check: { kind: 'value', selector: '[data-guide="setup-name"] input', minLength: 3 } },
    { id: 'runner-theme', title: 'Choose a look', body: 'Pick a theme for the boards and displays. You can change branding later, but choose one now.', hint: 'Tap Classic, Ocean or Street Fighter in the Theme row.', href: '/setup', target: '[data-guide="setup-theme"]', wizardStep: 0, check: { kind: 'action', key: 'setup-theme-picked' } },
    { id: 'runner-disciplines', title: 'Add a discipline', body: 'Add one discipline so the event has something to run. The defaults give you sensible timings and points.', hint: 'Press “Add discipline” in step 2, then fill in its name.', href: '/setup', target: '[data-guide="setup-disciplines"]', wizardStep: 1, check: { kind: 'action', key: 'setup-discipline-added' } },
    { id: 'runner-invites', title: 'Invite your people', body: 'Add at least one invitation in the People step. One line is enough: an email, role and club.', hint: 'Use the step buttons at the top to jump to People, then press “Add invitations”.', href: '/setup', target: '[data-guide="setup-invites"]', wizardStep: 3, check: { kind: 'action', key: 'setup-invites-added' } },
    { id: 'runner-launch', title: 'Review and launch', body: 'Jump to Review & launch. If the readiness check is green, launch the tournament. This creates the event and auto-generates the draw and officials.', hint: 'Fix any listed issues first — the guide will not let you skip them.', href: '/setup', target: '[data-guide="setup-launch"]', wizardStep: 4, check: { kind: 'action', key: 'tournament-created' } },
    { id: 'runner-pools', title: 'Inspect the generated draw', body: 'Fight pools and brackets were generated for you. Check that competitors are spread fairly across pools.', hint: 'This step completes when the fight-pool board loads.', href: '/pools', target: '[data-guide="pools-header"]', check: { kind: 'visible', selector: '[data-guide="pools-header"]' } },
    { id: 'runner-checkin', title: 'Check in and weigh in a fighter', body: 'Enter an official weight for one competitor. Out-of-class athletes are flagged automatically.', hint: 'Change any weight field in Check-in & weigh-in.', href: '/operations', target: '[data-guide="operations-weigh-input"]', check: { kind: 'action', key: 'check-in' } },
    { id: 'runner-mat-select', title: 'Pick a contest', body: 'The mat console only shows controls after a contest is selected. Choose a bout from the queue first.', hint: 'Tap any “Red vs Blue” contest button in the contest queue.', href: '/mat', target: '[data-guide="mat-contest-select"]', check: { kind: 'action', key: 'fight-selected' } },
    { id: 'runner-mat', title: 'Run the clock', body: 'Start the authoritative clock. Pause when the referee stops the bout, then resume from the same console.', hint: 'Use Start, Pause and Resume on the mat console.', href: '/mat', target: '[data-guide="mat-start"]', check: { kind: 'action', key: 'fight-started' } },
    { id: 'runner-score', title: 'Award points', body: 'Add a score to either corner. Every entry is committed immediately and appears on the scoreboard and audit trail.', hint: 'Use the Red +N or Blue +N buttons on the mat console.', href: '/mat', target: '[data-guide="mat-scores"]', check: { kind: 'action', key: 'score-added' } },
    { id: 'runner-submit', title: 'Submit the result', body: 'Choose the real winner or draw, then confirm. This locks the contest and publishes it everywhere.', hint: 'Pick the result in the Result dropdown, press Submit result and confirm.', href: '/mat', target: '[data-guide="mat-submit"]', check: { kind: 'action', key: 'result-submitted' } },
    { id: 'runner-control', title: 'Put screens where people need them', body: 'Display control opens every board in its own tab for TVs, projectors and judge tablets. Keep this page on the control-room machine.', hint: 'The step completes when Display control loads.', href: '/control', target: '[data-guide="control-header"]', check: { kind: 'route', path: '/control' } },
    { id: 'runner-public', title: 'Show the venue board', body: 'The public board is what spectators see: live contests, next-up callouts, clock and results.', hint: 'Open it here, then try it on a second screen.', href: '/public', target: '[data-guide="public-header"]', check: { kind: 'route', path: '/public' } },
    { id: 'runner-audit', title: 'Verify the audit trail', body: 'Every action you took is recorded with actor, time and effect. Review it before you trust a result.', hint: 'You have finished the runner guide when this page loads.', href: '/audit', target: '[data-guide="audit-header"]', check: { kind: 'route', path: '/audit' } },
  ],
};

const official: Guide = {
  id: 'official', label: 'Official / moderator', shortLabel: 'Officials', icon: '🎯',
  description: 'Pick a contest, score with confirmation, control the clock and publish a result.',
  steps: [
    { id: 'official-contest', title: 'Choose a contest', body: 'Select the bout you are officiating from the contest list.', hint: 'Change the “Select contest” dropdown.', href: '/official', target: '[data-guide="official-contest-select"]', check: { kind: 'action', key: 'official-contest-selected' } },
    { id: 'official-corner', title: 'Choose the active corner', body: 'Points and penalties always apply to the corner you select. Check the name on the button before scoring.', hint: 'Tap the RED or BLUE corner button.', href: '/official', target: '[data-guide="official-corner-red"]', check: { kind: 'action', key: 'official-corner' } },
    { id: 'official-search', title: 'Search the technique catalogue', body: 'Type an English name or romaji such as “seoi nage”, pick a family chip, then select a technique. Point values come from the event rules.', hint: 'Search, choose a result and review the points dropdown.', href: '/official', target: '[data-guide="official-tech-button"]', check: { kind: 'visible', selector: '[data-guide="official-tech-button"]' } },
    { id: 'official-technique', title: 'Log a technique', body: 'Confirm the technique for the active corner. Its points are resolved from the configured rule set, not guessed by the browser.', hint: 'Choose any technique and press “Log for RED”, then “Confirm & log”.', href: '/official', target: '[data-guide="official-tech-button"]', check: { kind: 'action', key: 'official-entry' } },
    { id: 'official-cards', title: 'Attribute a yellow or red card', body: 'Cards are recorded against the active corner with the correct shido or hansoku wording and a configurable penalty value.', hint: 'Press Yellow / shido or Red / hansoku, then confirm.', href: '/official', target: '[data-guide="official-cards"]', check: { kind: 'action', key: 'official-card' } },
    { id: 'official-clock', title: 'Control the authoritative clock', body: 'Start, pause and resume the bout clock here. The same timer is shown on every board and paused at the referee’s command.', hint: 'Use Start, Pause and Resume in the contest card.', href: '/official', target: '[data-guide="official-clock"]', check: { kind: 'action', key: 'timer-paused' } },
    { id: 'official-technical', title: 'Call a technical result', body: 'Record an injury, forfeit, no-show or disqualification with a reason and optional detail. The bout locks and every display agrees.', hint: 'Choose a reason in “Technical win / loss”, then press “Call technical result” and confirm.', href: '/official', target: '[data-guide="official-technical"]', check: { kind: 'action', key: 'official-result' } },
    { id: 'official-result', title: 'Publish an outcome', body: 'Finish the bout by recording a win, loss or draw. If scores are level the console prompts you to confirm the draw or call a technical decision.', hint: 'Use Red wins, Blue wins or Draw in “Confirm final decision”.', href: '/official', target: '[data-guide="official-result-win-red"]', check: { kind: 'action', key: 'official-result' } },
    { id: 'official-kata', title: 'Score a kata panel', body: 'Switch to the Kata tab and enter a panel score. The total is calculated from committed judge scores.', hint: 'Open Kata panel, then change any score field and confirm.', href: '/official', target: '[data-guide="official-tab-kata"]', check: { kind: 'action', key: 'kata-score' } },
    { id: 'official-tables', title: 'Check the published tables', body: 'Verify that your entries appear in the live standings and kata table.', hint: 'This step completes when the tables page loads.', href: '/tables', target: '[data-guide="tables-header"]', check: { kind: 'route', path: '/tables' } },
  ],
};

const participant: Guide = {
  id: 'participant', label: 'Participant / spectator', shortLabel: 'Participants', icon: '🥋',
  description: 'Find your contest, check the requirements, keep your profile current and follow the results.',
  steps: [
    { id: 'participant-board', title: 'Read the public board', body: 'This shows live contests, the clock and who is up next. Check it regularly on the day.', hint: 'The step completes when the public board loads.', href: '/public', target: '[data-guide="public-header"]', check: { kind: 'route', path: '/public' } },
    { id: 'participant-requirements', title: 'Check kit and rules', body: 'Read the required kit, event times, venue details and course rules before you travel.', hint: 'Open Event config and browse the tabs.', href: '/event', target: '[data-guide="event-header"]', check: { kind: 'route', path: '/event' } },
    { id: 'participant-profile', title: 'Update your profile', body: 'Keep your name, contact details, height, weight, rank and club correct so eligibility is accurate.', hint: 'Change any profile field, then press Save settings.', href: '/settings', target: '[data-guide="settings-profile"]', check: { kind: 'action', key: 'settings-saved' } },
    { id: 'participant-pools', title: 'Find your fight pool', body: 'Fight pools show your group and the contests scheduled within it.', hint: 'The step completes when Fight pools loads.', href: '/pools', target: '[data-guide="pools-header"]', check: { kind: 'route', path: '/pools' } },
    { id: 'participant-brackets', title: 'Follow progression', body: 'Brackets show who advances through each round, including third-place and bronze decisions.', hint: 'The step completes when Brackets loads.', href: '/brackets', target: '[data-guide="brackets-header"]', check: { kind: 'route', path: '/brackets' } },
    { id: 'participant-tables', title: 'See results and medals', body: 'Standings, club points and medals update live from confirmed results.', hint: 'The step completes when Tables & rankings loads.', href: '/tables', target: '[data-guide="tables-header"]', check: { kind: 'route', path: '/tables' } },
  ],
};

const clubLeader: Guide = {
  id: 'club_leader', label: 'Club leader', shortLabel: 'Club leaders', icon: '🏠',
  description: 'Invite members, check in your team and follow their results from the club area.',
  steps: [
    { id: 'club-home', title: 'Open your club area', body: 'Your club page lists members you can manage and links to the public boards.', hint: 'The step completes when My club loads.', href: '/club', target: '[data-guide="club-members"]', check: { kind: 'route', path: '/club' } },
    { id: 'club-invite', title: 'Invite a member', body: 'Enter a name and email, then invite them under your club affiliation.', hint: 'Fill both fields and press “Send club invite”.', href: '/club', target: '[data-guide="club-invite-submit"]', check: { kind: 'action', key: 'club-invite' } },
    { id: 'club-checkin', title: 'Check in your team', body: 'Record an official weight for one entrant so out-of-class flags are caught early.', hint: 'Change any weight field in Check-in & weigh-in.', href: '/operations', target: '[data-guide="operations-weigh-input"]', check: { kind: 'action', key: 'check-in' } },
    { id: 'club-public', title: 'Follow your fighters', body: 'The public board shows live contests and upcoming callouts for your club.', hint: 'The step completes when the public board loads.', href: '/public', target: '[data-guide="public-header"]', check: { kind: 'route', path: '/public' } },
    { id: 'club-tables', title: 'Check club points', body: 'See how your club is doing in the medal and team tables.', hint: 'The step completes when Tables & rankings loads.', href: '/tables', target: '[data-guide="tables-header"]', check: { kind: 'route', path: '/tables' } },
  ],
};

export const GUIDES: Record<Role, Guide> = {
  runner,
  official,
  club_leader: clubLeader,
  fighter: participant,
  spectator: participant,
};

/** Unique guides shown on the progress page. */
export const GUIDE_LIST: Guide[] = [runner, official, clubLeader, participant];

export function guideForRole(role: Role): Guide {
  return GUIDES[role] ?? runner;
}

const PROGRESS_KEY = 'j3-guide-progress-v1';
const LEGACY_KEY = 'j3-tour-progress';
const LEGACY_SEEN_KEY = 'j3-tour-seen';

export type GuideProgress = {
  completed: string[];
  skipped: boolean;
  updatedAt: number;
};

const emptyProgress = (): GuideProgress => ({ completed: [], skipped: false, updatedAt: 0 });

export function guideProgress(role: Role): GuideProgress {
  if (typeof localStorage === 'undefined') return emptyProgress();
  try {
    const all = JSON.parse(localStorage.getItem(PROGRESS_KEY) || '{}') as Record<string, GuideProgress>;
    const saved = all[role];
    if (!saved || !Array.isArray(saved.completed)) return emptyProgress();
    return { completed: saved.completed, skipped: Boolean(saved.skipped), updatedAt: Number(saved.updatedAt) || 0 };
  } catch {
    return emptyProgress();
  }
}

function saveProgress(role: Role, value: GuideProgress) {
  if (typeof localStorage === 'undefined') return;
  try {
    const all = JSON.parse(localStorage.getItem(PROGRESS_KEY) || '{}') as Record<string, GuideProgress>;
    all[role] = { ...value, updatedAt: Date.now() };
    localStorage.setItem(PROGRESS_KEY, JSON.stringify(all));
  } catch {
    localStorage.setItem(PROGRESS_KEY, JSON.stringify({ [role]: { ...value, updatedAt: Date.now() } }));
  }
  // Keep the old counter in sync so existing QA helpers still disable auto-start.
  try {
    const legacy = JSON.parse(localStorage.getItem(LEGACY_KEY) || '{}') as Record<string, number>;
    legacy[role] = Math.max(legacy[role] ?? 0, value.completed.length);
    localStorage.setItem(LEGACY_KEY, JSON.stringify(legacy));
  } catch {
    localStorage.setItem(LEGACY_KEY, JSON.stringify({ [role]: value.completed.length }));
  }
}

export function markStepDone(role: Role, stepId: string) {
  const guide = guideForRole(role);
  if (!guide.steps.some((step) => step.id === stepId)) return;
  const current = guideProgress(role);
  if (!current.completed.includes(stepId)) {
    saveProgress(role, { ...current, completed: [...current.completed, stepId] });
  }
}

export function setGuideSkipped(role: Role) {
  const guide = guideForRole(role);
  // Skipping is an explicit "I do not need this guide" decision: mark every
  // step complete so it reads as dismissed rather than half-finished, and stop
  // auto-starting it. The user can always relaunch it from Help or /guides.
  saveProgress(role, {
    completed: guide.steps.map((step) => step.id),
    skipped: true,
    updatedAt: Date.now(),
  });
}

export function resetGuide(role?: Role) {
  if (typeof localStorage === 'undefined') return;
  if (role) {
    try {
      const all = JSON.parse(localStorage.getItem(PROGRESS_KEY) || '{}') as Record<string, GuideProgress>;
      delete all[role];
      localStorage.setItem(PROGRESS_KEY, JSON.stringify(all));
      const legacy = JSON.parse(localStorage.getItem(LEGACY_KEY) || '{}') as Record<string, number>;
      delete legacy[role];
      localStorage.setItem(LEGACY_KEY, JSON.stringify(legacy));
    } catch { /* clearing is best-effort */ }
    return;
  }
  localStorage.removeItem(PROGRESS_KEY);
  localStorage.removeItem(LEGACY_KEY);
  localStorage.removeItem(LEGACY_SEEN_KEY);
}

export function guideStats(role: Role) {
  const guide = guideForRole(role);
  const progress = guideProgress(role);
  const completed = guide.steps.filter((step) => progress.completed.includes(step.id)).length;
  const nextIndex = guide.steps.findIndex((step) => !progress.completed.includes(step.id));
  return {
    guide,
    progress,
    completed,
    total: guide.steps.length,
    percent: Math.round((completed / guide.steps.length) * 100),
    nextIndex: nextIndex === -1 ? 0 : nextIndex,
    done: completed === guide.steps.length,
  };
}

export function stepIsComplete(role: Role, step: GuideStep) {
  if (guideProgress(role).completed.includes(step.id)) return true;
  switch (step.check.kind) {
    case 'action':
      return hasGuideAction(step.check.key);
    case 'route':
      return typeof window !== 'undefined' && window.location.pathname === step.check.path;
    case 'value': {
      if (typeof document === 'undefined') return false;
      const element = document.querySelector(step.check.selector) as HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement | null;
      return Boolean(element && (element.value || '').trim().length >= (step.check.minLength ?? 1));
    }
    case 'visible': {
      if (typeof document === 'undefined') return false;
      const element = document.querySelector(step.check.selector) as HTMLElement | null;
      if (!element) return false;
      const rect = element.getBoundingClientRect();
      return rect.width > 0 && rect.height > 0;
    }
  }
}

/** Old helpers kept so existing callers and QA scripts continue to work. */
export function progress(): Record<string, number> {
  if (typeof localStorage === 'undefined') return {};
  try { return JSON.parse(localStorage.getItem(LEGACY_KEY) || '{}'); } catch { return {}; }
}

export function stepDone(role: Role, index: number) {
  if (index < 0) return;
  const step = guideForRole(role).steps[index];
  if (step) markStepDone(role, step.id);
  else saveProgress(role, guideProgress(role));
}

export function isTourDone(role: Role) {
  return guideStats(role).done;
}

export function hasSeenTour() {
  return typeof localStorage !== 'undefined' && localStorage.getItem(LEGACY_SEEN_KEY) === '1';
}

/** True when a guide has been skipped, completed or suppressed by QA helpers. */
export function isGuideDismissed(role: Role) {
  return guideProgress(role).skipped || isTourDone(role) || hasSeenTour();
}
