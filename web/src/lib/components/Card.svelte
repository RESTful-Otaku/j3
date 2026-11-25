<script lang="ts">
  import { onMount } from 'svelte';
  import { isPhoneWidth } from '../nav';

  export let title = '';
  export let eyebrow = '';
  export let expandable = true;
  export let expanded = false;
  export let preview = '';

  let isPhone = false;

  onMount(() => {
    isPhone = isPhoneWidth();
  });

  function toggle() {
    if (!isPhone || !expandable) return;
    expanded = !expanded;
  }

  function onKey(event: KeyboardEvent) {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      toggle();
    }
  }

  $: shouldShow = !isPhone || !expandable || expanded;
</script>

<div class="card" class:collapsed={isPhone && expandable && !expanded} aria-expanded={expanded ? 'true' : 'false'} class:open={expanded}>
  {#if isPhone && expandable}
    <div class="card-head" tabindex="0" on:click|stopImmediatePropagation={toggle} on:keydown={onKey} role="button" aria-expanded={expanded ? 'true' : 'false'} aria-label={title}>
      {#if eyebrow}<span class="card-eyebrow">{eyebrow}</span>{/if}
      <span class="card-title">{title}</span>
      {#if preview && !expanded}
        <span class="card-preview">{preview}</span>
      {/if}
      <span class="card-chevron" class:rotated={expanded}>{expanded ? '▾' : '▸'}</span>
    </div>
  {:else}
    {#if eyebrow}<span class="card-eyebrow">{eyebrow}</span>{/if}
    {#if title}<h2 class="card-title">{title}</h2>{/if}
  {/if}

  {#if shouldShow}
    <div class="card-body" class:showing={expanded}><slot /></div>
  {:else}
    <div class="card-body" style="display:none;"></div>
  {/if}
</div>

<style>
  .card {
    display: flex;
    flex-direction: column;
    gap: 0;
    position: relative;
    overflow: hidden;
  }
  .card-head {
    display: flex;
    flex-direction: column;
    gap: 6px;
    cursor: pointer;
    padding: 4px 0;
    z-index: 2;
  }
  .card-head:hover {
    opacity: 0.92;
  }
  .card-eyebrow {
    text-transform: uppercase;
    letter-spacing: 0.16em;
    font-size: 0.68rem;
    font-weight: 900;
    color: var(--brand-accent);
  }
  .card-title {
    font-family: var(--font-display);
    font-weight: 800;
    font-size: 1.05rem;
    letter-spacing: -0.01em;
    min-width: 0;
    overflow-wrap: anywhere;
  }
  .card-preview {
    font-size: 0.82rem;
    color: var(--brand-muted);
    min-width: 0;
    overflow-wrap: anywhere;
  }
  .card-chevron {
    font-size: 0.7rem;
    font-weight: 800;
    color: var(--brand-muted);
    flex: none;
    transition: transform 0.22s cubic-bezier(0.2, 0, 0, 1);
  }
  .card-chevron.rotated { transform: rotate(180deg); }
  .card-body {
    padding-top: 10px;
    border-top: 1px solid var(--brand-border);
    min-width: 0;
    overflow: hidden;
    animation: cardFadeIn 0.34s cubic-bezier(0.18, 0.88, 0.32, 1.12);
  }
  @keyframes cardFadeIn {
    from { opacity: 0; transform: translateY(6px); }
    to   { opacity: 1; transform: translateY(0); }
  }
  @media (max-width: 700px) {
    .card-head { cursor: pointer; }
    .card.collapsed { border-color: var(--brand-border); }
    .card.collapsed .card-head:hover { background: color-mix(in srgb, var(--brand-primary) 6%, transparent); }
    .card.collapsed .card-body { display: none; }
    .card.open { border-color: var(--brand-primary); box-shadow: 0 10px 34px color-mix(in srgb, var(--brand-primary) 28%, transparent); }
  }
</style>
