<script lang="ts">
  import { onMount } from 'svelte';
  import { initials } from '../../../lib/format';
  import { session, uiRole } from '../../../lib/auth';
  import { phoneLimit, phoneRows } from '../../../lib/responsive';
  import { api, livePoll, type DemoFight, type EventDisplayInfo, type StandingRow, type ScheduleItem } from '../../../lib/api';
  import { eventTitle } from '../../../lib/event-title';
  import { replaySnapshot } from '../../../lib/demo-replay';

  let event: EventDisplayInfo | null = null;
  let fights: DemoFight[] = [];
  let standings: StandingRow[] = [];
  let schedule: ScheduleItem[] = [];
  let now = new Date();
  let canControlSchedule = false;
  let phoneSection: 'live' | 'next' | 'table' | 'results' = 'live';

  function syncRole() {
    const active = session();
    canControlSchedule = !!active && ['runner', 'official'].includes(uiRole(active.role));
  }

  onMount(() => {
    syncRole();
    window.addEventListener('j3-session-changed', syncRole);
    const stopEvent = livePoll(api.event, 10000, (data) => (event = data));
    const stopFights = livePoll(api.fights, 2000, (data) => (fights = data.fights));
    const stopStandings = livePoll(api.standings, 5000, (data) => (standings = data.standings));
    const stopSchedule = livePoll(api.schedule, 5000, (data) => (schedule = data.items));
    const clock = setInterval(() => (now = new Date()), 1000);
    return () => {
      stopEvent(); stopFights(); stopStandings(); stopSchedule(); clearInterval(clock);
      window.removeEventListener('j3-session-changed', syncRole);
    };
  });

  $: nextUp = schedule.find((item) => item.status === 'upcoming');
  $: offset = schedule[0]?.offsetMinutes ?? 0;
  $: live = fights.filter((f) => f.state === 'live');
  $: upcoming = fights.filter((f) => f.state === 'pending').slice(0, 12);
  $: finished = fights.filter((f) => f.state === 'complete');
  $: leaders = standings.slice(0, 10);
  $: replay = $replaySnapshot;
  async function drift(minutes: number) {
    const result = await api.setDrift(minutes);
    schedule = result.items;
  }
  async function complete(id: string) {
    const result = await api.completeScheduleItem(id);
    schedule = result.items;
  }
  function formatTime(value: string) {
    return new Date(value).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }
</script>

<svelte:head><title>{$eventTitle} · Public dashboard</title></svelte:head>

<div class="display-shell">
  <header class="display-header" data-guide="public-header">
    <div class="event-mark"><span class="crest">J3</span><span>{event?.name ?? 'Tournament'}</span></div>
    <div class="meta">Live now · {live.length} contest{live.length === 1 ? '' : 's'} · {fights.length} total · updates automatically</div>
  </header>

  <div class="brand-banner" aria-label="J3 banner"></div>
  {#if replay.active}
    <div class="replay-strip" aria-live="polite">
      <span class="eyebrow">Demo replay · {replay.index + 1}/{replay.total}</span>
      {#if replay.bout}
        <strong>{replay.bout.redId} {replay.bout.pointsRed} — {replay.bout.pointsBlue} {replay.bout.blueId}</strong>
        <span class="muted small">{replay.bout.stage} · {replay.bout.round} · {replay.bout.outcome.replaceAll('_', ' ')}</span>
      {:else}<strong>Tournament ready</strong>{/if}
    </div>
  {/if}
  <div class="clock-hero slide-in">
    <div>
      <div class="eyebrow">Venue time</div>
      <div class="big-clock">{now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })}</div>
    </div>
    <div style="flex:1; min-width:240px;">
      <div class="eyebrow">Next up</div>
      <div style="font-family:var(--font-display); font-size:clamp(1.2rem,2.4vw,2.4rem); font-weight:800;">
        {nextUp?.name ?? 'Schedule complete'}
      </div>
      <div class="muted">{nextUp ? `${nextUp.area} · projected ${formatTime(nextUp.projectedStart)}–${formatTime(nextUp.projectedEnd)}` : 'Well done everyone'}</div>
    </div>
    <div style="display:grid; gap:8px;">
      <span class="pill">{offset === 0 ? 'On schedule' : `${offset > 0 ? '+' : ''}${offset} min drift`}</span>
      {#if canControlSchedule}
        <div style="display:flex; gap:6px;">
          <button class="btn" on:click={() => drift(Math.max(-480, offset - 5))}>−5 min</button>
          <button class="btn" on:click={() => drift(0)}>Reset</button>
          <button class="btn" on:click={() => drift(Math.min(480, offset + 5))}>+5 min</button>
        </div>
      {:else}
        <span class="muted small">Projected times update automatically</span>
      {/if}
    </div>
  </div>

  <div class="phone-sections" aria-label="Dashboard sections">
    {#each [['live','Live'],['next','Next up'],['table','Table'],['results','Results']] as [id, label]}
      <button class="tab" class:active={phoneSection === id} on:click={() => (phoneSection = id as typeof phoneSection)}>{label}</button>
    {/each}
  </div>

  <section class="phone-section" class:active={phoneSection === 'live'}>
  <div class="section-title"><h2>Today's timeline</h2></div>
  <div class="card">
    <table style="width:100%; border-collapse:collapse; font-size:clamp(.9rem,1.2vw,1.3rem);">
      <thead><tr class="muted small" style="text-align:left;"><th style="padding:8px;">Projected</th><th>Event</th><th>Area</th><th>Status</th><th></th></tr></thead>
      <tbody use:phoneRows={5}>
        {#each schedule as item (item.id)}
          <tr style="border-top:1px solid var(--brand-border);" class:flash={item.id === nextUp?.id}>
            <td style="padding:8px; font-weight:700;">{formatTime(item.projectedStart)}–{formatTime(item.projectedEnd)}</td>
            <td>{item.name}</td><td>{item.area}</td>
            <td><span class="badge {item.status === 'complete' ? 'spectator' : 'fighter'}">{item.status}</span></td>
            <td>{#if canControlSchedule && item.status !== 'complete'}<button class="tab" on:click={() => complete(item.id)}>Mark done</button>{/if}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>

  <div class="section-title"><h2>On the mats now</h2><span class="pill"><span class="dot"></span>Live</span></div>
  <div class="grid" use:phoneLimit={3} style="grid-template-columns: repeat(auto-fit, minmax(min(100%, 420px), 1fr));">
    {#each live as fight (fight.id)}
      <div class="card slide-in" style="border-color:var(--brand-primary);">
        <div class="meta">{fight.mat} · {fight.category} · {fight.round}</div>
        <div class="live-duel" style="display:flex; justify-content:space-between; align-items:center; gap:18px; margin-top:12px;">
          <div style="display:flex; align-items:center; gap:12px;">
            <span class="portrait small-portrait" style="background:{fight.red.clubColour};">{initials(fight.red.name)}</span>
            <strong style="font-size:1.2rem;">{fight.red.name}</strong>
          </div>
          <span class="vs" style="font-family:var(--font-display); font-size:2rem;">vs</span>
          <div style="display:flex; align-items:center; gap:12px;">
            <strong style="font-size:1.2rem;">{fight.blue.name}</strong>
            <span class="portrait small-portrait" style="background:{fight.blue.clubColour};">{initials(fight.blue.name)}</span>
          </div>
        </div>
      </div>
    {:else}
      <div class="placeholder">No live contests at this moment.</div>
    {/each}
  </div>

  </section>

  <section class="phone-section" class:active={phoneSection === 'next'}>
  <div class="section-title"><h2>Next up — report to your mat</h2><span class="muted small">{upcoming.length} upcoming</span></div>
  <div class="grid" use:phoneLimit={4} style="grid-template-columns: repeat(auto-fit, minmax(min(100%, 300px), 1fr));">
    {#each upcoming as fight (fight.id)}
      <div class="stat flash">
        <div class="eyebrow">{fight.mat} · {fight.category}</div>
        <div style="font-size:1.1rem; font-weight:700;">{fight.red.name} <span class="muted">vs</span> {fight.blue.name}</div>
        <div class="small muted">{fight.round}</div>
      </div>
    {:else}
      <div class="placeholder">Schedule is clear.</div>
    {/each}
  </div>

  </section>

  <section class="phone-section" class:active={phoneSection === 'table'}>
  <div class="section-title"><h2>Leaderboard</h2><span class="muted small">Points: 3 per win, 1 per draw</span></div>
  <div class="card">
    <table style="width:100%; border-collapse:collapse; font-size:clamp(.9rem,1.2vw,1.3rem);">
      <thead><tr class="muted small" style="text-align:left;"><th style="padding:8px;">#</th><th>Fighter</th><th>Category</th><th>Club</th><th>W</th><th>D</th><th>L</th><th>Pts</th></tr></thead>
      <tbody use:phoneRows={5}>
        {#each leaders as row, i (row.fighterId)}
          <tr style="border-top:1px solid var(--brand-border);">
            <td style="padding:8px;">{i + 1}</td>
            <td><strong>{row.name}</strong> <span class="badge spectator">{row.belt}</span></td>
            <td>{row.category}</td>
            <td><span style="color:{row.clubColour};">●</span> {row.club}</td>
            <td style="color:var(--win);">{row.wins}</td><td style="color:var(--draw);">{row.draws}</td>
            <td style="color:var(--loss);">{row.losses}</td>
            <td><strong>{row.points}</strong></td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>

  </section>

  <section class="phone-section" class:active={phoneSection === 'results'}>
  <div class="section-title"><h2>Recent results</h2></div>
  <div class="grid" use:phoneLimit={4} style="grid-template-columns: repeat(auto-fit, minmax(min(100%, 260px), 1fr));">
    {#each finished.slice(-8) as fight (fight.id)}
      <div class="stat small">
        <div class="eyebrow">{fight.mat} · {fight.category}</div>
        <div>{fight.red.name} <span class="muted">—</span> {fight.blue.name}</div>
        <div class="muted">{fight.outcome === 'draw' ? 'Draw · DP / DP' : fight.outcome === 'win_red' ? 'Winner: ' + fight.red.name : 'Winner: ' + fight.blue.name}</div>
      </div>
    {/each}
  </div>
  </section>
</div>

<style>
  .small-portrait { width: 54px; height: 54px; border-radius: 14px; display: grid; place-items: center; color: #fff; font-weight: 800; }
  .brand-banner { height:clamp(90px,10vw,160px); border-radius:16px; background:var(--event-banner) center/contain no-repeat, var(--brand-primary-fill); background-color:var(--brand-surface-2); border:1px solid var(--brand-border); margin-bottom:16px; }
  .replay-strip { display:flex; align-items:center; gap:14px; flex-wrap:wrap; padding:10px 14px; margin-bottom:16px; border-radius:14px; border:1px solid var(--brand-accent); background:color-mix(in srgb, var(--brand-accent) 12%, var(--brand-surface-2)); }
  .clock-hero { display:flex; gap:clamp(14px,3vw,50px); align-items:center; flex-wrap:wrap; padding:clamp(16px,2vw,30px); border-radius:20px; border:1px solid var(--brand-primary); background:linear-gradient(135deg, rgba(59,130,246,.18), transparent), var(--brand-surface-2); box-shadow:var(--brand-shadow); }
  .big-clock { font-family:var(--font-display); font-size:clamp(3rem,8vw,8rem); font-weight:800; line-height:1; letter-spacing:.02em; color:var(--brand-accent); }
  .phone-sections { display:none; }
  @media (max-width: 700px) {
    .phone-sections { display:flex; gap:8px; overflow-x:visible; position:sticky; top:60px; z-index:6; padding:8px 0; margin-bottom:8px; background:var(--brand-surface); }
    .phone-sections .tab { flex:1; padding:8px 4px; text-align:center; }
    .phone-section { display:none; }
    .phone-section.active { display:block; animation:sectionIn .22s ease-out; }
    .live-duel { flex-direction:column; gap:8px; }
    .live-duel .vs { display:none; }
    @keyframes sectionIn { from { opacity:0; transform:translateY(8px); } to { opacity:1; transform:none; } }
  }
</style>
