<script lang="ts">
  import { onMount } from 'svelte';
  import { get } from 'svelte/store';
  import {
    appMode, loadReplay, replayQuery, replayResult, replaySnapshot, restartReplay, seekReplay,
    setAppMode, setReplaySpeed, stepReplay, toggleReplayPlayback,
  } from '../../../lib/demo-replay';
  import { notifySuccess } from '../../../lib/toast';

  let seed = 42;
  let fighters = 50;

  function load() {
    void loadReplay(seed, fighters).then(() => notifySuccess(`Simulated ${$replaySnapshot.total} bouts with ${fighters} fighters`));
  }
  onMount(() => {
    if (get(appMode) !== 'sim') setAppMode('sim');
    else if (!get(replayResult)) load();
  });

  function barWidth(value: number, max: number) { return `${Math.max(4, (value / (max || 1)) * 100)}%`; }
  function maxValue(entries: Array<[string, number]>) { return Math.max(1, ...entries.map(([, value]) => value)); }
  function medalColour(kind: string) {
    return kind === 'gold' || kind === 'champion' ? '#fbbf24' : kind === 'bronze' ? '#d97706' : '#cbd5e1';
  }
  function stageGroups(completed: typeof $replaySnapshot.bouts) {
    const groups = new Map<string, typeof completed>();
    for (const bout of completed) {
      const label = bout.stage === 'pool' ? `Pool · ${bout.round}` : bout.round;
      groups.set(label, [...(groups.get(label) ?? []), bout]);
    }
    return [...groups.entries()];
  }
  function eventLabel(bout: { stage: string; round: string; sequence: number }) {
    return bout.stage === 'pool' ? `Pool ${bout.round.replace(/^Pool\s*/, '')}` : bout.round;
  }
  $: state = $replaySnapshot;
  $: topMoves = state.topMoves;
  $: clubPoints = state.clubPoints;
  $: names = new Map(state.standings.map((row) => [row.fighterId, row.name]));
  $: fighterName = (id: string) => names.get(id) ?? id;
  $: groups = stageGroups(state.completed);
  $: clubMedals = state.medals.reduce<Array<[string, number]>>((rows, medal) => {
    const entry = rows.find(([club]) => club === (medal.club || 'Independent'));
    if (entry) entry[1] += 1;
    else rows.push([medal.club || 'Independent', 1]);
    return rows;
  }, []).sort((a, b) => b[1] - a[1]);
</script>

<svelte:head><title>J3 · Tournament replay console</title></svelte:head>

<div class="display-shell">
  <header class="display-header">
    <div>
      <div class="event-mark"><span class="crest">J3</span>Tournament replay</div>
      <div class="meta">Bout {state.total ? state.index + 1 : 0} of {state.total} · deterministic seed {state.seed} · {state.fighters} fighters</div>
    </div>
    <div class="controls">
      <input class="tab" type="number" style="color:var(--brand-ink); width:80px;" bind:value={seed} aria-label="Random seed" />
      <input class="tab" type="number" style="color:var(--brand-ink); width:84px;" bind:value={fighters} aria-label="Fighter count" />
      <button class="btn" on:click={load}>↻ Regenerate</button>
      <select class="tab" style="color:var(--brand-ink);" value={state.speed}
        on:change={(event) => setReplaySpeed(Number((event.target as HTMLSelectElement).value))} aria-label="Replay speed">
        <option value={80}>8×</option><option value={160}>4×</option><option value={450}>1×</option><option value={900}>½×</option>
      </select>
      <button class="btn" on:click={() => restartReplay()}>↻ Restart</button>
      <button class="btn" on:click={() => stepReplay(-1)}>← Back</button>
      <button class="btn" on:click={() => stepReplay(1)}>Step →</button>
      <button class="btn" on:click={() => setAppMode('live')}>■ Stop & go live</button>
      <button class="btn primary" on:click={toggleReplayPlayback}>{state.playing ? '⏸ Pause' : '▶ Play full tournament'}</button>
    </div>
  </header>

  <div class="playhead">
    <div class="progress-track"><div class="progress-fill" style="width:{state.progress}%"></div></div>
    <input class="scrub" type="range" min="0" max={Math.max(0, state.total - 1)} value={state.index}
      on:input={(event) => seekReplay(Number((event.target as HTMLInputElement).value))} aria-label="Scrub tournament replay" />
    <div class="playhead-meta">
      <span>{state.completed.length} results recorded</span>
      <span>{state.total - state.completed.length} upcoming</span>
      <span>{state.speed === 80 ? '8× speed' : state.speed === 160 ? '4× speed' : state.speed === 450 ? '1× speed' : '½× speed'}</span>
    </div>
  </div>

  {#if state.bout}
    <div class="card now slide-in" style="border-color:var(--brand-primary);">
      <div class="eyebrow">Bout {state.bout.sequence} · {state.bout.stage} · {state.bout.round}</div>
      <div class="replay">
        <div class="replay-side"><span class="portrait red">{state.bout.redId.slice(-2)}</span><strong>{fighterName(state.bout.redId)}</strong><span class="score">{state.bout.pointsRed}</span></div>
        <div class="vs">⚡ VS ⚡</div>
        <div class="replay-side"><span class="score">{state.bout.pointsBlue}</span><strong>{fighterName(state.bout.blueId)}</strong><span class="portrait blue">{state.bout.blueId.slice(-2)}</span></div>
      </div>
      <div class="capabilities">
        <span class="cap on">{state.bout.outcome.replaceAll('_', ' ')}</span>
        <span class="cap on">{state.bout.method}</span>
      </div>
      {#if state.bout.techniques?.length}
        <div class="techniques"><span class="eyebrow">Scoring actions</span>{state.bout.techniques.slice(-8).join(' · ')}</div>
      {/if}
      <div class="inspect-links">
        <span class="eyebrow">Inspect this moment in</span>
        <a href="/mat?{replayQuery(state)}">Mat board</a>
        <a href="/public?{replayQuery(state)}">Public board</a>
        <a href="/tables?{replayQuery(state)}">Tables</a>
        <a href="/brackets?{replayQuery(state)}">Brackets</a>
        <a href="/ceremony?{replayQuery(state)}">Ceremony</a>
      </div>
    </div>
  {/if}

  <div class="card timeline" style="margin-top:18px;">
    <div class="section-head"><h2>Event-by-event timeline</h2><span class="eyebrow">Click any bout to jump</span></div>
    <div class="event-strip">
      {#each state.bouts as bout, i (bout.id)}
        <button class:current={i === state.index} class:past={i < state.index} class="event-chip"
          on:click={() => seekReplay(i)} aria-label={`Jump to bout ${bout.sequence}, ${eventLabel(bout)}`}>
          <span>{bout.sequence}</span><small>{eventLabel(bout)}</small>
        </button>
      {/each}
    </div>
  </div>

  <div class="grid views" style="margin-top:18px;">
    <div class="card view-card">
      <div class="section-head"><h2>Mat at this point</h2><a href="/mat?{replayQuery(state)}">Open →</a></div>
      {#if state.bout}
        <div class="mini-score"><span>{fighterName(state.bout.redId)}</span><strong>{state.bout.pointsRed}–{state.bout.pointsBlue}</strong><span>{fighterName(state.bout.blueId)}</span></div>
        <p class="hint">{state.bout.outcome.replaceAll('_', ' ')} · {state.bout.method}</p>
      {:else}
        <p class="hint">Waiting for the first bout.</p>
      {/if}
    </div>
    <div class="card view-card">
      <div class="section-head"><h2>Table at this point</h2><a href="/tables?{replayQuery(state)}">Open →</a></div>
      {#each state.standings.slice(0, 5) as row, i (row.fighterId)}
        <div class="rank-row"><span>#{i + 1}</span><strong>{row.name}</strong><small>{row.wins}W · {row.points} pts</small></div>
      {/each}
    </div>
    <div class="card view-card">
      <div class="section-head"><h2>Bracket at this point</h2><a href="/brackets?{replayQuery(state)}">Open →</a></div>
      {#each groups.slice(-4) as [label, bouts] (label)}
        <div class="rank-row"><span>▸</span><strong>{label}</strong><small>{bouts.length} bouts</small></div>
      {:else}
        <p class="hint">No completed rounds yet.</p>
      {/each}
    </div>
    <div class="card view-card">
      <div class="section-head"><h2>Ceremony at this point</h2><a href="/ceremony?{replayQuery(state)}">Open →</a></div>
      {#each state.medals.slice(0, 6) as medal (medal.id)}
        <div class="rank-row"><span class="medal-dot" style="background:{medalColour(medal.kind)}"></span><strong>{medal.name}</strong><small>{medal.kind}</small></div>
      {:else}
        <p class="hint">Medals appear as the event completes.</p>
      {/each}
    </div>
  </div>

  <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(min(100%, 300px), 1fr)); margin-top:18px;">
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
  .controls { display:flex; gap:8px; flex-wrap:wrap; align-items:center; justify-content:flex-end; }
  .playhead { margin-top:14px; }
  .progress-track { height:10px; border-radius:999px; background:var(--brand-surface-3); overflow:hidden; }
  .progress-fill { height:100%; background:linear-gradient(90deg,var(--brand-primary),var(--brand-accent)); transition:width .2s linear; }
  .scrub { width:100%; margin-top:8px; accent-color:var(--brand-accent); }
  .playhead-meta { display:flex; justify-content:space-between; gap:12px; flex-wrap:wrap; font-size:.78rem; color:var(--brand-ink); opacity:.8; }
  .section-head { display:flex; align-items:center; justify-content:space-between; gap:12px; flex-wrap:wrap; }
  .section-head h2 { margin:0; }
  .section-head a { color:var(--brand-accent); font-weight:700; }
  .now { border-width:2px; }
  .replay { display:flex; align-items:center; justify-content:space-around; gap:18px; margin-top:18px; flex-wrap:wrap; }
  .replay-side { display:flex; align-items:center; gap:14px; font-size:1.2rem; }
  .portrait { width:70px; height:70px; border-radius:20px; display:grid; place-items:center; color:#fff; font-family:var(--font-display); font-weight:800; font-size:1.4rem; }
  .portrait.red { background:linear-gradient(145deg,#ef4444,#7f1d1d); }
  .portrait.blue { background:linear-gradient(145deg,#3b82f6,#1e3a8a); }
  .score { font-family:var(--font-display); font-size:2.8rem; font-weight:800; }
  .vs { color:var(--brand-accent); font-weight:800; font-size:1.4rem; }
  .techniques { margin-top:12px; font-size:.84rem; color:var(--brand-ink); opacity:.82; }
  .inspect-links { display:flex; gap:10px; flex-wrap:wrap; align-items:center; margin-top:14px; padding-top:12px; border-top:1px solid var(--brand-border); }
  .inspect-links a { padding:5px 12px; border-radius:999px; background:var(--brand-surface-3); color:var(--brand-ink); font-size:.82rem; font-weight:700; }
  .timeline { padding-bottom:14px; }
  .event-strip { display:flex; gap:6px; overflow-x:auto; padding:12px 2px 6px; scrollbar-width:thin; }
  .event-chip { flex:0 0 58px; min-height:48px; display:flex; flex-direction:column; align-items:center; justify-content:center; gap:2px; border:1px solid var(--brand-border); border-radius:10px; background:var(--brand-surface-2); color:var(--brand-ink); cursor:pointer; }
  .event-chip small { font-size:.62rem; opacity:.75; max-width:52px; white-space:nowrap; overflow:hidden; text-overflow:ellipsis; }
  .event-chip.past { border-color:var(--win); }
  .event-chip.current { background:var(--brand-primary); color:var(--brand-ink); border-color:var(--brand-accent); transform:translateY(-2px); }
  .views { grid-template-columns:repeat(auto-fit,minmax(min(100%,280px),1fr)); }
  .rank-row { display:grid; grid-template-columns:28px 1fr auto; gap:8px; align-items:center; margin-top:9px; font-size:.84rem; }
  .rank-row small { color:var(--brand-ink); opacity:.75; }
  .medal-dot { width:12px; height:12px; border-radius:50%; }
  .hint { color:var(--brand-ink); opacity:.8; font-size:.84rem; }
  .mini-score { display:flex; align-items:center; justify-content:space-between; gap:8px; margin-top:12px; font-size:.86rem; }
  .mini-score strong { font-family:var(--font-display); font-size:1.5rem; }
  .bar-row { display:grid; grid-template-columns:110px 1fr 34px; align-items:center; gap:10px; margin-top:10px; font-size:.86rem; }
  .bar { height:12px; border-radius:999px; background:var(--brand-surface-3); overflow:hidden; }
  .bar-fill { height:100%; background:var(--brand-primary); border-radius:999px; transition:width .4s ease; }
  .bar-fill.alt { background:var(--win); }
  .bar-fill.gold { background:var(--brand-accent); }
  .medal { display:inline-flex; padding:8px 14px; border-radius:999px; color:#0f172a; font-weight:800; text-transform:capitalize; }
</style>
