<script lang="ts">
  import { onMount } from 'svelte';
  import { api, type DemoIdentity } from '../api';
  import { landingForRole, session, switchDemoUser, uiRole } from '../auth';
  import { currentRole, itemsForRole, type Role } from '../nav';

  let expanded = false;
  let mobileOpen = false;
  let role: Role = 'runner';
  let items: ReturnType<typeof itemsForRole> = [];
  let isMobile = false;
  let identities: DemoIdentity[] = [];
  let identityId = '';
  let switching = false;
  let switchError = '';

  function syncViewport() {
    isMobile = window.innerWidth <= 820;
  }

  function syncSession() {
    const active = session();
    role = active ? uiRole(active.role) : currentRole();
    items = itemsForRole(role);
    identityId = active?.userId ?? '';
  }

  async function loadIdentities() {
    try {
      const response = await api.identities();
      identities = response.identities;
      if (response.current) identityId = response.current.id;
    } catch {
      identities = [];
    }
  }

  onMount(() => {
    expanded = localStorage.getItem('j3-nav-expanded') === '1';
    syncViewport();
    syncSession();
    void loadIdentities();
    window.addEventListener('resize', syncViewport);
    window.addEventListener('j3-session-changed', syncSession);
    return () => {
      window.removeEventListener('resize', syncViewport);
      window.removeEventListener('j3-session-changed', syncSession);
    };
  });

  function toggle() {
    if (isMobile) {
      mobileOpen = !mobileOpen;
    } else {
      expanded = !expanded;
      localStorage.setItem('j3-nav-expanded', expanded ? '1' : '0');
    }
  }
  function closeMobile() { mobileOpen = false; }
  function startTour() {
    mobileOpen = false;
    window.dispatchEvent(new CustomEvent('j3-start-tour', { detail: { role } }));
  }
  async function changeIdentity(value: string) {
    if (!value || value === identityId || switching) return;
    switching = true;
    switchError = '';
    try {
      const next = await switchDemoUser(value);
      role = uiRole(next.role);
      items = itemsForRole(role);
      window.location.assign(landingForRole(next.role));
    } catch (e) {
      switchError = e instanceof Error ? e.message : String(e);
    } finally {
      switching = false;
    }
  }
</script>

<svelte:window on:keydown={(event) => { if (event.key === 'Escape') closeMobile(); }} />

{#if isMobile}
  <header class="mobile-bar">
    <a class="brand-mark" href="/" aria-label="Dashboard" on:click={closeMobile}>J3</a>
    <span class="mobile-title">J Cubed</span>
    <button class="hamburger" aria-label="Open navigation" aria-expanded={mobileOpen} on:click={toggle}>☰</button>
  </header>

  {#if mobileOpen}
    <button class="scrim" aria-label="Close navigation" on:click={closeMobile}></button>
    <aside class="mobile-drawer" aria-label="Navigation">
      <div class="drawer-head">
        <span class="brand-name">J Cubed</span>
        <button class="hamburger" aria-label="Close navigation" on:click={closeMobile}>✕</button>
      </div>
      <nav>
        {#each items as item (item.href)}
          <a class="nav-item" href={item.href} on:click={closeMobile}><span class="icon">{item.icon}</span><span>{item.label}</span></a>
        {/each}
      </nav>
      <button class="help-btn" aria-label="Start guided tour" on:click={startTour}>❓ Help & guided tour</button>
      <div class="role-picker">
        <span class="muted small">QA demo · sign in as</span>
        <select
          value={identityId}
          disabled={switching || identities.length === 0}
          on:change={(event) => changeIdentity((event.target as HTMLSelectElement).value)}
          aria-label="Sign in as demo user"
        >
          {#if identities.length === 0}
            <option value="">Loading demo accounts…</option>
          {:else}
            {#each identities as identity (identity.id)}
              <option value={identity.id}>{identity.name} · {identity.roleLabel}</option>
            {/each}
          {/if}
        </select>
        {#if switchError}<span class="muted small">{switchError}</span>{/if}
      </div>
    </aside>
  {/if}
{:else}
  <aside class:collapsed={!expanded} class="nav-shell">
    <div class="nav-head">
      <a class="brand-mark" href="/" aria-label="Dashboard">J3</a>
      {#if expanded}<span class="brand-name">J Cubed</span>{/if}
      <button class="hamburger" aria-label={expanded ? 'Collapse navigation' : 'Expand navigation'} aria-expanded={expanded} on:click={toggle}>
        {expanded ? '✕' : '☰'}
      </button>
    </div>
    <button class="help-btn" aria-label="Start guided tour" title="Guided tour" on:click={startTour}>❓{#if expanded}<span>Help & tour</span>{/if}</button>

    <nav aria-label="Main navigation">
      {#each items as item (item.href)}
        <a class="nav-item" class:icon-only={!expanded} href={item.href} title={item.label} aria-label={item.label}>
          <span class="icon">{item.icon}</span>
          {#if expanded}<span class="label">{item.label}</span>{/if}
        </a>
      {/each}
    </nav>

    {#if expanded}
      <button class="help-btn" aria-label="Start guided tour" on:click={startTour}>❓ Help & guided tour</button>
      <div class="role-picker">
        <span class="muted small">QA demo · sign in as</span>
        <select
          value={identityId}
          disabled={switching || identities.length === 0}
          on:change={(event) => changeIdentity((event.target as HTMLSelectElement).value)}
          aria-label="Sign in as demo user"
        >
          {#if identities.length === 0}
            <option value="">Loading demo accounts…</option>
          {:else}
            {#each identities as identity (identity.id)}
              <option value={identity.id}>{identity.name} · {identity.roleLabel}</option>
            {/each}
          {/if}
        </select>
        <span class="muted small">{items.length} available views</span>
        {#if switchError}<span class="muted small">{switchError}</span>{/if}
      </div>
    {:else}
      <div class="collapsed-role" title="Viewing as {role}">{role.slice(0, 1).toUpperCase()}</div>
    {/if}
  </aside>
{/if}

<style>
  .nav-shell {
    position:fixed; top:0; left:0; bottom:0; z-index:800; width:228px;
    background:var(--brand-surface-2); border-right:1px solid var(--brand-border);
    display:flex; flex-direction:column; padding:18px 12px; gap:14px;
    box-shadow:8px 0 30px rgba(0,0,0,.25); transition:width .2s ease;
  }
  .nav-shell.collapsed { width:76px; padding:18px 10px; align-items:center; }
  .nav-head { display:flex; align-items:center; gap:10px; min-height:46px; padding:2px 4px 14px; }
  .nav-shell.collapsed .nav-head { flex-direction:column; gap:8px; width:100%; padding:2px 0 14px; }
  .brand-name { font-family:var(--font-display); font-weight:800; white-space:nowrap; font-size:1.05rem; }
  .hamburger { background:var(--brand-surface-3); border:1px solid var(--brand-border); color:var(--brand-ink); border-radius:9px; padding:8px 10px; cursor:pointer; line-height:1; font-size:1rem; min-width:38px; }
  .nav-shell.collapsed .hamburger { margin:0 auto; }
  nav { display:grid; gap:4px; overflow-y:auto; flex:1; width:100%; scrollbar-width:thin; }
  .nav-item { display:flex; gap:12px; align-items:center; min-height:44px; padding:10px 12px; border-radius:11px; color:var(--brand-muted); text-decoration:none; font-size:.88rem; white-space:nowrap; }
  .nav-item .icon { font-size:1.15rem; width:24px; text-align:center; flex:none; }
  .nav-item:hover, .nav-item:focus-visible { background:var(--brand-surface-3); color:var(--brand-ink); text-decoration:none; }
  .nav-item.icon-only { justify-content:center; padding:12px 0; }
  .help-btn { display:flex; gap:10px; align-items:center; width:100%; padding:10px 12px; border-radius:10px; border:1px solid var(--brand-border); background:var(--brand-surface-3); color:var(--brand-ink); cursor:pointer; font-size:.86rem; }
  .nav-shell.collapsed .help-btn { justify-content:center; padding:11px 0; }
  .role-picker { display:grid; gap:7px; padding:14px 8px 4px; border-top:1px solid var(--brand-border); }
  .role-picker select { padding:9px; border-radius:9px; background:var(--brand-surface-3); color:var(--brand-ink); border:1px solid var(--brand-border); }
  .collapsed-role { width:38px; height:38px; border-radius:50%; display:grid; place-items:center; background:var(--brand-primary); color:#fff; font-weight:800; margin-top:auto; }
  .mobile-bar { position:fixed; top:0; left:0; right:0; height:60px; z-index:810; display:flex; align-items:center; gap:12px; padding:0 14px; background:var(--brand-surface-2); border-bottom:1px solid var(--brand-border); }
  .mobile-title { font-family:var(--font-display); font-weight:800; }
  .mobile-bar .hamburger { margin-left:auto; }
  .scrim { position:fixed; inset:0; z-index:820; background:rgba(2,6,23,.65); border:0; cursor:pointer; }
  .mobile-drawer { position:fixed; top:0; right:0; bottom:0; width:min(86vw, 340px); z-index:830; background:var(--brand-surface-2); border-left:1px solid var(--brand-border); padding:18px; display:flex; flex-direction:column; gap:14px; box-shadow:-16px 0 44px rgba(0,0,0,.45); animation:slide .22s ease-out; }
  .drawer-head { display:flex; align-items:center; justify-content:space-between; }
  .mobile-drawer nav { display:grid; gap:4px; overflow-y:auto; flex:1; }
  .mobile-drawer .nav-item { min-height:48px; }
  @keyframes slide { from { transform:translateX(100%); } to { transform:none; } }
</style>
