<script lang="ts">
  export let open = false;
  export let title = 'Please confirm';
  export let message = '';
  export let confirmLabel = 'Confirm';
  export let danger = false;
  export let onConfirm: () => void = () => {};
</script>

{#if open}
  <div class="backdrop" role="presentation" on:click={() => (open = false)}>
    <div class="dialog" role="dialog" aria-modal="true" aria-label={title} tabindex="-1" on:click|stopPropagation on:keydown={(e) => e.key === 'Escape' && (open = false)}>
      <div class="eyebrow">Confirmation</div>
      <h2>{title}</h2>
      <p>{message}</p>
      <div class="actions">
        <button class="btn" on:click={() => (open = false)}>Cancel</button>
        <button class="btn" class:primary={!danger} class:danger={danger} on:click={() => { onConfirm(); open = false; }}>
          {confirmLabel}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .backdrop {
    position: fixed; inset: 0; z-index: 1000; display: grid; place-items: center;
    background: rgba(2, 6, 23, .72); backdrop-filter: blur(6px); padding: 20px;
  }
  .dialog {
    width: min(460px, 100%); padding: 26px; border-radius: 20px;
    background: var(--brand-surface-2); border: 1px solid var(--brand-border);
    box-shadow: 0 30px 90px rgba(0,0,0,.55); animation: pop .22s ease-out;
  }
  .eyebrow { text-transform: uppercase; letter-spacing: .14em; font-size: .7rem; font-weight: 700; color: var(--brand-accent); }
  h2 { margin: 8px 0 10px; }
  p { color: var(--brand-muted); font-size: .92rem; }
  .actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 22px; }
  .btn { padding: 10px 16px; border-radius: 10px; border: 1px solid var(--brand-border); background: var(--brand-surface-3); color: var(--brand-ink); cursor: pointer; }
  .btn.primary { background: var(--brand-primary); color: #fff; border-color: transparent; }
  .btn.danger { background: var(--loss); color: #fff; border-color: transparent; }
  @keyframes pop { from { transform: scale(.92) translateY(12px); opacity: 0; } to { transform: none; opacity: 1; } }
</style>
