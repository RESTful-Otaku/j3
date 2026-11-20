<script lang="ts">
  import { onMount } from 'svelte';
  import ViewWidget from '../../lib/components/ViewWidget.svelte';
  import { loadViews, type SavedView } from '../../lib/view-types';

  let view: SavedView | null = null;
  let message = '';

  onMount(() => {
    const name = new URLSearchParams(location.search).get('name');
    const found = loadViews().find((item) => item.name === name);
    if (found) view = found;
    else message = name ? `View “${name}” was not found on this device.` : 'No view selected.';
  });
</script>

<svelte:head><title>J3 · {view?.name ?? 'Custom view'}</title></svelte:head>

<div class="display-shell">
  <header class="display-header">
    <div><div class="event-mark"><span class="crest">J3</span>{view?.name ?? 'Custom view'}</div>
      <div class="meta">Live custom dashboard · {view?.widgets.length ?? 0} widgets</div></div>
    <a class="btn" href="/builder" target="_blank" rel="noreferrer">✏️ Edit view ↗</a>
  </header>

  {#if view}
    <div class="view-grid">
      {#each view.widgets as widget (widget.id)}
        <div class:wide={widget.span === 2} class:extra={widget.span === 3} class:tall={widget.rows === 2} class:xtall={widget.rows === 3}>
          <ViewWidget type={widget.type} title={widget.title} span={widget.span} />
        </div>
      {/each}
    </div>
  {:else}
    <div class="placeholder">{message}</div>
  {/if}
</div>

<style>
  .view-grid { display:grid; grid-template-columns:repeat(auto-fit, minmax(260px,1fr)); gap:18px; }
  .wide { grid-column:span 2; }
  .extra { grid-column:span 3; }
  .tall { grid-row:span 2; }
  .xtall { grid-row:span 3; }
  @media (max-width: 900px) { .wide, .extra { grid-column:span 1; } }
</style>
