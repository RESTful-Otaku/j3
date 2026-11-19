<script lang="ts">
  import { onMount } from 'svelte';
  import {
    api,
    livePoll,
    type DemoFight,
    type EventDisplayInfo,
    type DisplayOfficial,
    type DemoFightEvent,
  } from '../../../lib/api';
  import ConfirmDialog from '../../../lib/components/ConfirmDialog.svelte';

  let event: EventDisplayInfo | null = null;
  let fights: DemoFight[] = [];
  let officials: DisplayOfficial[] = [];
  let selectedId = '';
  let seconds = 180;
  let paused = true;
  let celebrate = false;
  let confirmOpen = false;
  let loading = true;

  onMount(() => {
    const fromUrl = new URLSearchParams(location.search).get('fight') ?? '';
    const stopFights = livePoll(api.fights, 1500, (data) => {
      fights = data.fights;
      if (!selectedId) {
        const chosen = fromUrl && data.fights.find((f) => f.id === fromUrl)
          ? fromUrl
          : (data.fights.find((f) => f.state === 'live') ?? data.fights[0])?.id ?? '';
        selectedId = chosen;
      }
      loading = false;
    });
    const stopEvent = livePoll(api.event, 10000, (data) => (event = data));
    const officialTimer = setInterval(async () => {
      if (!selectedId) return;
      try {
        officials = (await api.officials(selectedId)).officials;
      } catch {
        /* keep previous frame */
      }
    }, 1500);
    const clock = setInterval(() => {
      if (!paused && seconds > 0) seconds -= 1;
    }, 1000);
    return () => {
      stopFights();
      stopEvent();
      clearInterval(officialTimer);
      clearInterval(clock);
    };
  });

  function payload(e: DemoFightEvent): Record<string, unknown> {
    return (e.payload ?? {}) as Record<string, unknown>;
  }
  function score(fight: DemoFight | undefined, corner: 'red' | 'blue') {
    if (!fight) return 0;
    return fight.events
      .filter((e) => e.kind === 'score' && payload(e).corner === corner)
      .reduce((sum, e) => sum + Number(payload(e).points ?? 0), 0);
  }
  function moves(fight: DemoFight | undefined, corner: 'red' | 'blue') {
    return (fight?.events ?? []).filter((e) => payload(e).corner === corner);
  }
  function initials(name: string) {
    return name.replace(/\s*\(.*\)$/, '').split(' ').slice(0, 2).map((p) => p[0] ?? '').join('').toUpperCase();
  }
  function clockText(value: number) {
    return `${String(Math.floor(value / 60)).padStart(2, '0')}:${String(value % 60).padStart(2, '0')}`;
  }
  function method(fight: DemoFight | undefined, corner: 'red' | 'blue') {
    if (!fight?.outcome) return '';
    const isWinner = (corner === 'red' && fight.outcome === 'win_red') ||
      (corner === 'blue' && fight.outcome === 'win_blue');
    if (fight.outcome === 'draw') return 'DP';
    const submission = fight.events.some((e) => e.kind === 'submission');
    const technical = fight.outcome.startsWith('dq') || fight.state === 'forfeit';
    return `${isWinner ? 'W' : 'L'}${submission ? 'S' : technical ? 'T' : 'P'}`;
  }
  function celebrateResult() {
    celebrate = true;
    setTimeout(() => (celebrate = false), 2800);
  }
  function fullscreen() {
    document.documentElement.requestFullscreen?.();
  }

  $: selected = fights.find((f) => f.id === selectedId);
  $: matName = selected?.mat ?? event?.mats[0] ?? 'Mat 1';
</script>

<svelte:head><title>J3 · Full-screen scoreboard</title></svelte:head>

<div class="display-shell">
  <header class="display-header">
    <div>
      <div class="event-mark"><span class="crest">J3</span><span>{event?.name ?? 'Tournament'}</span></div>
      <div class="meta">{matName} · {selected?.category ?? 'Contest'} · {selected?.round ?? 'Live'}</div>
    </div>
    <div style="display:flex; gap:10px; align-items:center; flex-wrap:wrap;">
      <span class="pill" class:off={!selected || selected.state !== 'live'}>
        <span class="dot" class:off={!selected || selected.state !== 'live'}></span>{selected?.state ?? 'connecting…'}
      </span>
      <button class="btn" on:click={fullscreen}>⛶ Fullscreen</button>
      <button class="btn" on:click={() => (paused = !paused)}>{paused ? '▶ Resume' : '⏸ Pause (ref stoppage)'}</button>
      <button class="btn primary" on:click={() => (confirmOpen = true)}>Confirm result</button>
    </div>
  </header>

  {#if loading || !selected}
    <div class="placeholder">Connecting to the venue server…</div>
  {:else}
    <div class="hero-scoreboard">
      <section class="fighter-hero" class:celebrate={celebrate && method(selected, 'red').startsWith('W')}>
        <div class="portrait">
          {initials(selected.red.name || 'TBD')}
          {#if selected.red.logoPath}<img src={selected.red.logoPath} alt="" />{/if}
          <span class="club-crest" style="background:{selected.red.clubColour};">{initials(selected.red.club || 'IND')}</span>
        </div>
        <div class="who">
          <div class="name">{selected.red.name || 'TBD'}</div>
          <span class="belt" style="background:{selected.red.beltColour || '#fbbf24'};">{selected.red.belt || 'Independent'}</span>
          <div class="stat">{selected.red.heightCm ? `${Math.round(selected.red.heightCm)} cm` : '—'} · {selected.red.weightKg ? `${selected.red.weightKg.toFixed(1)} kg` : '—'}</div>
          <div class="stat">{selected.red.club || 'Independent'}</div>
          <div class="moves-side">
            {#each moves(selected, 'red') as event (event.seq)}
              <div class="move-chip" class:penalty={event.kind === 'penalty'}>
                <span>#{event.seq} {event.kind === 'score' ? String(payload(event).technique ?? 'Score') : `⚠ ${String(payload(event).card ?? 'penalty')}`}</span>
                <span class="pts">{event.kind === 'score' ? `+${payload(event).points ?? 0}` : '—'}</span>
              </div>
            {:else}<span class="stat">No moves recorded</span>{/each}
          </div>
        </div>
        <div class="big-score">{score(selected, 'red')}</div>
        {#if method(selected, 'red')}<span class="badge fighter">{method(selected, 'red')}</span>{/if}
      </section>

      <section class="clock-panel">
        <div class="round">{matName} · {selected.round}</div>
        <div class="clock" class:flash={paused}>{clockText(seconds)}</div>
        <div class="meta">{paused ? 'Referee stoppage · timer paused' : 'Fight in progress'}</div>
      </section>

      <section class="fighter-hero blue" class:celebrate={celebrate && method(selected, 'blue').startsWith('W')}>
        <div class="portrait">
          {initials(selected.blue.name || 'TBD')}
          {#if selected.blue.logoPath}<img src={selected.blue.logoPath} alt="" />{/if}
          <span class="club-crest" style="background:{selected.blue.clubColour};">{initials(selected.blue.club || 'IND')}</span>
        </div>
        <div class="who">
          <div class="name">{selected.blue.name || 'TBD'}</div>
          <span class="belt" style="background:{selected.blue.beltColour || '#fbbf24'};">{selected.blue.belt || 'Independent'}</span>
          <div class="stat">{selected.blue.heightCm ? `${Math.round(selected.blue.heightCm)} cm` : '—'} · {selected.blue.weightKg ? `${selected.blue.weightKg.toFixed(1)} kg` : '—'}</div>
          <div class="stat">{selected.blue.club || 'Independent'}</div>
          <div class="moves-side">
            {#each moves(selected, 'blue') as event (event.seq)}
              <div class="move-chip" class:penalty={event.kind === 'penalty'}>
                <span>#{event.seq} {event.kind === 'score' ? String(payload(event).technique ?? 'Score') : `⚠ ${String(payload(event).card ?? 'penalty')}`}</span>
                <span class="pts">{event.kind === 'score' ? `+${payload(event).points ?? 0}` : '—'}</span>
              </div>
            {:else}<span class="stat">No moves recorded</span>{/each}
          </div>
        </div>
        <div class="big-score">{score(selected, 'blue')}</div>
        {#if method(selected, 'blue')}<span class="badge fighter">{method(selected, 'blue')}</span>{/if}
      </section>
    </div>

    <div class="officials-strip">
      {#each officials as official (official.id + official.role)}
        <div class="official slide-in">
          <span class="face">{initials(official.name)}</span>
          <span><span class="role">{official.role}</span><br /><span class="name">{official.name}</span></span>
        </div>
      {:else}
        <span class="muted small">Officials not assigned for this contest.</span>
      {/each}
    </div>
  {/if}

  <div style="margin-top:24px; display:flex; gap:10px; flex-wrap:wrap;">
    {#each fights as fight}
      <button class="tab" class:active={fight.id === selectedId} on:click={() => (selectedId = fight.id)}>
        {fight.mat} · {fight.red.name} vs {fight.blue.name}
      </button>
    {/each}
  </div>
</div>

{#if celebrate && selected}
  <div class="celebration-backdrop" role="presentation">
    <div class="celebration">
      <div class="eyebrow">Contest decided</div>
      <div class="portrait celebrate">{initials(method(selected, 'red').startsWith('W') ? selected.red.name : selected.blue.name)}</div>
      <h1>{method(selected, 'red').startsWith('W') ? selected.red.name : selected.blue.name}</h1>
      <p>{method(selected, 'red').startsWith('W') ? method(selected, 'red') : method(selected, 'blue')} · advances</p>
    </div>
  </div>
{/if}

<ConfirmDialog
  open={confirmOpen}
  title="Confirm this result?"
  message="The winner, points and promotion will be committed to the tournament record and shown on all displays."
  confirmLabel="Confirm result"
  onConfirm={celebrateResult}
/>

<style>
  .celebration-backdrop { position: fixed; inset: 0; z-index: 950; display: grid; place-items: center; background: rgba(2,6,23,.82); backdrop-filter: blur(8px); }
  .celebration { text-align: center; animation: pop .5s cubic-bezier(.16,1,.3,1); }
  .celebration .eyebrow { color: var(--brand-accent); letter-spacing: .2em; text-transform: uppercase; font-weight: 800; }
  .celebration .portrait { width: clamp(160px, 24vw, 420px); height: clamp(160px, 24vw, 420px); margin: 22px auto; border-radius: 24%; display: grid; place-items: center; font-family: var(--font-display); font-size: clamp(4rem, 10vw, 10rem); font-weight: 800; color: #fff; background: linear-gradient(145deg, var(--brand-primary), var(--brand-accent)); }
  .celebration p { color: var(--brand-muted); font-size: 1.2rem; }
  @keyframes pop { from { transform: scale(.6); opacity: 0; } to { transform: scale(1); opacity: 1; } }
</style>
