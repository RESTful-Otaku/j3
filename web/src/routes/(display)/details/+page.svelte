<script lang="ts">
  import { onMount } from 'svelte';
  import { api, livePoll, type EventDisplayInfo } from '../../../lib/api';

  let event: EventDisplayInfo | null = null;
  let details: Record<string, any> = {};

  onMount(() => livePoll(api.event, 10000, (data) => {
    event = data;
    try { details = JSON.parse(data.detailsJson || '{}'); } catch { details = {}; }
  }));
</script>

<svelte:head><title>J3 · Event details & requirements</title></svelte:head>

<div class="display-shell">
  <header class="display-header">
    <div><div class="event-mark"><span class="crest">{event?.logoPath ? '🏅' : 'J3'}</span>{event?.name ?? 'Tournament'}</div>
      <div class="meta">Official details · rules · requirements · venue</div></div>
    <a class="btn" href="/public" target="_blank" rel="noreferrer">📺 Live dashboard ↗</a>
  </header>

  <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(min(100%, 300px), 1fr));">
    <div class="card slide-in">
      <div class="eyebrow">Venue & times</div><h2>{details.venue || 'Main hall'}</h2>
      <p class="muted">Opening {details.schedule?.[0]?.time ?? '08:30'} · last event {details.schedule?.at(-1)?.time ?? '16:30'}</p>
      <div style="margin-top:14px; display:grid; gap:8px;">
        {#each details.schedule ?? [] as item}
          <div class="stat small"><strong>{item.time}</strong> · {item.label} <span class="badge spectator">{item.kind}</span></div>
        {/each}
      </div>
    </div>

    <div class="card slide-in">
      <div class="eyebrow">Required protection</div><h2>Bring & wear</h2>
      <div class="capabilities" style="margin-top:14px;">
        {#each details.protection ?? [] as requirement}
          <span class="cap on">✓ {requirement}</span>
        {:else}<span class="muted small">No special equipment listed.</span>{/each}
      </div>
      <h3 style="margin-top:20px;">Amenities</h3>
      <div class="capabilities">
        {#each details.amenities ?? [] as amenity}<span class="cap on">{amenity}</span>{/each}
      </div>
    </div>

    <div class="card slide-in">
      <div class="eyebrow">Fighter identification</div><h2>Competition corners</h2>
      <div style="display:flex; gap:14px; margin-top:16px;">
        <div class="corner-chip" style="background:{details.cornerColours?.red ?? '#ef4444'}">RED</div>
        <div class="corner-chip" style="background:{details.cornerColours?.blue ?? '#3b82f6'}">BLUE</div>
      </div>
      <p class="muted small" style="margin-top:14px;">Colours are set by the tournament runner and shown on every mat, scoreboard and bracket.</p>
      <h3 style="margin-top:20px;">Mat areas</h3>
      <div class="capabilities">{#each event?.mats ?? [] as mat}<span class="cap on">{mat}</span>{/each}</div>
    </div>

    <div class="card slide-in">
      <div class="eyebrow">Scoring & discipline rules</div><h2>How contests are decided</h2>
      <p class="muted small">Each discipline carries its own point system, timers, round count, requirements and attribution. The values below are the active event defaults.</p>
      <div class="stat small" style="margin-top:12px;">Match length <strong>{details.matchSeconds ?? 300}s</strong></div>
      <div class="stat small" style="margin-top:8px;">Ippon <strong>{details.ippon ?? 3}</strong> · Waza-ari <strong>{details.wazaAri ?? 1}</strong></div>
      <div class="stat small" style="margin-top:8px;">Submission <strong>{details.submissionInstantWin ? 'instant win' : 'points only'}</strong></div>
      <div class="stat small" style="margin-top:8px;">Shido limit <strong>{details.shidoLimit ?? 3}</strong></div>
    </div>
  </div>
</div>

<style>
  .corner-chip { padding:18px 28px; border-radius:14px; color:#fff; font-family:var(--font-display); font-weight:800; font-size:1.4rem; box-shadow:0 12px 30px rgba(0,0,0,.35); }
</style>
