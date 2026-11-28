<script lang="ts">
  import { onMount } from 'svelte';
  import { getJson, postJson, type InviteLink } from '../../lib/api';
  import { notifySuccess, notifyError } from '../../lib/toast';
  import ConfirmDialog from '../../lib/components/ConfirmDialog.svelte';
  import Pagination from '../../lib/components/Pagination.svelte';
  import { paginate } from '../../lib/paginate';

  type Account = { id: string; name: string; email: string; role: string; status: string; club: string; roleRequest: string };
  let users: Account[] = [];
  let search = '';
  let role = '';
  let selected = new Set<string>();
  let confirmDeleteOpen = false;
  let pendingDeleteIds: string[] = [];
  let adminPage = 1; let adminPageSize = 15;
  const PARTICIPANT_ROLES = ['fighter','spectator']; // Participant roles on UI

  let inviteName = '';
  let inviteEmail = '';
  let inviteRole = 'fighter';
  let inviteClub = '';
  let lastInvite: InviteLink | null = null;
  let inviting = false;

  const inviteRoles = [
    ['admin', 'Admin / owner'],
    ['tournament_runner', 'Tournament runner'],
    ['official', 'Official'],
    ['club_leader', 'Club leader'],
    ['fighter', 'Fighter'],
    ['spectator', 'Spectator'],
  ];

  async function copyInvite() {
    if (!lastInvite) return;
    const url = new URL(lastInvite.inviteUrl ?? `/invite/${lastInvite.token}`, window.location.origin).toString();
    try {
      await navigator.clipboard.writeText(url);
      notifySuccess('Invite link copied');
    } catch {
      notifyError('Copy failed — select and copy the link manually');
    }
  }

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

  async function sendInvite() {
    if (!inviteEmail.trim()) { notifyError('Enter an email address'); return; }
    inviting = true;
    try {
      lastInvite = await postJson<InviteLink>('/api/v1/invites', {
        name: inviteName.trim(), email: inviteEmail.trim(), role: inviteRole,
        ...(inviteClub.trim() ? { clubName: inviteClub.trim() } : {}),
      });
      notifySuccess(`Invite sent to ${inviteEmail.trim()}`);
      inviteName = ''; inviteEmail = ''; inviteClub = '';
      await load();
    } catch (error) {
      notifyError(error instanceof Error ? error.message : 'Could not send invite');
    } finally { inviting = false; }
  }

  function toggleOne(id: string, on: boolean) { const n = new Set(selected); on ? n.add(id) : n.delete(id); selected = n; }
  function toggleAll(on: boolean) { selected = on ? new Set(filtered.map(u=>u.id)) : new Set(); }
  async function deleteUsers(ids: string[]) { for (const id of ids) { await fetch(`/api/v1/demo/admin/users/${id}`, { method: 'DELETE', headers: { 'Content-Type': 'application/json' } }).catch(()=>{}); } notifySuccess(`${ids.length} account(s) deleted`); selected = new Set(); await load(); }
  function askDelete(ids: string[]) { pendingDeleteIds = ids; confirmDeleteOpen = true; }

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

  <section class="card invite-card">
    <h2>Invite an account</h2>
    <p class="muted small">Invitations are bound to your authority: club leaders can only invite into their own club.</p>
    <div class="invite-grid">
      <label class="field">Name<input aria-label="Invite name" bind:value={inviteName} /></label>
      <label class="field">Email<input type="email" aria-label="Invite email" bind:value={inviteEmail} /></label>
      <label class="field">Role<select aria-label="Invite role" bind:value={inviteRole}>
        {#each inviteRoles as [value, label]}<option {value}>{label}</option>{/each}
      </select></label>
      <label class="field">Club name (optional)<input aria-label="Club name" bind:value={inviteClub} /></label>
    </div>
    <button class="btn primary" disabled={inviting} on:click={sendInvite}>Send invite</button>
    {#if lastInvite}
      <div class="invite-result" role="status">
        <strong>Invite ready for {lastInvite.email}</strong>
        <div><span class="muted small">Link</span> <code>{lastInvite.inviteUrl ?? `/invite/${lastInvite.token}`}</code></div>
        <div><span class="muted small">Token</span> <code>{lastInvite.token}</code></div>
        <button class="btn" on:click={copyInvite}>Copy link</button>
      </div>
    {/if}
  </section>

  <div class="card" style="margin-bottom:16px; display:flex; gap:10px; flex-wrap:wrap;">
    <input class="tab" style="color:var(--brand-ink); cursor:text;" placeholder="Search name, email or club…" bind:value={search} />
    <select class="tab" style="color:var(--brand-ink);" bind:value={role}>
      <option value="">All roles</option>
      {#each ['admin','tournament_runner','official','club_leader','fighter','spectator'] as r}<option value={r}>{r}</option>{/each}
    </select>
    <span class="pill">{filtered.length} shown</span>
  </div>

  <div class="card" style="margin-bottom:12px;">
    <span class="pill">{selected.size} of {filtered.length} selected</span>
    <button class="btn" on:click={() => toggleAll(true)}>Select all</button>
    <button class="btn" on:click={() => toggleAll(false)}>Clear</button>
    <button class="btn warn" disabled={selected.size===0} on:click={() => askDelete([...selected])}>Delete selected</button>
  </div>

  <div class="card" style="overflow-x:auto;">
    <table style="width:100%; border-collapse:collapse;">
      <thead><tr class="muted small" style="text-align:left;"><th style="padding:8px;"><input type="checkbox" aria-label="Select all" on:change={(e)=>toggleAll((e.target as HTMLInputElement).checked)} /></th><th style="padding:8px;">Name</th><th>Email</th><th>Role</th><th>Club</th><th>Status</th><th>Actions</th></tr></thead>
      <tbody>
        {#each paginate(filtered, adminPage, adminPageSize) as user (user.id)}
          <tr style="border-top:1px solid var(--brand-border);">
            <td style="padding:8px;"><input type="checkbox" checked={selected.has(user.id)} on:change={(e)=>toggleOne(user.id,(e.target as HTMLInputElement).checked)} aria-label={"Select "+user.name} /></td>
            <td style="padding:8px;"><strong>{user.name}</strong></td>
            <td>{user.email}</td><td>{user.role.replaceAll('_',' ')}</td><td>{user.club || 'Independent'}</td>
            <td><span class="badge {user.status === 'active' ? 'fighter' : 'spectator'}">{user.status}</span></td>
            <td style="display:flex; gap:5px; flex-wrap:wrap;">
              {#if user.status === 'active'}<button class="tab" on:click={() => status(user.id, 'suspended')}>Suspend</button>{/if}
              {#if user.status !== 'active'}<button class="tab" on:click={() => status(user.id, 'active')}>Unlock</button>{/if}
              <button class="tab" on:click={() => reset(user.id)}>Reset password</button>
              <button class="tab" on:click={() => askDelete([user.id])}>Delete</button>
              {#if user.roleRequest}<span class="muted small">requested {user.roleRequest}</span>{/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
  <Pagination total={filtered.length} page={adminPage} pageSize={adminPageSize} onChange={(p,s)=>{adminPage=p;adminPageSize=s;}} />
</div>

<ConfirmDialog bind:open={confirmDeleteOpen} title={pendingDeleteIds.length + " account(s) will be permanently deleted"} message="This cannot be undone. Temporary undo is not available after confirmation." confirmLabel="Delete" danger onConfirm={() => deleteUsers(pendingDeleteIds)} />

<style>
  .invite-card { margin-bottom:16px; }
  .invite-grid { display:grid; grid-template-columns:repeat(auto-fit, minmax(180px, 1fr)); gap:12px; margin:12px 0; }
  .field { display:grid; gap:4px; font-size:.8rem; color:var(--brand-muted); }
  .field input, .field select { padding:9px; border-radius:9px; background:var(--brand-surface-3); border:1px solid var(--brand-border); color:var(--brand-ink); }
  .invite-result { margin-top:14px; padding:12px; border-radius:12px; background:var(--brand-surface-3); display:grid; gap:6px; }
  code { background:var(--brand-surface-2); padding:2px 6px; border-radius:6px; word-break:break-all; }
  @media (max-width: 820px) { table { display:block; overflow-x:auto; white-space:nowrap; } }
</style>
