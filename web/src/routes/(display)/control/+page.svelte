<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { broadcastPA } from '../../../lib/api';
  import { setTheme, themeChoice } from '../../../lib/theme';

  type Template = { name: string; widgets: string[] };
  const runToday = [
    { path: '/operations', name: 'Check-in & weigh-in', icon: '📥', note: 'Arrivals, official weights and CSV' },
    { path: '/mat', name: 'Mat control', icon: '🎛️', note: 'Authoritative clock and result' },
    { path: '/display', name: 'Live scoreboard', icon: '🥊', note: 'Current fight, officials, moves' },
    { path: '/public', name: 'Public board', icon: '📺', note: 'Live, next up and standings' },
    { path: '/brackets', name: 'Brackets', icon: '🏆', note: 'Elimination and round robin' },
    { path: '/tables', name: 'Tables & rankings', icon: '📊', note: 'Standings, medals and kata' },
    { path: '/pools', name: 'Fight pools', icon: '🧩', note: 'Matchups and progression' },
    { path: '/ceremony', name: 'Medal ceremony', icon: '🏅', note: 'Animated podium callouts' },
    { path: '/event', name: 'Event config', icon: '📋', note: 'Rules, kit and venue details' },
    { path: '/overlay', name: 'Broadcast overlay', icon: '📡', note: 'Transparent scorebug for OBS' },
  ];
  const setupTools = [
    { path: '/setup', name: 'Create tournament', icon: '✨', note: 'Five-step setup wizard' },
    { path: '/', name: 'Operator dashboard', icon: '🎛️', note: 'Accounts, sign-up, pools, fights' },
    { path: '/simulate', name: 'Event simulator', icon: '🎬', note: 'Replay a full event + charts' },
    { path: '/builder', name: 'View builder', icon: '🧱', note: 'Drag, resize and save screens' },
    { path: '/ui', name: 'Component gallery', icon: '🧩', note: 'UI QA gallery' },
    { path: '/data', name: 'Data explorer', icon: '🔎', note: 'Search, filter and sort data' },
    { path: '/audit', name: 'Audit trail', icon: '📜', note: 'Who did what, when and how' },
  ];
  const widgetOptions = ['scoreboard', 'upcoming', 'standings', 'brackets', 'medals', 'kata', 'announcements'];

  const themes = ['dark', 'light', 'street'] as const;
  let templates: Template[] = [];
  let templateName = '';
  let selectedWidgets: string[] = ['scoreboard', 'upcoming', 'standings'];
  let paMessage = '';
  let paFighter = '';
  let paSeconds = 6;

  onMount(() => {
    try { templates = JSON.parse(localStorage.getItem('j3-templates') || '[]'); } catch { templates = []; }
  });

  function applyTheme(next: (typeof themes)[number]) { setTheme(next); }
  function saveTemplate() {
    if (!templateName.trim()) return;
    templates = [...templates.filter((t) => t.name !== templateName), { name: templateName, widgets: [...selectedWidgets] }];
    localStorage.setItem('j3-templates', JSON.stringify(templates));
    templateName = '';
  }
  function announce() {
    if (!paMessage.trim()) return;
    broadcastPA(paMessage, paFighter, paSeconds);
    if ('speechSynthesis' in window) {
      const utterance = new SpeechSynthesisUtterance(paMessage);
      utterance.rate = 1.05;
      window.speechSynthesis.speak(utterance);
    }
    paMessage = '';
    paFighter = '';
  }
  function openView(path: string) { window.open(path, '_blank', 'noopener,noreferrer'); }
  function showView(path: string) { void goto(path); }
</script>

<svelte:head><title>J3 · Display control</title></svelte:head>

<div class="display-shell">
  <header class="display-header" data-guide="control-header">
    <div><div class="event-mark"><span class="crest">J3</span>Display control</div><div class="meta">Open any view on any screen · everything stays live-synced</div></div>
    <div class="tabs">
      {#each [['dark','🌙 Dark gray'],['light','☀ Light'],['street','⚡ Street Fighter']] as [id, label]}
        <button class="tab" class:active={$themeChoice === id} on:click={() => applyTheme(id as (typeof themes)[number])}>{label}</button>
      {/each}
    </div>
  </header>

  <div class="section-title"><h2>Run today</h2><span class="muted small">Switch views here, or use ↗ to put a board on its own screen</span></div>
  <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(min(100%,260px), 1fr));">
    {#each runToday as view (view.path)}
      <div class="card slide-in view-card">
        <div style="font-size:2rem;">{view.icon}</div>
        <h3 style="margin-top:8px;">{view.name}</h3>
        <p class="muted small" style="margin-top:4px;">{view.note}</p>
        <div class="view-actions">
          <button class="btn primary" on:click={() => showView(view.path)}>Open view</button>
          <button class="btn" aria-label="Open {view.name} in a new tab" on:click={() => openView(view.path)}>↗ New tab</button>
        </div>
      </div>
    {/each}
  </div>

  <details class="setup-tools">
    <summary>Setup & QA tools <span class="muted small">({setupTools.length} views)</span></summary>
    <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(min(100%,260px), 1fr)); margin-top:14px;">
      {#each setupTools as view (view.path)}
        <div class="card view-card">
          <div style="font-size:1.6rem;">{view.icon}</div>
          <h3 style="margin-top:6px;">{view.name}</h3>
          <p class="muted small" style="margin-top:4px;">{view.note}</p>
          <div class="view-actions">
            <button class="btn primary" on:click={() => showView(view.path)}>Open view</button>
            <button class="btn" aria-label="Open {view.name} in a new tab" on:click={() => openView(view.path)}>↗ New tab</button>
          </div>
        </div>
      {/each}
    </div>
  </details>

  <div class="grid" style="grid-template-columns: 1fr 1fr; margin-top:30px;">
    <div class="card">
      <h2>Dashboard templates</h2>
      <p class="muted small">Choose widgets, name the layout, then assign it to a screen.</p>
      <div class="capabilities" style="margin-top:14px;">
        {#each widgetOptions as widget}
          <label class="cap" class:on={selectedWidgets.includes(widget)} style="cursor:pointer;">
            <input type="checkbox" value={widget} bind:group={selectedWidgets} /> {widget}
          </label>
        {/each}
      </div>
      <div style="display:flex; gap:8px; margin-top:16px;">
        <input class="tab" style="flex:1; color:var(--brand-ink); cursor:text;" placeholder="Template name…" bind:value={templateName} />
        <button class="btn primary" on:click={saveTemplate}>Save template</button>
      </div>
      <div style="margin-top:14px;">
        {#each templates as template (template.name)}
          <div class="stat small" style="margin-top:8px; display:flex; justify-content:space-between; gap:10px;">
            <span><strong>{template.name}</strong> · {template.widgets.join(', ')}</span>
            <button class="tab" on:click={() => openView('/public')}>Assign ↗</button>
          </div>
        {:else}<span class="muted small">No saved templates yet.</span>{/each}
      </div>
    </div>

    <div class="card">
      <h2>PA & announcer</h2>
      <p class="muted small">Callouts appear as toasts in this browser and other tabs on this device. Connect this screen to the venue speakers for venue-wide audio.</p>
      <div style="display:grid; gap:10px; margin-top:14px;">
        <input class="tab" style="color:var(--brand-ink); cursor:text;" placeholder="Message, e.g. 'Fighters to Mat 2'" bind:value={paMessage} />
        <input class="tab" style="color:var(--brand-ink); cursor:text;" placeholder="Fighter name (optional)" bind:value={paFighter} />
        <input class="tab" type="number" min="2" max="30" style="color:var(--brand-ink);" bind:value={paSeconds} />
      </div>
      <div style="display:flex; gap:8px; flex-wrap:wrap; margin-top:14px;">
        <button class="btn primary" on:click={announce}>📣 Announce</button>
        <button class="btn" on:click={() => { paMessage = 'Fighters to Mat 1 — prepare to compete'; announce(); }}>Call Mat 1</button>
        <button class="btn" on:click={() => { paMessage = ' Medal ceremony at the podium'; announce(); }}>Podium call</button>
      </div>
      <p class="muted small" style="margin-top:16px;">Open the public board in another tab on this device to preview the toast.</p>
    </div>
  </div>
</div>

<style>
  .view-card { display:flex; flex-direction:column; align-items:flex-start; gap:4px; }
  .view-card .view-actions { display:flex; gap:8px; flex-wrap:wrap; margin-top:auto; padding-top:14px; }
  .view-card .btn { font-size:.78rem; padding:7px 11px; }
  .setup-tools { margin-top:24px; border-top:1px solid var(--brand-border); padding-top:16px; }
  .setup-tools summary { cursor:pointer; color:var(--brand-muted); font-weight:800; font-size:.85rem; }
  .setup-tools summary:hover { color:var(--brand-ink); }
</style>
