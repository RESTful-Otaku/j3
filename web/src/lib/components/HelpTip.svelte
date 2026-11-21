<script lang="ts">
  export let text = '';
  let open = false;
</script>

<span class="help-tip">
  <button aria-label="Help" on:mouseenter={() => (open = true)} on:mouseleave={() => (open = false)}
    on:focus={() => (open = true)} on:blur={() => (open = false)} on:click={() => (open = !open)}>?</button>
  {#if open}<span class="tip" role="tooltip">{text}</span>{/if}
</span>

<style>
  .help-tip { position:relative; display:inline-flex; vertical-align:middle; }
  button { width:18px; height:18px; border-radius:50%; border:1px solid var(--brand-border); background:var(--brand-surface-3); color:var(--brand-muted); font-size:.7rem; cursor:help; padding:0; }
  .tip { position:absolute; bottom:calc(100% + 8px); left:50%; transform:translateX(-50%); width:max-content; max-width:240px; padding:8px 10px; border-radius:8px; background:var(--brand-ink); color:var(--brand-surface); font-size:.76rem; line-height:1.35; z-index:100; box-shadow:0 8px 24px rgba(0,0,0,.3); }
</style>
