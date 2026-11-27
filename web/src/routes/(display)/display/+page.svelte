<script lang="ts">
  import { onMount } from 'svelte';
  import { initials, clockText } from '../../../lib/format';
  import {
    api,
    livePoll,
    type DemoFight,
    type EventDisplayInfo,
    type DisplayOfficial,
    type DemoFightEvent,
    type FightState,
  } from '../../../lib/api';
  import ConfirmDialog from '../../../lib/components/ConfirmDialog.svelte';
  import { onLiveChange } from '../../../lib/live-updates';

  let event: EventDisplayInfo | null = null;
  let fights: DemoFight[] = [];
  let officials: DisplayOfficial[] = [];
  let selectedId = '';
  let seconds = 180;
  let paused = true;
  let celebrate = false;
  let confirmOpen = false;
  let loading = true;
  let liveState: FightState | null = null;

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
    const stopStream = api.eventStream((message) => {
      if (message.payload && (message.payload.fightId === selectedId || message.fightId === selectedId)) {
        liveState = message.payload;
      }
    });
    const stopLive = onLiveChange(['fights', 'officials', 'medals'], () => {
      void api.fights().then((data) => { fights = data.fights; loading = false; }).catch(() => {});
      if (selectedId) void api.officials(selectedId).then((data) => (officials = data.officials)).catch(() => {});
    });
    return () => {
      stopFights();
      stopEvent();
      stopStream();
      stopLive();
      clearInterval(officialTimer);
      clearInterval(clock);
    };
  });

  function payload(e: DemoFightEvent): Record<string, unknown> {
    return (e.payload ?? {}) as Record<string, unknown>;
  }
  function score(fight: DemoFight | undefined, corner: 'red' | 'blue') {
    if (!fight) return 0;
    if (liveState?.fightId === fight.id) return corner === 'red' ? liveState.redScore : liveState.blueScore;
    return fight.events
      .filter((e) => e.kind === 'score' && payload(e).corner === corner)
      .reduce((sum, e) => sum + Number(payload(e).points ?? 0), 0);
  }
  function moves(fight: DemoFight | undefined, corner: 'red' | 'blue') {
    return (fight?.events ?? []).filter((e) => e.kind === 'score' && payload(e).corner === corner);
  }
  function fouls(fight: DemoFight | undefined, corner: 'red' | 'blue') {
    return (fight?.events ?? []).filter((e) => e.kind === 'penalty' && payload(e).corner === corner);
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
  async function startFight() {
    if (selected) liveState = await api.startFight(selected.id, 180, 1);
  }
  async function pauseFight() {
    if (selected) liveState = await api.pauseFight(selected.id);
  }
  async function resumeFight() {
    if (selected) liveState = await api.resumeFight(selected.id);
  }
  async function submitResult(outcome: string) {
    if (selected) liveState = await api.submitFight(selected.id, outcome);
    confirmOpen = false;
    celebrateResult();
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
      <div class="meta">Event: {event?.name ?? 'Tournament'} · Category: {selected?.category ?? 'Contest'} · Round: {selected?.round ?? 'Live'} · {matName}</div>
    </div>
    <div style="display:flex; gap:10px; align-items:center; flex-wrap:wrap;">
      <span class="pill" class:off={!selected || selected.state !== 'live'}>
        <span class="dot" class:off={!selected || selected.state !== 'live'}></span>{selected?.state ?? 'connecting…'}
      </span>
      <button class="btn" on:click={fullscreen}>⛶ Fullscreen</button>
      <a class="btn" href="/mat">🎛️ Operator console</a>
    </div>
  </header>

  {#if loading || !selected}
    <div class="placeholder">Connecting to the venue server…</div>
  {:else}
    <div class="scoreboard-stage">
      <section class="clock-panel">
        <div class="round">{matName} · {selected.round} · {selected.category}</div>
        <div class="clock" class:paused-flash={liveState ? !liveState.running : paused}>{clockText(liveState?.seconds ?? seconds)}</div>
        <div class="meta">{(liveState ? !liveState.running : paused) ? 'Referee stoppage · timer paused' : 'Fight in progress'} · authoritative server clock</div>
        {#if liveState && liveState.holdSeconds > 0}
          <div class="hold-badge">HOLD · {liveState.holdCorner} · {liveState.holdSeconds}s</div>
        {/if}
      </section>

      <div class="fighter-row">
        <section class="fighter-hero" class:celebrate={celebrate && method(selected, 'red').startsWith('W')}>
          <div class="portrait">
            {#if selected.red.avatarPath}<img class="fighter-photo" src={selected.red.avatarPath} alt={selected.red.name} />{:else}{initials(selected.red.name || 'TBD')}{/if}
            <span class="club-crest" style="background:{selected.red.clubColour};">{#if selected.red.logoPath}<img src={selected.red.logoPath} alt="" />{:else}{initials(selected.red.club || 'IND')}{/if}</span>
          </div>
          <div class="who">
            <div class="name">{selected.red.name || 'TBD'}</div>
            <div class="badges">
              <span class="belt" style="background:{selected.red.beltColour || '#fbbf24'};">{selected.red.belt || 'Independent'}</span>
              {#if method(selected, 'red')}<span class="badge fighter">{method(selected, 'red')}</span>{/if}
            </div>
            <div class="stat">{selected.red.heightCm ? `${Math.round(selected.red.heightCm)} cm` : '—'} · {selected.red.weightKg ? `${selected.red.weightKg.toFixed(1)} kg` : '—'}</div>
            <div class="stat">{selected.red.club || 'Independent'}</div>
            <div class="moves-side">
              {#each moves(selected, 'red').slice(-3) as move (move.seq)}
                <div class="move-chip"><span>Last {move.seq} · {String(payload(move).technique ?? 'Score')}</span><span class="pts">+{payload(move).points ?? 0}</span></div>
              {:else}<span class="stat">No moves recorded</span>{/each}
            </div>
            <div class="fouls">
              {#each fouls(selected, 'red') as foul (foul.seq)}
                <span class="foul-card {String(payload(foul).card ?? 'yellow')}">{String(payload(foul).card ?? 'yellow').toLowerCase() === 'red' ? 'R' : 'Y'}</span>
              {/each}
            </div>
          </div>
          <div class="big-score">{score(selected, 'red')}</div>
        </section>

        <section class="fighter-hero blue" class:celebrate={celebrate && method(selected, 'blue').startsWith('W')}>
          <div class="big-score">{score(selected, 'blue')}</div>
          <div class="who">
            <div class="name">{selected.blue.name || 'TBD'}</div>
            <div class="badges">
              <span class="belt" style="background:{selected.blue.beltColour || '#fbbf24'};">{selected.blue.belt || 'Independent'}</span>
              {#if method(selected, 'blue')}<span class="badge fighter">{method(selected, 'blue')}</span>{/if}
            </div>
            <div class="stat">{selected.blue.heightCm ? `${Math.round(selected.blue.heightCm)} cm` : '—'} · {selected.blue.weightKg ? `${selected.blue.weightKg.toFixed(1)} kg` : '—'}</div>
            <div class="stat">{selected.blue.club || 'Independent'}</div>
            <div class="moves-side">
              {#each moves(selected, 'blue').slice(-3) as move (move.seq)}
                <div class="move-chip"><span>Last {move.seq} · {String(payload(move).technique ?? 'Score')}</span><span class="pts">+{payload(move).points ?? 0}</span></div>
              {:else}<span class="stat">No moves recorded</span>{/each}
            </div>
            <div class="fouls">
              {#each fouls(selected, 'blue') as foul (foul.seq)}
                <span class="foul-card {String(payload(foul).card ?? 'yellow')}">{String(payload(foul).card ?? 'yellow').toLowerCase() === 'red' ? 'R' : 'Y'}</span>
              {/each}
            </div>
          </div>
          <div class="portrait">
            {#if selected.blue.avatarPath}<img class="fighter-photo" src={selected.blue.avatarPath} alt={selected.blue.name} />{:else}{initials(selected.blue.name || 'TBD')}{/if}
            <span class="club-crest" style="background:{selected.blue.clubColour};">{#if selected.blue.logoPath}<img src={selected.blue.logoPath} alt="" />{:else}{initials(selected.blue.club || 'IND')}{/if}</span>
          </div>
        </section>
      </div>
    </div>

    <div class="officials-strip">
      {#each officials as official (official.id + official.role)}
        <div class="official slide-in">
          <span class="face">{#if (official as any).photoPath}<img src={(official as any).photoPath} alt="" />{:else}{initials(official.name)}{/if}</span>
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
  bind:open={confirmOpen}
  title="Confirm this result?"
  message="The winner, points and promotion will be committed to the tournament record and shown on all displays."
  confirmLabel="Confirm result"
  onConfirm={() => void submitResult('win_red')}
/>

<style>
  .celebration-backdrop { position: fixed; inset: 0; z-index: 950; display: grid; place-items: center; background: rgba(2,6,23,.82); backdrop-filter: blur(8px); }
  .celebration { text-align: center; animation: pop .5s cubic-bezier(.16,1,.3,1); }
  .celebration .eyebrow { color: var(--brand-accent); letter-spacing: .2em; text-transform: uppercase; font-weight: 800; }
  .celebration .portrait { width: clamp(160px, 24vw, 420px); height: clamp(160px, 24vw, 420px); margin: 22px auto; border-radius: 24%; display: grid; place-items: center; font-family: var(--font-display); font-size: clamp(4rem, 10vw, 10rem); font-weight: 800; color: #fff; background: linear-gradient(145deg, var(--brand-primary), var(--brand-accent)); }
  .celebration p { color: var(--brand-muted); font-size: 1.2rem; }
  .foul-card { display:inline-grid; place-items:center; width:34px; height:26px; border-radius:6px; font-size:.72rem; font-weight:900; color:#fff; box-shadow:0 4px 14px rgba(0,0,0,.35); }
  .hold-badge { display:inline-block; margin-top:8px; padding:6px 12px; border-radius:10px; background:var(--brand-accent); color:#0f172a; font-weight:800; }
  .foul-card.yellow { background:#facc15; color:#422006; }
  .foul-card.red { background:#ef4444; }
  :global(.paused-flash) { color: var(--draw) !important; animation: pausedPulse .9s ease-in-out infinite alternate; }
  @keyframes pausedPulse { from { opacity: 1; text-shadow: 0 0 0 rgba(245,158,11,0); } to { opacity: .45; text-shadow: 0 0 32px rgba(245,158,11,.9); } }
  @keyframes pop { from { transform: scale(.6); opacity: 0; } to { transform: scale(1); opacity: 1; } }
</style>
