<script lang="ts">
  import { onMount } from 'svelte';
  import { afterNavigate, goto } from '$app/navigation';
  import { api, type DemoIdentity } from '../api';
  import { landingForRole, session, switchDemoUser, uiRole } from '../auth';
  import {
    canUseOnPhone, currentRole, HOME_ROUTE, isPhoneWidth, itemsForRole,
    MOBILE_PRIMARY_HREFS, NAV_GROUPS, phoneHome, type NavItem, type Role,
  } from '../nav';

  let expanded = false;
  let mobileOpen = false;
  let role: Role = 'runner';
  let items: NavItem[] = [];
  let isMobile = false;
  let isPhone = false;
  let identities: DemoIdentity[] = [];
  let identityId = '';
  let switching = false;
  let switchError = '';
  let qaOpen = false;
  let activePath = '';

  afterNavigate(() => { activePath = window.location.pathname; });

  function visibleItems() {
    return itemsForRole(role).filter((item) => !isPhone || canUseOnPhone(item.href));
  }

  function syncViewport() {
    isMobile = window.innerWidth <= 820;
    isPhone = isPhoneWidth();
    items = visibleItems();
  }

  function syncNavWidth() {
    document.documentElement.dataset.nav = expanded ? 'wide' : 'compact';
  }

  $: mobileItems = (() => {
    const primary = MOBILE_PRIMARY_HREFS
      .map((href) => items.find((item) => item.href === href))
      .filter((item): item is NavItem => Boolean(item));
    const extras = items.filter((item) => item.group !== 'QA tools' && !primary.includes(item));
    return [...primary, ...extras].slice(0, 5);
  })();

  function syncSession() {
    const active = session();
    role = active ? uiRole(active.role) : currentRole();
    items = visibleItems();
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
    qaOpen = localStorage.getItem('j3-nav-qa') === '1';
    activePath = window.location.pathname;
    syncViewport();
    syncNavWidth();
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
      syncNavWidth();
    }
  }
  function closeMobile() {
    mobileOpen = false;
  }
  function isActive(href: string) {
    const path = activePath || window.location.pathname;
    return href === '/' ? path === '/' : path === href || path.startsWith(`${href}/`);
  }
  function toggleQa() {
    qaOpen = !qaOpen;
    localStorage.setItem('j3-nav-qa', qaOpen ? '1' : '0');
  }
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
      items = visibleItems();
      await goto(isPhone ? phoneHome(role) : landingForRole(next.role));
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
    <a class="brand-mark" href={HOME_ROUTE} aria-label="Public dashboard" on:click={closeMobile}>J3</a>
    <span class="mobile-title">J Cubed</span>
    <button class="hamburger" data-guide="nav-mobile-menu" aria-label="Open navigation" aria-expanded={mobileOpen} on:click={toggle}>☰</button>
  </header>

  {#if mobileOpen}
    <button class="scrim" aria-label="Close navigation" on:click={closeMobile}></button>
    <aside class="mobile-drawer" aria-label="Navigation">
      <div class="drawer-head">
        <span class="brand-name">J Cubed</span>
        <button class="hamburger" aria-label="Close navigation" on:click={closeMobile}>✕</button>
      </div>
      <nav aria-label="Main navigation">
        {#each NAV_GROUPS as group}
          {@const groupItemsForRole = items.filter((item) => item.group === group)}
          {#if groupItemsForRole.length}
            {#if group === 'QA tools'}
              <button class="group-toggle" aria-expanded={qaOpen} on:click={toggleQa}>{qaOpen ? '▾' : '▸'} {group}</button>
            {:else}
              <div class="group-label">{group}</div>
            {/if}
            {#if group !== 'QA tools' || qaOpen}
              {#each groupItemsForRole as item (item.href)}
                <a class="nav-item" class:active={isActive(item.href)} aria-current={isActive(item.href) ? 'page' : undefined}
                  href={item.href} on:click={closeMobile}>
                  <span class="icon">{item.icon}</span><span>{item.label}</span>
                </a>
              {/each}
            {/if}
          {/if}
        {/each}
      </nav>
      {#if isPhone}
        <p class="phone-note">Phone view shows public, participant, club and guardian screens. Scoring, setup and audit tools are available on tablet or desktop.</p>
      {/if}
      <button class="help-btn" data-guide="nav-help-full" aria-label="Start guided tour" on:click={startTour}>❓ Help & guided tour</button>
      <div class="role-picker">
        <span class="muted small">Demo identity · sign in as</span>
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
  <nav class="mobile-bottom" aria-label="Quick navigation">
    {#each mobileItems as item (item.href)}
      <a class="quick-item" class:active={isActive(item.href)} aria-current={isActive(item.href) ? 'page' : undefined}
        href={item.href} title={item.label}>
        <span class="icon">{item.icon}</span><span class="quick-label">{item.label}</span>
      </a>
    {/each}
    <button class="quick-item" data-guide="nav-mobile-more" aria-label="More navigation" aria-expanded={mobileOpen} on:click={toggle}>
      <span class="icon">☰</span><span class="quick-label">More</span>
    </button>
  </nav>
{:else}
  <aside class:collapsed={!expanded} class="nav-shell">
    <div class="nav-head">
      <a class="brand-mark" href={HOME_ROUTE} aria-label="Public dashboard">J3</a>
      {#if expanded}<span class="brand-name">J Cubed</span>{/if}
      <button class="hamburger" aria-label={expanded ? 'Collapse navigation' : 'Expand navigation'} aria-expanded={expanded} on:click={toggle}>
        {expanded ? '✕' : '☰'}
      </button>
    </div>
    <button class="help-btn" data-guide="nav-help-compact" aria-label="Start guided tour" title="Guided tour" on:click={startTour}>
      ❓{#if expanded}<span>Help & guided tour</span>{/if}
    </button>

    <nav aria-label="Main navigation">
      {#each NAV_GROUPS as group}
        {@const groupItemsForRole = items.filter((item) => item.group === group)}
        {#if groupItemsForRole.length}
          {#if expanded}
            {#if group === 'QA tools'}
              <button class="group-toggle" aria-expanded={qaOpen} on:click={toggleQa}>{qaOpen ? '▾' : '▸'} {group}</button>
            {:else}
              <div class="group-label">{group}</div>
            {/if}
          {:else}
            <div class="group-divider" aria-hidden="true"></div>
          {/if}
          {#if expanded && group === 'QA tools' && !qaOpen}
            <!-- collapsed by default: QA tools stay out of the event-day way -->
          {:else}
            {#each groupItemsForRole as item (item.href)}
              <a class="nav-item" class:icon-only={!expanded} class:active={isActive(item.href)}
                aria-current={isActive(item.href) ? 'page' : undefined}
                href={item.href} title={item.label} aria-label={item.label}>
                <span class="icon">{item.icon}</span>
                {#if expanded}<span class="label">{item.label}</span>{/if}
              </a>
            {/each}
          {/if}
        {/if}
      {/each}
    </nav>

    {#if expanded}
      <div class="role-picker">
        <span class="muted small">Demo identity · sign in as</span>
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
    position:fixed; top:0; left:0; bottom:0; z-index:800; width:244px;
    background:var(--brand-surface-2); border-right:1px solid var(--brand-border);
    display:flex; flex-direction:column; padding:18px 12px; gap:10px;
    box-shadow:8px 0 30px rgba(0,0,0,.25); transition:width .2s ease;
  }
  .nav-shell.collapsed { width:76px; padding:18px 10px; align-items:center; }
  .nav-head { display:flex; align-items:center; gap:10px; min-height:46px; padding:2px 4px 12px; }
  .nav-shell.collapsed .nav-head { flex-direction:column; gap:8px; width:100%; padding:2px 0 10px; }
  .brand-name { font-family:var(--font-display); font-weight:800; white-space:nowrap; font-size:1.05rem; }
  .hamburger { background:var(--brand-surface-3); border:1px solid var(--brand-border); color:var(--brand-ink); border-radius:9px; padding:8px 10px; cursor:pointer; line-height:1; font-size:1rem; min-width:38px; }
  .nav-shell.collapsed .hamburger { margin:0 auto; }
  nav { display:grid; gap:2px; overflow-y:auto; flex:1; width:100%; scrollbar-width:thin; }
  .group-label, .group-toggle { margin:12px 8px 4px; font-size:.65rem; letter-spacing:.14em; text-transform:uppercase; color:var(--brand-muted); font-weight:900; }
  .group-toggle { width:calc(100% - 16px); text-align:left; background:none; border:0; cursor:pointer; padding:4px 0; }
  .group-toggle:hover { color:var(--brand-ink); }
  .group-divider { height:1px; background:var(--brand-border); margin:8px 10px 5px; }
  .nav-item { display:flex; gap:12px; align-items:center; min-height:42px; padding:10px 12px; border-radius:11px; color:var(--brand-muted); text-decoration:none; font-size:.86rem; white-space:nowrap; border-left:3px solid transparent; }
  .nav-item .icon { font-size:1.1rem; width:24px; text-align:center; flex:none; }
  .nav-item:hover, .nav-item:focus-visible { background:var(--brand-surface-3); color:var(--brand-ink); text-decoration:none; }
  .nav-item.active { background:var(--brand-surface-3); color:var(--brand-ink); border-left-color:var(--brand-primary); font-weight:800; }
  .nav-item.icon-only { justify-content:center; padding:12px 0; }
  .nav-shell.collapsed .nav-item { padding-left:9px; }
  .help-btn { display:flex; gap:10px; align-items:center; width:100%; padding:10px 12px; border-radius:10px; border:1px solid var(--brand-border); background:var(--brand-surface-3); color:var(--brand-ink); cursor:pointer; font-size:.86rem; }
  .nav-shell.collapsed .help-btn { justify-content:center; padding:11px 0; }
  .role-picker { display:grid; gap:7px; padding:14px 8px 4px; border-top:1px solid var(--brand-border); }
  .role-picker select { padding:9px; border-radius:9px; background:var(--brand-surface-3); color:var(--brand-ink); border:1px solid var(--brand-border); }
  .collapsed-role { width:38px; height:38px; border-radius:50%; display:grid; place-items:center; background:var(--brand-primary); color:#fff; font-weight:800; margin-top:auto; }
  .mobile-bar { position:fixed; top:0; left:0; right:0; height:60px; z-index:810; display:flex; align-items:center; gap:12px; padding:0 14px; background:var(--brand-surface-2); border-bottom:1px solid var(--brand-border); }
  .mobile-title { font-family:var(--font-display); font-weight:800; }
  .mobile-bar .hamburger { margin-left:auto; }
  .mobile-bottom {
    position:fixed; left:0; right:0; bottom:0; z-index:815;
    display:grid; grid-template-columns:repeat(auto-fit, minmax(0,1fr)); gap:2px;
    padding:6px 6px calc(6px + env(safe-area-inset-bottom, 0px));
    background:var(--brand-surface-2); border-top:1px solid var(--brand-border);
    box-shadow:0 -8px 24px rgba(0,0,0,.28);
  }
  .quick-item {
    display:grid; place-items:center; gap:2px; min-width:0; padding:5px 2px;
    border:0; border-radius:10px; background:transparent; color:var(--brand-muted);
    text-decoration:none; font-size:.62rem; font-weight:700; cursor:pointer;
  }
  .quick-item .icon { font-size:1.05rem; line-height:1; }
  .quick-label { max-width:100%; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
  .quick-item.active { color:var(--brand-ink); background:var(--brand-surface-3); box-shadow:inset 0 -2px 0 var(--brand-primary); }
  .quick-item:hover, .quick-item:focus-visible { color:var(--brand-ink); text-decoration:none; }
  .scrim { position:fixed; inset:0; z-index:820; background:rgba(2,6,23,.65); border:0; cursor:pointer; }
  .mobile-drawer { position:fixed; top:0; right:0; bottom:0; width:min(86vw, 340px); z-index:830; background:var(--brand-surface-2); border-left:1px solid var(--brand-border); padding:18px; display:flex; flex-direction:column; gap:12px; box-shadow:-16px 0 44px rgba(0,0,0,.45); animation:slide .22s ease-out; }
  .drawer-head { display:flex; align-items:center; justify-content:space-between; }
  .mobile-drawer nav { display:grid; gap:3px; overflow-y:auto; flex:1; }
  .mobile-drawer .nav-item { min-height:46px; }
  .mobile-drawer .group-label, .mobile-drawer .group-toggle { margin-top:10px; }
  .phone-note { margin:0; padding:10px 12px; border-radius:9px; background:var(--brand-surface-3); border:1px solid var(--brand-border); color:var(--brand-muted); font-size:.74rem; line-height:1.4; }
  @keyframes slide { from { transform:translateX(100%); } to { transform:none; } }
</style>
