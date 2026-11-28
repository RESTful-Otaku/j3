<script lang="ts">
  import FighterPortrait from './FighterPortrait.svelte';

  export let name = 'TBD';
  export let avatarPath = '';
  export let club = 'Independent';
  export let clubLogo = '/media/club_logo_independent.svg';
  export let clubColour = '#64748b';
  export let belt = '';
  export let beltColour = '#94a3b8';
  export let category = '';
  export let mat = '';
  export let event = '';
  export let heightCm: number | null = null;
  export let weightKg: number | null = null;
  export let gender = '';
  export let state = '';
  export let score: number | null = null;
  export let stats: Record<string, string | number> = {};
  export let expandable = true;
  export let expanded = false; // default shrunk
  export let onToggle: () => void = () => {};

  $: clubLogoSafe = clubLogo && clubLogo.trim() ? clubLogo : '/media/club_logo_independent.svg';
  $: statEntries = Object.entries(stats);
  $: hasStats = statEntries.length > 0 || heightCm != null || weightKg != null || !!belt || !!category || !!mat || !!gender;

  function toggle() { if (expandable) { expanded = !expanded; onToggle(); } }
</script>

<div class="fighter-card" class:expanded class:collapsed={!expanded} class:hide-photo={false}>
  <button class="fighter-head" type="button" on:click={toggle} aria-expanded={expanded} aria-label={name}>
    <div class="profile-wrap">
      <FighterPortrait src={avatarPath} label={name} size="72px" fallback={name.split(' ').map(p=>p[0]).join('')} clubColour={clubColour} />
      {#if score != null}<span class="score-badge">{score}</span>{/if}
    </div>
    <div class="club-line">
      <img class="club-logo" src={clubLogoSafe} alt={club} loading="lazy" />
      <span class="name">{name}</span>
    </div>
    {#if state}<span class="state">{state}</span>{/if}
  </button>

  {#if expanded && hasStats}
    <div class="stats">
      {#if event}<div><span>Event</span><strong>{event}</strong></div>{/if}
      {#if category}<div><span>Category</span><strong>{category}</strong></div>{/if}
      {#if mat}<div><span>Mat</span><strong>{mat}</strong></div>{/if}
      {#if club}<div><span>Club</span><strong>{club}</strong></div>{/if}
      {#if belt}<div><span>Grade</span><strong><span class="belt-dot" style="background:{beltColour}"></span>{belt}</strong></div>{/if}
      {#if heightCm != null}<div><span>Height</span><strong>{heightCm} cm</strong></div>{/if}
      {#if weightKg != null}<div><span>Weight</span><strong>{weightKg} kg</strong></div>{/if}
      {#if gender}<div><span>Gender</span><strong>{gender}</strong></div>{/if}
      {#each statEntries as [k, v]}<div><span>{k}</span><strong>{v}</strong></div>{/each}
    </div>
  {/if}
</div>

<style>
  .fighter-card { display:grid; gap:8px; padding:12px; border-radius:0; border:1px solid var(--brand-border); background:var(--brand-surface-2); min-width:0; }
  .fighter-card.expanded { border-color:var(--brand-primary); }
  .fighter-head { display:grid; gap:8px; justify-items:center; align-items:center; background:transparent; border:0; color:var(--brand-ink); cursor:pointer; padding:0; min-width:0; }
  .profile-wrap { position:relative; }
  .score-badge { min-width:24px; height:24px; padding:0 6px; border-radius:0; background:var(--brand-primary); color:var(--brand-on-primary); font-weight:900; display:grid; place-items:center; font-size:.78rem; }
  .club-line { display:flex; align-items:center; justify-content:flex-start; gap:10px; min-width:0; width:100%; }
  .club-logo { width:48px; height:48px; border-radius:0; object-fit:contain; border:1px solid var(--brand-border); background:var(--brand-surface-3); flex:none; }
  .name { font-weight:800; min-width:0; overflow-wrap:anywhere; text-align:left; }
  .state { font-size:.68rem; text-transform:uppercase; letter-spacing:.06em; color:var(--brand-muted); }
  .stats { display:grid; grid-template-columns:1fr 1fr; gap:6px 12px; border-top:1px solid var(--brand-border); padding-top:8px; }
  .stats div { display:grid; gap:1px; min-width:0; }
  .stats span { font-size:.66rem; text-transform:uppercase; letter-spacing:.06em; color:var(--brand-muted); }
  .stats strong { font-size:.82rem; overflow-wrap:anywhere; }
  .belt-dot { width:10px; height:10px; border-radius:50%; display:inline-block; margin-right:4px; }
  @media (max-width:700px){ .fighter-card{ padding:10px; } .stats{ grid-template-columns:1fr; } }
</style>
