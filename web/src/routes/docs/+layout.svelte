<script lang="ts">
  import { onMount } from 'svelte';
  import { afterNavigate } from '$app/navigation';
  import { docs, docHref } from '../../lib/docs';
  import type { DocEntry, SearchEntry } from '../../lib/docs';

  let entries: DocEntry[] = [];
  let searchIndex: SearchEntry[] = [];
  let query = '';
  let menuOpen = false;
  let loadError = '';
  let activePath = '';

  $: groups = Array.from(new Set(entries.map((entry) => entry.group)));
  $: results = query.trim()
    ? searchIndex
        .filter((entry) => {
          const haystack = `${entry.title} ${entry.section} ${entry.text}`.toLowerCase();
          return query.toLowerCase().split(/\s+/).every((word) => haystack.includes(word));
        })
        .slice(0, 24)
    : [];
  $: pageSections = Array.from(
    new Map(
      searchIndex
        .filter((entry) => activePath && entry.url.startsWith(`${activePath}#`) && entry.title.includes(' · '))
        .map((entry) => [entry.url, { label: entry.title.split(' · ').slice(1).join(' · '), url: entry.url }]),
    ).values(),
  );
  afterNavigate(() => (activePath = window.location.pathname));

  onMount(() => {
    void docs.manifest().then((data) => (entries = data.docs)).catch((error) => (loadError = String(error)));
    void docs.search().then((data) => (searchIndex = data)).catch(() => {});
  });
</script>

<svelte:head>
  <meta name="description" content="J3 documentation: organiser guides, developer references and a live API explorer." />
</svelte:head>

<div class="docs-shell">
  <aside class:open={menuOpen} class="sidebar">
    <div class="side-head">
      <a class="logo" href="/docs" aria-label="J3 documentation home">J3 <span>/ docs</span></a>
      <button class="icon-btn mobile-only" aria-label="Close menu" on:click={() => (menuOpen = false)}>✕</button>
    </div>
    <label class="search">
      <span aria-hidden="true">🔎</span>
      <input aria-label="Search documentation" placeholder="Search guides and endpoints…" bind:value={query} />
    </label>

    {#if query.trim()}
      <div class="nav-group">Results <span class="count">{results.length}</span></div>
      <nav aria-label="Search results">
        {#each results as result}
          <a class="nav-link" href={result.url} on:click={() => (menuOpen = false)}>
            <strong>{result.title}</strong>
            <small>{result.text}</small>
          </a>
        {:else}
          <p class="empty">No matches. Try “pools”, “scoring” or “deploy”.</p>
        {/each}
      </nav>
    {:else}
      <div class="nav-group">Overview</div>
      <nav aria-label="Documentation pages">
        <a class="nav-link" class:active={activePath === '/docs'} href="/docs" on:click={() => (menuOpen = false)}>
          <strong>🏠 Documentation home</strong><small>Start here</small>
        </a>
      </nav>
      {#each groups as group}
        <div class="nav-group">{group}</div>
        <nav aria-label={group}>
          {#each entries.filter((entry) => entry.group === group) as entry}
            <a
              class="nav-link"
              class:active={activePath === docHref(entry) || (entry.slug === 'api' && activePath === '/apidocs')}
              href={docHref(entry)}
              on:click={() => (menuOpen = false)}
            >
              <strong>{entry.icon} {entry.title}</strong>
              <small>{entry.description}</small>
            </a>
          {/each}
        </nav>
      {/each}
      {#if pageSections.length}
        <div class="nav-group">Sections in this page</div>
        <nav aria-label="Sections in this guide">
          {#each pageSections as section (section.url)}
            <a class="nav-link section-link" href={section.url} on:click={() => (menuOpen = false)}>
              <strong>{section.label}</strong>
            </a>
          {/each}
        </nav>
      {/if}
      {#if loadError}<p class="empty">Could not load the page list: {loadError}</p>{/if}
    {/if}

    <div class="side-foot">
      <a href="/">← Back to J3</a>
      <span>v0.1 · local venue</span>
    </div>
  </aside>

  {#if menuOpen}<button class="scrim" aria-label="Close menu" on:click={() => (menuOpen = false)}></button>{/if}

  <main class="content">
    <header class="doc-topbar">
      <button class="icon-btn mobile-only" aria-label="Open menu" on:click={() => (menuOpen = true)}>☰</button>
      <span class="crumb">Documentation</span>
      <span class="spacer"></span>
      <a class="top-link" href="/docs/guide">Organiser guide</a>
      <a class="top-link" href="/docs/development">Developer guide</a>
      <a class="top-link primary" href="/docs/api">API reference ↗</a>
    </header>
    <div class="page-body"><slot /></div>
  </main>
</div>

<style>
  .docs-shell { display:grid; grid-template-columns:290px minmax(0,1fr); min-height:100vh; background:var(--brand-surface); }
  .sidebar { position:sticky; top:0; height:100vh; overflow-y:auto; padding:20px 16px; background:var(--brand-surface); border-right:1px solid var(--brand-border); }
  .side-head { display:flex; align-items:center; justify-content:space-between; margin-bottom:18px; }
  .logo { font-family:var(--font-display); font-size:1.5rem; font-weight:900; color:var(--brand-ink); text-decoration:none; }
  .logo span { font-family:var(--font); font-size:.8rem; font-weight:600; color:var(--brand-muted); }
  .search { display:flex; align-items:center; gap:8px; padding:9px 11px; border:1px solid var(--brand-border); border-radius:11px; background:var(--brand-surface-3); margin-bottom:18px; }
  .search input { flex:1; min-width:0; border:0; outline:0; background:transparent; color:var(--brand-ink); font:inherit; font-size:.82rem; }
  .nav-group { margin:18px 6px 8px; font-size:.68rem; font-weight:900; letter-spacing:.12em; text-transform:uppercase; color:var(--brand-muted); }
  .count { display:inline-block; padding:1px 6px; border-radius:99px; background:var(--brand-surface-3); }
  nav { display:grid; gap:4px; }
  .nav-link { display:block; padding:9px 11px; border-radius:10px; text-decoration:none; color:var(--brand-ink); border:1px solid transparent; }
  .nav-link:hover { background:var(--brand-surface-3); }
  .nav-link.active { background:var(--brand-surface-3); border-color:var(--brand-primary); }
  .nav-link strong { display:block; font-size:.84rem; }
  .nav-link small { display:block; margin-top:2px; font-size:.7rem; color:var(--brand-muted); line-height:1.35; }
  .section-link { padding:6px 11px 6px 20px; border-left:2px solid var(--brand-border); border-radius:0 10px 10px 0; }
  .section-link strong { font-size:.76rem; color:var(--brand-muted); }
  .empty { color:var(--brand-muted); font-size:.8rem; padding:4px 10px; }
  .side-foot { display:flex; justify-content:space-between; gap:8px; margin-top:28px; padding-top:14px; border-top:1px solid var(--brand-border); font-size:.7rem; color:var(--brand-muted); }
  .side-foot a { color:var(--brand-primary); }
  .content { min-width:0; }
  .doc-topbar { position:sticky; top:0; z-index:20; display:flex; align-items:center; gap:10px; padding:12px 28px; background:color-mix(in srgb, var(--brand-surface) 88%, transparent); backdrop-filter:blur(10px); border-bottom:1px solid var(--brand-border); }
  .crumb { font-weight:800; color:var(--brand-ink); }
  .spacer { flex:1; }
  .top-link { color:var(--brand-muted); text-decoration:none; font-size:.8rem; }
  .top-link.primary { color:var(--brand-primary); font-weight:800; }
  .page-body { padding:28px clamp(18px,4vw,56px) 80px; }
  .icon-btn { border:1px solid var(--brand-border); background:var(--brand-surface-3); color:var(--brand-ink); border-radius:9px; padding:7px 10px; cursor:pointer; }
  .mobile-only, .scrim { display:none; }
  @media (max-width: 900px) {
    .docs-shell { grid-template-columns:1fr; }
    .sidebar { position:fixed; z-index:60; left:0; top:0; width:min(320px,88vw); transform:translateX(-105%); transition:transform .2s ease; box-shadow:12px 0 40px rgba(0,0,0,.3); }
    .sidebar.open { transform:none; }
    .mobile-only { display:inline-flex; }
    .scrim { display:block; position:fixed; inset:0; z-index:50; border:0; background:rgba(0,0,0,.5); }
    .top-link { display:none; }
    .doc-topbar { padding:12px 16px; }
  }
</style>
