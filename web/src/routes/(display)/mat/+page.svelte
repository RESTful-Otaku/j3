<script lang="ts">
  import { onMount } from 'svelte';
  import { api, livePoll, type DemoFight, type EventDisplayInfo, type FightState } from '../../../lib/api';
  import { notifySuccess, notifyError } from '../../../lib/toast';
  import ConfirmDialog from '../../../lib/components/ConfirmDialog.svelte';

  let event: EventDisplayInfo | null = null;
  let fights: DemoFight[] = [];
  let selectedMat = 'Mat 1';
  let state: FightState | null = null;
  let confirmOpen = false;
  let outcome = 'win_red';

  async function loadState(id: string) {
    try { state = await api.fightState(id); } catch { state = null; }
  }
  onMount(() => {
    const stopEvent = livePoll(api.event, 10000, (data) => (event = data));
    const stopFights = livePoll(api.fights, 3000, (data) => {
      fights = data.fights;
      if (!selected) selected = (data.fights.find((f) => f.mat === selectedMat) ?? data.fights[0])?.id ?? '';
    });
    const stopStream = api.eventStream((message) => {
      if (message.payload && message.payload.fightId === selected) state = message.payload;
    });
    return () => { stopEvent(); stopFights(); stopStream(); };
  });

  let selected = '';
  $: matFights = fights.filter((fight) => fight.mat === selectedMat);
  $: selectedFight = fights.find((fight) => fight.id === selected);
  $: if (selected && !state) { void loadState(selected); }

  async function act(fn: () => Promise<FightState>, label: string) {
    try { state = await fn(); notifySuccess(label); }
    catch (error) { notifyError(error instanceof Error ? error.message : label); }
  }
  async function score(corner: 'red' | 'blue', points: number) {
    if (!selected) return;
    await act(() => api.fightEvent(selected, { corner, points, kind: 'score', technique: 'Score' }), `+${points} ${corner}`);
  }
  async function penalty(corner: 'red' | 'blue') {
    if (!selected) return;
    await act(() => api.fightEvent(selected, { corner, kind: 'penalty', card: 'yellow' }), 'Penalty recorded');
  }
  async function hold(corner: 'red' | 'blue') {
    if (!selected) return;
    await act(() => api.startHold(selected, corner, 20), 'Hold timer started');
  }
  async function submit() {
    if (!selected) return;
    await act(() => api.submitFight(selected, outcome), 'Result submitted');
    confirmOpen = false;
  }
</script>

<svelte:head><title>J3 · Mat scorekeeper</title></svelte:head>

<div class="display-shell">
  <header class="display-header">
    <div><div class="event-mark"><span class="crest">J3</span>Mat control</div>
      <div class="meta">Scorekeeper & timekeeper · authoritative board</div></div>
    <div style="display:flex; gap:8px; flex-wrap:wrap;">
      {#each event?.mats ?? [] as mat}
        <button class="tab" class:active={selectedMat === mat} on:click={() => { selectedMat = mat; selected = ''; state = null; }}>{mat}</button>
      {/each}
    </div>
  </header>

  <div class="card">
    <div class="section-title" style="margin-top:0;">
      <h2>{selectedMat} · contest queue</h2>
      <span class="pill">{matFights.length} contests</span>
    </div>
    <div class="tabs">
      {#each matFights as fight (fight.id)}
        <button class="tab" class:active={selected === fight.id} on:click={() => { selected = fight.id; void loadState(fight.id); }}>
          {fight.red.name} vs {fight.blue.name} · {fight.state}
        </button>
      {/each}
    </div>
  </div>

  {#if selectedFight && state}
    <div class="card" style="margin-top:18px;">
      <div class="eyebrow">{selectedFight.round} · {selectedFight.category}</div>
      <h2>{selectedFight.red.name} <span class="muted">vs</span> {selectedFight.blue.name}</h2>
      <div class="score-row">
        <div><span class="corner-label red">RED · {selectedFight.red.belt}</span><div class="big-score">{state.redScore}</div></div>
        <div style="text-align:center;">
          <div class="clock" class:paused={!state.running}>{String(Math.floor(state.seconds / 60)).padStart(2,'0')}:{String(state.seconds % 60).padStart(2,'0')}</div>
          {#if state.holdSeconds > 0}<div class="hold-badge">HOLD · {state.holdCorner} · {state.holdSeconds}s</div>{/if}
          <div class="muted small">Round {state.round} · {state.running ? 'running' : 'paused'}</div>
        </div>
        <div style="text-align:right;"><span class="corner-label blue">BLUE · {selectedFight.blue.belt}</span><div class="big-score">{state.blueScore}</div></div>
      </div>

      <div class="control-grid">
        <button class="btn primary" on:click={() => act(() => api.startFight(selected), 'Fight started')}>▶ Start</button>
        <button class="btn" on:click={() => act(() => api.pauseFight(selected), 'Timer paused')}>⏸ Pause</button>
        <button class="btn" on:click={() => act(() => api.resumeFight(selected), 'Timer resumed')}>▶ Resume</button>
        <button class="btn warn" on:click={() => (confirmOpen = true)}>✓ Submit result</button>
      </div>
      <div class="control-grid">
        {#each [1, 2, 3, 4] as points}
          <button class="btn" on:click={() => score('red', points)}>Red +{points}</button>
        {/each}
        {#each [1, 2, 3, 4] as points}
          <button class="btn" on:click={() => score('blue', points)}>Blue +{points}</button>
        {/each}
      </div>
      <div class="control-grid">
        <button class="btn" on:click={() => penalty('red')}>🟨 Red foul</button>
        <button class="btn" on:click={() => penalty('blue')}>🟨 Blue foul</button>
        <button class="btn" on:click={() => hold('red')}>⏱ Red hold</button>
        <button class="btn" on:click={() => hold('blue')}>⏱ Blue hold</button>
        <button class="btn" on:click={() => act(() => api.undoFightEvent(selected), 'Last event undone')}>↶ Undo</button>
      </div>
    </div>
  {:else}
    <div class="placeholder" style="margin-top:18px;">Choose a contest to control this mat.</div>
  {/if}
</div>

<ConfirmDialog
  open={confirmOpen}
  title="Submit final result?"
  message="This locks the contest, publishes it to every display and records the decision in the audit trail."
  confirmLabel="Submit result"
  onConfirm={submit}
/>

<style>
  .score-row { display:grid; grid-template-columns:1fr 1fr 1fr; align-items:center; gap:18px; margin:18px 0; }
  .corner-label { font-weight:800; font-size:1.2rem; }
  .corner-label.red { color:#ef4444; } .corner-label.blue { color:#3b82f6; }
  .big-score { font-family:var(--font-display); font-size:clamp(3rem,8vw,7rem); font-weight:800; }
  .clock { font-family:var(--font-display); font-size:2.6rem; font-weight:800; }
  .clock.paused { color:var(--draw); animation:pulse 1s infinite alternate; }
  .hold-badge { color:#fff; background:var(--brand-accent); border-radius:10px; padding:6px 10px; font-weight:800; }
  .control-grid { display:flex; gap:8px; flex-wrap:wrap; margin-top:12px; }
  @keyframes pulse { from { opacity:1; } to { opacity:.5; } }
</style>
