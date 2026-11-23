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

export type ReplayMedal = { id: string; fighterId: string; name: string; club: string; kind: string };

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
  completed: SimulationBout[];
  standings: ReplayStanding[];
  medals: ReplayMedal[];
  clubPoints: Array<[string, number]>;
  topMoves: Array<[string, number]>;
};

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
  const push = (fighterId: string, kind: string) => {
    if (!fighterId || medals.some((medal) => medal.fighterId === fighterId && medal.kind === kind)) return;
    const row = meta.get(fighterId);
    medals.push({ id: `${fighterId}-${kind}`, fighterId, name: row?.name ?? fighterId, club: row?.club ?? '', kind });
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
    for (const medal of result.medals) push(medal.fighterId, medal.kind);
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
