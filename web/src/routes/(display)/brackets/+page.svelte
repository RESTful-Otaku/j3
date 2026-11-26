<script lang="ts">
  import { onMount } from 'svelte';
  import { initials } from '../../../lib/format';
  import { api, livePoll, type DemoFight, type PoolView, type BracketStage, broadcastPA } from '../../../lib/api';
  import { notifySuccess, notifyError } from '../../../lib/toast';
  import { phoneLimit, phoneRows } from '../../../lib/responsive';
  import ConfirmDialog from '../../../lib/components/ConfirmDialog.svelte';
  import FighterPortrait from '../../../lib/components/FighterPortrait.svelte';
  import { onLiveChange } from '../../../lib/live-updates';
  import { cornerConfig } from '../../../lib/event-brand';

  type Slot = { id: string; name: string; colour: string; avatarPath: string; clubLogo: string };
  type BracketNode = { id: string; label: string; a: Slot | null; b: Slot | null; winner: string };

  let fights: DemoFight[] = [];
  let pools: PoolView[] = [];
  let bracket: BracketNode[][] = [];
  let mode: 'official' | 'elimination' | 'roundRobin' = 'official';
  let serverStages: BracketStage[] = [];
  let pendingPromotion: DemoFight | null = null;
  let confirmOpen = false;
  let buildConfirmOpen = false;
  let directBuild = false;
  let dragging: { round: number; node: number; corner: 'a' | 'b' } | null = null;
  let bracketContainer: HTMLDivElement;
  let connectorPaths: string[] = [];

  function matchesOf(stage: BracketStage) {
    const result: { id: string; slots: typeof stage.slots }[] = [];
    for (let i = 0; i < stage.slots.length; i += 2) {
      result.push({ id: stage.slots[i].id, slots: stage.slots.slice(i, i + 2) });
    }
    return result;
  }

  function drawConnectors() {
    if (!bracketContainer) return;
    const containerRect = bracketContainer.getBoundingClientRect();
    const columns = Array.from(bracketContainer.querySelectorAll('.bracket-column'));
    const paths: string[] = [];
    for (let i = 0; i < columns.length - 1; i++) {
      const from = Array.from(columns[i].querySelectorAll('.bracket-match'));
      const to = Array.from(columns[i + 1].querySelectorAll('.bracket-match'));
      from.forEach((element, index) => {
        const target = to[Math.floor(index / 2)];
        if (!target) return;
        const a = element.getBoundingClientRect();
        const b = target.getBoundingClientRect();
        const x1 = a.right - containerRect.left;
        const y1 = a.top - containerRect.top + a.height / 2;
        const x2 = b.left - containerRect.left;
        const y2 = b.top - containerRect.top + b.height / 2;
        const mid = x1 + (x2 - x1) / 2;
        paths.push(`M ${x1} ${y1} H ${mid} V ${y2} H ${x2}`);
      });
    }
    connectorPaths = paths;
  }

  function scheduleConnectors() {
    requestAnimationFrame(() => requestAnimationFrame(drawConnectors));
  }

  onMount(() => {
    const loadStages = () => api.bracketStages().then((result) => (serverStages = result.stages)).catch(() => {});
    const loadFights = () => api.fights().then((data) => {
      fights = data.fights;
      if (bracket.length > 0) buildBracket();
    }).catch(() => {});
    void loadStages();
    const stopFights = livePoll(api.fights, 2000, (data) => {
      fights = data.fights;
      if (bracket.length > 0) buildBracket();
    });
    const stopPools = livePoll(api.pools, 4000, (data) => {
      pools = data.pools;
      if (bracket.length === 0) buildBracket();
    });
    const stopLive = onLiveChange(['brackets', 'fights'], () => { void loadStages(); void loadFights(); });
    window.addEventListener('resize', drawConnectors);
    return () => { stopFights(); stopPools(); stopLive(); window.removeEventListener('resize', drawConnectors); };
  });

  async function buildOfficial(direct = false) {
    try {
      const result = await api.buildBracket({
        eventId: '', format: direct ? 'Single elimination' : 'Double elimination',
        targetPoolSize: 4, maxPools: 12, seed: 42, thirdPlace: true, direct,
      });
      serverStages = result.stages;
      notifySuccess(`Official bracket built · ${result.format} · ${serverStages.length} stages`);
    } catch (error) {
      notifyError(error instanceof Error ? error.message : 'Could not build bracket');
    }
  }
  function askBuild(direct: boolean) {
    directBuild = direct;
    buildConfirmOpen = true;
  }
  async function promoteOfficial(id: string, icon: string) {
    try {
      const result = await api.promoteSlot(id, { winner: icon.startsWith('W'), resultIcon: icon });
      serverStages = result.stages;
      notifySuccess(`Result ${icon} recorded with history retained`);
    } catch (error) {
      notifyError(error instanceof Error ? error.message : 'Could not record result');
    }
  }
  function toSlot(member: { id: string; name: string; clubColour: string; avatarPath: string; clubLogo: string }): Slot {
    return { id: member.id, name: member.name, colour: member.clubColour, avatarPath: member.avatarPath, clubLogo: member.clubLogo };
  }
  function winnerFor(a: Slot | null, b: Slot | null) {
    const fight = fights.find((f) =>
      (f.red.id === a?.id && f.blue.id === b?.id) || (f.red.id === b?.id && f.blue.id === a?.id));
    if (fight?.outcome === 'win_red') return a?.id ?? '';
    if (fight?.outcome === 'win_blue') return b?.id ?? '';
    if (fight?.outcome === 'draw') return 'draw';
    return a?.id ?? b?.id ?? '';
  }
  function node(id: string, label: string, a: Slot | null, b: Slot | null): BracketNode {
    return { id, label, a, b, winner: winnerFor(a, b) };
  }
  function buildBracket() {
    const first: BracketNode[] = [];
    pools.forEach((pool, poolIndex) => {
      for (let i = 0; i + 1 < pool.members.length; i += 2) {
        first.push(node(`r0-${poolIndex}-${i}`, `${pool.name} · ${i / 2 + 1}`,
          toSlot(pool.members[i]), toSlot(pool.members[i + 1])));
      }
    });
    const advance = (sources: BracketNode[]) => sources.map((source) =>
      source.winner === source.a?.id ? source.a : source.b);
    const semiWinners = advance(first);
    const semis: BracketNode[] = [];
    for (let i = 0; i + 1 < semiWinners.length && semis.length < 2; i += 2) {
      semis.push(node(`r1-${i}`, `Semi ${i / 2 + 1}`, semiWinners[i], semiWinners[i + 1]));
    }
    const finalists = advance(semis);
    const losers = semis.map((s) => s.winner === s.a?.id ? s.b : s.a);
    const last: BracketNode[] = [
      node('final', 'Final', finalists[0] ?? null, finalists[1] ?? null),
      node('third', 'Bronze decider', losers[0] ?? null, losers[1] ?? null),
    ];
    bracket = [first, semis, last];
  }
  function dropSlot(round: number, nodeIndex: number, corner: 'a' | 'b') {
    if (!dragging) return;
    const source = bracket[dragging.round]?.[dragging.node];
    const target = bracket[round]?.[nodeIndex];
    if (!source || !target) return;
    const from = dragging.corner === 'a' ? source.a : source.b;
    const to = corner === 'a' ? target.a : target.b;
    if (dragging.round === round && dragging.node === nodeIndex && dragging.corner === corner) return;
    if (dragging.corner === 'a') source.a = to; else source.b = to;
    if (corner === 'a') target.a = from; else target.b = from;
    source.winner = winnerFor(source.a, source.b);
    target.winner = winnerFor(target.a, target.b);
    bracket = [...bracket];
    dragging = null;
    notifySuccess('Bracket slot updated', () => buildBracket());
  }
  function outcome(fight: DemoFight, corner: 'red' | 'blue') {
    if (!fight.outcome) return { code: '', win: false };
    if (fight.outcome === 'draw') return { code: 'DP', win: false };
    const winner = fight.outcome === 'win_red' ? 'red' : 'blue';
    const win = corner === winner;
    const submission = fight.events.some((e) => e.kind === 'submission');
    const technical = fight.outcome.startsWith('dq') || fight.state === 'forfeit';
    return { code: `${win ? 'W' : 'L'}${submission ? 'S' : technical ? 'T' : 'P'}`, win };
  }
  function cards(fight: DemoFight, corner: 'red' | 'blue') {
    return fight.events.filter((e) => e.kind === 'penalty' && String((e.payload as Record<string, unknown>)?.corner) === corner)
      .map((e) => String((e.payload as Record<string, unknown>)?.card ?? ''));
  }
  function roundGroups() {
    const groups: Record<string, DemoFight[]> = {};
    for (const fight of fights) {
      const key = fight.round || 'Round 1';
      (groups[key] ??= []).push(fight);
    }
    return Object.entries(groups).sort(([a], [b]) => a.localeCompare(b));
  }
  function promote(fight: DemoFight) {
    pendingPromotion = fight;
    confirmOpen = true;
  }
  function winnerOf(fight: DemoFight) {
    if (fight.outcome === 'draw' || !fight.outcome) return null;
    return fight.outcome === 'win_blue' ? fight.blue : fight.red;
  }
  function confirmPromotion() {
    const winner = pendingPromotion ? winnerOf(pendingPromotion) : null;
    if (pendingPromotion && winner) {
      broadcastPA(`${winner.name} advances from ${pendingPromotion.round}`, winner.id, 5);
    }
  }

  $: groups = roundGroups();
  $: if (serverStages.length > 0) scheduleConnectors();
</script>

<svelte:head><title>J3 · Brackets</title></svelte:head>

<div class="display-shell">
  <header class="display-header" data-guide="brackets-header">
    <div><div class="event-mark"><span class="crest">J3</span>Brackets</div><div class="meta">Progression, results and history · live</div></div>
    <div class="tabs">
      <button class="tab" class:active={mode === 'official'} on:click={() => (mode = 'official')}>Official bracket</button>
      <button class="tab" class:active={mode === 'elimination'} on:click={() => (mode = 'elimination')}>Elimination bracket</button>
      <button class="tab" class:active={mode === 'roundRobin'} on:click={() => (mode = 'roundRobin')}>Round robin table</button>
    </div>
  </header>

  {#if mode === 'official'}
    <div style="display:flex; justify-content:space-between; gap:12px; flex-wrap:wrap; align-items:center; margin-bottom:14px;">
      <p class="muted small">Server-generated stages · results and promotion retained in every round</p>
      <button class="btn" on:click={() => askBuild(true)}>⚡ Direct elimination (BYEs)</button>
      <button class="btn primary" on:click={() => askBuild(false)}>⚙ Build official bracket</button>
    </div>
    <div class="bracket-wrap" bind:this={bracketContainer}>
      <svg class="bracket-lines" aria-hidden="true">
        {#each connectorPaths as path (path)}<path d={path} />{/each}
      </svg>
      {#each serverStages as stage (stage.id)}
        <section class="bracket-column">
          <div class="eyebrow">{stage.name} · {stage.kind}</div>
          <div class="bracket-matches">
            {#each matchesOf(stage) as match (match.id)}
              <div class="bracket-match">
                {#each match.slots as slot (slot.id)}
                  <div class="bracket-slot" class:advanced={slot.state === 'advanced'} class:out={slot.state === 'eliminated'}>
                    <FighterPortrait
                      src={slot.avatarPath}
                      label={slot.fighterName || 'TBD'}
                      size="42px"
                      fallback={initials(slot.fighterName || 'TBD')}
                      clubColour={slot.clubColour || '#64748b'}
                    />
                    <span style="flex:1; min-width:0;">
                      <strong>{slot.fighterName || 'TBD'}</strong>
                      <span class="muted small" style="display:block; white-space:nowrap; overflow:hidden; text-overflow:ellipsis;">{slot.club || 'Independent'} · {slot.belt}</span>
                    </span>
                    {#if slot.resultIcon}<span class="badge {slot.state === 'advanced' ? 'fighter' : 'spectator'}">{slot.resultIcon}</span>{/if}
                    {#if slot.fighterId}
                      <span class="slot-actions">
                        <button class="tab" title="Win on points" on:click={() => promoteOfficial(slot.id, 'WP')}>W</button>
                        <button class="tab" title="Win by submission" on:click={() => promoteOfficial(slot.id, 'WS')}>S</button>
                        <button class="tab" title="Loss" on:click={() => promoteOfficial(slot.id, 'LP')}>L</button>
                      </span>
                    {/if}
                  </div>
                {/each}
              </div>
            {/each}
          </div>
        </section>
      {:else}
        <div class="placeholder">No official bracket yet — press “Build official bracket”.</div>
      {/each}
    </div>
  {:else if mode === 'elimination'}
    <div class="bracket-rounds" style="display:flex; gap:clamp(14px,2vw,40px); padding-bottom:20px; align-items:flex-start;" use:phoneLimit={3}>
      {#each groups as [round, roundFights] (round)}
        <section style="min-width:clamp(280px,24vw,420px); flex:1;">
          <div class="eyebrow">{round}</div>
          <div style="display:grid; gap:14px; margin-top:12px;">
            {#each roundFights as fight (fight.id)}
              <div class="card slide-in" style="padding:14px;">
                <div class="small muted">{fight.mat} · {fight.category}</div>
                {#each [{ name: fight.red.name, avatarPath: fight.red.avatarPath, clubColour: fight.red.clubColour, id: fight.red.id }, { name: fight.blue.name, avatarPath: fight.blue.avatarPath, clubColour: fight.blue.clubColour, id: fight.blue.id }] as fighter, i}
                  {@const corner = i === 0 ? 'red' : 'blue'}
                  {@const result = outcome(fight, corner)}
                  <div class="match-row" style="display:flex; align-items:center; gap:10px; padding:8px 0; border-top:1px solid var(--brand-border);">
                    <FighterPortrait
                      src={fighter.avatarPath}
                      label={fighter.name}
                      size="42px"
                      fallback={initials(fighter.name)}
                      clubColour={fighter.clubColour}
                    />
                    <span class="match-name" style="flex:1; font-weight:600;">{fighter.name || 'TBD'}</span>
                    {#each cards(fight, corner) as card}<span class="badge {card === 'red' ? 'official' : 'runner'}">{card === 'red' ? '🟥' : '🟨'}</span>{/each}
                    {#if result.code}<span class="badge {result.win ? 'fighter' : 'spectator'}">{result.code}</span>{/if}
                  </div>
                {/each}
                {#if fight.state !== 'complete'}
                  <button class="btn" style="width:100%; margin-top:8px;" on:click={() => promote(fight)}>Announce winner →</button>
                {:else}
                  <div class="small muted" style="margin-top:8px;">Result committed · history retained at this round</div>
                {/if}
              </div>
            {/each}
          </div>
        </section>
      {:else}
        <div class="placeholder">No brackets generated yet.</div>
      {/each}
    </div>
  {:else}
    <div class="card">
      <h2>Round robin — every pairing and its result</h2>
      <table style="width:100%; border-collapse:collapse; margin-top:12px;">
        <thead><tr class="muted small" style="text-align:left;"><th style="padding:8px;">Round</th><th>Mat</th><th>{$cornerConfig.red.label}</th><th>{$cornerConfig.blue.label}</th><th>Result</th><th>State</th></tr></thead>
        <tbody use:phoneRows={10}>
          {#each fights as fight (fight.id)}
            <tr style="border-top:1px solid var(--brand-border);">
              <td style="padding:8px;">{fight.round}</td><td>{fight.mat}</td>
              <td>{fight.red.name} <span class="badge fighter">{outcome(fight, 'red').code}</span></td>
              <td>{fight.blue.name} <span class="badge spectator">{outcome(fight, 'blue').code}</span></td>
              <td>{fight.outcome || '—'}</td><td>{fight.state}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<ConfirmDialog
  bind:open={confirmOpen}
  title="Announce winner?"
  message="This calls the winner to the next area. Use the Official bracket tab to record a result and promote a fighter."
  confirmLabel="Announce"
  onConfirm={confirmPromotion}
/>

<ConfirmDialog
  bind:open={buildConfirmOpen}
  title="Rebuild the official bracket?"
  message="Existing stage slots and promotion history are replaced. Confirmed fight results stay in the audit trail."
  confirmLabel="Rebuild bracket"
  danger={true}
  onConfirm={() => void buildOfficial(directBuild)}
/>

<style>
  .bracket-wrap { position:relative; display:flex; gap:clamp(18px,2vw,44px); align-items:stretch; overflow-x:auto; padding:12px 0 32px; min-height:min(68vh, 720px); min-width:max-content; }
  .bracket-lines { position:absolute; inset:0; width:100%; height:100%; pointer-events:none; z-index:0; }
  .bracket-lines path { fill:none; stroke:var(--brand-accent); stroke-width:2; opacity:.7; }
  .bracket-column { min-width:clamp(230px,19vw,330px); flex:1; position:relative; z-index:1; display:flex; flex-direction:column; }
  .bracket-matches { display:flex; flex-direction:column; justify-content:space-around; flex:1; gap:12px; margin-top:12px; }
  .bracket-match { background:var(--brand-surface-2); border:1px solid var(--brand-border); border-radius:12px; padding:4px; box-shadow:var(--brand-shadow); }
  .bracket-slot { display:flex; align-items:center; gap:8px; padding:8px; border-radius:8px; }
  .bracket-slot + .bracket-slot { border-top:1px solid var(--brand-border); }
  .bracket-slot.advanced { background:rgba(34,197,94,.12); }
  .bracket-slot.out { opacity:.55; }
  .slot-actions { display:flex; gap:3px; opacity:0; transition:opacity .15s ease; }
  .bracket-slot:hover .slot-actions, .bracket-slot:focus-within .slot-actions { opacity:1; }
  @media (hover: none) { .slot-actions { opacity:1; } }
  .slot-actions .tab { padding:4px 7px; font-size:.72rem; }
  .match-name { min-width:0; overflow-wrap:anywhere; }
  @media (max-width: 700px) {
    .bracket-wrap { display:grid; grid-template-columns:1fr; min-width:0; overflow:visible; padding-bottom:8px; }
    .bracket-column { min-width:0; width:100%; }
    .bracket-lines { display:none; }
    .bracket-matches { justify-content:flex-start; }
    .bracket-rounds { display:grid !important; grid-template-columns:1fr; }
    .bracket-rounds > section { min-width:0; width:100%; }
    /* Touch controls move to their own line so fighter names keep the width. */
    .bracket-slot { flex-wrap:wrap; gap:6px; padding:7px; }
    .bracket-slot > span:nth-child(2) { flex:1 1 calc(100% - 48px); }
    .slot-actions { flex:1 1 100%; justify-content:flex-end; margin-top:2px; }
    .slot-actions .tab { min-width:42px; padding:5px 8px; font-size:.76rem; }
    .match-row { gap:8px; padding:7px 0; }
    .match-row .badge { flex:none; padding:2px 6px; }
  }
</style>
