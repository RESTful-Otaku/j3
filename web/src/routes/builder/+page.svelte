<script lang="ts">
  import ViewWidget from '../../lib/components/ViewWidget.svelte';
  import { notifySuccess, notifyWarning } from '../../lib/toast';
  import {
    WIDGET_PALETTE, loadViews, saveViews,
    type WidgetType, type ViewWidgetConfig, type SavedView,
  } from '../../lib/view-types';

  let widgets: ViewWidgetConfig[] = [];
  let views: SavedView[] = [];
  let name = '';
  let dragType: WidgetType | null = null;
  let dragIndex = -1;

  function init() {
    views = loadViews();
    if (widgets.length === 0 && views.length > 0) widgets = structuredClone(views[0].widgets);
  }
  init();

  const uid = () => Math.random().toString(36).slice(2, 9);
  function addWidget(type: WidgetType) {
    widgets = [...widgets, { id: uid(), type, span: 1 }];
  }
  function removeWidget(id: string) { widgets = widgets.filter((widget) => widget.id !== id); }
  function resize(id: string, span: 1 | 2 | 3) {
    widgets = widgets.map((widget) => (widget.id === id ? { ...widget, span } : widget));
  }
  function move(from: number, to: number) {
    if (from === to || from < 0 || to < 0) return;
    const list = [...widgets];
    const [item] = list.splice(from, 1);
    list.splice(to, 0, item);
    widgets = list;
  }
  function paletteDrag(event: DragEvent, type: WidgetType) {
    dragType = type;
    dragIndex = -1;
    event.dataTransfer?.setData('text/plain', type);
  }
  function widgetDrag(event: DragEvent, index: number) {
    dragIndex = index;
    dragType = null;
    event.dataTransfer?.setData('text/plain', 'widget');
  }
  function dropOn(index: number, event: DragEvent) {
    event.preventDefault();
    if (dragType) {
      const list = [...widgets];
      list.splice(index, 0, { id: uid(), type: dragType, span: 1 });
      widgets = list;
    } else if (dragIndex >= 0) {
      move(dragIndex, index);
    }
    dragType = null;
    dragIndex = -1;
  }
  function persist(next: SavedView[]) { views = next; saveViews(next); }
  function save() {
    if (!name.trim()) { notifyWarning('Give the view a name first'); return; }
    persist([
      { name: name.trim(), widgets: structuredClone(widgets), updatedAt: new Date().toLocaleString() },
      ...views.filter((view) => view.name !== name.trim()),
    ]);
    notifySuccess(`View “${name}” saved`);
  }
  function load(view: SavedView) { widgets = structuredClone(view.widgets); name = view.name; }
  function clone(view: SavedView) {
    const copy = { name: `${view.name} copy`, widgets: structuredClone(view.widgets), updatedAt: new Date().toLocaleString() };
    persist([copy, ...views]);
    notifySuccess(`Cloned as “${copy.name}”`);
  }
  function remove(nameToDelete: string) {
    persist(views.filter((view) => view.name !== nameToDelete));
    notifySuccess('View deleted');
  }
  function open(nameToOpen: string) { window.open(`/view?name=${encodeURIComponent(nameToOpen)}`, '_blank', 'noopener,noreferrer'); }
</script>

<svelte:head><title>J3 · View builder</title></svelte:head>

<div class="shell">
  <header class="topbar">
    <div class="brand"><a class="brand-mark" href="/" aria-label="Dashboard">J3</a><div><div class="brand-name">View builder</div><div class="brand-sub">Drag widgets, resize, save and open on any screen</div></div></div>
    <div style="display:flex; gap:8px; flex-wrap:wrap;">
      <input class="tab" style="color:var(--brand-ink); cursor:text;" placeholder="View name…" bind:value={name} />
      <button class="btn" on:click={save}>💾 Save view</button>
      <button class="btn primary" on:click={() => open(name || 'current')}>↗ Open in new tab</button>
    </div>
  </header>

  <div class="grid" style="grid-template-columns: 240px 1fr 260px; align-items:start;">
    <div class="card">
      <div class="eyebrow">Palette</div><h2>Widgets</h2>
      <p class="muted small">Drag onto the canvas.</p>
      <div style="display:grid; gap:8px; margin-top:12px;">
        {#each WIDGET_PALETTE as item (item.type)}
          <button class="palette" draggable="true" aria-label="Drag {item.label}"
            on:dragstart={(event) => paletteDrag(event, item.type)}
            on:click={() => addWidget(item.type)}>
            <span>{item.icon}</span> {item.label}
          </button>
        {/each}
      </div>
    </div>

    <div class="card">
      <div class="section-title" style="margin-top:0;">
        <h2>Canvas</h2><span class="pill">{widgets.length} widget{widgets.length === 1 ? '' : 's'}</span>
      </div>
      <div class="canvas" on:dragover={(event) => event.preventDefault()} on:drop={(event) => dropOn(widgets.length, event)} role="list" aria-label="View canvas">
        {#each widgets as widget, index (widget.id)}
          <div class="canvas-item" role="listitem" draggable="true"
            on:dragstart={(event) => widgetDrag(event, index)}
            on:dragover={(event) => event.preventDefault()}
            on:drop={(event) => dropOn(index, event)}>
            <div class="widget-controls">
              <span class="drag-handle" aria-label="Drag to reorder">⠿</span>
              <button class="tab" on:click={() => resize(widget.id, 1)} class:active={widget.span === 1}>S</button>
              <button class="tab" on:click={() => resize(widget.id, 2)} class:active={widget.span === 2}>M</button>
              <button class="tab" on:click={() => resize(widget.id, 3)} class:active={widget.span === 3}>L</button>
              <button class="tab" on:click={() => move(index, Math.max(0, index - 1))}>←</button>
              <button class="tab" on:click={() => move(index, Math.min(widgets.length - 1, index + 1))}>→</button>
              <button class="tab" on:click={() => removeWidget(widget.id)}>✕</button>
            </div>
            <ViewWidget type={widget.type} span={widget.span} />
          </div>
        {:else}
          <div class="placeholder">Drag a widget here to start building your view.</div>
        {/each}
      </div>
    </div>

    <div class="card">
      <div class="eyebrow">Saved views</div><h2>Library</h2>
      <div style="display:grid; gap:8px; margin-top:12px;">
        {#each views as view (view.name)}
          <div class="saved">
            <strong>{view.name}</strong>
            <span class="muted small">{view.widgets.length} widgets · {view.updatedAt}</span>
            <div style="display:flex; gap:5px; flex-wrap:wrap; margin-top:6px;">
              <button class="tab" on:click={() => load(view)}>Load</button>
              <button class="tab" on:click={() => open(view.name)}>Open</button>
              <button class="tab" on:click={() => clone(view)}>Clone</button>
              <button class="tab" on:click={() => remove(view.name)}>Delete</button>
            </div>
          </div>
        {:else}<span class="muted small">No saved views yet.</span>{/each}
      </div>
    </div>
  </div>
</div>

<style>
  .palette { display:flex; align-items:center; gap:8px; padding:10px 12px; border-radius:10px; border:1px solid var(--brand-border); background:var(--brand-surface-3); color:var(--brand-ink); cursor:grab; font-size:.86rem; text-align:left; }
  .palette:hover { border-color:var(--brand-primary); }
  .canvas { display:grid; grid-template-columns:repeat(auto-fit, minmax(240px,1fr)); gap:14px; min-height:280px; padding:12px; border:1px dashed var(--brand-border); border-radius:14px; }
  .canvas-item { position:relative; }
  .canvas-item :global(.widget.wide) { grid-column:span 2; }
  .canvas-item :global(.widget.extra) { grid-column:span 3; }
  .widget-controls { display:flex; gap:4px; align-items:center; flex-wrap:wrap; margin-bottom:6px; }
  .widget-controls .tab { padding:4px 7px; font-size:.72rem; }
  .drag-handle { cursor:grab; padding:3px 6px; color:var(--brand-accent); }
  .saved { padding:10px; border-radius:10px; background:var(--brand-surface-3); border:1px solid var(--brand-border); display:grid; gap:3px; }
  @media (max-width: 1100px) { .grid { grid-template-columns:1fr !important; } }
</style>
