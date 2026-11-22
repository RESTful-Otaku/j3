<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import {
    login, passkey, verify, session, landingForRole, uiRole, clearSession, switchDemoUser,
    changePassword,
  } from '../../lib/auth';
  import { setRole } from '../../lib/nav';
  import { api, type DemoCredentialAccount, type DemoCredentialsSheet } from '../../lib/api';
  import { notifyError, notifySuccess } from '../../lib/toast';

  let email = 'tony@okami.j3';
  let password = 'demo';
  let step: 'method' | 'mfa' | 'password' = 'method';
  let demoCode = '';
  let code = '';
  let busy = false;
  let currentPassword = '';
  let newPassword = '';
  let confirmPassword = '';
  let forcedNotice = false;

  let credentials: DemoCredentialsSheet | null = null;
  let credentialError = '';
  let switchingId = '';

  if (session()) goto('/');

  onMount(() => {
    forcedNotice = new URLSearchParams(window.location.search).get('password') === '1';
    const onForced = () => {
      forcedNotice = true;
      step = 'method';
      notifyError('Set a new password before using this account.');
    };
    window.addEventListener('j3-password-change-required', onForced);
    api.credentials()
      .then((sheet) => { credentials = sheet; credentialError = ''; })
      .catch((error: unknown) => {
        const status = error instanceof Error && /^404\b/.test(error.message);
        credentialError = status
          ? 'Demo accounts are only available in a demo venue.'
          : 'Demo accounts could not be loaded right now.';
      });
    return () => window.removeEventListener('j3-password-change-required', onForced);
  });

  async function useAccount(account: DemoCredentialAccount) {
    if (switchingId) return;
    switchingId = account.id;
    try {
      const next = await switchDemoUser(account.id);
      notifySuccess(`Signed in as ${account.name} · ${account.roleLabel}`);
      await goto(landingForRole(next.role));
    } catch (error) {
      notifyError(error instanceof Error ? error.message : 'Could not switch account');
    } finally {
      switchingId = '';
    }
  }

  async function start(provider: string) {
    if (!email.trim()) { notifyError('Enter your email first'); return; }
    busy = true;
    try {
      const result = provider === 'passkey' ? await passkey(email) : await login(email, provider, password);
      demoCode = result.demoCode;
      code = result.demoCode;
      step = 'mfa';
      notifySuccess('MFA code sent');
    } catch (error) {
      notifyError(error instanceof Error ? error.message : 'Sign-in failed');
    } finally { busy = false; }
  }

  async function confirm() {
    busy = true;
    try {
      const result = await verify(email, code);
      setRole(uiRole(result.role) as any);
      if (result.mustChangePassword) {
        currentPassword = password;
        newPassword = '';
        confirmPassword = '';
        step = 'password';
        notifySuccess('Sign-in verified — choose a new password to continue.');
        return;
      }
      notifySuccess(`Welcome, ${result.email}`);
      await goto(landingForRole(result.role));
    } catch (error) {
      notifyError(error instanceof Error ? error.message : 'Verification failed');
    } finally { busy = false; }
  }

  async function finishPasswordChange() {
    if (newPassword.length < 8) { notifyError('Use at least 8 characters'); return; }
    if (newPassword !== confirmPassword) { notifyError('New passwords do not match'); return; }
    busy = true;
    try {
      const current = session();
      await changePassword(currentPassword, newPassword);
      currentPassword = '';
      newPassword = '';
      confirmPassword = '';
      forcedNotice = false;
      notifySuccess('Password updated');
      await goto(landingForRole(current?.role ?? 'spectator'));
    } catch (error) {
      notifyError(error instanceof Error ? error.message : 'Could not change password');
    } finally { busy = false; }
  }

  function logout() { clearSession(); notifySuccess('Signed out'); }
</script>

<svelte:head><title>J3 · Sign in</title></svelte:head>

<div class="auth-shell">
  <div class="auth-card">
    <div class="brand"><span class="brand-mark">J3</span><div><h1>J Cubed</h1><p class="muted small">Tournament creator & operator</p></div></div>
    {#if forcedNotice}<p class="notice">Your account needs a new password before you can continue.</p>{/if}

    {#if step === 'password'}
      <h2>Choose a new password</h2>
      <p class="muted small">The one-time password can only be used once. Replace it now to unlock your account.</p>
      <label class="field">Current password<input type="password" bind:value={currentPassword} /></label>
      <label class="field">New password<input type="password" bind:value={newPassword} /></label>
      <label class="field">Confirm new password<input type="password" bind:value={confirmPassword} /></label>
      <button class="btn primary wide" disabled={busy} on:click={finishPasswordChange}>Update password & continue</button>
    {:else if step === 'method'}
      <label class="field">Email<input type="email" bind:value={email} /></label>
      <label class="field">Password<input type="password" bind:value={password} /></label>
      <button class="btn primary wide" disabled={busy} on:click={() => start('password')}>Sign in with password</button>
      <div class="divider"><span>or continue with</span></div>
      <div class="providers">
        <button class="btn provider" disabled={busy} on:click={() => start('google')}><span class="g">G</span> Google</button>
        <button class="btn provider" disabled={busy} on:click={() => start('apple')}><span class="a"></span> Apple</button>
      </div>
      <button class="btn wide" disabled={busy} on:click={() => start('passkey')}>🔑 Passkey / biometrics</button>
      <p class="muted small">Demo mode: Google/Apple/passkey issue an MFA code shown on screen so the full flow can be tested without external credentials.</p>
    {:else}
      <h2>Two-factor verification</h2>
      <p class="muted small">Enter the 6-digit code for {email}.</p>
      {#if demoCode}<div class="demo-code">Demo code: <strong>{demoCode}</strong></div>{/if}
      <label class="field">6-digit code<input inputmode="numeric" maxlength="6" bind:value={code} /></label>
      <button class="btn primary wide" disabled={busy || code.length !== 6} on:click={confirm}>Verify & sign in</button>
      <button class="btn wide" on:click={() => (step = 'method')}>← Back</button>
    {/if}
    {#if session()}<button class="btn wide" on:click={logout}>Sign out</button>{/if}
  </div>

  <section class="demo-sheet" aria-label="Demo accounts">
    <div class="demo-head">
      <div>
        <h2>Demo accounts — one click to test any role</h2>
        <p class="muted small">Every demo account shares the password <strong>{credentials?.password ?? 'demo'}</strong>. Switch identity to see the exact permissions and displays for that user.</p>
      </div>
      <a class="btn" href="/demo-accounts">View all {credentials?.total ?? ''} accounts →</a>
    </div>

    {#if credentialError}
      <p class="muted">{credentialError}</p>
    {:else if credentials}
      <div class="table-wrap">
        <table>
          <thead><tr><th>Role</th><th>Name</th><th>Email</th><th>Password</th><th>Club</th><th></th></tr></thead>
          <tbody>
            {#each credentials.quickAccounts as account (account.id)}
              <tr>
                <td><span class="badge spectator">{account.roleLabel}</span></td>
                <td>{account.name}</td>
                <td>{account.email}</td>
                <td><code>{account.password}</code></td>
                <td>{account.club || 'Independent'}</td>
                <td><button class="btn primary small" disabled={!!switchingId} on:click={() => useAccount(account)}>Use account</button></td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {:else}
      <p class="muted small">Loading demo accounts…</p>
    {/if}
  </section>
</div>

<style>
  .auth-shell { min-height:100vh; display:grid; place-items:center; align-content:center; gap:20px; padding:20px; }
  .auth-card { width:min(420px, 100%); padding:28px; border-radius:20px; background:var(--brand-surface-2); border:1px solid var(--brand-border); box-shadow:var(--brand-shadow); display:grid; gap:12px; }
  .brand { display:flex; gap:12px; align-items:center; margin-bottom:10px; }
  h1 { font-size:1.5rem; }
  .field { display:grid; gap:5px; font-size:.82rem; color:var(--brand-muted); }
  .field input { padding:11px; border-radius:10px; background:var(--brand-surface-3); border:1px solid var(--brand-border); color:var(--brand-ink); }
  .wide { width:100%; justify-content:center; }
  .providers { display:grid; grid-template-columns:1fr 1fr; gap:10px; }
  .provider { justify-content:center; }
  .g { color:#4285f4; font-weight:800; font-size:1.2rem; }
  .a { width:14px; height:16px; background:var(--brand-ink); clip-path:polygon(50% 0,65% 24%,100% 24%,78% 58%,88% 100%,50% 78%,12% 100%,22% 58%,0 24%,35% 24%); }
  .divider { display:flex; align-items:center; gap:10px; color:var(--brand-muted); font-size:.76rem; }
  .divider::before, .divider::after { content:''; flex:1; height:1px; background:var(--brand-border); }
  .demo-code { padding:10px; border-radius:9px; background:rgba(245,158,11,.12); font-size:.9rem; }
  .notice { padding:10px 12px; border-radius:9px; background:rgba(239,68,68,.12); border:1px solid rgba(239,68,68,.35); font-size:.85rem; }
  .demo-sheet { width:min(1040px, 100%); padding:20px; border-radius:18px; background:var(--brand-surface-2); border:1px solid var(--brand-border); box-shadow:var(--brand-shadow); }
  .demo-head { display:flex; align-items:flex-start; justify-content:space-between; gap:14px; flex-wrap:wrap; margin-bottom:12px; }
  .demo-head h2 { font-size:1.1rem; }
  .table-wrap { overflow-x:auto; }
  table { width:100%; border-collapse:collapse; font-size:.82rem; }
  th, td { text-align:left; padding:8px 10px; border-bottom:1px solid var(--brand-border); white-space:nowrap; }
  th { color:var(--brand-muted); font-weight:600; font-size:.74rem; text-transform:uppercase; letter-spacing:.04em; }
  code { background:var(--brand-surface-3); padding:2px 6px; border-radius:6px; }
</style>
