<script lang="ts">
  import { onMount } from 'svelte';
  import { afterNavigate, goto } from '$app/navigation';
  import { api, type DemoIdentity } from '../api';
  import { landingForRole, session, switchDemoUser, uiRole } from '../auth';
  import { eventTitle } from '../event-title';
  import { appMode, setAppMode } from '../demo-replay';
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
  let navWidth = 244;
  let resizing = false;

  afterNavigate(() => { activePath = window.location.pathname; });

  function visibleItems() {
    return itemsForRole(role).filter((item) => !isPhone || canUseOnPhone(item.href));
  }

  function syncViewport() {
    isMobile = window.innerWidth <= 820;
    isPhone = isPhoneWidth();
    items = visibleItems();
    syncNavWidth();
  }

  function syncNavWidth() {
    const width = isMobile ? 0 : expanded ? navWidth : 76;
    const scale = Math.min(1.08, Math.max(0.82, width / 244));
    document.documentElement.dataset.nav = expanded ? 'wide' : 'compact';
    document.documentElement.style.setProperty('--nav-width', `${width}px`);
    document.documentElement.style.setProperty('--nav-scale', scale.toFixed(2));
  }

  function clampWidth(value: number) {
    return Math.max(180, Math.min(380, Math.round(value)));
  }

  function persistWidth() {
    localStorage.setItem('j3-nav-width', String(navWidth));
    syncNavWidth();
  }

  function startResize(event: PointerEvent) {
    event.preventDefault();
    resizing = true;
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';
    window.addEventListener('pointermove', resizeMove);
    window.addEventListener('pointerup', stopResize);
  }

  function resizeMove(event: PointerEvent) {
    if (!resizing) return;
    navWidth = clampWidth(event.clientX);
    persistWidth();
  }

  function stopResize() {
    resizing = false;
    document.body.style.cursor = '';
    document.body.style.userSelect = '';
    window.removeEventListener('pointermove', resizeMove);
    window.removeEventListener('pointerup', stopResize);
  }

  function nudgeResize(event: KeyboardEvent) {
    if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight') return;
    event.preventDefault();
    navWidth = clampWidth(navWidth + (event.key === 'ArrowRight' ? 16 : -16));
    persistWidth();
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
    navWidth = clampWidth(Number(localStorage.getItem('j3-nav-width')) || 244);
    qaOpen = localStorage.getItem('j3-nav-qa') === '1';
    activePath = window.location.pathname;
    syncViewport();
    syncNavWidth();
    syncSession();
    void loadIdentities();
    window.addEventListener('resize', syncViewport);
    window.addEventListener('j3-session-changed', syncSession);
    return () => {
      stopResize();
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
    <div class="mobile-bar-row">
      <a class="brand-mark" href={HOME_ROUTE} aria-label="Public dashboard" on:click={closeMobile}>J3</a>
      <div class="mode-switch compact" role="group" aria-label="Application mode">
        <button class:active={$appMode === 'live'} aria-label="Switch to live event mode" aria-pressed={$appMode === 'live'} on:click={() => setAppMode('live')}>●</button>
        <button class:active={$appMode === 'sim'} aria-label="Switch to simulated tournament mode" aria-pressed={$appMode === 'sim'} on:click={() => setAppMode('sim')}>🎬</button>
      </div>
      <button class="hamburger" data-guide="nav-mobile-menu" aria-label="Open navigation" aria-expanded={mobileOpen} on:click={toggle}>☰</button>
    </div>
    <div class="mobile-event" title={$eventTitle}>
      <span class="mobile-event-label">Event</span>
      <span class="mobile-event-title">{$eventTitle}</span>
    </div>
  </header>

  {#if mobileOpen}
    <button class="scrim" aria-label="Close navigation" on:click={closeMobile}></button>
    <aside class="mobile-drawer" aria-label="Navigation">
      <div class="drawer-head" title={$eventTitle}>
        <span class="event-card-title">{$eventTitle}</span>
        <button class="hamburger" aria-label="Close navigation" on:click={closeMobile}>✕</button>
      </div>
      <div class="mode-switch drawer" role="group" aria-label="Application mode">
        <button class:active={$appMode === 'live'} aria-pressed={$appMode === 'live'} on:click={() => { setAppMode('live'); closeMobile(); }}>● <span>Live event</span></button>
        <button class:active={$appMode === 'sim'} aria-pressed={$appMode === 'sim'} on:click={() => { setAppMode('sim'); closeMobile(); }}>🎬 <span>Simulated</span></button>
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
      <button class="hamburger" aria-label={expanded ? 'Collapse navigation' : 'Expand navigation'} aria-expanded={expanded} on:click={toggle}>
        {expanded ? '✕' : '☰'}
      </button>
    </div>
    {#if expanded}
      <div class="event-card" title={$eventTitle}>
        <span class="event-card-label">Event</span>
        <span class="event-card-title">{$eventTitle}</span>
      </div>
    {/if}
    <div class="mode-switch" class:compact={!expanded} role="group" aria-label="Application mode">
      <button class:active={$appMode === 'live'} aria-label="Switch to live event mode" aria-pressed={$appMode === 'live'} title="Live event" on:click={() => setAppMode('live')}>
        ● {#if expanded}<span>Live event</span>{/if}
      </button>
      <button class:active={$appMode === 'sim'} aria-label="Switch to simulated tournament mode" aria-pressed={$appMode === 'sim'} title="Simulated tournament" on:click={() => setAppMode('sim')}>
        🎬 {#if expanded}<span>Simulated</span>{/if}
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
    {#if expanded && !isMobile}
      <button type="button" class="resize-handle" aria-label="Resize navigation"
        title="Drag or use arrow keys to resize navigation" on:pointerdown={startResize} on:keydown={nudgeResize}></button>
    {/if}
  </aside>
{/if}

<style>
  .nav-shell {
    position:fixed; top:0; left:0; bottom:0; z-index:800;
    width:var(--nav-width, 244px); font-size:calc(1rem * var(--nav-scale, 1));
    background:var(--brand-surface-2); border-right:1px solid var(--brand-border);
    display:flex; flex-direction:column; padding:18px 12px; gap:10px;
    box-shadow:8px 0 30px rgba(0,0,0,.25); transition:width .12s ease;
  }
  .nav-shell.collapsed { width:76px; padding:18px 10px; align-items:center; }
  .resize-handle {
    position:absolute; top:0; right:-3px; bottom:0; width:8px; cursor:col-resize;
    background:transparent; border:0; border-right:3px solid transparent; padding:0; transition:border-color .16s ease;
  }
  .resize-handle:hover, .resize-handle:focus-visible, .resize-handle:active {
    border-right-color:var(--brand-primary); outline:none;
    box-shadow:6px 0 18px rgba(0,0,0,.3);
  }
  .nav-head { display:flex; align-items:center; justify-content:space-between; gap:10px; min-height:46px; padding:2px 4px 8px; }
  .nav-shell.collapsed .nav-head { flex-direction:column; gap:8px; width:100%; padding:2px 0 10px; }
  .event-card { display:grid; gap:2px; min-width:0; padding:10px 12px; border-radius:12px; background:var(--brand-surface-3); border:1px solid var(--brand-border); }
  .event-card-label, .mobile-event-label { text-transform:uppercase; letter-spacing:.14em; font-size:.58rem; font-weight:900; color:var(--brand-muted); }
  .event-card-title { font-family:var(--font-display); font-weight:800; font-size:.88em; line-height:1.25; min-width:0; overflow:hidden; overflow-wrap:anywhere; display:-webkit-box; -webkit-line-clamp:2; line-clamp:2; -webkit-box-orient:vertical; }
  .mode-switch { display:grid; grid-template-columns:1fr 1fr; gap:4px; padding:4px; border-radius:12px; background:var(--brand-surface-3); border:1px solid var(--brand-border); }
  .mode-switch button { display:flex; align-items:center; justify-content:center; gap:5px; min-width:0; padding:7px 6px; border-radius:9px; border:1px solid transparent; background:transparent; color:var(--brand-muted); cursor:pointer; font-size:.72em; font-weight:800; white-space:nowrap; }
  .mode-switch button:hover { color:var(--brand-ink); }
  .mode-switch button.active { border-color:var(--brand-accent); background:var(--brand-primary-fill); color:var(--brand-on-primary); box-shadow:0 0 14px color-mix(in srgb, var(--brand-accent) 30%, transparent); }
  .mode-switch.compact { grid-template-columns:1fr; width:52px; margin:0 auto; }
  .mode-switch.compact button { padding:8px 0; font-size:.95em; }
  .mode-switch.compact button.active { box-shadow:inset 3px 0 0 var(--brand-accent); }
  .mode-switch.drawer { grid-template-columns:1fr 1fr; }
  .hamburger { background:var(--brand-surface-3); border:1px solid var(--brand-border); color:var(--brand-ink); border-radius:9px; padding:8px 10px; cursor:pointer; line-height:1; font-size:1em; min-width:38px; }
  .nav-shell.collapsed .hamburger { margin:0 auto; }
  nav { display:grid; gap:2px; overflow-y:auto; flex:1; width:100%; scrollbar-width:thin; }
  .group-label, .group-toggle { margin:12px 8px 4px; font-size:.65em; letter-spacing:.14em; text-transform:uppercase; color:var(--brand-muted); font-weight:900; }
  .group-toggle { width:calc(100% - 16px); text-align:left; background:none; border:0; cursor:pointer; padding:4px 0; }
  .group-toggle:hover { color:var(--brand-ink); }
  .group-divider { height:1px; background:var(--brand-border); margin:8px 10px 5px; }
  .nav-item { display:flex; gap:12px; align-items:center; min-height:42px; padding:10px 12px; border-radius:11px; color:var(--brand-muted); text-decoration:none; font-size:.86em; white-space:nowrap; border-left:3px solid transparent; }
  .nav-item .icon { font-size:1.1em; width:24px; text-align:center; flex:none; }
  .nav-item:hover, .nav-item:focus-visible { background:var(--brand-surface-3); color:var(--brand-ink); text-decoration:none; }
  .nav-item.active {
    background:linear-gradient(90deg, color-mix(in srgb, var(--brand-primary) 30%, var(--brand-surface-3)), var(--brand-surface-3));
    color:var(--brand-ink); border-left-color:var(--brand-primary); font-weight:800;
    box-shadow:inset 4px 0 0 var(--brand-accent), 0 0 0 1px color-mix(in srgb, var(--brand-primary) 45%, transparent);
  }
  :global([data-theme='street']) .nav-item.active {
    border-left-color:var(--brand-accent);
    box-shadow:inset 4px 0 0 var(--brand-accent), 0 0 20px color-mix(in srgb, var(--brand-accent) 35%, transparent);
  }
  .nav-item.icon-only { justify-content:center; padding:12px 0; }
  .nav-shell.collapsed .nav-item { padding-left:9px; }
  .help-btn { display:flex; gap:10px; align-items:center; width:100%; padding:10px 12px; border-radius:10px; border:1px solid var(--brand-border); background:var(--brand-surface-3); color:var(--brand-ink); cursor:pointer; font-size:.86em; }
  .nav-shell.collapsed .help-btn { justify-content:center; padding:11px 0; }
  .role-picker { display:grid; gap:7px; min-width:0; padding:14px 8px 4px; border-top:1px solid var(--brand-border); overflow:hidden; }
  .role-picker select { width:100%; max-width:100%; min-width:0; padding:9px; border-radius:9px; background:var(--brand-surface-3); color:var(--brand-ink); border:1px solid var(--brand-border); font-size:.86em; }
  .collapsed-role { width:38px; height:38px; border-radius:50%; display:grid; place-items:center; background:var(--brand-primary); color:var(--brand-on-primary); font-weight:800; margin-top:auto; }
  .mobile-bar { position:fixed; top:0; left:0; right:0; z-index:810; display:grid; gap:3px; padding:7px 14px 8px; background:var(--brand-surface-2); border-bottom:1px solid var(--brand-border); }
  .mobile-bar-row { display:flex; align-items:center; justify-content:space-between; gap:12px; min-height:34px; }
  .mobile-event { display:grid; gap:1px; min-width:0; }
  .mobile-event-title { font-family:var(--font-display); font-weight:800; font-size:.8rem; line-height:1.2; min-width:0; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
  .mobile-bar .hamburger { margin-left:auto; }
  .mobile-bar .mode-switch.compact { display:flex; width:auto; }
  .mobile-bar .mode-switch.compact button { min-width:36px; padding:7px 8px; }
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
  .quick-item.active {
    color:var(--brand-ink); background:var(--brand-surface-3);
    box-shadow:inset 0 -3px 0 var(--brand-accent), 0 0 0 1px color-mix(in srgb, var(--brand-primary) 45%, transparent);
  }
  .quick-item:hover, .quick-item:focus-visible { color:var(--brand-ink); text-decoration:none; }
  .scrim { position:fixed; inset:0; z-index:820; background:rgba(2,6,23,.65); border:0; cursor:pointer; }
  .mobile-drawer { position:fixed; top:0; right:0; bottom:0; width:min(86vw, 340px); z-index:830; background:var(--brand-surface-2); border-left:1px solid var(--brand-border); padding:18px; display:flex; flex-direction:column; gap:12px; box-shadow:-16px 0 44px rgba(0,0,0,.45); animation:slide .22s ease-out; }
  .drawer-head { display:flex; align-items:center; justify-content:space-between; gap:10px; min-width:0; }
  .drawer-head .event-card-title { flex:1; }
  .mobile-drawer nav { display:grid; gap:3px; overflow-y:auto; flex:1; }
  .mobile-drawer .nav-item { min-height:46px; }
  .mobile-drawer .group-label, .mobile-drawer .group-toggle { margin-top:10px; }
  .phone-note { margin:0; padding:10px 12px; border-radius:9px; background:var(--brand-surface-3); border:1px solid var(--brand-border); color:var(--brand-muted); font-size:.74rem; line-height:1.4; }
  @keyframes slide { from { transform:translateX(100%); } to { transform:none; } }
</style>
