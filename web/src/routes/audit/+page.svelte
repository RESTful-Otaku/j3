<script lang="ts">
  import { onMount } from 'svelte';
  import { api, type AuditEntry } from '../../lib/api';
  import { Pagination } from '../../lib/components/Pagination.svelte';
  import { paginate } from '../../lib/paginate';

  let entries: AuditEntry[] = [];
  let actor = '';
  let action = '';
  let expanded = 0;
  const PAGE_SIZES=[10,15,20,50];

  let from = '';
  let to = '';

  async function load() {
    try {
      entries = (await api.audit({ actor, action, limit: 200, from, to })).entries; } catch { entries = []; }
    }
  onMount(() => {
    void load();
    const handle = setInterval(() => void load(), 4000);
    return () => clearInterval(handle);
  });
  $: page = 1;
  $: pageSize = 15;
  $: paged = paginate(entries, page, pageSize);
  $: total = entries.length;
  function pageInfo(start:number, end:number,total:number) { return `${start}–${end} of ${total}`; }
  function pretty(value: string) {
    if (!value) return '—';
    try { return JSON.stringify(JSON.parse(value), null, 2); } catch { return value; }
  }
</script>

<svelte:head><title>J3 · Audit trail</title></svelte:head>

<div class="shell">
  <header class="display-header" data-guide="audit-header">
    <div class="brand-mark">J3</div>
    <div class="brand-name">Audit trail</div>
    <div class="brand-sub">Who · when · what · how — for runners and officials</div>
  </header>

  <div class="card">
    <div class="section-title"><h2>Filters</h2></div>
    <div class="field-row">
      <input class="tab" style="color:var(--brand-ink); cursor:text;" placeholder="Filter by actor…" bind:value={actor} on:input={() => void load()} />
      <input class="tab" style="color:var(--brand-ink); cursor:text;" placeholder="Filter by action…" bind:value={action} on:input={() => void load()} />
      <input class="tab" type="date" bind:value={from} on:change={() => void load()} />
      <input class="tab" type="date" bind:value={to} on:change={() => void load()} />
      <button class="btn" on:click={() => void load()}>Refresh</button>
    </div>
  </div>

  <div class="card" style="overflow-x:auto;">
    <table style="width:100%; border-collapse:collapse;">
      <thead><tr class="muted small" style="text-align:left;">
        <th style="padding:8px;">Time</th><th>Actor</th><th>Role</th><th>Action</th><th>Target</th><th>Details</th>
      </tr></thead>
      <tbody>
        {#each paged as entry (entry.id)}
          <tr style="border-top:1px solid var(--brand-border);">
            <td style="padding:8px;"><span class="small muted">{entry.createdAt}</span></td>
            <td style="padding:8px;"><strong>{entry.actorUserId || 'system'}</strong></td>
            <td style="padding:8px;"><span class="badge spectator">{entry.actorRole}</span></td>
            <td style="padding:8px;"><strong>{entry.action}</strong></td>
            <td style="padding:8px;"><span class="small">{entry.targetType}</span> {entry.targetId}</td>
            <td style="padding:8px;"><button class="tab" on:click={() => (expanded = expanded === entry.id ? 0 : entry.id)}>{expanded === entry.id ? 'Hide details' : 'Show details'}</button></td>
          </tr>
          {#if expanded === entry.id}
            <tr style="border-top:1px solid var(--brand-border);">
              <td style="padding:8px;" colspan="6">
                <div class="json-card">
                  <div class="json-row"><strong>Before</strong></div>
                  <pre>{pretty(entry.beforeJson)}</pre>
                  <div class="json-row"><strong>After</strong></div>
                  <pre>{pretty(entry.afterJson)}</pre>
                </div>
              </td>
            </tr>
          {/if}
        {/each}
      </tbody>
    </table>
  </div>

  <Pagination total={entries.length} page={page} pageSize={pageSize} pageSizeOptions={PAGE_SIZES} onChange={(p,s)=>{page=p;pageSize=s;}} />
</div>

<style>
  .field-row { display:flex; gap:10px; flex-wrap:wrap; align-items:center; margin-top:12px; margin-bottom:12px; }
  .field-row > * { min-width:0; }
  .json-card { border:1px solid var(--brand-border); border-radius:0; padding:12px 14px; background:var(--brand-surface-3); margin-top:8px; }
  .json-row { font-size:.68rem; text-transform:uppercase; letter-spacing:.14em; font-weight:900; color:var(--brand-accent); margin-top:8px; }
  .json-row:first-child { margin-top:0; }
  pre { white-space:pre-wrap; font-size:.78rem; margin:6px 0; color:var(--brand-muted); }
</style>
