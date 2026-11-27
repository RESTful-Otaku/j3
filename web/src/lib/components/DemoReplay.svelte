<script lang="ts">
  import {
    appMode, replayQuery, replaySnapshot, restartReplay, seekReplay, setAppMode,
    setReplaySpeed, stepReplay, toggleReplayPlayback,
  } from '../demo-replay';

  let view: 'mat' | 'table' | 'bracket' | 'ceremony' = 'mat';
  let seed = 42;
  let fighters = 50;
  let collapsed = false;

  function toggleCollapsed() {
    collapsed = !collapsed;
    try { localStorage.setItem('j3-replay-collapsed', collapsed ? '1' : '0'); } catch { /* ignore */ }
  }

  // Restore collapsed preference.
  (function initCollapsed() {
    try { collapsed = localStorage.getItem('j3-replay-collapsed') === '1'; } catch { /* ignore */ }
  })();

  function stageGroups(completed: typeof $replaySnapshot.bouts) {
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

{#if $appMode === 'sim'}
  {#if collapsed}
    <button class="replay-chip" on:click={toggleCollapsed} aria-label="Show replay panel" title="Show replay panel">
      🎬 <span>Replay · {state.index + 1}/{state.total}</span>
    </button>
  {:else}
  <section class="replay-panel" aria-label="Simulated tournament controls">
    <header>
      <div>
        <div class="eyebrow">Virtual event · global playback</div>
        <strong>🎬 Simulated tournament</strong>
      </div>
      <div class="header-actions">
        <button class="mini" on:click={toggleCollapsed} aria-label="Minimize replay panel" title="Minimize replay panel">−</button>
        <button class="mini exit" on:click={() => setAppMode('live')} aria-label="Return to live mode">● Live</button>
      </div>
    </header>

    {#if state.loading}
      <p class="hint">Loading a deterministic tournament…</p>
    {:else if state.error}
      <p class="error">{state.error}</p>
      <button class="wide-btn" on:click={() => setAppMode('sim')}>Retry simulation</button>
    {:else if !state.active}
      <p class="hint">Simulation is starting…</p>
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
          <option value={80}>8×</option>
          <option value={160}>4×</option>
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
          {#if state.bout.techniques?.length}<p class="hint small">{state.bout.techniques.slice(-6).join(' · ')}</p>{/if}
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

      <div class="inspect-links">
        <span>Inspect this moment</span>
        <a href="/public?{replayQuery(state)}">Public</a>
        <a href="/tables?{replayQuery(state)}">Tables</a>
        <a href="/brackets?{replayQuery(state)}">Brackets</a>
        <a href="/ceremony?{replayQuery(state)}">Ceremony</a>
      </div>
      <a class="wide-btn console-link" href="/simulate?{replayQuery(state)}">🎛 Open full replay console</a>
      <p class="hint small">Seed {state.seed} · {state.fighters} fighters · {state.speed === 80 ? '8×' : state.speed === 160 ? '4×' : state.speed === 450 ? '1×' : state.speed === 900 ? '½×' : '¼×'} playback</p>
    {/if}
  </section>
  {/if}
{/if}

<style>
  .replay-panel {
    position:fixed; right:calc(12px + var(--safe-right)); bottom:calc(52px + var(--safe-bottom)); z-index:890; width:min(370px, calc(100vw - 24px - var(--safe-left) - var(--safe-right)));
    max-height:min(72vh, 660px); overflow:auto; display:grid; gap:10px; padding:14px;
    border-radius:16px; border:1px solid var(--brand-accent); background:var(--brand-surface-2);
    box-shadow:0 0 0 2px color-mix(in srgb, var(--brand-accent) 20%, transparent), var(--brand-shadow);
  }
  header { display:flex; justify-content:space-between; align-items:center; gap:10px; }
  header strong { font-family:var(--font-display); font-size:1rem; }
  .eyebrow { text-transform:uppercase; letter-spacing:.12em; font-size:.6rem; font-weight:900; color:var(--brand-muted); }
  .hint { margin:0; color:var(--brand-muted); font-size:.76rem; line-height:1.4; }
  .hint.small { font-size:.68rem; }
  .error { margin:0; color:var(--loss); font-size:.76rem; }
  .wide-btn, .mini {
    padding:8px 10px; border-radius:9px; border:1px solid var(--brand-border);
    background:var(--brand-surface-3); color:var(--brand-ink); cursor:pointer; font-size:.78rem; font-weight:700;
  }
  .wide-btn { width:100%; text-align:center; text-decoration:none; }
  .console-link { border-color:var(--brand-accent); color:var(--brand-accent); }
  .mini.exit { border-color:var(--win); }
  .header-actions { display:flex; gap:6px; align-items:center; }
  .replay-chip { position:fixed; right:calc(12px + var(--safe-right)); bottom:calc(52px + var(--safe-bottom)); z-index:890; display:flex; align-items:center; gap:8px; padding:10px 14px; border-radius:999px; border:1px solid var(--brand-accent); background:var(--brand-surface-2); color:var(--brand-ink); font-size:.78rem; font-weight:800; cursor:pointer; box-shadow:var(--brand-shadow); }
  .replay-chip:hover { border-color:var(--brand-primary); }
  @media (max-width: 700px) { .replay-chip { right:calc(8px + var(--safe-right)); bottom:calc(132px + var(--safe-bottom)); } }
  .mini.primary { background:var(--brand-primary-fill); color:var(--brand-on-primary); border-color:transparent; }
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
  .corner { display:grid; place-items:center; min-width:54px; padding:5px; border-radius:8px; font-size:.68rem; font-weight:800; box-shadow:inset 0 0 0 1px rgba(15,23,42,.3); }
  .corner.red { background:var(--red-corner); color:var(--on-red-corner); }
  .corner.blue { background:var(--blue-corner); color:var(--on-blue-corner); }
  .rank { display:grid; grid-template-columns:28px 1fr auto 30px; gap:8px; align-items:center; padding:5px 8px; border-radius:8px; background:var(--brand-surface-3); font-size:.72rem; }
  .rank em { color:var(--brand-muted); font-style:normal; }
  .rank b { text-align:right; color:var(--brand-accent); }
  .group { display:grid; gap:3px; padding:7px 9px; border-radius:8px; background:var(--brand-surface-3); font-size:.7rem; }
  .group span { color:var(--brand-muted); }
  .inspect-links { display:flex; gap:5px; flex-wrap:wrap; align-items:center; }
  .inspect-links span { font-size:.68rem; color:var(--brand-muted); }
  .inspect-links a { padding:4px 8px; border-radius:999px; background:var(--brand-surface-3); color:var(--brand-ink); font-size:.66rem; font-weight:700; text-decoration:none; }
  @media (max-width: 700px) {
    .replay-panel { right:calc(8px + var(--safe-right)); bottom:calc(132px + var(--safe-bottom)); max-height:45vh; width:calc(100vw - 16px - var(--safe-left) - var(--safe-right)); }
  }
</style>
