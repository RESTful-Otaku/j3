<script lang="ts">
  import FighterCard from './FighterCard.svelte';

  export let event = '';
  export let category = '';
  export let mat = '';
  export let red: any = {};
  export let blue: any = {};
  export let redScore = 0;
  export let blueScore = 0;
  export let state = '';
  export let expandable = true;
  export let expanded = false;

  function fighter(f: any) {
    return {
      name: f?.name ?? f?.displayName ?? 'TBD',
      avatarPath: f?.avatarPath ?? f?.photoPath ?? '',
      club: f?.club ?? 'Independent',
      clubLogo: f?.clubLogo ?? f?.logoPath ?? '/media/club_logo_independent.svg',
      clubColour: f?.clubColour ?? '#64748b',
      belt: f?.belt ?? '',
      beltColour: f?.beltColour ?? '#94a3b8',
      heightCm: f?.heightCm ?? null,
      weightKg: f?.weightKg ?? null,
      gender: f?.gender ?? '',
    };
  }
  $: r = fighter(red); $: b = fighter(blue);
  function toggle() { if (expandable) expanded = !expanded; }
</script>

<div class="versus-card" class:expanded>
  <button class="versus-meta" type="button" on:click={toggle} aria-expanded={expanded}>
    <span class="m-event"><span>Event</span>{event || '—'}</span>
    <span class="m-cat"><span>Category</span>{category || '—'}</span>
    <span class="m-mat"><span>Mat</span>{mat || '—'}</span>
  </button>

  <div class="versus-row">
    <FighterCard {...r} score={redScore} {event} {category} {mat} expanded={expanded} expandable={false} />
    <div class="score-vs">
      <div class="scores"><strong>{redScore}</strong><span>vs</span><strong>{blueScore}</strong></div>
      {#if state}<span class="state">{state}</span>{/if}
    </div>
    <FighterCard {...b} score={blueScore} {event} {category} {mat} expanded={expanded} expandable={false} />
  </div>
</div>

<style>
  .versus-card {
    display:grid; gap:10px; padding:12px; border-radius:0;
    border:1px solid var(--brand-border); background:var(--brand-surface-2); min-width:0;
  }
  .versus-meta {
    display:flex; flex-wrap:wrap; gap:8px; justify-content:center;
    background:transparent; border:0; color:var(--brand-ink);
    cursor:pointer; padding:0;
  }
  .versus-meta span { display:inline-flex; align-items:center; gap:6px; font-weight:700; font-size:.82rem; }
  .versus-meta span span {
    font-size:.62rem; text-transform:uppercase; letter-spacing:.08em;
    color:var(--brand-muted);
  }
  .versus-row { display:grid; grid-template-columns:1fr auto 1fr; gap:10px; align-items:center; }
  .score-vs { display:grid; gap:4px; justify-items:center; }
  .scores { display:flex; align-items:center; gap:8px; font-family:var(--font-display); font-size:1.8rem; font-weight:900; }
  .scores span { font-size:.8rem; color:var(--brand-muted); }
  .state { font-size:.66rem; text-transform:uppercase; letter-spacing:.06em; color:var(--brand-muted); }
  @media (max-width:700px){
    .versus-row{ grid-template-columns:1fr; }
    .score-vs{ order:0; }
  }
</style>
