<script lang="ts">
  import { onMount } from 'svelte';
  import { getJson, postJson } from '../../lib/api';
  import { notifySuccess, notifyError } from '../../lib/toast';
  import { guideAction } from '../../lib/guide';

  type Member = { id: string; name: string; clubName: string; role: string; status: string };
  let members: Member[] = [];
  let name = '';
  let email = '';

  async function load() {
    try {
      const data = await getJson<{ accounts: Member[] }>('/api/v1/demo/accounts');
      members = data.accounts.filter((account) => account.role === 'fighter');
    } catch { members = []; }
  }
  onMount(load);

  async function addMember() {
    if (!name.trim() || !email.trim()) return;
    try {
      await postJson('/api/v1/tournaments', { name: `${name} entry`, disciplines: [{ name: 'club entry' }], categories: [{ name: 'club' }], invites: [{ email, role: 'fighter' }] });
      notifySuccess(`${name} invited under your club`);
      guideAction('club-invite');
      name = ''; email = '';
    } catch (error) { notifyError(error instanceof Error ? error.message : 'Could not invite member'); }
  }
</script>

<svelte:head><title>J3 · My club</title></svelte:head>

<div class="shell">
  <header class="topbar"><div class="brand"><a class="brand-mark" href="/public" aria-label="Public">J3</a><div><div class="brand-name">My club</div><div class="brand-sub">Manage your own members and view public boards</div></div></div>
    <a class="btn" href="/public">📺 Public board</a></header>

  <div class="grid" style="grid-template-columns:1fr .8fr;">
    <div class="card" data-guide="club-members"><h2>Club members</h2>
      <div style="display:grid; gap:8px; margin-top:12px; max-height:520px; overflow:auto;">
        {#each members.slice(0, 30) as member (member.id)}
          <div class="row"><strong>{member.name}</strong><span class="muted small">{member.clubName || 'Independent'} · {member.status}</span></div>
        {:else}<span class="muted">No members loaded yet.</span>{/each}
      </div>
    </div>
    <div class="card"><h2>Invite a member</h2>
      <label class="field">Name<input bind:value={name} /></label>
      <label class="field" style="margin-top:10px;">Email<input type="email" bind:value={email} /></label>
      <button class="btn primary" data-guide="club-invite-submit" style="margin-top:14px;" on:click={addMember}>Send club invite</button>
      <a class="btn desktop-only" style="margin-top:10px;" href="/operations">📥 Check-in & import</a>
    </div>
  </div>
</div>

<style>
  .field { display:grid; gap:4px; font-size:.82rem; color:var(--brand-muted); }
  .field input { padding:9px; border-radius:9px; background:var(--brand-surface-3); border:1px solid var(--brand-border); color:var(--brand-ink); }
  .row { display:flex; justify-content:space-between; gap:10px; padding:8px 0; border-top:1px solid var(--brand-border); }
  @media (max-width: 820px) { .grid { grid-template-columns:1fr !important; } }
</style>
