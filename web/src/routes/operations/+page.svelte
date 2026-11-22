<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '../../lib/api';
  import { notifySuccess, notifyError } from '../../lib/toast';
  import { guideAction } from '../../lib/guide';

  type Entrant = { id: string; fighter_id: string; category_id: string; status: string };
  type Fighter = { id: string; display_name: string; weight_kg: number };
  let entrants: Entrant[] = [];
  let fighters: Fighter[] = [];
  let teams: { club: string; gold: number; silver: number; bronze: number; points: number }[] = [];
  let name = '';
  let club = '';
  let weight = 70;
  let message = '';
  let search = '';
  let statusFilter = 'all';

  onMount(async () => {
    const [entrantData, fighterData, teamData] = await Promise.all([
      api.query({ table: 'entrants', limit: 500 }),
      api.query({ table: 'fighters', limit: 500 }),
      api.teamRankings(),
    ]);
    entrants = entrantData.rows as unknown as Entrant[];
    fighters = fighterData.rows as unknown as Fighter[];
    teams = teamData.teams;
  });

  function fighterName(id: string) {
    return fighters.find((fighter) => fighter.id === id)?.display_name ?? id;
  }
  $: visibleEntrants = entrants.filter((entrant) => {
    const matchesSearch = !search.trim() || fighterName(entrant.fighter_id).toLowerCase().includes(search.trim().toLowerCase());
    const matchesStatus = statusFilter === 'all' || entrant.status === statusFilter;
    return matchesSearch && matchesStatus;
  });
  async function downloadCsv(type: string) {
    const result = await api.exportCsv(type);
    if (!result.ok) notifyError(result.error ?? `Export (${type}) failed`);
  }
  async function checkIn(entrant: Entrant, event: Event) {
    const input = event.target as HTMLInputElement;
    try {
      const result = await api.checkIn(entrant.id, Number(input.value));
      message = result.outOfClass
        ? `⚠ ${fighterName(entrant.fighter_id)} is out of class at ${result.weightKg} kg`
        : `✅ ${fighterName(entrant.fighter_id)} checked in at ${result.weightKg} kg`;
      notifySuccess(message);
      entrants = entrants.map((item) => item.id === entrant.id
        ? { ...item, status: result.outOfClass ? 'flagged' : 'checked_in' } : item);
      guideAction('check-in', { entrantId: entrant.id, weightKg: result.weightKg });
    } catch (error) {
      notifyError(error instanceof Error ? error.message : 'Check-in failed');
    }
  }
  async function addFighter() {
    if (!name.trim()) return;
    try {
      await api.importEntries([{ name, club, weightKg: weight }]);
      notifySuccess(`${name} added`);
      guideAction('fighter-added');
      name = '';
    } catch (error) {
      notifyError(error instanceof Error ? error.message : 'Import failed');
    }
  }
</script>

<svelte:head><title>J3 · Operations</title></svelte:head>

<div class="shell">
  <header class="topbar">
    <div class="brand"><a class="brand-mark" href="/public" aria-label="Public dashboard">J3</a><div><div class="brand-name">Check-in & weigh-in</div><div class="brand-sub">Arrivals, official weights, club rankings and data exchange</div></div></div>
    <a class="btn" href="/control">🖥️ Displays</a>
  </header>

  <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(300px,1fr));">
    <div class="card">
      <div class="eyebrow">Arrivals</div><h2>Arrivals & weigh-in</h2>
      <p class="muted small">Search for an arrival, enter their official weight and press Enter. Out-of-class athletes are flagged automatically.</p>
      {#if message}<p class="small" style="margin-top:10px;">{message}</p>{/if}
      <div style="display:flex; gap:8px; margin-top:12px; flex-wrap:wrap;">
        <input class="tab" style="flex:1; min-width:140px; color:var(--brand-ink); cursor:text;" placeholder="Search fighter…" bind:value={search} aria-label="Search arrivals" />
        <select class="tab" bind:value={statusFilter} aria-label="Filter by check-in status" style="color:var(--brand-ink);">
          <option value="all">All status</option>
          <option value="checked_in">Checked in</option>
          <option value="flagged">Flagged</option>
          <option value="pending">Pending</option>
        </select>
      </div>
      <div style="margin-top:12px; max-height:420px; overflow:auto;">
        {#each visibleEntrants as entrant, entrantIndex (entrant.id)}
          <div class="row">
            <span><strong>{fighterName(entrant.fighter_id)}</strong><span class="muted small" style="display:block;">{entrant.status.replaceAll('_', ' ')}</span></span>
            <input class="tab" type="number" step="0.1" placeholder="kg" data-guide={entrantIndex === 0 ? 'operations-weigh-input' : undefined} style="width:80px; color:var(--brand-ink);" aria-label="Weight for {fighterName(entrant.fighter_id)}" on:change={(event) => checkIn(entrant, event)} />
          </div>
        {:else}<p class="muted small">No entrants match this filter.</p>{/each}
      </div>
    </div>

    <div class="card">
      <div class="eyebrow">Club competition</div><h2>Team rankings</h2>
      <p class="muted small">Gold 3 · Silver 2 · Bronze 1</p>
      <table style="width:100%; border-collapse:collapse; margin-top:12px;">
        <thead><tr class="muted small" style="text-align:left;"><th>Club</th><th>🥇</th><th>🥈</th><th>🥉</th><th>Pts</th></tr></thead>
        <tbody>
          {#each teams as team (team.club)}
            <tr style="border-top:1px solid var(--brand-border);"><td>{team.club}</td><td>{team.gold}</td><td>{team.silver}</td><td>{team.bronze}</td><td><strong>{team.points}</strong></td></tr>
          {:else}<tr><td colspan="5" class="muted">No medals awarded yet.</td></tr>{/each}
        </tbody>
      </table>
    </div>

    <div class="card">
      <div class="eyebrow">Bulk data</div><h2>Import & export</h2>
      <label class="field" style="display:grid; gap:5px; margin-top:12px;">Name<input bind:value={name} /></label>
      <label class="field" style="display:grid; gap:5px; margin-top:10px;">Club<input bind:value={club} /></label>
      <label class="field" style="display:grid; gap:5px; margin-top:10px;">Weight (kg)<input type="number" bind:value={weight} /></label>
      <button class="btn primary" style="margin-top:12px;" on:click={addFighter}>Add fighter</button>
      <div class="action-row" style="margin-top:18px;">
        <button class="btn" on:click={() => void downloadCsv('fights')}>⬇ Fights CSV</button>
        <button class="btn" on:click={() => void downloadCsv('entrants')}>⬇ Entrants CSV</button>
        <button class="btn" on:click={() => void downloadCsv('standings')}>⬇ Standings CSV</button>
      </div>
    </div>
  </div>
</div>

<style>
  .row { display:flex; justify-content:space-between; align-items:center; gap:10px; padding:9px 0; border-bottom:1px solid var(--brand-border); }
</style>
