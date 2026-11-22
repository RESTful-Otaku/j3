<script lang="ts">
  import { onMount } from 'svelte';
  import { session, uiRole } from '../../lib/auth';
  import { clearGuideActions } from '../../lib/guide';
  import { currentRole } from '../../lib/nav';
  import { notifySuccess } from '../../lib/toast';
  import { guideForRole, guideStats, resetGuide, type Guide, type Role } from '../../lib/tour';

  const entries: { role: Role; guide: Guide }[] = [
    { role: 'runner', guide: guideForRole('runner') },
    { role: 'official', guide: guideForRole('official') },
    { role: 'club_leader', guide: guideForRole('club_leader') },
    { role: 'fighter', guide: guideForRole('fighter') },
  ];

  let role: Role = 'runner';
  let selectedRole: Role = 'runner';
  let version = 0;

  // Reading `version` makes Svelte recompute after a reset.
  $: cards = entries.map((entry) => {
    void version;
    return { ...entry, stats: guideStats(entry.role) };
  });
  $: selected = cards.find((card) => card.role === selectedRole) ?? cards[0];
  $: currentStats = guideStats(role);

  onMount(() => {
    const active = session();
    role = active && uiRole(active.role) ? uiRole(active.role)! : currentRole();
    selectedRole = role === 'spectator' ? 'fighter' : role;
    const listener = () => {
      const next = session() && uiRole(session()!.role) ? uiRole(session()!.role)! : currentRole();
      role = next;
      selectedRole = next === 'spectator' ? 'fighter' : next;
      version += 1;
    };
    window.addEventListener('j3-session-changed', listener);
    return () => window.removeEventListener('j3-session-changed', listener);
  });

  function start(targetRole: Role, stepIndex?: number) {
    window.dispatchEvent(new CustomEvent('j3-start-tour', {
      detail: { role: targetRole, stepIndex },
    }));
  }

  function reset(targetRole: Role, label: string) {
    resetGuide(targetRole);
    version += 1;
    notifySuccess(`${label} progress reset`);
  }

  function resetAll() {
    resetGuide();
    clearGuideActions();
    version += 1;
    notifySuccess('All guide progress reset');
  }
</script>

<svelte:head><title>J3 · Guides & progress</title></svelte:head>

<div class="shell">
  <header class="topbar">
    <div class="brand">
      <div class="brand-mark">J3</div>
      <div><div class="brand-name">Guides & progress</div><div class="brand-sub">Learn by doing · skip or resume whenever you like</div></div>
    </div>
    <a class="btn" href="/settings">⚙️ Settings</a>
  </header>

  <section class="card hero" data-guide="guides-hero">
    <div class="eyebrow">Your current role · {guideForRole(role).label}</div>
    <h1>Learn the whole tournament by running one.</h1>
    <p class="muted">
      These guides do not just point at buttons. Each step waits until you have actually performed the
      action — filled a field, confirmed a score, published a result — then unlocks the next step.
    </p>
    <div class="hero-row">
      <div class="ring" style="--pct:{currentStats.percent}%"><span>{currentStats.percent}%</span></div>
      <div class="hero-copy">
        <strong>{currentStats.completed} of {currentStats.total} steps complete</strong>
        <div class="progress"><span style="width:{currentStats.percent}%"></span></div>
        <span class="muted small">{currentStats.progress.skipped ? 'Guide skipped — you can restart anytime.' : currentStats.done ? 'Guide complete. Well done!' : 'Pick up where you left off below.'}</span>
      </div>
      <div class="hero-actions">
        <button class="btn primary" data-guide="guides-start" on:click={() => start(role)}>
          {currentStats.done ? '↻ Repeat guide' : '▶ Resume guide'}
        </button>
        <button class="btn" on:click={() => reset(role, guideForRole(role).label)}>Reset progress</button>
      </div>
    </div>
  </section>

  <h2 class="section-heading">Choose a guide</h2>
  <div class="guide-grid">
    {#each cards as card (card.role)}
      <button class="guide-card" class:active={selectedRole === card.role} on:click={() => (selectedRole = card.role)}>
        <div class="guide-icon">{card.guide.icon}</div>
        <strong>{card.guide.label}</strong>
        <span class="muted small">{card.guide.description}</span>
        <div class="progress"><span style="width:{card.stats.percent}%"></span></div>
        <span class="muted small">{card.stats.completed}/{card.stats.total} steps</span>
      </button>
    {/each}
  </div>

  <section class="card checklist" data-guide="guides-checklist">
    <div class="section-title">
      <div>
        <div class="eyebrow">{selected.guide.label} · checklist</div>
        <h2>{selected.guide.steps.length} hands-on steps</h2>
      </div>
      <button class="btn" on:click={() => resetAll()}>Reset all guides</button>
    </div>
    <p class="muted small">Jump straight to any step to refresh your memory. Jumping does not erase steps you have already completed.</p>
    <ol>
      {#each selected.guide.steps as step, i (step.id)}
        <li class:done={selected.stats.progress.completed.includes(step.id)}>
          <span class="step-mark">{selected.stats.progress.completed.includes(step.id) ? '✓' : i + 1}</span>
          <span class="step-copy"><strong>{step.title}</strong><span class="muted small">{step.hint}</span></span>
          <button class="btn small" on:click={() => start(selected.role, i)}>Jump here →</button>
        </li>
      {/each}
    </ol>
  </section>
</div>

<style>
  .hero { margin-top:18px; }
  h1 { font-family:var(--font-display); font-size:clamp(1.6rem,4vw,2.6rem); margin:6px 0; }
  .hero-row { display:flex; align-items:center; gap:22px; margin-top:20px; flex-wrap:wrap; }
  .ring { width:104px; height:104px; border-radius:50%; display:grid; place-items:center; flex:0 0 auto; background:conic-gradient(var(--brand-accent,#38bdf8) var(--pct), var(--brand-surface-3,#1e293b) 0); }
  .ring span { width:78px; height:78px; border-radius:50%; display:grid; place-items:center; background:var(--brand-surface-2,#0f172a); font-weight:800; font-size:1.3rem; }
  .hero-copy { flex:1 1 260px; }
  .hero-actions { display:flex; gap:8px; flex-wrap:wrap; }
  .progress { height:6px; background:var(--brand-surface-3,#1e293b); border-radius:999px; overflow:hidden; margin:8px 0; }
  .progress span { display:block; height:100%; background:linear-gradient(90deg,var(--brand-accent,#38bdf8),var(--brand-orange,#fb923c)); }
  .section-heading { margin:26px 0 12px; font-family:var(--font-display); }
  .guide-grid { display:grid; grid-template-columns:repeat(auto-fit,minmax(220px,1fr)); gap:12px; }
  .guide-card { display:grid; gap:7px; text-align:left; padding:16px; border-radius:16px; border:1px solid var(--brand-border,#334155); background:var(--brand-surface-2,#0f172a); color:var(--brand-ink,#e2e8f0); cursor:pointer; transition:transform .18s ease,border-color .18s ease; }
  .guide-card:hover { transform:translateY(-3px); border-color:var(--brand-accent,#38bdf8); }
  .guide-card.active { border-color:var(--brand-accent,#38bdf8); box-shadow:0 0 0 2px rgba(56,189,248,.22); }
  .guide-icon { font-size:1.8rem; }
  .checklist { margin-top:22px; }
  ol { list-style:none; margin:16px 0 0; padding:0; display:grid; gap:8px; counter-reset:step; }
  li { display:flex; align-items:center; gap:12px; padding:11px 12px; border:1px solid var(--brand-border,#334155); border-radius:12px; background:var(--brand-surface-3,#1e293b); }
  li.done { border-color:rgba(74,222,128,.5); }
  .step-mark { width:30px; height:30px; border-radius:50%; display:grid; place-items:center; flex:0 0 auto; background:var(--brand-surface-2,#0f172a); font-weight:800; }
  li.done .step-mark { background:#166534; color:#dcfce7; }
  .step-copy { display:grid; gap:2px; flex:1; }
  .btn.small { padding:6px 10px; font-size:.78rem; }
  @media (max-width: 700px) { .hero-row { gap:14px; } .hero-actions { width:100%; } li { flex-wrap:wrap; } }
</style>
