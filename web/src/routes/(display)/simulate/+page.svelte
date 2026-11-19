<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { api, type SimulationResult, type StatsResult } from '../../../lib/api';
  import { notifySuccess } from '../../../lib/toast';

  let result: SimulationResult | null = null;
  let stats: StatsResult | null = null;
  let index = 0;
  let playing = false;
  let speed = 320;
  let seed = 42;
  let fighters = 50;
  let timer: ReturnType<typeof setInterval> | null = null;

  async function load() {
    [result, stats] = await Promise.all([api.simulate(seed, fighters), api.stats(seed, fighters)]);
    index = 0;
    pause();
    notifySuccess(`Simulated ${result.bouts.length} bouts with ${fighters} fighters`);
  }
  function play() {
    if (!result || playing) return;
    if (index >= result.bouts.length) index = 0;
    playing = true;
    timer = setInterval(() => {
      if (!result || index >= result.bouts.length - 1) { pause(); return; }
      index += 1;
    }, speed);
  }
  function pause() {
    playing = false;
    if (timer) clearInterval(timer);
    timer = null;
  }
  function step() {
    if (!result) return;
    pause();
    index = Math.min(result.bouts.length - 1, index + 1);
  }
  onMount(() => { void load(); });
  onDestroy(() => pause());

  function barWidth(value: number, max: number) { return `${Math.max(4, (value / (max || 1)) * 100)}%`; }
  function maxValue(record: Record<string, number>) { return Math.max(1, ...Object.values(record)); }
  function medalColour(kind: string) {
    return kind === 'gold' || kind === 'champion' ? '#fbbf24' : kind === 'bronze' ? '#d97706' : '#cbd5e1';
  }
  $: bout = result?.bouts[index];
  $: progress = result ? ((index + 1) / result.bouts.length) * 100 : 0;
  $: topMoves = stats ? Object.entries(stats.topMoves).sort((a, b) => b[1] - a[1]) : [];
  $: clubPoints = stats ? Object.entries(stats.clubPoints).sort((a, b) => b[1] - a[1]) : [];
  $: clubMedals = stats ? Object.entries(stats.clubMedals).sort((a, b) => b[1] - a[1]) : [];
</script>

<svelte:head><title>J3 · Tournament simulator</title></svelte:head>

<div class="display-shell">
  <header class="display-header">
    <div><div class="event-mark"><span class="crest">J3</span>Tournament simulator</div>
      <div class="meta">Deterministic full event · seed {seed} · {fighters} fighters · {result?.bouts.length ?? 0} bouts</div></div>
    <div style="display:flex; gap:8px; flex-wrap:wrap; align-items:center;">
      <input class="tab" type="number" style="color:var(--brand-ink); width:86px;" bind:value={seed} aria-label="Random seed" />
      <input class="tab" type="number" style="color:var(--brand-ink); width:86px;" bind:value={fighters} aria-label="Fighter count" />
      <button class="btn" on:click={() => void load()}>↻ Regenerate</button>
      <button class="btn" on:click={step}>Step →</button>
      <button class="btn primary" on:click={playing ? pause : play}>{playing ? '⏸ Pause' : '▶ Play full tournament'}</button>
    </div>
  </header>

  <div class="progress-track"><div class="progress-fill" style="width:{progress}%"></div></div>

  {#if bout}
    <div class="card slide-in" style="margin-top:18px; border-color:var(--brand-primary);">
      <div class="eyebrow">Bout {bout.sequence} · {bout.stage} · {bout.round}</div>
      <div class="replay">
        <div class="replay-side"><span class="portrait red">{bout.redId.slice(-2)}</span><strong>{bout.redId}</strong><span class="score">{bout.pointsRed}</span></div>
        <div class="vs">⚡ VS ⚡</div>
        <div class="replay-side"><span class="score">{bout.pointsBlue}</span><strong>{bout.blueId}</strong><span class="portrait blue">{bout.blueId.slice(-2)}</span></div>
      </div>
      <div class="capabilities" style="margin-top:14px;">
        <span class="cap on">{bout.outcome.replaceAll('_', ' ')}</span>
        <span class="cap on">{bout.method}</span>
      </div>
    </div>
  {/if}

  <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(min(100%, 300px), 1fr)); margin-top:24px;">
    <div class="card">
      <h2>Top moves & techniques</h2>
      {#each topMoves as [move, count] (move)}
        <div class="bar-row"><span>{move}</span><div class="bar"><div class="bar-fill" style="width:{barWidth(count, maxValue(stats?.topMoves ?? {}))}"></div></div><strong>{count}</strong></div>
      {/each}
    </div>
    <div class="card">
      <h2>Highest scoring clubs</h2>
      {#each clubPoints as [club, points] (club)}
        <div class="bar-row"><span>{club}</span><div class="bar"><div class="bar-fill alt" style="width:{barWidth(points, maxValue(stats?.clubPoints ?? {}))}"></div></div><strong>{points}</strong></div>
      {/each}
    </div>
    <div class="card">
      <h2>Medals by club</h2>
      {#each clubMedals as [club, count] (club)}
        <div class="bar-row"><span>{club}</span><div class="bar"><div class="bar-fill gold" style="width:{barWidth(count, maxValue(stats?.clubMedals ?? {}))}"></div></div><strong>{count}</strong></div>
      {/each}
    </div>
  </div>

  <div class="card" style="margin-top:24px;">
    <h2>Final standings</h2>
    <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(min(100%,250px),1fr)); margin-top:12px;">
      {#each (stats?.standings ?? []).slice(0, 10) as row, i (row.fighterId)}
        <div class="stat small"><div class="value" style="font-size:1.4rem;">#{i + 1} {row.name}</div><div class="label">{row.club} · {row.points} pts · {row.wins}W {row.draws}D {row.losses}L</div></div>
      {/each}
    </div>
  </div>

  {#if result && index >= result.bouts.length - 1}
    <div class="celebrate" style="margin-top:24px; text-align:center;">
      <div class="eyebrow">Tournament complete</div>
      <h1>🏆 Champions & medalists</h1>
      <div class="capabilities" style="justify-content:center; margin-top:14px;">
        {#each result.medals as medal (medal.fighterId + medal.kind)}
          <span class="medal" style="background:{medalColour(medal.kind)}">{medal.kind} · {medal.fighterId}</span>
        {/each}
      </div>
    </div>
  {/if}
</div>

<style>
  .progress-track { height:10px; border-radius:999px; background:var(--brand-surface-3); overflow:hidden; }
  .progress-fill { height:100%; background:linear-gradient(90deg,var(--brand-primary),var(--brand-accent)); transition:width .2s linear; }
  .replay { display:flex; align-items:center; justify-content:space-around; gap:18px; margin-top:18px; flex-wrap:wrap; }
  .replay-side { display:flex; align-items:center; gap:14px; font-size:1.3rem; }
  .portrait { width:76px; height:76px; border-radius:20px; display:grid; place-items:center; color:#fff; font-family:var(--font-display); font-weight:800; font-size:1.6rem; }
  .portrait.red { background:linear-gradient(145deg,#ef4444,#7f1d1d); }
  .portrait.blue { background:linear-gradient(145deg,#3b82f6,#1e3a8a); }
  .score { font-family:var(--font-display); font-size:3rem; font-weight:800; }
  .vs { color:var(--brand-accent); font-weight:800; font-size:1.4rem; }
  .bar-row { display:grid; grid-template-columns:110px 1fr 34px; align-items:center; gap:10px; margin-top:10px; font-size:.86rem; }
  .bar { height:12px; border-radius:999px; background:var(--brand-surface-3); overflow:hidden; }
  .bar-fill { height:100%; background:var(--brand-primary); border-radius:999px; transition:width .4s ease; }
  .bar-fill.alt { background:var(--win); }
  .bar-fill.gold { background:var(--brand-accent); }
  .medal { display:inline-flex; padding:8px 14px; border-radius:999px; color:#0f172a; font-weight:800; text-transform:capitalize; }
</style>
