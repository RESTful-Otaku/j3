<script lang="ts">
  import { onMount } from 'svelte';
  import { initials } from '../../../lib/format';
  import { session, uiRole } from '../../../lib/auth';
  import { phoneLimit, phoneRows } from '../../../lib/responsive';
  import Card from '../../../lib/components/Card.svelte';
  import { api, livePoll, type DemoFight, type EventDisplayInfo, type StandingRow, type ScheduleItem } from '../../../lib/api';
  import { eventTitle } from '../../../lib/event-title';
  import { replaySnapshot } from '../../../lib/demo-replay';
  import { onLiveChange } from '../../../lib/live-updates';

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
    const stopLive = onLiveChange(['fights', 'standings', 'schedule', 'medals'], () => {
      void api.fights().then((data) => (fights = data.fights)).catch(() => {});
      void api.standings().then((data) => (standings = data.standings)).catch(() => {});
      void api.schedule().then((data) => (schedule = data.items)).catch(() => {});
    });
    return () => {
      stopEvent(); stopFights(); stopStandings(); stopSchedule(); stopLive(); clearInterval(clock);
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

  <section class="phone-section view-enter class:active={phoneSection === 'live'}">
    <Card title="Today's timeline" eyebrow="Schedule" expandable preview={nextUp ? `${nextUp.name} · ${formatTime(nextUp.projectedStart)}–${formatTime(nextUp.projectedEnd)}` : 'Schedule complete'}>
    <div class="card">
      <table style="width:100%; border-collapse:collapse; font-size:clamp(.9rem,1.2vw,1.3rem);">
        <thead><tr class="muted small" style="text-align:left;"><th style="padding:8px;">Projected</th><th>Event</th><th>Area</th><th>Status</th><th></th></tr></thead>
        <tbody use:phoneRows={5}>
          {#each schedule as item (item.id)}
            <tr style="border-top:1px solid var(--brand-border);" class:flash={item.id === nextUp?.id} class="reveal-item">
              <td style="padding:8px; font-weight:700;">{formatTime(item.projectedStart)}–{formatTime(item.projectedEnd)}</td>
              <td>{item.name}</td><td>{item.area}</td>
              <td><span class="badge {item.status === 'complete' ? 'spectator' : 'fighter'}">{item.status}</span></td>
              <td>{#if canControlSchedule && item.status !== 'complete'}<button class="tab" on:click={() => complete(item.id)}>Mark done</button>{/if}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
    </Card>

    <Card title="On the mats now" eyebrow="Live" expandable preview={live.length ? `${live.length} contest${live.length === 1 ? '' : 's'}` : 'No live contests'}>
    <div class="section-title"><span class="pill"><span class="dot"></span>Live</span></div>
    <div class="grid" use:phoneLimit={3} style="grid-template-columns: repeat(auto-fit, minmax(min(100%, 420px), 1fr));">
      {#each live as fight (fight.id)}
        <div class="card slide-in reveal-item" style="border-color:var(--brand-primary);">
          <div class="meta">{fight.mat} · {fight.category} · {fight.round}</div>
          <div class="live-duel">
            <div class="duel-side red">
              <span class="portrait small-portrait" style="background:{fight.red.clubColour};">{initials(fight.red.name)}</span>
              <strong class="duel-name">{fight.red.name}</strong>
            </div>
            <span class="vs">vs</span>
            <div class="duel-side blue">
              <strong class="duel-name">{fight.blue.name}</strong>
              <span class="portrait small-portrait" style="background:{fight.blue.clubColour};">{initials(fight.blue.name)}</span>
            </div>
          </div>
        </div>
      {:else}
        <div class="placeholder">No live contests at this moment.</div>
      {/each}
    </div>
    </Card>
  </section>

  <section class="phone-section view-enter class:active={phoneSection === 'next'}">
    <Card title="Next up — report to your mat" eyebrow="Upcoming" expandable preview={upcoming.length ? `${upcoming.length} bout${upcoming.length === 1 ? '' : 's'}` : 'Schedule is clear'}>
    <div class="section-title"><span class="muted small">{upcoming.length} upcoming</span></div>
    <div class="grid" use:phoneLimit={4} style="grid-template-columns: repeat(auto-fit, minmax(min(100%, 300px), 1fr));">
      {#each upcoming as fight (fight.id)}
        <div class="stat flash reveal-item">
          <div class="eyebrow">{fight.mat} · {fight.category}</div>
          <div class="matchup">
            <span class="fighter red">{fight.red.name}</span>
            <span class="sep">vs</span>
            <span class="fighter blue">{fight.blue.name}</span>
          </div>
          <div class="small muted">{fight.round}</div>
        </div>
      {:else}
        <div class="placeholder">Schedule is clear.</div>
      {/each}
    </div>
    </Card>
  </section>

  <section class="phone-section view-enter class:active={phoneSection === 'table'}">
    <Card title="Leaderboard" eyebrow="Standings" expandable preview={leaders.length ? `#1 ${leaders[0]?.name ?? '—'}` : 'No standings yet'}>
    <div class="section-title"><span class="muted small">Points: 3 per win, 1 per draw</span></div>
    <div class="card">
      <table style="width:100%; border-collapse:collapse; font-size:clamp(.9rem,1.2vw,1.3rem);">
        <thead><tr class="muted small" style="text-align:left;"><th style="padding:8px;">#</th><th>Fighter</th><th>Category</th><th>Club</th><th>W</th><th>D</th><th>L</th><th>Pts</th></tr></thead>
        <tbody use:phoneRows={5}>
          {#each leaders as row, i (row.fighterId)}
            <tr style="border-top:1px solid var(--brand-border);" class="reveal-item">
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
    </Card>
  </section>

  <section class="phone-section view-enter class:active={phoneSection === 'results'}">
    <Card title="Recent results" eyebrow="Finished" expandable preview={finished.length ? `Last: ${finished[finished.length-1]?.red.name ?? '—'} vs ${finished[finished.length-1]?.blue.name ?? '—'}` : 'No results yet'}>
    <div class="section-title"></div>
    <div class="grid" use:phoneLimit={4} style="grid-template-columns: repeat(auto-fit, minmax(min(100%, 260px), 1fr));">
      {#each finished.slice(-8) as fight (fight.id)}
        <div class="stat small reveal-item">
          <div class="eyebrow">{fight.mat} · {fight.category}</div>
          <div class="matchup compact">
            <span class="fighter red">{fight.red.name}</span>
            <span class="sep">—</span>
            <span class="fighter blue">{fight.blue.name}</span>
          </div>
          <div class="muted">{fight.outcome === 'draw' ? 'Draw · DP / DP' : fight.outcome === 'win_red' ? 'Winner: ' + fight.red.name : 'Winner: ' + fight.blue.name}</div>
        </div>
      {/each}
    </div>
    </Card>
  </section>
</div>

<style>
  .small-portrait { width: 54px; height: 54px; border-radius: 14px; display: grid; place-items: center; color: #fff; font-weight: 800; }
  .live-duel { display:flex; justify-content:space-between; align-items:center; gap:18px; margin-top:12px; }
  .duel-side { display:flex; align-items:center; gap:12px; min-width:0; flex:1; }
  .duel-side.blue { justify-content:flex-end; }
  .duel-name { font-size:1.2rem; min-width:0; overflow-wrap:anywhere; }
  .vs { font-family:var(--font-display); font-size:2rem; }
  .matchup { display:flex; align-items:center; gap:6px; font-size:1.05rem; font-weight:700; }
  .matchup .fighter { min-width:0; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
  .matchup .fighter.red { flex:1; text-align:right; }
  .matchup .fighter.blue { flex:1; }
  .matchup .sep { flex:none; color:var(--brand-muted); font-size:.78rem; font-weight:600; }
  .matchup.compact { font-size:.92rem; }
  /* Host banner is full-bleed across the display content area — it belongs to the organiser/club, not a card. */
  .brand-banner {
    height:clamp(90px,10vw,160px); width:calc(100% + clamp(28px,4vw,80px));
    margin:0 calc(-1 * clamp(14px,2vw,40px)) 16px; border-radius:0;
    border:0; border-bottom:1px solid var(--brand-border);
    background:var(--event-banner) center/contain no-repeat, var(--brand-primary-fill);
    background-color:var(--brand-surface-2);
  }
  .replay-strip { display:flex; align-items:center; gap:14px; flex-wrap:wrap; padding:10px 14px; margin-bottom:16px; border-radius:14px; border:1px solid var(--brand-accent); background:color-mix(in srgb, var(--brand-accent) 12%, var(--brand-surface-2)); }
  .clock-hero { display:flex; gap:clamp(14px,3vw,50px); align-items:center; flex-wrap:wrap; padding:clamp(16px,2vw,30px); border-radius:20px; border:1px solid var(--brand-primary); background:linear-gradient(135deg, rgba(59,130,246,.18), transparent), var(--brand-surface-2); box-shadow:var(--brand-shadow); }
  .big-clock { font-family:var(--font-display); font-size:clamp(3rem,8vw,8rem); font-weight:800; line-height:1; letter-spacing:.02em; color:var(--brand-accent); }
  .phone-sections { display:none; }
  @media (max-width: 700px) {
    .phone-sections {
      display:flex; gap:6px; overflow-x:auto; position:sticky;
      top:calc(var(--mobile-bar-height) + var(--safe-top)); z-index:6;
      padding:8px 0; margin-bottom:8px; background:var(--brand-surface);
      scrollbar-width:none;
    }
    .phone-sections::-webkit-scrollbar { display:none; }
    .phone-sections .tab {
      flex:1 0 auto; min-width:0; max-width:none; padding:8px 4px; text-align:center;
      white-space:nowrap; overflow:hidden; text-overflow:ellipsis;
    }
    .phone-section { display:none; }
    .phone-section.active { display:block; animation:sectionIn .22s ease-out; }
    .live-duel { gap:7px; margin-top:8px; }
    .duel-side { gap:6px; }
    .small-portrait { width:34px; height:34px; border-radius:10px; font-size:.68rem; }
    .duel-name { font-size:.94rem; line-height:1.15; }
    .live-duel .vs { display:none; }
    .matchup { font-size:.92rem; gap:4px; }
    .matchup.compact { font-size:.84rem; }
    /* Phone board: compact the hero clock and banner so the actual fights land
       higher on the screen and there is far less scrolling before content. */
    .brand-banner { height: clamp(56px, 7vw, 96px); margin: 0 calc(-1 * clamp(10px, 2vw, 24px)) 12px; }
    .clock-hero { padding: clamp(10px, 2vw, 18px); gap: clamp(10px, 3vw, 26px); }
    .big-clock { font-size: clamp(1.7rem, 9vw, 4.2rem); }
    .clock-hero .eyebrow { font-size: .7rem; }
    .phone-sections { top: calc(var(--mobile-bar-height) + var(--safe-top)); }
    @keyframes sectionIn { from { opacity:0; transform:translateY(8px); } to { opacity:1; transform:none; } }
  }
</style>
