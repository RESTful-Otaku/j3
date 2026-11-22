<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { goto } from '$app/navigation';
  import ViewWidget from '../../lib/components/ViewWidget.svelte';
  import { notifySuccess, notifyWarning } from '../../lib/toast';
  import {
    WIDGET_PALETTE, clampSize, insertionIndex, loadViews, saveViews, sizeFromDelta,
    type SavedView, type ViewWidgetConfig, type WidgetSize, type WidgetType,
  } from '../../lib/view-types';

  type DragState = {
    kind: 'palette' | 'reorder';
    pointerId: number;
    type?: WidgetType;
    id?: string;
    from: number;
    startX: number;
    startY: number;
    x: number;
    y: number;
    moved: boolean;
    dropIndex: number;
  };
  type SizeAxis = 'x' | 'y' | 'xy';
  type SizingState = {
    pointerId: number;
    id: string;
    axis: SizeAxis;
    label: string;
    startX: number;
    startY: number;
    x: number;
    y: number;
    startSpan: WidgetSize;
    startRows: WidgetSize;
    span: WidgetSize;
    rows: WidgetSize;
  };
  type Ghost = { x: number; y: number; label: string; detail: string };

  let widgets: ViewWidgetConfig[] = [];
  let views: SavedView[] = [];
  let name = '';
  let canvasEl: HTMLDivElement;
  let drag: DragState | null = null;
  let sizing: SizingState | null = null;
  let suppressPaletteClick = false;
  let reduceMotion = false;

  onMount(() => {
    reduceMotion = window.matchMedia?.('(prefers-reduced-motion: reduce)').matches ?? false;
    init();
  });

  function init() {
    views = loadViews();
    if (widgets.length === 0 && views.length > 0) widgets = structuredClone(views[0].widgets);
  }

  const uid = () => Math.random().toString(36).slice(2, 9);
  const sizeLabel = (size: WidgetSize) => (size === 3 ? 'L' : size === 2 ? 'M' : 'S');
  const paletteLabel = (type?: WidgetType) => WIDGET_PALETTE.find((item) => item.type === type)?.label ?? 'Widget';
  const widgetLabel = (id?: string) => widgets.find((widget) => widget.id === id)?.type
    ? paletteLabel(widgets.find((widget) => widget.id === id)?.type)
    : 'Widget';
  const previewSpan = (widget: ViewWidgetConfig) => (sizing?.id === widget.id ? sizing.span : widget.span);
  const previewRows = (widget: ViewWidgetConfig) => (sizing?.id === widget.id ? sizing.rows : widget.rows);

  $: ghost = ((): Ghost | null => {
    if (drag) {
      const label = drag.type ? paletteLabel(drag.type) : widgetLabel(drag.id);
      return {
        x: drag.x, y: drag.y, label,
        detail: drag.moved ? `Drop at position ${drag.dropIndex + 1}` : 'Drag to place',
      };
    }
    if (sizing) {
      return {
        x: sizing.x, y: sizing.y, label: sizing.label,
        detail: `${sizeLabel(sizing.span)} · ${sizing.rows} row${sizing.rows === 1 ? '' : 's'}`,
      };
    }
    return null;
  })();

  function addWidget(type: WidgetType) {
    widgets = [...widgets, { id: uid(), type, span: 1, rows: 1 }];
  }
  function insertWidgetAt(type: WidgetType, index: number) {
    const list = [...widgets];
    list.splice(Math.max(0, Math.min(index, list.length)), 0, { id: uid(), type, span: 1, rows: 1 });
    widgets = list;
  }
  function removeWidget(id: string) { widgets = widgets.filter((widget) => widget.id !== id); }
  function resize(id: string, span: WidgetSize) {
    widgets = widgets.map((widget) => (widget.id === id ? { ...widget, span } : widget));
  }
  function resizeHeight(id: string, rows: WidgetSize) {
    widgets = widgets.map((widget) => (widget.id === id ? { ...widget, rows } : widget));
  }
  function applySize(id: string, span: WidgetSize, rows: WidgetSize) {
    widgets = widgets.map((widget) => (widget.id === id ? { ...widget, span, rows } : widget));
  }
  function move(from: number, to: number) {
    if (from === to || from < 0 || to < 0 || from >= widgets.length) return;
    const list = [...widgets];
    const [item] = list.splice(from, 1);
    list.splice(Math.max(0, Math.min(to, list.length)), 0, item);
    widgets = list;
  }

  function gridMetrics() {
    if (!canvasEl) return { column: 240, row: 130, gapX: 14, gapY: 14 };
    const styles = getComputedStyle(canvasEl);
    const positive = (value: string) => Number.parseFloat(value) || 0;
    const columns = styles.gridTemplateColumns.split(' ').map(positive).filter((value) => value > 0);
    const rows = styles.gridTemplateRows.split(' ').map(positive).filter((value) => value > 0);
    return {
      column: columns[0] || 240,
      row: rows[0] || 130,
      gapX: positive(styles.columnGap) || 14,
      gapY: positive(styles.rowGap) || 14,
    };
  }

  function itemRects() {
    if (!canvasEl) return [];
    return Array.from(canvasEl.querySelectorAll<HTMLElement>('[data-widget-id]')).map((element) => {
      const rect = element.getBoundingClientRect();
      return { x: rect.left, y: rect.top, width: rect.width, height: rect.height };
    });
  }

  function snapshotRects() {
    if (!canvasEl) return new Map<string, DOMRect>();
    return new Map(
      Array.from(canvasEl.querySelectorAll<HTMLElement>('[data-widget-id]'))
        .map((element) => [element.dataset.widgetId ?? '', element.getBoundingClientRect()] as const)
        .filter(([id]) => Boolean(id)),
    );
  }

  async function withSnap(mutate: () => void) {
    const before = snapshotRects();
    mutate();
    await tick();
    if (reduceMotion) return;
    before.forEach((rect, id) => {
      const element = canvasEl?.querySelector<HTMLElement>(`[data-widget-id="${id}"]`);
      if (!element) return;
      const next = element.getBoundingClientRect();
      const dx = rect.left - next.left;
      const dy = rect.top - next.top;
      const sx = next.width ? rect.width / next.width : 1;
      const sy = next.height ? rect.height / next.height : 1;
      if (Math.abs(dx) < 2 && Math.abs(dy) < 2 && Math.abs(sx - 1) < 0.02 && Math.abs(sy - 1) < 0.02) return;
      element.animate(
        [
          { transform: `translate(${dx}px, ${dy}px) scale(${sx}, ${sy})` },
          { transform: 'none' },
        ],
        { duration: 260, easing: 'cubic-bezier(.16,1,.3,1)' },
      );
    });
  }

  function beginPalette(event: PointerEvent, type: WidgetType) {
    if (event.button !== 0) return;
    drag = {
      kind: 'palette', pointerId: event.pointerId, type, from: -1,
      startX: event.clientX, startY: event.clientY, x: event.clientX, y: event.clientY,
      moved: false, dropIndex: widgets.length,
    };
  }

  function beginReorder(event: PointerEvent, index: number, id: string) {
    if (event.button !== 0) return;
    event.preventDefault();
    event.stopPropagation();
    drag = {
      kind: 'reorder', pointerId: event.pointerId, id, from: index,
      startX: event.clientX, startY: event.clientY, x: event.clientX, y: event.clientY,
      moved: false, dropIndex: index,
    };
  }

  function beginResize(event: PointerEvent, widget: ViewWidgetConfig, axis: SizeAxis) {
    if (event.button !== 0) return;
    event.preventDefault();
    event.stopPropagation();
    sizing = {
      pointerId: event.pointerId, id: widget.id, axis,
      label: paletteLabel(widget.type),
      startX: event.clientX, startY: event.clientY, x: event.clientX, y: event.clientY,
      startSpan: widget.span, startRows: widget.rows, span: widget.span, rows: widget.rows,
    };
  }

  function onPointerMove(event: PointerEvent) {
    if (drag && event.pointerId === drag.pointerId) {
      const moved = drag.moved || Math.hypot(event.clientX - drag.startX, event.clientY - drag.startY) > 5;
      const dropIndex = moved
        ? insertionIndex({ x: event.clientX, y: event.clientY }, itemRects())
        : drag.dropIndex;
      drag = { ...drag, x: event.clientX, y: event.clientY, moved, dropIndex };
      return;
    }
    if (sizing && event.pointerId === sizing.pointerId) {
      const metrics = gridMetrics();
      const span = sizing.axis === 'y'
        ? sizing.startSpan
        : sizeFromDelta(sizing.startSpan, event.clientX - sizing.startX, metrics.column, metrics.gapX);
      const rows = sizing.axis === 'x'
        ? sizing.startRows
        : sizeFromDelta(sizing.startRows, event.clientY - sizing.startY, metrics.row, metrics.gapY);
      sizing = { ...sizing, x: event.clientX, y: event.clientY, span, rows };
    }
  }

  async function commitDrag(finished: DragState) {
    if (!finished.moved) return;
    suppressPaletteClick = true;
    if (finished.kind === 'palette' && finished.type) {
      await withSnap(() => insertWidgetAt(finished.type!, finished.dropIndex));
    } else if (finished.kind === 'reorder') {
      await withSnap(() => move(finished.from, finished.dropIndex));
    }
    setTimeout(() => (suppressPaletteClick = false), 0);
  }

  function onPointerUp(event: PointerEvent) {
    if (drag && event.pointerId === drag.pointerId) {
      const finished = drag;
      drag = null;
      void commitDrag(finished);
      return;
    }
    if (sizing && event.pointerId === sizing.pointerId) {
      const finished = sizing;
      sizing = null;
      void withSnap(() => applySize(finished.id, finished.span, finished.rows));
    }
  }

  function onPointerCancel(event: PointerEvent) {
    if (drag?.pointerId === event.pointerId) drag = null;
    if (sizing?.pointerId === event.pointerId) sizing = null;
  }

  function onSizeKey(event: KeyboardEvent, widget: ViewWidgetConfig, axis: SizeAxis) {
    const delta = event.key === 'ArrowRight' || event.key === 'ArrowDown' ? 1 : event.key === 'ArrowLeft' || event.key === 'ArrowUp' ? -1 : 0;
    if (!delta) return;
    event.preventDefault();
    if (axis !== 'y') resize(widget.id, clampSize(widget.span + delta));
    if (axis !== 'x') resizeHeight(widget.id, clampSize(widget.rows + delta));
  }

  function paletteClick(type: WidgetType) {
    if (suppressPaletteClick) { suppressPaletteClick = false; return; }
    addWidget(type);
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
  function open(nameToOpen: string) { void goto(`/view?name=${encodeURIComponent(nameToOpen)}`); }
  function openNewTab(nameToOpen: string) {
    window.open(`/view?name=${encodeURIComponent(nameToOpen)}`, '_blank', 'noopener,noreferrer');
  }
</script>

<svelte:window on:pointermove={onPointerMove} on:pointerup={onPointerUp} on:pointercancel={onPointerCancel} />

<svelte:head><title>J3 · View builder</title></svelte:head>

<div class="shell">
  <header class="topbar">
    <div class="brand"><a class="brand-mark" href="/public" aria-label="Public dashboard">J3</a><div><div class="brand-name">View builder</div><div class="brand-sub">Drag widgets, resize, save and open on any screen</div></div></div>
    <div style="display:flex; gap:8px; flex-wrap:wrap;">
      <input class="tab" style="color:var(--brand-ink); cursor:text;" placeholder="View name…" bind:value={name} />
      <button class="btn" on:click={save}>💾 Save view</button>
      <button class="btn primary" on:click={() => openNewTab(name || 'current')}>↗ Open in new tab</button>
    </div>
  </header>

  <div class="grid" style="grid-template-columns: 240px 1fr 260px; align-items:start;">
    <div class="card">
      <div class="eyebrow">Palette</div><h2>Widgets</h2>
      <p class="muted small">Click to add, or drag onto the canvas to choose a position.</p>
      <div style="display:grid; gap:8px; margin-top:12px;">
        {#each WIDGET_PALETTE as item (item.type)}
          <button class="palette" data-widget-type={item.type} aria-label="Add {item.label}"
            on:pointerdown={(event) => beginPalette(event, item.type)}
            on:click={() => paletteClick(item.type)}>
            <span>{item.icon}</span> {item.label}
          </button>
        {/each}
      </div>
    </div>

    <div class="card">
      <div class="section-title" style="margin-top:0;">
        <h2>Canvas</h2><span class="pill">{widgets.length} widget{widgets.length === 1 ? '' : 's'}</span>
      </div>
      <div
        class="canvas"
        class:dragging={Boolean(drag?.moved)}
        class:resizing={Boolean(sizing)}
        class:drop-end={Boolean(drag?.moved && drag.dropIndex === widgets.length)}
        bind:this={canvasEl}
        role="list"
        aria-label="View canvas"
      >
        {#each widgets as widget, index (widget.id)}
          <div
            class="canvas-item"
            class:wide={previewSpan(widget) === 2}
            class:extra={previewSpan(widget) === 3}
            class:tall={previewRows(widget) === 2}
            class:xtall={previewRows(widget) === 3}
            class:dragging={drag?.moved && drag.id === widget.id}
            class:drop-target={drag?.moved && drag.dropIndex === index}
            class:resizing={sizing?.id === widget.id}
            data-widget-id={widget.id}
            data-widget-type={widget.type}
            data-span={previewSpan(widget)}
            data-rows={previewRows(widget)}
            role="listitem"
          >
            <div class="widget-controls">
              <button class="drag-handle" aria-label="Drag {paletteLabel(widget.type)} to reorder"
                title="Drag to reorder"
                on:pointerdown={(event) => beginReorder(event, index, widget.id)}
                on:keydown={(event) => {
                  if (event.key === 'ArrowLeft') { event.preventDefault(); move(index, Math.max(0, index - 1)); }
                  if (event.key === 'ArrowRight') { event.preventDefault(); move(index, Math.min(widgets.length - 1, index + 1)); }
                }}>⠿</button>
              <button class="tab" on:click={() => resize(widget.id, 1)} class:active={widget.span === 1}>S</button>
              <button class="tab" on:click={() => resize(widget.id, 2)} class:active={widget.span === 2}>M</button>
              <button class="tab" on:click={() => resize(widget.id, 3)} class:active={widget.span === 3}>L</button>
              <button class="tab" on:click={() => resizeHeight(widget.id, 1)} class:active={widget.rows === 1}>H1</button>
              <button class="tab" on:click={() => resizeHeight(widget.id, 2)} class:active={widget.rows === 2}>H2</button>
              <button class="tab" on:click={() => resizeHeight(widget.id, 3)} class:active={widget.rows === 3}>H3</button>
              <button class="tab" on:click={() => move(index, Math.max(0, index - 1))}>←</button>
              <button class="tab" on:click={() => move(index, Math.min(widgets.length - 1, index + 1))}>→</button>
              <button class="tab" on:click={() => removeWidget(widget.id)}>✕</button>
            </div>
            <ViewWidget type={widget.type} span={previewSpan(widget)} />
            <button class="resize-handle-x" aria-label="Drag to resize {paletteLabel(widget.type)} width"
              title="Drag sideways or use ←/→" data-resize-axis="x"
              on:pointerdown={(event) => beginResize(event, widget, 'x')}
              on:keydown={(event) => onSizeKey(event, widget, 'x')}></button>
            <button class="resize-handle-y" aria-label="Drag to resize {paletteLabel(widget.type)} height"
              title="Drag vertically or use ↑/↓" data-resize-axis="y"
              on:pointerdown={(event) => beginResize(event, widget, 'y')}
              on:keydown={(event) => onSizeKey(event, widget, 'y')}></button>
            <button class="resize-handle-corner" aria-label="Drag to resize {paletteLabel(widget.type)} in both directions"
              title="Drag diagonally" data-resize-axis="xy"
              on:pointerdown={(event) => beginResize(event, widget, 'xy')}
              on:keydown={(event) => onSizeKey(event, widget, 'xy')}></button>
          </div>
        {:else}
          <div class="placeholder">Drag a widget here to start building your view.</div>
        {/each}
      </div>
      <p class="muted small" style="margin-top:10px;">Tip: pull the right or bottom edge of a card to resize; pull the ⠿ handle to reorder.</p>
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
              <button class="tab" on:click={() => openNewTab(view.name)}>↗</button>
              <button class="tab" on:click={() => clone(view)}>Clone</button>
              <button class="tab" on:click={() => remove(view.name)}>Delete</button>
            </div>
          </div>
        {:else}<span class="muted small">No saved views yet.</span>{/each}
      </div>
    </div>
  </div>
</div>

{#if ghost}
  <div class="drag-ghost" style="left:{ghost.x + 16}px; top:{ghost.y + 16}px;" aria-hidden="true">
    <strong>{ghost.label}</strong><span>{ghost.detail}</span>
  </div>
{/if}

<style>
  .palette { display:flex; align-items:center; gap:8px; padding:10px 12px; border-radius:10px; border:1px solid var(--brand-border); background:var(--brand-surface-3); color:var(--brand-ink); cursor:grab; font-size:.86rem; text-align:left; touch-action:none; }
  .palette:hover, .palette:focus-visible { border-color:var(--brand-primary); }
  .canvas { display:grid; grid-template-columns:repeat(auto-fit, minmax(240px,1fr)); grid-auto-rows:minmax(130px,auto); gap:14px; min-height:280px; padding:12px; border:1px dashed var(--brand-border); border-radius:14px; transition:box-shadow .2s ease; }
  .canvas.drop-end { box-shadow:inset 0 -4px 0 var(--brand-primary); }
  .canvas.dragging, .canvas.resizing { user-select:none; }
  .canvas-item { position:relative; min-width:0; transition:opacity .18s ease, box-shadow .18s ease; }
  .canvas-item.wide, .canvas-item :global(.widget.wide) { grid-column:span 2; }
  .canvas-item.extra, .canvas-item :global(.widget.extra) { grid-column:span 3; }
  .canvas-item.tall { grid-row:span 2; }
  .canvas-item.xtall { grid-row:span 3; }
  .canvas-item.dragging { opacity:.45; }
  .canvas-item.resizing { box-shadow:0 0 0 2px var(--brand-primary); }
  .canvas-item.drop-target { box-shadow:inset 0 3px 0 var(--brand-primary); }
  .widget-controls { display:flex; gap:4px; align-items:center; flex-wrap:wrap; margin-bottom:6px; }
  .widget-controls .tab { padding:4px 7px; font-size:.72rem; }
  .drag-handle { padding:3px 7px; border:1px solid var(--brand-border); border-radius:7px; background:var(--brand-surface-3); color:var(--brand-accent); cursor:grab; touch-action:none; }
  .drag-handle:active { cursor:grabbing; }
  .resize-handle-x, .resize-handle-y, .resize-handle-corner {
    position:absolute; z-index:3; border:0; padding:0; background:transparent;
    touch-action:none; transition:background .16s ease, transform .16s ease;
  }
  .resize-handle-x { top:42px; bottom:10px; right:-7px; width:16px; cursor:ew-resize; border-radius:8px; }
  .resize-handle-y { left:10px; right:10px; bottom:-7px; height:16px; cursor:ns-resize; border-radius:8px; }
  .resize-handle-corner { right:-7px; bottom:-7px; width:20px; height:20px; cursor:nwse-resize; border-radius:50%; background:var(--brand-primary); box-shadow:0 2px 10px rgba(0,0,0,.35); }
  .resize-handle-x::after, .resize-handle-y::after { content:'⋮'; position:absolute; inset:0; display:grid; place-items:center; color:var(--brand-muted); font-size:.75rem; opacity:0; transition:opacity .16s ease; }
  .resize-handle-y::after { content:'⋯'; }
  .resize-handle-x:hover, .resize-handle-y:hover, .resize-handle-x:focus-visible, .resize-handle-y:focus-visible { background:rgba(56,189,248,.18); }
  .resize-handle-x:hover::after, .resize-handle-y:hover::after, .resize-handle-x:focus-visible::after, .resize-handle-y:focus-visible::after { opacity:1; }
  .resize-handle-corner:hover { transform:scale(1.18); }
  .drag-ghost { position:fixed; z-index:860; pointer-events:none; display:grid; gap:2px; padding:9px 13px; border-radius:11px; background:var(--brand-primary); color:#fff; box-shadow:0 14px 36px rgba(0,0,0,.45); font-size:.78rem; animation:ghostIn .18s ease-out both; }
  .drag-ghost span { opacity:.85; font-size:.7rem; }
  .saved { padding:10px; border-radius:10px; background:var(--brand-surface-3); border:1px solid var(--brand-border); display:grid; gap:3px; }
  @keyframes ghostIn { from { transform:scale(.86); opacity:0; } to { transform:none; opacity:1; } }
  @media (max-width: 1100px) { .grid { grid-template-columns:1fr !important; } }
</style>
