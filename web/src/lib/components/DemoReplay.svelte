<script lang="ts">
  import { onMount } from 'svelte';
  import { api, type SimulationBout } from '../api';
  import {
    loadReplay, pauseReplay, playReplay, replaySnapshot, restartReplay, seekReplay,
    setReplaySpeed, startReplay, stepReplay, stopReplay, toggleReplayPlayback,
  } from '../demo-replay';

  let demoAvailable = false;
  let open = false;
  let view: 'mat' | 'table' | 'bracket' | 'ceremony' = 'mat';
  let seed = 42;
  let fighters = 50;

  onMount(() => {
    void api.version().then((info) => { demoAvailable = info.demo === true; }).catch(() => {});
    const unsub = replaySnapshot.subscribe((state) => {
      if (state.active) seed = state.seed;
      if (state.active) fighters = state.fighters;
    });
    return unsub;
  });

  function load() {
    void loadReplay(Number(seed) || 42, Number(fighters) || 50, true);
  }

  function stageGroups(completed: SimulationBout[]) {
    const groups = new Map<string, typeof completed>();
    for (const bout of completed) {
      const label = bout.stage === 'pool' ? `Pool · ${bout.round}` : bout.round;
      groups.set(label, [...(groups.get(label) ?? []), bout]);
    }
    return [...groups.entries()];
  }

  $: state = $replaySnapshot;
  $: groups = stageGroups(state.completed);
</script>

{#if demoAvailable}
  {#if !open}
    <button class="replay-launcher" class:active={state.active} on:click={() => (open = true)}
      aria-label="Open demo replay controls">
      🎬 <span>Demo replay</span>{#if state.active}<em>{state.index + 1}/{state.total}</em>{/if}
    </button>
  {:else}
    <section class="replay-panel" aria-label="Demo tournament replay">
      <header>
        <div>
          <div class="eyebrow">Demo playback</div>
          <strong>Tournament replay</strong>
        </div>
        <div style="display:flex; gap:6px;">
          {#if state.active}
            <button class="mini" on:click={() => stopReplay()} aria-label="Exit replay">✕</button>
          {/if}
          <button class="mini" on:click={() => (open = false)} aria-label="Minimise replay controls">–</button>
        </div>
      </header>

      {#if !state.active}
        <p class="hint">Run a deterministic event, then scrub any point in time. The controls follow you across every view.</p>
        <div class="setup">
          <label>Seed<input type="number" bind:value={seed} aria-label="Replay seed" /></label>
          <label>Fighters<input type="number" min="8" max="200" bind:value={fighters} aria-label="Replay fighter count" /></label>
          <button class="mini primary" on:click={load} disabled={state.loading}>{state.loading ? 'Loading…' : 'Start'}</button>
        </div>
        <button class="wide-btn" on:click={startReplay}>▶ Replay last event</button>
        {#if state.error}<p class="error">{state.error}</p>{/if}
      {:else}
        <div class="timeline">
          <div class="timeline-top">
            <span>Bout {state.index + 1} of {state.total}</span>
            <span>{Math.round(state.progress)}%</span>
          </div>
          <input type="range" min="0" max={Math.max(0, state.total - 1)} value={state.index}
            on:input={(event) => seekReplay(Number((event.target as HTMLInputElement).value))}
            aria-label="Scrub tournament replay" />
        </div>

        <div class="controls">
          <button class="mini" on:click={() => restartReplay()} aria-label="Restart replay">↻</button>
          <button class="mini" on:click={() => stepReplay(-1)} aria-label="Previous bout">⏮</button>
          <button class="mini primary" on:click={toggleReplayPlayback} aria-label={state.playing ? 'Pause replay' : 'Play replay'}>
            {state.playing ? '⏸' : '▶'}
          </button>
          <button class="mini" on:click={() => stepReplay(1)} aria-label="Next bout">⏭</button>
          <select value={state.speed} on:change={(event) => setReplaySpeed(Number((event.target as HTMLSelectElement).value))}
            aria-label="Replay speed">
            <option value={120}>4×</option>
            <option value={450}>1×</option>
            <option value={900}>½×</option>
            <option value={1800}>¼×</option>
          </select>
        </div>

        <div class="tabs" role="tablist" aria-label="Replay state views">
          {#each [['mat','Mat'],['table','Table'],['bracket','Bracket'],['ceremony','Ceremony']] as [id, label] (id)}
            <button role="tab" aria-selected={view === id} class:active={view === id} on:click={() => (view = id as typeof view)}>{label}</button>
          {/each}
        </div>

        <div class="replay-body">
          {#if view === 'mat' && state.bout}
            <div class="duel">
              <div><span class="corner red">{state.bout.redId}</span><strong>{state.bout.pointsRed}</strong></div>
              <em>vs</em>
              <div><strong>{state.bout.pointsBlue}</strong><span class="corner blue">{state.bout.blueId}</span></div>
            </div>
            <p class="hint">{state.bout.stage} · {state.bout.round} · {state.bout.outcome.replaceAll('_', ' ')} · {state.bout.method}</p>
          {:else if view === 'mat'}
            <p class="hint">No bout at this point in the tournament.</p>
          {:else if view === 'table'}
            {#each state.standings.slice(0, 5) as row, i (row.fighterId)}
              <div class="rank"><span>#{i + 1}</span><strong>{row.name}</strong><em>{row.club}</em><b>{row.points}</b></div>
            {/each}
          {:else if view === 'bracket'}
            {#each groups.slice(-3) as [label, bouts] (label)}
              <div class="group"><strong>{label}</strong>
                {#each bouts.slice(-2) as bout (bout.id)}
                  <span>{bout.redId} vs {bout.blueId} → {bout.outcome.replaceAll('_', ' ')}</span>
                {/each}
              </div>
            {:else}<p class="hint">No bracket results yet.</p>{/each}
          {:else}
            {#each ['gold', 'silver', 'bronze'] as kind (kind)}
              <div class="group"><strong>{kind}</strong>
                {#each state.medals.filter((medal) => medal.kind === kind) as medal (medal.id)}
                  <span>{medal.name}</span>
                {:else}<span class="hint">To be decided</span>{/each}
              </div>
            {/each}
          {/if}
        </div>

        <p class="hint small">Seed {state.seed} · {state.fighters} fighters · changes follow every display tab.</p>
      {/if}
    </section>
  {/if}
{/if}

<style>
  .replay-launcher {
    position:fixed; right:12px; bottom:52px; z-index:890; display:flex; align-items:center; gap:8px;
    padding:9px 14px; border-radius:999px; border:1px solid var(--brand-border);
    background:var(--brand-surface-2); color:var(--brand-ink); box-shadow:var(--brand-shadow);
    cursor:pointer; font-size:.8rem; font-weight:800;
  }
  .replay-launcher.active { border-color:var(--brand-accent); box-shadow:0 0 18px color-mix(in srgb, var(--brand-accent) 35%, transparent); }
  .replay-launcher em { font-style:normal; color:var(--brand-accent); }
  .replay-panel {
    position:fixed; right:12px; bottom:52px; z-index:890; width:min(370px, calc(100vw - 24px));
    max-height:min(70vh, 640px); overflow:auto; display:grid; gap:10px; padding:14px;
    border-radius:16px; border:1px solid var(--brand-border); background:var(--brand-surface-2);
    box-shadow:var(--brand-shadow);
  }
  header { display:flex; justify-content:space-between; align-items:center; gap:10px; }
  header strong { font-family:var(--font-display); font-size:1rem; }
  .hint { margin:0; color:var(--brand-muted); font-size:.76rem; line-height:1.4; }
  .hint.small { font-size:.68rem; }
  .error { margin:0; color:var(--loss); font-size:.76rem; }
  .setup { display:grid; grid-template-columns:1fr 1fr auto; gap:8px; align-items:end; }
  .setup label { display:grid; gap:3px; font-size:.68rem; color:var(--brand-muted); }
  .setup input { width:100%; padding:7px; border-radius:8px; background:var(--brand-surface-3); color:var(--brand-ink); border:1px solid var(--brand-border); }
  .wide-btn, .mini {
    padding:8px 10px; border-radius:9px; border:1px solid var(--brand-border);
    background:var(--brand-surface-3); color:var(--brand-ink); cursor:pointer; font-size:.78rem; font-weight:700;
  }
  .wide-btn { width:100%; }
  .mini.primary { background:var(--brand-primary-fill); color:var(--brand-on-primary); border-color:transparent; }
  .mini:disabled { opacity:.55; cursor:wait; }
  .timeline { display:grid; gap:4px; }
  .timeline-top { display:flex; justify-content:space-between; font-size:.72rem; color:var(--brand-muted); }
  .timeline input { width:100%; accent-color:var(--brand-accent); }
  .controls { display:grid; grid-template-columns:repeat(4, auto) 1fr; gap:6px; }
  .controls select { min-width:0; padding:7px; border-radius:9px; background:var(--brand-surface-3); color:var(--brand-ink); border:1px solid var(--brand-border); }
  .tabs { display:grid; grid-template-columns:repeat(4, 1fr); gap:4px; }
  .tabs button { padding:6px 2px; border-radius:8px; border:1px solid var(--brand-border); background:transparent; color:var(--brand-muted); font-size:.68rem; cursor:pointer; }
  .tabs button.active { background:var(--brand-primary-fill); color:var(--brand-on-primary); border-color:transparent; font-weight:800; }
  .replay-body { display:grid; gap:7px; min-height:90px; max-height:220px; overflow:auto; }
  .duel { display:flex; align-items:center; justify-content:center; gap:12px; }
  .duel div { display:flex; align-items:center; gap:8px; }
  .duel strong { font-family:var(--font-display); font-size:1.5rem; }
  .duel em { color:var(--brand-accent); font-style:normal; font-weight:800; }
  .corner { display:grid; place-items:center; min-width:54px; padding:5px; border-radius:8px; color:#fff; font-size:.68rem; font-weight:800; }
  .corner.red { background:#b91c1c; }
  .corner.blue { background:#1d4ed8; }
  .rank { display:grid; grid-template-columns:28px 1fr auto 30px; gap:8px; align-items:center; padding:5px 8px; border-radius:8px; background:var(--brand-surface-3); font-size:.72rem; }
  .rank em { color:var(--brand-muted); font-style:normal; }
  .rank b { text-align:right; color:var(--brand-accent); }
  .group { display:grid; gap:3px; padding:7px 9px; border-radius:8px; background:var(--brand-surface-3); font-size:.7rem; }
  .group span { color:var(--brand-muted); }
  @media (max-width: 700px) {
    .replay-launcher, .replay-panel { right:8px; bottom:132px; }
    .replay-panel { max-height:45vh; width:calc(100vw - 16px); }
  }
</style>
