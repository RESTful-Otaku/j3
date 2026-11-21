<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { TOURS, isTourDone, stepDone, type Role } from '../tour';

  let active = false;
  let index = 0;
  let role: Role = 'runner';
  let target: HTMLElement | null = null;
  let box = { top: 0, left: 0, width: 0, height: 0 };

  export function start(selected?: Role) {
    role = selected ?? role;
    if (isTourDone(role)) index = 0;
    active = true;
    index = 0;
    stepDone(role, -1);
    void navigate();
  }

  onMount(() => {
    const saved = localStorage.getItem('j3-role');
    if (saved) role = saved as Role;
    const listener = (event: Event) => {
      const detail = (event as CustomEvent<{ role?: Role }>).detail;
      start(detail?.role);
    };
    window.addEventListener('j3-start-tour', listener);
    if (!isTourDone(role)) setTimeout(() => start(), 600);
    return () => window.removeEventListener('j3-start-tour', listener);
  });

  async function navigate() {
    const step = TOURS[role].steps[index];
    if (step.href && window.location.pathname !== step.href) await goto(step.href);
    setTimeout(locate, 200);
  }
  function locate() {
    const step = TOURS[role].steps[index];
    target = document.querySelector(step.selector) as HTMLElement | null;
    if (target) {
      const rect = target.getBoundingClientRect();
      box = { top: rect.top - 8, left: rect.left - 8, width: rect.width + 16, height: rect.height + 16 };
    } else {
      box = { top: 80, left: window.innerWidth / 2 - 160, width: 320, height: 120 };
    }
  }
  async function next() {
    stepDone(role, index);
    if (index + 1 >= TOURS[role].steps.length) { finish(); return; }
    index += 1;
    await navigate();
  }
  function finish() {
    stepDone(role, TOURS[role].steps.length - 1);
    active = false;
  }
</script>

{#if active}
  <div class="tour-root" role="dialog" aria-modal="true" aria-label="Guided walkthrough">
    <div class="spotlight" style="top:{box.top}px; left:{box.left}px; width:{box.width}px; height:{box.height}px;"></div>
    <div class="tour-card" style="top:{Math.min(box.top + box.height + 14, window.innerHeight - 230)}px; left:{Math.max(12, Math.min(box.left, window.innerWidth - 350))}px;">
      <div class="eyebrow">Guided walkthrough · {index + 1}/{TOURS[role].steps.length}</div>
      <h2>{TOURS[role].steps[index].title}</h2>
      <p>{TOURS[role].steps[index].body}</p>
      <div class="progress"><span style="width:{((index + 1) / TOURS[role].steps.length) * 100}%"></span></div>
      <div class="actions">
        <button class="btn" on:click={finish}>Skip tour</button>
        <button class="btn primary" on:click={next}>{index + 1 === TOURS[role].steps.length ? 'Finish' : 'Next →'}</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .tour-root { position:fixed; inset:0; z-index:1200; }
  .spotlight { position:fixed; border:3px solid var(--brand-accent); border-radius:14px; box-shadow:0 0 0 9999px rgba(2,6,23,.72); pointer-events:none; transition:all .25s ease; }
  .tour-card { position:fixed; width:min(340px, 92vw); padding:18px; border-radius:14px; background:var(--brand-surface-2); border:1px solid var(--brand-accent); box-shadow:0 20px 60px rgba(0,0,0,.5); }
  h2 { margin:8px 0; font-size:1.15rem; }
  p { color:var(--brand-muted); font-size:.88rem; }
  .progress { height:4px; border-radius:999px; background:var(--brand-surface-3); margin:14px 0; overflow:hidden; }
  .progress span { display:block; height:100%; background:var(--brand-accent); }
  .actions { display:flex; justify-content:space-between; gap:8px; }
  @media (max-width: 700px) { .tour-card { left:12px !important; top:auto !important; bottom:12px; } }
</style>
