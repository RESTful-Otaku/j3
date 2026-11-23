import { derived, get, writable } from 'svelte/store';
import { api, type SimulationBout, type SimulationResult } from './api';

export type ReplayStanding = {
  fighterId: string;
  name: string;
  club: string;
  category: string;
  belt: string;
  clubColour: string;
  wins: number;
  losses: number;
  draws: number;
  points: number;
};

export type ReplayMedal = {
  id: string; fighterId: string; name: string; club: string; kind: string;
  category?: string; discipline?: string; eventType?: string;
};

export type ReplaySnapshot = {
  active: boolean;
  playing: boolean;
  loading: boolean;
  error: string;
  seed: number;
  fighters: number;
  index: number;
  total: number;
  progress: number;
  speed: number;
  bout: SimulationBout | null;
  bouts: SimulationBout[];
  completed: SimulationBout[];
  standings: ReplayStanding[];
  medals: ReplayMedal[];
  clubPoints: Array<[string, number]>;
  topMoves: Array<[string, number]>;
};

export type AppMode = 'live' | 'sim';

/** Global operating mode. Live is the real event; sim replays a deterministic tournament. */
export const appMode = writable<AppMode>('live');
export const replayActive = writable(false);
export const replayPlaying = writable(false);
export const replayLoading = writable(false);
export const replayError = writable('');
export const replayResult = writable<SimulationResult | null>(null);
export const replayIndex = writable(0);
export const replaySpeed = writable(450);
export const replaySeed = writable(42);
export const replayFighters = writable(50);

let timer: ReturnType<typeof setInterval> | null = null;

function clamp(value: number, max: number) {
  return Math.max(0, Math.min(max, Math.round(value)));
}

function redWon(bout: SimulationBout) {
  return bout.outcome.includes('red');
}

function winnerOf(bout: SimulationBout) {
  return redWon(bout) ? bout.redId : bout.blueId;
}

function loserOf(bout: SimulationBout) {
  return redWon(bout) ? bout.blueId : bout.redId;
}

function isFinal(bout: SimulationBout) {
  return bout.stage === 'knockout' && bout.round === 'Final';
}

function isSemi(bout: SimulationBout) {
  return bout.stage === 'knockout' && bout.round === 'Semi-final';
}

function clubColour(club: string) {
  const palette = ['#2563eb', '#b91c1c', '#047857', '#7c3aed', '#b45309'];
  const total = [...club].reduce((sum, char) => sum + char.charCodeAt(0), 0);
  return palette[total % palette.length];
}

function standingsAt(result: SimulationResult, completed: SimulationBout[]): ReplayStanding[] {
  const byId = new Map<string, ReplayStanding>();
  for (const row of result.standings) {
    byId.set(row.fighterId, {
      ...row, category: 'Demo replay', belt: '', clubColour: clubColour(row.club),
      wins: 0, losses: 0, draws: 0, points: 0,
    });
  }
  const rowFor = (id: string): ReplayStanding =>
    byId.get(id) ?? {
      fighterId: id, name: id, club: '', category: 'Demo replay', belt: '',
      clubColour: '#64748b', wins: 0, losses: 0, draws: 0, points: 0,
    };

  for (const bout of completed) {
    if (bout.outcome === 'draw') {
      for (const id of [bout.redId, bout.blueId]) {
        const row = rowFor(id);
        row.draws += 1;
        row.points += 1;
      }
      continue;
    }
    const winner = rowFor(winnerOf(bout));
    const loser = rowFor(loserOf(bout));
    winner.wins += 1;
    winner.points += 3;
    loser.losses += 1;
  }
  return [...byId.values()].sort((a, b) =>
    b.points - a.points || b.wins - a.wins || a.name.localeCompare(b.name));
}

function medalsAt(result: SimulationResult, completed: SimulationBout[], index: number, total: number): ReplayMedal[] {
  const rows = standingsAt(result, completed);
  const meta = new Map(rows.map((row) => [row.fighterId, row]));
  const medals: ReplayMedal[] = [];
  const push = (fighterId: string, kind: string, context?: { category?: string; discipline?: string; eventType?: string }) => {
    if (!fighterId) return;
    const existing = medals.find((medal) => medal.fighterId === fighterId && medal.kind === kind);
    if (existing) {
      if (context) {
        existing.category = context.category ?? existing.category;
        existing.discipline = context.discipline ?? existing.discipline;
        existing.eventType = context.eventType ?? existing.eventType;
      }
      return;
    }
    const row = meta.get(fighterId);
    medals.push({
      id: `${fighterId}-${kind}`, fighterId, name: row?.name ?? fighterId, club: row?.club ?? '', kind,
      category: context?.category ?? 'Open', discipline: context?.discipline ?? 'JJ Fighting System',
      eventType: context?.eventType ?? 'fight',
    });
  };

  const final = [...completed].reverse().find(isFinal);
  if (final) {
    push(winnerOf(final), 'gold');
    push(loserOf(final), 'silver');
  }
  for (const bout of completed.filter(isSemi)) push(loserOf(bout), 'bronze');

  // Once the bracket is complete, trust the server-resolved medal list (it also
  // handles no-rematch bronze rules and shared bronzes).
  if (total > 0 && index >= total - 1) {
    for (const medal of result.medals) push(medal.fighterId, medal.kind, medal);
  }
  return medals;
}

function clubPointsAt(result: SimulationResult, completed: SimulationBout[]): Array<[string, number]> {
  const clubOf = new Map(result.standings.map((row) => [row.fighterId, row.club]));
  const points = new Map<string, number>();
  for (const bout of completed) {
    if (bout.outcome === 'draw') continue;
    const club = clubOf.get(winnerOf(bout)) ?? 'Independent';
    points.set(club, (points.get(club) ?? 0) + 3);
  }
  return [...points.entries()].sort((a, b) => b[1] - a[1]);
}

function topMovesAt(completed: SimulationBout[]): Array<[string, number]> {
  const moves = new Map<string, number>();
  for (const bout of completed) moves.set(bout.method, (moves.get(bout.method) ?? 0) + 1);
  return [...moves.entries()].sort((a, b) => b[1] - a[1]);
}

const REPLAY_QUERY_KEYS = ['mode', 'replay', 'seed', 'fighters', 'speed'] as const;
let urlSyncStarted = false;

/** Query string that keeps a replay position when opening another view or a new tab. */
export function replayQuery(state: ReplaySnapshot = get(replaySnapshot)): string {
  const params = new URLSearchParams();
  params.set('mode', 'sim');
  params.set('replay', String(state.index));
  params.set('seed', String(state.seed));
  params.set('fighters', String(state.fighters));
  if (state.speed !== 450) params.set('speed', String(state.speed));
  return params.toString();
}

export const replaySnapshot = derived(
  [replayActive, replayPlaying, replayLoading, replayError, replaySeed, replayFighters,
    replayResult, replayIndex, replaySpeed],
  ([$active, $playing, $loading, $error, $seed, $fighters, $result, $index, $speed]): ReplaySnapshot => {
    const bouts = $result?.bouts ?? [];
    const total = bouts.length;
    const index = clamp($index, Math.max(0, total - 1));
    const completed = bouts.slice(0, total ? index + 1 : 0);
    return {
      active: $active,
      playing: $playing,
      loading: $loading,
      error: $error,
      seed: $seed,
      fighters: $fighters,
      index,
      total,
      progress: total ? ((index + 1) / total) * 100 : 0,
      speed: $speed,
      bout: bouts[index] ?? null,
      bouts,
      completed,
      standings: $result ? standingsAt($result, completed) : [],
      medals: $result ? medalsAt($result, completed, index, total) : [],
      clubPoints: $result ? clubPointsAt($result, completed) : [],
      topMoves: topMovesAt(completed),
    };
  },
);

function clearTimer() {
  if (timer) clearInterval(timer);
  timer = null;
}

export function pauseReplay() {
  clearTimer();
  replayPlaying.set(false);
}

export function playReplay() {
  const result = get(replayResult);
  if (!result || result.bouts.length === 0) return;
  clearTimer();
  if (get(replayIndex) >= result.bouts.length - 1) replayIndex.set(0);
  replayPlaying.set(true);
  timer = setInterval(() => {
    const current = get(replayResult);
    if (!current || get(replayIndex) >= current.bouts.length - 1) {
      pauseReplay();
      return;
    }
    replayIndex.update((value) => value + 1);
  }, get(replaySpeed));
}

export function toggleReplayPlayback() {
  if (get(replayPlaying)) pauseReplay();
  else playReplay();
}

export function seekReplay(value: number) {
  const total = get(replayResult)?.bouts.length ?? 0;
  replayIndex.set(clamp(value, Math.max(0, total - 1)));
  if (get(replayPlaying)) playReplay();
}

export function stepReplay(delta: number) {
  seekReplay(get(replayIndex) + delta);
}

export function restartReplay() {
  replayIndex.set(0);
  playReplay();
}

export function setReplaySpeed(ms: number) {
  replaySpeed.set(Math.max(80, Math.min(3000, ms)));
  if (get(replayPlaying)) playReplay();
}

export async function loadReplay(seed = get(replaySeed), fighters = get(replayFighters), autoplay = false) {
  replayLoading.set(true);
  replayError.set('');
  try {
    const result = await api.simulate(seed, fighters);
    replaySeed.set(seed);
    replayFighters.set(fighters);
    replayResult.set(result);
    replayIndex.set(0);
    replayActive.set(true);
    if (autoplay) playReplay();
    else pauseReplay();
  } catch (error) {
    replayError.set(error instanceof Error ? error.message : 'Could not load the demo tournament');
    replayActive.set(false);
    pauseReplay();
  } finally {
    replayLoading.set(false);
  }
}

export function startReplay() {
  if (get(replayResult)) {
    replayActive.set(true);
    restartReplay();
    return;
  }
  void loadReplay(get(replaySeed), get(replayFighters), true);
}

export function stopReplay() {
  pauseReplay();
  replayActive.set(false);
}

function persistAppMode(mode: AppMode) {
  appMode.set(mode);
  if (typeof window !== 'undefined') window.localStorage.setItem('j3-app-mode', mode);
}

/** Switch the whole app between the real event and the deterministic replay. */
export function setAppMode(mode: AppMode) {
  if (mode === 'live') {
    persistAppMode('live');
    stopReplay();
    return;
  }
  persistAppMode('sim');
  if (!get(replayResult)) {
    void loadReplay(get(replaySeed), get(replayFighters), true);
    return;
  }
  if (!get(replayActive)) startReplay();
  else if (!get(replayPlaying)) playReplay();
}

export function toggleAppMode() {
  setAppMode(get(appMode) === 'sim' ? 'live' : 'sim');
}

/**
 * Restore a replay from `?replay=N` and keep the URL in sync while it plays.
 * Called once from the root layout so every view can be inspected at the same
 * event, including when opened in a new tab.
 */
export function initReplayUrl(): void {
  if (typeof window === 'undefined' || urlSyncStarted) return;
  urlSyncStarted = true;

  const url = new URL(window.location.href);
  const modeParam = url.searchParams.get('mode');
  const storedMode = window.localStorage.getItem('j3-app-mode');
  const simRequested = modeParam === 'sim' || url.searchParams.has('replay') || storedMode === 'sim';
  persistAppMode(modeParam === 'live' ? 'live' : simRequested ? 'sim' : 'live');

  if (url.searchParams.has('replay')) {
    const seed = Number(url.searchParams.get('seed')) || get(replaySeed);
    const fighters = Number(url.searchParams.get('fighters')) || get(replayFighters);
    const speed = Number(url.searchParams.get('speed')) || get(replaySpeed);
    const index = Math.max(0, Number(url.searchParams.get('replay')) || 0);
    replaySeed.set(seed);
    replayFighters.set(fighters);
    setReplaySpeed(speed);
    void loadReplay(seed, fighters, false).then(() => seekReplay(index));
  } else if (simRequested) {
    void loadReplay(get(replaySeed), get(replayFighters), true);
  }

  replaySnapshot.subscribe(applyReplayUrl);
}

function applyReplayUrl(state: ReplaySnapshot) {
  if (typeof window === 'undefined') return;
  const next = new URL(window.location.href);
  if (!state.active) {
    let changed = false;
    for (const key of REPLAY_QUERY_KEYS) {
      if (next.searchParams.has(key)) { next.searchParams.delete(key); changed = true; }
    }
    if (changed) window.history.replaceState(window.history.state, '', next);
    return;
  }
  const query = replayQuery(state);
  if (next.search !== `?${query}`) {
    for (const key of REPLAY_QUERY_KEYS) next.searchParams.delete(key);
    for (const [key, value] of new URLSearchParams(query)) next.searchParams.set(key, value);
    window.history.replaceState(window.history.state, '', next);
  }
}

/** Re-apply the current playhead after client-side navigation (e.g. a paused replay). */
export function syncReplayUrl(): void {
  if (!urlSyncStarted) return;
  applyReplayUrl(get(replaySnapshot));
}
