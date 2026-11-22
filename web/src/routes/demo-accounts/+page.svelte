<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { api, type DemoCredentialAccount, type DemoCredentialsSheet } from '../../lib/api';
  import { landingForRole, switchDemoUser } from '../../lib/auth';
  import { notifyError, notifySuccess } from '../../lib/toast';

  let sheet: DemoCredentialsSheet | null = null;
  let error = '';
  let loading = true;
  let search = '';
  let roleFilter = '';
  let clubFilter = '';
  let switchingId = '';

  onMount(() => {
    api.credentials()
      .then((value) => { sheet = value; loading = false; })
      .catch((cause: unknown) => {
        const production = cause instanceof Error && /^404\b/.test(cause.message);
        error = production
          ? 'Demo accounts are only available when J3 runs in demo mode. Ask your organiser for an invitation instead.'
          : 'The demo account sheet could not be loaded. Check that the venue server is running.';
        loading = false;
      });
  });

  $: accounts = sheet?.accounts ?? [];
  $: roles = [...new Set(accounts.map((account) => account.role))].sort();
  $: clubs = [...new Set(accounts.map((account) => account.club).filter(Boolean))].sort();
  $: filtered = accounts.filter((account) => {
    const term = `${account.name} ${account.email} ${account.club} ${account.roleLabel}`.toLowerCase();
    return (!search || term.includes(search.toLowerCase()))
      && (!roleFilter || account.role === roleFilter)
      && (!clubFilter || account.club === clubFilter);
  });

  async function useAccount(account: DemoCredentialAccount) {
    if (switchingId) return;
    switchingId = account.id;
    try {
      const next = await switchDemoUser(account.id);
      notifySuccess(`Signed in as ${account.name} · ${account.roleLabel}`);
      await goto(landingForRole(next.role));
    } catch (cause) {
      notifyError(cause instanceof Error ? cause.message : 'Could not switch account');
    } finally {
      switchingId = '';
    }
  }
</script>

<svelte:head><title>J3 · Demo accounts</title></svelte:head>

<div class="shell">
  <header class="topbar">
    <div class="brand"><a class="brand-mark" href="/public" aria-label="Public dashboard">J3</a><div><div class="brand-name">Demo accounts</div><div class="brand-sub">One click to test any role, club or participant view</div></div></div>
    <a class="btn" href="/login">← Sign in</a>
  </header>

  <h1 class="page-title">Demo accounts</h1>

  {#if error}
    <div class="card"><h2>Demo accounts unavailable</h2><p class="muted">{error}</p></div>
  {:else if loading}
    <div class="card"><p class="muted">Loading demo accounts…</p></div>
  {:else if sheet}
    <div class="card password-card">
      <strong>Shared demo password:</strong> <code>{sheet.password}</code>
      <span class="muted small">{sheet.total} seeded accounts · every account uses this password.</span>
    </div>

    <div class="card filters">
      <label class="field">Search<input aria-label="Search demo accounts" placeholder="Name, email, club or role…" bind:value={search} /></label>
      <label class="field">Role<select aria-label="Filter by role" bind:value={roleFilter}>
        <option value="">All roles</option>
        {#each roles as role}<option value={role}>{sheet.accounts.find((account) => account.role === role)?.roleLabel ?? role}</option>{/each}
      </select></label>
      <label class="field">Club<select aria-label="Filter by club" bind:value={clubFilter}>
        <option value="">All clubs</option>
        {#each clubs as club}<option value={club}>{club}</option>{/each}
      </select></label>
      <span class="pill">{filtered.length} shown</span>
    </div>

    <div class="card table-wrap">
      <table>
        <thead><tr><th>Role</th><th>Name</th><th>Email</th><th>Password</th><th>Club</th><th>Status</th><th></th></tr></thead>
        <tbody>
          {#each filtered as account (account.id)}
            <tr>
              <td><span class="badge spectator">{account.roleLabel}</span></td>
              <td><strong>{account.name}</strong></td>
              <td>{account.email}</td>
              <td><code>{account.password}</code></td>
              <td>{account.club || 'Independent'}</td>
              <td>{account.status}</td>
              <td><button class="btn primary small" disabled={!!switchingId} on:click={() => useAccount(account)}>Use account</button></td>
            </tr>
          {:else}
            <tr><td colspan="7" class="muted">No accounts match those filters.</td></tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<style>
  .password-card { display:flex; align-items:center; gap:10px; flex-wrap:wrap; margin-bottom:14px; }
  code { background:var(--brand-surface-3); padding:3px 8px; border-radius:7px; font-weight:700; }
  .filters { display:flex; align-items:end; gap:12px; flex-wrap:wrap; margin-bottom:14px; }
  .field { display:grid; gap:4px; font-size:.8rem; color:var(--brand-muted); min-width:180px; }
  .field input, .field select { padding:9px; border-radius:9px; background:var(--brand-surface-3); border:1px solid var(--brand-border); color:var(--brand-ink); }
  .table-wrap { overflow-x:auto; }
  table { width:100%; border-collapse:collapse; font-size:.82rem; }
  th, td { text-align:left; padding:8px 10px; border-bottom:1px solid var(--brand-border); white-space:nowrap; }
  th { color:var(--brand-muted); font-size:.74rem; text-transform:uppercase; letter-spacing:.04em; }
  @media (max-width: 820px) { .field { min-width:0; width:100%; } }
</style>
