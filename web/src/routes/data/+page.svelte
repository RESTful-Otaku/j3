<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '../../lib/api';
  import { notifyError } from '../../lib/toast';

  const tables = ['fighters', 'clubs', 'events', 'categories', 'fights', 'pools', 'belt_ranks', 'users', 'audit_log'];
  let table = 'fighters';
  let search = '';
  let sort = '';
  let order: 'asc' | 'desc' = 'asc';
  let limit = 100;
  let offset = 0;
  let filterColumn = '';
  let filterValue = '';
  let columns: string[] = [];
  let rows: Record<string, unknown>[] = [];
  let count = 0;

  async function load() {
    const params: Record<string, string | number> = { table, limit, offset, order };
    if (search) params.search = search;
    if (sort) params.sort = sort;
    if (filterColumn && filterValue) params[filterColumn] = filterValue;
    try {
      const result = await api.query(params);
      columns = result.columns;
      rows = result.rows;
      count = result.count;
      if (!sort) sort = columns[0] ?? '';
    } catch (error) {
      notifyError(error instanceof Error ? error.message : 'Query failed');
    }
  }
  function sortBy(column: string) {
    if (sort === column) order = order === 'asc' ? 'desc' : 'asc';
    else { sort = column; order = 'asc'; }
    void load();
  }
  function changeTable() {
    sort = ''; offset = 0; filterColumn = ''; filterValue = '';
    void load();
  }
  onMount(() => { void load(); });

</script>

<svelte:head><title>J3 · Data explorer</title></svelte:head>

<div class="shell">
  <header class="topbar">
    <div class="brand"><div class="brand-mark">J3</div><div><div class="brand-name">Data explorer</div><div class="brand-sub">Sort, filter and search every dataset</div></div></div>
    <a class="btn" href="/control" target="_blank" rel="noreferrer">🖥️ Displays ↗</a>
  </header>

  <div class="card">
    <div class="controls">
      <select bind:value={table} on:change={changeTable}>{#each tables as item}<option value={item}>{item}</option>{/each}</select>
      <input bind:value={search} placeholder="Search…" on:input={() => { offset = 0; void load(); }} />
      <select bind:value={filterColumn}><option value="">Filter column…</option>{#each columns as column}<option value={column}>{column}</option>{/each}</select>
      <input bind:value={filterValue} placeholder="Value" on:input={() => { offset = 0; void load(); }} />
      <select bind:value={sort} on:change={() => void load()}>{#each columns as column}<option value={column}>Sort: {column}</option>{/each}</select>
      <button class="btn" on:click={() => void load()}>Apply</button>
    </div>
    <p class="muted small">{count} row{count === 1 ? '' : 's'} · page {Math.floor(offset / limit) + 1}</p>
    <div style="overflow-x:auto; margin-top:12px;">
      <table>
        <thead><tr>{#each columns as column}<th><button on:click={() => sortBy(column)}>{column} {sort === column ? (order === 'asc' ? '↑' : '↓') : ''}</button></th>{/each}</tr></thead>
        <tbody>
          {#each rows as row}
            <tr>{#each columns as column}<td>{row[column] ?? '—'}</td>{/each}</tr>
          {:else}<tr><td colspan={columns.length || 1} class="muted">No matching rows.</td></tr>{/each}
        </tbody>
      </table>
    </div>
    <div class="pager">
      <button class="btn" disabled={offset === 0} on:click={() => { offset = Math.max(0, offset - limit); void load(); }}>← Previous</button>
      <button class="btn" disabled={count < limit} on:click={() => { offset += limit; void load(); }}>Next →</button>
    </div>
  </div>
</div>

<style>
  .controls { display:flex; gap:9px; flex-wrap:wrap; }
  .controls select, .controls input { padding:9px 11px; border-radius:9px; border:1px solid var(--brand-border); background:var(--brand-surface-3); color:var(--brand-ink); font-size:.9rem; }
  table { width:100%; border-collapse:collapse; font-size:.85rem; }
  th { text-align:left; padding:8px; border-bottom:1px solid var(--brand-border); }
  th button { background:transparent; border:none; color:var(--brand-primary); cursor:pointer; font-weight:700; }
  td { padding:8px; border-bottom:1px solid var(--brand-border); }
  .pager { display:flex; gap:10px; margin-top:16px; }
</style>
