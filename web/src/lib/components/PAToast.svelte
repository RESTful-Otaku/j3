<script lang="ts">
  import { onMount } from 'svelte';

  type Toast = { id: number; message: string; fighterId: string };
  let toasts: Toast[] = [];

  function add(detail: { message: string; fighterId?: string; seconds?: number }) {
    if (!detail?.message) return;
    const id = Date.now() + Math.random();
    toasts = [...toasts, { id, message: detail.message, fighterId: detail.fighterId ?? '' }];
    setTimeout(() => (toasts = toasts.filter((t) => t.id !== id)), (detail.seconds ?? 6) * 1000);
  }

  onMount(() => {
    const fromOtherTab = (event: StorageEvent) => {
      if (event.key === 'j3-pa' && event.newValue) {
        try { add(JSON.parse(event.newValue)); } catch { /* ignore malformed payloads */ }
      }
    };
    const sameTab = (event: Event) => add((event as CustomEvent).detail);
    window.addEventListener('storage', fromOtherTab);
    window.addEventListener('j3-pa', sameTab);
    return () => {
      window.removeEventListener('storage', fromOtherTab);
      window.removeEventListener('j3-pa', sameTab);
    };
  });
</script>

<div class="toast-stack" aria-live="polite">
  {#each toasts as toast (toast.id)}
    <div class="toast flash">
      <span class="spark">📣</span>
      <div>
        <strong>{toast.message}</strong>
        {#if toast.fighterId}<small>{toast.fighterId} — report to your mat</small>{/if}
      </div>
    </div>
  {/each}
</div>

<style>
  .toast-stack { position: fixed; top: 18px; right: 18px; z-index: 900; display: grid; gap: 10px; width: min(420px, 90vw); }
  .toast {
    display: flex; gap: 12px; align-items: center; padding: 16px 18px; border-radius: 14px;
    background: linear-gradient(135deg, var(--brand-primary), var(--brand-primary-soft));
    color: #fff; box-shadow: 0 18px 50px rgba(0,0,0,.45); animation: slide .3s ease-out;
  }
  .toast strong { display: block; font-size: 1rem; }
  .toast small { display: block; opacity: .85; margin-top: 2px; }
  .spark { font-size: 1.6rem; }
  .flash { animation: slide .3s ease-out, glow 1s ease-in-out infinite alternate; }
  @keyframes slide { from { transform: translateX(40px); opacity: 0; } to { transform: none; opacity: 1; } }
  @keyframes glow { from { box-shadow: 0 0 0 0 rgba(251,191,36,.5); } to { box-shadow: 0 0 0 10px rgba(251,191,36,0); } }
</style>
