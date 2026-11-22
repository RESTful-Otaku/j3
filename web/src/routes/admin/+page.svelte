<script lang="ts">
  import { onMount } from 'svelte';
  import { getJson, postJson } from '../../lib/api';
  import { notifySuccess, notifyError } from '../../lib/toast';

  type Account = { id: string; name: string; email: string; role: string; status: string; club: string; roleRequest: string };
  let users: Account[] = [];
  let search = '';
  let role = '';

  async function load() {
    try { users = (await getJson<{ users: Account[] }>('/api/v1/demo/admin/users')).users; }
    catch (error) { notifyError(error instanceof Error ? error.message : 'Could not load accounts'); }
  }
  onMount(load);

  async function status(id: string, next: string) {
    try {
      await postJson(`/api/v1/demo/admin/users/${id}/status`, { status: next });
      notifySuccess(`Account ${next}`);
      await load();
    } catch (error) { notifyError(error instanceof Error ? error.message : 'Update failed'); }
  }
  async function reset(id: string) {
    try {
      const result = await postJson<{ resetToken: string }>(`/api/v1/demo/admin/users/${id}/reset`, {});
      notifySuccess(`Password reset sent · demo token ${result.resetToken}`);
    } catch (error) { notifyError(error instanceof Error ? error.message : 'Reset failed'); }
  }

  $: filtered = users.filter((user) =>
    (!role || user.role === role) &&
    `${user.name} ${user.email} ${user.club}`.toLowerCase().includes(search.toLowerCase()));
</script>

<svelte:head><title>J3 · Account administration</title></svelte:head>

<div class="shell">
  <header class="topbar">
    <div class="brand"><a class="brand-mark" href="/public" aria-label="Public dashboard">J3</a><div><div class="brand-name">Account administration</div><div class="brand-sub">{users.length} accounts · all actions audited</div></div></div>
    <a class="btn" href="/permissions">🔐 Permissions</a>
  </header>

  <div class="card" style="margin-bottom:16px; display:flex; gap:10px; flex-wrap:wrap;">
    <input class="tab" style="color:var(--brand-ink); cursor:text;" placeholder="Search name, email or club…" bind:value={search} />
    <select class="tab" style="color:var(--brand-ink);" bind:value={role}>
      <option value="">All roles</option>
      {#each ['tournament_runner','official','club_leader','fighter','spectator'] as r}<option value={r}>{r}</option>{/each}
    </select>
    <span class="pill">{filtered.length} shown</span>
  </div>

  <div class="card" style="overflow-x:auto;">
    <table style="width:100%; border-collapse:collapse;">
      <thead><tr class="muted small" style="text-align:left;"><th style="padding:8px;">Name</th><th>Email</th><th>Role</th><th>Club</th><th>Status</th><th>Actions</th></tr></thead>
      <tbody>
        {#each filtered as user (user.id)}
          <tr style="border-top:1px solid var(--brand-border);">
            <td style="padding:8px;"><strong>{user.name}</strong></td>
            <td>{user.email}</td><td>{user.role.replaceAll('_',' ')}</td><td>{user.club || 'Independent'}</td>
            <td><span class="badge {user.status === 'active' ? 'fighter' : 'spectator'}">{user.status}</span></td>
            <td style="display:flex; gap:5px; flex-wrap:wrap;">
              {#if user.status === 'active'}<button class="tab" on:click={() => status(user.id, 'suspended')}>Suspend</button>{/if}
              {#if user.status !== 'active'}<button class="tab" on:click={() => status(user.id, 'active')}>Unlock</button>{/if}
              <button class="tab" on:click={() => reset(user.id)}>Reset password</button>
              {#if user.roleRequest}<span class="muted small">requested {user.roleRequest}</span>{/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>

<style>
  @media (max-width: 820px) { table { display:block; overflow-x:auto; white-space:nowrap; } }
</style>
