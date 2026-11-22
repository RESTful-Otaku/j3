<script lang="ts">
  import { tick } from 'svelte';

  export let open = false;
  export let title = 'Please confirm';
  export let message = '';
  export let confirmLabel = 'Confirm';
  export let danger = false;
  export let onConfirm: () => void = () => {};
  export let onCancel: () => void = () => {};

  let dialog: HTMLDivElement;
  let cancelButton: HTMLButtonElement;
  let previouslyFocused: HTMLElement | null = null;

  async function focusDialog() {
    previouslyFocused = document.activeElement as HTMLElement | null;
    await tick();
    cancelButton?.focus();
  }

  function close(cancelled: boolean) {
    open = false;
    if (cancelled) onCancel();
    previouslyFocused?.focus();
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
      event.preventDefault();
      close(true);
      return;
    }
    if (event.key !== 'Tab') return;
    const focusable = Array.from(
      dialog.querySelectorAll<HTMLElement>('button:not([disabled]), [href], input, select, textarea'),
    );
    if (focusable.length === 0) return;
    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault();
      last.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first.focus();
    }
  }

  $: if (open) void focusDialog();
</script>

<svelte:window on:keydown={open ? handleKeydown : () => {}} />

{#if open}
  <div class="backdrop" role="presentation" on:mousedown={(event) => event.target === event.currentTarget && close(true)}>
    <div class="dialog" role="dialog" aria-modal="true" aria-label={title} aria-describedby="confirm-message" tabindex="-1"
      bind:this={dialog} on:mousedown|stopPropagation>
      <div class="eyebrow">Confirmation</div>
      <h2>{title}</h2>
      <p id="confirm-message">{message}</p>
      <div class="actions">
        <button class="btn" bind:this={cancelButton} on:click={() => close(true)}>Cancel</button>
        <button class="btn" data-guide="confirm-action" class:primary={!danger} class:danger={danger} on:click={() => { onConfirm(); close(false); }}>
          {confirmLabel}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .backdrop { position: fixed; inset: 0; z-index: 1000; display: grid; place-items: center; background: rgba(2,6,23,.72); backdrop-filter: blur(6px); padding: 20px; }
  .dialog { width: min(460px, 100%); padding: 26px; border-radius: 20px; background: var(--brand-surface-2); border: 1px solid var(--brand-border); box-shadow: 0 30px 90px rgba(0,0,0,.55); animation: pop .22s ease-out; }
  .eyebrow { text-transform: uppercase; letter-spacing: .14em; font-size: .7rem; font-weight: 700; color: var(--brand-accent); }
  h2 { margin: 8px 0 10px; }
  p { color: var(--brand-muted); font-size: .92rem; }
  .actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 22px; }
  .btn { padding: 10px 16px; border-radius: 10px; border: 1px solid var(--brand-border); background: var(--brand-surface-3); color: var(--brand-ink); cursor: pointer; }
  .btn.primary { background: var(--brand-primary); color: #fff; border-color: transparent; }
  .btn.danger { background: var(--loss); color: #fff; border-color: transparent; }
  @keyframes pop { from { transform: scale(.92) translateY(12px); opacity: 0; } to { transform: none; opacity: 1; } }
</style>
