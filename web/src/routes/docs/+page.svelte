<script lang="ts">
  import { onMount } from 'svelte';
  import { docs, docHref } from '../../lib/docs';
  import type { DocEntry } from '../../lib/docs';

  let entries: DocEntry[] = [];
  let status = 'Checking local venue server…';

  onMount(() => {
    void docs.manifest().then((data) => (entries = data.docs)).catch(() => {});
    void fetch('/healthz')
      .then((response) => (status = response.ok ? 'Venue server online · demo data ready' : 'Venue server offline'))
      .catch(() => (status = 'Venue server offline'));
  });

  const tasks = [
    { icon: '✨', title: 'Create an event', href: '/docs/guide#create-your-event-in-five-steps', text: 'Five guided steps with sensible defaults.' },
    { icon: '🧩', title: 'Build pools & brackets', href: '/docs/guide#let-j3-build-the-draw', text: 'Balanced pools, BYEs and automatic promotion.' },
    { icon: '🥊', title: 'Score a contest', href: '/docs/guide#score-a-contest', text: 'Timers, techniques, penalties and instant wins.' },
    { icon: '📺', title: 'Set up displays', href: '/docs/guide#screens-and-displays', text: 'One URL per screen, all live-synced.' },
    { icon: '🎓', title: 'Follow hands-on guides', href: '/guides', text: 'Role-specific walkthroughs that check every step.' },
  ];
</script>

<svelte:head><title>J3 · Documentation</title></svelte:head>

<section class="hero">
  <div class="eyebrow">J3 · J Cubed</div>
  <h1>Everything you need to run J3.</h1>
  <p>
    Plain-language guides for organisers, deep references for developers and a live API explorer —
    all in one place, and all usable offline at the venue.
  </p>
  <div class="hero-actions">
    <a class="btn primary" href="/docs/guide">Start with the organiser guide →</a>
    <a class="btn" href="/docs/api">Explore the API</a>
    <a class="btn" href="/docs/development">Developer guide</a>
  </div>
  <div class="status"><span class:online={status.includes('online')}></span>{status}</div>
</section>

<section>
  <h2>Popular tasks</h2>
  <div class="task-grid">
    {#each tasks as task}
      <a class="task" href={task.href}>
        <span class="task-icon">{task.icon}</span>
        <strong>{task.title}</strong>
        <small>{task.text}</small>
      </a>
    {/each}
  </div>
</section>

<section>
  <h2>All documentation</h2>
  <div class="card-grid">
    {#each entries as entry}
      <a class="doc-card" href={docHref(entry)}>
        <div class="card-icon">{entry.icon}</div>
        <div class="card-group">{entry.group}</div>
        <h3>{entry.title}</h3>
        <p>{entry.description}</p>
        <span class="card-link">Open →</span>
      </a>
    {:else}
      <p class="loading">Loading documentation…</p>
    {/each}
  </div>
</section>

<section class="quickstart">
  <div>
    <h2>Try it in two minutes</h2>
    <p>The local demo is seeded with clubs, fighters, officials, scored fights, kata results and medals.</p>
    <pre>bun run dev</pre>
    <p class="small">Choose <strong>▶ Run full local demo</strong>, then open
      <a href="http://localhost:5173">localhost:5173</a>. The TUI menu also opens these docs.</p>
  </div>
  <div class="quick-links">
    <a href="/docs/faq">❓ Frequently asked questions</a>
    <a href="/docs/ci">🚦 CI & releases</a>
    <a href="/openapi.yaml" target="_blank" rel="noreferrer">📄 Raw OpenAPI 3.1 contract ↗</a>
  </div>
</section>

<style>
  .hero { padding:34px 0 26px; border-bottom:1px solid var(--brand-border); margin-bottom:28px; }
  .eyebrow { color:var(--brand-accent); font-size:.72rem; letter-spacing:.16em; font-weight:900; text-transform:uppercase; }
  h1 { font-size:clamp(2.1rem,5vw,3.6rem); line-height:1.02; margin:12px 0 14px; }
  .hero p { max-width:70ch; color:var(--brand-muted); font-size:1.02rem; line-height:1.65; }
  .hero-actions { display:flex; flex-wrap:wrap; gap:10px; margin:22px 0; }
  .status { display:inline-flex; align-items:center; gap:8px; padding:7px 12px; border-radius:99px; background:var(--brand-surface-3); color:var(--brand-muted); font-size:.76rem; }
  .status span { width:8px; height:8px; border-radius:50%; background:var(--loss); }
  .status span.online { background:var(--win); }
  section { margin-bottom:36px; }
  h2 { font-size:1.35rem; margin-bottom:14px; }
  .task-grid, .card-grid { display:grid; grid-template-columns:repeat(auto-fit,minmax(210px,1fr)); gap:12px; }
  .task { display:grid; gap:6px; padding:16px; border-radius:14px; background:var(--brand-surface-3); border:1px solid var(--brand-border); color:var(--brand-ink); text-decoration:none; }
  .task:hover, .doc-card:hover { border-color:var(--brand-primary); transform:translateY(-2px); }
  .task-icon, .card-icon { font-size:1.5rem; }
  .task small, .doc-card p { color:var(--brand-muted); font-size:.8rem; line-height:1.5; }
  .doc-card { display:block; padding:18px; border-radius:16px; border:1px solid var(--brand-border); background:var(--brand-surface-2); color:var(--brand-ink); text-decoration:none; transition:transform .15s ease,border-color .15s ease; }
  .doc-card h3 { margin:8px 0 6px; font-size:1.05rem; }
  .card-group { margin-top:10px; font-size:.65rem; font-weight:900; letter-spacing:.12em; text-transform:uppercase; color:var(--brand-accent); }
  .card-link { display:inline-block; margin-top:12px; color:var(--brand-primary); font-weight:800; font-size:.78rem; }
  .quickstart { display:grid; grid-template-columns:minmax(0,2fr) minmax(220px,1fr); gap:24px; padding:22px; border-radius:18px; background:var(--brand-surface-2); border:1px solid var(--brand-border); }
  .quickstart p { color:var(--brand-muted); line-height:1.6; }
  .quickstart pre { display:inline-block; padding:9px 14px; border-radius:9px; background:var(--brand-surface-3); color:var(--brand-ink); }
  .small { font-size:.82rem; }
  .quick-links { display:grid; align-content:start; gap:8px; }
  .quick-links a { padding:10px 12px; border-radius:10px; background:var(--brand-surface-3); color:var(--brand-ink); text-decoration:none; font-size:.82rem; }
  .loading { color:var(--brand-muted); }
  @media (max-width:760px) { .quickstart { grid-template-columns:1fr; } }
</style>
