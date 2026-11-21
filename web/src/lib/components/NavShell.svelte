<script lang="ts">
  import { onMount } from 'svelte';
  import { currentRole, itemsForRole, setRole, type Role } from '../nav';

  let expanded = false;
  let role: Role = 'runner';
  let items: ReturnType<typeof itemsForRole> = [];

  onMount(() => {
    role = currentRole();
    items = itemsForRole(role);
    const saved = localStorage.getItem('j3-nav-expanded');
    expanded = saved === '1';
  });

  function toggle() {
    expanded = !expanded;
    localStorage.setItem('j3-nav-expanded', expanded ? '1' : '0');
  }
  function changeRole(value: string) {
    role = value as Role;
    setRole(role);
    items = itemsForRole(role);
  }
</script>

<aside class:collapsed={!expanded} class="nav-shell">
  <div class="nav-head">
    <a class="brand-mark" href="/" aria-label="Dashboard">J3</a>
    {#if expanded}<span class="brand-name">J Cubed</span>{/if}
    <button class="hamburger" aria-label={expanded ? 'Collapse navigation' : 'Expand navigation'} aria-expanded={expanded} on:click={toggle}>
      {expanded ? '✕' : '☰'}
    </button>
  </div>

  <nav aria-label="Main navigation">
    {#each items as item (item.href)}
      <a class="nav-item" href={item.href} title={item.label} aria-label={item.label}>
        <span class="icon">{item.icon}</span>
        {#if expanded}<span>{item.label}</span>{/if}
      </a>
    {/each}
  </nav>

  {#if expanded}
    <div class="role-picker">
      <span class="muted small">Viewing as</span>
      <select value={role} on:change={(event) => changeRole((event.target as HTMLSelectElement).value)} aria-label="Account role">
        <option value="runner">Runner / organiser</option>
        <option value="official">Official / moderator</option>
        <option value="club_leader">Club leader</option>
        <option value="fighter">Fighter</option>
        <option value="spectator">Spectator</option>
      </select>
      <span class="muted small">{items.length} available views</span>
    </div>
  {/if}
</aside>

<style>
  .nav-shell {
    position:fixed; top:0; left:0; bottom:0; z-index:800; width:230px;
    background:var(--brand-surface-2); border-right:1px solid var(--brand-border);
    display:flex; flex-direction:column; padding:14px 10px; gap:8px;
    box-shadow:8px 0 30px rgba(0,0,0,.25); transition:width .2s ease;
  }
  .nav-shell.collapsed { width:64px; }
  .nav-head { display:flex; align-items:center; gap:8px; padding:4px 4px 12px; }
  .brand-name { font-family:var(--font-display); font-weight:800; white-space:nowrap; }
  .hamburger { margin-left:auto; background:transparent; border:1px solid var(--brand-border); color:var(--brand-ink); border-radius:8px; padding:5px 8px; cursor:pointer; }
  nav { display:grid; gap:4px; overflow-y:auto; flex:1; }
  .nav-item { display:flex; gap:10px; align-items:center; padding:9px 10px; border-radius:10px; color:var(--brand-muted); text-decoration:none; font-size:.88rem; white-space:nowrap; }
  .nav-item:hover, .nav-item:focus-visible { background:var(--brand-surface-3); color:var(--brand-ink); text-decoration:none; }
  .nav-item .icon { font-size:1.1rem; width:22px; text-align:center; flex:none; }
  .role-picker { display:grid; gap:6px; padding:10px 6px; border-top:1px solid var(--brand-border); }
  .role-picker select { padding:7px; border-radius:8px; background:var(--brand-surface-3); color:var(--brand-ink); border:1px solid var(--brand-border); }
  @media (max-width: 700px) { .nav-shell { width:60px; } .brand-name, .role-picker { display:none; } }
</style>
