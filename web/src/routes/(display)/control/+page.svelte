<script lang="ts">
  import { onMount } from 'svelte';
  import { broadcastPA } from '../../../lib/api';

  type Template = { name: string; widgets: string[] };
  const views = [
    { path: '/setup', name: 'Create tournament', icon: '✨', note: 'Five-step setup wizard' },
    { path: '/', name: 'Operator QA', icon: '🎛️', note: 'Accounts, sign-up, pools, fights' },
    { path: '/simulate', name: 'Tournament simulator', icon: '🎬', note: 'Replay a full event + charts' },
    { path: '/data', name: 'Data explorer', icon: '🔎', note: 'Search, filter, sort any dataset' },
    { path: '/audit', name: 'Audit trail', icon: '📜', note: 'Who did what, when and how' },
    { path: '/mat', name: 'Mat scorekeeper console', icon: '🎛️', note: 'Authoritative control per mat' },
    { path: '/display', name: 'Full-screen scoreboard', icon: '🥊', note: 'Current fight, officials, moves' },
    { path: '/ceremony', name: 'Medal ceremony', icon: '🏆', note: 'Animated podium + PA call' },
    { path: '/public', name: 'Public dashboard', icon: '📺', note: 'Live, next up, leaderboard' },
    { path: '/details', name: 'Event details', icon: '📋', note: 'Rules, protection, schedule, corners' },
    { path: '/brackets', name: 'Brackets', icon: '🏆', note: 'Elimination + round robin' },
    { path: '/tables', name: 'Tables & rankings', icon: '📊', note: 'Overall, medals, kata panel' },
    { path: '/pools', name: 'Fight pools', icon: '🧩', note: 'Matchups and progression' },
  ];
  const widgetOptions = ['scoreboard', 'upcoming', 'standings', 'brackets', 'medals', 'kata', 'announcements'];

  let theme: 'dark' | 'light' | 'street' = 'dark';
  let templates: Template[] = [];
  let templateName = '';
  let selectedWidgets: string[] = ['scoreboard', 'upcoming', 'standings'];
  let paMessage = '';
  let paFighter = '';
  let paSeconds = 6;

  onMount(() => {
    theme = (localStorage.getItem('j3-theme') as typeof theme) || 'dark';
    document.documentElement.dataset.theme = theme;
    try { templates = JSON.parse(localStorage.getItem('j3-templates') || '[]'); } catch { templates = []; }
  });

  function applyTheme(next: typeof theme) {
    theme = next;
    localStorage.setItem('j3-theme', next);
    document.documentElement.dataset.theme = next;
  }
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
</script>

<svelte:head><title>J3 · Display control</title></svelte:head>

<div class="display-shell">
  <header class="display-header">
    <div><div class="event-mark"><span class="crest">J3</span>Display control</div><div class="meta">Open any view on any screen · everything stays live-synced</div></div>
    <div class="tabs">
      {#each [['dark','🌙 Dark'],['light','☀ Light'],['street','⚡ Street Fighter']] as [id, label]}
        <button class="tab" class:active={theme === id} on:click={() => applyTheme(id as typeof theme)}>{label}</button>
      {/each}
    </div>
  </header>

  <div class="section-title"><h2>Open a screen</h2><span class="muted small">Each opens in its own tab — drag it to a TV or projector</span></div>
  <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(min(100%,260px), 1fr));">
    {#each views as view (view.path)}
      <button class="card slide-in" style="text-align:left; cursor:pointer; color:var(--brand-ink);" on:click={() => openView(view.path)}>
        <div style="font-size:2rem;">{view.icon}</div>
        <h3 style="margin-top:8px;">{view.name}</h3>
        <p class="muted small" style="margin-top:4px;">{view.note}</p>
        <div class="eyebrow" style="margin-top:14px;">Open in new tab ↗</div>
      </button>
    {/each}
  </div>

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
      <p class="muted small">Shoutouts and callouts appear as timed toasts on every open screen.</p>
      <div style="display:grid; gap:10px; margin-top:14px;">
        <input class="tab" style="color:var(--brand-ink); cursor:text;" placeholder="Message, e.g. 'Fighters to Mat 2'" bind:value={paMessage} />
        <input class="tab" style="color:var(--brand-ink); cursor:text;" placeholder="Fighter ID (optional)" bind:value={paFighter} />
        <input class="tab" type="number" min="2" max="30" style="color:var(--brand-ink);" bind:value={paSeconds} />
      </div>
      <div style="display:flex; gap:8px; flex-wrap:wrap; margin-top:14px;">
        <button class="btn primary" on:click={announce}>📣 Announce</button>
        <button class="btn" on:click={() => { paMessage = 'Fighters to Mat 1 — prepare to compete'; announce(); }}>Call Mat 1</button>
        <button class="btn" on:click={() => { paMessage = ' Medal ceremony at the podium'; announce(); }}>Podium call</button>
      </div>
      <p class="muted small" style="margin-top:16px;">Open the public dashboard in another tab to see the toast appear live.</p>
    </div>
  </div>
</div>
