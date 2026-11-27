<script lang="ts">
  import { pageCount, clampPage } from '../paginate';

  export let total = 0;
  export let page = 1;          // 1-based
  export let pageSize = 10;
  export let pageSizeOptions: number[] = [10, 25, 50, 100];
  export let onChange: (page: number, pageSize: number) => void = () => {};

  $: pages = pageCount(total, pageSize);
  $: page = clampPage(page, total, pageSize);
  $: from = total === 0 ? 0 : (page - 1) * pageSize + 1;
  $: to = Math.min(total, page * pageSize);

  function setPage(p: number) { page = clampPage(p, total, pageSize); onChange(page, pageSize); }
  function setSize(s: number) { pageSize = s; page = 1; onChange(page, pageSize); }
</script>

{#if total > 0}
<div class="pagination" role="navigation" aria-label="Pagination">
  <div class="info">{from}–{to} of {total}</div>
  <div class="controls">
    <button class="page-btn" aria-label="First page" disabled={page <= 1} on:click={() => setPage(1)}>«</button>
    <button class="page-btn" aria-label="Previous page" disabled={page <= 1} on:click={() => setPage(page - 1)}>‹</button>
    <span class="page-indicator">{page} / {pages}</span>
    <button class="page-btn" aria-label="Next page" disabled={page >= pages} on:click={() => setPage(page + 1)}>›</button>
    <button class="page-btn" aria-label="Last page" disabled={page >= pages} on:click={() => setPage(pages)}>»</button>
  </div>
  <label class="size">Rows
    <select bind:value={pageSize} on:change={(e) => setSize(Number((e.target as HTMLSelectElement).value))} aria-label="Rows per page">
      {#each pageSizeOptions as opt}<option value={opt}>{opt}</option>{/each}
    </select>
  </label>
</div>
{/if}

<style>
  .pagination { display:flex; flex-wrap:wrap; align-items:center; justify-content:space-between; gap:10px; margin:14px 0; padding:10px 12px; border-radius:12px; border:1px solid var(--brand-border); background:var(--brand-surface-2); }
  .info { font-size:.78rem; color:var(--brand-muted); }
  .controls { display:flex; align-items:center; gap:6px; }
  .page-btn { min-width:44px; min-height:44px; border-radius:10px; border:1px solid var(--brand-border); background:var(--brand-surface-3); color:var(--brand-ink); font-size:1rem; font-weight:800; cursor:pointer; }
  .page-btn:disabled { opacity:.4; cursor:not-allowed; }
  .page-indicator { font-size:.8rem; font-weight:800; min-width:64px; text-align:center; }
  .size { display:flex; align-items:center; gap:6px; font-size:.76rem; color:var(--brand-muted); }
  .size select { padding:8px 10px; border-radius:9px; border:1px solid var(--brand-border); background:var(--brand-surface-3); color:var(--brand-ink); min-height:44px; }
  @media (max-width:700px){ .pagination{ flex-direction:column; align-items:stretch; } .controls{ justify-content:center; } .size{ justify-content:center; } }
</style>
