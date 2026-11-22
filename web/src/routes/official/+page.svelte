<script lang="ts">
  import { onMount } from 'svelte';
  import {
    api, livePoll, getJson, type DemoFight, type FightState, type KataPerformance,
  } from '../../lib/api';
  import { notifySuccess, notifyError } from '../../lib/toast';
  import ConfirmDialog from '../../lib/components/ConfirmDialog.svelte';
  import { guideAction } from '../../lib/guide';

  type Technique = { id: string; english: string; romaji: string; points: number; kind: string };
  let fights: DemoFight[] = [];
  let techniques: Technique[] = [];
  let selected = '';
  let state: FightState | null = null;
  let corner: 'red' | 'blue' = 'red';
  let search = '';
  let tab: 'fight' | 'kata' = 'fight';
  let kataPerformances: KataPerformance[] = [];

  type Pending = { kind: 'event' | 'result' | 'submission' | 'kata'; label: string; action: () => Promise<unknown> };
  let pending: Pending | null = null;
  let confirmOpen = false;

  onMount(() => {
    void getJson<{ techniques: Technique[] }>('/api/v1/demo/techniques')
      .then((data) => (techniques = data.techniques)).catch(() => {});
    const stopFights = livePoll(api.fights, 2000, (data) => {
      fights = data.fights;
      if (!selected) selected = (data.fights.find((fight) => fight.state === 'live') ?? data.fights[0])?.id ?? '';
    });
    const stopKata = livePoll(api.kata, 5000, (data) => (kataPerformances = data.performances));
    const stopStream = api.eventStream((message) => {
      if (message.payload && message.payload.fightId === selected) state = message.payload;
    });
    return () => { stopFights(); stopKata(); stopStream(); };
  });

  $: selectedFight = fights.find((fight) => fight.id === selected);
  let loadedFor = '';
  $: if (selected && selected !== loadedFor) {
    loadedFor = selected;
    state = null;
    void api.fightState(selected).then((next) => (state = next)).catch(() => {});
  }
  $: filtered = techniques.filter((technique) =>
    `${technique.english} ${technique.romaji}`.toLowerCase().includes(search.toLowerCase()));

  function ask(item: Pending) { pending = item; confirmOpen = true; }
  async function run() {
    if (!pending) return;
    const kind = pending.kind;
    try {
      await pending.action();
      notifySuccess(`Confirmed: ${pending.label}`);
      if (kind === 'event') guideAction('official-entry');
      if (kind === 'result' || kind === 'submission') guideAction('official-result');
      if (kind === 'kata') guideAction('kata-score');
    } catch (error) {
      notifyError(error instanceof Error ? error.message : 'Could not record action');
    }
    pending = null;
  }
  function logTechnique(technique: Technique) {
    const label = `${corner.toUpperCase()} · ${technique.english} / ${technique.romaji} · ${technique.points} pts`;
    ask({ kind: 'event', label, action: async () => {
      state = await api.fightEvent(selected, {
        corner, kind: technique.kind, points: technique.points, technique: `${technique.english} (${technique.romaji})`,
        card: technique.kind === 'penalty' ? technique.id : '',
      });
    } });
  }
  function submitResult(outcome: string) {
    const label = outcome === 'win_red' ? 'Red wins' : outcome === 'win_blue' ? 'Blue wins' : 'Draw';
    ask({ kind: 'result', label: `Result: ${label}`, action: async () => {
      state = await api.submitFight(selected, outcome);
    } });
  }
  function submissionOverride(winner: 'red' | 'blue') {
    ask({ kind: 'submission', label: `Instant win by submission — ${winner}`, action: async () => {
      state = await api.fightEvent(selected, { corner: winner, kind: 'submission', points: 0, technique: 'Tap-out' });
      state = await api.submitFight(selected, winner === 'red' ? 'win_red' : 'win_blue');
    } });
  }
  async function scoreKata(id: string, judge: number, section: number, value: number) {
    if (value < 0 || value > 10) return;
    ask({ kind: 'kata', label: `Kata section ${section} · judge ${judge} · ${value}/10`, action: async () => {
      kataPerformances = (await api.kataScore(id, { judge, score: value })).performances;
    } });
  }
</script>

<svelte:head><title>J3 · Official console</title></svelte:head>

<div class="shell">
  <header class="topbar">
    <div class="brand"><a class="brand-mark" href="/public" aria-label="Public dashboard">J3</a><div><div class="brand-name">Official console</div><div class="brand-sub">One tap to log · confirm before it counts</div></div></div>
    <div class="tabs">
      <button class="tab" class:active={tab === 'fight'} on:click={() => (tab = 'fight')}>🥊 Fight</button>
      <button class="tab" data-guide="official-tab-kata" class:active={tab === 'kata'} on:click={() => (tab = 'kata')}>🥋 Kata panel</button>
    </div>
  </header>

  {#if tab === 'fight'}
    <div class="grid" style="grid-template-columns:repeat(auto-fit,minmax(280px,1fr));">
      <div class="card">
        <div class="eyebrow">1 · Select contest</div>
        <h2>{selectedFight?.mat ?? 'Mat'} · {selectedFight?.round ?? ''}</h2>
        <select class="tab" data-guide="official-contest-select" style="width:100%; color:var(--brand-ink);" bind:value={selected} aria-label="Select contest" on:change={() => guideAction('official-contest-selected')}>
          {#each fights as fight (fight.id)}<option value={fight.id}>{fight.mat} · {fight.red.name} vs {fight.blue.name}</option>{/each}
        </select>
        {#if selectedFight}
          <div class="scoreline">
            <strong style="color:var(--red-corner);">{selectedFight.red.name}</strong>
            <span>{state?.redScore ?? 0} — {state?.blueScore ?? 0}</span>
            <strong style="color:var(--blue-corner);">{selectedFight.blue.name}</strong>
          </div>
        {/if}
        <div class="action-row">
          <button class="btn" on:click={() => api.startFight(selected)}>▶ Start</button>
          <button class="btn" on:click={() => api.pauseFight(selected)}>⏸ Pause</button>
          <button class="btn" on:click={() => api.resumeFight(selected)}>▶ Resume</button>
        </div>
      </div>

      <div class="card">
        <div class="eyebrow">2 · Whose point?</div>
        <h2>Active corner</h2>
        <div class="corner-picker">
          <button class="corner-btn red" data-guide="official-corner-red" class:active={corner === 'red'} on:click={() => { corner = 'red'; guideAction('official-corner'); }}>🔴 RED<br />{selectedFight?.red.name ?? ''}</button>
          <button class="corner-btn blue" class:active={corner === 'blue'} on:click={() => { corner = 'blue'; guideAction('official-corner'); }}>🔵 BLUE<br />{selectedFight?.blue.name ?? ''}</button>
        </div>
        <input class="tab" style="width:100%; color:var(--brand-ink); margin-top:12px;" placeholder="Search technique / romaji…" bind:value={search} />
        <div class="tech-grid">
          {#each filtered as technique, techniqueIndex (technique.id)}
            <button class="tech" data-guide={techniqueIndex === 0 ? 'official-tech-button' : undefined} class:penalty={technique.kind === 'penalty'} on:click={() => logTechnique(technique)}>
              <strong>{technique.english}</strong><span class="muted">{technique.romaji}</span>
              <span class="points">{technique.kind === 'submission' ? 'WIN' : technique.kind === 'penalty' ? 'CARD' : `+${technique.points}`}</span>
            </button>
          {/each}
        </div>
      </div>

      <div class="card">
        <div class="eyebrow">3 · Finish & override</div>
        <h2>Confirm outcome</h2>
        <div class="action-row">
          <button class="btn" data-guide="official-result-win-red" on:click={() => submitResult('win_red')}>Red wins</button>
          <button class="btn" on:click={() => submitResult('win_blue')}>Blue wins</button>
          <button class="btn" on:click={() => submitResult('draw')}>Draw</button>
        </div>
        <div class="eyebrow" style="margin-top:20px;">Submission override — instant win, points negated</div>
        <div class="action-row">
          <button class="btn warn" on:click={() => submissionOverride('red')}>🔴 Red submits</button>
          <button class="btn warn" on:click={() => submissionOverride('blue')}>🔵 Blue submits</button>
        </div>
        <button class="btn" style="margin-top:18px;" on:click={() => api.undoFightEvent(selected)}>↶ Undo last entry</button>
        <div class="muted small" style="margin-top:14px;">Every action is attributed and audited. Nothing is final until confirmed.</div>
      </div>
    </div>
  {:else}
    <div class="card">
      <div class="eyebrow">Kata judging · 3 judges · score /10</div>
      <h2>Section scores with automatic tie retry</h2>
      <div style="display:grid; gap:12px; margin-top:14px;">
        {#each kataPerformances as performance (performance.id)}
          <div class="kata-card">
            <div style="display:flex; justify-content:space-between; gap:12px; flex-wrap:wrap;">
              <strong>{performance.name}</strong>
              <span class="badge fighter">{performance.kataName} · {performance.total.toFixed(1)}</span>
            </div>
            <div class="kata-grid">
              {#each ['Etiquette','Demonstration','Would the application work'] as section, i}
                {#each [1,2,3] as judge}
                  <label class="small">J{judge} · {section}
                    <input type="number" min="0" max="10" step="0.5" placeholder="—"
                      on:change={(event) => scoreKata(performance.id, judge, i+1, Number((event.target as HTMLInputElement).value))} />
                  </label>
                {/each}
              {/each}
            </div>
          </div>
        {:else}<span class="muted small">No kata performances yet.</span>{/each}
      </div>
    </div>
  {/if}
</div>

<ConfirmDialog
  bind:open={confirmOpen}
  title="Confirm this entry?"
  message={pending?.label ?? ''}
  confirmLabel="Confirm & log"
  onConfirm={run}
/>

<style>
  .scoreline { display:flex; justify-content:space-between; gap:10px; margin-top:14px; font-size:1.1rem; }
  .corner-picker { display:grid; grid-template-columns:1fr 1fr; gap:10px; margin-top:12px; }
  .corner-btn { padding:14px; border-radius:12px; border:2px solid transparent; cursor:pointer; background:var(--brand-surface-3); color:var(--brand-ink); }
  .corner-btn.red.active { border-color:var(--red-corner); }
  .corner-btn.blue.active { border-color:var(--blue-corner); }
  .tech-grid { display:grid; grid-template-columns:repeat(auto-fill,minmax(150px,1fr)); gap:8px; margin-top:12px; }
  .tech { display:grid; gap:2px; text-align:left; padding:9px 10px; border-radius:10px; border:1px solid var(--brand-border); background:var(--brand-surface-3); color:var(--brand-ink); cursor:pointer; }
  .tech:hover { border-color:var(--brand-primary); }
  .tech.penalty { border-color:var(--draw); }
  .tech .points { color:var(--brand-accent); font-weight:800; }
  .kata-card { padding:14px; border-radius:12px; background:var(--brand-surface-3); border:1px solid var(--brand-border); }
  .kata-grid { display:grid; grid-template-columns:repeat(auto-fit,minmax(120px,1fr)); gap:8px; margin-top:10px; }
  .kata-grid input { width:100%; padding:7px; border-radius:8px; background:var(--brand-surface-2); color:var(--brand-ink); border:1px solid var(--brand-border); }
</style>
