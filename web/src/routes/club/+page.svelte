<script lang="ts">
  import { onMount } from 'svelte';
  import { getJson, postJson, type InviteLink } from '../../lib/api';
  import { notifySuccess, notifyError } from '../../lib/toast';
  import { guideAction } from '../../lib/guide';

  type Member = { id: string; name: string; clubName: string; role: string; status: string };
  let members: Member[] = [];
  let name = '';
  let email = '';
  let lastInvite: InviteLink | null = null;
  let inviting = false;

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
    try {
      const data = await getJson<{ accounts: Member[] }>('/api/v1/demo/accounts');
      members = data.accounts.filter((account) => account.role === 'fighter');
    } catch { members = []; }
  }
  onMount(load);

  async function addMember() {
    if (!name.trim() || !email.trim()) return;
    inviting = true;
    try {
      // The server derives the club from the verified session, so a leader can
      // never invite a fighter into somebody else's club.
      lastInvite = await postJson<InviteLink>('/api/v1/invites', {
        name: name.trim(), email: email.trim(), role: 'fighter',
      });
      notifySuccess(`${name} invited under your club`);
      guideAction('club-invite');
      name = ''; email = '';
      await load();
    } catch (error) { notifyError(error instanceof Error ? error.message : 'Could not invite member'); }
    finally { inviting = false; }
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
      <button class="btn primary" data-guide="club-invite-submit" style="margin-top:14px;" disabled={inviting} on:click={addMember}>Send club invite</button>
      {#if lastInvite}
        <div class="invite-result" role="status">
          <strong>Invite ready for {lastInvite.email}</strong>
          <div><span class="muted small">Link</span> <code>{lastInvite.inviteUrl ?? `/invite/${lastInvite.token}`}</code></div>
          <div><span class="muted small">Token</span> <code>{lastInvite.token}</code></div>
          <button class="btn" on:click={copyInvite}>Copy link</button>
        </div>
      {/if}
      <a class="btn desktop-only" style="margin-top:10px;" href="/operations">📥 Check-in & import</a>
    </div>
  </div>
</div>

<style>
  .field { display:grid; gap:4px; font-size:.82rem; color:var(--brand-muted); }
  .field input { padding:9px; border-radius:9px; background:var(--brand-surface-3); border:1px solid var(--brand-border); color:var(--brand-ink); }
  .row { display:flex; justify-content:space-between; gap:10px; padding:8px 0; border-top:1px solid var(--brand-border); }
  .invite-result { margin-top:14px; padding:12px; border-radius:12px; background:var(--brand-surface-3); display:grid; gap:6px; }
  code { background:var(--brand-surface-2); padding:2px 6px; border-radius:6px; word-break:break-all; }
  @media (max-width: 820px) { .grid { grid-template-columns:1fr !important; } }
</style>
