<script lang="ts">
  import { onMount } from 'svelte';
  import {
    loadReplay, replaySnapshot, restartReplay, seekReplay, setReplaySpeed,
    stepReplay, toggleReplayPlayback,
  } from '../../../lib/demo-replay';
  import { notifySuccess } from '../../../lib/toast';

  let seed = 42;
  let fighters = 50;

  function load() {
    void loadReplay(seed, fighters).then(() => notifySuccess(`Simulated ${$replaySnapshot.total} bouts with ${fighters} fighters`));
  }
  onMount(load);

  function barWidth(value: number, max: number) { return `${Math.max(4, (value / (max || 1)) * 100)}%`; }
  function maxValue(entries: Array<[string, number]>) { return Math.max(1, ...entries.map(([, value]) => value)); }
  function medalColour(kind: string) {
    return kind === 'gold' || kind === 'champion' ? '#fbbf24' : kind === 'bronze' ? '#d97706' : '#cbd5e1';
  }
  $: state = $replaySnapshot;
  $: topMoves = state.topMoves;
  $: clubPoints = state.clubPoints;
  $: clubMedals = state.medals.reduce<Array<[string, number]>>((rows, medal) => {
    const entry = rows.find(([club]) => club === (medal.club || 'Independent'));
    if (entry) entry[1] += 1;
    else rows.push([medal.club || 'Independent', 1]);
    return rows;
  }, []).sort((a, b) => b[1] - a[1]);
</script>

<svelte:head><title>J3 · Tournament simulator</title></svelte:head>

<div class="display-shell">
  <header class="display-header">
    <div><div class="event-mark"><span class="crest">J3</span>Tournament simulator</div>
      <div class="meta">Deterministic full event · seed {state.seed} · {state.fighters} fighters · {state.total} bouts</div></div>
    <div style="display:flex; gap:8px; flex-wrap:wrap; align-items:center;">
      <input class="tab" type="number" style="color:var(--brand-ink); width:86px;" bind:value={seed} aria-label="Random seed" />
      <input class="tab" type="number" style="color:var(--brand-ink); width:86px;" bind:value={fighters} aria-label="Fighter count" />
      <button class="btn" on:click={load}>↻ Regenerate</button>
      <button class="btn" on:click={() => restartReplay()}>↻ Restart</button>
      <button class="btn" on:click={() => stepReplay(-1)}>← Back</button>
      <button class="btn" on:click={() => stepReplay(1)}>Step →</button>
      <select class="tab" style="color:var(--brand-ink);" value={state.speed}
        on:change={(event) => setReplaySpeed(Number((event.target as HTMLSelectElement).value))} aria-label="Replay speed">
        <option value={120}>4×</option><option value={450}>1×</option><option value={900}>½×</option>
      </select>
      <button class="btn primary" on:click={toggleReplayPlayback}>{state.playing ? '⏸ Pause' : '▶ Play full tournament'}</button>
    </div>
  </header>

  <div class="progress-track"><div class="progress-fill" style="width:{state.progress}%"></div></div>
  <input class="scrub" type="range" min="0" max={Math.max(0, state.total - 1)} value={state.index}
    on:input={(event) => seekReplay(Number((event.target as HTMLInputElement).value))} aria-label="Scrub tournament replay" />

  {#if state.bout}
    <div class="card slide-in" style="margin-top:18px; border-color:var(--brand-primary);">
      <div class="eyebrow">Bout {state.bout.sequence} · {state.bout.stage} · {state.bout.round}</div>
      <div class="replay">
        <div class="replay-side"><span class="portrait red">{state.bout.redId.slice(-2)}</span><strong>{state.bout.redId}</strong><span class="score">{state.bout.pointsRed}</span></div>
        <div class="vs">⚡ VS ⚡</div>
        <div class="replay-side"><span class="score">{state.bout.pointsBlue}</span><strong>{state.bout.blueId}</strong><span class="portrait blue">{state.bout.blueId.slice(-2)}</span></div>
      </div>
      <div class="capabilities" style="margin-top:14px;">
        <span class="cap on">{state.bout.outcome.replaceAll('_', ' ')}</span>
        <span class="cap on">{state.bout.method}</span>
      </div>
    </div>
  {/if}

  <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(min(100%, 300px), 1fr)); margin-top:24px;">
    <div class="card">
      <h2>Top moves & techniques</h2>
      {#each topMoves as [move, count] (move)}
        <div class="bar-row"><span>{move}</span><div class="bar"><div class="bar-fill" style="width:{barWidth(count, maxValue(topMoves))}"></div></div><strong>{count}</strong></div>
      {/each}
    </div>
    <div class="card">
      <h2>Highest scoring clubs</h2>
      {#each clubPoints as [club, points] (club)}
        <div class="bar-row"><span>{club}</span><div class="bar"><div class="bar-fill alt" style="width:{barWidth(points, maxValue(clubPoints))}"></div></div><strong>{points}</strong></div>
      {/each}
    </div>
    <div class="card">
      <h2>Medals by club</h2>
      {#each clubMedals as [club, count] (club)}
        <div class="bar-row"><span>{club}</span><div class="bar"><div class="bar-fill gold" style="width:{barWidth(count, maxValue(clubMedals))}"></div></div><strong>{count}</strong></div>
      {/each}
    </div>
  </div>

  <div class="card" style="margin-top:24px;">
    <h2>Standings at this point</h2>
    <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(min(100%,250px),1fr)); margin-top:12px;">
      {#each state.standings.slice(0, 10) as row, i (row.fighterId)}
        <div class="stat small"><div class="value" style="font-size:1.4rem;">#{i + 1} {row.name}</div><div class="label">{row.club} · {row.points} pts · {row.wins}W {row.draws}D {row.losses}L</div></div>
      {/each}
    </div>
  </div>

  {#if state.active && state.total > 0 && state.index >= state.total - 1}
    <div class="celebrate" style="margin-top:24px; text-align:center;">
      <div class="eyebrow">Tournament complete</div>
      <h1>🏆 Champions & medalists</h1>
      <div class="capabilities" style="justify-content:center; margin-top:14px;">
        {#each state.medals as medal (medal.id)}
          <span class="medal" style="background:{medalColour(medal.kind)}">{medal.kind} · {medal.name}</span>
        {/each}
      </div>
    </div>
  {/if}
</div>

<style>
  .progress-track { height:10px; border-radius:999px; background:var(--brand-surface-3); overflow:hidden; }
  .progress-fill { height:100%; background:linear-gradient(90deg,var(--brand-primary),var(--brand-accent)); transition:width .2s linear; }
  .scrub { width:100%; margin-top:8px; accent-color:var(--brand-accent); }
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
