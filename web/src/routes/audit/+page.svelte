<script lang="ts">
  import { onMount } from 'svelte';
  import { api, type AuditEntry } from '../../lib/api';

  let entries: AuditEntry[] = [];
  let actor = '';
  let action = '';
  let expanded = 0;

  async function load() {
    try { entries = (await api.audit({ actor, action, limit: 250 })).entries; } catch { /* keep last frame */ }
  }
  onMount(() => {
    void load();
    const handle = setInterval(() => void load(), 4000);
    return () => clearInterval(handle);
  });
  function pretty(value: string) {
    if (!value) return '—';
    try { return JSON.stringify(JSON.parse(value), null, 2); } catch { return value; }
  }
</script>

<svelte:head><title>J3 · Audit trail</title></svelte:head>

<div class="shell">
  <header class="topbar">
    <div class="brand" data-guide="audit-header"><div class="brand-mark">J3</div><div><div class="brand-name">Audit trail</div><div class="brand-sub">Who · when · what · how — for runners and officials</div></div></div>
    <a class="btn" href="/control" target="_blank" rel="noreferrer">🖥️ Displays ↗</a>
  </header>

  <div class="card">
    <div style="display:flex; gap:10px; flex-wrap:wrap;">
      <input bind:value={actor} placeholder="Filter by actor ID…" on:input={() => void load()} />
      <input bind:value={action} placeholder="Filter by action…" on:input={() => void load()} />
      <button class="btn" on:click={() => void load()}>Refresh</button>
    </div>
    <div style="display:grid; gap:8px; margin-top:16px;">
      {#each entries as entry (entry.id)}
        <div class="entry">
          <div style="display:flex; justify-content:space-between; gap:12px; flex-wrap:wrap;">
            <div><strong>{entry.action}</strong> <span class="badge {entry.actorRole === 'tournament_runner' ? 'runner' : 'spectator'}">{entry.actorRole}</span></div>
            <span class="muted small">{entry.createdAt} · by {entry.actorUserId || 'system'}</span>
          </div>
          <div class="muted small">{entry.targetType} · {entry.targetId}</div>
          {#if expanded === entry.id}
            <pre>{pretty(entry.beforeJson)} → {pretty(entry.afterJson)}</pre>
          {/if}
          <button class="tab" on:click={() => (expanded = expanded === entry.id ? 0 : entry.id)}>{expanded === entry.id ? 'Hide change' : 'View change'}</button>
        </div>
      {:else}<div class="placeholder">No audit entries match these filters.</div>{/each}
    </div>
  </div>
</div>

<style>
  input { padding:9px 11px; border-radius:9px; border:1px solid var(--brand-border); background:var(--brand-surface-3); color:var(--brand-ink); }
  .entry { padding:12px 14px; border-radius:10px; border:1px solid var(--brand-border); background:var(--brand-surface-3); }
  pre { white-space:pre-wrap; font-size:.78rem; margin:10px 0; color:var(--brand-muted); }
</style>
