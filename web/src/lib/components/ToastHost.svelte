<script lang="ts">
  import { onMount } from 'svelte';
  import { fly } from 'svelte/transition';
  import { runUndo, type ToastDetail } from '../toast';

  type ActiveToast = ToastDetail;
  let toasts: ActiveToast[] = [];

  const icons: Record<string, string> = {
    success: '✅', error: '⛔', warning: '⚠️', info: '💡',
  };

  onMount(() => {
    const handler = (event: Event) => {
      const detail = (event as CustomEvent<ToastDetail>).detail;
      toasts = [...toasts, detail];
      setTimeout(() => dismiss(detail.id), Math.max(detail.duration ?? 4000, 1500)); // cooldown min 1.5s; always dismissible
    };
    window.addEventListener('j3-toast', handler);
    return () => window.removeEventListener('j3-toast', handler);
  });

  function dismiss(id: string) {
    toasts = toasts.filter((toast) => toast.id !== id);
  }
</script>

<div class="toast-stack" aria-live="polite">
  {#each toasts as toast (toast.id)}
    <div class="toast {toast.type}" role="status" in:fly={{ x: 60, duration: 280 }} out:fly={{ x: 60, duration: 200 }}>
      <span class="icon">{icons[toast.type] ?? '💡'}</span>
      <span class="message">{toast.message}</span>
      {#if toast.undoable}
        <button class="undo" on:click={() => { runUndo(toast.id); dismiss(toast.id); }}>Undo</button>
      {/if}
      <button class="close" aria-label="Dismiss notification" on:click={() => dismiss(toast.id)}>×</button>
      <span class="progress" style="animation-duration: {toast.duration}ms"></span>
    </div>
  {/each}
</div>

<style>
  .toast-stack {
    position: fixed; right: calc(18px + var(--safe-right)); bottom: calc(18px + var(--safe-bottom));
    z-index: 100000; display: grid; gap: 10px; width: min(440px, calc(100vw - 24px - var(--safe-left) - var(--safe-right)));
  }
  @media (max-width: 700px) {
    .toast-stack { right: calc(10px + var(--safe-right)); bottom: calc(84px + var(--safe-bottom)); width: calc(100vw - 20px - var(--safe-left) - var(--safe-right)); }
  }
  .toast {
    position: relative; display: flex; gap: 10px; align-items: center; overflow: hidden;
    padding: 13px 15px 17px; border-radius: 0; font-size: .9rem;
    background: var(--brand-surface, #0b0d10); color: var(--brand-ink); opacity:1; backdrop-filter:none;
    border: 1px solid var(--brand-border); border-left: 4px solid var(--brand-primary);
    box-shadow: 0 18px 46px rgba(0,0,0,.4);
  }
  .toast.success { border-left-color: var(--win); }
  .toast.error { border-left-color: var(--loss); }
  .toast.warning { border-left-color: var(--draw); }
  .icon { font-size: 1.1rem; }
  .message { flex: 1; }
  .undo { background: transparent; border: 1px solid var(--brand-primary); color: var(--brand-primary); border-radius: 7px; padding: 4px 9px; cursor: pointer; font-weight: 700; }
  .close { background: transparent; border: none; color: var(--brand-muted); cursor: pointer; font-size: 1.2rem; line-height: 1; padding: 0 2px; }
  .progress { position: absolute; left: 0; bottom: 0; height: 3px; width: 100%; background: var(--brand-primary); transform-origin: left; animation: drain linear forwards; }
  @keyframes drain { from { transform: scaleX(1); } to { transform: scaleX(0); } }
</style>
