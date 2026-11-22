<script lang="ts">
  import { onMount } from 'svelte';
  import {
    api, livePoll, broadcastPA,
    type DemoFight, type FightState, type StandingRow, type AuditEntry,
    type KataPerformance, type PoolView, type BracketStage,
  } from '../api';

  export let type = 'clock';
  export let title = '';
  export let span: 1 | 2 | 3 = 1;

  let fights: DemoFight[] = [];
  let state: FightState | null = null;
  let standings: StandingRow[] = [];
  let medals: { id: string; fighterId: string; name: string; club: string; clubColour: string; kind: string }[] = [];
  let audit: AuditEntry[] = [];
  let kata: KataPerformance[] = [];
  let pools: PoolView[] = [];
  let stages: BracketStage[] = [];
  let stats: { topMoves: Record<string, number>; clubMedals: Record<string, number> } | null = null;
  let eventName = '';
  let now = new Date();

  const labels: Record<string, string> = {
    scoreboard: '🥊 Live scoreboard', upcoming: '📋 Up next', standings: '🏆 Standings',
    brackets: '🌳 Bracket', medals: '🥇 Medals', charts: '📊 Charts', clock: '🕒 Clock',
    kata: '🥋 Kata panel', pools: '🧩 Fight pools', pa: '📣 Announcer', details: '📋 Event', audit: '📜 Audit',
  };

  onMount(() => {
    // Only subscribe to the data this widget actually renders.
    const stops: (() => void)[] = [];
    if (type === 'scoreboard' || type === 'upcoming') stops.push(livePoll(api.fights, 3000, (data) => (fights = data.fights)));
    if (type === 'standings') stops.push(livePoll(api.standings, 5000, (data) => (standings = data.standings)));
    if (type === 'medals') stops.push(livePoll(api.medals, 5000, (data) => (medals = data.medals)));
    if (type === 'kata') stops.push(livePoll(api.kata, 5000, (data) => (kata = data.performances)));
    if (type === 'pools') stops.push(livePoll(api.pools, 5000, (data) => (pools = data.pools)));
    if (type === 'brackets') stops.push(livePoll(api.bracketStages, 5000, (data) => (stages = data.stages)));
    if (type === 'charts') stops.push(livePoll(api.stats, 8000, (data) => (stats = data)));
    if (type === 'audit') stops.push(livePoll(api.audit, 8000, (data) => (audit = data.entries)));
    if (type === 'details') stops.push(livePoll(api.event, 10000, (data) => (eventName = data.name)));
    const stopStream = type === 'scoreboard'
      ? api.eventStream((message) => { if (message.payload) state = message.payload; })
      : () => {};
    const clock = type === 'clock' ? setInterval(() => (now = new Date()), 1000) : null;
    return () => { stops.forEach((stop) => stop()); stopStream(); if (clock) clearInterval(clock); };
  });

  $: liveFight = fights.find((fight) => fight.state === 'live') ?? fights[0];
  $: upcoming = fights.filter((fight) => fight.state === 'pending').slice(0, 5);
  $: heading = title || labels[type] || type;
  $: moveRows = stats ? Object.entries(stats.topMoves).sort((a, b) => b[1] - a[1]).slice(0, 4) : [];
  $: clubRows = stats ? Object.entries(stats.clubMedals).sort((a, b) => b[1] - a[1]).slice(0, 4) : [];
</script>

<section class="widget" data-widget-type={type} class:wide={span === 2} class:extra={span === 3}>
  <div class="eyebrow">{heading}</div>

  {#if type === 'clock'}
    <div class="big-clock">{now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })}</div>
    <div class="muted small">{now.toLocaleDateString([], { weekday: 'long', day: 'numeric', month: 'long' })}</div>
  {:else if type === 'scoreboard'}
    {#if liveFight}
      <div class="score-line"><strong>{liveFight.red.name}</strong><span class="score">{state?.redScore ?? 0}</span><span class="muted">vs</span><span class="score">{state?.blueScore ?? 0}</span><strong>{liveFight.blue.name}</strong></div>
      <div class="muted small">{liveFight.mat} · {liveFight.category} · {state?.running ? 'live' : 'paused'}</div>
    {:else}<span class="muted small">No contests yet.</span>{/if}
  {:else if type === 'upcoming'}
    {#each upcoming as fight (fight.id)}<div class="row"><strong>{fight.red.name}</strong> vs <strong>{fight.blue.name}</strong><span class="muted small">{fight.mat}</span></div>{/each}
  {:else if type === 'standings'}
    {#each standings.slice(0, 5) as row, i (row.fighterId)}<div class="row"><span>{i + 1}. <strong>{row.name}</strong></span><span class="muted small">{row.points} pts</span></div>{/each}
  {:else if type === 'brackets'}
    <div class="chips">{#each stages as stage (stage.id)}<span class="chip">{stage.name}</span>{/each}</div>
    <div class="muted small">{stages.reduce((sum, stage) => sum + stage.slots.length, 0)} slots across {stages.length} rounds</div>
  {:else if type === 'medals'}
    <div class="chips">
      {#each medals as medal (medal.id)}<span class="medal-dot" style="background:{medal.clubColour}" title={medal.name}>{medal.kind === 'gold' ? '🥇' : medal.kind === 'silver' ? '🥈' : '🥉'}</span>{/each}
    </div>
    <div class="muted small">{medals.length} medals awarded</div>
  {:else if type === 'charts'}
    {#each moveRows as [name, value] (name)}<div class="bar-row"><span>{name}</span><div class="bar"><div class="bar-fill" style="width:{Math.min(100, value * 8)}%"></div></div><strong>{value}</strong></div>{/each}
    <div class="muted small" style="margin-top:8px;">Clubs: {clubRows.map(([name, value]) => `${name} ${value}`).join(' · ')}</div>
  {:else if type === 'kata'}
    {#each kata.slice(0, 5) as performance (performance.id)}<div class="row"><strong>{performance.name}</strong><span class="muted small">{performance.kataName} · {performance.total.toFixed(2)}</span></div>{/each}
  {:else if type === 'pools'}
    <div class="chips">{#each pools as pool (pool.id)}<span class="chip">{pool.name} · {pool.members.length}</span>{/each}</div>
  {:else if type === 'pa'}
    <div class="action-row">
      <button class="btn" on:click={() => broadcastPA('Competitors to the mat — prepare to fight')}>Call to mat</button>
      <button class="btn" on:click={() => broadcastPA('Medal ceremony at the podium')}>Podium call</button>
    </div>
  {:else if type === 'details'}
    <strong>{eventName || 'Tournament'}</strong>
    <div class="muted small">Rules, requirements and venue details are published on the details screen.</div>
    <a class="btn" href="/details">Open details</a>
  {:else if type === 'audit'}
    {#each audit.slice(0, 5) as entry (entry.id)}<div class="row"><span class="muted small">{entry.action}</span><span class="muted small">{entry.actorRole}</span></div>{/each}
  {/if}
</section>

<style>
  .widget { padding:16px; border-radius:14px; background:var(--brand-surface-2); border:1px solid var(--brand-border); box-shadow:var(--brand-shadow); min-height:150px; }
  .widget.wide { grid-column:span 2; }
  .widget.extra { grid-column:span 3; }
  .big-clock { font-family:var(--font-display); font-size:clamp(2.4rem,5vw,4rem); font-weight:800; color:var(--brand-accent); line-height:1.1; }
  .score-line { display:flex; align-items:center; justify-content:space-between; gap:10px; flex-wrap:wrap; margin-bottom:8px; }
  .score { font-family:var(--font-display); font-size:2rem; font-weight:800; color:var(--brand-primary); }
  .row { display:flex; justify-content:space-between; gap:10px; padding:6px 0; border-top:1px solid var(--brand-border); font-size:.88rem; }
  .chips { display:flex; gap:6px; flex-wrap:wrap; margin-bottom:8px; }
  .chip { padding:5px 9px; border-radius:999px; background:var(--brand-surface-3); font-size:.74rem; }
  .medal-dot { width:34px; height:34px; border-radius:50%; display:grid; place-items:center; }
  .bar-row { display:grid; grid-template-columns:90px 1fr 28px; align-items:center; gap:8px; font-size:.8rem; margin-top:6px; }
  .bar { height:10px; border-radius:999px; background:var(--brand-surface-3); overflow:hidden; }
  .bar-fill { height:100%; background:var(--brand-primary); border-radius:999px; }
</style>
