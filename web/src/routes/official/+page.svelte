<script lang="ts">
  import { onMount } from 'svelte';
  import {
    api, livePoll, getJson,
    type DemoFight, type FightState, type KataPerformance,
    type Technique, type TechniqueCatalogue, type TechniquePointOption,
    type TechnicalResultRequest,
  } from '../../lib/api';
  import { notifySuccess, notifyError } from '../../lib/toast';
  import ConfirmDialog from '../../lib/components/ConfirmDialog.svelte';
  import { guideAction } from '../../lib/guide';
  import { onLiveChange } from '../../lib/live-updates';

  let contests: DemoFight[] = [];
  let selected = '';
  let role = '';
  let state: FightState | null = null;
  let corner: 'red' | 'blue' = 'red';
  let tab: 'fight' | 'kata' = 'fight';
  let kataPerformances: KataPerformance[] = [];

  let catalogue: TechniqueCatalogue | null = null;
  let search = '';
  let activeGroup = '';
  let selectedTechnique: Technique | null = null;
  let selectedPointLabel = '';
  let detail = '';
  let technicalReason: TechnicalResultRequest['reason'] = 'injury';
  let technicalWinner: 'red' | 'blue' = 'red';

  type Pending = {
    kind: 'event' | 'card' | 'result' | 'submission' | 'technical' | 'kata';
    label: string;
    action: () => Promise<unknown>;
  };
  let pending: Pending | null = null;
  let confirmOpen = false;
  let stopContestFallback: (() => void) | null = null;

  const reasonLabels: Record<TechnicalResultRequest['reason'], string> = {
    injury: 'Injury',
    forfeit: 'Forfeit',
    cheating: 'Cheating / unsportsmanlike conduct',
    dq: 'Disqualification',
    no_show: 'No show',
    withdrawal: 'Withdrawal',
    referee: 'Referee decision',
    other: 'Other recorded reason',
  };

  onMount(() => {
    const loadContests = () => api.myContests()
      .then((data) => {
        contests = data.contests;
        role = data.role;
        if (!selected) selected = (contests.find((fight) => fight.state === 'live') ?? contests[0])?.id ?? '';
      })
      .catch(() => {
        stopContestFallback = livePoll(api.fights, 2000, (data) => {
          contests = data.fights;
          if (!selected) selected = (data.fights.find((fight) => fight.state === 'live') ?? data.fights[0])?.id ?? '';
        });
      });
    void loadContests();
    void api.techniques().then((data) => {
      catalogue = data;
      const first = data.techniques.find((item) => item.kind === 'score');
      if (first) selectTechnique(first, data.pointOptions);
    }).catch(() => {});
    const stopKata = livePoll(api.kata, 5000, (data) => (kataPerformances = data.performances));
    const stopStream = api.eventStream((message) => {
      if (message.payload && message.payload.fightId === selected) state = message.payload;
    });
    const clock = setInterval(() => {
      if (!selected) return;
      void api.fightState(selected).then((next) => (state = next)).catch(() => {});
    }, 1000);
    const stopLive = onLiveChange(['fights', 'officials'], () => void loadContests());
    const stopKataLive = onLiveChange(['kata', 'medals'], () => void api.kata().then((data) => (kataPerformances = data.performances)).catch(() => {}));
    return () => { stopContestFallback?.(); stopKata(); stopStream(); stopLive(); stopKataLive(); clearInterval(clock); };
  });

  $: selectedFight = contests.find((fight) => fight.id === selected);
  let loadedFor = '';
  $: if (selected && selected !== loadedFor) {
    loadedFor = selected;
    state = null;
    void api.fightState(selected).then((next) => (state = next)).catch(() => {});
  }
  $: needle = search.trim().toLowerCase();
  $: filtered = (catalogue?.techniques ?? []).filter((item) => {
    if (activeGroup && item.group !== activeGroup) return false;
    if (!needle) return true;
    return `${item.english} ${item.romaji} ${item.japanese ?? ''} ${item.group} ${item.tags.join(' ')}`
      .toLowerCase().includes(needle);
  });
  $: penaltyYellow = (catalogue?.techniques ?? []).find((item) => item.kind === 'penalty' && item.card === 'yellow');
  $: penaltyRed = (catalogue?.techniques ?? []).find((item) => item.kind === 'penalty' && item.card === 'red');
  $: levelScores = Boolean(state) && state!.redScore === state!.blueScore;

  function ask(item: Pending) { pending = item; confirmOpen = true; }

  async function run() {
    if (!pending) return;
    const kind = pending.kind;
    try {
      await pending.action();
      notifySuccess(`Confirmed: ${pending.label}`);
      if (kind === 'event') guideAction('official-entry');
      if (kind === 'card') guideAction('official-card');
      if (kind === 'result' || kind === 'submission' || kind === 'technical') guideAction('official-result');
      if (kind === 'kata') guideAction('kata-score');
    } catch (error) {
      notifyError(error instanceof Error ? error.message : 'Could not record action');
    }
    pending = null;
  }

  function selectTechnique(technique: Technique, options: TechniquePointOption[] = catalogue?.pointOptions ?? []) {
    selectedTechnique = technique;
    const aliases: Record<string, string[]> = {
      strike: ['strike', 'punch', 'kick', 'tsuki', 'geri'],
      kick: ['strike', 'kick', 'geri'],
      throw: ['throw', 'nage', 'kaeshi'],
      takedown: ['takedown', 'taoshi', 'sweep', 'reap'],
      sweep: ['takedown', 'sweep', 'harai'],
      reap: ['takedown', 'reap', 'harai'],
      hook: ['takedown', 'strike', 'hook'],
      hold: ['hold', 'pin', 'osae', 'kesa', 'control'],
      ground: ['hold', 'ground', 'pin'],
      control: ['hold', 'control', 'guard'],
      break: ['hold', 'break', 'escape'],
      standing: ['strike', 'standing', 'throw'],
      submission: ['submission', 'tap', 'choke', 'lock'],
      ippon: ['ippon'],
      wazaari: ['waza'],
      penalty: ['penalty', 'card', 'shido', 'hansoku'],
      kata: ['kata', 'excellent', 'very good'],
    };
    const terms = aliases[technique.pointKey] ?? [technique.pointKey];
    const match = options.find((option) => {
      const label = `${option.key} ${option.label}`.toLowerCase();
      return terms.some((term) => label.includes(term));
    });
    selectedPointLabel = match?.label ?? options[0]?.label ?? '';
  }

  function pointOptionsFor(technique: Technique | null): TechniquePointOption[] {
    if (!technique) return [];
    const all = catalogue?.pointOptions ?? [];
    if (technique.kind === 'penalty') {
      return all.filter((option) => /card|shido|foul|hansoku|penalty/i.test(`${option.key} ${option.label}`));
    }
    const terms: Record<string, string[]> = {
      strike: ['strike', 'punch', 'kick', 'tsuki', 'geri', 'hiji', 'hiza'],
      kick: ['strike', 'kick', 'geri'],
      throw: ['throw', 'nage', 'kaeshi'],
      takedown: ['takedown', 'taoshi', 'sweep', 'reap', 'hook'],
      sweep: ['takedown', 'sweep', 'harai'],
      reap: ['takedown', 'reap', 'harai'],
      hook: ['takedown', 'strike', 'hook'],
      hold: ['hold', 'pin', 'osae', 'kesa', 'control', 'guard'],
      ground: ['hold', 'ground', 'pin'],
      control: ['hold', 'control', 'guard'],
      break: ['hold', 'break', 'escape'],
      standing: ['strike', 'standing', 'throw'],
      submission: ['submission', 'tap', 'choke', 'lock', 'armbar'],
      ippon: ['ippon'], wazaari: ['waza'], kata: ['kata', 'excellent'],
    };
    const keys = terms[technique.pointKey] ?? [technique.pointKey];
    const matches = all.filter((option) => {
      const label = `${option.key} ${option.label}`.toLowerCase();
      return keys.some((key) => label.includes(key));
    });
    return matches.length ? matches : all;
  }

  function logTechnique() {
    if (!selectedTechnique || !selected) return;
    const label = `${corner.toUpperCase()} · ${selectedTechnique.english} / ${selectedTechnique.romaji}`;
    ask({
      kind: 'event',
      label,
      action: () => api.fightEvent(selected, {
        corner,
        techniqueId: selectedTechnique!.id,
        pointLabel: selectedPointLabel,
        technique: selectedTechnique!.instantWin
          ? `${selectedTechnique!.english} (${selectedTechnique!.romaji}) — instant win`
          : `${selectedTechnique!.english} (${selectedTechnique!.romaji})`,
      }),
    });
  }

  function logCard(technique: Technique | undefined, card: 'yellow' | 'red') {
    if (!selected) return;
    ask({
      kind: 'card',
      label: `${corner.toUpperCase()} · ${card === 'red' ? 'Red card' : 'Yellow card'}`,
      action: async () => {
        const next = await api.fightEvent(selected, {
          corner,
          kind: 'penalty',
          card,
          techniqueId: technique?.id,
          pointLabel: selectedPointLabel,
          technique: technique ? `${technique.english} (${technique.romaji})` : `${card} card`,
          points: card === 'red' ? -2 : -1,
        });
        state = next;
        return next;
      },
    });
  }

  function submitResult(outcome: string) {
    const label = outcome === 'win_red' ? 'Red wins' : outcome === 'win_blue' ? 'Blue wins' : 'Draw';
    ask({ kind: 'result', label: `Result: ${label}`, action: () => api.submitFight(selected, outcome) });
  }

  function callTechnical() {
    if (!selected) return;
    ask({
      kind: 'technical',
      label: `Technical result · ${technicalWinner.toUpperCase()} · ${reasonLabels[technicalReason]}${detail ? ` · ${detail}` : ''}`,
      action: () => api.technicalResult(selected, { winner: technicalWinner, reason: technicalReason, detail }),
    });
  }

  async function scoreKata(id: string, judge: number, section: number, value: number) {
    if (value < 0 || value > 10) return;
    ask({ kind: 'kata', label: `Kata section ${section} · judge ${judge} · ${value}/10`, action: async () => {
      kataPerformances = (await api.kataScore(id, { judge, score: value })).performances;
    } });
  }

  function timerAction(action: 'start' | 'pause' | 'resume') {
    if (!selected) return;
    const run = action === 'start'
      ? () => api.startFight(selected)
      : action === 'pause' ? () => api.pauseFight(selected) : () => api.resumeFight(selected);
    void run().then((next) => {
      state = next;
      notifySuccess(action === 'start' ? 'Clock started' : action === 'pause' ? 'Clock paused' : 'Clock resumed');
      guideAction(action === 'start' ? 'fight-started' : action === 'pause' ? 'timer-paused' : 'timer-resumed');
    }).catch((error) => notifyError(error instanceof Error ? error.message : 'Clock update failed'));
  }

  const clockText = (seconds: number) =>
    `${String(Math.floor(Math.max(0, seconds) / 60)).padStart(2, '0')}:${String(Math.max(0, seconds) % 60).padStart(2, '0')}`;
</script>

<svelte:head><title>J3 · Official console</title></svelte:head>

<div class="shell">
  <header class="topbar">
    <div class="brand"><a class="brand-mark" href="/public" aria-label="Public dashboard">J3</a><div><div class="brand-name">Official console</div><div class="brand-sub">Assigned contests · search every technique · confirm before it counts</div></div></div>
    <div class="tabs">
      <button class="tab" class:active={tab === 'fight'} on:click={() => (tab = 'fight')}>🥊 Fight</button>
      <button class="tab" data-guide="official-tab-kata" class:active={tab === 'kata'} on:click={() => (tab = 'kata')}>🥋 Kata panel</button>
    </div>
  </header>

  {#if tab === 'fight'}
    <div class="console-grid">
      <section class="card contest-card" data-guide="official-contest-select">
        <div class="eyebrow">1 · Assigned contest</div>
        <h2>{selectedFight?.mat ?? 'Mat'} · {selectedFight?.round ?? ''}</h2>
        <select class="tab select-full" bind:value={selected} aria-label="Select contest" on:change={() => guideAction('official-contest-selected')}>
          {#each contests as fight (fight.id)}
            <option value={fight.id}>{fight.mat} · {fight.red.name} vs {fight.blue.name} · {fight.state}</option>
          {/each}
        </select>
        {#if selectedFight}
          <div class="scoreline">
            <span style="color:var(--red-corner);">{selectedFight.red.name}</span>
            <strong>{state?.redScore ?? 0} <span class="muted">—</span> {state?.blueScore ?? 0}</strong>
            <span style="color:var(--blue-corner);">{selectedFight.blue.name}</span>
          </div>
          <div class="score-meta">
            <span>Belt {selectedFight.red.belt} · {selectedFight.red.weightKg} kg</span>
            <span>{selectedFight.blue.weightKg} kg · Belt {selectedFight.blue.belt}</span>
          </div>
        {/if}
        <div class="clock-panel" data-guide="official-clock">
          <div class="clock" class:paused={state && !state.running}>{clockText(state?.seconds ?? 180)}</div>
          <div class="clock-meta">
            <span class="badge" class:live={state?.running}>{state?.running ? '● RUNNING' : 'Ⅱ PAUSED / READY'}</span>
            <span>Round {state?.round ?? 1} · {state?.state ?? 'pending'}</span>
          </div>
          {#if state?.holdSeconds}
            <div class="hold">HOLD · {state.holdCorner.toUpperCase()} · {state.holdSeconds}s</div>
          {/if}
        </div>
        <div class="action-row">
          <button class="btn primary" data-guide="mat-start" on:click={() => timerAction('start')}>▶ Start</button>
          <button class="btn" data-guide="mat-pause" on:click={() => timerAction('pause')}>⏸ Pause</button>
          <button class="btn" on:click={() => timerAction('resume')}>▶ Resume</button>
        </div>
        <div class="muted small" style="margin-top:12px;">
          Clock is authoritative on the venue server. Every start, pause and score is attributed and audited.
        </div>
      </section>

      <section class="card technique-card">
        <div class="eyebrow">2 · Technique search</div>
        <h2>Find a technique, choose points</h2>
        <div class="corner-picker">
          <button class="corner-btn red" data-guide="official-corner-red" class:active={corner === 'red'} on:click={() => { corner = 'red'; guideAction('official-corner'); }}>🔴 RED<br />{selectedFight?.red.name ?? ''}</button>
          <button class="corner-btn blue" class:active={corner === 'blue'} on:click={() => { corner = 'blue'; guideAction('official-corner'); }}>🔵 BLUE<br />{selectedFight?.blue.name ?? ''}</button>
        </div>
        <input class="tab search-input" type="search" placeholder="Search English, romaji, Japanese or type…" bind:value={search} aria-label="Search techniques" />
        <div class="chips">
          <button class="chip" class:active={!activeGroup} on:click={() => (activeGroup = '')}>All</button>
          {#each catalogue?.groups ?? [] as group (group.id)}
            <button class="chip" class:active={activeGroup === group.id} on:click={() => (activeGroup = group.id)}>{group.label}</button>
          {/each}
        </div>
        {#if selectedTechnique}
          <div class="selection">
            <div>
              <strong>{selectedTechnique.english}</strong>
              <span class="muted">{selectedTechnique.romaji}{selectedTechnique.japanese ? ` · ${selectedTechnique.japanese}` : ''}</span>
            </div>
            <label class="point-picker">
              <span>Points ({selectedTechnique.group})</span>
              <select bind:value={selectedPointLabel} aria-label="Configured points">
                {#each pointOptionsFor(selectedTechnique) as option (option.source + option.key + option.label)}
                  <option value={option.label}>{option.label} · {option.value} ({option.source})</option>
                {/each}
              </select>
            </label>
            <button class="btn primary" data-guide="official-tech-button" on:click={logTechnique}>
              Log for {corner.toUpperCase()} {selectedTechnique.instantWin ? '· INSTANT WIN' : ''}
            </button>
          </div>
        {/if}
        <div class="tech-grid">
          {#each filtered as technique (technique.id)}
            <button class="tech" class:selected={selectedTechnique?.id === technique.id} class:penalty={technique.kind === 'penalty'} class:win={technique.instantWin} on:click={() => selectTechnique(technique)}>
              <strong>{technique.english}</strong>
              <span class="muted">{technique.romaji}</span>
              <span class="tech-type">{technique.group.replace(/-/g, ' ')}</span>
              <span class="points">{technique.instantWin ? 'WIN' : technique.kind === 'penalty' ? 'CARD' : `+${technique.resolvedPoints}`}</span>
            </button>
          {:else}
            <div class="muted small">No techniques match “{search}”. Try a romaji spelling or clear the group filter.</div>
          {/each}
        </div>
      </section>

      <section class="card finish-card">
        <div class="eyebrow">3 · Cards & technical results</div>
        <h2>Attribution and outcomes</h2>
        <div class="card-grid" data-guide="official-cards">
          <button class="btn card-btn yellow" on:click={() => logCard(penaltyYellow, 'yellow')}>🟨 Yellow / shido</button>
          <button class="btn card-btn red" on:click={() => logCard(penaltyRed, 'red')}>🟥 Red / hansoku</button>
        </div>
        <div class="technical" data-guide="official-technical">
          <div class="eyebrow">Technical win / loss</div>
          <div class="technical-row">
            <select bind:value={technicalWinner} aria-label="Technical winner">
              <option value="red">Red wins</option>
              <option value="blue">Blue wins</option>
            </select>
            <select bind:value={technicalReason} aria-label="Technical reason">
              {#each Object.entries(reasonLabels) as [value, label]}<option {value}>{label}</option>{/each}
            </select>
          </div>
          <input bind:value={detail} placeholder="Optional detail (e.g. shoulder injury, refused to continue)" aria-label="Technical result detail" />
          <button class="btn warn" on:click={callTechnical}>Call technical result</button>
        </div>

        <div class="eyebrow" style="margin-top:20px;">Confirm final decision</div>
        {#if levelScores}<div class="level-note">Scores are level — confirm a draw or call a technical decision.</div>{/if}
        <div class="action-row">
          <button class="btn" data-guide="official-result-win-red" on:click={() => submitResult('win_red')}>Red wins</button>
          <button class="btn" on:click={() => submitResult('win_blue')}>Blue wins</button>
          <button class="btn" on:click={() => submitResult('draw')}>Draw</button>
        </div>
        <button class="btn quiet" style="margin-top:14px;" on:click={() => api.undoFightEvent(selected).then((next) => { state = next; notifySuccess('Last entry undone'); }).catch((error) => notifyError(error.message))}>↶ Undo last entry</button>
        <div class="muted small" style="margin-top:12px;">Role: {role || 'loading'} · assigned entries only. Results publish to every display after confirmation.</div>
      </section>
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
  .console-grid { display:grid; grid-template-columns:minmax(280px,0.85fr) minmax(340px,1.5fr) minmax(280px,0.95fr); gap:16px; align-items:start; }
  .select-full, .search-input { width:100%; color:var(--brand-ink); }
  .scoreline { display:flex; justify-content:space-between; gap:8px; margin-top:14px; font-size:1.05rem; font-weight:700; }
  .scoreline strong { font-family:var(--font-display); font-size:1.5rem; }
  .score-meta { display:flex; justify-content:space-between; gap:8px; font-size:.72rem; color:var(--brand-muted); }
  .clock-panel { margin-top:16px; padding:16px; border-radius:14px; background:var(--brand-surface-2); border:1px solid var(--brand-border); text-align:center; }
  .clock { font-family:var(--font-display); font-size:clamp(3rem,7vw,5rem); font-weight:800; line-height:1; letter-spacing:.04em; }
  .clock.paused { color:var(--draw); animation:pulse 1s infinite alternate; }
  .clock-meta { display:flex; justify-content:center; gap:10px; flex-wrap:wrap; margin-top:8px; font-size:.76rem; color:var(--brand-muted); }
  .badge { padding:3px 8px; border-radius:999px; background:var(--brand-surface-3); }
  .badge.live { color:#16a34a; font-weight:800; }
  .hold { margin-top:8px; color:#fff; background:var(--brand-accent); border-radius:8px; padding:5px 8px; font-weight:800; font-size:.78rem; }
  .action-row { display:flex; gap:8px; flex-wrap:wrap; margin-top:14px; }
  .corner-picker { display:grid; grid-template-columns:1fr 1fr; gap:10px; }
  .corner-btn { padding:12px; border-radius:12px; border:2px solid transparent; cursor:pointer; background:var(--brand-surface-3); color:var(--brand-ink); font-weight:700; }
  .corner-btn.red.active { border-color:var(--red-corner); }
  .corner-btn.blue.active { border-color:var(--blue-corner); }
  .search-input { margin-top:12px; padding:11px 12px; border-radius:10px; border:1px solid var(--brand-border); background:var(--brand-surface-2); }
  .chips { display:flex; gap:6px; flex-wrap:wrap; margin-top:10px; max-height:76px; overflow:auto; }
  .chip { padding:5px 9px; border-radius:999px; border:1px solid var(--brand-border); background:var(--brand-surface-3); color:var(--brand-muted); font-size:.72rem; cursor:pointer; }
  .chip.active { background:var(--brand-primary); color:#fff; border-color:var(--brand-primary); }
  .selection { display:grid; grid-template-columns:1fr auto; gap:10px; align-items:end; margin-top:12px; padding:12px; border-radius:12px; background:var(--brand-surface-2); border:1px solid var(--brand-primary); }
  .selection > div:first-child { display:grid; gap:2px; }
  .selection .muted { font-size:.78rem; }
  .point-picker { display:grid; gap:4px; font-size:.7rem; color:var(--brand-muted); }
  .point-picker select, .technical select, .technical input { padding:8px; border-radius:8px; border:1px solid var(--brand-border); background:var(--brand-surface-3); color:var(--brand-ink); font:inherit; }
  .selection .btn { grid-column:1 / -1; }
  .tech-grid { display:grid; grid-template-columns:repeat(auto-fill,minmax(150px,1fr)); gap:8px; margin-top:12px; max-height:520px; overflow:auto; padding-right:4px; }
  .tech { display:grid; gap:1px; text-align:left; padding:9px 10px; border-radius:10px; border:1px solid var(--brand-border); background:var(--brand-surface-3); color:var(--brand-ink); cursor:pointer; }
  .tech:hover, .tech.selected { border-color:var(--brand-primary); box-shadow:0 0 0 2px color-mix(in srgb, var(--brand-primary) 25%, transparent); }
  .tech.penalty { border-color:var(--draw); }
  .tech.win { border-color:var(--brand-accent); }
  .tech .points { color:var(--brand-accent); font-weight:800; }
  .tech-type { font-size:.66rem; text-transform:capitalize; color:var(--brand-muted); }
  .card-grid { display:grid; grid-template-columns:1fr 1fr; gap:8px; }
  .card-btn.yellow { border-color:#eab308; }
  .card-btn.red { border-color:#ef4444; }
  .technical { display:grid; gap:8px; margin-top:14px; padding:12px; border-radius:12px; background:var(--brand-surface-2); border:1px solid var(--brand-border); }
  .technical-row { display:grid; grid-template-columns:1fr 1.4fr; gap:8px; }
  .level-note { margin-top:10px; padding:8px 10px; border-radius:8px; background:color-mix(in srgb, var(--draw) 18%, transparent); font-size:.78rem; }
  .btn.quiet { background:transparent; border-color:var(--brand-border); }
  .kata-card { padding:14px; border-radius:12px; background:var(--brand-surface-3); border:1px solid var(--brand-border); }
  .kata-grid { display:grid; grid-template-columns:repeat(auto-fit,minmax(120px,1fr)); gap:8px; margin-top:10px; }
  .kata-grid input { width:100%; padding:7px; border-radius:8px; background:var(--brand-surface-2); color:var(--brand-ink); border:1px solid var(--brand-border); }
  @keyframes pulse { from { opacity:1; } to { opacity:.5; } }
  @media (max-width:1100px) { .console-grid { grid-template-columns:1fr 1fr; } .technique-card { grid-column:1 / -1; grid-row:2; } }
  @media (max-width:720px) { .console-grid { grid-template-columns:1fr; } .technique-card { grid-column:auto; grid-row:auto; } .selection { grid-template-columns:1fr; } }
</style>
