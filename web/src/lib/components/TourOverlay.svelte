<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { afterNavigate, goto } from '$app/navigation';
  import { session, uiRole } from '../auth';
  import { GUIDE_ACTION_EVENT } from '../guide';
  import {
    guideForRole, guideStats, isGuideDismissed, markStepDone,
    setGuideSkipped, stepIsComplete, type GuideStep, type Role,
  } from '../tour';

  let active = false;
  let index = 0;
  let role: Role = 'runner';
  let satisfied = false;
  let target: HTMLElement | null = null;
  let box = { top: 80, left: 80, width: 300, height: 140 };
  let vw = 1200;
  let vh = 800;
  let verifyTimer: ReturnType<typeof setInterval> | undefined;

  $: guide = guideForRole(role);
  $: step = guide.steps[Math.min(index, guide.steps.length - 1)];
  $: stats = guideStats(role);
  $: stepAlreadyDone = Boolean(step && stats.progress.completed.includes(step.id));
  $: displayPercent = stats.total
    ? Math.round(((stats.completed + (satisfied && !stepAlreadyDone ? 1 : 0)) / stats.total) * 100)
    : 0;
  $: cardLeft = Math.max(12, Math.min(box.left, vw - 366));
  $: preferredTop = box.top + box.height + 14;
  $: cardStyle = vw <= 820
    ? ''
    : preferredTop + 330 <= vh
      ? `top:${Math.max(12, preferredTop)}px; left:${cardLeft}px;`
      : `bottom:calc(12px + env(safe-area-inset-bottom, 0px)); left:${cardLeft}px;`;

  // Manual navigation or a guide jump can change the target without a resize.
  afterNavigate(() => {
    if (active) {
      locate();
      void verify();
    }
  });

  function syncViewport() {
    vw = window.innerWidth;
    vh = window.innerHeight;
  }

  async function start(selected?: Role, stepIndex?: number) {
    // Auto-login can finish after this component mounts; always re-read the
    // verified session so an admin follows the runner guide with one progress key.
    const activeSession = session();
    role = selected ?? (activeSession ? uiRole(activeSession.role) : role);
    const current = guideForRole(role);
    const next = typeof stepIndex === 'number' && Number.isFinite(stepIndex)
      ? Math.max(0, Math.min(stepIndex, current.steps.length - 1))
      : guideStats(role).nextIndex;
    index = next;
    active = true;
    satisfied = stepIsComplete(role, current.steps[index]);
    syncViewport();
    await navigateToStep();
    void verify();
  }

  async function prepareStep(current: GuideStep) {
    // Several steps live inside collapsed wizard tabs or behind a selected
    // contest. Open the right tab before trying to spotlight the control.
    if (typeof current.wizardStep !== 'number') return;
    const tab = document.querySelector(
      `[data-guide="setup-stepper"] [data-step="${current.wizardStep}"]`,
    ) as HTMLButtonElement | null;
    if (tab && !tab.classList.contains('active')) {
      tab.click();
      await tick();
    }
  }

  async function navigateToStep() {
    const current = guideForRole(role).steps[index];
    if (!current) return;
    if (window.location.pathname !== current.href) await goto(current.href);
    await tick();
    await prepareStep(current);
    setTimeout(() => { locate(true); void verify(); }, 220);
  }

  function computeBox() {
    if (!target) {
      box = { top: 80, left: Math.max(12, vw / 2 - 150), width: 300, height: 120 };
      return;
    }
    const rect = target.getBoundingClientRect();
    box = {
      top: Math.max(8, rect.top - 8),
      left: Math.max(8, rect.left - 8),
      width: Math.min(vw - 16, rect.width + 16),
      height: Math.min(vh - 16, rect.height + 16),
    };
  }

  function locate(scroll = false) {
    const current = guideForRole(role).steps[index];
    if (!current) return;
    const selector = vw <= 820 && current.mobileTarget ? current.mobileTarget : current.target;
    target = document.querySelector(selector) as HTMLElement | null;
    if (target && scroll) target.scrollIntoView({ block: 'center', behavior: 'smooth' });
    computeBox();
    if (target && scroll) setTimeout(computeBox, 180);
  }

  async function verify() {
    if (!active) return;
    const current = guideForRole(role).steps[index];
    if (!current) return;
    const passed = stepIsComplete(role, current);
    if (passed !== satisfied) satisfied = passed;
  }

  async function next() {
    if (!active || !satisfied || !step) return;
    markStepDone(role, step.id);
    if (index + 1 >= guide.steps.length) {
      active = false;
      return;
    }
    index += 1;
    satisfied = stepIsComplete(role, guide.steps[index]);
    await navigateToStep();
  }

  function back() {
    if (index === 0) return;
    index -= 1;
    satisfied = stepIsComplete(role, guide.steps[index]);
    void navigateToStep();
  }

  function skip() {
    setGuideSkipped(role);
    active = false;
  }

  function onAction(event: Event) {
    const detail = (event as CustomEvent<{ key?: string }>).detail;
    if (detail?.key) void verify();
  }

  function onFieldChange() { void verify(); }

  function onResize() {
    syncViewport();
    computeBox();
  }

  function onKeydown(event: KeyboardEvent) {
    if (active && event.key === 'Escape') skip();
  }

  onMount(() => {
    syncViewport();
    const activeSession = session();
    if (activeSession) {
      const mapped = uiRole(activeSession.role);
      if (mapped) role = mapped;
    }
    const startListener = (event: Event) => {
      const detail = (event as CustomEvent<{ role?: Role; stepIndex?: number }>).detail;
      start(detail?.role, detail?.stepIndex).catch(() => undefined);
    };
    const sessionListener = () => {
      const activeRole = session() ? uiRole(session()!.role) : null;
      if (activeRole && activeRole !== role && !active) role = activeRole;
    };
    window.addEventListener('j3-start-tour', startListener);
    window.addEventListener('j3-session-changed', sessionListener);
    window.addEventListener(GUIDE_ACTION_EVENT, onAction);
    window.addEventListener('input', onFieldChange);
    window.addEventListener('change', onFieldChange);
    window.addEventListener('resize', onResize);
    window.addEventListener('scroll', computeBox, true);
    window.addEventListener('keydown', onKeydown);

    const docsRoute = window.location.pathname.startsWith('/docs') || window.location.pathname === '/apidocs';
    if (!docsRoute && !isGuideDismissed(role)) {
      setTimeout(() => { void start(); }, 600);
    }

    verifyTimer = setInterval(() => void verify(), 250);
    return () => {
      window.removeEventListener('j3-start-tour', startListener);
      window.removeEventListener('j3-session-changed', sessionListener);
      window.removeEventListener(GUIDE_ACTION_EVENT, onAction);
      window.removeEventListener('input', onFieldChange);
      window.removeEventListener('change', onFieldChange);
      window.removeEventListener('resize', onResize);
      window.removeEventListener('scroll', computeBox, true);
      window.removeEventListener('keydown', onKeydown);
      if (verifyTimer) clearInterval(verifyTimer);
    };
  });
</script>

<svelte:window on:resize={syncViewport} />

{#if active && step}
  <div class="tour-root" role="dialog" aria-modal="true" aria-label="Guided walkthrough" data-guide="tour-dialog">
    <div class="spotlight" style="top:{box.top}px; left:{box.left}px; width:{box.width}px; height:{box.height}px;"></div>
    <div class="tour-card" style={cardStyle}>
      <div class="eyebrow">Hands-on guide · {guide.shortLabel} · {index + 1}/{guide.steps.length}</div>
      <h2>{step.title}</h2>
      <p>{step.body}</p>
      <p class="hint"><strong>Do this:</strong> {step.hint}</p>
      <div class="progress" aria-label="Guide progress"><span style="width:{displayPercent}%"></span></div>
      <div class="status" class:complete={satisfied} aria-live="polite">
        {satisfied ? '✅ Step complete — you can move on' : '🔒 Try the step above to unlock Next'}
      </div>
      <div class="actions">
        <button class="btn ghost" data-guide="tour-skip" on:click={skip}>Skip guide</button>
        <div class="right-actions">
          {#if index > 0}<button class="btn" on:click={back}>← Back</button>{/if}
          <button class="btn primary" data-guide="tour-next" disabled={!satisfied} on:click={() => void next()}>
            {index + 1 === guide.steps.length ? 'Finish ✓' : 'Next →'}
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
  .tour-root { position:fixed; inset:0; z-index:950; pointer-events:none; }
  .spotlight { position:fixed; border:3px solid var(--brand-accent, #38bdf8); border-radius:14px; box-shadow:0 0 0 9999px rgba(2,6,23,.68), 0 0 34px rgba(56,189,248,.6); pointer-events:none; transition:top .25s ease,left .25s ease,width .25s ease,height .25s ease; animation:pulse 1.8s ease-in-out infinite; }
  .tour-card { position:fixed; width:min(352px, calc(100vw - 24px - var(--safe-left) - var(--safe-right))); max-height:calc(100vh - 24px - var(--safe-top) - var(--safe-bottom)); overflow:auto; padding:18px; border-radius:16px; background:var(--brand-surface-2,#0f172a); border:1px solid var(--brand-accent,#38bdf8); box-shadow:0 24px 80px rgba(0,0,0,.6); pointer-events:auto; }
  h2 { margin:8px 0; font-size:1.15rem; }
  p { color:var(--brand-muted,#94a3b8); font-size:.88rem; line-height:1.5; margin:6px 0; }
  .hint { padding:8px 10px; border-radius:10px; background:var(--brand-surface-3,#1e293b); color:var(--brand-ink,#e2e8f0); font-size:.82rem; }
  .progress { height:5px; border-radius:999px; background:var(--brand-surface-3,#1e293b); margin:14px 0 8px; overflow:hidden; }
  .progress span { display:block; height:100%; background:linear-gradient(90deg,var(--brand-accent,#38bdf8),var(--brand-orange,#fb923c)); transition:width .3s ease; }
  .status { font-size:.78rem; font-weight:700; color:#fbbf24; margin-bottom:12px; }
  .status.complete { color:#4ade80; }
  .actions { display:flex; justify-content:space-between; gap:8px; align-items:center; }
  .right-actions { display:flex; gap:8px; }
  .btn { padding:8px 12px; font-size:.82rem; }
  .btn.ghost { background:transparent; border-color:var(--brand-border,#334155); }
  .btn:disabled { opacity:.55; cursor:not-allowed; }
  @media (max-width: 820px) {
    .tour-card { left:calc(12px + var(--safe-left)) !important; right:calc(12px + var(--safe-right)) !important; top:auto !important; bottom:calc(84px + var(--safe-bottom)) !important; width:auto; max-height:58vh; overflow:auto; }
    .actions { flex-wrap:wrap; }
  }
  @keyframes pulse { 0%,100% { filter:brightness(1); } 50% { filter:brightness(1.35); } }
</style>
